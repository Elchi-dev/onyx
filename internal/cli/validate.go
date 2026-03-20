package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Elchi-dev/onyx/internal/config"
)

func newValidateCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:     "validate",
		Short:   "Validate the Onyx config file without starting",
		Example: "  onyx validate\n  onyx validate --config /etc/onyx/onyx.toml",
		RunE: func(cmd *cobra.Command, args []string) error {
			if configPath == "" {
				var err error
				configPath, err = findConfigPath()
				if err != nil {
					return err
				}
			}

			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("✗ Config invalid: %w", err)
			}

			fmt.Printf("✓ Config valid: %s\n", configPath)
			fmt.Printf("  HTTP port:   %d\n", cfg.Server.HTTPPort)
			fmt.Printf("  Dashboard:   port %d, enabled=%v\n", cfg.Dashboard.Port, cfg.Dashboard.Enabled)
			fmt.Printf("  Routes:      %d configured\n", len(cfg.Routes))
			for _, r := range cfg.Routes {
				status := "enabled"
				if !r.Enabled {
					status = "disabled"
				}
				fmt.Printf("    %-40s → %s (%s)\n", r.Host, r.Target, status)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to onyx.toml")
	return cmd
}

// findConfigPath searches standard locations for onyx.toml.
// Duplicated from internal/app intentionally to keep the cli package
// free of an app dependency and avoid an import cycle.
func findConfigPath() (string, error) {
	candidates := []string{"onyx.toml", "/etc/onyx/onyx.toml"}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "onyx", "onyx.toml"))
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no onyx.toml found -- run 'onyx setup' first")
}
