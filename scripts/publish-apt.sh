#!/usr/bin/env bash
# scripts/publish-apt.sh
# Updates the apt repository for a given version.
# Usage: bash scripts/publish-apt.sh 0.1.3   (no leading v)
set -euo pipefail

GREEN='\033[0;32m'; CYAN='\033[0;36m'; NC='\033[0m'
ok()   { echo -e "${GREEN}✓${NC} $*"; }
info() { echo -e "${CYAN}→${NC} $*"; }

export ONYX_VERSION="${1:-}"
[[ -n "$ONYX_VERSION" ]] || { echo "Usage: $0 <version>  (e.g. $0 0.1.3)"; exit 1; }

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

publish_arch() {
    local ARCH="$1"
    info "Building .deb for $ARCH..."

    export ARCH
    envsubst < nfpm.yaml | nfpm pkg --config /dev/stdin \
        --target "dist/onyx_${ONYX_VERSION}_${ARCH}.deb"
    ok "dist/onyx_${ONYX_VERSION}_${ARCH}.deb"

    mkdir -p "apt/pool/main/o/onyx/${ARCH}"
    cp "dist/onyx_${ONYX_VERSION}_${ARCH}.deb" \
       "apt/pool/main/o/onyx/${ARCH}/"

    info "Generating Packages index for $ARCH..."
    mkdir -p "apt/dists/stable/main/binary-${ARCH}"
    (cd apt && dpkg-scanpackages "pool/main/o/onyx/${ARCH}" /dev/null \
        > "dists/stable/main/binary-${ARCH}/Packages")
    gzip -9c "apt/dists/stable/main/binary-${ARCH}/Packages" \
        > "apt/dists/stable/main/binary-${ARCH}/Packages.gz"
    ok "Packages index for $ARCH"
}

publish_arch "amd64"
publish_arch "arm64"

info "Updating Release file..."
cd apt

AMD64_SHA=$(sha256sum dists/stable/main/binary-amd64/Packages | cut -d' ' -f1)
AMD64_SIZE=$(stat -c%s dists/stable/main/binary-amd64/Packages)
AMD64GZ_SHA=$(sha256sum dists/stable/main/binary-amd64/Packages.gz | cut -d' ' -f1)
AMD64GZ_SIZE=$(stat -c%s dists/stable/main/binary-amd64/Packages.gz)
ARM64_SHA=$(sha256sum dists/stable/main/binary-arm64/Packages | cut -d' ' -f1)
ARM64_SIZE=$(stat -c%s dists/stable/main/binary-arm64/Packages)
ARM64GZ_SHA=$(sha256sum dists/stable/main/binary-arm64/Packages.gz | cut -d' ' -f1)
ARM64GZ_SIZE=$(stat -c%s dists/stable/main/binary-arm64/Packages.gz)

cat > dists/stable/Release << EOF
Origin: Onyx Repository
Label: Onyx
Suite: stable
Codename: stable
Architectures: amd64 arm64
Components: main
Description: Modular reverse proxy with a live dashboard
Date: $(date -uR)
SHA256:
 ${AMD64_SHA} ${AMD64_SIZE} main/binary-amd64/Packages
 ${AMD64GZ_SHA} ${AMD64GZ_SIZE} main/binary-amd64/Packages.gz
 ${ARM64_SHA} ${ARM64_SIZE} main/binary-arm64/Packages
 ${ARM64GZ_SHA} ${ARM64GZ_SIZE} main/binary-arm64/Packages.gz
EOF

info "Signing Release..."
rm -f dists/stable/InRelease dists/stable/Release.gpg
gpg --clearsign -o dists/stable/InRelease dists/stable/Release
gpg -abs -o dists/stable/Release.gpg dists/stable/Release
ok "Signed"

cd ..
ok "apt repository updated for v${ONYX_VERSION}"
