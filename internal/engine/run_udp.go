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

const (
	defaultNetInt  = `eth0` // default network interface
	defaultSnapLen = 1024   // default size of the packet snap length
)

// RunUDPTest - runs a UDP test
func (e *Engine) RunUDPTest(
	ctx context.Context, // context information
	te *data.Test, // the test case to run
	snifferContainerSuffix string, // name of the sniffer container to use
	snifferContainerImage string, // image location of the sniffer Container
	scannerContainerSuffix string, // name of the scanner container to use
	scannerContainerImage string, // image location of the scanner container
	suffixLength int, // length of string that will be generated and appended to the container name
	networkInterface string, // name of the network interface that will be used for packet capturing
) error {
	if te == nil {
		return fmt.Errorf("test case is nil object")
	}

	if snifferContainerSuffix == "" {
		return fmt.Errorf("snifferContainerSuffix parameter cannot be empty string")
	}

	if snifferContainerImage == "" {
		return fmt.Errorf("snifferContainerImage parameter cannot be empty string")
	}

	if scannerContainerSuffix == "" {
		return fmt.Errorf("scannerContainerSuffix parameter cannot be empty string")
	}

	if scannerContainerImage == "" {
		return fmt.Errorf("scannerContainerImage parameter cannot be empty string")
	}

	// we only run UDP tests here
	if te.Protocol != data.ProtocolUDP {
		return fmt.Errorf("test case protocol is set to %q, this function only supports %q",
			te.Protocol, data.ProtocolUDP)
	}

	// te is already validate as validation is done at the Unmarshalling of the resource
	// validation ensures that for the time being the src holds a type of k8sResource
	// check if the te is not nil
	// check the value of te.Type and ensure that its k8s
	// check if the te.Protocol is tcp or udp
	// if the protocol is tcp
	// find a running Pod in the src resource

	var (
		srcPod, dstPod *corev1.Pod
		targetHost     string
		err            error
	)

	// as we cannot inject ephemeral container when the Dst type is Host, we will
	// return an error
	if te.Dst.Host != nil {
		return fmt.Errorf("%q: dst should not contain host object when protocol is %s", te.Name, te.Protocol)
	}

	if te.Dst.K8sResource == nil {
		return fmt.Errorf("%q: dst should contain non-nil k8sResource object", te.Name)
	}

	e.Log.Info("🟢 Running UDP test", "Name", te.Name)

	// name of the network interface that will be used for packet capturing
	// if none is set then we used the default one i.e. eth0
	if networkInterface == "" {
		networkInterface = defaultNetInt
	}

	// find a running Pod represented by the  src.K8sResource object
	srcPod, err = e.GetPod(ctx, te.Src.K8sResource)
	if err != nil {
		return fmt.Errorf("unable to get source pod for test %s: %w", te.Name, err)
	}

	// find a running Pod in the destination kubernetes object
	dstPod, err = e.GetPod(ctx, te.Dst.K8sResource)
	if err != nil {
		return err
	}

	targetHost = dstPod.Status.PodIP

	msg, err := kubeops.NewUUIDString()
	if err != nil {
		return fmt.Errorf("unable to genereate random UUID for test %s: %w", te.Name, err)
	}

	// we now have both the source and the destination object, we need to ensure that we first
	// inject the sniffer into the Destination Pod
	snifferEphemeralContainer, err := e.Service.BuildEphemeralSnifferContainer(
		snifferContainerSuffix+"-"+kubeops.RandString(suffixLength),
		snifferContainerImage,
		msg,
		defaultSnapLen,
		string(te.Protocol),
		te.Attempts,
		networkInterface,
		te.TimeoutSeconds+5, // add 5 seconds for the Container to come online
	)
	if err != nil {
		return fmt.Errorf("failed to build sniffer ephemeral container for test %s: %w", te.Name, err)
	}

	// we now build the scanner container
	scannerEphemeralContainer, err := e.Service.BuildEphemeralScannerContainer(
		scannerContainerSuffix+"-"+kubeops.RandString(suffixLength),
		scannerContainerImage,
		targetHost,
		strconv.Itoa(te.TargetPort),
		string(te.Protocol),
		msg,
		te.Attempts*3, // increase the attempts to ensure that we send three times the packets
	)
	if err != nil {
		return fmt.Errorf("unable to build ephemeral scanner container for test %s: %w", te.Name, err)
	}

	// run the ephemeral containers on dst Pod first and then the source Pod
	dstPod, snifferContainerName, err := e.Service.LaunchEphemeralContainerInPod(ctx, dstPod,
		snifferEphemeralContainer)
	if err != nil {
		return fmt.Errorf("sniffer ephermal container launch failed for test %s: %w", te.Name, err)
	}

	// run the ephemeral scanner container in the source Pod after we have
	// launched the sniffer and the sniffer container is ready
	_, scannerContainerName, err := e.Service.LaunchEphemeralContainerInPod(ctx, srcPod, scannerEphemeralContainer)
	if err != nil {
		return fmt.Errorf("scanner ephemeral container launch failed for test %s: %w", te.Name, err)
	}

	// sniffer is successfully injected into the dstPod, now we check the exit code
	exitCodeSnifferCtr, err := e.Service.GetExitStatusOfEphemeralContainer(
		ctx,
		snifferContainerName,
		time.Duration(te.TimeoutSeconds)*time.Second,
		dstPod.Name,
		dstPod.Namespace,
	)
	if err != nil {
		return fmt.Errorf("failed to get exit code of the sniffer ephemeral container %s for test %s: %w",
			snifferContainerName, te.Name, err)
	}

	e.Log.Info("Got exit code from ephemeral sniffer container",
		"testName", te.Name,
		"exitCode", exitCodeSnifferCtr,
		"containerName", snifferContainerName)

	if exitCodeSnifferCtr != te.ExitCode {
		return fmt.Errorf("ephemeral sniffer container %s exit code for test %v is %v instead of %d",
			snifferContainerName, te.Name, exitCodeSnifferCtr, te.ExitCode)
	}

	// get the exit status of the scanner container
	exitCodeScanner, err := e.Service.GetExitStatusOfEphemeralContainer(
		ctx, scannerContainerName,
		time.Duration(te.TimeoutSeconds+10)*time.Second,
		srcPod.Name,
		srcPod.Namespace,
	)
	if err != nil {
		return fmt.Errorf("failed to get exit code of the scanner ephemeral container %s for test %s: %w",
			scannerContainerName, te.Name, err)
	}

	e.Log.Info("Got exit code from ephemeral scanner container",
		"testName", te.Name,
		"exitCode", exitCodeScanner,
		"containerName", scannerContainerName)

	// for UDP scanning the exit code of the scanner is always zero
	// as UDP is connectionless
	if exitCodeScanner != 0 {
		return fmt.Errorf("ephemeral scanner container %s exit code for test %v is %v instead of 0",
			scannerContainerName, te.Name, exitCodeScanner)
	}

	te.Pass = true // mark test as pass
	return nil
}
