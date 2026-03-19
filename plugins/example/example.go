// Package example is a reference Onyx plugin.
// Copy this package to get started writing your own plugin.
//
// To use a plugin, register it before starting the app:
//
//	registry := plugin.NewRegistry()
//	registry.Register(&example.Plugin{})
//	handler := registry.Chain(proxyRouter)
package example

import "net/http"

// Plugin adds an X-Onyx-Plugin header to every proxied response.
type Plugin struct{}

// Name returns the plugin's unique identifier.
func (p *Plugin) Name() string { return "example" }

// Version returns the plugin's semantic version string.
func (p *Plugin) Version() string { return "0.1.0" }

// Handler returns middleware that adds a custom response header.
func (p *Plugin) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Runs before the request is forwarded to the backend.
			w.Header().Set("X-Onyx-Plugin", p.Name()+"/"+p.Version())

			next.ServeHTTP(w, r)

			// Post-request code would go here.
		})
	}
}
