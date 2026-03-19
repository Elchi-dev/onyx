#!/usr/bin/env bash
# Called by the package manager before removing the package.
set -e
if command -v systemctl &>/dev/null; then
  systemctl is-active --quiet onyx 2>/dev/null && systemctl stop onyx || true
  systemctl disable onyx 2>/dev/null || true
fi
