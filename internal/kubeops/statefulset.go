package kubeops

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

// GetPodInStatefulSet - Returns a running Pod in a statefulset
func (svc *Service) GetPodInStatefulSet(ctx context.Context, name, namespace string) (*corev1.Pod, error) {
	// check if the statefulset actually exists
	ss, err := svc.Client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("unable to find statefulset %s in namespace %s: %w", name, namespace, err)
		}
		return nil, err
	}

	// ensure that the statefulset has replia count set to at least 1

	if *ss.Spec.Replicas == 0 {
		svc.Log.Error("Stateful set has zero replicas", "Statefulset",
			name, "Namespace", namespace)
		return nil, fmt.Errorf("zero replicas were found in the statefulset %s in namespace %s",
			name, namespace)
	}

	fieldSelector := fields.OneTermEqualSelector(
		"status.phase",
		"Running",
	).String()

	// Grab the labels from the embedded Pod spec templates
	// podLabels := labels.FormatLabels(ss.Spec.Template.Labels)
	podLabels := labels.FormatLabels(ss.Spec.Selector.MatchLabels)

	pods, err := svc.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: podLabels,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		svc.Log.Error("Unable list Pods", "Namespace", namespace, "error", err)
		return nil, fmt.Errorf("unable to list Pods in namespace: %w", err)
	}

	pod, err := getRandomPodFromPodList(ss, pods)
	if err != nil {
		return nil, fmt.Errorf("unable to find Pod owned by Statefulset %s in namespace %s: %w",
			name, namespace, err)
	}

	svc.Log.Info("Found running Pod owned by StatefulSet", "StatefulSetName", name,
		"Namespace", namespace, "Pod", pod.Name, "PodIP", pod.Status.PodIP)

	return pod, nil
}
