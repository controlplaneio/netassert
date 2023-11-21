package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// validateCmdConfig - config for validate sub-command
type validateCmdConfig struct {
	TestCasesFile string
	TestCasesDir  string
}

var (
	validateCmdCfg validateCmdConfig // config for validate sub-command that will be used in the package

	validateCmd = &cobra.Command{
		Use: "validate",
		Short: "verify the syntax and semantic correctness of netassert test(s) in a test file or folder. Only one of the " +
			"two flags (--input-file and --input-dir) can be used at a time.",
		Run:     validateTestCases,
		Version: rootCmd.Version,
	}
)

// validateTestCases - validates test cases from file or directory
func validateTestCases(cmd *cobra.Command, args []string) {

	_, err := loadTestCases(validateCmdCfg.TestCasesFile, validateCmdCfg.TestCasesDir)
	if err != nil {
		fmt.Println("❌ Validation of test cases failed", "error", err)
		os.Exit(1)
	}

	fmt.Println("✅ All test cases are valid syntax-wise and semantically")
}

func init() {
	validateCmd.Flags().StringVarP(&validateCmdCfg.TestCasesFile, "input-file", "f", "", "input test file that contains a list of netassert tests")
	validateCmd.Flags().StringVarP(&validateCmdCfg.TestCasesDir, "input-dir", "d", "", "input test directory that contains a list of netassert test files")
}
