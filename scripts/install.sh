#!/usr/bin/env bash
# =============================================================================
#  Onyx — Universal Installer
#  Usage:  curl -fsSL https://elchi-dev.github.io/onyx/install.sh | bash
#  Or:     sudo bash scripts/install.sh
#
#  Supports: Ubuntu 20.04+, Debian 11+, Arch Linux, macOS
# =============================================================================
set -euo pipefail

ONYX_VERSION="${ONYX_VERSION:-v0.1.0-early}"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/onyx"
SYSTEMD_DIR="/etc/systemd/system"
REPO="Elchi-dev/onyx"

# ── Colours ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'
info()    { echo -e "${CYAN}→${NC} $*"; }
ok()      { echo -e "${GREEN}✓${NC} $*"; }
warn()    { echo -e "${YELLOW}⚠${NC} $*"; }
die()     { echo -e "${RED}✗${NC} $*" >&2; exit 1; }

# ── Banner ───────────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}  Onyx Installer${NC} — ${ONYX_VERSION}"
echo "  Modular Reverse Proxy"
echo ""

# ── Root check ───────────────────────────────────────────────────────────────
[[ $EUID -eq 0 ]] || die "Run as root: sudo bash install.sh"

# ── Detect OS ────────────────────────────────────────────────────────────────
OS_ID=""
if [[ -f /etc/os-release ]]; then
  . /etc/os-release
  OS_ID="$ID"
  info "OS: $PRETTY_NAME"
elif [[ "$(uname)" == "Darwin" ]]; then
  OS_ID="macos"
  info "OS: macOS $(sw_vers -productVersion)"
else
  die "Unsupported OS. Build from source: https://github.com/${REPO}"
fi

# ── Detect architecture ───────────────────────────────────────────────────────
ARCH_SUFFIX=""
case "$(uname -m)" in
  x86_64)  ARCH_SUFFIX="amd64" ;;
  aarch64|arm64) ARCH_SUFFIX="arm64" ;;
  *) die "Unsupported architecture: $(uname -m)" ;;
esac
info "Arch: $(uname -m) → ${ARCH_SUFFIX}"

# ── Install dependencies ──────────────────────────────────────────────────────
case "$OS_ID" in
  ubuntu|debian)
    apt-get update -qq
    apt-get install -y -qq curl ca-certificates
    ;;
  arch|manjaro)
    pacman -Sy --noconfirm --needed curl ca-certificates 2>/dev/null || true
    ;;
  macos)
    command -v curl >/dev/null || die "curl not found. Install Xcode command line tools."
    ;;
esac

# ── Install binary ────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GOOS="linux"
[[ "$OS_ID" == "macos" ]] && GOOS="darwin"

LOCAL_BIN="$SCRIPT_DIR/../build/onyx-${GOOS}-${ARCH_SUFFIX}"

if [[ -f "$LOCAL_BIN" ]]; then
  info "Installing local build..."
  cp "$LOCAL_BIN" "${INSTALL_DIR}/onyx"
else
  info "Downloading Onyx ${ONYX_VERSION}..."
  ASSET="onyx-${GOOS}-${ARCH_SUFFIX}"
  URL="https://github.com/${REPO}/releases/download/${ONYX_VERSION}/${ASSET}"
  if ! curl -fsSL "$URL" -o "${INSTALL_DIR}/onyx"; then
    die "Download failed. Build from source:\n  git clone https://github.com/${REPO}\n  cd onyx && make build"
  fi
fi

chmod +x "${INSTALL_DIR}/onyx"
ok "Binary installed: ${INSTALL_DIR}/onyx ($(onyx --version 2>/dev/null || echo 'version unknown'))"

# ── Create config directory ───────────────────────────────────────────────────
mkdir -p "$CONFIG_DIR"
chmod 750 "$CONFIG_DIR"
ok "Config directory: ${CONFIG_DIR}"

# ── Systemd service (Linux only) ──────────────────────────────────────────────
if command -v systemctl &>/dev/null && [[ "$OS_ID" != "macos" ]]; then
  # Create dedicated system user.
  if ! id -u onyx &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin onyx
    ok "System user 'onyx' created."
  fi
  chown -R onyx:onyx "$CONFIG_DIR"

  # Install service file.
  SERVICE_SRC="$SCRIPT_DIR/../onyx.service"
  if [[ -f "$SERVICE_SRC" ]]; then
    cp "$SERVICE_SRC" "${SYSTEMD_DIR}/onyx.service"
  else
    cat > "${SYSTEMD_DIR}/onyx.service" << 'SVCEOF'
[Unit]
Description=Onyx Reverse Proxy
After=network-online.target
Wants=network-online.target
Documentation=https://github.com/Elchi-dev/onyx

[Service]
Type=simple
User=onyx
Group=onyx
ExecStart=/usr/local/bin/onyx start --config /etc/onyx/onyx.toml
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure
RestartSec=5s
TimeoutStopSec=30s
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/etc/onyx
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
SVCEOF
  fi

  systemctl daemon-reload
  systemctl enable onyx
  ok "Systemd service installed and enabled."
fi

# ── First-time setup ──────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}  Running first-time setup…${NC}"
echo "  ────────────────────────────────────"
"${INSTALL_DIR}/onyx" setup

# ── Done ──────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}${BOLD}  Onyx ${ONYX_VERSION} is installed!${NC}"
echo ""
echo "  Start now:        ${BOLD}onyx start${NC}"
if command -v systemctl &>/dev/null; then
echo "  Start as service: ${BOLD}sudo systemctl start onyx${NC}"
echo "  View logs:        ${BOLD}journalctl -u onyx -f${NC}"
fi
echo "  Add a route:      ${BOLD}onyx route add --host example.com --target http://localhost:3000${NC}"
echo "  Dashboard:        ${BOLD}http://localhost:8080${NC}"
echo ""
