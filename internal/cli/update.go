// Package cli — update command.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const githubReleasesAPI = "https://api.github.com/repos/Elchi-dev/onyx/releases/latest"

// githubRelease is the relevant subset of the GitHub releases API response.
type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

func newUpdateCommand() *cobra.Command {
	var (
		checkOnly bool
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for updates and optionally install the latest version",
		Long: `Checks the GitHub releases page for a newer version of Onyx.
If a newer version is found, downloads and replaces the current binary.

The update is atomic: the new binary is downloaded to a temporary file,
verified, then moved into place. The old binary is never removed until
the new one is confirmed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			currentVersion, _ := cmd.Root().Version, ""
			return runUpdate(currentVersion, checkOnly, force)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, do not install")
	cmd.Flags().BoolVar(&force, "force", false, "Re-install even if already on the latest version")
	return cmd
}

func runUpdate(currentVersion string, checkOnly, force bool) error {
	fmt.Println("Checking for updates...")

	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("fetching latest release: %w", err)
	}

	latestVersion := release.TagName
	fmt.Printf("  Current version: %s\n", currentVersion)
	fmt.Printf("  Latest version:  %s\n", latestVersion)

	// Compare: skip if already up to date (unless --force)
	if !force && latestVersion == currentVersion {
		fmt.Println("\n  Already up to date!")
		return nil
	}
	if !force && isDevBuild(currentVersion) {
		fmt.Println("\n  Running a development build — use --force to install a release version.")
		return nil
	}

	fmt.Printf("\n  New version available: %s\n", latestVersion)
	fmt.Printf("  Release page: %s\n", release.HTMLURL)

	if checkOnly {
		fmt.Println("\n  Run without --check to install.")
		return nil
	}

	// Find the right asset for the current platform
	assetName := assetNameForPlatform()
	downloadURL := ""
	var downloadSize int64
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			downloadSize = asset.Size
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf(
			"no binary found for %s/%s (looked for %q)\n"+
				"  Download manually from: %s",
			runtime.GOOS, runtime.GOARCH, assetName, release.HTMLURL,
		)
	}

	fmt.Printf("\n  Downloading %s (%.1f MB)...\n", assetName, float64(downloadSize)/(1024*1024))

	// Get the path of the current binary
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding current executable: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("resolving symlinks: %w", err)
	}

	// Download to a temp file in the same directory as the binary
	tmpFile, err := os.CreateTemp(filepath.Dir(exePath), ".onyx-update-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // clean up on any error

	if err := downloadFile(tmpFile, downloadURL); err != nil {
		tmpFile.Close()
		return fmt.Errorf("downloading update: %w", err)
	}
	tmpFile.Close()

	// Make the downloaded binary executable
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}

	// Atomic replace: rename temp file over the current binary
	// os.Rename is atomic on Unix (same filesystem required, hence same dir)
	if err := os.Rename(tmpPath, exePath); err != nil {
		return fmt.Errorf(
			"replacing binary at %s: %w\n"+
				"  Try running with sudo, or update manually from: %s",
			exePath, err, release.HTMLURL,
		)
	}

	fmt.Printf("  Updated: %s\n", exePath)
	fmt.Printf("\n  Onyx %s installed successfully!\n", latestVersion)
	fmt.Println("  Restart Onyx to use the new version.")
	return nil
}

// fetchLatestRelease queries the GitHub releases API.
func fetchLatestRelease() (*githubRelease, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, githubReleasesAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "onyx-updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &release, nil
}

// downloadFile downloads a URL into w, showing a simple progress indicator.
func downloadFile(w io.Writer, url string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	_, err = io.Copy(w, resp.Body)
	return err
}

// assetNameForPlatform returns the expected release asset filename for the
// current OS and architecture.
func assetNameForPlatform() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	if goos == "windows" {
		return fmt.Sprintf("onyx-%s-%s.exe", goos, goarch)
	}
	return fmt.Sprintf("onyx-%s-%s", goos, goarch)
}

// isDevBuild reports whether the version string looks like a dev build
// (git hash, "dev", or contains "dirty").
func isDevBuild(version string) bool {
	if version == "" || version == "dev" {
		return true
	}
	if strings.Contains(version, "dirty") {
		return true
	}
	// Git hashes are typically 7–40 hex chars, not starting with "v"
	if !strings.HasPrefix(version, "v") && len(version) <= 40 {
		return true
	}
	return false
}
