package kubeops

import (
	"context"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetPod(t *testing.T) {
	// create a fake clientset that will store state in memory
	fakeClient := fake.NewSimpleClientset()

	ctx := context.Background()
	// initialise our Service
	svc := Service{
		Client: fakeClient,
		Log:    hclog.NewNullLogger(),
	}

	// create a testNamespace and testPod
	testNamespace := "foo-ns"
	testPodName := "bar-pod"
	testPodIP := "192.168.168.100"
	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testPodName,
			Namespace: testNamespace,
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: testPodIP,
		},
	}

	r := require.New(t)

	// create and then add the test pod to the fake clientset
	_, err := fakeClient.CoreV1().Pods(testNamespace).Create(context.Background(), testPod, metav1.CreateOptions{})

	r.NoError(err, "unable to add test Pod to the fake client")

	testCases := []struct {
		name      string
		namespace string
		podName   string
		podIP     string
		phase     corev1.PodPhase
		wantErr   bool
	}{
		{
			name:      "GetPod with valid inputs should return the correct pod",
			namespace: testNamespace,
			podName:   testPodName,
			podIP:     testPodIP,
			phase:     corev1.PodRunning,
			wantErr:   false,
		},
		{
			name:      "GetPod should return an error when the pod is not in running state",
			namespace: testNamespace,
			podName:   testPodName,
			podIP:     testPodIP,
			phase:     corev1.PodPending,
			wantErr:   true,
		},
		{
			name:      "GetPod should return an error when the pod does not have an IP address",
			namespace: testNamespace,
			podName:   testPodName,
			podIP:     "",
			phase:     corev1.PodRunning,
			wantErr:   true,
		},
		{
			name:      "GetPod should return an error when the pod is not found",
			namespace: testNamespace,
			podName:   "this-pod-does-not-exist",
			podIP:     "",
			phase:     corev1.PodPending,
			wantErr:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// update the test pod with the desired phase and podIP
			testPod.Status.Phase = testCase.phase
			testPod.Status.PodIP = testCase.podIP

			// update the fake clientset with the updated pod that has some fields missing or empty
			_, err := fakeClient.CoreV1().Pods(testNamespace).Update(ctx, testPod, metav1.UpdateOptions{})

			r.NoError(err, "unable to update the Pod in the fake clientset")
			result, err := svc.GetPod(ctx, testCase.podName, testCase.namespace)

			if testCase.wantErr {
				r.Error(err, "wanted an error, but got nil")
				return
			}

			r.NoError(err, "wanted no error, but got %v", err)
			r.NotNil(result, "wanted a pod, but got nil")
			r.Equal(testCase.podName, result.Name, "wanted pod name to be %s, but got %s", testCase.podName, result.Name)
		})
	}
}
