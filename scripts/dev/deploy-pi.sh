#!/usr/bin/env bash
set -euo pipefail
PI="sami@192.168.5.2"
GREEN='\033[0;32m'; CYAN='\033[0;36m'; NC='\033[0m'
echo -e "${CYAN}→${NC} Cross-compiling for ARM64..."
GOOS=linux GOARCH=arm64 go build \
  -ldflags "-X main.version=v0.1.1-local -s -w" \
  -buildvcs=false \
  -o /tmp/onyx-arm64 ./cmd/onyx
echo -e "${GREEN}✓${NC} Built: $(du -sh /tmp/onyx-arm64 | cut -f1)"
echo -e "${CYAN}→${NC} Deploying to Pi..."
ssh "$PI" "mkdir -p ~/onyx"
scp /tmp/onyx-arm64 "$PI:~/onyx/onyx"
ssh "$PI" "chmod +x ~/onyx/onyx && sudo setcap cap_net_bind_service=+ep ~/onyx/onyx"
echo -e "${GREEN}✓${NC} Deployed to $PI:~/onyx/onyx (cap_net_bind_service granted)"
