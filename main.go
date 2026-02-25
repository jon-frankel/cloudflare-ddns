package main

import (
	"fmt"
	"os"

	"github.com/jonathan/cloudflare-ddns/cmd"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	cmd.SetVersion(version)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
