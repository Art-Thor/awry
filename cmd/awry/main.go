package main

import (
	"fmt"
	"os"
)

// version is set via ldflags at build time.
var version = "dev"

func main() {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
