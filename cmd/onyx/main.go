// Package main is the binary entrypoint for Onyx.
// It is intentionally minimal — all logic lives in internal packages.
package main

import (
	"fmt"
	"os"

	"github.com/Elchi-dev/onyx/internal/cli"
)

// version is injected at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := cli.NewRootCommand(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
