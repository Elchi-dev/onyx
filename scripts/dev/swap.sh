#!/usr/bin/env bash
# =============================================================================
#  Onyx Swap — replace project files from a zip without touching dev tooling
#
#  Usage: bash scripts/dev/swap.sh ~/Downloads/onyx-v0.3.1.zip
#
#  Preserved (never touched):
#    scripts/dev/   .git/   build/   dist/   go.sum   .vscode/
# =============================================================================
set -euo pipefail

GREEN='\033[0;32m'; RED='\033[0;31m'; CYAN='\033[0;36m'; YELLOW='\033[1;33m'; NC='\033[0m'
ok()   { echo -e "${GREEN}✓${NC} $*"; }
info() { echo -e "${CYAN}→${NC} $*"; }
die()  { echo -e "${RED}✗${NC} $*" >&2; exit 1; }

ZIP="${1:-}"
[[ -n "$ZIP" ]] || die "Usage: $0 <path-to-zip>"
[[ -f "$ZIP" ]] || die "File not found: $ZIP"

ZIP_NAME=$(basename "$ZIP" .zip)

echo ""
echo -e "  Onyx Swap — ${CYAN}${ZIP_NAME}.zip${NC}"
echo ""

# ── Extract to temp dir ───────────────────────────────────────────────────────
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

info "Extracting zip to temp directory..."
unzip -q "$ZIP" -d "$TMPDIR"

# Find the project root inside the zip (first subdirectory).
PROJECT_ROOT=""
for d in "$TMPDIR"/*/; do
  if [[ -d "$d" ]]; then
    PROJECT_ROOT="$d"
    break
  fi
done

[[ -n "$PROJECT_ROOT" ]] || die "Could not find project root in zip"
ok "Found project root in zip: $(basename "${PROJECT_ROOT%/}")"

# ── Preview changes ───────────────────────────────────────────────────────────
info "Files that will change:"
rsync --dry-run --archive --checksum \
  --exclude='.git/' \
  --exclude='scripts/dev/' \
  --exclude='build/' \
  --exclude='dist/' \
  --exclude='go.sum' \
  --exclude='.vscode/' \
  --exclude='apt/' \
  --out-format="    %f" \
  "$PROJECT_ROOT" . 2>/dev/null | grep -v '/$' || true

echo ""
read -rp "  Apply swap? (y/N): " confirm
[[ "${confirm,,}" == "y" ]] || { echo "  Aborted."; exit 0; }

# ── Sync files ────────────────────────────────────────────────────────────────
info "Syncing files..."
rsync --archive --checksum \
  --exclude='.git/' \
  --exclude='scripts/dev/' \
  --exclude='build/' \
  --exclude='dist/' \
  --exclude='go.sum' \
  --exclude='.vscode/' \
  --exclude='apt/' \
  "$PROJECT_ROOT" .
ok "Files synced"

# ── go mod tidy ───────────────────────────────────────────────────────────────
info "Running go mod tidy..."
go mod tidy
ok "go.sum updated"

# ── Build ─────────────────────────────────────────────────────────────────────
info "Building..."
make build
ok "Build complete: $(./build/onyx --version)"

echo ""
echo -e "${GREEN}  Swap complete!${NC}"
echo "  Next: git add . && git commit -m 'chore: update to ${ZIP_NAME}'"
echo ""
