# Task: Delete user URLs

## 1. Context and goal
Authenticated users can list their URLs on `/` but cannot delete them.
The API endpoint has been added already, but it's missing unit tests, also add a delete action in the URL list UI alongside its unit tests.

## 2. Starting points
- Backend: `server/internal/service/urls_test.go`, `server/internal/api/v1/urls_test.go`
- Route: docs: `server/docs/openapi.yaml`, README table
- Frontend: `frontend/src/services/urls.ts`, `URLListItem.tsx`, `App.tsx`

## 3. Acceptance criteria
- [x] `go test ./... -cover -race` and `golangci-lint run` pass in `server/`
- [x] `npm test`, `npm run lint`, `npm run format-check`, `npm run build` pass in `frontend/`
- [x] Deleting your own URL returns 204 and removes it from the list without a reload
- [x] Deleting a code you don't own or that doesn't exist returns 404 (same response for both)
- [x] Unauthenticated request returns 401
- [x] Tests cover all three cases above, in service, handler and frontend

## 4. Restrictions
- No new migration; no changes to `/shorten`, resolve, or claim behavior.
- No new dependencies.
- Guest links (localStorage) are out of scope for now.