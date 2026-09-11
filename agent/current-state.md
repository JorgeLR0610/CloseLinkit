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
     - `created_at TIMESTAMPTZ NOT NULL`
     - `expires_at TIMESTAMPTZ NOT NULL`
     - `revoked_at TIMESTAMPTZ`
     - Index on `user_id` (`idx_refresh_tokens_user_id`)
   - `server/db/migrations/004_add_user_id_to_urls.sql`:
     - Added nullable `user_id UUID` column to `urls` table.
     - Foreign key constraint `user_fk` referencing `users(id)` with `ON DELETE SET NULL`.
     - Index on `urls.user_id` (`idx_urls_user_id`).

3. **Environment Configuration:**
   - Added `JWT_SECRET` to `.env` (used for signing and validating JWT access tokens).
   - *(Note: Ensure `.env.example` includes a placeholder `JWT_SECRET=` if syncing configuration templates).*

---

## Active Task & Next Steps

The overarching objective is to implement **JWT-based User Authentication** starting from the server side.

### Immediate Step 1: SQL Queries for SQLC
Define the necessary queries in `server/db/queries/`:

- **Create `server/db/queries/users.sql`:**
  - `CreateUser`: Insert a new user returning the created record.
  - `GetUserByEmail`: Query user by lowercase email.
  - `GetUserByID`: Query user by UUID.
  - `UpdateUserPassword`: Update `hashed_password` and `updated_at`.
  - `MarkEmailVerified`: Set `email_verified_at` and `updated_at`.
- **Create `server/db/queries/refresh_tokens.sql`:**
  - `CreateRefreshToken`: Insert a refresh token entry.
  - `GetRefreshTokenByHash`: Retrieve active/non-revoked token by hash.
  - `RevokeRefreshToken`: Set `revoked_at = NOW()` by token hash or ID.
  - `RevokeAllUserRefreshTokens`: Revoke all active tokens for a `user_id`.
  - `DeleteExpiredTokens`: Periodic cleanup of expired/revoked tokens.
- **Update `server/db/queries/urls.sql`:**
  - Update `CreateURL` to accept optional `user_id` (`sql.Null[uuid.UUID]` or nullable UUID).
  - Add `GetURLsByUserID`: List shortened URLs owned by a specific authenticated user.

### Immediate Step 2: SQLC Code Generation
Run the code generator to update the Go repository layer:
```bash
make sqlc-generate
# or
cd server && go generate ./...
```
Verify the generated files in `server/internal/repository/`.

### Subsequent Step 3: Service Layer Implementation
- Add password hashing utilities (using `golang.org/x/crypto/argon2`).
- Implement JWT token generation, claims handling, and validation in `server/internal/service/` (e.g. `auth.go`).
- Implement refresh token issuance, rotation, and revocation logic.

### Subsequent Step 4: Handler & Transport Layer
- Create handlers for auth routes:
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`
  - `POST /api/v1/auth/refresh`
  - `POST /api/v1/auth/logout`
- Implement JWT middleware (`server/internal/middleware/auth.go`) to authenticate requests and inject `user_id` into request context.
- Update `POST /api/v1/shorten` to optionally associate the created URL with the authenticated user.

### Subsequent Step 5: Testing & Documentation
- Write unit tests for new service methods, middleware, and handlers.
- Update `server/docs/openapi.yaml` to document the new auth endpoints.
