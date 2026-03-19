package auth_test

import (
	"testing"

	"github.com/Elchi-dev/onyx/internal/auth"
)

func TestHashAndCheck(t *testing.T) {
	hash, err := auth.HashPassword("supersecret99")
	if err != nil {
		t.Fatal(err)
	}
	if !auth.CheckPassword(hash, "supersecret99") {
		t.Error("correct password should match")
	}
	if auth.CheckPassword(hash, "wrongpassword") {
		t.Error("wrong password should not match")
	}
}

func TestGenerateKey(t *testing.T) {
	key, err := auth.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 64 {
		t.Errorf("want 64 hex chars, got %d", len(key))
	}
	key2, _ := auth.GenerateKey()
	if key == key2 {
		t.Error("two generated keys should not be identical")
	}
}
