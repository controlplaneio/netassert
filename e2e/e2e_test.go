package e2e

import (
	"bytes"
	"context"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/controlplaneio/netassert/v2/e2e/helpers"
	"github.com/controlplaneio/netassert/v2/internal/data"
	"github.com/controlplaneio/netassert/v2/internal/engine"
	"github.com/controlplaneio/netassert/v2/internal/kubeops"
	"github.com/controlplaneio/netassert/v2/internal/logger"
)

const (
	suffixLength            = 9 // suffix length of the random string to be appended to the container name
	snifferContainerImage   = "docker.io/controlplane/netassertv2-packet-sniffer:latest"
	snifferContainerPrefix  = "netassertv2-sniffer"
	scannerContainerImage   = "docker.io/controlplane/netassertv2-l4-client:latest"
	scannerContainerPrefix  = "netassertv2-client"
	pauseInSeconds          = 5 // time to pause before each test case
	packetCaputureInterface = `eth0`
	testCasesFile           = `./manifests/test-cases.yaml`
	resultFile              = "result.log" // where we write the results
)

var (
	envVarKind          = `KIND_E2E_TESTS`
	envVarGKEWithVPC    = `GKE_VPC_E2E_TESTS`
	envVarGKEWithDPv2   = `GKE_DPV2_E2E_TESTS`
	envVarEKSWithVPC    = `EKS_VPC_E2E_TESTS`
	envVarEKSWithCalico = `EKS_CALICO_E2E_TESTS`
)

type MinimalK8sObject struct {
	Kind     string `json:"kind"`
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
}

var denyAllPolicyBody = `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
`

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestKind(t *testing.T) {
	t.Parallel()

	if os.Getenv(envVarKind) == "" {
		t.Skipf("skipping test associated with Kind as %q environment variable was not set", envVarKind)
	}

	kind := helpers.NewKindCluster(t, "./clusters/kind", "kind-calico", helpers.Calico)
	createTestDestroy(t, kind)
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

func waitUntilManifestReady(t *testing.T, svc *kubeops.Service, manifestPath string) []string {
	timeout := 20 * time.Minute
	pollTime := 30 * time.Second

	data, err := os.ReadFile(manifestPath)
	require.NoError(t, err, "Failed to read manifest file")

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	namespaces := make([]string, 0)

	for {
		var obj MinimalK8sObject
		err := decoder.Decode(&obj)

		if err == io.EOF {
			break
		}
		require.NoError(t, err, "Failed to parse YAML document")

		if obj.Kind == "" || obj.Metadata.Name == "" {
			t.Fatalf("Found malformed kubernetes  document in YAML")
			continue
		}

		kind := strings.ToLower(obj.Kind)

		switch kind {
		case "deployment", "daemonset", "replicaset", "statefulset", "pod":
		default:
			continue
		}

		targetNs := strings.ToLower(obj.Metadata.Namespace)
		if targetNs == "" {
			targetNs = "default"
		}

		if !slices.Contains(namespaces, targetNs) {
			namespaces = append(namespaces, targetNs)
		}

		if err := svc.WaitForPodInResourceReady(obj.Metadata.Name, targetNs, kind, pollTime, timeout); err != nil {
			t.Fatalf("Error while waiting for resource %s to become ready: %s", obj.Metadata.Name, err.Error())
		}
	}
	return namespaces
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

	// we apply all the manifests and then run
	k8s.KubectlApply(t, options, "./manifests/workload.yaml")
	namespaces := waitUntilManifestReady(t, svc, "./manifests/workload.yaml")

	netAssertTestCases, err := data.ReadTestsFromFile(testCasesFile)
	if err != nil {
		t.Fatal(err)
	}

	// create the network policies
	k8s.KubectlApply(t, options, "./manifests/networkpolicies.yaml")

	// run the sample tests
	runTests(ctx, t, svc, netAssertTestCases)

	if gc.SkipNetPolTests() {
		return
	}

	// read the tests again for a fresh start
	netAssertTestCases, err = data.ReadTestsFromFile(testCasesFile)
	if err != nil {
		t.Fatal(err)
	}

	// set the exit to 1 since this time the network policies will block the traffic
	for _, tc := range netAssertTestCases {
		tc.ExitCode = 1
	}

	for _, ns := range namespaces {
		nsKubeOptions := k8s.NewKubectlOptions("", kubeConfig, ns)

		k8s.KubectlApplyFromString(t, nsKubeOptions, denyAllPolicyBody)

		k8s.WaitUntilNetworkPolicyAvailable(t, nsKubeOptions, "default-deny-all", 10, 5*time.Second)
		require.NoError(t, err, "Error, the NetworkPolicy should exist in namespace %s", ns)
	}

	// run the tests with network policies blocking everything
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
