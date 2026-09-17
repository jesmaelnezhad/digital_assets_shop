#!/usr/bin/env bash
set -euo pipefail
cd /root/project/frontend/account-mfe
rm -rf dist && mkdir -p dist/assets
cp src/assets/favicon.svg dist/assets/ 2>/dev/null || true
cp src/index.html dist/ 2>/dev/null || true
cp src/account.html dist/ 2>/dev/null || true
cp src/cart.html dist/ 2>/dev/null || true
cp src/wishlist.html dist/ 2>/dev/null || true
cp src/referrals.html dist/ 2>/dev/null || true
cp /root/project/shared/vendor/alpine.min.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/theme/theme.css dist/assets/ 2>/dev/null || true
cp /root/project/shared/lib/api.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/chrome/app.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/env.template.js dist/ 2>/dev/null || true
cp /root/project/shared/nginx-mfe.conf dist/ 2>/dev/null || true
cp /root/project/shared/docker-entrypoint-env.sh dist/ 2>/dev/null || true
docker build -t 130.185.123.156:30099/pawradise/account-mfe:latest .
echo "Built account-mfe"
