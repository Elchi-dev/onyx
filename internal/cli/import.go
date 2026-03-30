package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Elchi-dev/onyx/internal/database"
	"github.com/Elchi-dev/onyx/internal/nginx"
)

func newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import routes from other sources",
	}
	cmd.AddCommand(newImportNginxCommand())
	return cmd
}

func newImportNginxCommand() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:     "nginx <path>",
		Short:   "Import routes from nginx config file or directory",
		Example: "  onyx import nginx /etc/nginx/sites-available/",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("cannot access %s: %w", path, err)
			}

			var routes []database.Route
			var errs []error

			if info.IsDir() {
				routes, errs = nginx.ParseDir(path)
			} else {
				r, err := nginx.ParseFile(path)
				if err != nil {
					errs = append(errs, err)
				} else {
					routes = r
				}
			}

			if len(errs) > 0 {
				for _, e := range errs {
					fmt.Fprintf(os.Stderr, "  warning: %v\n", e)
				}
			}

			if len(routes) == 0 {
				fmt.Println("No importable routes found.")
				return nil
			}

			fmt.Printf("\n  Found %d route(s):\n\n", len(routes))
			for _, r := range routes {
				https := ""
				if r.HTTPS {
					https = " [HTTPS]"
				}
				paths := ""
				if len(r.Paths) > 0 {
					paths = fmt.Sprintf(" (%d path rules)", len(r.Paths))
				}
				static := ""
				if r.StaticRoot != "" {
					static = fmt.Sprintf(" [static: %s]", filepath.Base(r.StaticRoot))
				}
				fmt.Printf("    %s → %s%s%s%s\n", r.Host, r.Target, https, paths, static)
			}
			fmt.Println()

			if dryRun {
				fmt.Println("  Dry run — no changes written.")
				return nil
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			imported := 0
			for _, r := range routes {
				if err := db.UpsertRoute(r); err != nil {
					fmt.Fprintf(os.Stderr, "  error saving %s: %v\n", r.Host, err)
					continue
				}
				imported++
			}

			fmt.Printf("  Imported %d/%d routes. Run 'onyx start' to apply.\n\n", imported, len(routes))
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be imported without saving")
	return cmd
}
