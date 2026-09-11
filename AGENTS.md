# AGENTS.md

## Project Overview

CloseLinkit is a URL shortening web application implemented as a monorepo.

The project consists of:

* `server/`: Go backend and database layer.
* `frontend/`: React + TypeScript + Vite frontend.
* `docs/`: Architecture documentation and Architecture Decision Records.
* `scripts/`: Development and automation scripts.

The application uses PostgreSQL as its database and Docker Compose for the local development environment.

## Architecture

The backend follows a layered architecture:

1. Handler/transport layer
2. Service/business logic layer
3. Repository/data access layer
4. PostgreSQL database

Do not bypass these layers without a clear reason.

See `docs/architecture.md` for the overall system architecture.

## Important Documentation

Before making architectural changes, read:

* `docs/architecture.md`
* Relevant files under `docs/ADR/`

Before modifying the current implementation or continuing an unfinished task, read:

* `agent/current-state.md`

Before introducing a new technical decision that may affect the architecture, review:

* `agent/decisions.md`
* Existing ADRs under `docs/ADR/`

## Backend Guidelines

* Use idiomatic Go.
* Keep HTTP handlers focused on transport concerns.
* Keep business rules in the service layer.
* Keep database access in the repository layer.
* Prefer dependency injection through interfaces where appropriate.
* Do not place business logic directly in handlers.
* Do not access PostgreSQL directly from the service layer.

## Frontend Guidelines

* Use TypeScript.
* Keep API communication separate from presentation components.
* Follow the existing project conventions before introducing new abstractions.
* Reuse existing types and utilities when possible.

## Database Guidelines

* PostgreSQL is the source of truth for persisted data.
* Database schema changes must be implemented through migrations.
* SQL queries should follow the existing `sqlc` workflow.
* Do not manually modify generated `sqlc` files unless there is an explicit reason.

## API Guidelines

* Preserve the existing REST API contract unless the task explicitly requires changing it.
* Check the OpenAPI documentation when modifying endpoints.
* When changing request or response structures, update all affected backend and frontend types.

## Testing and Validation

Before considering a backend change complete:

```bash
cd server
go test ./...
```

For frontend changes:

```bash
cd frontend
npm test
npm run lint
npm run build
```

Run additional checks when relevant, including race detection or vulnerability scanning.

## Working Rules

* Inspect the existing implementation before proposing a new abstraction.
* Prefer small, focused changes.
* Do not rewrite working code unnecessarily.
* Preserve existing architectural decisions unless the task explicitly changes them.
* When behavior is intentionally changed, update the relevant documentation and tests.
* Keep `agent/current-state.md` updated when completing or leaving work partially implemented.

## Current Project State

See `agent/current-state.md`.

## Additional Documentation

The `README.md` contains setup instructions, environment variables, API endpoints, testing commands, and the current high-level roadmap.