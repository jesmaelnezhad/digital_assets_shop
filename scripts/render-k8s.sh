#!/usr/bin/env bash
# Copy the apply-set YAML into k8s/generated/ with values from config/site.env.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/site-env.sh
. "$SCRIPT_DIR/lib/site-env.sh"

need() { [[ -n "${!1:-}" ]] || { echo "missing \$$1 in config/site.env" >&2; exit 1; }; }
need STAGING_HOST
need PRODUCTION_HOST
need RED_HOST
need REGISTRY_HOST
need INGRESS_HTTP_NODEPORT
need IMAGE_REPO
: "${INGRESS_NGINX_IMAGE:=registry.k8s.io/ingress-nginx/controller:v1.11.2}"
: "${DB_HOST:=$RED_HOST}"
: "${POSTGRES_PASSWORD:=}"
: "${MONGO_USER:=paw}"
: "${MONGO_PORT:=27017}"
: "${MONGO_DB:=events}"
: "${MONGO_HOST:=$RED_HOST}"
: "${MONGO_PASSWORD:=}"

OUT="$REPO_ROOT/k8s/generated"
rm -rf "$OUT"
mkdir -p "$OUT/manifests"

export RENDER_FILES="$(cat <<'EOF'
k8s/registry.yaml
k8s/ingress-nginx-rbac.yaml
k8s/ingress-nginx-configmap.yaml
k8s/ingress-nginx-deploy.yaml
k8s/ingress-nginx-svc.yaml
k8s/api-gateway-staging.yaml
k8s/frontend-ingress-staging.yaml
k8s/frontend-deployments.yaml
k8s/frontend-ingress.yaml
k8s/namespace.yaml
k8s/manifests/staging-deployments.yaml
k8s/manifests/production-deployments.yaml
k8s/manifests/production-ingress.yaml
EOF
)"

python3 - "$REPO_ROOT" "$OUT" <<'PY'
import os, sys
from pathlib import Path
from urllib.parse import quote

root, out = Path(sys.argv[1]), Path(sys.argv[2])
red = os.environ["RED_HOST"]
mongo_user = os.environ.get("MONGO_USER", "paw")
mongo_pw = quote(os.environ.get("MONGO_PASSWORD", ""), safe="")
mongo_host = os.environ.get("MONGO_HOST") or red
mongo_port = os.environ.get("MONGO_PORT", "27017")
mongo_db = os.environ.get("MONGO_DB", "events")
mongo_uri = f"mongodb://{mongo_user}:{mongo_pw}@{mongo_host}:{mongo_port}/{mongo_db}?authSource=admin"
repl = [
    ("{{STAGING_HOST}}", os.environ["STAGING_HOST"]),
    ("{{PRODUCTION_HOST}}", os.environ["PRODUCTION_HOST"]),
    ("{{RED_HOST}}", red),
    ("{{REGISTRY_HOST}}", os.environ["REGISTRY_HOST"]),
    ("{{INGRESS_HTTP_NODEPORT}}", os.environ["INGRESS_HTTP_NODEPORT"]),
    ("{{IMAGE_REPO}}", os.environ["IMAGE_REPO"]),
    ("{{INGRESS_NGINX_IMAGE}}", os.environ.get("INGRESS_NGINX_IMAGE", "registry.k8s.io/ingress-nginx/controller:v1.11.2")),
    ("{{DB_HOST}}", os.environ.get("DB_HOST") or red),
    ("{{POSTGRES_PASSWORD}}", os.environ.get("POSTGRES_PASSWORD", "")),
    ("{{MONGO_URI}}", mongo_uri),
]
files = [f for f in os.environ["RENDER_FILES"].splitlines() if f.strip()]
for rel in files:
    src = root / rel
    if not src.is_file():
        print(f"skip missing {rel}", file=sys.stderr)
        continue
    text = src.read_text()
    for k, v in repl:
        text = text.replace(k, v)
    dest = out / Path(rel).relative_to("k8s")
    dest.parent.mkdir(parents=True, exist_ok=True)
    dest.write_text(text)
    print(dest.relative_to(root))
PY

echo "wrote $OUT"
