# Changelog

All notable changes to Onyx are documented here.
Format: [Keep a Changelog](https://keepachangelog.com/) · Versioning: [SemVer](https://semver.org/)

---

## [Unreleased]

---

## [v0.3.0] — 2026-03-30

### Added

- **Path-based routing** — define `[[routes.paths]]` blocks to route URL prefixes
  to different backends. `/api/` can go to one service, `/ws` to another, all
  under the same hostname. Longest-prefix match, consistent with nginx behavior.
- **Static file serving** — set `static_root` on any route to serve files
  directly from a directory. No separate file server needed.
- **SPA mode** — set `static_spa = true` to fall back to `index.html` for any
  missing path, enabling client-side routing (React, Vue, Svelte, etc.).
- **Explicit WebSocket proxying** — the proxy now correctly forwards `Upgrade`
  and `Connection` headers, with `FlushInterval = -1` for streaming. WebSocket
  and SSE connections work out of the box without any extra configuration.
- **Gzip compression** — set `gzip = true` per route to compress responses.
  Respects `Accept-Encoding`, strips `Content-Length` correctly.
- **Per-route rate limiting** — `[routes.rate_limit]` now works independently
  per route instead of sharing the global limiter.
- **Per-route body size limit** — set `max_body_size` (bytes) per route,
  overriding the global default for that route.
- **Per-route timeouts** — set `timeout` (seconds) per route for
  `ResponseHeaderTimeout`, useful for slow backends like ML inference endpoints.
- **Custom response headers** — set `[routes.headers]` to inject arbitrary
  response headers (CORS, CSP, cache-control, etc.) per route.
- **www redirect** — set `www_redirect = "strip"` to redirect `www.` to the
  bare domain, or `"add"` to do the reverse. 301 redirect, automatic.
- **`onyx import nginx`** — new CLI command. Pass a file or directory
  (`/etc/nginx/sites-available/`) and Onyx parses server blocks, proxy_pass
  targets, location rules, headers, body size limits, and gzip settings,
  converting them to Onyx routes automatically.
- **nginx Import UI** — "Import nginx" button in the Routes view. Paste a
  server block config, preview what will be imported, confirm.
- **Route Edit Modal** — click the edit button on any route in the dashboard
  to update all settings (target, HTTPS, gzip, paths, headers, static root,
  timeouts) without deleting and re-adding the route.
- **Docker support** — `Dockerfile` (multi-stage, `FROM scratch` final image)
  and `docker-compose.yml` example added to the repository.

### Dashboard

- Routes table now shows feature badges per route (HTTPS, Gzip, path count,
  www redirect mode, static directory).
- Add Route form has an expandable Advanced section covering all new fields.
- Edit modal for full in-place route editing.
- nginx Import modal with paste-and-preview flow.

### Internal

- `internal/nginx/` — new package with a from-scratch nginx config tokenizer
  and parser. No external dependencies.
- `internal/proxy/router.go` — fully rewritten. Handler chain is now built
  per-route and composed of: www redirect → body limit → rate limit → gzip →
  response headers → path router → static files / proxy.
- `internal/database/database.go` — 9 new columns added via safe `ALTER TABLE`
  migrations. Paths and response headers stored as JSON.
- `internal/config/config.go` — `RouteConfig` extended with all new fields.

---

## [v0.2.0] — 2026-03-29

### Added

- **Automatic HTTPS via Let's Encrypt (ACME)** — enable HTTPS per route with
  `https = true` in `onyx.toml` or via the dashboard toggle. Public domains get
  a real certificate from Let's Encrypt automatically; local domains (`.test`,
  `.local`, `localhost`, IPs) get a self-signed certificate generated on first
  start. Certificates are stored in `~/.config/onyx/certs/` and renewed
  automatically before expiry.
- **`internal/tls` package** — new certificate manager handling ACME via
  `autocert`, self-signed cert generation, dynamic host registration, and cert
  status reporting.
- **HTTP → HTTPS redirect** — when HTTPS is active, port 80 automatically
  redirects all traffic to HTTPS and serves ACME HTTP-01 challenges.
- **Certificates view** — new dashboard page showing TLS status per route
  (valid / expiring soon / pending / error), certificate mode (Let's Encrypt
  vs self-signed), and expiry date.

### Dashboard

- **Complete UI redesign** — new modern look with refined typography, spacing,
  and color palette. Smooth view-transition animations throughout.
- **Overview page** — new landing page with stat cards, a live requests/minute
  sparkline chart (Chart.js), and a recent traffic feed.
- **Live Traffic page** — dedicated full-page feed with filter bar (by host,
  method, status code) and a pause/resume button.
- **Analytics page** — requests-by-route and error-rate bar charts (Chart.js),
  plus per-route breakdown cards with traffic bars.
- **Settings page** — change dashboard password directly from the browser
  without touching the terminal.
- **Toast notifications** — all actions (add route, delete, toggle, password
  change) now show slide-in toast messages instead of alert bars.
- **HTTPS toggle per route** — enable or disable HTTPS on any route directly
  from the Routes table.

### API

- `POST /api/routes` now accepts an `https` boolean field.
- `PATCH /api/routes/{host}` now accepts an `https` boolean field.
- `GET /api/certs` — new endpoint returning TLS certificate status for all
  HTTPS-enabled routes.
- `POST /api/settings/password` — new endpoint for in-dashboard password
  changes.
- `GET /api/stats` now includes `req_per_min` (average requests per minute
  since startup).

---

## [v0.1.2] — 2026-03-29

### Fixed

- **Live route wiring** — routes added or removed via the dashboard are now
  immediately active in the running proxy without a restart. Previously the
  router was never notified of dashboard changes; a new `RouteManager` interface
  and `SetEventHandler()` method wire the dashboard directly to the proxy router.
- **Double startup logging** — routes were logged twice on start because
  `needsSetup()` in `start.go` instantiated the full app just to check for a
  config file. Replaced with a simple file-existence check.
- **`setcap` preserved on deploy** — `scripts/dev/deploy-pi.sh` now runs
  `sudo setcap cap_net_bind_service=+ep` automatically after copying the binary,
  so port 80 binding survives every redeploy without a manual step.

### Added

- **`onyx update`** — new CLI command that checks the GitHub Releases API for a
  newer version, downloads the correct binary for the current OS and
  architecture, and replaces the running binary atomically via `os.Rename()`.
  Supports `--check` (print latest version without updating) and `--force`
  (re-install even if already on the latest version).

### Developer

- **`scripts/dev/swap.sh`** — replaces project files from a zip while
  preserving `scripts/dev/`, `.git/`, `build/`, `go.sum`, and `.vscode/`.
  Accepts the zip path as an argument, shows a diff preview, asks for
  confirmation, then runs `go mod tidy` and `make build` automatically.
- **`scripts/dev/release.sh`** — local release helper. Runs the test suite,
  stamps `CHANGELOG.md` with the version and date, opens `$EDITOR` for release
  notes, commits, tags, and pushes. GitHub Actions does the rest.

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
