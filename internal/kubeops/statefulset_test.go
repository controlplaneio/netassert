package kubeops

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/pointer"
)

// createStatefulSet - creates a statefulset and returns the same
func createStatefulSet(client kubernetes.Interface, name, namespace string, replicas int32, t *testing.T) *appsv1.StatefulSet {
	labels := map[string]string{
		"app": "nginx",
	}
	selector := &metav1.LabelSelector{
		MatchLabels: labels,
	}

	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx",
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
	}

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    pointer.Int32(replicas),
			ServiceName: "nginx-service",
			Selector:    selector,
			Template:    podTemplate,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	}

	obj, err := client.AppsV1().StatefulSets(namespace).Create(context.Background(), statefulSet, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create stateful set %s: %v", name, err)
	}

	return obj
}

// createStatefulSetPod - creates a new pod in the statefulset with index set to index
func createStatefulSetPod(
	client kubernetes.Interface,
	statefulSet *appsv1.StatefulSet,
	ipAddress string,
	index int,
	t *testing.T,
) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", statefulSet.Name, index),
			Namespace: statefulSet.Namespace,
			Labels:    statefulSet.Spec.Template.ObjectMeta.Labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(statefulSet, appsv1.SchemeGroupVersion.WithKind("StatefulSet")),
			},
		},
		Spec: statefulSet.Spec.Template.Spec,
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: ipAddress,
		},
	}

	podObj, err := client.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create pod owned by statefulset %v: %v", statefulSet.Name, err)
	}

	return podObj
}

func TestGetPodInStatefulSet(t *testing.T) {
	testCases := []struct {
		name              string
		ipAddress         string
		ssName            string
		ssNamespace       string
		replicas          int32
		createPod         bool
		createStatefulset bool
		wantErr           bool
		errMsg            string
	}{
		{
			name:              "when both StatefulSet and Pod does not exist",
			ipAddress:         "192.168.0.1",
			ssName:            "web",
			ssNamespace:       "default",
			createPod:         false,
			replicas:          0,
			createStatefulset: false,
			wantErr:           true,
			errMsg:            `unable to find statefulset`,
		},
		{
			name:              "when StatefulSet exists but the Pod does not exist",
			ipAddress:         "192.168.0.1",
			ssName:            "web",
			ssNamespace:       "default",
			createPod:         false,
			replicas:          0,
			createStatefulset: true,
			wantErr:           true,
			errMsg:            `zero replicas were found in the statefulset`,
		},
		{
			name:              "when both StatefulSet and Pod exist",
			ipAddress:         "192.168.0.1",
			ssName:            "web",
			ssNamespace:       "default",
			createPod:         true,
			replicas:          1,
			createStatefulset: true,
			wantErr:           false,
		},
	}

	for _, testCase := range testCases {
		tc := testCase

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)
			client := fake.NewSimpleClientset()

			svc := New(client, hclog.NewNullLogger())

			var (
				ss  *appsv1.StatefulSet
				pod *corev1.Pod
			)

			if tc.createStatefulset {
				ss = createStatefulSet(client, tc.ssName, tc.ssNamespace, tc.replicas, t)
			}

			if tc.createPod {
				pod = createStatefulSetPod(client, ss, tc.ipAddress, 0, t)
			}

			gotPod, err := svc.GetPodInStatefulSet(context.Background(), tc.ssName, tc.ssNamespace)

			// if we do not expect and error in the test case
			if !tc.wantErr {
				r.NoError(err)
				// if we have created both pods and statefulset
				if tc.createPod && tc.createStatefulset {
					r.Equal(gotPod, pod)
				}
				return
			}

			// if tc.wantErr {
			// we are expecting an error
			r.Error(err)
			// check error message with the one defined in the test case
			if tc.errMsg != "" {
				r.Contains(err.Error(), tc.errMsg)
			}
			//}
		})
	}
}
