package main

import (
	"os"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use: "validate",
	Short: "verify the syntax and semantic correctness of netassert test(s) in a test file or folder.Only one of the " +
		"two flags (--input-file and --input-dir) can be used at a time.",
	Run:     validateTestCases,
	Version: rootCmd.Version,
}

// validateTestCases - validates test cases from file or directory
func validateTestCases(cmd *cobra.Command, args []string) {
	if _, err := loadTestCases(); err != nil {
		lg.Error("❌ Validation failed", "error", err)
		os.Exit(1)
	}

	lg.Info("✅ All test cases are valid")
}
