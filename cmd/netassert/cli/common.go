package main

import (
	"errors"

	"github.com/controlplaneio/netassert/v2/internal/data"
	"github.com/controlplaneio/netassert/v2/internal/kubeops"
	"github.com/hashicorp/go-hclog"
)

// loadTestCases - Reads test from a file or Directory
func loadTestCases(testCasesFile, testCasesDir string) (data.Tests, error) {
	if testCasesFile == "" && testCasesDir == "" {
		return nil, errors.New("either an input file or an input dir containing the tests must be provided using " +
			"flags (--input-file or --input-dir)")
	}

	if testCasesFile != "" && testCasesDir != "" {
		return nil, errors.New("input must be either a file or a directory but not both i.e use one of " +
			"the flags --input-file or --input-dir")
	}

	var (
		testCases data.Tests
		err       error
	)

	switch {
	case testCasesDir != "":
		testCases, err = data.ReadTestsFromDir(testCasesDir)
	case testCasesFile != "":
		testCases, err = data.ReadTestsFromFile(testCasesFile)
	}

	return testCases, err
}

// createService - creates a new kubernetes operations service
func createService(kubeconfigPath string, l hclog.Logger) (*kubeops.Service, error) {
	// if the user has supplied a kubeConfig file location then
	if kubeconfigPath != "" {
		return kubeops.NewServiceFromKubeConfigFile(kubeconfigPath, l)
	}

	return kubeops.NewDefaultService(l)
}
