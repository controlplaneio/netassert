package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"

	"github.com/controlplaneio/netassert/v2/internal/engine"
	"github.com/controlplaneio/netassert/v2/internal/logger"
)

// RunConfig - configuration for the run command
type runCmdConfig struct {
	TapFile                string
	SuffixLength           int
	SnifferContainerImage  string
	SnifferContainerPrefix string
	ScannerContainerImage  string
	ScannerContainerPrefix string
	PauseInSeconds         int
	PacketCaptureInterface string
	KubeConfig             string
	TestCasesFile          string
	TestCasesDir           string
	LogLevel               string
}

// Initialize with default values
var runCmdCfg = runCmdConfig{
	TapFile:                "results.tap", // name of the default TAP file where the results will be written
	SuffixLength:           9,             // suffix length of the random string to be appended to the container name
	SnifferContainerImage:  "docker.io/controlplane/netassertv2-packet-sniffer:latest",
	SnifferContainerPrefix: "netassertv2-sniffer",
	ScannerContainerImage:  "docker.io/controlplane/netassertv2-l4-client:latest",
	ScannerContainerPrefix: "netassertv2-client",
	PauseInSeconds:         1,      // seconds to pause before each test case
	PacketCaptureInterface: `eth0`, // the interface used by the sniffer image to capture traffic
	LogLevel:               "info", // log level
}

var runCmd = &cobra.Command{
	Use: "run",
	Short: "Run the program with the specified source file or source directory. Only one of the two " +
		"flags (--input-file and --input-dir) can be used at a time. The --input-dir " +
		"flag only reads the first level of the directory and does not recursively scan it.",
	Long: "Run the program with the specified source file or source directory. Only one of the two " +
		"flags (--input-file and --input-dir) can be used at a time. The --input-dir " +
		"flag only reads the first level of the directory and does not recursively scan it.",
	Run: func(cmd *cobra.Command, args []string) {
		lg := logger.NewHCLogger(runCmdCfg.LogLevel, fmt.Sprintf("%s-%s", appName, version), os.Stdout)
		if err := runTests(lg); err != nil {
			lg.Error(" ❌ Failed to successfully run all the tests", "error", err)
			os.Exit(1)
		}
	},

	Version: rootCmd.Version,
}

// run - runs the netAssert Test(s)
func runTests(lg hclog.Logger) error {
	testCases, err := loadTestCases(runCmdCfg.TestCasesFile, runCmdCfg.TestCasesDir)
	if err != nil {
		return fmt.Errorf("unable to load test cases: %w", err)
	}

	//lg := logger.NewHCLogger(runCmdCfg.LogLevel, fmt.Sprintf("%s-%s", appName, version), os.Stdout)
	k8sSvc, err := createService(runCmdCfg.KubeConfig, lg)
	if err != nil {
		return fmt.Errorf("failed to build K8s client: %w", err)
	}

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	// ping the kubernetes cluster and check to see if
	// it is alive and that it has support for ephemeral container(s)
	ping(ctx)

	// initialise our test runner
	testRunner := engine.New(k8sSvc, lg)
	// initialise our done signal
	done := make(chan struct{})

	// add our test runner to the wait group
	go func() {
		defer func() {
			// once all our go routines have finished notify the done channel
			done <- struct{}{}
		}()

		// run the tests
		testRunner.RunTests(
			ctx,                              // context to use
			testCases,                        // net assert test cases
			runCmdCfg.SnifferContainerPrefix, // prefix used for the sniffer container name
			runCmdCfg.SnifferContainerImage,  // sniffer container image location
			runCmdCfg.ScannerContainerPrefix, // scanner container prefix used in the container name
			runCmdCfg.ScannerContainerImage,  // scanner container image location
			runCmdCfg.SuffixLength,           // length of random string that will be appended to the snifferContainerPrefix and scannerContainerPrefix
			time.Duration(runCmdCfg.PauseInSeconds)*time.Second, // pause duration between each test
			runCmdCfg.PacketCaptureInterface,                    // the interface used by the sniffer image to capture traffic
		)
	}()

	// Wait for the tests to finish or for the context to be canceled
	select {
	case <-done:
		// all our tests have finished running
	case <-ctx.Done():
		lg.Info("Received signal from OS", "msg", ctx.Err())
		// context has been cancelled, we wait for our test runner to finish
		<-done
	}

	return genResult(testCases, runCmdCfg.TapFile, lg)
}

func init() {
	// Bind flags to the runCmd
	runCmd.Flags().StringVarP(&runCmdCfg.TapFile, "tap", "t", runCmdCfg.TapFile, "output tap file containing the tests results")
	runCmd.Flags().IntVarP(&runCmdCfg.SuffixLength, "suffix-length", "s", runCmdCfg.SuffixLength, "length of the random suffix that will appended to the scanner/sniffer containers")
	runCmd.Flags().StringVarP(&runCmdCfg.SnifferContainerImage, "sniffer-image", "i", runCmdCfg.SnifferContainerImage, "container image to be used as sniffer")
	runCmd.Flags().StringVarP(&runCmdCfg.SnifferContainerPrefix, "sniffer-prefix", "p", runCmdCfg.SnifferContainerPrefix, "prefix of the sniffer container")
	runCmd.Flags().StringVarP(&runCmdCfg.ScannerContainerImage, "scanner-image", "c", runCmdCfg.ScannerContainerImage, "container image to be used as scanner")
	runCmd.Flags().StringVarP(&runCmdCfg.ScannerContainerPrefix, "scanner-prefix", "x", runCmdCfg.ScannerContainerPrefix, "prefix of the scanner debug container name")
	runCmd.Flags().IntVarP(&runCmdCfg.PauseInSeconds, "pause-sec", "P", runCmdCfg.PauseInSeconds, "number of seconds to pause before running each test case")
	runCmd.Flags().StringVarP(&runCmdCfg.PacketCaptureInterface, "interface", "n", runCmdCfg.PacketCaptureInterface, "the network interface used by the sniffer container to capture packets")
	runCmd.Flags().StringVarP(&runCmdCfg.TestCasesFile, "input-file", "f", runCmdCfg.TestCasesFile, "input test file that contains a list of netassert tests")
	runCmd.Flags().StringVarP(&runCmdCfg.TestCasesDir, "input-dir", "d", runCmdCfg.TestCasesDir, "input test directory that contains a list of netassert test files")
	runCmd.Flags().StringVarP(&runCmdCfg.KubeConfig, "kubeconfig", "k", runCmdCfg.KubeConfig, "path to kubeconfig file")
	runCmd.Flags().StringVarP(&runCmdCfg.LogLevel, "log-level", "l", "info", "set log level (info, debug or trace)")
}
