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

func TestLaunchEphemeralContainerInPod_InvalidEphemeralContainer(t *testing.T) {
	// create a fake pod with no ephemeral containers
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-pod",
			Namespace: "default",
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

	ctx := context.Background()

	fakeClient := fake.NewSimpleClientset(pod)

	// initialise our Service
	svc := Service{
		Client: fakeClient,
		Log:    hclog.NewNullLogger(),
	}

	t.Run("Inject valid ephemeral container", func(t *testing.T) {
		r := require.New(t)

		ephContainerName := "foo-container"

		// create an invalid ephemeral container
		ec, err := svc.BuildEphemeralSnifferContainer(
			ephContainerName, // name of the ephemeral container
			"foo:12",         // image location of the container
			"foo",            // search for this string in the captured packet
			1024,             // snapLength to capture
			"tcp",            // protocol to capture
			3,                // no. of matches that triggers an exit with status 0
			"eth0",           // the network interface to read the packets from
			3,                // timeout for the ephemeral container
		)

		r.NoError(err, "failed to Build Ephemeral Container ")

		pod, _, err := svc.LaunchEphemeralContainerInPod(ctx, pod, ec)

		r.NoError(err)

		gotName := pod.Spec.EphemeralContainers[0].EphemeralContainerCommon.Name

		r.Equal(ephContainerName, gotName)

		r.Equal(len(pod.Spec.EphemeralContainers), 1)
	})

	t.Run("inject valid ephemeral container in a Pod that does not exist", func(t *testing.T) {
		r := require.New(t)

		tmpPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
			},
		}

		// create an invalid ephemeral container
		ec, err := svc.BuildEphemeralSnifferContainer(
			"eph1",    // name of the ephemeral container
			"foo:1.2", // image location of the container
			"foo",     // search for this string in the captured packet
			1024,      // snapLength to capture
			"tcp",     // protocol to capture
			3,         // no. of matches that triggers an exit with status 0
			"eth0",    // the network interface to read the packets from
			3,         // timeout for the ephemeral container
		)

		r.NoError(err, "failed to Build Ephemeral Container ")

		gotPod, _, err := svc.LaunchEphemeralContainerInPod(ctx, tmpPod, ec)
		r.Nil(gotPod)
		r.Error(err)
		r.Contains(err.Error(), `pods "test-pod" not found`)
	})
}
