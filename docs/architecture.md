# Architecture

## Overview

CloseLinkit is a URL shortening service composed of a Go backend, a React frontend and a PostgreSQL database.

The system exposes a REST API and a web client that allows users to create, retrieve and resolve shortened URLs.

---

## Technology Stack

| Component | Technology |
|----------|------------|
| Backend | Go (net/http) |
| Frontend | React + TypeScript + Vite |
| Database | PostgreSQL |
| SQL Code Generation | sqlc |
| Migrations Tool | goose |
| Security & Auth | Argon2id, JWT (golang-jwt/jwt/v5) |
| Containerization | Docker Compose & Distroless Static (ADR 0011) |

---

## System Components

### Backend (Go)

Implements the REST API, business logic, security/auth primitives, and communication with the database.

### Frontend (React)

Provides the graphical user interface (GUI) and communicates with the backend through HTTP requests.

### PostgreSQL

Persists application data.

### Docker Compose

Provides the local development environment by orchestrating the application services.

---

## Backend Layers

### Handler Layer

Receives HTTP requests, validates input, invokes the service layer, and builds HTTP responses.

### Middleware Layer

Provides cross-cutting HTTP request handling including JWT authentication (`RequireAuth`, `OptionalAuth`), CORS policies, and IP-based rate limiting.

### Service Layer

Implements the application's business logic, user authentication, URL claiming, and coordinates domain operations.

### Security Layer

Provides Argon2id password hashing and constant-time verification against timing attacks, as well as JWT access and refresh token generation and cryptographic validation (`internal/security/`).

### Repository Layer

Provides database access through SQLC-generated queries.

### Database (PostgreSQL)

Persists application data.

---

## Authentication

CloseLinkit uses short-lived JWT access tokens and rotating refresh tokens.
See ADR 0010 for the rationale.

### Flow

```text
Login / Register
  Client ── POST /api/v1/auth/login ──▶ API
  Client ◀── access_token (JSON body) + refresh_token (HttpOnly cookie) ── API

Authenticated request
  Client ── Authorization: Bearer <access_token> ──▶ API

Session restore / token expiry
  Client ── POST /api/v1/auth/refresh (cookie sent automatically) ──▶ API
  Client ◀── new access_token + user + rotated refresh cookie ── API

Logout
  Client ── POST /api/v1/auth/logout ──▶ API (revokes token, clears cookie)
```

### Token storage
- The access token is kept only in memory on the client (never in localStorage).
- The refresh token is an opaque random value; only its SHA-256 hash is stored in
  PostgreSQL. The cookie is HttpOnly, Secure, SameSite, scoped to `Path=/api/v1/auth`.
- Each refresh revokes the used token and issues a new one (rotation).
- Expired refresh tokens are deleted by an hourly background job.

### Route protection
| Route | Middleware | Behavior |
|-------|-----------|----------|
| `POST /api/v1/shorten` | `OptionalAuth` | Anonymous: URL expires in 7 days. Authenticated: permanent, owned by the user. |
| `GET /api/v1/urls` | `RequireAuth` | Lists the authenticated user's URLs. |
| `POST /api/v1/urls/claim` | `RequireAuth` | Transfers anonymous URLs to the user. |
| `DELETE /api/v1/urls/{shortCode}` | `RequireAuth` | Deletes a shortened URL owned by the authenticated user. |
| `GET /api/v1/{shortCode}/stats` | `RequireAuth` | Retrieves access counts and statistics for a short code. |

### Claiming anonymous URLs
Guests keep their links in localStorage (`history`). After login, registration, or
session restore, the client sends the short codes to `POST /api/v1/urls/claim`. The
backend assigns ownership only to URLs with no owner and clears their expiration.

---

## Request Flow

```text
                Browser
                   │
                   ▼
             React Frontend
                   │
              HTTP / JSON
                   │
                   ▼
             Handler Layer
                   │
                   ▼
             Service Layer
                   │
                   ▼
            Repository Layer
                   │
                   ▼
               PostgreSQL
```

The response follows the same path in reverse.

---

## Roadmap

### Implemented
- User authentication (JWT + refresh tokens), user accounts, per-user URL listing
- Anonymous URL expiration (7 days) and claiming
- User URL deletion (`DELETE /api/v1/urls/{shortCode}`) and protected analytics (`GET /api/v1/{shortCode}/stats`)
- Hardened Distroless runtime image (ADR 0011)

### Planned
- Link editing (updating target URL)
- Custom short URLs (user-specified aliases)

## Infrastructure

### Implemented
- CI: backend (tests, golangci-lint, govulncheck), frontend (tests, lint, format check),
  secrets scan.
- Minimal static distroless runtime image (`gcr.io/distroless/static-debian13:nonroot`, ADR 0011).
- API Docker image published to GHCR on `v*.*.*` tags (`docker-publish.yaml`);
  pull requests build the image without pushing.

### Planned
- AWS deployment (EC2, S3, CloudFront)
- Kubernetes manifests