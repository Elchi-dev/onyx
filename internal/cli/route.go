package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Elchi-dev/onyx/internal/database"
)

func newRouteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route",
		Short: "Manage proxy routes",
	}
	cmd.AddCommand(
		newRouteAddCommand(),
		newRouteListCommand(),
		newRouteRemoveCommand(),
	)
	return cmd
}

func newRouteAddCommand() *cobra.Command {
	var host, target string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new proxy route",
		Example: "  onyx route add --host api.example.com --target http://localhost:3000",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()
			if err := db.UpsertRoute(host, target, true); err != nil {
				return err
			}
			fmt.Printf("✓ Route added: %s → %s\n", host, target)
			fmt.Println("  Restart Onyx (or hot-reload in v0.3.0) to apply.")
			return nil
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Incoming hostname (e.g. api.example.com)")
	cmd.Flags().StringVar(&target, "target", "", "Backend target URL (e.g. http://localhost:3000)")
	_ = cmd.MarkFlagRequired("host")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}

func newRouteListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered routes",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()
			routes, err := db.ListRoutes()
			if err != nil {
				return err
			}
			if len(routes) == 0 {
				fmt.Println("No routes registered.")
				fmt.Println("Add one: onyx route add --host example.com --target http://localhost:3000")
				return nil
			}
			fmt.Printf("\n  %-40s %-40s %s\n", "HOST", "TARGET", "ENABLED")
			fmt.Printf("  %-40s %-40s %s\n", "────────────────────────────────────────",
				"────────────────────────────────────────", "───────")
			for _, r := range routes {
				enabled := "yes"
				if !r.Enabled {
					enabled = "no"
				}
				fmt.Printf("  %-40s %-40s %s\n", r.Host, r.Target, enabled)
			}
			fmt.Println()
			return nil
		},
	}
}

func newRouteRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "remove [host]",
		Short:   "Remove a route by hostname",
		Example: "  onyx route remove api.example.com",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()
			if err := db.DeleteRoute(args[0]); err != nil {
				return err
			}
			fmt.Printf("✓ Route removed: %s\n", args[0])
			return nil
		},
	}
}

// openDB locates and opens the Onyx database from the default data directory.
func openDB() (*database.DB, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("finding home dir: %w", err)
	}
	dbPath := filepath.Join(home, ".config", "onyx", "onyx.db")
	return database.Open(dbPath)
}
