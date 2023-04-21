package kubeops

import (
	"context"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func createDaemonSet(client kubernetes.Interface,
	name, namespace string,
	numberAvailable int32,
	t *testing.T,
) *appsv1.DaemonSet {
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       uuid.NewUUID(),
		},
		Status: appsv1.DaemonSetStatus{
			NumberAvailable: numberAvailable,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "foobar",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "foobar",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx-ofcourse",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}

	ctx := context.Background()
	obj, err := client.AppsV1().DaemonSets(namespace).Create(ctx, ds, metav1.CreateOptions{})
	if err != nil {
		t.Fatal("failed to create daemon set", err)
	}

	return obj
}

func deaemonSetPod(
	client kubernetes.Interface,
	ownerDaemonSet *appsv1.DaemonSet,
	podName string,
	podNamespace string,
	t *testing.T,
	podPhase corev1.PodPhase,
	ipAddress string,
) *corev1.Pod {
	if ownerDaemonSet == nil {
		t.Fatal("ownerDaemonSet is set to nil")
	}

	// Create the Pod with the appropriate owner reference
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: podNamespace,
			Labels:    ownerDaemonSet.Spec.Selector.MatchLabels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ownerDaemonSet, appsv1.SchemeGroupVersion.WithKind("DaemonSet")),
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "has-to-be-nginx",
					Image: "nginx:latest",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: podPhase,
			PodIP: ipAddress,
		},
	}

	// Create the Pod in the Kubernetes cluster.
	pod, err := client.CoreV1().Pods(podNamespace).Create(
		context.Background(), pod, metav1.CreateOptions{},
	)
	if err != nil {
		t.Fatalf("unable to create pod: %v", err)
	}

	return pod
}

func TestGetPodInDaemonSet(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name            string
		dsName          string // daemonSet name
		namespace       string // namespace for both the Pod and daemonset
		createPod       bool
		createDS        bool // should the daemonset be created
		ipAddress       string
		podPhase        corev1.PodPhase
		wantErr         bool
		numberAvailable int32
		errMsg          string
	}{
		{
			name:            "when both the DaemonSet and Pod do not exist",
			dsName:          "foo",
			namespace:       "bar",
			createPod:       false,
			createDS:        false,
			ipAddress:       "192.168.168.33",
			podPhase:        corev1.PodRunning,
			wantErr:         true,
			numberAvailable: 1,
			errMsg:          `daemonsets.apps "foo" not found`,
		},
		{
			name:            "when the DaemonSet exists but the Pod does not",
			dsName:          "foo",
			namespace:       "bar",
			createPod:       false,
			createDS:        true,
			ipAddress:       "192.168.168.33",
			podPhase:        corev1.PodRunning,
			wantErr:         true,
			numberAvailable: 1,
			errMsg:          `unable to find any Pod`,
		},
		{
			name:            "when the DaemonSet and Pod exists but the Pod is not in running state",
			dsName:          "foo",
			namespace:       "bar",
			createPod:       true,
			createDS:        true,
			ipAddress:       "192.168.168.33",
			podPhase:        corev1.PodPending,
			wantErr:         true,
			numberAvailable: 1,
			errMsg:          `unable to find any Pod`,
		},
		{
			name:            "when the DaemonSet and Pod exists and the Pod is in running state",
			dsName:          "foo",
			namespace:       "bar",
			createPod:       true,
			createDS:        true,
			ipAddress:       "192.168.168.33",
			podPhase:        corev1.PodRunning,
			wantErr:         false,
			numberAvailable: 1,
			errMsg:          ``,
		},
		{
			name:            "when the DaemonSet and Pod exists, the Pod is in running state but does not have valid IP address",
			dsName:          "foo",
			namespace:       "bar",
			createPod:       true,
			createDS:        true,
			ipAddress:       "",
			podPhase:        corev1.PodRunning,
			wantErr:         true,
			numberAvailable: 1,
			errMsg:          `unable to find any Pod owned by daemonset`,
		},
		{
			name:            "when the DaemonSet exists, but has Status.NumberAvailable set to zero",
			dsName:          "foo",
			namespace:       "bar",
			createPod:       false,
			createDS:        true,
			ipAddress:       "",
			podPhase:        corev1.PodRunning,
			wantErr:         true,
			numberAvailable: 0,
			errMsg:          `zero Pods are available in the daemonset`,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			k8sClient := fake.NewSimpleClientset()

			var ds *appsv1.DaemonSet // daemon set

			if tc.createDS {
				tmpDs := createDaemonSet(k8sClient, tc.dsName, tc.namespace, tc.numberAvailable, t)
				r.NotNil(tmpDs)
				ds = tmpDs
			}

			if tc.createPod {
				_ = deaemonSetPod(k8sClient, ds, "foo-pod", tc.namespace, t, tc.podPhase, tc.ipAddress)
			}

			svc := New(k8sClient, hclog.NewNullLogger())

			gotPod, err := svc.GetPodInDaemonSet(context.Background(), tc.dsName, tc.namespace)

			if !tc.wantErr { // if we do not want an error then
				r.NoError(err)
				if tc.createPod {
					r.Equal(gotPod.Status.PodIP, tc.ipAddress)
				}
				return
			}

			if tc.wantErr {
				r.Error(err)
				if tc.errMsg != "" {
					r.Contains(err.Error(), tc.errMsg)
				}

			}
		})
	}
}
