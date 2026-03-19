package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Elchi-dev/onyx/internal/wizard"
)

func newSetupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Run the interactive first-time setup wizard",
		Long: `Guides you through:
  - Choosing a data directory
  - Setting a dashboard password
  - Configuring proxy and dashboard ports
  - Optionally adding your first route`,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := wizard.New().Run()
			if err != nil {
				return err
			}
			fmt.Printf("\n  Config saved to: %s\n", path)
			fmt.Println("  Run 'onyx start' to launch.\n")
			return nil
		},
	}
}
