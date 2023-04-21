package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/controlplaneio/netassert/v2/internal/engine"
	"github.com/spf13/cobra"
)

var (
	tapFile                 = "results.tap" // name of the default TAP file where the results will be written
	suffixLength            = 9             // suffix length of the random string to be appended to the container name
	snifferContainerImage   = "docker.io/controlplane/netassertv2-packet-sniffer:latest"
	snifferContainerPrefix  = "netassertv2-sniffer"
	scannerContainerImage   = "docker.io/controlplane/netassertv2-l4-client:latest"
	scannerContainerPrefix  = "netassertv2-client"
	pauseInSeconds          = 1 // time to pause before each test case
	packetCaputureInterface = `eth0`
)

var runCmd = &cobra.Command{
	Use: "run",
	Short: "Run the program with the specified source file or source directory. Only one of the two " +
		"flags (--input-file and --input-dir) can be used at a time. The --input-dir " +
		"flag only reads the first level of the directory and does not recursively scan it",
	Long: "Run the program with the specified source file or source directory. Only one of the two " +
		"flags (--input-file and --input-dir) can be used at a time. The --input-dir " +
		"flag only reads the first level of the directory and does not recursively scan it.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(cmd, args); err != nil {
			lg.Error("❌ ❌ Run failed", "error", err)
			os.Exit(1)
		}
	},
	Version: rootCmd.Version,
}

// run - runs the netAssert Test(s)
func run(cmd *cobra.Command, args []string) error {
	testCases, err := loadTestCases()
	if err != nil {
		return fmt.Errorf("unable to load test cases: %w", err)
	}

	k8sSvc, err := createService()
	if err != nil {
		return fmt.Errorf("failed to build K8s client: %w", err)
	}

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	// ping the kubernetes cluster and check to see if
	// it is alive and that it has support for ephemeral container(s)
	ping()

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
			ctx,                    // context to use
			testCases,              // net assert test cases
			snifferContainerPrefix, // prefix used for the sniffer container name
			snifferContainerImage,  // sniffer container image location
			scannerContainerPrefix, // scanner container prefix used in the container name
			scannerContainerImage,  // scanner container image location
			suffixLength,           // length of random string that will be appended to the snifferContainerPrefix and scannerContainerPrefix
			time.Duration(pauseInSeconds)*time.Second, // pause duration between each test
			packetCaputureInterface,                   // the interface used by the sniffer image to capture traffic
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

	return genResult(testCases)
}

func init() {
	runCmd.Flags().StringVarP(&tapFile, "tap", "t", tapFile, "output tap file containing the tests results")
	runCmd.Flags().IntVarP(&suffixLength, "suffix-length", "", suffixLength, "length of the random suffix that will appended to the scanner/sniffer containers")
	runCmd.Flags().StringVarP(&snifferContainerImage, "sniffer-image", "", snifferContainerImage, "container image to be used as sniffer")
	runCmd.Flags().StringVarP(&snifferContainerPrefix, "sniffer-suffix", "", snifferContainerPrefix, "prefix of the sniffer container")
	runCmd.Flags().StringVarP(&scannerContainerImage, "scanner-image", "", scannerContainerImage, "container image to be used as scanner")
	runCmd.Flags().StringVarP(&scannerContainerPrefix, "scanner-suffix", "", scannerContainerPrefix, "prefix of the scanner debug container name")
	runCmd.Flags().IntVarP(&pauseInSeconds, "pause-sec", "", 1, "no. of seconds to pause before running each test case")
	runCmd.Flags().StringVarP(&packetCaputureInterface, "interface", "", "eth0", "the network interface used by the sniffer container to capture packets")
}
