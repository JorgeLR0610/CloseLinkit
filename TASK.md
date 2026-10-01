# Task: Delete user URLs

## 1. Context and goal
Authenticated users can list their URLs on `/` but cannot delete them.
Add `DELETE /api/v1/urls/{code}` (RequireAuth) and a delete action in the URL list UI.
Deleting removes the row from `urls`; the short code then returns 404 on resolve.

## 2. Starting points
- Query: `server/db/queries/urls.sql` (new `DeleteURLByCode` with `AND user_id = $2`)
- Service: `server/internal/service/urls.go`; handler: `server/internal/api/v1/urls.go`
- Route: `server/cmd/CloseLinkit/main.go`; docs: `server/docs/openapi.yaml`, README table
- Frontend: `frontend/src/services/urls.ts`, `URLListItem.tsx`, `App.tsx`

## 3. Acceptance criteria
- [ ] `go test ./... -cover -race` and `golangci-lint run` pass in `server/`
- [ ] `npm test`, `npm run lint`, `npm run format-check`, `npm run build` pass in `frontend/`
- [ ] Deleting your own URL returns 204 and removes it from the list without a reload
- [ ] Deleting a code you don't own or that doesn't exist returns 404 (same response for both)
- [ ] Unauthenticated request returns 401
- [ ] Tests cover all three cases above, in service, handler and frontend

## 4. Restrictions
- No new migration; no changes to `/shorten`, resolve, or claim behavior.
- No new dependencies.
- Guest links (localStorage) are out of scope for now.