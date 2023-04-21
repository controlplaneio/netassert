package engine

import (
	"context"
	"fmt"
	"testing"

	"github.com/controlplaneio/netassert/v2/internal/data"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/mock/gomock"
)

func TestEngine_GetPod_Pod(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		podName   = "foo-pod"
		namespace = "default"
	)

	t.Run("test GetPod from statefulSet when Pod exists", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		res := data.K8sResource{
			Kind:      data.KindPod,
			Name:      podName,
			Namespace: namespace,
		}

		mockRunner := NewMockNetAssertTestRunner(mockCtrl)

		mockRunner.EXPECT().
			GetPod(ctx, podName, namespace).
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

	t.Run("GetPod when Pod does not exists", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		res := data.K8sResource{
			Kind:      data.KindPod,
			Name:      podName,
			Namespace: namespace,
		}

		mockRunner := NewMockNetAssertTestRunner(mockCtrl)

		mockRunner.EXPECT().
			GetPod(ctx, podName, namespace).
			Return(&corev1.Pod{}, fmt.Errorf("pod does not exist"))

		eng := New(mockRunner, hclog.NewNullLogger())

		_, err := eng.GetPod(ctx, &res)

		require.Error(t, err)
	})
}
