package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "correct-password" {
		t.Fatal("hash must not equal the plaintext password")
	}
	if err := CheckPassword(hash, "correct-password"); err != nil {
		t.Fatalf("expected password to match: %v", err)
	}
	if err := CheckPassword(hash, "wrong-password"); err == nil {
		t.Fatal("expected password mismatch")
	}
}
