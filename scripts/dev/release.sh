#!/usr/bin/env bash
# =============================================================================
#  Onyx — Master Release Script
#
#  Usage: bash scripts/dev/release.sh v0.1.3
#
#  What this does:
#    1.  Runs full test suite
#    2.  Cross-compiles all 5 platform binaries
#    3.  Builds .deb (amd64 + arm64) and .rpm
#    4.  Updates + signs the apt repository
#    5.  Generates checksums
#    6.  Opens CHANGELOG.md in your editor
#    7.  Commits everything (binaries stay in dist/, apt/ goes into git)
#    8.  Creates the git tag
#    9.  Pushes branch + tag  →  GitHub Actions publishes the GitHub Release
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
[[ -n "$VERSION" ]] || die "Usage: $0 <version>  (e.g. $0 v0.1.3)"
[[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9] ]] || \
    die "Version must follow vMAJOR.MINOR.PATCH[-suffix]"

# Strip leading 'v' for deb/apt (deb versions don't use v prefix)
VERSION_BARE="${VERSION#v}"

BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || die "Not a git repository")
[[ "$BRANCH" == "main" ]] || die "Releases must be cut from main. You are on: ${BRANCH}"

echo ""
echo -e "${BOLD}  Onyx Master Release — ${CYAN}${VERSION}${NC}"
echo "  ─────────────────────────────────────"
echo ""

# ── Clean working tree ────────────────────────────────────────────────────────
if ! git diff --quiet || ! git diff --cached --quiet; then
    die "Working tree has uncommitted changes. Commit or stash first."
fi

if git tag -l | grep -q "^${VERSION}$"; then
    die "Tag ${VERSION} already exists. Delete it first: git tag -d ${VERSION}"
fi

# ── Step 1: Tests ─────────────────────────────────────────────────────────────
info "Running full test suite..."
go test ./... -race -count=1 || die "Tests failed. Fix before releasing."
ok "Tests passed"

# ── Step 2: Cross-compile ─────────────────────────────────────────────────────
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

go build -ldflags "$LDFLAGS" $BFLAGS -o build/onyx ./cmd/onyx
ok "Local binary at build/onyx"

# ── Step 3: .deb + apt repo ───────────────────────────────────────────────────
if command -v nfpm &>/dev/null && command -v dpkg-scanpackages &>/dev/null; then
    info "Building .deb packages and updating apt repository..."
    bash scripts/publish-apt.sh "$VERSION_BARE"
    ok "apt repository updated"
else
    warn "nfpm or dpkg-scanpackages not found — skipping .deb and apt update"
    warn "Install: go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest"
    warn "         sudo pacman -S dpkg  (or apt install dpkg-dev)"
fi

# ── Step 4: .rpm ──────────────────────────────────────────────────────────────
if command -v nfpm &>/dev/null; then
    info "Building .rpm package..."
    ONYX_VERSION="${VERSION_BARE}" ARCH="amd64" \
        envsubst < nfpm.yaml | nfpm pkg --config /dev/stdin \
        --packager rpm --target dist/ 2>/dev/null \
        && ok ".rpm built" || warn ".rpm build failed (non-fatal)"
fi

# ── Step 5: Checksums ─────────────────────────────────────────────────────────
info "Generating checksums..."
cd dist
sha256sum * > checksums.txt
ok "checksums.txt generated"
cd ..

# ── Step 6: Edit CHANGELOG ────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}  ──────────────────────────────────────────────────${NC}"
echo -e "${YELLOW}  Now edit CHANGELOG.md for ${VERSION}${NC}"
echo -e "${YELLOW}  ──────────────────────────────────────────────────${NC}"
echo ""
read -rp "  Press ENTER to open CHANGELOG.md (Ctrl+C to abort)..."

DATE=$(date +%Y-%m-%d)
if ! grep -q "\[${VERSION}\]" CHANGELOG.md; then
    sed -i "s/^## \[Unreleased\]/## [Unreleased]\n\n---\n\n## [${VERSION}] — ${DATE}/" CHANGELOG.md
fi
"${EDITOR:-nano}" CHANGELOG.md

echo ""
read -rp "  Changelog looks good? Commit and tag? (y/N): " confirm
[[ "${confirm,,}" == "y" ]] || die "Aborted."

# ── Step 7: Commit ────────────────────────────────────────────────────────────
git add CHANGELOG.md apt/ dist/checksums.txt
git diff --cached --quiet && warn "Nothing new to commit" || \
    git commit -m "chore: release ${VERSION}"
ok "Release commit created"

# ── Step 8: Tag ───────────────────────────────────────────────────────────────
git tag -a "${VERSION}" -m "Release ${VERSION}"
ok "Tag ${VERSION} created"

# ── Step 9: Push ──────────────────────────────────────────────────────────────
info "Pushing to origin..."
git push origin main
git push origin "${VERSION}"
ok "Pushed — GitHub Actions will publish the release"

echo ""
echo -e "${GREEN}${BOLD}  Release ${VERSION} is on its way!${NC}"
echo ""
echo "  Actions:  https://github.com/Elchi-dev/onyx/actions"
echo "  Release:  https://github.com/Elchi-dev/onyx/releases/tag/${VERSION}"
echo ""
