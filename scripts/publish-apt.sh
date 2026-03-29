#!/bin/bash
# scripts/publish-apt.sh

set -e

# Pfade definieren
REPO_ROOT=$(pwd)
APT_DIR="$REPO_ROOT/apt"
DISTS_DIR="$APT_DIR/dists/stable"
BINARY_DIR="$DISTS_DIR/main/binary-amd64"

echo "Update Onyx APT Repository..."

# 1. Packages Liste neu generieren
cd "$APT_DIR"
dpkg-scanpackages pool/main/o/onyx/ /dev/null > "$BINARY_DIR/Packages"
gzip -9c "$BINARY_DIR/Packages" > "$BINARY_DIR/Packages.gz"

# 2. Hashes für das Release-File berechnen
MD5=$(md5sum "$BINARY_DIR/Packages" | cut -d' ' -f1)
SHA256=$(sha256sum "$BINARY_DIR/Packages" | cut -d' ' -f1)
SIZE=$(stat -c%s "$BINARY_DIR/Packages")

# 3. Release File schreiben
cat <<EOF > "$DISTS_DIR/Release"
Origin: Onyx Repository
Label: Onyx
Suite: stable
Codename: stable
Architectures: amd64 arm64
Components: main
Description: Modular reverse proxy with a live dashboard
MD5Sum:
 $MD5 $SIZE main/binary-amd64/Packages
SHA256:
 $SHA256 $SIZE main/binary-amd64/Packages
EOF

# 4. Signieren (ersetze 'Samuel Krauß' falls nötig)
rm -f "$DISTS_DIR/InRelease" "$DISTS_DIR/Release.gpg"
gpg --clearsign -o "$DISTS_DIR/InRelease" "$DISTS_DIR/Release"
gpg -abs -o "$DISTS_DIR/Release.gpg" "$DISTS_DIR/Release"

echo "Done! Ready to git push."
