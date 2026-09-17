# Release Note: 03-Landing Page Deployment

Date: 2026-09-03

## Objective
Deploy a lightweight landing page to both `staging` and `production` namespaces for testing and promotion workflow.

## Actions Performed

### 1. Created namespaces
- `staging` - for testing changes
- `production` - for live traffic

### 2. Built lightweight landing page
- Used `nginx:alpine` image (lightweight, ~20MB)
- Simple HTML page with environment indicator
- ConfigMap-based content (no custom image needed)
- Minimal resource requests: 16Mi memory, 10m CPU

### 3. Deployed to both namespaces
- Deployment with 1 replica each
- ClusterIP service for internal access
- Exposed via NodePort for external access:
  - Staging: `http://130.185.121.83:30081`
  - Production: `http://130.185.121.83:30082`

## Outcome
Both environments are running and accessible. Staging should be used for testing before promoting to production.
