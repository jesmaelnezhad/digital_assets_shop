#!/usr/bin/env bash
# Provision RED from BLUE: Docker, host DBs, k3s, namespaces, secrets, ingress-nginx,
# host nginx, registry, then app manifests. Does not disable SSH passwords.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/site-env.sh
. "$SCRIPT_DIR/lib/site-env.sh"

need() { [[ -n "${!1:-}" ]] || { echo "missing \$$1" >&2; exit 1; }; }
need RED_HOST
need STAGING_HOST
need JWT_SECRET
need ADMIN_TOKEN
need POSTGRES_PASSWORD
need MONGO_PASSWORD
: "${STAGING_NS:=staging}"
: "${PRODUCTION_NS:=production}"
: "${REGISTRY_NS:=registry}"
: "${INGRESS_NS:=ingress-nginx}"

echo "1/8 Docker on RED"
# shellcheck disable=SC2087
red_ssh bash -s <<'EOF'
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
if ! command -v docker >/dev/null 2>&1; then
  apt-get update
  apt-get install -y ca-certificates curl gnupg
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  . /etc/os-release
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu ${VERSION_CODENAME} stable" >/etc/apt/sources.list.d/docker.list
  apt-get update
  apt-get install -y docker-ce docker-ce-cli containerd.io
fi
docker info >/dev/null
EOF

echo "2/8 host Postgres + Mongo"
"$SCRIPT_DIR/install-host-db.sh"

echo "3/8 k3s (Traefik off)"
red_ssh bash -s <<'EOF'
set -euo pipefail
if ! command -v k3s >/dev/null 2>&1; then
  curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='--disable traefik' sh -
fi
k3s kubectl get nodes
EOF

echo "4/8 namespaces + secrets"
# shellcheck disable=SC2087
red_ssh bash -s <<EOF
set -euo pipefail
for ns in ${STAGING_NS} ${PRODUCTION_NS} ${REGISTRY_NS} ${INGRESS_NS}; do
  k3s kubectl create namespace "\$ns" --dry-run=client -o yaml | k3s kubectl apply -f -
done
MONGO_URI="mongodb://${MONGO_USER}:${MONGO_PASSWORD}@${MONGO_HOST}:${MONGO_PORT}/${MONGO_DB}?authSource=admin"
for ns in ${STAGING_NS} ${PRODUCTION_NS}; do
  k3s kubectl -n "\$ns" create secret generic store4bots-secrets \\
    --from-literal=JWT_SECRET='${JWT_SECRET}' \\
    --from-literal=ADMIN_TOKEN='${ADMIN_TOKEN}' \\
    --from-literal=DB_PASSWORD='${POSTGRES_PASSWORD}' \\
    --from-literal=MONGO_URI="\$MONGO_URI" \\
    --dry-run=client -o yaml | k3s kubectl apply -f -
done
EOF

echo "5/8 ingress-nginx"
"$SCRIPT_DIR/render-k8s.sh" >/dev/null
for f in ingress-nginx-rbac.yaml ingress-nginx-configmap.yaml ingress-nginx-deploy.yaml ingress-nginx-svc.yaml; do
  red_ssh "k3s kubectl apply -f -" < "$REPO_ROOT/k8s/generated/$f"
done
red_ssh "k3s kubectl -n ${INGRESS_NS} rollout status deploy/ingress-nginx-controller --timeout=180s"

echo "6/8 host nginx + TLS"
"$SCRIPT_DIR/install-host-nginx.sh"

echo "7/8 in-cluster registry"
"$SCRIPT_DIR/setup-registry.sh"

echo "8/8 app manifests"
for f in \
  namespace.yaml \
  manifests/staging-deployments.yaml \
  manifests/production-deployments.yaml \
  frontend-deployments.yaml \
  api-gateway-staging.yaml \
  frontend-ingress-staging.yaml \
  frontend-ingress.yaml \
  manifests/production-ingress.yaml
do
  if [[ -f "$REPO_ROOT/k8s/generated/$f" ]]; then
    red_ssh "k3s kubectl apply -f -" < "$REPO_ROOT/k8s/generated/$f"
  fi
done

echo "RED base is up. Next: scripts/migrate.sh && scripts/seed-staging.sh && scripts/build-push-deploy.sh"
echo "After a second SSH key session works: scripts/harden-ssh.sh red && scripts/harden-firewall.sh red --apply"
