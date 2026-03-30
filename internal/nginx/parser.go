// Package nginx parses nginx configuration files and converts them to Onyx routes.
package nginx

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/Elchi-dev/onyx/internal/database"
)

// ParseFile parses a single nginx config file and returns Onyx routes.
func ParseFile(path string) ([]database.Route, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return ParseConfig(string(data))
}

// ParseDir parses all files in a directory (nginx sites-available style).
func ParseDir(dir string) ([]database.Route, []error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, []error{fmt.Errorf("reading directory %s: %w", dir, err)}
	}
	var routes []database.Route
	var errs []error
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		path := filepath.Join(dir, e.Name())
		r, err := ParseFile(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", e.Name(), err))
			continue
		}
		routes = append(routes, r...)
	}
	return routes, errs
}

// ParseConfig parses nginx config text and returns Onyx routes.
func ParseConfig(input string) ([]database.Route, error) {
	tokens := tokenize(input)
	blocks, err := parseBlocks(tokens)
	if err != nil {
		return nil, err
	}
	var routes []database.Route
	for _, b := range blocks {
		if b.name != "server" {
			continue
		}
		r, ok := blockToRoute(b)
		if !ok {
			continue
		}
		routes = append(routes, r)
	}
	return routes, nil
}

// ── AST ───────────────────────────────────────────────────────────────────────

type block struct {
	name       string
	args       []string
	directives []directive
	children   []block
}

type directive struct {
	name string
	args []string
}

// ── Tokenizer ─────────────────────────────────────────────────────────────────

type token struct {
	kind  string // "word", "{", "}", ";"
	value string
}

func tokenize(input string) []token {
	var tokens []token
	i := 0
	for i < len(input) {
		c := rune(input[i])
		// Skip comments.
		if c == '#' {
			for i < len(input) && input[i] != '\n' {
				i++
			}
			continue
		}
		// Skip whitespace.
		if unicode.IsSpace(c) {
			i++
			continue
		}
		if c == '{' {
			tokens = append(tokens, token{"{"," {"})
			i++
			continue
		}
		if c == '}' {
			tokens = append(tokens, token{"}", "}"})
			i++
			continue
		}
		if c == ';' {
			tokens = append(tokens, token{";", ";"})
			i++
			continue
		}
		// Quoted string.
		if c == '"' || c == '\'' {
			quote := input[i]
			i++
			start := i
			for i < len(input) && input[i] != quote {
				i++
			}
			tokens = append(tokens, token{"word", input[start:i]})
			if i < len(input) {
				i++ // closing quote
			}
			continue
		}
		// Word (everything else until whitespace or special char).
		start := i
		for i < len(input) {
			ch := input[i]
			if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' ||
				ch == '{' || ch == '}' || ch == ';' || ch == '#' {
				break
			}
			i++
		}
		if i > start {
			tokens = append(tokens, token{"word", input[start:i]})
		}
	}
	return tokens
}

// ── Parser ────────────────────────────────────────────────────────────────────

func parseBlocks(tokens []token) ([]block, error) {
	blocks, _, err := parseBlockList(tokens, 0)
	return blocks, err
}

func parseBlockList(tokens []token, start int) ([]block, int, error) {
	var blocks []block
	i := start
	for i < len(tokens) {
		if tokens[i].kind == "}" {
			break
		}
		if tokens[i].kind != "word" {
			i++
			continue
		}
		name := tokens[i].value
		i++
		// Collect args until { or ;
		var args []string
		for i < len(tokens) && tokens[i].kind == "word" {
			args = append(args, tokens[i].value)
			i++
		}
		if i >= len(tokens) {
			break
		}
		if tokens[i].kind == ";" {
			// Directive.
			blocks = append(blocks, block{name: name, args: args})
			i++
		} else if tokens[i].kind == "{" {
			// Block.
			i++
			children, next, err := parseBlockList(tokens, i)
			if err != nil {
				return nil, 0, err
			}
			i = next
			if i < len(tokens) && tokens[i].kind == "}" {
				i++
			}
			b := block{name: name, args: args}
			for _, child := range children {
				if len(child.children) > 0 {
					b.children = append(b.children, child)
				} else {
					b.directives = append(b.directives, directive{name: child.name, args: child.args})
				}
			}
			blocks = append(blocks, b)
		} else {
			i++
		}
	}
	return blocks, i, nil
}

// ── nginx → Onyx conversion ───────────────────────────────────────────────────

func blockToRoute(b block) (database.Route, bool) {
	var r database.Route
	r.Enabled = true
	r.RespHeaders = map[string]string{}

	for _, d := range b.directives {
		switch d.name {
		case "server_name":
			if len(d.args) > 0 {
				// Use the first non-wildcard name.
				for _, name := range d.args {
					if name != "_" && !strings.HasPrefix(name, "~") {
						r.Host = name
						break
					}
				}
			}
		case "listen":
			for _, arg := range d.args {
				if arg == "443" || strings.Contains(arg, "443") {
					r.HTTPS = true
				}
				if arg == "ssl" {
					r.HTTPS = true
				}
			}
		case "add_header":
			if len(d.args) >= 2 {
				// Skip "always" modifier.
				val := d.args[1]
				if len(d.args) > 2 && d.args[2] != "always" {
					val = strings.Join(d.args[1:len(d.args)-1], " ")
				}
				// Don't import SSL/security headers — Onyx handles those.
				name := d.args[0]
				skip := []string{"Strict-Transport-Security", "X-Frame-Options",
					"X-Content-Type-Options", "X-XSS-Protection",
					"ssl_certificate", "ssl_dhparam"}
				skipIt := false
				for _, s := range skip {
					if strings.EqualFold(name, s) {
						skipIt = true
						break
					}
				}
				if !skipIt {
					r.RespHeaders[name] = val
				}
			}
		case "client_max_body_size":
			if len(d.args) > 0 {
				r.MaxBodySize = parseSize(d.args[0])
			}
		case "gzip":
			if len(d.args) > 0 && d.args[0] == "on" {
				r.Gzip = true
			}
		}
	}

	// Parse location blocks for path rules and default proxy.
	for _, child := range b.children {
		if child.name != "location" {
			continue
		}
		locPath := ""
		if len(child.args) > 0 {
			locPath = child.args[len(child.args)-1]
		}

		for _, d := range child.directives {
			switch d.name {
			case "proxy_pass":
				if len(d.args) > 0 {
					target := d.args[0]
					if locPath == "/" || locPath == "" {
						r.Target = target
					} else {
						r.Paths = append(r.Paths, database.PathEntry{
							Path:   locPath,
							Target: target,
						})
					}
				}
			case "root":
				if len(d.args) > 0 && r.StaticRoot == "" {
					r.StaticRoot = d.args[0]
				}
			case "try_files":
				for _, arg := range d.args {
					if arg == "/index.html" || strings.HasSuffix(arg, ".html") {
						r.StaticSPA = true
					}
				}
			case "return":
				if len(d.args) >= 2 && (d.args[0] == "301" || d.args[0] == "302") {
					// Detect www → non-www pattern.
					dest := d.args[1]
					if strings.Contains(dest, "$host") || strings.Contains(dest, "www.") {
						// Simple heuristic.
						if strings.Contains(dest, "//www.") {
							r.WWWRedirect = "add"
						}
					}
				}
			}
		}
	}

	// Need at least a host and some target.
	if r.Host == "" {
		return r, false
	}
	if r.Target == "" && r.StaticRoot == "" {
		return r, false
	}
	// Skip pure redirect-only server blocks (certbot temp blocks, etc.)
	if r.Target == "" && r.StaticRoot == "" {
		return r, false
	}

	return r, true
}

func parseSize(s string) int64 {
	s = strings.ToLower(strings.TrimSpace(s))
	multipliers := map[string]int64{
		"k": 1024, "kb": 1024,
		"m": 1024 * 1024, "mb": 1024 * 1024,
		"g": 1024 * 1024 * 1024, "gb": 1024 * 1024 * 1024,
	}
	for suffix, mult := range multipliers {
		if strings.HasSuffix(s, suffix) {
			num, err := strconv.ParseFloat(strings.TrimSuffix(s, suffix), 64)
			if err == nil {
				return int64(num * float64(mult))
			}
		}
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
