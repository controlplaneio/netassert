package engine

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/controlplaneio/netassert/v2/internal/data"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var sampleTest = `
- name: busybox-deploy-to-echoserver-deploy
  type: k8s
  protocol: tcp
  targetPort: 8080
  timeoutSeconds: 20
  attempts: 3
  exitCode: 0
  src:
    k8sResource:
      kind: deployment
      name: busybox
      namespace: busybox
  dst:
    k8sResource:
      kind: deployment
      name: echoserver
      namespace: echoserver
`

func TestEngine_RunTCPTest(t *testing.T) {
	t.Run("error when BuildEphemeralScannerContainer fails", func(t *testing.T) {
		r := require.New(t)
		ctx := context.Background()
		testCases, err := data.NewFromReader(strings.NewReader(sampleTest))
		r.Nil(err)

		// we only expect a single testCase to be present

		r.Equal(len(testCases), 1)

		tc := testCases[0]
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockRunner := NewMockNetAssertTestRunner(mockCtrl)

		mockRunner.EXPECT().
			GetPodInDeployment(ctx, tc.Src.K8sResource.Name, tc.Src.K8sResource.Namespace).
			Return(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "busybox",
					Namespace: "busybox",
				},
				Spec:   corev1.PodSpec{},
				Status: corev1.PodStatus{},
			}, nil)

		mockRunner.EXPECT().
			GetPodInDeployment(ctx, tc.Dst.K8sResource.Name, tc.Dst.K8sResource.Namespace).
			Return(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "echoserver",
					Namespace: "echoserver",
				},
				Spec:   corev1.PodSpec{},
				Status: corev1.PodStatus{},
			}, nil)

		mockRunner.EXPECT().
			BuildEphemeralScannerContainer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&corev1.EphemeralContainer{},
				fmt.Errorf("failed to build ephemeral scanner container"))

		eng := New(mockRunner, hclog.NewNullLogger())

		err = eng.RunTCPTest(ctx, testCases[0],
			"scanner-container-name",
			"scanner-container-image", 7)

		r.Error(err)
		wantErrMsg := `unable to build ephemeral scanner container for test ` + tc.Name
		r.Contains(err.Error(), wantErrMsg)
	})
}
