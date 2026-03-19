// Package plugin defines the Onyx plugin API.
//
// A plugin is any type that implements the Plugin interface.
// Plugins are composable middleware — they wrap the proxy handler and can
// inspect or modify requests and responses.
//
// Implementing a plugin:
//
//	type MyPlugin struct{}
//	func (p *MyPlugin) Name()    string { return "my-plugin" }
//	func (p *MyPlugin) Version() string { return "1.0.0" }
//	func (p *MyPlugin) Handler() func(http.Handler) http.Handler {
//	    return func(next http.Handler) http.Handler {
//	        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	            // pre-request logic
//	            next.ServeHTTP(w, r)
//	            // post-request logic
//	        })
//	    }
//	}
package plugin

import "net/http"

// Plugin is the interface every Onyx plugin must implement.
type Plugin interface {
	// Name returns a unique lowercase identifier for the plugin.
	Name() string
	// Version returns the plugin's semantic version string (e.g. "1.0.0").
	Version() string
	// Handler returns middleware that wraps the given next handler.
	Handler() func(http.Handler) http.Handler
}

// Registry holds registered plugins and applies them as a middleware chain.
type Registry struct {
	plugins []Plugin
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry { return &Registry{} }

// Register adds a plugin to the registry.
// Plugins are applied in registration order — first registered is outermost.
func (r *Registry) Register(p Plugin) {
	r.plugins = append(r.plugins, p)
}

// Chain wraps h with all registered plugin middlewares.
// Execution order for plugins [A, B, C]:
//
//	Request  →  A  →  B  →  C  →  h
//	Response ←  A  ←  B  ←  C  ←  h
func (r *Registry) Chain(h http.Handler) http.Handler {
	for i := len(r.plugins) - 1; i >= 0; i-- {
		h = r.plugins[i].Handler()(h)
	}
	return h
}

// List returns "name@version" strings for every registered plugin.
func (r *Registry) List() []string {
	out := make([]string, len(r.plugins))
	for i, p := range r.plugins {
		out[i] = p.Name() + "@" + p.Version()
	}
	return out
}

// Len returns the number of registered plugins.
func (r *Registry) Len() int { return len(r.plugins) }
