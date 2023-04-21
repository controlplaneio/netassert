package kubeops

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

const (
	ephContainerGroup   = ""                         // api group which provides the ephemeral container resource
	ephContainerVersion = "v1"                       // api version which provides the ephemeral container resource
	ephContainerKind    = "Pod"                      // core API Kind that provides the ephemeral container resource
	ephContainerRes     = "pods/ephemeralcontainers" // name of the ephemeral container subresource
)

// GetPod - Returns a Running Pod that has an IP address allocated to it
func (svc *Service) GetPod(ctx context.Context, name, namespace string) (*corev1.Pod, error) {
	// Get the Pod that matches the name and namespace
	pod, err := svc.Client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})

	if err != nil && apierrors.IsNotFound(err) {
		svc.Log.Error("unable to find Pod", "namespace", namespace, "error", err)
		return nil, fmt.Errorf("unable to find Pod %s in namespace %s: %w", name, namespace, err)
	}

	if err != nil {
		svc.Log.Error("Unable to find Pod", "namespace", namespace, "error", err)
		return nil, err
	}

	if pod.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("pod %s in namespace %s is not in running state: %s",
			name,
			namespace,
			pod.Status.Phase,
		)
	}

	if pod.Status.PodIP == "" {
		return nil, fmt.Errorf("pod %s in namespace %s does not have an IP address",
			name,
			namespace,
		)
	}

	return pod, nil
}

// getPodFromPodList - returns a random pod from PodList object
func getRandomPodFromPodList(ownerObj metav1.Object, podList *corev1.PodList) (*corev1.Pod, error) {
	if ownerObj == nil {
		return &corev1.Pod{}, fmt.Errorf("parameter ownerObj cannot be nil")
	}

	if podList == nil {
		return &corev1.Pod{}, fmt.Errorf("parameter podList cannot be nil")
	}

	var pods []*corev1.Pod

	for _, pod := range podList.Items {

		tmpPod := pod
		if !metav1.IsControlledBy(&tmpPod, ownerObj) {
			continue
		}

		if tmpPod.Status.Phase != corev1.PodRunning {
			continue
		}

		if tmpPod.Status.PodIP == "" {
			continue
		}

		// this pod is owned by the parent object
		// the pod is in running state
		// the pod also has IP address allocated by CNI
		pods = append(pods, &tmpPod)
	}

	if len(pods) == 0 {
		return &corev1.Pod{}, fmt.Errorf("unable to find a Pod")
	}

	index := rand.Intn(len(pods))
	return pods[index], nil
}

// CheckEphemeralContainerSupport - Checks support for ephemeral containers
func (svc *Service) CheckEphemeralContainerSupport(ctx context.Context) error {
	// check if the kubernetes server has support for ephemeral containers

	found, err := checkResourceSupport(ctx, svc.Client, ephContainerGroup, ephContainerVersion,
		ephContainerKind, ephContainerRes)
	if err != nil {
		return err
	}

	// we have not found the resource
	if !found {
		return fmt.Errorf("unable to find K8s resource=%q, Group=%q, Version=%q Kind=%q",
			ephContainerRes, ephContainerGroup, ephContainerVersion, ephContainerKind)
	}

	return nil
}

// checkResourceSupport - Checks support for a specific resource
func checkResourceSupport(
	ctx context.Context,
	k8sClient kubernetes.Interface,
	group, version, kind, resourceName string,
) (bool, error) {
	groupVersion := schema.GroupVersion{Group: group, Version: version}

	dClient := k8sClient.Discovery()

	resourceList, err := dClient.ServerResourcesForGroupVersion(groupVersion.String())
	if err != nil {
		return false, fmt.Errorf("failed to get API resource list for %q: %w", groupVersion.String(), err)
	}

	for _, resource := range resourceList.APIResources {
		if resource.Kind == kind && resource.Name == resourceName {
			return true, nil
		}
	}

	return false, nil
}

// PingHealthEndpoint - pings a single endpoint of the apiServer using HTTP
func (svc *Service) PingHealthEndpoint(ctx context.Context, endpoint string) error {
	pingRequest := svc.Client.CoreV1().RESTClient().Get().AbsPath(endpoint)

	if err := pingRequest.Do(ctx).Error(); err != nil {
		return fmt.Errorf("unable to HTTP ping "+endpoint+" of the API server: %w", err)
	}

	return nil
}

func (svc *Service) WaitForPodInResourceReady(name, namespace, resourceType string,
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
				log.Printf("Found name=%s namespace=%s  resourceType=%s", name, namespace, resourceType)
				return nil
			}
			log.Println("polling for object", name, namespace, resourceType, err)
			svc.Log.Info("polling", "name", name, "namespace", namespace)
		}
	}
}
