package security_test

import (
	"errors"
	"testing"

	"github.com/JorgeLR0610/CloseLinkit/internal/security"
)

func TestHashAndVerifyPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "standard password",
			password: "MySecurePassword123!",
		},
		{
			name:     "empty password",
			password: "",
		},
		{
			name:     "unicode password",
			password: "Secure🔐Password_123",
		},
		{
			name:     "long password",
			password: "this-is-a-very-long-password-exceeding-typical-lengths-to-test-argon2id-bounds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := security.HashPasswordWithParams(tt.password, security.FastArgon2Params)
			if err != nil {
				t.Fatalf("unexpected error hashing password: %v", err)
			}

			if hash == "" {
				t.Fatal("expected non-empty hash string")
			}

			// Verify correct password
			match, err := security.VerifyPassword(tt.password, hash)
			if err != nil {
				t.Fatalf("unexpected error verifying password: %v", err)
			}
			if !match {
				t.Errorf("expected password to match hash, but it did not")
			}

			// Verify wrong password
			matchWrong, err := security.VerifyPassword("wrong-password", hash)
			if err != nil {
				t.Fatalf("unexpected error verifying wrong password: %v", err)
			}
			if matchWrong {
				t.Errorf("expected wrong password to not match hash")
			}
		})
	}
}

func TestVerifyPassword_InvalidHashes(t *testing.T) {
	tests := []struct {
		name        string
		encodedHash string
		expectedErr error
	}{
		{
			name:        "empty string",
			encodedHash: "",
			expectedErr: security.ErrInvalidHash,
		},
		{
			name:        "invalid algorithm",
			encodedHash: "$bcrypt$v=19$m=64,t=1,p=1$c2FsdA$aGFzaA",
			expectedErr: security.ErrInvalidHash,
		},
		{
			name:        "invalid parts count",
			encodedHash: "$argon2id$v=19$m=64,t=1,p=1$c2FsdA",
			expectedErr: security.ErrInvalidHash,
		},
		{
			name:        "incompatible version",
			encodedHash: "$argon2id$v=99$m=64,t=1,p=1$c2FsdA$aGFzaA",
			expectedErr: security.ErrIncompatibleVersion,
		},
		{
			name:        "invalid params format",
			encodedHash: "$argon2id$v=19$invalid_params$c2FsdA$aGFzaA",
			expectedErr: security.ErrInvalidHash,
		},
		{
			name:        "invalid salt base64",
			encodedHash: "$argon2id$v=19$m=64,t=1,p=1$!!!$aGFzaA",
			expectedErr: security.ErrInvalidHash,
		},
		{
			name:        "invalid key base64",
			encodedHash: "$argon2id$v=19$m=64,t=1,p=1$c2FsdA$!!!",
			expectedErr: security.ErrInvalidHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := security.VerifyPassword("password", tt.encodedHash)
			if match {
				t.Errorf("expected match to be false on invalid hash")
			}
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
