package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

// health endpoint for the K8s server
var apiServerHealthEndpoint = `/healthz`

var pingCmd = &cobra.Command{
	Use: "ping",
	Short: "pings the K8s API server over HTTP to see if it is alive and also checks if the server has support for " +
		"ephemeral containers",
	Long: "pings the K8s API server over HTTP to see if it is alive and also checks if the server has support for " +
		"ephemeral/debug containers",
	Run: func(cmd *cobra.Command, args []string) {
		ping()
	},
	Version: rootCmd.Version,
}

// checkEphemeralContainerSupport checks to see if ephemeral containers are supported by the K8s server
func ping() {
	k8sSvc, err := createService()
	if err != nil {
		lg.Error("Failed to build K8s Client", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	if err := k8sSvc.PingHealthEndpoint(ctx, apiServerHealthEndpoint); err != nil {
		lg.Error("Ping failed", "error", err)
		os.Exit(1)
	}

	lg.Info("✅ Successfully pinged " + apiServerHealthEndpoint + " endpoint of the Kubernetes server")

	if err := k8sSvc.CheckEphemeralContainerSupport(ctx); err != nil {
		lg.Error("❌ Ephemeral containers are not supported by the Kubernetes server",
			"error", err)
		os.Exit(1)
	}

	lg.Info("✅ Ephemeral containers are supported by the Kubernetes server")
}
