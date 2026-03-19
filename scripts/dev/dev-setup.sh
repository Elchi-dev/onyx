#!/usr/bin/env bash
# =============================================================================
#  Onyx — Developer Environment Setup (Arch Linux / Hyprland)
#  Run once after cloning the repo.
#  This script is gitignored — it's just for you.
# =============================================================================
set -euo pipefail

GREEN='\033[0;32m'; CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'
ok()   { echo -e "${GREEN}✓${NC} $*"; }
info() { echo -e "${CYAN}→${NC} $*"; }

echo ""
echo -e "${BOLD}  Onyx Dev Setup${NC}"
echo ""

# ── Go tools ─────────────────────────────────────────────────────────────────
info "Installing Go dev tools..."

go install github.com/air-verse/air@latest
ok "air (live reload)"

go install golang.org/x/tools/cmd/goimports@latest
ok "goimports"

go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
ok "nfpm (deb/rpm packaging)"

# ── golangci-lint ─────────────────────────────────────────────────────────────
if ! command -v golangci-lint &>/dev/null; then
  info "Installing golangci-lint..."
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" latest
  ok "golangci-lint"
else
  ok "golangci-lint already installed: $(golangci-lint version --short 2>/dev/null || echo 'unknown')"
fi

# ── Git hooks ─────────────────────────────────────────────────────────────────
info "Installing git pre-commit hook..."
mkdir -p .git/hooks
cat > .git/hooks/pre-commit << 'HOOK'
#!/usr/bin/env bash
set -e
echo "→ Running pre-commit checks..."
go fmt ./...
go vet ./...
go test ./... -short -count=1
echo "✓ Pre-commit passed"
HOOK
chmod +x .git/hooks/pre-commit
ok "Pre-commit hook installed (fmt + vet + tests on every commit)"

# ── VSCode settings ───────────────────────────────────────────────────────────
info "Writing .vscode/settings.json..."
mkdir -p .vscode
cat > .vscode/settings.json << 'VSCODE'
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.formatTool": "goimports",
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    }
  },
  "editor.rulers": [100],
  "files.trimTrailingWhitespace": true
}
VSCODE
ok ".vscode/settings.json written"

# ── Tidy deps ─────────────────────────────────────────────────────────────────
info "Tidying Go modules..."
go mod tidy
ok "go.sum up to date"

echo ""
echo -e "${GREEN}${BOLD}  Dev environment ready!${NC}"
echo ""
echo "  make build   — build the binary"
echo "  make test    — run all tests"
echo "  make dev     — start with live reload (air)"
echo ""
