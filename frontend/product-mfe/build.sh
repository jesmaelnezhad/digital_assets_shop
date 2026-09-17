#!/usr/bin/env bash
set -euo pipefail
cd /root/project/frontend/product-mfe
rm -rf dist && mkdir -p dist/product dist/assets
cp src/assets/favicon.svg dist/assets/ 2>/dev/null || true
cp src/product.html dist/product.html 2>/dev/null || true
cp src/product.html dist/product/product.html 2>/dev/null || true
cp src/bundle.html dist/product/bundle.html 2>/dev/null || true
cp src/request.html dist/product/request.html 2>/dev/null || true
cp /root/project/shared/vendor/alpine.min.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/theme/theme.css dist/assets/ 2>/dev/null || true
cp /root/project/shared/lib/api.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/chrome/app.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/env.template.js dist/ 2>/dev/null || true
cp src/nginx-mfe.conf dist/ 2>/dev/null || true
cp /root/project/shared/docker-entrypoint-env.sh dist/ 2>/dev/null || true
docker build -t 130.185.123.156:30099/pawradise/product-mfe:latest .
echo "Built product-mfe"