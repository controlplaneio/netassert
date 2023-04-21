package kubeops

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

// GetPodInDeployment - Returns a running Pod in a deployment
func (svc *Service) GetPodInDeployment(ctx context.Context, name, namespace string) (*corev1.Pod, error) {
	// check if the deployment actually exists
	deploy, err := svc.Client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// get the list of replicaset in the current namespace
	replicaSets, err := svc.Client.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var rsList []*appsv1.ReplicaSet

	var rs *appsv1.ReplicaSet

	// Loop through the ReplicaSets and check if they are owned by the deployment
	for index := range replicaSets.Items {
		if metav1.IsControlledBy(&replicaSets.Items[index], deploy) {
			r := replicaSets.Items[index]
			if *r.Spec.Replicas == 0 {
				svc.Log.Info("replicaSet size set to zero",
					"name", r.Name,
					"deployment", name,
					"namespace", namespace)
				continue
			}
			rsList = append(rsList, &r)
		}
	}

	if len(rsList) == 0 {
		return nil, fmt.Errorf("could not find the replicaSet with replica count >=1 owned by the "+
			"deployment %s in namespace %s", name, namespace)
	}

	// we use the first replicaset that has zero size
	rs = rsList[0]

	if rs == nil || rs.Spec.Replicas == nil {
		return nil, fmt.Errorf("could not find the replicaSet owned by the deployment %s in namespace %s",
			name, namespace)
	}

	if *rs.Spec.Replicas < 1 {
		svc.Log.Info("replicaSet size set to zero",
			"deployment", name,
			"namespace", namespace)

		return nil, fmt.Errorf("deployment %s in namespace %s has repliaset size set to zero",
			name, namespace)
	}

	// we only want the Pods that are in running state
	fieldSelector := fields.OneTermEqualSelector(
		"status.phase",
		"Running",
	)

	// grab the labels associated with the ReplicaSet object
	// we only want to select the Pods, whose labels match that of the parent replicaSet
	// selector, err := labels.ValidatedSelectorFromSet(rs.Labels)

	selector, err := labels.ValidatedSelectorFromSet(rs.Spec.Selector.MatchLabels)
	if err != nil {
		return nil, err
	}

	// we now get a list of pods and find the ones owned by the replicaset
	pods, err := svc.Client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}

	pod, err := getRandomPodFromPodList(rs, pods)
	if err != nil {
		return nil, fmt.Errorf("unable to find any Pod associated with deployment %s in namespace %s: %w",
			name, namespace, err)
	}

	svc.Log.Info("Found Pod", "Parent-deployment", name, "Pod", pod.Name,
		"Namespace", pod.Namespace, "IP", pod.Status.PodIP)

	return pod, nil
}
