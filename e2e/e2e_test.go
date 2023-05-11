package e2e

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/controlplaneio/netassert/v2/e2e/helpers"
	"github.com/controlplaneio/netassert/v2/internal/data"
	"github.com/controlplaneio/netassert/v2/internal/engine"
	"github.com/controlplaneio/netassert/v2/internal/kubeops"
	"github.com/controlplaneio/netassert/v2/internal/logger"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/hashicorp/go-hclog"
)

const (
	suffixLength            = 9 // suffix length of the random string to be appended to the container name
	snifferContainerImage   = "docker.io/controlplane/netassertv2-packet-sniffer:latest"
	snifferContainerPrefix  = "netassertv2-sniffer"
	scannerContainerImage   = "docker.io/controlplane/netassertv2-l4-client:latest"
	scannerContainerPrefix  = "netassertv2-client"
	pauseInSeconds          = 1 // time to pause before each test case
	packetCaputureInterface = `eth0`
	testCasesFile           = `./manifests/test-cases.yaml`
	resultFile              = "result.log" // where we write the results
)

var (
	envVarGKEWithVPC    = `GKE_VPC_E2E_TESTS`
	envVarGKEWithDPv2   = `GKE_DPV2_E2E_TESTS`
	envVarEKSWithVPC    = `EKS_VPC_E2E_TESTS`
	envVarEKSWithCalico = `EKS_CALICO_E2E_TESTS`
)

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestGKEWithVPC(t *testing.T) {
	t.Parallel()

	if os.Getenv(envVarGKEWithVPC) == "" {
		t.Skipf("skipping test associated with GKE VPC as %q environment variable was not set", envVarGKEWithVPC)
	}

	gke := helpers.NewGKECluster(t, "./clusters/gke-vpc", "vpc", helpers.VPC)
	createTestDestroy(t, gke)
}

func TestGKEWithDataPlaneV2(t *testing.T) {
	t.Parallel()

	if os.Getenv(envVarGKEWithDPv2) == "" {
		t.Skipf("skipping test associated with GKE DataPlaneV2 as %q environment variable was not set", envVarGKEWithDPv2)
	}

	gke := helpers.NewGKECluster(t, "./clusters/gke-dataplanev2", "dataplanev2", helpers.DataPlaneV2)
	createTestDestroy(t, gke)
}

func TestEKSWithVPC(t *testing.T) {
	t.Parallel()

	pn := os.Getenv(envVarEKSWithVPC)
	if pn == "" {
		t.Skipf("skipping test associated with EKS VPC CNI as %q environment variable was not set", envVarEKSWithVPC)
	}

	eks := helpers.NewEKSCluster(t, "./clusters/eks-with-vpc-cni/terraform", "vpc", helpers.VPC)
	createTestDestroy(t, eks)
}

func TestEKSWithCalico(t *testing.T) {
	t.Parallel()

	pn := os.Getenv(envVarEKSWithCalico)
	if pn == "" {
		t.Skipf("skipping test associated with EKS Calico CNI as environment %q was not set", envVarEKSWithCalico)
	}

	eks := helpers.NewEKSCluster(t, "./clusters/eks-with-calico-cni/terraform", "calico", helpers.Calico)
	createTestDestroy(t, eks)
}

func createTestDestroy(t *testing.T, gc helpers.GenericCluster) {
	defer gc.Destroy(t) // safe to call also when the cluster has not been created
	gc.Create(t)

	ctx := context.Background()
	kubeConfig := gc.KubeConfigGet()
	svc, err := kubeops.NewServiceFromKubeConfigFile(kubeConfig, hclog.NewNullLogger())
	if err != nil {
		t.Logf("Failed to build kubernetes client: %s", err)
		t.Fatal(err)
	}

	if err := svc.PingHealthEndpoint(ctx, "/healthz"); err != nil {
		t.Logf("Failed to ping kubernetes server: %s", err)
		t.Fatal(err)
	}

	t.Log("successfully pinged the k8s server")

	// create a new Kubernetes client
	options := k8s.NewKubectlOptions("", kubeConfig, "")

	// let's wait for all the nodes to be ready
	k8s.WaitUntilAllNodesReady(t, options, 20, 1*time.Minute)

	timeout := 5 * time.Minute
	pollTime := 30 * time.Second
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
			filePath:  "./manifests/daemonset.yaml",
			objType:   "daemonset",
		},
		{
			name:      "echoserver",
			namespace: "echoserver",
			filePath:  "./manifests/deployment.yaml",
			objType:   "deployment",
		},
		{
			name:      "busybox",
			namespace: "busybox",
			filePath:  "./manifests/deployment.yaml",
			objType:   "deployment",
		},
		{
			name:      "pod1",
			namespace: "pod1",
			filePath:  "./manifests/pod1-pod2.yaml",
			objType:   "pod",
		},
		{
			name:      "pod2",
			namespace: "pod2",
			filePath:  "./manifests/pod1-pod2.yaml",
			objType:   "pod",
		},
		{
			name:      "web",
			namespace: "web",
			filePath:  "./manifests/statefulset.yaml",
			objType:   "statefulset",
		},
	}

	// we apply all the manifests and then run

	for _, v := range k8sManifests {
		// apply the manifest
		k8s.KubectlApply(t, options, v.filePath)
		// wait for the object to have at least one pod healthy
		err = svc.WaitForPodInResourceReady(v.name, v.namespace,
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

	// run the tests without network policies
	runTests(ctx, t, svc, netAssertTestCases)

	if gc.SkipNetPolTests() {
		return
	}

	// create the network policies
	k8s.KubectlApply(t, options, "./manifests/networkpolicies.yaml")

	// read the tests again for a fresh start
	netAssertTestCases, err = data.ReadTestsFromFile(testCasesFile)
	if err != nil {
		t.Fatal(err)
	}

	// set the exit to 1 since this time the network policies will block the traffic
	for _, tc := range netAssertTestCases {
		tc.ExitCode = 1
	}

	// run the tests with network policies
	runTests(ctx, t, svc, netAssertTestCases)
}

func runTests(ctx context.Context, t *testing.T, svc *kubeops.Service, netAssertTestCases data.Tests) {
	lg := logger.NewHCLogger("INFO", "netassertv2-e2e", os.Stdout)
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
		t.Fatal(err)
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
		t.Fatal("e2e tests have failed", err)
	}
}
