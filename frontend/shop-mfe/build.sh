#!/usr/bin/env bash
set -euo pipefail
cd /root/project/frontend/shop-mfe
rm -rf dist && mkdir -p dist/assets
cp src/index.html dist/ 2>/dev/null || true
cp src/home.html dist/ 2>/dev/null || true
cp src/category.html dist/ 2>/dev/null || true
mkdir -p dist/assets
cp src/assets/favicon.svg dist/assets/ 2>/dev/null || true
cp /root/project/shared/vendor/alpine.min.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/theme/theme.css dist/assets/ 2>/dev/null || true
cp /root/project/shared/lib/api.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/chrome/app.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/env.template.js dist/ 2>/dev/null || true
cp /root/project/shared/nginx-mfe.conf dist/ 2>/dev/null || true
cp /root/project/shared/docker-entrypoint-env.sh dist/ 2>/dev/null || true
docker build -t 130.185.123.156:30099/pawradise/shop-mfe:latest .
echo "Built shop-mfe"
