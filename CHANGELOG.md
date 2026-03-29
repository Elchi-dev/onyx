# Changelog

All notable changes to Onyx are documented here.
Format: [Keep a Changelog](https://keepachangelog.com/) · Versioning: [SemVer](https://semver.org/)

---

## [Unreleased]

---

## [v0.1.2] — 2025

### Fixed

- **Dashboard route management** — routes added, deleted, or toggled via the
  dashboard UI now take effect in the live proxy router immediately, without
  requiring a restart. Previously they were saved to the database but the
  running proxy was never updated.
- **Double logging on startup** — `onyx start` was instantiating the full App
  (including database and all routes) twice during the setup check. Replaced
  with a lightweight file-existence check so routes are only logged once.
- **`html.go` build error** — JavaScript template literals (backticks) inside
  Go raw string constants were terminating the string early, causing a cascade
  of parse errors. Rewrote all JS to use string concatenation.
- **`middleware/ratelimit.go` unused import** — removed unused `"net/http"` import.
- **Deprecated `ssh/terminal`** — replaced `golang.org/x/crypto/ssh/terminal`
  (removed in newer x/crypto) with `golang.org/x/term`.
- **`cli/validate.go` undefined `findConfig`** — `findConfig` was defined in
  `internal/app` but called from `internal/cli`, a different package. Added a
  local `findConfigPath()` to `validate.go` to avoid the cross-package call.
- **VCS stamping error on fresh clone** — added `-buildvcs=false` to all
  Makefile build targets so the build never fails before `git init` is run.

### Added

- **`onyx update`** — new command that checks the GitHub releases API for a
  newer version and replaces the current binary atomically. Supports
  `--check` (check only, no install) and `--force` (re-install same version).
- **`scripts/dev/release.sh`** — developer release script. Runs tests,
  cross-compiles all 5 platform binaries, builds `.deb` and `.rpm`, generates
  `checksums.txt`, opens the editor for the changelog, commits, tags, and
  pushes. GitHub Actions publishes the actual GitHub Release on tag push.
- **`scripts/dev/swap.sh`** — quickly replace project files from a zip without
  touching `scripts/dev/`, `.git/`, `build/`, `go.sum`, or `.vscode/`. Shows
  a dry-run diff before applying and runs `go mod tidy` + `make build` after.

---

## [v0.1.1] — 2025

### Security

- Login rate limiting (max 5 attempts/min per IP)
- WebSocket origin validation
- Request body size limit (10 MiB default)
- Session persistence in SQLite (survives restarts)
- Backend connection + response timeouts
- TCP connection pooling
- Friendly port-permission error messages

### Dashboard

- Route management UI (add, enable/disable, delete from browser)
- Server-side persistent statistics (survive page refresh)
- Per-route analytics with request/error/latency breakdown
- Error spike alert bar
- "Remember me" login option (7-day session)
- Logout button
- Favicon
- Mobile responsive layout
- WebSocket auto-reconnect
- `routes_changed` event syncs all open tabs

### CLI

- `onyx validate` — validate config without starting

### Packaging

- AUR `PKGBUILD` + `onyx.install`
- Full install table in README (9 methods)

---

## [v0.1.0-early] — 2025

First early-access release.

### Added

- Reverse proxy core (host-based HTTP routing)
- Live WebSocket dashboard
- Dashboard auth (bcrypt + sessions)
- Interactive setup wizard
- SQLite storage
- TOML configuration
- Per-IP rate limiting (token bucket)
- Weighted traffic splitting
- Plugin API
- Graceful shutdown
- Universal install script
- Systemd service

---

## Roadmap

| Version | Focus |
|---|---|
| v0.2.0 | Automatic HTTPS via Let's Encrypt (ACME) |
| v0.3.0 | Config hot-reload, Homebrew tap, Docker image |
| v0.4.0 | Response caching, compression, webhook notifications |
| v1.0.0 | Stable release |
