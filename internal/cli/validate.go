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
				return fmt.Errorf("config invalid: %w", err)
			}

			fmt.Printf("\n  Config valid: %s\n\n", configPath)
			fmt.Printf("  %-20s %d\n", "HTTP port:", cfg.Server.HTTPPort)
			fmt.Printf("  %-20s %d\n", "HTTPS port:", cfg.Server.HTTPSPort)
			fmt.Printf("  %-20s %s\n", "Data dir:", cfg.Server.DataDir)
			fmt.Printf("  %-20s port %d, enabled=%v\n\n", "Dashboard:", cfg.Dashboard.Port, cfg.Dashboard.Enabled)

			if len(cfg.Routes) == 0 {
				fmt.Println("  Routes: none configured in TOML (check database for DB routes)")
				fmt.Println()
				return nil
			}

			fmt.Printf("  Routes (%d configured):\n\n", len(cfg.Routes))
			for _, r := range cfg.Routes {
				status := "enabled"
				if !r.Enabled {
					status = "disabled"
				}
				https := ""
				if r.HTTPS {
					https = " [HTTPS]"
				}
				gzip := ""
				if r.Gzip {
					gzip = " [gzip]"
				}
				target := r.Target
				if r.StaticRoot != "" {
					target = "static:" + r.StaticRoot
				}
				fmt.Printf("    %-35s -> %-35s (%s%s%s)\n", r.Host, target, status, https, gzip)
				for _, p := range r.Paths {
					pt := p.Target
					if pt == "" {
						pt = r.Target
					}
					fmt.Printf("      %-33s -> %s\n", p.Path, pt)
				}
			}
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to onyx.toml")
	return cmd
}

// findConfigPath searches standard locations for onyx.toml.
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
