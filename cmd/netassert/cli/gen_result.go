package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"

	"github.com/controlplaneio/netassert/v2/internal/data"
)

// genResult - Prints results to Stdout and writes it to a Tap file
func genResult(testCases data.Tests, tapFile string, lg hclog.Logger) error {
	failedTestCases := 0

	for _, v := range testCases {
		// increment the no. of test cases
		if v.Pass {
			lg.Info("✅ Test Result", "Name", v.Name, "Pass", v.Pass)
			continue
		}

		lg.Info("❌ Test Result", "Name", v.Name, "Pass", v.Pass, "FailureReason", v.FailureReason)
		failedTestCases++
	}

	tf, err := os.Create(tapFile)
	if err != nil {
		return fmt.Errorf("unable to create tap file %q: %w", tapFile, err)
	}

	if err := testCases.TAPResult(tf); err != nil {
		return fmt.Errorf("unable to generate tap results: %w", err)
	}

	if err := tf.Close(); err != nil {
		return fmt.Errorf("unable to close tap file %q: %w", tapFile, err)
	}

	lg.Info("✍ Wrote test result in a TAP File", "fileName", tapFile)

	if failedTestCases > 0 {
		return fmt.Errorf("total %v test cases have failed", failedTestCases)
	}

	return nil
}
