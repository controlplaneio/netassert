//go:generate mockgen -destination=engine_mocks_test.go -package=engine github.com/controlplaneio/netassert/internal/engine NetAssertTestRunner

package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	corev1 "k8s.io/api/core/v1"

	"github.com/controlplaneio/netassert/v2/internal/data"
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

	for i, tc := range te {
		wg.Add(1)
		go func(tc *data.Test, wg *sync.WaitGroup) {
			defer wg.Done()
			// run the test case
			err := e.RunTest(ctx, tc, snifferContainerPrefix, snifferContainerImage,
				scannerContainerPrefix, scannerContainerImage, suffixLength, packetCaptureInterface)
			if err != nil {
				e.Log.Error("Test execution failed", "Name", tc.Name, "error", err)
				tc.FailureReason = err.Error()
			}
		}(tc, &wg)

		if i < len(te)-1 { // do not pause after the last test
			cancellableDelay(ctx, pause)
		}
		// If the context is cancelled, we need to break out of the loop
		if ctx.Err() != nil {
			break
		}
	}
	wg.Wait()
}

// cancellableDelay introduces a delay that can be interrupted by context cancellation.
// This function is useful when you want to pause execution for a specific duration,
// but also need the ability to respond quickly if an interrupt signal (like CTRL + C) is received.
func cancellableDelay(ctx context.Context, duration time.Duration) {
	select {
	case <-time.After(duration):
		// The case of time.After(duration) is selected when the specified duration has elapsed.
		// This means the function completes its normal delay without any interruption.

	case <-ctx.Done():
		// The ctx.Done() case is selected if the context is cancelled before the duration elapses.
		// This could happen if an interrupt signal is received.
		// Returning early from the function allows the program to quickly respond to the cancellation signal,
		// such as cleaning up resources, stopping further processing, etc.

		// No specific action is needed here other than returning from the function,
		// as the cancellation of the context is handled by the caller.
		return
	}
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
