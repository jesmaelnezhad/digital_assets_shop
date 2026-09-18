#!/usr/bin/env bash
# Install host nginx on RED and write store4bots-ssl from the template + site.env.
# Requests certbot certificates when they are missing.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/site-env.sh
. "$SCRIPT_DIR/lib/site-env.sh"

need() { [[ -n "${!1:-}" ]] || { echo "missing \$$1" >&2; exit 1; }; }
need STAGING_HOST
need PRODUCTION_HOST
need INGRESS_HTTP_NODEPORT
EMAIL="${LETSENCRYPT_EMAIL:-}"

CONF="$(mktemp)"
trap 'rm -f "$CONF"' EXIT
python3 - "$SCRIPT_DIR/templates/store4bots-ssl.conf" "$CONF" <<'PY'
import os, pathlib, sys
src, dest = sys.argv[1], sys.argv[2]
text = pathlib.Path(src).read_text()
for k in ("STAGING_HOST", "PRODUCTION_HOST", "INGRESS_HTTP_NODEPORT"):
    text = text.replace("${" + k + "}", os.environ[k])
pathlib.Path(dest).write_text(text)
PY

# shellcheck disable=SC2087
red_ssh bash -s <<'EOF'
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y nginx certbot
mkdir -p /var/www/certbot /var/www/production/assets /etc/nginx/sites-available /etc/nginx/sites-enabled
rm -f /etc/nginx/sites-enabled/default
EOF

python3 -c "import sys; sys.stdout.buffer.write(open('$CONF','rb').read())" | red_ssh "cat > /etc/nginx/sites-available/store4bots-ssl"
red_ssh "ln -sfn /etc/nginx/sites-available/store4bots-ssl /etc/nginx/sites-enabled/store4bots-ssl"

# Bootstrap HTTP-only until certs exist (template references live cert paths).
red_ssh bash -s <<EOF
set -euo pipefail
if [[ ! -f /etc/letsencrypt/live/${STAGING_HOST}/fullchain.pem ]]; then
  cat > /etc/nginx/sites-available/store4bots-ssl <<'NGX'
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name ${STAGING_HOST} ${PRODUCTION_HOST};
    location /.well-known/acme-challenge/ { root /var/www/certbot; }
    location / { return 200 'store4bots acme\\n'; add_header Content-Type text/plain; }
}
NGX
fi
nginx -t
if systemctl is-active --quiet nginx; then
  nginx -s reload || kill -HUP "\$(cat /run/nginx.pid)"
else
  systemctl start nginx || nginx
fi
pidof nginx | awk '{print \$1}' >/run/nginx.pid || true
EOF

if [[ -n "$EMAIL" ]]; then
  red_ssh bash -s <<EOF
set -euo pipefail
if [[ ! -d /etc/letsencrypt/live/${STAGING_HOST} ]]; then
  certbot certonly --webroot -w /var/www/certbot -d '${STAGING_HOST}' --agree-tos -m '${EMAIL}' --non-interactive
fi
if [[ -n '${PRODUCTION_HOST}' && ! -d /etc/letsencrypt/live/${PRODUCTION_HOST} ]]; then
  certbot certonly --webroot -w /var/www/certbot -d '${PRODUCTION_HOST}' --agree-tos -m '${EMAIL}' --non-interactive || true
fi
EOF
fi

# Restore full TLS site file now that paths may exist.
python3 -c "import sys; sys.stdout.buffer.write(open('$CONF','rb').read())" | red_ssh "cat > /etc/nginx/sites-available/store4bots-ssl"
if red_ssh "test -f /etc/letsencrypt/live/${STAGING_HOST}/fullchain.pem"; then
  red_ssh "nginx -t && (nginx -s reload || kill -HUP \$(cat /run/nginx.pid))"
else
  echo "TLS certs not present yet. Point DNS at \$RED_HOST, set LETSENCRYPT_EMAIL, re-run this script." >&2
fi
echo "host nginx site: /etc/nginx/sites-available/store4bots-ssl"
