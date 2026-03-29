#!/usr/bin/env bash
# =============================================================================
#  Onyx — Release Script
#  Builds all binaries and packages locally, then tags and pushes.
#  GitHub Actions does the actual GitHub Release publishing.
#
#  Usage: bash scripts/dev/release.sh v0.1.2
#
#  What this script does:
#    1. Runs full test suite
#    2. Cross-compiles all 5 platform binaries
#    3. Builds .deb and .rpm packages
#    4. Generates checksums
#    5. Opens CHANGELOG.md in your editor
#    6. Commits the changelog
#    7. Creates the git tag
#    8. Pushes branch + tag  →  GitHub Actions takes over from here
#
#  What this script does NOT do:
#    - Publish the GitHub Release (Actions does that on tag push)
#    - Write your changelog for you
# =============================================================================
set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'

ok()   { echo -e "${GREEN}✓${NC} $*"; }
info() { echo -e "${CYAN}→${NC} $*"; }
warn() { echo -e "${YELLOW}⚠${NC} $*"; }
die()  { echo -e "${RED}✗${NC} $*" >&2; exit 1; }

# ── Args ──────────────────────────────────────────────────────────────────────
VERSION="${1:-}"
[[ -n "$VERSION" ]] || die "Usage: $0 <version>  (e.g. $0 v0.1.2)"
[[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9] ]] || die "Version must follow vMAJOR.MINOR.PATCH[-suffix]"

BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || die "Not a git repository")
[[ "$BRANCH" == "main" ]] || die "Releases must be cut from main. You are on: ${BRANCH}"

echo ""
echo -e "${BOLD}  Onyx Release Builder — ${CYAN}${VERSION}${NC}"
echo "  ─────────────────────────────────────"
echo ""

# ── Working tree must be clean ────────────────────────────────────────────────
if ! git diff --quiet || ! git diff --cached --quiet; then
    die "Working tree has uncommitted changes. Commit or stash first."
fi

# ── Tag must not already exist ────────────────────────────────────────────────
if git tag -l | grep -q "^${VERSION}$"; then
    die "Tag ${VERSION} already exists."
fi

# ── Step 1: Full test suite ───────────────────────────────────────────────────
info "Running full test suite..."
go test ./... -race -count=1 || die "Tests failed. Fix before releasing."
ok "Tests passed"

# ── Step 2: Cross-compile all binaries ───────────────────────────────────────
info "Cross-compiling all platform binaries..."
mkdir -p dist build

LDFLAGS="-s -w -X main.version=${VERSION}"
BFLAGS="-buildvcs=false"

targets=(
    "linux   amd64  onyx-linux-amd64"
    "linux   arm64  onyx-linux-arm64"
    "darwin  amd64  onyx-darwin-amd64"
    "darwin  arm64  onyx-darwin-arm64"
    "windows amd64  onyx-windows-amd64.exe"
)

for target in "${targets[@]}"; do
    read -r goos goarch outname <<< "$target"
    outpath="dist/${outname}"
    GOOS=$goos GOARCH=$goarch go build \
        -ldflags "$LDFLAGS" $BFLAGS \
        -o "$outpath" ./cmd/onyx
    size=$(du -sh "$outpath" | cut -f1)
    ok "${outname} (${size})"
done

# Also build the local binary for current platform
go build -ldflags "$LDFLAGS" $BFLAGS -o build/onyx ./cmd/onyx
ok "Local binary at build/onyx"

# ── Step 3: Build .deb and .rpm ───────────────────────────────────────────────
if command -v nfpm &>/dev/null; then
    info "Building .deb package..."
    ONYX_VERSION="${VERSION}" nfpm package \
        --config nfpm.yaml \
        --packager deb \
        --target dist/ 2>/dev/null && ok ".deb built" || warn ".deb build failed (non-fatal)"

    info "Building .rpm package..."
    ONYX_VERSION="${VERSION}" nfpm package \
        --config nfpm.yaml \
        --packager rpm \
        --target dist/ 2>/dev/null && ok ".rpm built" || warn ".rpm build failed (non-fatal)"
else
    warn "nfpm not found — skipping .deb/.rpm. Install: go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest"
fi

# ── Step 4: Checksums ─────────────────────────────────────────────────────────
info "Generating checksums..."
cd dist
sha256sum * > checksums.txt
ok "checksums.txt generated:"
cat checksums.txt | sed 's/^/    /'
cd ..

echo ""
echo -e "${BOLD}  All artifacts ready in dist/${NC}"
ls -lh dist/
echo ""

# ── Step 5: Edit CHANGELOG ────────────────────────────────────────────────────
echo -e "${YELLOW}  ──────────────────────────────────────────────────${NC}"
echo -e "${YELLOW}  Now edit CHANGELOG.md for ${VERSION}${NC}"
echo -e "${YELLOW}  Add your release notes under ## [${VERSION}]${NC}"
echo -e "${YELLOW}  ──────────────────────────────────────────────────${NC}"
echo ""
read -rp "  Press ENTER to open CHANGELOG.md in your editor (or Ctrl+C to abort)..."

# Insert version header into CHANGELOG if not already there
DATE=$(date +%Y-%m-%d)
if ! grep -q "\[${VERSION}\]" CHANGELOG.md; then
    # Insert after ## [Unreleased]
    sed -i "s/^## \[Unreleased\]/## [Unreleased]\n\n---\n\n## [${VERSION}] — ${DATE}/" CHANGELOG.md
fi

"${EDITOR:-nano}" CHANGELOG.md

echo ""
read -rp "  Changelog looks good? Commit and tag? (y/N): " confirm
[[ "${confirm,,}" == "y" ]] || die "Aborted."

# ── Step 6: Commit changelog ──────────────────────────────────────────────────
git add CHANGELOG.md
git diff --cached --quiet && warn "No changes to CHANGELOG.md — committing anyway" || true
git commit -m "chore: release ${VERSION}"
ok "Changelog committed"

# ── Step 7: Tag ───────────────────────────────────────────────────────────────
git tag -a "${VERSION}" -m "Release ${VERSION}"
ok "Tag ${VERSION} created"

# ── Step 8: Push ─────────────────────────────────────────────────────────────
info "Pushing main and tag to origin..."
git push origin main
git push origin "${VERSION}"
ok "Pushed — GitHub Actions will now publish the release automatically"

echo ""
echo -e "${GREEN}${BOLD}  Release ${VERSION} is on its way!${NC}"
echo ""
echo "  Watch the build:  https://github.com/Elchi-dev/onyx/actions"
echo "  Release page:     https://github.com/Elchi-dev/onyx/releases/tag/${VERSION}"
echo "  Artifacts in:     dist/"
echo ""
