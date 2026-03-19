# Configuration Reference

Onyx reads `onyx.toml` from one of these locations (in order):

1. `./onyx.toml`
2. `~/.config/onyx/onyx.toml`
3. `/etc/onyx/onyx.toml`

Override with `onyx start --config /path/to/onyx.toml`.

---

## Full example

```toml
[server]
http_port  = 80
https_port = 443
data_dir   = "~/.config/onyx"

[dashboard]
enabled = true
port    = 8080

[[routes]]
host    = "api.example.com"
target  = "http://localhost:3000"
enabled = true

  [routes.rate_limit]
  requests_per_second = 100.0
  burst               = 50

[[routes]]
host    = "app.example.com"
target  = "http://localhost:5173"
enabled = false
```

---

## [server]

| Key | Type | Default | Description |
|---|---|---|---|
| `http_port` | int | `80` | Port for the HTTP proxy |
| `https_port` | int | `443` | Port for HTTPS (v0.2.0) |
| `data_dir` | string | `~/.config/onyx` | Database and config storage |

## [dashboard]

| Key | Type | Default | Description |
|---|---|---|---|
| `enabled` | bool | `true` | Enable the dashboard server |
| `port` | int | `8080` | Dashboard HTTP port |

The dashboard password is **never** stored in this file.
Set it with `onyx setup`.

## [[routes]]

| Key | Type | Description |
|---|---|---|
| `host` | string | Hostname to match (required) |
| `target` | string | Backend URL (required) |
| `enabled` | bool | Set `false` to disable without removing |

### [[routes.rate_limit]]

| Key | Type | Description |
|---|---|---|
| `requests_per_second` | float | Token bucket refill rate |
| `burst` | int | Maximum burst size |

Set to 0 to disable rate limiting for a route.
