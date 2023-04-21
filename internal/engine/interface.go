package engine

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// PodGetter - gets a running Pod from various kubernetes resources
type PodGetter interface {
	GetPodInDaemonSet(context.Context, string, string) (*corev1.Pod, error)
	GetPodInDeployment(context.Context, string, string) (*corev1.Pod, error)
	GetPodInStatefulSet(context.Context, string, string) (*corev1.Pod, error)
	GetPod(context.Context, string, string) (*corev1.Pod, error)
}

// EphemeralContainerOperator - various operations related to the ephemeral container(s)
type EphemeralContainerOperator interface {
	BuildEphemeralScannerContainer(
		name string, // name of the ephemeral container
		image string, // image location of the container
		targetHost string, // host to connect to
		targetPort string, // target Port to connect to
		protocol string, // protocol to used for connection
		message string, // message to pass to the remote target
		attempts int, // Number of attempts
	) (*corev1.EphemeralContainer, error)

	GetExitStatusOfEphemeralContainer(
		ctx context.Context, // context passed to the function
		containerName string, // name of the ephemeral container
		timeOut time.Duration, // maximum duration to poll for the ephemeral container status
		podName string, // name of the pod that houses the ephemeral container
		podNamespace string, // namespace of the pod that houses the ephemeral container
	) (int, error)

	BuildEphemeralSnifferContainer(
		name string, // name of the ephemeral container
		image string, // image location of the container
		search string, // search for this string in the captured packet
		snapLen int, // snapLength to capture
		protocol string, // protocol to capture
		numberOfmatches int, // no. of matches
		intFace string, // the network interface to read the packets from
		timeoutSec int, // timeout for the ephemeral container in seconds
	) (*corev1.EphemeralContainer, error)

	LaunchEphemeralContainerInPod(
		ctx context.Context, // the context
		pod *corev1.Pod, // the target Pod
		ec *corev1.EphemeralContainer, // the ephemeralContainer that needs to be injected to the Pod
	) (*corev1.Pod, string, error)
}

// NetAssertTestRunner - runs netassert test case(s)
type NetAssertTestRunner interface {
	EphemeralContainerOperator
	PodGetter
}
