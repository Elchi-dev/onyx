<div align="center">

```
 ██████╗ ███╗   ██╗██╗   ██╗██╗  ██╗
██╔═══██╗████╗  ██║╚██╗ ██╔╝╚██╗██╔╝
██║   ██║██╔██╗ ██║ ╚████╔╝  ╚███╔╝
██║   ██║██║╚██╗██║  ╚██╔╝   ██╔██╗
╚██████╔╝██║ ╚████║   ██║   ██╔╝ ██╗
 ╚═════╝ ╚═╝  ╚═══╝   ╚═╝   ╚═╝  ╚═╝
```

**Modular reverse proxy with a live WebSocket dashboard.**

[![CI](https://github.com/Elchi-dev/onyx/actions/workflows/ci.yml/badge.svg)](https://github.com/Elchi-dev/onyx/actions/workflows/ci.yml)
[![Go 1.24+](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/Elchi-dev/onyx?include_prereleases)](https://github.com/Elchi-dev/onyx/releases)

</div>

---

## Install

### Ubuntu / Debian / Pop!_OS / Linux Mint

```bash
# Add the Onyx apt repository (one-time setup)
curl -fsSL https://elchi-dev.github.io/onyx/gpg.key \
  | sudo gpg --dearmor -o /usr/share/keyrings/onyx.gpg

echo "deb [signed-by=/usr/share/keyrings/onyx.gpg] \
  https://elchi-dev.github.io/onyx/apt stable main" \
  | sudo tee /etc/apt/sources.list.d/onyx.list

sudo apt update && sudo apt install onyx
```

Future updates: `sudo apt upgrade onyx`

---

### Arch Linux / Manjaro / EndeavourOS

```bash
# Using yay
yay -S onyx

# Using paru
paru -S onyx

# Manual AUR install
git clone https://aur.archlinux.org/onyx.git
cd onyx && makepkg -si
```

---

### Fedora / RHEL / CentOS / AlmaLinux / Rocky Linux

```bash
# Add the Onyx rpm repository
sudo curl -fsSL https://elchi-dev.github.io/onyx/onyx.repo \
  -o /etc/yum.repos.d/onyx.repo

sudo dnf install onyx
```

---

### openSUSE

```bash
sudo zypper addrepo https://elchi-dev.github.io/onyx/onyx.repo onyx
sudo zypper install onyx
```

---

### macOS (Homebrew)

```bash
# Add the Onyx tap (one-time)
brew tap Elchi-dev/onyx

brew install onyx
```

Future updates: `brew upgrade onyx`

---

### Any Linux / macOS — curl installer

```bash
curl -fsSL https://elchi-dev.github.io/onyx/install.sh | bash
```

Detects your OS and architecture automatically. Works on any system with `curl`.

---

### Docker

```bash
docker run -d \
  -p 80:80 -p 8080:8080 \
  -v ~/.config/onyx:/data \
  ghcr.io/elchi-dev/onyx:latest
```

---

### Build from source

Requires Go 1.24+.

```bash
git clone https://github.com/Elchi-dev/onyx.git
cd onyx
go mod tidy
make build          # binary at ./build/onyx
sudo make install   # copies to /usr/local/bin/onyx
```

---

### Download binary directly

Download the pre-built binary for your platform from the
[Releases page](https://github.com/Elchi-dev/onyx/releases):

| Platform | File |
|---|---|
| Linux x86_64 | `onyx-linux-amd64` |
| Linux ARM64 | `onyx-linux-arm64` |
| macOS x86_64 | `onyx-darwin-amd64` |
| macOS Apple Silicon | `onyx-darwin-arm64` |
| Windows x86_64 | `onyx-windows-amd64.exe` |

```bash
# Example for Linux amd64
curl -fsSL https://github.com/Elchi-dev/onyx/releases/latest/download/onyx-linux-amd64 \
  -o /usr/local/bin/onyx
chmod +x /usr/local/bin/onyx
```

---

## Quick start

```bash
# Interactive first-time setup (creates config, sets password, adds first route)
onyx setup

# Start proxy + live dashboard
onyx start

# Add a route
onyx route add --host api.example.com --target http://localhost:3000

# Open the dashboard in your browser
open http://localhost:8080
```

---

## Features

| Feature | Status |
|---|---|
| 🔀 Host-based reverse proxy | ✅ v0.1.0 |
| 📊 Live WebSocket dashboard | ✅ v0.1.0 |
| 🔒 Dashboard auth (bcrypt + sessions) | ✅ v0.1.0 |
| 🧙 Interactive setup wizard | ✅ v0.1.0 |
| 🗄️ SQLite storage (zero dependencies) | ✅ v0.1.0 |
| 📝 TOML configuration | ✅ v0.1.0 |
| 🚦 Per-IP rate limiting (token bucket) | ✅ v0.1.0 |
| ✂️ Weighted traffic splitting | ✅ v0.1.0 |
| 🔌 Plugin API (composable middleware) | ✅ v0.1.0 |
| 🔁 Graceful shutdown | ✅ v0.1.0 |
| 🛣️ Route management from dashboard UI | ✅ v0.1.1 |
| 📈 Server-side persistent stats | ✅ v0.1.1 |
| 📊 Per-route breakdown & analytics | ✅ v0.1.1 |
| 🔐 Login rate limiting (brute-force protection) | ✅ v0.1.1 |
| 🔐 WebSocket origin validation | ✅ v0.1.1 |
| 🔐 Session persistence (survives restart) | ✅ v0.1.1 |
| 🔐 Request body size limit | ✅ v0.1.1 |
| ⏱️ Backend connection + response timeouts | ✅ v0.1.1 |
| ♻️ TCP connection pooling to backends | ✅ v0.1.1 |
| ✅ `onyx validate` — config check | ✅ v0.1.1 |
| 📦 apt / .deb packaging | ✅ v0.1.1 |
| 📦 rpm packaging | ✅ v0.1.1 |
| 📦 AUR (Arch) package | ✅ v0.1.1 |
| 📱 Responsive mobile layout | ✅ v0.1.1 |
| 🔄 Live route updates (no restart needed) | ✅ v0.1.2 |
| ⬆️ Self-update via `onyx update` | ✅ v0.1.2 |
| 🔒 Automatic HTTPS (Let's Encrypt) | ✅ v0.2.0 |
| 🔒 Self-signed certs for local domains | ✅ v0.2.0 |
| 🛣️ Path-based routing | ✅ v0.3.0 |
| 📁 Static file serving + SPA mode | ✅ v0.3.0 |
| 🔌 Explicit WebSocket proxying | ✅ v0.3.0 |
| 🗜️ Gzip compression per route | ✅ v0.3.0 |
| 📋 Custom response headers per route | ✅ v0.3.0 |
| 🔁 www redirect (strip/add) | ✅ v0.3.0 |
| 📥 nginx config importer (CLI + dashboard) | ✅ v0.3.0 |
| ✏️ Route edit modal in dashboard | ✅ v0.3.0 |
| 🐳 Docker support | ✅ v0.3.0 |
| 🏥 Health checks + auto-failover | 📅 v0.3.2 |
| 🔐 Basic auth per route | 📅 v0.3.3 |
| 📊 Prometheus metrics | 📅 v0.4.0 |
| ♻️ Config hot-reload | 📅 v0.4.0 |
| 📦 Homebrew tap | 📅 v0.4.0 |

---

## CLI reference

```
onyx start [--config PATH] [--dev]       Start proxy + dashboard
onyx setup                               Interactive first-time setup
onyx validate [--config PATH]            Check config without starting
onyx status [--url URL]                  Check if Onyx is running
onyx update [--check] [--force]          Update Onyx to the latest release
onyx route add --host H --target T       Add a proxy route
onyx route list                          List all routes
onyx route remove <host>                 Remove a route
onyx import nginx <file|dir>             Import routes from nginx config
```

---

## Configuration

```toml
[server]
http_port = 80
data_dir  = "~/.config/onyx"

[dashboard]
enabled = true
port    = 8080

[[routes]]
host    = "api.example.com"
target  = "http://localhost:3000"
enabled = true

  [routes.rate_limit]
  requests_per_second = 100
  burst               = 50
```

Full reference: [docs/configuration.md](docs/configuration.md)

---

## Plugin API

```go
type Plugin interface {
    Name()    string
    Version() string
    Handler() func(http.Handler) http.Handler
}
```

See [`plugins/example/`](plugins/example/) and [docs/plugins.md](docs/plugins.md).

---

## Architecture

```
internal/app/         Dependency wiring — connects all packages
internal/middleware/  Composable middleware: Recovery, BodyLimit, Logger, Headers, RateLimit
internal/proxy/       Router (with connection pooling) + server + request events
internal/dashboard/   WebSocket hub + HTTP server + login limiter + session cleanup
internal/auth/        Password hashing + session tokens
internal/database/    SQLite: settings, routes, sessions, route_stats
internal/config/      TOML loader + validation
internal/ratelimit/   Per-IP token bucket rate limiter
internal/traffic/     Weighted traffic splitter
internal/plugin/      Plugin registry + middleware chain
internal/wizard/      Interactive CLI setup wizard
plugins/example/      Reference plugin — copy to build your own
docs/                 User documentation
scripts/              install.sh, uninstall.sh, pkg/, dev/
```

### Middleware stack

Every proxied request passes through (outermost first):

```
Recovery → BodyLimit → RequestLogger → SecureHeaders → RateLimit → Router → Backend
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) and [docs/plugins.md](docs/plugins.md).

---

## License

MIT — [LICENSE](./LICENSE) · built by [Elchi-dev](https://github.com/Elchi-dev)
