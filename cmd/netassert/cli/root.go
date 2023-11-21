package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   = "v2.0.0-dev" // netassert version
	appName   = "NetAssert"  // name of the application
	gitHash   = ""           // the git hash of the build
	buildDate = ""           // build date, will be injected by the build system
)

var rootCmd = &cobra.Command{
	Use:   "netassert",
	Short: "NetAssert is a command line utility to test network connectivity between kubernetes objects",
	Long: "NetAssert is a command line utility to test network connectivity between kubernetes objects. " +
		"It currently supports Deployment, Pod, Statefulset and Daemonset. You can check the traffic flow between these objects or from these " +
		"objects to a remote host or an IP address.",

	Version: fmt.Sprintf("\nNetAssert by control-plane.io\n"+
		"Version: %s\nCommit Hash: %s\nBuild Date: %s\n",
		version, gitHash, buildDate),
}

func init() {
	// add our subcommands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(pingCmd)
}
