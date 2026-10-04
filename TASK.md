# Task: Add require auth to stats endpoint

## 1. Context and goal
Right now, the `GET /api/v1/{shortCode}/stats` endpoint does not require auth.
Refactor this endpoint to require auth and ajust all related components as required, such as docs, tests and UI button to only be displayed when user is authenticated

## 2. Starting points
- Backend: `server/cmd/CloseLinkit/main.go`, `server/internal/api/v1/urls_test.go`
- Route: docs: `server/docs/openapi.yaml`, README table
- Frontend: `frontend/src/components/URLList/URLListItem.tsx`

## 3. Acceptance criteria
- [x] `go test ./... -cover -race` and `golangci-lint run` pass in `server/`
- [x] `npm test`, `npm run lint`, `npm run format-check`, `npm run build` pass in `frontend/`
- [x] Unauthenticated request returns 401
- [x] Tests cover all three cases above, in service, handler and frontend

## 4. Restrictions
- No new migration; no changes to `/shorten`, resolve, or claim behavior.
- No new dependencies.