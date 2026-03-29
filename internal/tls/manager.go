// Package tls manages TLS certificates for Onyx.
// Public domains get automatic certificates from Let's Encrypt via ACME.
// Local/private domains get auto-generated self-signed certificates.
package tls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// CertStatus describes the TLS certificate state for one host.
type CertStatus struct {
	Host      string    `json:"host"`
	Mode      string    `json:"mode"`   // "acme", "self_signed"
	Status    string    `json:"status"` // "valid", "expiring_soon", "pending", "error"
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// Manager handles certificate provisioning for all HTTPS-enabled routes.
type Manager struct {
	ac      *autocert.Manager
	log     *slog.Logger
	dataDir string
	hosts   []string
}

// New creates a Manager for the given hosts.
// Certificates are stored under dataDir/certs/.
func New(dataDir string, hosts []string, log *slog.Logger) *Manager {
	certsDir := filepath.Join(dataDir, "certs")
	_ = os.MkdirAll(certsDir, 0o700)

	publicHosts := filterPublicHosts(hosts)

	var ac *autocert.Manager
	if len(publicHosts) > 0 {
		ac = &autocert.Manager{
			Cache:      autocert.DirCache(certsDir),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(publicHosts...),
		}
	}

	return &Manager{ac: ac, log: log, dataDir: dataDir, hosts: hosts}
}

// TLSConfig returns a *tls.Config that serves ACME and self-signed certs.
func (m *Manager) TLSConfig() *tls.Config {
	return &tls.Config{
		GetCertificate: m.getCertificate,
		MinVersion:     tls.VersionTLS12,
	}
}

// HTTPHandler wraps fallback with the ACME HTTP-01 challenge handler.
// All non-challenge requests are passed through to fallback unchanged.
func (m *Manager) HTTPHandler(fallback http.Handler) http.Handler {
	if m.ac != nil {
		return m.ac.HTTPHandler(fallback)
	}
	return fallback
}

// AddHost dynamically adds a host to the managed set.
func (m *Manager) AddHost(host string) {
	for _, h := range m.hosts {
		if h == host {
			return
		}
	}
	m.hosts = append(m.hosts, host)
	if m.ac != nil && !IsLocalHost(host) {
		m.ac.HostPolicy = autocert.HostWhitelist(filterPublicHosts(m.hosts)...)
	}
}

// CertStatuses returns the current cert status for every managed host.
func (m *Manager) CertStatuses() []CertStatus {
	certsDir := filepath.Join(m.dataDir, "certs")
	var out []CertStatus
	for _, host := range m.hosts {
		s := CertStatus{Host: host}
		if IsLocalHost(host) {
			s.Mode = "self_signed"
			certPath := filepath.Join(certsDir, host+".crt")
			if exp, err := certExpiry(certPath); err == nil {
				s.ExpiresAt = exp
				switch {
				case time.Now().After(exp):
					s.Status = "error"
				case time.Until(exp) < 14*24*time.Hour:
					s.Status = "expiring_soon"
				default:
					s.Status = "valid"
				}
			} else {
				s.Status = "pending"
			}
		} else {
			s.Mode = "acme"
			cachePath := filepath.Join(certsDir, host)
			if exp, err := certExpiry(cachePath); err == nil {
				s.ExpiresAt = exp
				switch {
				case time.Now().After(exp):
					s.Status = "error"
				case time.Until(exp) < 14*24*time.Hour:
					s.Status = "expiring_soon"
				default:
					s.Status = "valid"
				}
			} else {
				s.Status = "pending"
			}
		}
		out = append(out, s)
	}
	return out
}

// getCertificate is the tls.Config.GetCertificate callback.
func (m *Manager) getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	host := hello.ServerName
	if IsLocalHost(host) {
		return m.selfSignedCert(host)
	}
	if m.ac != nil {
		return m.ac.GetCertificate(hello)
	}
	return nil, nil
}

// selfSignedCert generates or loads a self-signed certificate for host.
func (m *Manager) selfSignedCert(host string) (*tls.Certificate, error) {
	certsDir := filepath.Join(m.dataDir, "certs")
	certPath := filepath.Join(certsDir, host+".crt")
	keyPath := filepath.Join(certsDir, host+".key")

	// Load and validate existing cert.
	if _, err := os.Stat(certPath); err == nil {
		if exp, err := certExpiry(certPath); err == nil && time.Now().Before(exp) {
			cert, err := tls.LoadX509KeyPair(certPath, keyPath)
			if err == nil {
				return &cert, nil
			}
		}
	}

	// Generate a fresh self-signed cert.
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating key: %w", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: host},
		DNSNames:     []string{host},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("creating certificate: %w", err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("marshaling key: %w", err)
	}

	if err := writePEM(certPath, "CERTIFICATE", certDER); err != nil {
		return nil, err
	}
	if err := writePEM(keyPath, "EC PRIVATE KEY", keyDER); err != nil {
		return nil, err
	}

	m.log.Info("generated self-signed certificate", "host", host,
		"expires", tmpl.NotAfter.Format("2006-01-02"))

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	return &cert, err
}

// IsLocalHost reports whether a host cannot use ACME (local/private/IP).
func IsLocalHost(host string) bool {
	if host == "localhost" || host == "" {
		return true
	}
	if net.ParseIP(host) != nil {
		return true
	}
	for _, suffix := range []string{".local", ".internal", ".test", ".lan", ".home", ".localdomain"} {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	return false
}

// RedirectToHTTPS returns an http.Handler that sends a 301 redirect to https://
// for all requests. Use this as the fallback when HTTPS is fully enabled.
func RedirectToHTTPS(httpsPort int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := "https://" + stripPort(r.Host)
		if httpsPort != 443 {
			target += fmt.Sprintf(":%d", httpsPort)
		}
		target += r.RequestURI
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})
}

func stripPort(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}

func filterPublicHosts(hosts []string) []string {
	var out []string
	for _, h := range hosts {
		if !IsLocalHost(h) {
			out = append(out, h)
		}
	}
	return out
}

func certExpiry(path string) (time.Time, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return time.Time{}, fmt.Errorf("no PEM block in %s", path)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}
	return cert.NotAfter, nil
}

func writePEM(path, pemType string, data []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: pemType, Bytes: data})
}
