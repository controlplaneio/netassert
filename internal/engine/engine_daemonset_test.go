package engine

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/controlplaneio/netassert/v2/internal/data"
)

func TestEngine_GetPod_DaemonSet(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(func() {
		mockCtrl.Finish()
	})

	var (
		podName       = "foo-pod"
		namespace     = "default"
		daemonSetName = "foo-ds"
	)

	t.Run("GetPod from DaemonSet when Pod exists", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		res := data.K8sResource{
			Kind:      data.KindDaemonSet,
			Name:      daemonSetName,
			Namespace: namespace,
		}

		mockRunner := NewMockNetAssertTestRunner(mockCtrl)

		mockRunner.EXPECT().
			GetPodInDaemonSet(ctx, daemonSetName, namespace).
			Return(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
				},
				Spec:   corev1.PodSpec{},
				Status: corev1.PodStatus{},
			}, nil)

		eng := New(mockRunner, hclog.NewNullLogger())

		pod, err := eng.GetPod(ctx, &res)

		require.NoError(t, err)
		require.Equal(t, pod.Namespace, namespace)
		require.Equal(t, pod.Name, podName)
	})

	//
	t.Run("GetPod from DaemonSet when Pod does not exist", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		res := data.K8sResource{
			Kind:      data.KindDaemonSet,
			Name:      daemonSetName,
			Namespace: namespace,
		}

		mockRunner := NewMockNetAssertTestRunner(mockCtrl)

		mockRunner.
			EXPECT().
			GetPodInDaemonSet(ctx, daemonSetName, namespace).
			Return(
				&corev1.Pod{},
				fmt.Errorf("pod not found"),
			)

		eng := New(mockRunner, hclog.NewNullLogger())

		_, err := eng.GetPod(ctx, &res)

		require.Error(t, err)
	})
}
