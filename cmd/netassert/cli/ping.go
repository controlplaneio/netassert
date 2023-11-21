package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/controlplaneio/netassert/v2/internal/logger"
)

const (
	apiServerHealthEndpoint = `/healthz` // health endpoint for the K8s server
)

type pingCmdConfig struct {
	KubeConfig  string
	PingTimeout time.Duration
}

var pingCmdCfg = pingCmdConfig{}

var pingCmd = &cobra.Command{
	Use: "ping",
	Short: "pings the K8s API server over HTTP(S) to see if it is alive and also checks if the server has support for " +
		"ephemeral containers.",
	Long: "pings the K8s API server over HTTP(S) to see if it is alive and also checks if the server has support for " +
		"ephemeral/debug containers.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), pingCmdCfg.PingTimeout)
		defer cancel()
		ping(ctx)
	},
	Version: rootCmd.Version,
}

// checkEphemeralContainerSupport checks to see if ephemeral containers are supported by the K8s server
func ping(ctx context.Context) {
	lg := logger.NewHCLogger("info", fmt.Sprintf("%s-%s", appName, version), os.Stdout)

	k8sSvc, err := createService(pingCmdCfg.KubeConfig, lg)

	if err != nil {
		lg.Error("Ping failed, unable to build K8s Client", "error", err)
		os.Exit(1)
	}

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

func init() {
	pingCmd.Flags().DurationVarP(&pingCmdCfg.PingTimeout, "timeout", "t", 60*time.Second,
		"Timeout for the ping command")
	pingCmd.Flags().StringVarP(&pingCmdCfg.KubeConfig, "kubeconfig", "k", "", "path to kubeconfig file")
}
