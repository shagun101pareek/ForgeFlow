package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestIssueAndParseToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)

	signed, err := IssueToken(secret, userID, now)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	got, err := ParseToken(secret, signed)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if got != userID {
		t.Fatalf("user id = %s, want %s", got, userID)
	}

	if _, err := ParseToken("other-secret", signed); err == nil {
		t.Fatal("expected token signed with a different secret to fail")
	}
}

func TestParseTokenRejectsExpiredToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * tokenTTL)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})
	signed, err := expired.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}
	if _, err := ParseToken(secret, signed); err == nil {
		t.Fatal("expected expired token to fail")
	}
}
