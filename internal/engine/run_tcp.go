package engine

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/controlplaneio/netassert/v2/internal/data"
	"github.com/controlplaneio/netassert/v2/internal/kubeops"
	corev1 "k8s.io/api/core/v1"
)

// RunTCPTest - runs a TCP test
func (e *Engine) RunTCPTest(
	ctx context.Context, // context information
	te *data.Test, // test case we want to run
	scannerContainerName string, // name of the scanner container to use
	scannerContainerImage string, // docker image location of the scanner container image
	suffixLength int, // length of random string that will be appended to the ephemeral container name
) error {
	var (
		dstPod     *corev1.Pod
		targetHost string
	)

	if te.Dst.Host != nil && te.Dst.K8sResource != nil {
		return fmt.Errorf("both Dst.Host and Dst.K8sResource cannot be set at the same time")
	}

	if te.Dst.Host == nil && te.Dst.K8sResource == nil {
		return fmt.Errorf("both Dst.Host and Dst.K8sResource are nil")
	}

	if scannerContainerName == "" {
		return fmt.Errorf("scannerContainerName parameter cannot be empty string")
	}

	if scannerContainerImage == "" {
		return fmt.Errorf("scannerContainerImage parameter cannot be empty string")
	}

	e.Log.Info("🟢 Running TCP test", "Name", te.Name)

	srcPod, err := e.GetPod(ctx, te.Src.K8sResource)
	if err != nil {
		return err
	}

	// if Destination K8sResource is not set to nil
	if te.Dst.K8sResource != nil {
		// we need to find a running Pod  with IP Address in the Dst K8sResource
		dstPod, err = e.GetPod(ctx, te.Dst.K8sResource)
		if err != nil {
			return err
		}
		targetHost = dstPod.Status.PodIP
	} else {
		targetHost = te.Dst.Host.Name
	}

	// build ephemeral container with details of the IP addresses
	msg, err := kubeops.NewUUIDString()
	if err != nil {
		return fmt.Errorf("unable to genereate random UUID for test %s: %w", te.Name, err)
	}

	debugContainer, err := e.Service.BuildEphemeralScannerContainer(
		scannerContainerName+"-"+kubeops.RandString(suffixLength),
		scannerContainerImage,
		targetHost,
		strconv.Itoa(te.TargetPort),
		string(te.Protocol),
		msg,
		te.Attempts,
	)
	if err != nil {
		return fmt.Errorf("unable to build ephemeral scanner container for test %s: %w", te.Name, err)
	}

	// run the ephemeral/debug container
	// grab the exit code
	// make sure that the exit code matches the one that is specified in the test
	srcPod, ephContainerName, err := e.Service.LaunchEphemeralContainerInPod(ctx, srcPod, debugContainer)
	if err != nil {
		return fmt.Errorf("ephemeral container launch failed for test %s: %w", te.Name, err)
	}

	err = e.CheckExitStatusOfEphContainer(
		ctx,
		ephContainerName,
		te.Name,
		srcPod.Name,
		srcPod.Namespace,
		time.Duration(te.TimeoutSeconds)*time.Second,
		te.ExitCode,
	)

	if err != nil {
		return err
	}

	te.Pass = true // set the test as pass
	return nil
}

// CheckExitStatusOfEphContainer - returns an error if exit code of the ephemeral container does not match expExitCode
func (e *Engine) CheckExitStatusOfEphContainer(
	ctx context.Context, // context to pass to our function
	ephContainerName string, // name of the ephemeral container
	testCaseName string, // name of the test case
	podName string, // name of the pod that houses the ephemeral container
	podNamespace string, // namespace of the pod that houses the ephemeral container
	timeout time.Duration, // timeout for the exit status to reach the desired exit code
	expExitCode int, // expected exit code from the ephemeral container
) error {
	containerExitCode, err := e.Service.GetExitStatusOfEphemeralContainer(
		ctx,
		ephContainerName,
		timeout,
		podName,
		podNamespace,
	)
	if err != nil {
		return fmt.Errorf("failed to get exit code of the ephemeral container %s for test %s: %w",
			ephContainerName, testCaseName, err)
	}

	e.Log.Info("Got exit code from ephemeral container",
		"testName", testCaseName,
		"exitCode", containerExitCode,
		"container", ephContainerName,
	)

	if containerExitCode != expExitCode {
		e.Log.Error("Got exit code from ephemeral container",
			"testName", testCaseName,
			"exitCode", containerExitCode,
			"expectedExitCode", expExitCode,
			"container", ephContainerName,
		)
		return fmt.Errorf("ephemeral container %s exit code for test %v is %v instead of %v",
			ephContainerName, testCaseName, containerExitCode, expExitCode)
	}

	return nil
}
