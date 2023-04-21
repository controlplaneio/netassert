package main

import (
	"fmt"
	"os"

	"github.com/controlplaneio/netassert/v2/internal/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

var (
	logLevel      string         // our log leve
	lg            hclog.Logger   // command logger
	kubeConfig    string         // location of kubeconfig file
	testCasesFile string         // location of test cases file
	testCasesDir  string         // location of directory containing the test cases
	version       = "v2.0.0-dev" // netassert version
	appName       = "NetAssert"  // name of the application
	gitHash       = ""           // the git hash of the build
	buildDate     = ""           // build date, will be injected by the build system
)

var rootCmd = &cobra.Command{
	Use:   "netassert",
	Short: "NetAssert is a command line utility to test network connectivity between kubernetes objects",
	Long: "NetAssert is a command line utility to test network connectivity between kubernetes objects. " +
		"It currently supports Deployment, Pod, Statefulset and Daemonset. You can check the traffic flow between these objects or from these " +
		"objects to a remote host or an IP address",

	PersistentPreRun: initCommon,
	Version: fmt.Sprintf("\nNetAssert by control-plane.io\n"+
		"Version: %s\nCommit Hash: %s\nBuild Date: %s\n",
		version, gitHash, buildDate),
}

// init the logger and log basic app info
func initCommon(cmd *cobra.Command, args []string) {
	lg = logger.NewHCLogger(logLevel, fmt.Sprintf("%s-%s", appName, version), os.Stdout)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "k", kubeConfig, "kubeconfig path")
	rootCmd.PersistentFlags().StringVarP(&testCasesFile, "input-file", "f", "", "input test file that contains a list of netassert tests")
	rootCmd.PersistentFlags().StringVarP(&testCasesDir, "input-dir", "d", "", "input test directory that contains a list of netassert test files")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "set log level (info, debug or trace)")

	// add our subcommands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(pingCmd)
}
