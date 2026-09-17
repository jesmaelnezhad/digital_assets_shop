#!/bin/sh
set -eu
envsubst '${API_BASE} ${IS_STAGING} ${ENV_NAME}' \
  < /usr/share/nginx/html/env.template.js \
  > /usr/share/nginx/html/env.js
exec nginx -g "daemon off;"
