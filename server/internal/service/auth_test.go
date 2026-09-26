package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
	"github.com/JorgeLR0610/CloseLinkit/internal/security"
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockAuthRepository struct {
	CreateUserFunc                 func(ctx context.Context, arg repository.CreateUserParams) (repository.User, error)
	GetUserByEmailFunc             func(ctx context.Context, lower string) (repository.User, error)
	GetUserByIDFunc                func(ctx context.Context, id pgtype.UUID) (repository.User, error)
	UpdateUserPasswordFunc         func(ctx context.Context, arg repository.UpdateUserPasswordParams) error
	MarkEmailVerifiedFunc          func(ctx context.Context, id pgtype.UUID) error
	CreateRefreshTokenFunc         func(ctx context.Context, arg repository.CreateRefreshTokenParams) error
	GetRefreshTokenByHashFunc      func(ctx context.Context, tokenHash string) (repository.RefreshToken, error)
	RevokeRefreshTokenFunc         func(ctx context.Context, tokenHash string) error
	RevokeRefreshTokenByIDFunc     func(ctx context.Context, id pgtype.UUID) error
	RevokeAllUserRefreshTokensFunc func(ctx context.Context, userID pgtype.UUID) error
	DeleteExpiredTokensFunc        func(ctx context.Context) error

	revokeRefreshTokenCalls int
	createRefreshTokenCalls int
}

func (m *mockAuthRepository) CreateUser(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, arg)
	}
	return repository.User{}, nil
}

func (m *mockAuthRepository) GetUserByEmail(ctx context.Context, lower string) (repository.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, lower)
	}
	return repository.User{}, nil
}

func (m *mockAuthRepository) GetUserByID(ctx context.Context, id pgtype.UUID) (repository.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return repository.User{}, nil
}

func (m *mockAuthRepository) UpdateUserPassword(ctx context.Context, arg repository.UpdateUserPasswordParams) error {
	if m.UpdateUserPasswordFunc != nil {
		return m.UpdateUserPasswordFunc(ctx, arg)
	}
	return nil
}

func (m *mockAuthRepository) MarkEmailVerified(ctx context.Context, id pgtype.UUID) error {
	if m.MarkEmailVerifiedFunc != nil {
		return m.MarkEmailVerifiedFunc(ctx, id)
	}
	return nil
}

func (m *mockAuthRepository) CreateRefreshToken(ctx context.Context, arg repository.CreateRefreshTokenParams) error {
	m.createRefreshTokenCalls++
	if m.CreateRefreshTokenFunc != nil {
		return m.CreateRefreshTokenFunc(ctx, arg)
	}
	return nil
}

func (m *mockAuthRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (repository.RefreshToken, error) {
	if m.GetRefreshTokenByHashFunc != nil {
		return m.GetRefreshTokenByHashFunc(ctx, tokenHash)
	}
	return repository.RefreshToken{}, nil
}

func (m *mockAuthRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	m.revokeRefreshTokenCalls++
	if m.RevokeRefreshTokenFunc != nil {
		return m.RevokeRefreshTokenFunc(ctx, tokenHash)
	}
	return nil
}

func (m *mockAuthRepository) RevokeRefreshTokenByID(ctx context.Context, id pgtype.UUID) error {
	if m.RevokeRefreshTokenByIDFunc != nil {
		return m.RevokeRefreshTokenByIDFunc(ctx, id)
	}
	return nil
}

func (m *mockAuthRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID pgtype.UUID) error {
	if m.RevokeAllUserRefreshTokensFunc != nil {
		return m.RevokeAllUserRefreshTokensFunc(ctx, userID)
	}
	return nil
}

func (m *mockAuthRepository) DeleteExpiredTokens(ctx context.Context) error {
	if m.DeleteExpiredTokensFunc != nil {
		return m.DeleteExpiredTokensFunc(ctx)
	}
	return nil
}

func defaultTestAuthConfig() service.AuthConfig {
	return service.AuthConfig{
		JWTSecret:       []byte("test-jwt-secret-key-123456789012"),
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Argon2Params:    security.FastArgon2Params,
	}
}

func TestAuthService_Register(t *testing.T) {
	cfg := defaultTestAuthConfig()
	testUUID := uuid.New()

	tests := []struct {
		name        string
		email       string
		password    string
		setupRepo   func() *mockAuthRepository
		expectedErr error
	}{
		{
			name:     "successful registration",
			email:    "User@Example.com",
			password: "password123",
			setupRepo: func() *mockAuthRepository {
				return &mockAuthRepository{
					CreateUserFunc: func(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
						if arg.Email != "user@example.com" {
							t.Errorf("expected email to be lowercased to user@example.com, got %s", arg.Email)
						}
						return repository.User{
							ID:        pgtype.UUID{Bytes: testUUID, Valid: true},
							Email:     arg.Email,
							CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
							UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
						}, nil
					},
				}
			},
			expectedErr: nil,
		},
		{
			name:     "duplicate email error",
			email:    "duplicate@example.com",
			password: "password123",
			setupRepo: func() *mockAuthRepository {
				return &mockAuthRepository{
					CreateUserFunc: func(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
						return repository.User{}, &pgconn.PgError{
							Code: "23505",
						}
					},
				}
			},
			expectedErr: service.ErrUserAlreadyExists,
		},
		{
			name:     "invalid email",
			email:    "not-an-email",
			password: "password123",
			setupRepo: func() *mockAuthRepository {
				return &mockAuthRepository{}
			},
			expectedErr: service.ErrInvalidEmail,
		},
		{
			name:     "password too short",
			email:    "user@example.com",
			password: "123",
			setupRepo: func() *mockAuthRepository {
				return &mockAuthRepository{}
			},
			expectedErr: service.ErrPasswordTooShort,
		},
		{
			name:     "password too long",
			email:    "user@example.com",
			password: strings.Repeat("a", 73),
			setupRepo: func() *mockAuthRepository {
				return &mockAuthRepository{}
			},
			expectedErr: service.ErrPasswordTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			srv := service.NewAuthService(repo, cfg)

			userRes, err := srv.Register(context.Background(), tt.email, tt.password)
			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if userRes == nil {
					t.Fatal("expected userRes not to be nil")
				}
				if userRes.Email != strings.ToLower(tt.email) {
					t.Errorf("expected email %s, got %s", strings.ToLower(tt.email), userRes.Email)
				}
				if userRes.ID != testUUID {
					t.Errorf("expected user ID %v, got %v", testUUID, userRes.ID)
				}
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	cfg := defaultTestAuthConfig()
	testUUID := uuid.New()
	correctPassword := "password123"
	hashedPassword, err := security.HashPasswordWithParams(correctPassword, security.FastArgon2Params)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	tests := []struct {
		name        string
		email       string
		password    string
		setupRepo   func() *mockAuthRepository
		expectedErr error
	}{
		{
			name:     "successful login",
			email:    "user@example.com",
			password: correctPassword,
			setupRepo: func() *mockAuthRepository {
				return &mockAuthRepository{
					GetUserByEmailFunc: func(ctx context.Context, lower string) (repository.User, error) {
						return repository.User{
							ID:             pgtype.UUID{Bytes: testUUID, Valid: true},
							Email:          "user@example.com",
							HashedPassword: hashedPassword,
							CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
							UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
						}, nil
					},
					CreateRefreshTokenFunc: func(ctx context.Context, arg repository.CreateRefreshTokenParams) error {
						return nil
					},
				}
			},
			expectedErr: nil,
		},
		{
			name:     "user not found",
			email:    "unknown@example.com",
			password: "password123",
			setupRepo: func() *mockAuthRepository {
				return &mockAuthRepository{
					GetUserByEmailFunc: func(ctx context.Context, lower string) (repository.User, error) {
						return repository.User{}, pgx.ErrNoRows
					},
				}
			},
			expectedErr: service.ErrInvalidCredentials,
		},
		{
			name:     "wrong password",
			email:    "user@example.com",
			password: "wrongpassword",
			setupRepo: func() *mockAuthRepository {
				return &mockAuthRepository{
					GetUserByEmailFunc: func(ctx context.Context, lower string) (repository.User, error) {
						return repository.User{
							ID:             pgtype.UUID{Bytes: testUUID, Valid: true},
							Email:          "user@example.com",
							HashedPassword: hashedPassword,
						}, nil
					},
				}
			},
			expectedErr: service.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			srv := service.NewAuthService(repo, cfg)

			tokens, userRes, err := srv.Login(context.Background(), tt.email, tt.password)
			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tokens == nil || tokens.AccessToken == "" || tokens.RefreshToken == "" {
					t.Fatal("expected valid tokens returned")
				}
				if userRes == nil || userRes.Email != tt.email {
					t.Fatal("expected valid user response")
				}
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	cfg := defaultTestAuthConfig()
	testUserID := uuid.New()
	rawRefreshToken := "some-raw-refresh-token-value-12345"

	t.Run("successful rotation", func(t *testing.T) {
		repo := &mockAuthRepository{
			GetRefreshTokenByHashFunc: func(ctx context.Context, tokenHash string) (repository.RefreshToken, error) {
				return repository.RefreshToken{
					ID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
					UserID:    pgtype.UUID{Bytes: testUserID, Valid: true},
					TokenHash: tokenHash,
					ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(1 * time.Hour), Valid: true},
				}, nil
			},
			RevokeRefreshTokenFunc: func(ctx context.Context, tokenHash string) error {
				return nil
			},
			GetUserByIDFunc: func(ctx context.Context, id pgtype.UUID) (repository.User, error) {
				return repository.User{
					ID:    pgtype.UUID{Bytes: testUserID, Valid: true},
					Email: "user@example.com",
				}, nil
			},
			CreateRefreshTokenFunc: func(ctx context.Context, arg repository.CreateRefreshTokenParams) error {
				return nil
			},
		}

		srv := service.NewAuthService(repo, cfg)
		tokens, err := srv.RefreshToken(context.Background(), rawRefreshToken)
		if err != nil {
			t.Fatalf("unexpected error refreshing token: %v", err)
		}

		if tokens.AccessToken == "" || tokens.RefreshToken == "" {
			t.Fatal("expected new access and refresh tokens")
		}
		if tokens.RefreshToken == rawRefreshToken {
			t.Error("expected rotated refresh token to be different from original")
		}
		if repo.revokeRefreshTokenCalls != 1 {
			t.Errorf("expected 1 call to RevokeRefreshToken, got %d", repo.revokeRefreshTokenCalls)
		}
		if repo.createRefreshTokenCalls != 1 {
			t.Errorf("expected 1 call to CreateRefreshToken, got %d", repo.createRefreshTokenCalls)
		}
	})

	t.Run("invalid or non-existent token", func(t *testing.T) {
		repo := &mockAuthRepository{
			GetRefreshTokenByHashFunc: func(ctx context.Context, tokenHash string) (repository.RefreshToken, error) {
				return repository.RefreshToken{}, pgx.ErrNoRows
			},
		}

		srv := service.NewAuthService(repo, cfg)
		_, err := srv.RefreshToken(context.Background(), rawRefreshToken)
		if !errors.Is(err, service.ErrInvalidRefreshToken) {
			t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
		}
	})

	t.Run("expired refresh token", func(t *testing.T) {
		repo := &mockAuthRepository{
			GetRefreshTokenByHashFunc: func(ctx context.Context, tokenHash string) (repository.RefreshToken, error) {
				return repository.RefreshToken{
					ID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
					UserID:    pgtype.UUID{Bytes: testUserID, Valid: true},
					TokenHash: tokenHash,
					ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour), Valid: true}, // expired
				}, nil
			},
			RevokeRefreshTokenFunc: func(ctx context.Context, tokenHash string) error {
				return nil
			},
		}

		srv := service.NewAuthService(repo, cfg)
		_, err := srv.RefreshToken(context.Background(), rawRefreshToken)
		if !errors.Is(err, service.ErrExpiredRefreshToken) {
			t.Errorf("expected ErrExpiredRefreshToken, got %v", err)
		}
	})

	t.Run("empty refresh token", func(t *testing.T) {
		srv := service.NewAuthService(&mockAuthRepository{}, cfg)
		_, err := srv.RefreshToken(context.Background(), "   ")
		if !errors.Is(err, service.ErrInvalidRefreshToken) {
			t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
		}
	})
}

func TestAuthService_Logout(t *testing.T) {
	cfg := defaultTestAuthConfig()

	t.Run("successful logout", func(t *testing.T) {
		repo := &mockAuthRepository{
			RevokeRefreshTokenFunc: func(ctx context.Context, tokenHash string) error {
				return nil
			},
		}

		srv := service.NewAuthService(repo, cfg)
		err := srv.Logout(context.Background(), "some-token")
		if err != nil {
			t.Fatalf("unexpected error logging out: %v", err)
		}
		if repo.revokeRefreshTokenCalls != 1 {
			t.Errorf("expected RevokeRefreshToken to be called once, got %d", repo.revokeRefreshTokenCalls)
		}
	})

	t.Run("empty token logout", func(t *testing.T) {
		srv := service.NewAuthService(&mockAuthRepository{}, cfg)
		err := srv.Logout(context.Background(), "")
		if !errors.Is(err, service.ErrInvalidRefreshToken) {
			t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
		}
	})
}

func TestAuthService_RevokeAllUserSessions(t *testing.T) {
	cfg := defaultTestAuthConfig()
	userID := uuid.New()
	called := false

	repo := &mockAuthRepository{
		RevokeAllUserRefreshTokensFunc: func(ctx context.Context, id pgtype.UUID) error {
			called = true
			if id.Bytes != userID {
				t.Errorf("expected userID %v, got %v", userID, id.Bytes)
			}
			return nil
		},
	}

	srv := service.NewAuthService(repo, cfg)
	err := srv.RevokeAllUserSessions(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error revoking user sessions: %v", err)
	}
	if !called {
		t.Error("expected RevokeAllUserRefreshTokens to be called")
	}
}

func TestAuthService_ValidateAccessToken(t *testing.T) {
	cfg := defaultTestAuthConfig()
	userID := uuid.New()
	email := "user@example.com"

	tokenString, err := security.GenerateAccessToken(cfg.JWTSecret, userID, email, 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	srv := service.NewAuthService(&mockAuthRepository{}, cfg)
	claims, err := srv.ValidateAccessToken(tokenString)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	actualUserID, err := claims.UserID()
	if err != nil {
		t.Fatalf("unexpected error getting actual userID: %v", err)
	}

	if actualUserID != userID || claims.Email != email {
		t.Errorf("claims do not match expected values")
	}
}

func TestAuthService_StartRefreshTokenCleanup(t *testing.T) {
	cfg := defaultTestAuthConfig()
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("initial cleanup runs and context cancellation terminates", func(t *testing.T) {
		var callCount atomic.Int32
		repo := &mockAuthRepository{
			DeleteExpiredTokensFunc: func(ctx context.Context) error {
				callCount.Add(1)
				return nil
			},
		}

		srv := service.NewAuthService(repo, cfg)
		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})
		go func() {
			srv.StartRefreshTokenCleanup(ctx, 1*time.Hour, testLogger)
			close(done)
		}()

		// Wait for initial cleanup to be invoked
		for range 50 {
			if callCount.Load() >= 1 {
				break
			}
			time.Sleep(2 * time.Millisecond)
		}

		if callCount.Load() < 1 {
			t.Errorf("expected at least 1 call for initial cleanup, got %d", callCount.Load())
		}

		cancel()

		select {
		case <-done:
			// Success, exited after cancellation
		case <-time.After(1 * time.Second):
			t.Fatal("StartRefreshTokenCleanup did not exit upon context cancellation")
		}
	})

	t.Run("periodic cleanup runs repeatedly on ticker", func(t *testing.T) {
		var callCount atomic.Int32
		repo := &mockAuthRepository{
			DeleteExpiredTokensFunc: func(ctx context.Context) error {
				callCount.Add(1)
				return nil
			},
		}

		srv := service.NewAuthService(repo, cfg)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		done := make(chan struct{})
		go func() {
			srv.StartRefreshTokenCleanup(ctx, 5*time.Millisecond, testLogger)
			close(done)
		}()

		// Wait until ticker fires multiple times (initial + at least 2 ticks)
		for range 100 {
			if callCount.Load() >= 3 {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}

		if count := callCount.Load(); count < 3 {
			t.Errorf("expected at least 3 calls, got %d", count)
		}

		cancel()

		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("StartRefreshTokenCleanup did not exit upon context cancellation")
		}
	})

	t.Run("handles error from DeleteExpiredTokens gracefully without panic", func(t *testing.T) {
		var callCount atomic.Int32
		repo := &mockAuthRepository{
			DeleteExpiredTokensFunc: func(ctx context.Context) error {
				callCount.Add(1)
				return errors.New("database unavailable")
			},
		}

		srv := service.NewAuthService(repo, cfg)
		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})
		go func() {
			srv.StartRefreshTokenCleanup(ctx, 5*time.Millisecond, testLogger)
			close(done)
		}()

		// Wait for both initial and periodic calls with error
		for range 100 {
			if callCount.Load() >= 2 {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}

		if count := callCount.Load(); count < 2 {
			t.Errorf("expected at least 2 calls despite errors, got %d", count)
		}

		cancel()

		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("StartRefreshTokenCleanup did not exit upon cancellation")
		}
	})

	t.Run("exits immediately when context is already canceled", func(t *testing.T) {
		repo := &mockAuthRepository{
			DeleteExpiredTokensFunc: func(ctx context.Context) error {
				return ctx.Err()
			},
		}

		srv := service.NewAuthService(repo, cfg)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		done := make(chan struct{})
		go func() {
			srv.StartRefreshTokenCleanup(ctx, 1*time.Hour, testLogger)
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("StartRefreshTokenCleanup did not exit immediately with pre-canceled context")
		}
	})

	t.Run("resilient with nil logger and non-positive interval", func(t *testing.T) {
		repo := &mockAuthRepository{
			DeleteExpiredTokensFunc: func(ctx context.Context) error {
				return errors.New("db error")
			},
		}

		srv := service.NewAuthService(repo, cfg)
		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})
		go func() {
			// interval <= 0 and logger == nil must not panic
			srv.StartRefreshTokenCleanup(ctx, 0, nil)
			close(done)
		}()

		time.Sleep(10 * time.Millisecond)
		cancel()

		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("StartRefreshTokenCleanup did not exit")
		}
	})
}

func TestNewAuthService_Defaults(t *testing.T) {
	testSecret := []byte("secret-key-12345678901234567890")
	// Pass config with zero TTLs and zero Argon2Params
	srv := service.NewAuthService(&mockAuthRepository{}, service.AuthConfig{
		JWTSecret: testSecret,
	})

	testUserID := uuid.New()
	tokenString, err := security.GenerateAccessToken(testSecret, testUserID, "test@example.com", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	claims, err := srv.ValidateAccessToken(tokenString)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", claims.Email)
	}

	// Also verify that a service with empty secret rejects validation
	srvEmptySecret := service.NewAuthService(&mockAuthRepository{}, service.AuthConfig{})
	_, err = srvEmptySecret.ValidateAccessToken(tokenString)
	if !errors.Is(err, security.ErrEmptySecretKey) {
		t.Errorf("expected ErrEmptySecretKey, got %v", err)
	}
}

func TestAuthService_RevokeAllUserSessions_Error(t *testing.T) {
	cfg := defaultTestAuthConfig()
	dbErr := errors.New("db connection lost")
	repo := &mockAuthRepository{
		RevokeAllUserRefreshTokensFunc: func(ctx context.Context, id pgtype.UUID) error {
			return dbErr
		},
	}

	srv := service.NewAuthService(repo, cfg)
	err := srv.RevokeAllUserSessions(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Errorf("expected wrapped dbErr, got %v", err)
	}
}

func TestAuthService_Logout_Error(t *testing.T) {
	cfg := defaultTestAuthConfig()
	dbErr := errors.New("db error")
	repo := &mockAuthRepository{
		RevokeRefreshTokenFunc: func(ctx context.Context, tokenHash string) error {
			return dbErr
		},
	}

	srv := service.NewAuthService(repo, cfg)
	err := srv.Logout(context.Background(), "some-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Errorf("expected wrapped dbErr, got %v", err)
	}
}

func TestAuthService_Register_EmailVerifiedMapping(t *testing.T) {
	cfg := defaultTestAuthConfig()
	testUUID := uuid.New()
	verifiedAt := time.Now().Truncate(time.Second)

	repo := &mockAuthRepository{
		GetUserByEmailFunc: func(ctx context.Context, lower string) (repository.User, error) {
			return repository.User{}, pgx.ErrNoRows
		},
		CreateUserFunc: func(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
			return repository.User{
				ID:              pgtype.UUID{Bytes: testUUID, Valid: true},
				Email:           arg.Email,
				HashedPassword:  arg.HashedPassword,
				EmailVerifiedAt: pgtype.Timestamptz{Time: verifiedAt, Valid: true},
				CreatedAt:       pgtype.Timestamptz{Time: verifiedAt, Valid: true},
				UpdatedAt:       pgtype.Timestamptz{Time: verifiedAt, Valid: true},
			}, nil
		},
	}

	srv := service.NewAuthService(repo, cfg)
	res, err := srv.Register(context.Background(), "verified@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error registering user: %v", err)
	}

	if res.EmailVerifiedAt == nil {
		t.Fatal("expected EmailVerifiedAt to be non-nil")
	}
	if !res.EmailVerifiedAt.Equal(verifiedAt) {
		t.Errorf("expected EmailVerifiedAt %v, got %v", verifiedAt, *res.EmailVerifiedAt)
	}
}

