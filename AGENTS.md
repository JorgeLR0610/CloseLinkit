# AGENTS.md

CloseLinkit is a URL shortener. Monorepo: `server/` (Go backend + DB layer),
`frontend/` (React + TypeScript + Vite), `docs/` (architecture, ADRs),
`scripts/` (dev automation), `.github` (CI workflows and dependabot configuration),
`agent/` (working notes for agents).
Stack: Go (net/http), PostgreSQL, Goose migrations, sqlc, Docker Compose.

## Read first
- Every session: `agent/current-state.md` (active task, next steps).
- If `TASK.md` has unchecked items, it is the spec for this session.
- Before architectural changes: `docs/architecture.md`, `docs/ADR/`, `agent/decisions.md`.

## Commands
Backend, from `server/` (mirrors CI):
- `go test ./... -cover -race`
- `golangci-lint run`
Backend, from repo root:
- `make govulncheck`
- `make sqlc-generate` (after editing `server/db/queries/`)
- `make migrate-up` / `make migrate-status`
Frontend, from `frontend/`:
- `npm test`, `npm run lint`, `npm run format-check` (fix with `npm run format`), `npm run build`
Local environment: `docker compose up -d`, then `make migrate-up`.
A task is done only when the checks for the areas you touched pass.

## Architecture rules
- Backend layers: handler -> service -> repository -> PostgreSQL (see ADR 0009).
  Business logic only in services. SQL only in `server/db/queries/` via sqlc.
- Never hand-edit generated files in `server/internal/repository/`.
- Schema changes go in a NEW Goose migration. Never edit an existing one.
- Any endpoint change must update, in the same change: `server/docs/openapi.yaml`,
  the README endpoint table, backend DTOs, and frontend types in `frontend/src/types/`.
- Frontend: API calls live in `frontend/src/services/`, not in components.

## Do not undo these decisions
- Access token lives only in memory on the client. Never store tokens in localStorage.
- Refresh token is an HttpOnly cookie on `Path=/api/v1/auth`, rotated on every refresh (ADR 0010).
- Anonymous URLs expire in 7 days; authenticated URLs never expire; claiming clears `expires_at`.
- Only http/https URLs are accepted; loopback, private and link-local hosts are rejected.
- More detail in `agent/decisions.md`.

## Never
- Commit, push, or switch branches unless asked.
- Print, read aloud, or modify `.env`. Add new variables to `.env.example` with placeholders.
- Run `scripts/setup.sh` (it overwrites `.env`).
- Add dependencies without asking.
- Change existing API contracts unless the task says so.
- Create or push git tags. Tags matching `v*.*.*` publish a Docker image to GHCR.

## Maintenance
When finishing or pausing work, update only "Active task" and "Next steps" in
`agent/current-state.md`, and add one line to `CHANGELOG.md` for user-visible changes.
Record new non-obvious decisions in `agent/decisions.md`.