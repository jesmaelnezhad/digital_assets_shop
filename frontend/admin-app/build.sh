#!/usr/bin/env bash
set -euo pipefail
cd /root/project/frontend/admin-app
rm -rf dist && mkdir -p dist/admin dist/assets
cp src/assets/favicon.svg dist/assets/ 2>/dev/null || true
cp src/index.html dist/ 2>/dev/null || true
cp src/index.html dist/admin.html 2>/dev/null || true
cp src/index.html dist/admin/admin.html 2>/dev/null || true
cp /root/project/shared/vendor/alpine.min.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/theme/theme.css dist/assets/ 2>/dev/null || true
cp /root/project/shared/lib/api.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/chrome/app.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/env.template.js dist/ 2>/dev/null || true
cp /root/project/shared/nginx-mfe.conf dist/ 2>/dev/null || true
cp /root/project/shared/docker-entrypoint-env.sh dist/ 2>/dev/null || true
docker build -t 130.185.123.156:30099/pawradise/admin-app:latest .
echo "Built admin-app"
