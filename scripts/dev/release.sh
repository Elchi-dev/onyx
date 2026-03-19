#!/usr/bin/env bash
# =============================================================================
#  Onyx — Release Script
#  Bumps version, updates CHANGELOG header, tags, and pushes.
#  GitHub Actions release.yml does the rest automatically.
#  This script is gitignored — it's just for you.
#
#  Usage: ./scripts/dev/release.sh v0.2.0
# =============================================================================
set -euo pipefail

GREEN='\033[0;32m'; RED='\033[0;31m'; CYAN='\033[0;36m'; YELLOW='\033[1;33m'
BOLD='\033[1m'; NC='\033[0m'

ok()   { echo -e "${GREEN}✓${NC} $*"; }
info() { echo -e "${CYAN}→${NC} $*"; }
die()  { echo -e "${RED}✗${NC} $*" >&2; exit 1; }

VERSION="${1:-}"
[[ -n "$VERSION" ]] || die "Usage: $0 <version>  (e.g. $0 v0.2.0)"
[[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9] ]] || die "Version must follow vMAJOR.MINOR.PATCH (e.g. v0.2.0)"

BRANCH=$(git rev-parse --abbrev-ref HEAD)
[[ "$BRANCH" == "main" ]] || die "Releases must be cut from main. Currently on: ${BRANCH}"

# ── Check working tree is clean ───────────────────────────────────────────────
if ! git diff --quiet || ! git diff --cached --quiet; then
  die "Working tree is dirty. Commit or stash changes first."
fi

echo ""
echo -e "${BOLD}  Onyx Release${NC} — ${CYAN}${VERSION}${NC}"
echo ""

# ── Run full test suite ───────────────────────────────────────────────────────
info "Running full test suite..."
go test ./... -race -count=1 || die "Tests failed. Fix before releasing."
ok "Tests passed"

# ── Update CHANGELOG.md ───────────────────────────────────────────────────────
DATE=$(date +%Y-%m-%d)
info "Updating CHANGELOG.md with ${VERSION} / ${DATE}..."

# Insert a new unreleased section marker after the first # Changelog line.
sed -i "s/^## \[Unreleased\]/## [${VERSION}] — ${DATE}\n\n### Changed\n\n- (describe changes here)\n\n---\n\n## [Unreleased]/" CHANGELOG.md 2>/dev/null || true

# If there's no [Unreleased] section, just prepend one for next time.
if ! grep -q "\[Unreleased\]" CHANGELOG.md; then
  sed -i "2i\\\\n## [Unreleased]\\n" CHANGELOG.md
fi

ok "CHANGELOG.md updated. Opening for you to fill in release notes..."
sleep 1

# Open CHANGELOG in $EDITOR or nano.
"${EDITOR:-nano}" CHANGELOG.md

# ── Commit and tag ────────────────────────────────────────────────────────────
git add CHANGELOG.md
git commit -m "chore: release ${VERSION}"
ok "Release commit created"

git tag -a "${VERSION}" -m "Release ${VERSION}"
ok "Tag ${VERSION} created"

# ── Push ──────────────────────────────────────────────────────────────────────
info "Pushing to origin/main and pushing tag..."
git push origin main
git push origin "${VERSION}"
ok "Pushed — GitHub Actions will build and publish the release automatically."

echo ""
echo -e "${GREEN}${BOLD}  Release ${VERSION} is on its way!${NC}"
echo "  Watch: https://github.com/Elchi-dev/onyx/actions"
echo ""
