package data

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadTestFile(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		confFilepath   string
		wantErrMatches []string
	}{
		"existing valid file": {
			confFilepath: "./testdata/valid/empty.yaml",
		},
		"existing invalid file": {
			confFilepath:   "./testdata/invalid/not-a-list.yaml",
			wantErrMatches: []string{"failed to unmarshal tests"},
		},
		"not existing file": {
			confFilepath:   "./testdata/fake-dir/fake-file.yaml",
			wantErrMatches: []string{"no such file or directory"},
		},
		"empty file path": {
			confFilepath:   "",
			wantErrMatches: []string{"input fileName parameter can not be empty string"},
		},
	}

	for name, tt := range tests {
		tc := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := ReadTestsFromFile(tc.confFilepath)

			require.NotEqualf(t, len(tc.wantErrMatches) > 0, err == nil, "expecting an error: %v, got: %v",
				tc.wantErrMatches, err)
			for _, wem := range tc.wantErrMatches {
				require.Equalf(t, strings.Contains(err.Error(), wem), true,
					"expecting error to contain: %s, got: %v", wem, err)
			}
		})
	}
}

func TestReadTestsFromDir(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		confDir string
		errMsg  string
		wantErr bool
	}{
		"existing dir valid tests": {
			confDir: "./testdata/valid",
		},
		"dir without yaml files": {
			confDir: "./testdata/dir-without-yaml-files",
		},
		"existing dir with invalid tests": {
			confDir: "./testdata/invalid",
			errMsg:  "failed to unmarshal tests",
			wantErr: true,
		},
		"not existing dir": {
			confDir: "./testdata/fake-dir",
			errMsg:  "no such file or directory",
			wantErr: true,
		},
		"duplicated test names in different files": {
			confDir: "./testdata/invalid-duplicated-names",
			errMsg:  "duplicate test name found",
			wantErr: true,
		},
		"empty file path": {
			confDir: "",
			errMsg:  "input dir parameter cannot be empty string",
			wantErr: true,
		},
	}
	for name, tt := range tests {
		tc := tt
		t.Run(name, func(t *testing.T) {
			r := require.New(t)
			t.Parallel()

			_, err := ReadTestsFromDir(tc.confDir)

			if tc.wantErr {
				// we are expecting an error here
				r.Error(err)
				r.Contains(err.Error(), tc.errMsg)
				return
			}

			r.NoError(err)
		})
	}
}
