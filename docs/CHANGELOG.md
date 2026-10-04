# Changelog

## v0.3.0 (released)
- Users and JWT auth: register, login, refresh with rotation, logout (Argon2id passwords).
- Constant-time password verification simulation during login to prevent user enumeration and timing attacks.
- Refresh token as HttpOnly cookie; access token in memory only (ADR 0010).
- `GET /api/v1/urls`: list the authenticated user's URLs.
- `DELETE /api/v1/urls/{shortCode}`: delete shortened URLs owned by the authenticated user.
- Anonymous URLs expire after 7 days; periodic cleanup of expired URLs (12h) and expired refresh tokens (1h).
- `POST /api/v1/shorten` returns `expires_at` (null for authenticated users, 7 days for guests).
- `POST /api/v1/urls/claim`: transfer anonymous URLs to the user on login/signup.
- Require authentication for `GET /api/v1/{shortCode}/stats` and restrict analytics UI to authenticated users.
- Frontend: login/signup pages, React Router, unified `/` view for guests and users, URL deletion with instant UI update, toast notifications.
- Containerization: switch API runtime image to Distroless static nonroot (ADR 0011).
- CI/CD: API Docker image published to GHCR on version tags.
- Migrations 002-005: users, refresh_tokens, urls.user_id, urls.expires_at.