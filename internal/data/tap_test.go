package data

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTests_TAPResult(t *testing.T) {
	tests := []struct {
		name    string
		tests   Tests
		want    string
		wantErr bool
	}{
		{
			name: "multiple tests",
			tests: Tests{
				&Test{Name: "test1", Pass: false, FailureReason: "example failure reason"},
				&Test{Name: "pod2pod", Pass: true},
				&Test{Name: "---", Pass: true},
				&Test{Name: "don'tknow", Pass: false},
			},
			want: `TAP version 14
1..4
not ok 1 - test1
  ---
  reason: example failure reason
  ...
ok 2 - pod2pod
ok 3 - ---
not ok 4 - don'tknow
  ---
  reason: ""
  ...
`,
			wantErr: false,
		},
		{
			name:    "emptytests",
			tests:   Tests{},
			wantErr: true,
		},
		{
			name:    "niltests",
			tests:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := tt.tests.TAPResult(w)
			if !tt.wantErr {
				require.NoErrorf(t, err, "Tests.TAPResult() error = %v, wantErr %v", err, tt.wantErr)
			}
			gotW := w.String()
			require.Equalf(t, tt.want, gotW, "Tests.TAPResult() = %v, want %v")
		})
	}
}
