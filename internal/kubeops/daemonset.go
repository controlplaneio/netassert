package kubeops

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

// GetPodInDaemonSet - Returns a running Pod with IP address in a DaemonSet
func (svc *Service) GetPodInDaemonSet(ctx context.Context, name, namespace string) (*corev1.Pod, error) {
	// check if the DaemonSet actually exists
	ds, err := svc.Client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if ds.Status.NumberAvailable == 0 {
		svc.Log.Error("Found zero available Pods", "DaemonSet", name, "Namespace", namespace)
		return nil, fmt.Errorf("zero Pods are available in the daemonset %s", name)
	}

	fieldSelector := fields.OneTermEqualSelector(
		"status.phase",
		"Running",
	).String()

	// Grab the labels from the DaemonSet selector
	podLabels := labels.FormatLabels(ds.Spec.Selector.MatchLabels)

	pods, err := svc.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: podLabels,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		svc.Log.Error("Unable to list Pods", "namespace", namespace, "error", err)
		return nil, fmt.Errorf("unable to list Pods in namespace %s: %w", namespace, err)
	}

	pod, err := getRandomPodFromPodList(ds, pods)
	if err != nil {
		return nil, fmt.Errorf("unable to find any Pod owned by daemonset %s in namespace %s",
			name, namespace)
	}

	return pod, nil
}
