package security_test

import (
	"errors"
	"testing"
	"time"

	"github.com/JorgeLR0610/CloseLinkit/internal/security"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	secret := []byte("super-secret-key-for-testing-1234")
	userID := uuid.New()
	email := "test@example.com"
	ttl := 15 * time.Minute

	tokenString, err := security.GenerateAccessToken(secret, userID, email, ttl)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	claims, err := security.ValidateAccessToken(secret, tokenString)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}

	actualUserID, err := claims.UserID()
	if err != nil {
		t.Fatalf("unexpected error getting actual userID: %v", err)
	}

	if actualUserID != userID {
		t.Errorf("expected userID %v, got %v", userID, actualUserID)
	}
	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}
	if claims.Issuer != security.TokenIssuer {
		t.Errorf("expected issuer %s, got %s", security.TokenIssuer, claims.Issuer)
	}
	if claims.Subject != userID.String() {
		t.Errorf("expected subject %s, got %s", userID.String(), claims.Subject)
	}
}

func TestValidateAccessToken_Expired(t *testing.T) {
	secret := []byte("super-secret-key-for-testing-1234")
	userID := uuid.New()
	email := "test@example.com"

	// Create an expired token
	now := time.Now().Add(-1 * time.Hour)
	claims := security.Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    security.TokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)), // expired 50 minutes ago
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("unexpected error signing token: %v", err)
	}

	_, err = security.ValidateAccessToken(secret, tokenString)
	if err == nil {
		t.Fatal("expected error on expired token, got nil")
	}
	if !errors.Is(err, security.ErrInvalidToken) {
		t.Errorf("expected error wrapping ErrInvalidToken, got %v", err)
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	secretA := []byte("secret-key-A-12345678901234567890")
	secretB := []byte("secret-key-B-12345678901234567890")
	userID := uuid.New()
	email := "test@example.com"

	tokenString, err := security.GenerateAccessToken(secretA, userID, email, 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	_, err = security.ValidateAccessToken(secretB, tokenString)
	if err == nil {
		t.Fatal("expected error with wrong secret, got nil")
	}
	if !errors.Is(err, security.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateAccessToken_NoneAlgorithm(t *testing.T) {
	secret := []byte("secret-key-12345678901234567890")
	userID := uuid.New()
	email := "test@example.com"

	claims := security.Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID.String(),
		},
	}
	// Sign using None algorithm
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("unexpected error creating token: %v", err)
	}

	_, err = security.ValidateAccessToken(secret, tokenString)
	if err == nil {
		t.Fatal("expected error rejecting 'none' algorithm, got nil")
	}
	if !errors.Is(err, security.ErrInvalidAlgorithm) {
		t.Errorf("expected error wrapping ErrInvalidAlgorithm, got %v", err)
	}
}

func TestValidateAccessToken_EmptySecret(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"

	_, err := security.GenerateAccessToken([]byte(""), userID, email, 15*time.Minute)
	if !errors.Is(err, security.ErrEmptySecretKey) {
		t.Errorf("expected ErrEmptySecretKey, got %v", err)
	}

	_, err = security.ValidateAccessToken([]byte(""), "some-token")
	if !errors.Is(err, security.ErrEmptySecretKey) {
		t.Errorf("expected ErrEmptySecretKey, got %v", err)
	}
}

func TestGenerateAndHashRefreshToken(t *testing.T) {
	token1, err := security.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error generating refresh token: %v", err)
	}
	if token1 == "" {
		t.Fatal("expected non-empty refresh token")
	}

	token2, err := security.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error generating refresh token 2: %v", err)
	}
	if token1 == token2 {
		t.Fatal("expected two generated refresh tokens to be distinct")
	}

	hash1 := security.HashRefreshToken(token1)
	hash1Again := security.HashRefreshToken(token1)
	if hash1 != hash1Again {
		t.Errorf("expected deterministic hash for same token")
	}

	hash2 := security.HashRefreshToken(token2)
	if hash1 == hash2 {
		t.Errorf("expected different hashes for different tokens")
	}
}
