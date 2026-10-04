# Current State

**Version:** v0.3.0 (in progress) · **Branch:** feature/add-auth
**Stack:** Go + PostgreSQL (Goose, sqlc) + React/TS/Vite. See `docs/architecture.md`.

## Done
Auth (register/login/refresh/logout), user URL listing and deletion, anonymous URL expiry
with cleanup jobs, cookie-based refresh, unified `/` view, URL claiming, stats authentication,
minimal Distroless API container (ADR 0011), and documentation updated for v0.3.0.
History: `docs/CHANGELOG.md`. Endpoints: `server/docs/openapi.yaml`.

## Active task
None (Documentation update for v0.3.0 completed)

## Next steps
1. Edit URLs / per-link analytics.
2. Custom aliases (open question in `agent/decisions.md`).
3. Infra: deployment on AWS, Kubernetes manifests.

## Watch out
- Cleanup goroutines (refresh tokens hourly, expired URLs every 12h) are started in
  `server/cmd/CloseLinkit/main.go`.