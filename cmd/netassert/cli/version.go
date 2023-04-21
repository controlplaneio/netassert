package main

import (
	"os"

	"github.com/spf13/cobra"
	//
)

var versionCmd = &cobra.Command{
	Use:          "version",
	Short:        "Prints the version and other details associated with the program",
	SilenceUsage: false,
	Run:          versionDetails,
}

// versionDetails - prints build information to the STDOUT
func versionDetails(cmd *cobra.Command, args []string) {
	root := cmd.Root()
	root.SetArgs([]string{"--version"})
	if err := root.Execute(); err != nil {
		lg.Error("Failed to get version details", "error", err)
		os.Exit(1)
	}
}
