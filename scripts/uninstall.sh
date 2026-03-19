#!/usr/bin/env bash
# =============================================================================
#  Onyx — Uninstall Script
# =============================================================================
set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
ok()   { echo -e "${GREEN}✓${NC} $*"; }
warn() { echo -e "${YELLOW}⚠${NC} $*"; }

[[ $EUID -eq 0 ]] || { echo -e "${RED}✗${NC} Run as root: sudo bash uninstall.sh" >&2; exit 1; }

echo ""
echo "  Uninstalling Onyx…"
echo ""

if command -v systemctl &>/dev/null; then
  systemctl is-active --quiet onyx 2>/dev/null && systemctl stop onyx && ok "Service stopped."
  systemctl is-enabled --quiet onyx 2>/dev/null && systemctl disable onyx && ok "Service disabled."
  [[ -f /etc/systemd/system/onyx.service ]] && rm /etc/systemd/system/onyx.service && systemctl daemon-reload && ok "Service file removed."
fi

[[ -f /usr/local/bin/onyx ]] && rm /usr/local/bin/onyx && ok "Binary removed."

echo ""
warn "Config and data were NOT removed:"
warn "  /etc/onyx/  and  ~/.config/onyx/"
warn "To fully purge: rm -rf /etc/onyx ~/.config/onyx"
echo ""
echo "  Onyx uninstalled."
echo ""
