package service

import (
	"errors"
	"testing"
	"time"
)

func TestAuthTokenRoundTrip(t *testing.T) {
	svc := newAuthService(nil, nil, AuthOptions{Secret: "test-secret", AccessTTLSeconds: 60})

	token, err := svc.signToken("user-1", "access", time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	userID, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if userID != "user-1" {
		t.Fatalf("userID = %q, want user-1", userID)
	}
}

func TestValidatePassword(t *testing.T) {
	if err := validatePassword("Secret123!"); err != nil {
		t.Fatal(err)
	}
	if err := validatePassword("secret123"); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("err = %v, want ErrInvalidPassword", err)
	}
}
