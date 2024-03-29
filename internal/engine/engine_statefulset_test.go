package engine

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"go.uber.org/mock/gomock"

	"github.com/controlplaneio/netassert/v2/internal/data"
)

func TestEngine_GetPod_StatefulSet(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Cleanup(func() { mockCtrl.Finish() })

	var (
		podName         = "foo-pod"
		namespace       = "default"
		statefulSetName = "foo-statefulset"
	)

	t.Run("test GetPod from statefulSet when Pod exists", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		res := data.K8sResource{
			Kind:      data.KindStatefulSet,
			Name:      statefulSetName,
			Namespace: namespace,
		}

		mockRunner := NewMockNetAssertTestRunner(mockCtrl)

		mockRunner.EXPECT().
			GetPodInStatefulSet(ctx, statefulSetName, namespace).
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

	t.Run("test GetPod from statefulSet when Pod does not exist", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		res := data.K8sResource{
			Kind:      data.KindStatefulSet,
			Name:      statefulSetName,
			Namespace: namespace,
		}

		mockRunner := NewMockNetAssertTestRunner(mockCtrl)

		mockRunner.
			EXPECT().
			GetPodInStatefulSet(ctx, statefulSetName, namespace).
			Return(
				&corev1.Pod{},
				fmt.Errorf("pod not found"),
			)

		eng := New(mockRunner, hclog.NewNullLogger())

		_, err := eng.GetPod(ctx, &res)

		require.Error(t, err)
	})
}
