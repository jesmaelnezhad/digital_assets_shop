# Release Note: 07-ReactJS Frontend + Shared PostgreSQL

Date: 2026-09-03

## Objective
Replace the lightweight HTML/JS frontend with a real ReactJS app while keeping Go backend and PostgreSQL database, deployed in `staging` only.

## Actions Performed

### 1. ReactJS frontend
- Built React 18 app with Vite in `/root/project/frontend`
- Auth UI with register/login forms, JWT handling, environment badge
- API calls to `http://backend.staging:8080/api`
- Built production bundle: `dist/index.html` + `dist/assets/index-*.js`
- Lightweight Docker image: `registry.local/pawradise-frontend:react`
- Mounted nginx config ConfigMap for serving `/index.html`

### 2. Go backend
- No changes to backend code
- Existing endpoints: `POST /api/register`, `POST /api/login`, `GET /api/me`
- JWT auth with bcrypt password hashing
- Image: `registry.local/pawradise-backend:local`

### 3. Shared PostgreSQL in `database` namespace
- Single PostgreSQL 16 instance with separate logical databases:
  - `appdb_staging` (used by staging backend)
  - `appdb_production` (reserved for production)
- Secret-backed credentials
- Resources: 128Mi/256Mi memory, 50m/200m CPU

### 4. Staging deployment
- Namespace: `staging`
- Backend NodePort: 30083
- Frontend NodePort: 30084
- Firewall ports opened: 30083, 30084

## Outcome
Complete full-stack auth app (React + Go + PostgreSQL) is live in staging:
- Frontend: http://pawradise.ir:30084
- Backend API: http://pawradise.ir:30083

Tested: registration and login flows work. Awaiting your review before production promotion.
