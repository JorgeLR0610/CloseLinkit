# Changelog

## v0.3.0 (unreleased)
- API Docker image published to GHCR on version tags.
- Users and JWT auth: register, login, refresh with rotation, logout (Argon2id passwords).
- Refresh token as HttpOnly cookie; access token in memory only.
- `GET /api/v1/urls`: list the authenticated user's URLs.
- Anonymous URLs expire after 7 days; periodic cleanup of expired URLs and refresh tokens.
- `POST /api/v1/shorten` now returns `expires_at`.
- `POST /api/v1/urls/claim`: transfer anonymous URLs to the user on login/signup.
- Frontend: login/signup pages, React Router, unified `/` view for guests and users.
- Migrations 002-005: users, refresh_tokens, urls.user_id, urls.expires_at.