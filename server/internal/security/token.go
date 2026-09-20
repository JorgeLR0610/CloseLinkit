package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrInvalidAlgorithm    = errors.New("unexpected signing algorithm")
	ErrEmptySecretKey      = errors.New("jwt secret key cannot be empty")
	ErrInvalidRefreshToken = errors.New("invalid or revoked refresh token")
	ErrExpiredRefreshToken = errors.New("refresh token has expired")
)

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

const (
	DefaultAccessTokenTTL  = 15 * time.Minute
	DefaultRefreshTokenTTL = 7 * 24 * time.Hour
	TokenIssuer            = "CloseLinkit"
)

func (c *Claims) UserID() (uuid.UUID, error) {
	return uuid.Parse(c.Subject)
}

func GenerateAccessToken(secret []byte, userID uuid.UUID, email string, ttl time.Duration) (string, error) {
	if len(secret) == 0 {
		return "", ErrEmptySecretKey
	}
	if ttl <= 0 {
		ttl = DefaultAccessTokenTTL
	}

	now := time.Now()
	claims := Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    TokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("could not sign access token: %w", err)
	}

	return signedToken, nil
}

func ValidateAccessToken(secret []byte, tokenString string) (*Claims, error) {
	if len(secret) == 0 {
		return nil, ErrEmptySecretKey
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrInvalidAlgorithm, t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("could not generate refresh token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
