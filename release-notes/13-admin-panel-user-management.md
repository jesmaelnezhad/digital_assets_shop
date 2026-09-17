# Release Note: 13-Admin Panel for User Management

Date: 2026-09-03

## Purpose
Add an admin panel to manage users: list, delete, and reset passwords.

## Admin API Endpoints (backend)
- `GET /admin/users` — List all users (requires admin auth)
- `DELETE /admin/users/:id` — Delete a user
- `POST /admin/users/:id/reset-password` — Reset password to random string

## Admin Authentication
- Admin endpoints are protected by bearer token
- Token configured via `ADMIN_TOKEN` env var (defaults to `admin-secret-token-change-in-production` if not set)
- When `ADMIN_ENABLED=true`, admin routes are mounted under `/admin`

## Backend Changes
- Added `handlers/admin.go` with:
  - `AdminAuthMiddleware()` for token-based auth
  - `ListUsers()` — queries all users from DB
  - `DeleteUser()` — deletes by ID
  - `ResetPassword()` — generates random password (base64 of 16 random bytes), hashes with bcrypt, updates DB
- Updated `main.go` to conditionally mount admin routes

## Frontend Changes
- Added `public/admin.html` — admin panel UI
- Uses vanilla JS (no framework dependency)
- Features:
  - Login with admin token
  - Table of all users (ID, email, name, created date)
  - Reset password: generates random password, shows it, copies to clipboard on click
  - Delete user: confirmation dialog before deletion
  - Toast notifications for success/error
  - Environment badge (staging/production)
  - Logout functionality

## Nginx Configuration
- Admin page served at `/admin.html`
- Admin API proxied at `/admin-api/` (proxies to `/admin/*` on backend)
- Authorization header forwarded to backend

## Deployment
- Backend: `ADMIN_ENABLED=true` env var set in staging deployment
- Frontend: deployed with admin.html included
- Admin token: `admin-secret-token-change-in-production` (for staging)

## Testing
- Verified admin API directly via curl:
  - `GET /admin/users` returns list of users
  - `POST /admin/users/:id/reset-password` returns new random password
  - `DELETE /admin/users/:id` deletes user or returns 404

## Files Changed
- `/root/project/backend/handlers/admin.go` (new)
- `/root/project/backend/main.go` (updated)
- `/root/project/frontend/public/admin.html` (new)
- `/root/project/frontend/Dockerfile` (updated to include admin.html)
- `/root/project/k8s/staging/frontend-config.yaml` (updated nginx config)
- `/root/project/k8s/staging/backend.yaml` (updated with ADMIN_ENABLED=true and correct DB password)
