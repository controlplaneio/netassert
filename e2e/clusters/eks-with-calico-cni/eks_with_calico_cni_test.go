package e2e

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/go-hclog"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/controlplaneio/netassert/v2/internal/data"
	"github.com/controlplaneio/netassert/v2/internal/engine"
	"github.com/controlplaneio/netassert/v2/internal/kubeops"
	"github.com/controlplaneio/netassert/v2/internal/logger"
)

var envVarName = `AWS_EKS_E2E_CALICO_CNI`

const (
	suffixLength            = 9 // suffix length of the random string to be appended to the container name
	snifferContainerImage   = "docker.io/controlplane/netassertv2-packet-sniffer:1.0.0"
	snifferContainerPrefix  = "netassertv2-sniffer"
	scannerContainerImage   = "docker.io/controlplane/netassertv2-l4-client:1.0.0"
	scannerContainerPrefix  = "netassertv2-client"
	pauseInSeconds          = 1 // time to pause before each test case
	packetCaputureInterface = `eth0`
	testCasesFile           = `../../manifests/test-cases.yaml`
	resultFile              = "result.log" // where we write the results
)

func TestEKSWith_AWS_VPC_CNI(t *testing.T) {
	var (
		region         = "eu-west-1"
		clusterName    = "netassert-calico-" + strings.ToLower(random.UniqueId())
		clusterVersion = "1.25"
		kubeConfig     = clusterName + "-kubeconfig"
		timeout        = 5 * time.Minute
		pollTime       = 20 * time.Second
	)

	fmt.Printf("Region=%s\n, EKSVersion=%s\n, ClusterName=%s\n, kubeConfig=%s\n",
		region, clusterVersion, clusterName, kubeConfig)

	val := os.Getenv(envVarName)

	if val == "" {
		t.Skipf("skipping test associated with %q because %q environment was not set",
			clusterName, envVarName)
	}

	tfOpt := &terraform.Options{
		TerraformDir: "./terraform",
		Vars: map[string]interface{}{
			"region":          region,
			"cluster_version": clusterVersion,
			"cluster_name":    clusterName,
			"kubeconfig_file": kubeConfig,
			"desired_size":    0,
			"node_group_name": "group",
		},
	}

	// we first spin up a cluster with zero worker nodes
	terraformOptions := terraform.WithDefaultRetryableErrors(t, tfOpt)

	// terraform init
	if _, err := terraform.InitE(t, terraformOptions); err != nil {
		t.Fatalf("failed to iniitialise terraform: %s", err)
	}

	if _, err := terraform.ApplyE(t, terraformOptions); err != nil {
		t.Fatalf("failed to run terraform apply: %s", err)
	}

	// clean up the resources later
	defer terraform.Destroy(t, terraformOptions)

	// once the cluster is ready, we need to follow the instructions here
	// https://docs.tigera.io/calico/3.25/getting-started/kubernetes/managed-public-cloud/eks
	ctx := context.Background()
	kubeConfigPath := "./terraform/" + kubeConfig

	lg := logger.NewHCLogger("INFO", "netassertv2-e2e-calico", os.Stdout)

	svc, err := kubeops.NewServiceFromKubeConfigFile(kubeConfigPath, lg)
	if err != nil {
		t.Fatalf("Failed to build kubernetes client: %s", err)
	}

	// kubectl delete daemonset -n kube-system aws-node
	err = svc.Client.AppsV1().DaemonSets("kube-system").Delete(
		ctx, "aws-node", metav1.DeleteOptions{})

	if err != nil {
		if !apierrors.IsNotFound(err) {
			t.Fatalf("failed to delete daemonset aws-node in the kube-system namespace")
		}
	}

	svc.Log.Info("AWS-CNI", "msg", "Deleted daemonset aws-node in the kube-system namespace")
	// create a new Kubernetes client using the terratest package
	options := k8s.NewKubectlOptions("", kubeConfigPath, "")

	// we now apply calico CNI manifest
	k8s.KubectlApply(t, options, "./calico-3.26.4.yaml")

	// update the desired_size variable to 3
	tfOpt.Vars["desired_size"] = 3
	tfOpt.Vars["node_group_name"] = "calico"

	newTFOptions := terraform.WithDefaultRetryableErrors(t, tfOpt)
	// terraform apply the new options
	// this terraform apply should scale up the worker nodes with Calico CNI
	if _, err := terraform.InitAndApplyE(t, newTFOptions); err != nil {
		t.Fatalf("failed to run terraform init and apply: %s", err)
	}

	// let's wait for all the nodes to be ready, these worker nodes should use calico
	// as the CNI
	k8s.WaitUntilAllNodesReady(t, options, 12, 1*time.Minute)

	// ping the cluster Endpoint and see if things are working as expected
	if err := svc.PingHealthEndpoint(ctx, "/healthz"); err != nil {
		t.Fatalf("Failed to ping kubernetes server: %s", err)
	}

	type k8sManifest struct {
		name      string
		namespace string
		filePath  string
		objType   string
	}

	k8sManifests := []k8sManifest{
		{
			name:      "fluentd",
			namespace: "fluentd",
			filePath:  "../../manifests/daemonset.yaml",
			objType:   "daemonset",
		},
		{
			name:      "echoserver",
			namespace: "echoserver",
			filePath:  "../../manifests/deployment.yaml",
			objType:   "deployment",
		},
		{
			name:      "busybox",
			namespace: "busybox",
			filePath:  "../../manifests/deployment.yaml",
			objType:   "deployment",
		},
		{
			name:      "pod1",
			namespace: "pod1",
			filePath:  "../../manifests/pod1-pod2.yaml",
			objType:   "pod",
		},
		{
			name:      "pod2",
			namespace: "pod2",
			filePath:  "../../manifests/pod1-pod2.yaml",
			objType:   "pod",
		},
		{
			name:      "web",
			namespace: "web",
			filePath:  "../../manifests/statefulset.yaml",
			objType:   "statefulset",
		},
	}

	// we apply all the manifests and then check for a Pod in the resource to be ready
	for _, v := range k8sManifests {
		// apply the manifest
		k8s.KubectlApply(t, options, v.filePath)
		// wait for the object to have at least one pod healthy
		err = waitForPodInResourceReady(svc, v.name, v.namespace,
			v.objType, pollTime, timeout)
		// if the Pods in the objects are not ready within allocated time, fail the test
		if err != nil {
			t.Fatal(err)
		}
	}

	netAssertTestCases, err := data.ReadTestsFromFile(testCasesFile)
	if err != nil {
		t.Fatal(err)
	}

	// run the tests
	testRunner := engine.New(svc, lg)

	testRunner.RunTests(
		ctx,                    // context to use
		netAssertTestCases,     // net assert test cases
		snifferContainerPrefix, // prefix used for the sniffer container name
		snifferContainerImage,  // sniffer container image location
		scannerContainerPrefix, // scanner container prefix used in the container name
		scannerContainerImage,  // scanner container image location
		suffixLength,           // length of random string that will be appended to the snifferContainerPrefix and scannerContainerPrefix
		time.Duration(pauseInSeconds)*time.Second, // time to pause between each test
		packetCaputureInterface,                   // the interface used by the sniffer image to capture traffic
	)

	fh, err := os.Create(resultFile)
	if err != nil {
		t.Log("failed to create file", resultFile, err)
		t.Fail()
	}

	mr := io.MultiWriter(fh, os.Stdout)
	lg = logger.NewHCLogger("INFO", "netassertv2-e2e", mr)

	failedTestCases := 0

	for _, v := range netAssertTestCases {
		// increment the no. of test cases
		if v.Pass {
			lg.Info("✅ Test Result", "Name", v.Name, "Pass", v.Pass)
			continue
		}

		lg.Info("❌ Test Result", "Name", v.Name, "Pass", v.Pass, "FailureReason", v.FailureReason)
		failedTestCases++
	}

	if failedTestCases > 0 {
		t.Fatal("e2e tests have failed")
	}
}

func waitForPodInResourceReady(svc *kubeops.Service, name, namespace, resourceType string,
	poll, timeout time.Duration,
) error {
	var fn func(context.Context, string, string) (*corev1.Pod, error)

	switch strings.ToLower(resourceType) {
	case "deployment":
		fn = svc.GetPodInDeployment
	case "daemonset":
		fn = svc.GetPodInDaemonSet
	case "statefulset":
		fn = svc.GetPodInStatefulSet
	case "pod":
		fn = svc.GetPod
	default:
		return fmt.Errorf("unsupported resource type %q", resourceType)
	}

	timeOutCh := time.After(timeout)
	ticker := time.NewTicker(poll)

	for {
		select {
		case <-timeOutCh:
			return fmt.Errorf("timed out getting pod for %q - %s/%s, timeout duration=%v", resourceType,
				namespace, name, timeout.String())
		case <-ticker.C:
			_, err := fn(context.Background(), name, namespace)
			if err == nil {
				svc.Log.Info("polling", "found", hclog.Fmt("name=%s namespace=%s "+
					" resourceType=%s", name, namespace, resourceType))
				return nil
			}
			svc.Log.Info("polling", "name", name, "namespace", namespace)
		}
	}
}
