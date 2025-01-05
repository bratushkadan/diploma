package auth_test

import (
	"testing"

	"github.com/bratushkadan/floral/pkg/auth"
)

func TestPassordHasher(t *testing.T) {
	pass := "foobar12345"
	incorrectPass := "fueo$9"
	h := auth.NewPasswordHasher("verysecretphrase")

	hashedPass, err := h.Hash(pass)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if pass == hashedPass {
		t.Fatal("hashed password must not be equal to password")
	}

	if h.Check(incorrectPass, hashedPass) {
		t.Fatal("hashed password check must not succeed for incorrect password")
	}
	if !h.Check(pass, hashedPass) {
		t.Fatal("hashed password check must succeed for correct password")
	}
}
