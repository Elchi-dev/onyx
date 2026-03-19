package cli

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/Elchi-dev/onyx/internal/app"
	"github.com/Elchi-dev/onyx/internal/logger"
	"github.com/Elchi-dev/onyx/internal/wizard"
)

func newStartCommand() *cobra.Command {
	var (
		configPath string
		dev        bool
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Onyx proxy and dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			version, _ := cmd.Root().Version, ""
			return runStart(configPath, dev, version)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to onyx.toml (auto-detected if omitted)")
	cmd.Flags().BoolVar(&dev, "dev", false, "Development mode: verbose text logs, debug level")
	return cmd
}

func runStart(configPath string, dev bool, version string) error {
	log := logger.New(0, dev)

	// Auto-run setup if no config or DB exists yet.
	if needsSetup(configPath) {
		log.Info("no configuration found — starting first-time setup")
		w := wizard.New()
		var err error
		configPath, err = w.Run()
		if err != nil {
			return fmt.Errorf("setup failed: %w", err)
		}
	}

	a, err := app.NewFromConfigPath(configPath, dev, version)
	if err != nil {
		return err
	}
	defer a.Close()

	if err := a.Run(); err != nil {
		return friendlyError(err)
	}
	return nil
}

// needsSetup returns true when there is no usable config or database yet.
func needsSetup(configPath string) bool {
	if configPath != "" {
		return false
	}
	_, err := app.NewFromConfigPath("", false, "")
	if err != nil {
		return true
	}
	return false
}

// friendlyError converts low-level errors into helpful user-facing messages.
func friendlyError(err error) error {
	if errors.Is(err, syscall.EACCES) || errors.Is(err, syscall.EPERM) {
		return fmt.Errorf(
			"%w\n\n"+
				"  Ports below 1024 require elevated privileges.\n"+
				"  Options:\n"+
				"  1. Use a port above 1024 in onyx.toml (e.g. http_port = 8080)\n"+
				"  2. Run as root: sudo onyx start\n"+
				"  3. Grant capability: sudo setcap cap_net_bind_service=+ep $(which onyx)\n"+
				"  4. Use systemd: systemctl start onyx (the service handles this automatically)",
			err,
		)
	}
	return err
}
