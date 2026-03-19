#!/usr/bin/env bash
# =============================================================================
#  Onyx — Smart Push Script
#  Runs quality checks, commits with a prompted message, and pushes.
#  This script is gitignored — it's just for you.
# =============================================================================
set -euo pipefail

GREEN='\033[0;32m'; RED='\033[0;31m'; CYAN='\033[0;36m'; YELLOW='\033[1;33m'
BOLD='\033[1m'; NC='\033[0m'

ok()   { echo -e "${GREEN}✓${NC} $*"; }
info() { echo -e "${CYAN}→${NC} $*"; }
warn() { echo -e "${YELLOW}⚠${NC} $*"; }
die()  { echo -e "${RED}✗${NC} $*" >&2; exit 1; }

BRANCH=$(git rev-parse --abbrev-ref HEAD)

echo ""
echo -e "${BOLD}  Onyx Push${NC} → ${CYAN}${BRANCH}${NC}"
echo ""

# ── Format ────────────────────────────────────────────────────────────────────
info "Formatting..."
go fmt ./...
goimports -w . 2>/dev/null || true
ok "Formatted"

# ── Vet ───────────────────────────────────────────────────────────────────────
info "go vet..."
go vet ./... || die "go vet failed. Fix errors before pushing."
ok "Vet passed"

# ── Tests ─────────────────────────────────────────────────────────────────────
info "Running tests..."
go test ./... -short -count=1 || die "Tests failed. Fix before pushing."
ok "Tests passed"

# ── Lint (optional — skip with SKIP_LINT=1) ──────────────────────────────────
if [[ "${SKIP_LINT:-0}" != "1" ]]; then
  if command -v golangci-lint &>/dev/null; then
    info "Linting..."
    golangci-lint run ./... || die "Lint failed. Fix or run SKIP_LINT=1 ./scripts/dev/push.sh"
    ok "Lint passed"
  else
    warn "golangci-lint not found — skipping. Run scripts/dev/dev-setup.sh to install it."
  fi
fi

# ── Staged changes check ──────────────────────────────────────────────────────
if git diff --quiet && git diff --cached --quiet; then
  warn "Nothing to commit."
  echo ""
  read -rp "  Push anyway? (y/N): " push_anyway
  [[ "${push_anyway,,}" == "y" ]] || exit 0
else
  # ── Commit message ──────────────────────────────────────────────────────────
  echo ""
  echo "  Commit prefixes: feat | fix | refactor | docs | test | chore"
  echo ""
  read -rp "  Commit message: " msg
  [[ -n "$msg" ]] || die "Commit message cannot be empty."

  git add -A
  git commit -m "$msg"
  ok "Committed: $msg"
fi

# ── Push ──────────────────────────────────────────────────────────────────────
info "Pushing to origin/${BRANCH}..."
git push origin "$BRANCH"
ok "Pushed to origin/${BRANCH}"
echo ""
