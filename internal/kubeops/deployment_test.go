package kubeops

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/utils/pointer"

	"k8s.io/client-go/kubernetes/fake"
)

func getDeploymentObject(name, namespace string, replicaSize int32) *appsv1.Deployment {
	deploymentName := name
	imageName := "nginx"
	imageTag := "latest"
	replicas := pointer.Int32(replicaSize)

	// Create the Deployment object
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			UID:       uuid.NewUUID(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  imageName,
							Image: fmt.Sprintf("%s:%s", imageName, imageTag),
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment
}

// replicaSetWithOwnerSetToDeployment - creates a replicaset spec with owner reference to deploy
func replicaSetWithOwnerSetToDeployment(deploy *appsv1.Deployment, size int32) *appsv1.ReplicaSet {
	replicaSet := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "replicaset-" + deploy.Name,
			Namespace: "default",
			Labels:    deploy.Spec.Selector.MatchLabels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(deploy, appsv1.SchemeGroupVersion.WithKind("Deployment")),
			},
		},
		Spec: appsv1.ReplicaSetSpec{
			Selector: deploy.Spec.Selector,
			Template: deploy.Spec.Template,
			Replicas: pointer.Int32(size),
		},
	}

	replicaSet.UID = deploy.UID

	return replicaSet
}

// podWithOwnerSetToReplicaSet - creates a Pod spec. and sets the ownder reference to rs
func podWithOwnerSetToReplicaSet(rs *appsv1.ReplicaSet) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rs.Name + "-pod-foo",
			Namespace: rs.Namespace,
			Labels:    rs.Labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(rs, appsv1.SchemeGroupVersion.WithKind("ReplicaSet")),
			},
		},
		Spec: rs.Spec.Template.Spec,
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: "192.168.0.1",
		},
	}

	return pod
}

func TestGetPodInDeployment(t *testing.T) {
	r := require.New(t)
	// build the new service object
	svc := New(fake.NewSimpleClientset(), hclog.NewNullLogger())

	name := "deploy1"
	namespace := "default"
	deploySpec := getDeploymentObject(name, namespace, 1)
	ctx := context.Background()
	deployObj, err := svc.Client.AppsV1().Deployments(namespace).Create(ctx, deploySpec, metav1.CreateOptions{})
	r.NoError(err)

	// deployment will not create a replicaSet so when we call GetPodInDeployment, it should
	// give us an error specific to that
	_, err = svc.GetPodInDeployment(ctx, name, namespace)
	r.Error(err)
	rsNotFoundMessage := fmt.Sprintf("could not find the replicaSet with replica count >=1 owned "+
		"by the deployment %s in namespace %s",
		name, namespace)
	r.Equal(rsNotFoundMessage, err.Error())

	// now we create the replicaSet that is owned by the deployment that has zero replicas
	rsSpec := replicaSetWithOwnerSetToDeployment(deployObj, 0)
	_, err = svc.Client.AppsV1().ReplicaSets(namespace).Create(ctx, rsSpec, metav1.CreateOptions{})
	r.NoError(err, "failed to create replicaSet with size set to zero")
	_, err = svc.GetPodInDeployment(ctx, name, namespace)
	r.Error(err)
	// rsSizeSetToZeroMsg := fmt.Sprintf("deployment %s in namespace %s has repliaset size set to zero",
	//	name, namespace)
	r.Equal(rsNotFoundMessage, err.Error())

	// we now modify the existing replicaSet and set the replica size to 1
	rsSpec.Spec.Replicas = pointer.Int32(1)
	rsObj, err := svc.Client.AppsV1().ReplicaSets(namespace).Update(ctx, rsSpec, metav1.UpdateOptions{})
	r.NoError(err, "failed to update replicaset size set 1")
	_, err = svc.GetPodInDeployment(ctx, name, namespace)
	podNotFound := fmt.Sprintf("unable to find any Pod associated with deployment %s in namespace %s",
		name, namespace)
	r.Contains(err.Error(), podNotFound)

	// we now create Pod that has no IP address set to it
	// but is in running stage and is owned by the replicaSet
	// this should also trigger an error
	podSpec := podWithOwnerSetToReplicaSet(rsObj)
	podSpec.Status.PodIP = ""
	_, err = svc.Client.CoreV1().Pods(namespace).Create(ctx, podSpec, metav1.CreateOptions{})
	r.NoError(err)
	_, err = svc.GetPodInDeployment(ctx, name, namespace)
	r.Contains(err.Error(), podNotFound)

	// finally we update the Pod to simulate an IP address allocation by CNI
	podSpec.Status.PodIP = "192.168.0.1"
	createdPodObj, err := svc.Client.CoreV1().Pods(namespace).Update(ctx, podSpec, metav1.UpdateOptions{})
	r.NoError(err)
	gotPodObj, err := svc.GetPodInDeployment(ctx, name, namespace)
	r.NoError(err)
	r.Equal(createdPodObj, gotPodObj)
}
