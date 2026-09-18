#!/usr/bin/env bash
# Install BLUE build tools on this machine (Git, Docker, Go 1.22+, Node 20+, Playwright).
# Does not change sshd. After a second key-based SSH session works: scripts/harden-ssh.sh blue
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/site-env.sh
. "$SCRIPT_DIR/lib/site-env.sh"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "run as root on BLUE" >&2
  exit 1
fi

export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y git curl ca-certificates gnupg make python3 python3-venv openssl

if ! command -v docker >/dev/null 2>&1; then
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg
  # shellcheck disable=SC1091
  . /etc/os-release
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu ${VERSION_CODENAME} stable" >/etc/apt/sources.list.d/docker.list
  apt-get update
  apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
fi

if ! command -v go >/dev/null 2>&1 || ! go version | grep -qE 'go1\.(2[2-9]|[3-9][0-9])'; then
  GOVER="${GO_VERSION:-1.22.10}"
  curl -fsSL "https://go.dev/dl/go${GOVER}.linux-amd64.tar.gz" -o /tmp/go.tgz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf /tmp/go.tgz
  ln -sf /usr/local/go/bin/go /usr/local/bin/go
  ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt
fi

if ! command -v node >/dev/null 2>&1; then
  curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
  apt-get install -y nodejs
fi

mkdir -p "${REPO_DIR:-/root/digital_assets_shop}"
if [[ ! -d "$REPO_ROOT/.git" ]]; then
  echo "clone the repo to ${REPO_DIR:-$REPO_ROOT} before running tests" >&2
fi

if [[ -f "$REPO_ROOT/tests/package.json" ]]; then
  (cd "$REPO_ROOT/tests" && npm ci && npx playwright install --with-deps chromium)
fi

umask 077
if [[ -n "${REGISTRY_PUSH_PASS:-}" ]]; then
  cat > /root/.registry-auth <<EOF
REGISTRY_HOST=${REGISTRY_HOST}
REGISTRY_PUSH_USER=${REGISTRY_PUSH_USER}
REGISTRY_PUSH_PASS=${REGISTRY_PUSH_PASS}
REGISTRY_PULL_USER=${REGISTRY_PULL_USER}
REGISTRY_PULL_PASS=${REGISTRY_PULL_PASS}
EOF
  chmod 600 /root/.registry-auth
  if curl -sS -o /dev/null -w '%{http_code}' "https://${REGISTRY_HOST}/v2/" | grep -qx 401; then
    echo "$REGISTRY_PUSH_PASS" | docker login "$REGISTRY_HOST" -u "$REGISTRY_PUSH_USER" --password-stdin
  else
    echo "registry not reachable yet; skip docker login (run again after setup-registry.sh)"
  fi
fi

echo "go $(go version)"
docker version --format '{{.Server.Version}}' >/dev/null
echo "node $(node -v)"
git --version
echo "BLUE packages installed. Harden later with: scripts/harden-ssh.sh blue"
