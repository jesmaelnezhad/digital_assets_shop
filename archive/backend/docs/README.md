# Backend Documentation

## Overview
Lightweight Go API for user authentication (register/login).

## Tech Stack
- Go 1.24
- PostgreSQL 16 (shared between staging and production)
- JWT for authentication

## Project Structure
```
backend/
├── main.go              # Entry point
├── database/
│   └── db.go            # Database connection
├── models/
│   └── user.go          # User model
├── handlers/
│   └── auth.go          # Register/login handlers
├── middleware/
│   └── auth.go          # JWT middleware
└── docs/
    └── README.md        # This file
```

## Build & Run
```bash
go mod init backend
go get github.com/gin-gonic/gin
go get github.com/golang-jwt/jwt/v5
go get github.com/lib/pq
go build -o api .
./api
```

## API Endpoints
- `POST /api/register` - Register new user
- `POST /api/login` - Login and get JWT token
- `GET /api/me` - Get current user (protected)

## Environment Variables
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT signing

## Deployment
- Image: `registry.example.com/backend:latest`
- Port: 8080
- Replicas: 1
- Resources: 64Mi memory, 50m CPU (requests), 128Mi memory, 200m CPU (limits)
