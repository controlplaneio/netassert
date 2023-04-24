//go:generate mockgen -destination=engine_mocks_test.go -package=engine github.com/controlplaneio/netassert/internal/engine NetAssertTestRunner

package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/controlplaneio/netassert/v2/internal/data"
	"github.com/hashicorp/go-hclog"
	corev1 "k8s.io/api/core/v1"
)

// Engine - type responsible for running the netAssert test(s)
type Engine struct {
	Service NetAssertTestRunner
	Log     hclog.Logger
}

// New - Returns a new instance of Engine
func New(service NetAssertTestRunner, log hclog.Logger) *Engine {
	return &Engine{Service: service, Log: log}
}

// GetPod - returns a running Pod defined by the K8sResource
func (e *Engine) GetPod(ctx context.Context, res *data.K8sResource) (*corev1.Pod, error) {
	if res == nil {
		return &corev1.Pod{}, fmt.Errorf("res parameter is nil")
	}

	switch res.Kind {
	case data.KindDeployment:
		return e.Service.GetPodInDeployment(ctx, res.Name, res.Namespace)
	case data.KindStatefulSet:
		return e.Service.GetPodInStatefulSet(ctx, res.Name, res.Namespace)
	case data.KindDaemonSet:
		return e.Service.GetPodInDaemonSet(ctx, res.Name, res.Namespace)
	case data.KindPod:
		return e.Service.GetPod(ctx, res.Name, res.Namespace)
	default:
		e.Log.Error("", hclog.Fmt("%s is not supported K8sResource", res.Kind))
		return &corev1.Pod{}, fmt.Errorf("%s is not supported K8sResource", res.Kind)
	}
}

// RunTests - runs a list of net assert test cases
func (e *Engine) RunTests(
	ctx context.Context, // context information
	te data.Tests, // the list of tests we are running
	snifferContainerPrefix string, // name of the sniffer container to use
	snifferContainerImage string, // image location of the sniffer Container
	scannerContainerPrefix string, // name of the scanner container to use
	scannerContainerImage string, // image location of the scanner container
	suffixLength int, // length of the random string that will be generated  and appended to the container name
	pause time.Duration, // time to pause before running a test
	packetCaptureInterface string, // the network interface used to capture traffic by the sniffer container
) {
	var wg sync.WaitGroup

	for _, tc := range te {

		wg.Add(1)

		go func(tc *data.Test, wg *sync.WaitGroup) {
			// time to sleep before running each test
			defer wg.Done()

			err := e.RunTest(ctx, tc, snifferContainerPrefix, snifferContainerImage,
				scannerContainerPrefix, scannerContainerImage, suffixLength, packetCaptureInterface)
			if err != nil {
				e.Log.Error("Test execution failed", "Name", tc.Name, "error", err)
				tc.FailureReason = err.Error()
			}
		}(tc, &wg)

		time.Sleep(pause)
	}
	wg.Wait()
}

// RunTest - Runs a single netAssert test case
func (e *Engine) RunTest(
	ctx context.Context, // context passed to this function
	te *data.Test, // test cases to execute
	snifferContainerPrefix string, // name of the sniffer container to use
	snifferContainerImage string, // image location of the sniffer Container
	scannerContainerPrefix string, // name of the scanner container to use
	scannerContainerImage string, // image location of the scanner container
	suffixLength int, // length of string that will be generated and appended to the container name
	packetCaptureInterface string, // the network interface used to capture traffic by the sniffer container
) error {
	if te.Type != data.K8sTest {
		return fmt.Errorf("only k8s test type is supported at this time: %s", te.Type)
	}

	switch te.Protocol {
	case data.ProtocolTCP:
		return e.RunTCPTest(ctx, te, scannerContainerPrefix, scannerContainerImage, suffixLength)
	case data.ProtocolUDP:
		return e.RunUDPTest(ctx,
			te,
			snifferContainerPrefix,
			snifferContainerImage,
			scannerContainerPrefix,
			scannerContainerImage,
			suffixLength,
			packetCaptureInterface,
		)
	default:
		e.Log.Error("error", hclog.Fmt("Only TCP/UDP protocol is supported at this time and not %s", te.Protocol))
		return fmt.Errorf("only TCP/UDP protocol is supported at this time and not %v", te.Protocol)
	}
}
