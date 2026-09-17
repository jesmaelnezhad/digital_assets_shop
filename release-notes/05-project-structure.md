# Release Note: 05-Project Structure

Date: 2026-09-03

## Objective
Create a well-documented, lightweight project structure for the full-stack application.

## Actions Performed

### 1. Created directory structure
```
/root/project/
├── backend/           # Go API
│   ├── main.go
│   ├── database/db.go
│   ├── models/user.go
│   ├── handlers/auth.go
│   ├── go.mod
│   └── docs/README.md
├── frontend/          # React app
│   ├── public/index.html
│   ├── src/
│   │   ├── components/
│   │   └── pages/
│   └── docs/
└── k8s/               # Kubernetes manifests
    ├── database/      # Shared PostgreSQL
    ├── staging/
    └── production/
```

### 2. Backend documentation
- Tech stack: Go 1.22, PostgreSQL 16, JWT
- API endpoints: /register, /login, /me
- Build/run instructions
- Resource limits: 64Mi/128Mi memory, 50m/200m CPU

### 3. Frontend documentation
- Lightweight vanilla JS (no React build step needed)
- Environment detection via hostname
- Register/login forms with JWT handling

## Outcome
Project is structured and documented for future maintenance. Backend and frontend code is scaffolded.
