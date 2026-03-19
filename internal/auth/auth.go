// Package auth handles password hashing, session token generation,
// and encryption key management for Onyx.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 12

	// SettingKeyPasswordHash is the database key for the bcrypt password hash.
	SettingKeyPasswordHash = "dashboard_password_hash"
	// SettingKeyEncryptionKey is the database key for the AES-256 encryption key.
	SettingKeyEncryptionKey = "encryption_key"

	// SessionDurationShort is the session lifetime for a normal login (24 hours).
	SessionDurationShort = 24 * time.Hour
	// SessionDurationLong is the session lifetime when "remember me" is checked (7 days).
	SessionDurationLong = 7 * 24 * time.Hour
)

// Session is an authenticated dashboard session.
type Session struct {
	Token     string
	ExpiresAt time.Time
}

// HashPassword returns a bcrypt hash of the given plaintext password.
func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(h), nil
}

// CheckPassword reports whether the plaintext password matches the stored hash.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateKey returns a cryptographically secure random 32-byte hex-encoded key.
func GenerateKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random key: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// GenerateSessionToken returns a secure random 32-byte session token as hex.
func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating session token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// NewSession creates a new session valid for the given duration.
func NewSession(rememberMe bool) (*Session, error) {
	token, err := GenerateSessionToken()
	if err != nil {
		return nil, err
	}
	dur := SessionDurationShort
	if rememberMe {
		dur = SessionDurationLong
	}
	return &Session{
		Token:     token,
		ExpiresAt: time.Now().Add(dur),
	}, nil
}
