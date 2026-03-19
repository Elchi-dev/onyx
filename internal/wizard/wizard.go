// Package wizard implements the interactive first-run setup wizard for Onyx.
// It guides a new user through creating the data directory, setting a dashboard
// password, choosing ports, and optionally registering a first route.
package wizard

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"

	"github.com/Elchi-dev/onyx/internal/auth"
	"github.com/Elchi-dev/onyx/internal/config"
	"github.com/Elchi-dev/onyx/internal/database"
)

// Wizard orchestrates the interactive first-run setup.
type Wizard struct {
	reader *bufio.Reader
}

// New creates a Wizard that reads from stdin.
func New() *Wizard {
	return &Wizard{reader: bufio.NewReader(os.Stdin)}
}

// Run executes the setup sequence and returns the path to the written config file.
func (w *Wizard) Run() (string, error) {
	w.banner()

	cfg := config.Defaults()

	// ── Step 1: data directory ──────────────────────────────────────────────
	dataDir, err := w.promptDefault("Data directory", cfg.Server.DataDir)
	if err != nil {
		return "", err
	}
	cfg.Server.DataDir = dataDir

	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return "", fmt.Errorf("creating data directory: %w", err)
	}

	// ── Step 2: open / create database ─────────────────────────────────────
	db, err := database.Open(filepath.Join(dataDir, "onyx.db"))
	if err != nil {
		return "", fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	// ── Step 3: one-time encryption key (skip if already generated) ─────────
	if _, exists, _ := db.GetSetting(auth.SettingKeyEncryptionKey); !exists {
		key, err := auth.GenerateKey()
		if err != nil {
			return "", fmt.Errorf("generating encryption key: %w", err)
		}
		if err := db.SetSetting(auth.SettingKeyEncryptionKey, key); err != nil {
			return "", err
		}
		fmt.Println("  ✓ Encryption key generated.")
	}

	// ── Step 4: dashboard password ──────────────────────────────────────────
	if err := w.stepPassword(db); err != nil {
		return "", err
	}

	// ── Step 5: proxy port ──────────────────────────────────────────────────
	if port, err := w.promptInt("HTTP proxy port", cfg.Server.HTTPPort); err == nil {
		cfg.Server.HTTPPort = port
	}

	// ── Step 6: dashboard port ──────────────────────────────────────────────
	if port, err := w.promptInt("Dashboard port", cfg.Dashboard.Port); err == nil {
		cfg.Dashboard.Port = port
	}

	// ── Step 7: first route (optional) ──────────────────────────────────────
	if err := w.stepRoute(db); err != nil {
		return "", err
	}

	// ── Write config ────────────────────────────────────────────────────────
	configPath := filepath.Join(dataDir, "onyx.toml")
	if err := cfg.Write(configPath); err != nil {
		return "", fmt.Errorf("writing config: %w", err)
	}

	fmt.Println()
	fmt.Println("  ✓ Setup complete!")
	fmt.Printf("  ✓ Config: %s\n", configPath)
	fmt.Println()
	fmt.Println("  Run  onyx start  to launch Onyx.")
	fmt.Println()

	return configPath, nil
}

func (w *Wizard) banner() {
	fmt.Println()
	fmt.Println("  ╔════════════════════════════════════╗")
	fmt.Println("  ║    Welcome to Onyx Setup  🪨        ║")
	fmt.Println("  ║    Modular Reverse Proxy            ║")
	fmt.Println("  ╚════════════════════════════════════╝")
	fmt.Println()
}

func (w *Wizard) stepPassword(db *database.DB) error {
	fmt.Println()
	fmt.Print("  Dashboard password (min 8 chars): ")

	password, err := readPassword()
	if err != nil {
		// Fall back to plain input if not in a real terminal.
		password, err = w.readLine()
		if err != nil {
			return err
		}
	} else {
		fmt.Println()
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	if err := db.SetSetting(auth.SettingKeyPasswordHash, hash); err != nil {
		return err
	}
	fmt.Println("  ✓ Dashboard password set.")
	return nil
}

func (w *Wizard) stepRoute(db *database.DB) error {
	fmt.Println()
	fmt.Print("  Add a proxy route now? (y/N): ")
	ans, err := w.readLine()
	if err != nil {
		return err
	}
	if strings.ToLower(ans) != "y" {
		return nil
	}

	fmt.Print("  Host   (e.g. api.example.com): ")
	host, err := w.readLine()
	if err != nil {
		return err
	}

	fmt.Print("  Target (e.g. http://localhost:3000): ")
	target, err := w.readLine()
	if err != nil {
		return err
	}

	if host == "" || target == "" {
		fmt.Println("  ⚠  Skipped (empty input).")
		return nil
	}

	if err := db.UpsertRoute(host, target, true); err != nil {
		return fmt.Errorf("saving route: %w", err)
	}
	fmt.Printf("  ✓ Route %s → %s saved.\n", host, target)
	return nil
}

// promptDefault shows a prompt with a default value. Returns default on empty input.
func (w *Wizard) promptDefault(label, defaultVal string) (string, error) {
	fmt.Printf("  %s [%s]: ", label, defaultVal)
	input, err := w.readLine()
	if err != nil {
		return "", err
	}
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

// promptInt shows a numeric prompt with a default. Returns default on empty input.
func (w *Wizard) promptInt(label string, defaultVal int) (int, error) {
	fmt.Printf("  %s [%d]: ", label, defaultVal)
	input, err := w.readLine()
	if err != nil {
		return defaultVal, err
	}
	if input == "" {
		return defaultVal, nil
	}
	var port int
	if _, err := fmt.Sscanf(input, "%d", &port); err != nil || port < 1 || port > 65535 {
		fmt.Printf("  ⚠  Invalid value %q — using default %d.\n", input, defaultVal)
		return defaultVal, nil
	}
	return port, nil
}

func (w *Wizard) readLine() (string, error) {
	line, err := w.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}
	return strings.TrimSpace(line), nil
}

// readPassword reads a password from the terminal with echo suppressed.
func readPassword() (string, error) {
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	return string(b), nil
}
