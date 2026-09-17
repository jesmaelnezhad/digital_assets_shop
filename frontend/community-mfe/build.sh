#!/bin/bash
set -euo pipefail
cd /root/project/frontend/community-mfe
rm -rf dist && mkdir -p dist/assets
cp src/assets/favicon.svg dist/assets/ 2>/dev/null || true
cp src/index.html dist/ 2>/dev/null || true
cp src/community.html dist/ 2>/dev/null || true
cp src/post.html dist/ 2>/dev/null || true
cp src/profile.html dist/ 2>/dev/null || true
cp /root/project/shared/vendor/alpine.min.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/theme/theme.css dist/assets/ 2>/dev/null || true
cp /root/project/shared/lib/api.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/chrome/app.js dist/assets/ 2>/dev/null || true
cp /root/project/shared/env.template.js dist/ 2>/dev/null || true
cp src/nginx-mfe.conf dist/ 2>/dev/null || true
cp /root/project/shared/docker-entrypoint-env.sh dist/ 2>/dev/null || true
docker build -t 130.185.123.156:30099/pawradise/community-mfe:latest .
echo "Built community-mfe"
