#!/usr/bin/env bash
# =============================================================================
#  Onyx — Swap Script
#  Replaces project files from a zip while preserving your local state.
#
#  Usage: bash scripts/dev/swap.sh <path-to-zip>
#  e.g.:  bash scripts/dev/swap.sh ~/Downloads/onyx-v0.1.2.zip
#
#  PRESERVED (never touched):
#    scripts/dev/    — your private dev scripts (including this one)
#    .git/           — git history
#    build/          — compiled binaries
#    go.sum          — dependency lock file
#    .vscode/        — your editor settings
#    onyx.toml       — local config if any
# =============================================================================
set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'

ok()   { echo -e "${GREEN}✓${NC} $*"; }
info() { echo -e "${CYAN}→${NC} $*"; }
warn() { echo -e "${YELLOW}⚠${NC} $*"; }
die()  { echo -e "${RED}✗${NC} $*" >&2; exit 1; }

# ── Args ──────────────────────────────────────────────────────────────────────
ZIP="${1:-}"
[[ -n "$ZIP" ]] || die "Usage: $0 <path-to-zip>  (e.g. $0 ~/Downloads/onyx-v0.1.2.zip)"
[[ -f "$ZIP" ]] || die "File not found: $ZIP"

# Resolve to absolute path before we cd anywhere
ZIP="$(realpath "$ZIP")"

PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$PROJECT_ROOT"

echo ""
echo -e "${BOLD}  Onyx Swap${NC} — ${CYAN}$(basename "$ZIP")${NC}"
echo ""

# ── Temp directory ────────────────────────────────────────────────────────────
TMPDIR_SWAP=$(mktemp -d)
trap 'rm -rf "$TMPDIR_SWAP"' EXIT

info "Extracting zip to temp directory..."
unzip -q "$ZIP" -d "$TMPDIR_SWAP"

# Find the root of the extracted content (handle nested folder in zip)
EXTRACTED_ROOT=$(find "$TMPDIR_SWAP" -mindepth 1 -maxdepth 2 -name "go.mod" | head -1 | xargs dirname)
[[ -n "$EXTRACTED_ROOT" ]] || die "Could not find go.mod in zip — is this an Onyx zip?"
ok "Found project root in zip: $(basename "$EXTRACTED_ROOT")"

# ── Define what to preserve ───────────────────────────────────────────────────
# These paths are NEVER overwritten — they are local-only state
PRESERVE=(
    "scripts/dev"
    ".git"
    "build"
    "go.sum"
    ".vscode"
    "onyx.toml"
)

# Build rsync exclude args
EXCLUDE_ARGS=()
for p in "${PRESERVE[@]}"; do
    EXCLUDE_ARGS+=(--exclude="/$p")
done

# ── Dry run preview ───────────────────────────────────────────────────────────
info "Files that will change:"
rsync -av --dry-run --delete \
    "${EXCLUDE_ARGS[@]}" \
    "$EXTRACTED_ROOT/" \
    "$PROJECT_ROOT/" \
    | grep -v "^sending\|^sent\|^total\|^$\|/$" \
    | head -40 \
    | sed 's/^/    /'
echo ""

read -rp "  Apply swap? (y/N): " confirm
[[ "${confirm,,}" == "y" ]] || { warn "Aborted — no changes made."; exit 0; }

# ── Sync ─────────────────────────────────────────────────────────────────────
info "Syncing files..."
rsync -a --delete \
    "${EXCLUDE_ARGS[@]}" \
    "$EXTRACTED_ROOT/" \
    "$PROJECT_ROOT/"
ok "Files synced"

# ── Post-swap: tidy and build ─────────────────────────────────────────────────
info "Running go mod tidy..."
go mod tidy
ok "go.sum updated"

info "Building..."
make build
ok "Build complete: $(./build/onyx --version 2>/dev/null || echo 'built')"

echo ""
echo -e "${GREEN}${BOLD}  Swap complete!${NC}"
echo ""
echo "  Next: git add . && git commit -m 'chore: update to $(basename $ZIP .zip)'"
echo ""
