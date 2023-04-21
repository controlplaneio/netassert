package data

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFromReader(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		confFile       string
		want           Tests
		wantErrMatches []string
	}{
		"empty": {
			confFile: "empty.yaml",
			want:     Tests{},
		},
		"not a list": {
			confFile:       "not-a-list.yaml",
			wantErrMatches: []string{"cannot unmarshal"},
		},
		"wrong test values": {
			confFile: "wrong-test-values.yaml",
			wantErrMatches: []string{
				"invalid protocol",
				"targetPort out of range",
				"attempts must",
				"timeoutSeconds must",
				"k8sResource invalid kind",
				"host field is set to empty string",
			},
		},
		"missing fields": {
			confFile: "missing-fields.yaml",
			wantErrMatches: []string{
				"name field is missing",
				"targetPort out of range",
				"attempts must",
				"timeoutSeconds must",
				"invalid",
				"src is currently",
				"dst block must",
			},
		},
		"host as a source": {
			confFile:       "host-as-source.yaml",
			wantErrMatches: []string{"k8sResource field in src is currently the only source allowed"},
		},
		"multiple destination blocks": {
			confFile:       "multiple-dst-blocks.yaml",
			wantErrMatches: []string{"dst field only supports K8sResource or Host but not both"},
		},
		"duplicated test names": {
			confFile:       "duplicated-names.yaml",
			wantErrMatches: []string{"duplicate test name found"},
		},
		"empty resources": {
			confFile: "empty-resources.yaml",
			wantErrMatches: []string{
				"k8sResource name is missing",
				"k8sResource kind is missing",
				"k8sResource namespace is missing",
				"k8sResource invalid kind",
			},
		},
		"host as a destination with udp": {
			confFile:       "host-as-dst-udp.yaml",
			wantErrMatches: []string{"with udp tests the destination must be a k8sResource"},
		},
		"multi valid": {
			confFile: "multi.yaml",
			want: Tests{
				&Test{
					Name:           "testname",
					Type:           "k8s",
					Protocol:       ProtocolTCP,
					Attempts:       3,
					TimeoutSeconds: 15,
					TargetPort:     80,
					ExitCode:       0,
					Src: &Src{
						K8sResource: &K8sResource{
							Name:      "deployment1",
							Kind:      KindDeployment,
							Namespace: "ns1",
						},
					},
					Dst: &Dst{
						Host: &Host{
							Name: "1.1.1.1",
						},
					},
				},
				&Test{
					Name:           "testname2",
					Type:           "k8s",
					Protocol:       ProtocolUDP,
					Attempts:       20,
					TimeoutSeconds: 50,
					TargetPort:     8080,
					ExitCode:       1,
					Src: &Src{
						K8sResource: &K8sResource{
							Name:      "statefulset1",
							Kind:      KindStatefulSet,
							Namespace: "ns1",
						},
					},
					Dst: &Dst{
						K8sResource: &K8sResource{
							Name:      "mypod",
							Kind:      KindPod,
							Namespace: "ns2",
						},
					},
				},
			},
		},
	}
	for name, tt := range tests {
		tc := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var testFile string
			if len(tc.wantErrMatches) > 0 {
				testFile = filepath.Join("./testdata/invalid", tc.confFile)
			} else {
				testFile = filepath.Join("./testdata/valid", tc.confFile)
			}
			c, err := os.Open(filepath.Clean(testFile))
			require.NoError(t, err, "cannot open test file %s, err: %v", testFile, err)

			got, err := NewFromReader(c)

			errc := c.Close()
			require.NoError(t, errc, "cannot close test file %s, err: %v", testFile, errc)

			require.NotEqualf(t, len(tc.wantErrMatches) > 0, err == nil, "expecting an error: %v, got: %v", tc.wantErrMatches, err)
			for _, wem := range tc.wantErrMatches {
				require.Equalf(t, strings.Contains(err.Error(), wem), true, "expecting error to contain: %s, got: %v", wem, err)
			}

			require.Equalf(t, tc.want, got, "expecting config %v, got: %v", tc.want, got)
		})
	}
}
