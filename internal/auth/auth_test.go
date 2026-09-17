package auth

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func TestPasswords(t *testing.T) {
	password := "Long-password-123"
	a, b := HashPassword(password), HashPassword(password)
	if a == b || !strings.HasPrefix(a, "$2") || !VerifyPassword(a, password) || VerifyPassword(a, "wrong") {
		t.Fatal("bcrypt verification failed")
	}
	legacy := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	if !VerifyPassword(legacy, password) || VerifyPassword(legacy, "wrong") || VerifyPassword(a, strings.Repeat("a", 73)) {
		t.Fatal("legacy/length validation failed")
	}
}
