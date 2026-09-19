# Current Project State

## Project Overview

- **Project Name:** CloseLinkit
- **Current Version:** v0.2.0
- **Repository Structure:** Monorepo (`server/`, `frontend/`, `docs/`, `scripts/`, `agent/`)

---

## Architecture & Technology Stack

- **Backend:** Go 1.26.2 (`net/http`, layered architecture: handlers $\rightarrow$ service $\rightarrow$ repository $\rightarrow$ PostgreSQL).
- **Frontend:** React 19.2.7 + TypeScript + Vite.
- **Database:** PostgreSQL 18.4 with migrations managed by Goose (`server/db/migrations/`).
- **SQL Code Generation:** `sqlc` v1.31.1 mapping queries in `server/db/queries/` to `server/internal/repository/`.
- **Testing & Code Quality Suites:**
   - **Backend:** Full unit tests (`go test ./... -cover -race`), security scans (`make govulncheck`), and GitHub Actions CI workflow (`backend-ci.yaml`).
   - **Frontend:** Full unit/integration tests using Vitest, React Testing Library, and MSW (`npm test`), linting (`oxlint`), formatting (`oxfmt`), and GitHub Actions CI workflow (`frontend-ci.yaml`).
- **Infrastructure:** Docker Compose (`compose.yaml`), GitHub Actions CI/CD workflows.

---

## Recent Additions (v0.2.0)

1. **Dependency Maintenance:**
   - Minor package updates across backend and frontend dependencies.

2. **Database Migrations (Goose):**
   - `server/db/migrations/001_urls.sql`: Base `urls` table.
   - `server/db/migrations/002_users.sql`: Created `users` table:
     - `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
     - `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
     - `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
     - `email TEXT UNIQUE NOT NULL` (with unique index `idx_users_email` on `LOWER(email)`)
     - `email_verified_at TIMESTAMPTZ`
     - `hashed_password TEXT NOT NULL`
   - `server/db/migrations/003_refresh_tokens.sql`: Created `refresh_tokens` table:
     - `id UUID PRIMARY KEY`
     - `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
     - `token_hash TEXT NOT NULL UNIQUE`
      - `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
      - `expires_at TIMESTAMPTZ NOT NULL`
      - `revoked_at TIMESTAMPTZ`
      - Index on `user_id` (`idx_refresh_tokens_user_id`)
    - `server/db/migrations/004_add_user_id_to_urls.sql`:
      - Added nullable `user_id UUID` column to `urls` table.
      - Foreign key constraint `user_fk` referencing `users(id)` with `ON DELETE SET NULL`.
      - Index on `urls.user_id` (`idx_urls_user_id`).

3. **Environment Configuration:**
   - Added `JWT_SECRET` to `.env` (used for signing and validating JWT access tokens) and placeholder to `.env.example`.

4. **SQL Queries & SQLC Repository Layer (Step 1 & Step 2 Completed):**
   - Created `server/db/queries/users.sql` (`CreateUser`, `GetUserByEmail`, `GetUserByID`, `UpdateUserPassword`, `MarkEmailVerified`).
   - Created `server/db/queries/refresh_tokens.sql` (`CreateRefreshToken`, `GetRefreshTokenByHash`, `RevokeRefreshToken`, `RevokeRefreshTokenByID`, `RevokeAllUserRefreshTokens`, `DeleteExpiredTokens`).
   - Updated `server/db/queries/urls.sql` (`CreateURL` accepts optional `user_id`, added `GetURLsByUserID`).
   - Regenerated repository code via `make sqlc-generate` in `server/internal/repository/` (`models.go`, `urls.sql.go`, `users.sql.go`, `refresh_tokens.sql.go`).

5. **Authentication Security & Service Layer (Step 3 Completed):**
   - Created `server/internal/security/password.go` (`HashPassword`, `HashPasswordWithParams`, `VerifyPassword` using Argon2id with OWASP-recommended parameters and constant-time comparison).
   - Created `server/internal/security/token.go` (JWT access token generation and validation using `golang-jwt/jwt/v5`, cryptographically secure refresh token issuance and SHA-256 hashing).
   - Created `server/internal/service/auth.go` (`AuthService` and `AuthRepository` interface implementing `Register`, `Login`, `RefreshToken` with rotation, `Logout`, `RevokeAllUserSessions`, and `ValidateAccessToken`).
   - Added comprehensive unit test suites in `security/` (`password_test.go`, `token_test.go`) and `service/` (`auth_test.go`).

6. **Authentication Handler & Transport Layer (Step 4 Completed):**
   - Created `server/internal/api/v1/auth.go` (handlers for `POST /api/v1/auth/register`, `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`, and `POST /api/v1/auth/logout`).
   - Created `server/internal/middleware/auth.go` (`RequireAuth` and `OptionalAuth` middlewares with context injection of `user_id`).
   - Updated `server/internal/service/urls.go` to optionally associate the created shortened URL with authenticated users via `UserIDFromContext`.
   - Wired auth routes and `OptionalAuth` into `server/cmd/CloseLinkit/main.go`.
   - Added comprehensive unit test suites (`server/internal/api/v1/auth_test.go` and `server/internal/middleware/auth_test.go`).

---

7. **Testing & Documentation (Step 5 Completed):**
   - Updated `server/docs/openapi.yaml` to document:
     - Security scheme: `bearerAuth` (HTTP Bearer JWT).
     - New endpoints: `POST /api/v1/auth/register`, `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`, `POST /api/v1/auth/logout`.
     - Updated `POST /api/v1/shorten` with optional Bearer authentication scheme and updated description.
     - New component schemas: `RegisterRequest`, `LoginRequest`, `LoginResponse`, `RefreshTokenRequest`, `RefreshTokenResponse`, `LogoutRequest`, `UserResponse`.
   - Updated `README.md` API Endpoints reference table with authentication endpoints.
   - Verified OpenAPI embedding and full backend test suite (`go test -count=1 ./... -race`) with all tests passing.
   - Verified zero vulnerabilities with `make govulncheck`.

---

## Active Task & Next Steps

The backend JWT-based User Authentication milestones (Steps 1–5) are fully completed (ready for `v0.3.0`).

### Next Steps:
- Frontend Authentication Integration:
  - Auth context and token management (handling access tokens, refresh token rotation, in-memory/secure cookie storage).
  - Register & Login UI components and forms.
  - Authenticated user session state in frontend navigation.
  - User Dashboard to view and manage user-created short links.

