# Release Note: 09-Fix White Page / SPA Routing

Date: 2026-09-03

## Issue
Visiting http://staging.pawradise.ir:30084/ showed a white page.

## Root Cause
Nginx default config returned 404 for non-root paths or SPA navigation. React needs `index.html` served for all routes.

## Fix
Updated `frontend-nginx-config` ConfigMap in `staging`:
```nginx
location / {
  try_files $uri $uri/ /index.html;
}
```

Rolled out new frontend pods.

## Verification
- Frontend reachable and serves React app
- External E2E tests pass again
