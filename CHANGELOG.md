# Changelog

All notable changes to Onyx are documented here.
Format: [Keep a Changelog](https://keepachangelog.com/) · Versioning: [SemVer](https://semver.org/)

---

## [Unreleased]

---

## [v0.1.1] — 2025

### Security

- **Login rate limiting** — max 5 attempts per minute per IP on the `/login`
  endpoint. Excess attempts receive a 429 page with a `Retry-After` header.
- **WebSocket origin validation** — the WebSocket upgrader now rejects
  connections from a different origin than the dashboard's own host, preventing
  cross-site WebSocket hijacking.
- **Request body size limit** — incoming proxy requests are capped at 10 MiB by
  default (configurable). Prevents runaway request bodies from holding goroutines.
- **Session persistence** — dashboard sessions are now stored in SQLite instead
  of an in-memory map. Sessions survive server restarts and are cleaned up hourly.
- **Backend connection timeouts** — the proxy transport now enforces a 10-second
  dial timeout, 60-second response header timeout, and 10-second TLS handshake
  timeout. A slow or unresponsive backend can no longer hang a goroutine
  indefinitely.
- **TCP connection pooling** — the proxy reuses backend connections (max 200
  idle, 20 per host, 90-second idle timeout), significantly reducing latency
  and resource usage under load.
- **Friendly port-permission errors** — binding port 80/443 without sufficient
  privileges now prints a clear, actionable error instead of a raw syscall error.

### Dashboard

- **Route management UI** — add, enable/disable, and delete proxy routes
  directly from the browser. No terminal required after initial setup.
- **Server-side persistent statistics** — request counts, error counts, and
  average latency are stored in SQLite and survive page refreshes and restarts.
- **Per-route analytics** — the Statistics view shows a breakdown card per
  route with request count, error count, average latency, and a relative
  traffic bar.
- **Error spike alert** — a red alert bar appears at the top when 5xx errors
  spike significantly.
- **"Remember me" login option** — checking "Keep me signed in" extends the
  session from 24 hours to 7 days.
- **Logout button** — header now includes a sign-out action that deletes the
  session from the database.
- **Favicon** — inline SVG favicon, no external requests.
- **Mobile responsive layout** — the dashboard is usable on phones and tablets.
- **WebSocket auto-reconnect** — dashboard reconnects automatically with
  exponential backoff if the connection drops.
- **`routes_changed` WebSocket event** — all connected dashboard tabs refresh
  their route list automatically when a route is added, removed, or toggled.

### CLI

- **`onyx validate`** — new command that parses and validates the config file
  without starting any servers. Shows parsed ports, routes, and their status.

### Packaging

- **AUR package** (`PKGBUILD` + `onyx.install`) for Arch Linux / Manjaro /
  EndeavourOS. Install with `yay -S onyx` or `paru -S onyx`.
- **README** — full install table covering apt, dnf/rpm, AUR, Homebrew,
  curl installer, Docker, binary download, and build from source.

---

## [v0.1.0-early] — 2025

First early-access release. Core architecture complete and functional.

### Added

- Reverse proxy core — host-based HTTP routing
- Live WebSocket dashboard
- Dashboard auth (bcrypt + session cookies)
- Interactive setup wizard
- SQLite storage (zero external dependencies)
- TOML configuration
- Per-IP rate limiting (token bucket)
- Weighted traffic splitting
- Plugin API (composable middleware)
- Graceful shutdown (SIGTERM / SIGINT)
- `scripts/install.sh` — universal installer
- Systemd service with capability-based port binding
- CI workflow (test + lint)
- Release workflow (cross-compile + GitHub Release + .deb)

---

## Roadmap

| Version | Focus |
|---|---|
| v0.2.0 | Automatic HTTPS via Let's Encrypt (ACME) |
| v0.3.0 | Config hot-reload, Homebrew tap, Docker image |
| v0.4.0 | Response caching, compression, webhook notifications |
| v1.0.0 | Stable release |
