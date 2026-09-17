# Release Note: 06-Full-Stack Auth App in Staging

Date: 2026-09-03

## Objective
Deploy a lightweight full-stack auth app (register/login) in staging using Go backend, vanilla JS frontend, and shared PostgreSQL, with minimal resource usage on a ~2GB server.

## Actions Performed

### 1. Database: PostgreSQL 16 in `database` namespace
- Single PostgreSQL instance for both staging and production
- Separate logical DBs per environment: `appdb_staging`, `appdb_production`
- Secret-backed credentials
- Local-path volume with 2Gi storage
- Lightweight requests: 128Mi memory, 50m CPU; limits: 256Mi memory, 200m CPU

### 2. Backend: Go API in `staging` namespace
- Endpoints: `POST /api/register`, `POST /api/login`, `GET /api/me`
- JWT auth with bcrypt password hashing
- CORS enabled
- Uses `postgres://app:app_password@postgres.database:5432/appdb_staging`
- Local image: `registry.local/pawradise-backend:local`
- Service exposed on NodePort 30083

### 3. Frontend: lightweight HTML/JS in `staging` namespace
- Single-page auth UI (register/login)
- Proxies API calls to `http://backend.staging:8080`
- Local image: `registry.local/pawradise-frontend:local`
- Service exposed on NodePort 30084

### 4. Firewall
- Opened ports 30083 and 30084 in ufw

## Outcome
Staging environment is fully functional and accessible at:
- Frontend: http://pawradise.ir:30084
- Backend API: http://pawradise.ir:30083

Tested: registration and login flows work. Awaiting your review before production promotion.
