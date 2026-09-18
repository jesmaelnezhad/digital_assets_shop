#!/bin/sh
# 40-store4bots-env.sh — runs before nginx starts, generates env.js from env.template.js
set -eu
envsubst '${API_BASE} ${IS_STAGING} ${ENV_NAME}' \
  < /usr/share/nginx/html/env.template.js \
  > /usr/share/nginx/html/env.js
