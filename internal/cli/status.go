package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	var dashURL string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check whether Onyx is running",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, dashURL+"/api/stats", nil)
			if err != nil {
				return err
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "✗ Onyx is not reachable at %s\n", dashURL)
				return err
			}
			defer resp.Body.Close()
			fmt.Printf("✓ Onyx is running — dashboard at %s (HTTP %d)\n", dashURL, resp.StatusCode)
			return nil
		},
	}

	cmd.Flags().StringVar(&dashURL, "url", "http://localhost:8080", "Dashboard URL to check")
	return cmd
}
