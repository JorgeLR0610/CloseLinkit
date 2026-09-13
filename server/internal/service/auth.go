package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
	"github.com/JorgeLR0610/CloseLinkit/internal/security"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrUserAlreadyExists   = errors.New("user with this email already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidEmail        = errors.New("invalid email address")
	ErrPasswordTooShort    = errors.New("password must be at least 8 characters long")
	ErrPasswordTooLong     = errors.New("password must be at most 72 characters long")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidRefreshToken = security.ErrInvalidRefreshToken
	ErrExpiredRefreshToken = security.ErrExpiredRefreshToken
)

type Claims = security.Claims
type Argon2Params = security.Argon2Params

type AuthRepository interface {
	CreateUser(ctx context.Context, arg repository.CreateUserParams) (repository.User, error)
	GetUserByEmail(ctx context.Context, lower string) (repository.User, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (repository.User, error)
	UpdateUserPassword(ctx context.Context, arg repository.UpdateUserPasswordParams) error
	MarkEmailVerified(ctx context.Context, id pgtype.UUID) error

	CreateRefreshToken(ctx context.Context, arg repository.CreateRefreshTokenParams) (repository.RefreshToken, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (repository.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeRefreshTokenByID(ctx context.Context, id pgtype.UUID) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID pgtype.UUID) error
	DeleteExpiredTokens(ctx context.Context) error
}

type AuthConfig struct {
	JWTSecret       []byte
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Argon2Params    security.Argon2Params
}

type UserResponse struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type AuthService struct {
	repo AuthRepository
	cfg  AuthConfig
}

func NewAuthService(repo AuthRepository, cfg AuthConfig) *AuthService {
	if cfg.AccessTokenTTL <= 0 {
		cfg.AccessTokenTTL = security.DefaultAccessTokenTTL
	}
	if cfg.RefreshTokenTTL <= 0 {
		cfg.RefreshTokenTTL = security.DefaultRefreshTokenTTL
	}
	if cfg.Argon2Params.Memory == 0 {
		cfg.Argon2Params = security.DefaultArgon2Params
	}

	return &AuthService{
		repo: repo,
		cfg:  cfg,
	}
}

func isValidEmail(email string) bool {
	if email == "" || len(email) > 254 {
		return false
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

func mapUserToResponse(user repository.User) UserResponse {
	userID, _ := uuid.FromBytes(user.ID.Bytes[:])
	var verifiedAt *time.Time
	if user.EmailVerifiedAt.Valid {
		t := user.EmailVerifiedAt.Time
		verifiedAt = &t
	}

	return UserResponse{
		ID:              userID,
		Email:           user.Email,
		EmailVerifiedAt: verifiedAt,
		CreatedAt:       user.CreatedAt.Time,
		UpdatedAt:       user.UpdatedAt.Time,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*UserResponse, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}
	if len(password) > 72 {
		return nil, ErrPasswordTooLong
	}

	hashedPassword, err := security.HashPasswordWithParams(password, s.cfg.Argon2Params)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, repository.CreateUserParams{
		Email:          email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.SQLState() == uniqueViolation {
				return nil, ErrUserAlreadyExists
			}
		}
		return nil, fmt.Errorf("could not create user: %w", err)
	}

	res := mapUserToResponse(user)
	return &res, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, *UserResponse, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !isValidEmail(email) {
		return nil, nil, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("could not query user: %w", err)
	}

	match, err := security.VerifyPassword(password, user.HashedPassword)
	if err != nil || !match {
		return nil, nil, ErrInvalidCredentials
	}

	userID, err := uuid.FromBytes(user.ID.Bytes[:])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid user id: %w", err)
	}

	tokenPair, err := s.generateAndStoreTokens(ctx, userID, user.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate tokens: %w", err)
	}

	userRes := mapUserToResponse(user)
	return tokenPair, &userRes, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, rawRefreshToken string) (*TokenPair, error) {
	rawRefreshToken = strings.TrimSpace(rawRefreshToken)
	if rawRefreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	tokenHash := security.HashRefreshToken(rawRefreshToken)

	tokenRecord, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("could not fetch refresh token: %w", err)
	}

	if time.Now().After(tokenRecord.ExpiresAt.Time) {
		_ = s.repo.RevokeRefreshToken(ctx, tokenHash)
		return nil, ErrExpiredRefreshToken
	}

	// Token rotation: revoke old token
	if err := s.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("could not revoke old refresh token: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, tokenRecord.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("could not fetch user for token: %w", err)
	}

	userID, err := uuid.FromBytes(user.ID.Bytes[:])
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	tokenPair, err := s.generateAndStoreTokens(ctx, userID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("could not issue new tokens: %w", err)
	}

	return tokenPair, nil
}

func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	rawRefreshToken = strings.TrimSpace(rawRefreshToken)
	if rawRefreshToken == "" {
		return ErrInvalidRefreshToken
	}

	tokenHash := security.HashRefreshToken(rawRefreshToken)
	if err := s.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return fmt.Errorf("could not revoke token: %w", err)
	}

	return nil
}

func (s *AuthService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.RevokeAllUserRefreshTokens(ctx, pgtype.UUID{Bytes: userID, Valid: true}); err != nil {
		return fmt.Errorf("could not revoke user sessions: %w", err)
	}
	return nil
}

func (s *AuthService) ValidateAccessToken(tokenString string) (*security.Claims, error) {
	return security.ValidateAccessToken(s.cfg.JWTSecret, tokenString)
}

func (s *AuthService) generateAndStoreTokens(ctx context.Context, userID uuid.UUID, email string) (*TokenPair, error) {
	accessToken, err := security.GenerateAccessToken(s.cfg.JWTSecret, userID, email, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("could not generate access token: %w", err)
	}

	rawRefreshToken, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("could not generate refresh token: %w", err)
	}

	tokenHash := security.HashRefreshToken(rawRefreshToken)
	refreshTokenID := uuid.New()
	expiresAt := time.Now().Add(s.cfg.RefreshTokenTTL)

	_, err = s.repo.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		ID:        pgtype.UUID{Bytes: refreshTokenID, Valid: true},
		UserID:    pgtype.UUID{Bytes: userID, Valid: true},
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("could not store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
	}, nil
}
