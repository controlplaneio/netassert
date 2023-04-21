package kubeops

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/utils/pointer"
)

// LaunchEphemeralContainerInPod - Launches an ephemeral container in running Pod
func (svc *Service) LaunchEphemeralContainerInPod(
	ctx context.Context, // the context
	pod *corev1.Pod, // the target Pod
	ec *corev1.EphemeralContainer, // the ephemeralContainer that needs to be injected to the Pod
) (*corev1.Pod, string, error) {
	// grab the JSON of the original Pod
	originalPodJSON, err := json.Marshal(pod)
	if err != nil {
		svc.Log.Error("Unable to marshal the original pod into JSON object")
		return nil, "", err
	}

	podCopy := pod.DeepCopy()

	// Add the ephemeral container to the Pod spec of existing ephemeral containers
	podCopy.Spec.EphemeralContainers = append(podCopy.Spec.EphemeralContainers, *ec)

	podCopyJSON, err := json.Marshal(podCopy)
	if err != nil {
		return nil, "", err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(originalPodJSON, podCopyJSON, pod)
	if err != nil {
		return nil, "", err
	}

	svc.Log.Debug("Generated JSON patch for the Pod", "JSONPatch", string(patch))
	svc.Log.Info("Patching Pod", "Pod", pod.Name, "Namespace", pod.Namespace, "Container", ec.Name)

	// we now patch the Pod with ephemeral container
	newPod, err := svc.Client.CoreV1().Pods(pod.Namespace).Patch(
		ctx,
		pod.Name,
		types.StrategicMergePatchType,
		patch,
		metav1.PatchOptions{},
		"ephemeralcontainers",
	)
	if err != nil {
		return nil, "", err
	}

	return newPod, ec.Name, nil
}

// BuildEphemeralSnifferContainer - builds an ephemeral sniffer container
func (svc *Service) BuildEphemeralSnifferContainer(
	name string, // name of the ephemeral container
	image string, // image location of the container
	search string, // search for this string in the captured packet
	snapLen int, // snapLength to capture
	protocol string, // protocol to capture
	numberMatches int, // no. of matches that triggers an exit with status 0
	intFace string, // the network interface to read the packets from
	timeoutSec int, // timeout for the ephemeral container
) (*corev1.EphemeralContainer, error) {
	ec := corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:  name,
			Image: image,
			Env: []corev1.EnvVar{
				{
					Name:  "TIMEOUT_SECONDS",
					Value: strconv.Itoa(timeoutSec),
				},
				{
					Name:  "IFACE",
					Value: intFace,
				},
				{
					Name:  "SNAPLEN",
					Value: strconv.Itoa(snapLen),
				},
				{
					Name:  "SEARCH_STRING",
					Value: search,
				},
				{
					Name:  "PROTOCOL",
					Value: protocol,
				},
				{
					Name:  "MATCHES",
					Value: strconv.Itoa(numberMatches),
				},
			},
			Stdin:     false,
			StdinOnce: false,
			TTY:       false,
			SecurityContext: &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{"NET_RAW"},
				},
				AllowPrivilegeEscalation: pointer.Bool(false),
				RunAsNonRoot:             pointer.Bool(true),
			},
		},
		// empty string forces the container to run in the namespace of the Pod, rather than the container
		// this is default value, only added here for readability
		TargetContainerName: "",
	}

	return &ec, nil
}

// BuildEphemeralScannerContainer - builds an ephemeral scanner container
func (svc *Service) BuildEphemeralScannerContainer(
	name string, // name of the ephemeral container
	image string, // image location of the container
	targetHost string, // host to connect to
	targetPort string, // target Port to connect to
	protocol string, // protocol to used for connection
	message string, // message to pass to the remote target
	attempts int, // Number of attempts
) (*corev1.EphemeralContainer, error) {
	ec := corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:  name,
			Image: image,
			Env: []corev1.EnvVar{
				{
					Name:  "TARGET_HOST",
					Value: targetHost,
				},
				{
					Name:  "TARGET_PORT",
					Value: targetPort,
				},
				{
					Name:  "PROTOCOL",
					Value: protocol,
				},
				{
					Name:  "MESSAGE",
					Value: message,
				},
				{
					Name:  "ATTEMPTS",
					Value: strconv.Itoa(attempts),
				},
			},
			Stdin:     false,
			StdinOnce: false,
			TTY:       false,
			SecurityContext: &corev1.SecurityContext{
				RunAsNonRoot:             pointer.Bool(true),
				AllowPrivilegeEscalation: pointer.Bool(false),
			},
		},
		// empty string forces the container to run in the namespace of the Pod, rather than the container
		// this is default value, only added here for readability
		TargetContainerName: "",
	}

	return &ec, nil
}

// GetExitStatusOfEphemeralContainer - returns the exit status of an EphemeralContainer in a pod
func (svc *Service) GetExitStatusOfEphemeralContainer(
	ctx context.Context, // the context
	containerName string, // name of the ephemeral container
	timeOut time.Duration, // maximum duration to poll for the container status
	podName string, // name of the pod which has the ephemeral container
	podNamespace string, // namespace of the pod which has the ephemeral container
) (int, error) {
	// we only want the Pods that are in running state
	// and are in specific namespace
	fieldSelector := fields.AndSelectors(
		fields.OneTermEqualSelector(
			"status.phase",
			"Running",
		),
		fields.OneTermEqualSelector(
			"metadata.name",
			podName,
		),
		fields.OneTermEqualSelector(
			"metadata.namespace",
			podNamespace,
		),
	)

	podWatcher, err := svc.Client.CoreV1().Pods(podNamespace).Watch(ctx, metav1.ListOptions{
		TypeMeta:      metav1.TypeMeta{},
		FieldSelector: fieldSelector.String(),
	})

	defer func() {
		if podWatcher != nil {
			podWatcher.Stop()
		}
	}()

	if err != nil {
		return -1, err
	}

	timer := time.NewTimer(timeOut)
	defer func() {
		if ok := timer.Stop(); !ok {
			svc.Log.Info("Unable to close the timer channel")
		}
	}()

	for {
		select {
		case event := <-podWatcher.ResultChan():

			pod, ok := event.Object.(*corev1.Pod)
			// if this is not a pod object, then we skip
			if !ok {
				break // breaks from the select
			}

			svc.Log.Debug("Polling the status of ephemeral container",
				"pod", pod.Name,
				"namespace", pod.Namespace,
				"container", containerName,
			)

			for _, v := range pod.Status.EphemeralContainerStatuses {

				if v.Name != containerName {
					continue
				}

				if v.State.Waiting != nil {
					svc.Log.Debug("Container state", "container", containerName, "state", "Waiting")
					continue
				}

				if v.State.Running != nil {
					svc.Log.Debug("Container state", "container", containerName, "state", "Running")
					continue
				}

				if v.State.Terminated != nil {
					svc.Log.Info("Ephemeral container has finished executing", "name", containerName)
					svc.Log.Debug("", "ContainerName", v.Name)
					svc.Log.Debug("", "ExitCode", v.State.Terminated.ExitCode)
					svc.Log.Debug("", "ContainerID", v.State.Terminated.ContainerID)
					svc.Log.Debug("", "FinishedAt", v.State.Terminated.FinishedAt)
					svc.Log.Debug("", "StartedAt", v.State.Terminated.StartedAt)
					svc.Log.Debug("", "Message", v.State.Terminated.Message)
					svc.Log.Debug("", "Reason", v.State.Terminated.Reason)
					svc.Log.Debug("", "Signal", v.State.Terminated.Signal)
					return int(v.State.Terminated.ExitCode), nil
				}

			}

		case <-timer.C:
			return -1, fmt.Errorf("container %v did not reach termination state in %v seconds", containerName, timeOut.Seconds())
		case <-ctx.Done():
			return -1, fmt.Errorf("process was cancelled: %w", ctx.Err())
		}
	}
}
