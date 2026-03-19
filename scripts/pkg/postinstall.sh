#!/usr/bin/env bash
# Called by the .deb/.rpm installer after the package is installed.
set -e
# Create onyx system user if it doesn't exist.
if ! id -u onyx &>/dev/null; then
  useradd --system --no-create-home --shell /usr/sbin/nologin onyx
fi
mkdir -p /etc/onyx
chown -R onyx:onyx /etc/onyx
chmod 750 /etc/onyx

if command -v systemctl &>/dev/null; then
  systemctl daemon-reload
  systemctl enable onyx
fi

echo ""
echo "  Onyx installed. Run 'onyx setup' then 'onyx start' (or 'systemctl start onyx')."
echo ""
