# Plugin API

Onyx middleware plugins wrap the proxy router and can inspect or modify
requests and responses. The plugin API is stable from v0.1.0 onward.

---

## The Plugin interface

```go
// import "github.com/Elchi-dev/onyx/internal/plugin"

type Plugin interface {
    Name()    string                               // unique lowercase id
    Version() string                               // semver e.g. "1.0.0"
    Handler() func(http.Handler) http.Handler      // middleware function
}
```

---

## Quickstart

Copy `plugins/example/` and adapt it:

```go
package myplugin

import "net/http"

type Plugin struct{}

func (p *Plugin) Name()    string { return "my-plugin" }
func (p *Plugin) Version() string { return "0.1.0" }

func (p *Plugin) Handler() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Runs BEFORE the request is forwarded to the backend.
            w.Header().Set("X-My-Plugin", "hello")

            next.ServeHTTP(w, r)

            // Runs AFTER the backend responds.
        })
    }
}
```

---

## Registering a plugin

Plugins are currently registered in Go code. In a future release Onyx will
support loading plugins from a directory automatically.

```go
registry := plugin.NewRegistry()
registry.Register(&myplugin.Plugin{})

// Wrap the proxy router:
handler := registry.Chain(proxyRouter)
```

---

## Execution order

`Chain(router, A, B, C)` results in:

```
Request  →  A  →  B  →  C  →  router  →  backend
Response ←  A  ←  B  ←  C  ←  router  ←  backend
```

---

## Built-in middleware

These live in `internal/middleware/` and are applied automatically:

| Middleware | What it does |
|---|---|
| `Recovery` | Catches panics, returns 500 |
| `RequestLogger` | Logs method/path/status/latency |
| `SecureHeaders` | Adds X-Content-Type-Options etc. |
| `RateLimit` | Enforces per-IP token bucket |
| `CORS` | Adds CORS headers |

You can use these in your own plugins too — they're just functions.

---

## Plugin ideas

- Request ID injection (`X-Request-ID`)
- Bearer token validation
- IP allowlist / denylist
- Gzip response compression
- Cache-Control header injection
- Geo-blocking
