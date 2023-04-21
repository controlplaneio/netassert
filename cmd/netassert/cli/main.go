package main

import (
	"os"

	_ "go.uber.org/automaxprocs"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
