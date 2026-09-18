#!/usr/bin/env bash
# In-cluster registry on RED: apply k8s/registry.yaml, htpasswd secret, k3s pull mirror.
# Unauthenticated GET https://$REGISTRY_HOST/v2/ must be 401.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/site-env.sh
. "$SCRIPT_DIR/lib/site-env.sh"

need() { [[ -n "${!1:-}" ]] || { echo "missing \$$1" >&2; exit 1; }; }
need STAGING_HOST
need REGISTRY_HOST
need REGISTRY_PULL_USER
need REGISTRY_PUSH_USER
need REGISTRY_PULL_PASS
need REGISTRY_PUSH_PASS

"$SCRIPT_DIR/render-k8s.sh" >/dev/null

echo "applying registry manifests"
red_ssh "k3s kubectl apply -f -" < "$REPO_ROOT/k8s/generated/registry.yaml"

hash_apr1() { openssl passwd -apr1 "$1"; }
PULL_HASH="$(hash_apr1 "$REGISTRY_PULL_PASS")"
PUSH_HASH="$(hash_apr1 "$REGISTRY_PUSH_PASS")"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
printf '%s:%s\n%s:%s\n' "$REGISTRY_PULL_USER" "$PULL_HASH" "$REGISTRY_PUSH_USER" "$PUSH_HASH" >"$TMP/all"
printf '%s:%s\n' "$REGISTRY_PUSH_USER" "$PUSH_HASH" >"$TMP/push"

echo "applying registry-auth secret"
python3 - "$TMP/all" "$TMP/push" <<'PY' | red_ssh "k3s kubectl apply -f -"
import base64, pathlib, sys
all_b = pathlib.Path(sys.argv[1]).read_bytes()
push_b = pathlib.Path(sys.argv[2]).read_bytes()
print("apiVersion: v1")
print("kind: Secret")
print("metadata:")
print("  name: registry-auth")
print("  namespace: registry")
print("type: Opaque")
print("data:")
print("  all:", base64.b64encode(all_b).decode())
print("  push:", base64.b64encode(push_b).decode())
PY

echo "writing /etc/rancher/k3s/registries.yaml (pull user only)"
python3 - <<'PY' | red_ssh "install -d -m 0700 /etc/rancher/k3s && umask 077 && cat > /etc/rancher/k3s/registries.yaml"
import json, os
host = os.environ["STAGING_HOST"]
user = os.environ["REGISTRY_PULL_USER"]
pw = os.environ["REGISTRY_PULL_PASS"]
print("mirrors:")
print(f"  {json.dumps(host)}:")
print("    endpoint:")
print(f"      - {json.dumps('https://' + host + '/registry')}")
print("configs:")
print(f"  {json.dumps(host)}:")
print("    auth:")
print(f"      username: {json.dumps(user)}")
print(f"      password: {json.dumps(pw)}")
PY
red_ssh "systemctl restart k3s"

echo "waiting for registry pod"
red_ssh "k3s kubectl -n ${REGISTRY_NS:-registry} rollout status deploy/registry --timeout=180s"

echo "probe unauthenticated /v2/ (expect 401)"
code="$(curl -sS -o /dev/null -w '%{http_code}' "https://${REGISTRY_HOST}/v2/" || true)"
if [[ "$code" != "401" ]]; then
  echo "GET https://${REGISTRY_HOST}/v2/ returned ${code:-curl-fail}, expected 401" >&2
  echo "Check host nginx proxies /v2, ingress, and registry-auth." >&2
  exit 1
fi
echo "registry ready (401 without auth)"
