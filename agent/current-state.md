# Current State

**Version:** v0.2.0 (in progress) · **Branch:** feature/add-auth
**Stack:** Go + PostgreSQL (Goose, sqlc) + React/TS/Vite. See `docs/architecture.md`.

## Done
Auth (register/login/refresh/logout), user URL listing, anonymous URL expiry with
cleanup jobs, cookie-based refresh, unified `/` view, URL claiming.
History: `CHANGELOG.md`. Endpoints: `server/docs/openapi.yaml`.

## Active task
None (Delete user URLs completed)

## Next steps
1. Edit URLs / per-link analytics.
2. Custom aliases (open question in `agent/decisions.md`).
3. Infra: deployment on AWS, Kubernetes manifests.

## Watch out
- Cleanup goroutines (refresh tokens hourly, expired URLs every 12h) are started in
  `server/cmd/CloseLinkit/main.go`.