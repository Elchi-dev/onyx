// Package cli defines the Onyx command-line interface.
// Each command lives in its own file and stays thin —
// all real logic lives in internal/app and other internal packages.
package cli

import "github.com/spf13/cobra"

// NewRootCommand builds and returns the root cobra.Command for Onyx.
func NewRootCommand(version string) *cobra.Command {
	root := &cobra.Command{
		Use:     "onyx",
		Short:   "Onyx — modular reverse proxy with a live dashboard",
		Version: version,
	}
	root.AddCommand(
		newStartCommand(),
		newSetupCommand(),
		newStatusCommand(),
		newValidateCommand(),
		newRouteCommand(),
	)
	return root
}
