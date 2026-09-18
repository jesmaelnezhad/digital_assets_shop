#!/usr/bin/env bash
# Build on this machine (BLUE), push to https://$REGISTRY_HOST, roll the staging Deployment.
# Usage: scripts/build-push-deploy.sh <name> <tag> [namespace]
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/site-env.sh
. "$SCRIPT_DIR/lib/site-env.sh"

NAME="${1:?usage: build-push-deploy.sh <service-or-mfe> <tag> [namespace]}"
TAG="${2:?usage: build-push-deploy.sh <name> <tag> [namespace]}"
NS="${3:-${STAGING_NS:-staging}}"
IMAGE_REF="$(IMAGE "$NAME"):${TAG}"
export GOFLAGS="${GOFLAGS:--buildvcs=false}"

need() { [[ -n "${!1:-}" ]] || { echo "missing \$$1" >&2; exit 1; }; }
need REGISTRY_HOST
need REGISTRY_PUSH_USER
need REGISTRY_PUSH_PASS

echo "$REGISTRY_PUSH_PASS" | docker login "$REGISTRY_HOST" -u "$REGISTRY_PUSH_USER" --password-stdin

MFES="shop-mfe auth-mfe product-mfe account-mfe checkout-mfe community-mfe admin-app"
is_mfe=0
for m in $MFES; do
  [[ "$m" == "$NAME" ]] && is_mfe=1
done

if [[ "$is_mfe" -eq 1 ]]; then
  IMAGE_TAG="$IMAGE_REF" bash "$REPO_ROOT/shared/build-mfe.sh" "$NAME"
else
  dir="$REPO_ROOT/services/$NAME"
  [[ -d "$dir" ]] || { echo "unknown target $NAME" >&2; exit 1; }
  (
    cd "$dir"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test ./handlers -count=1
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "${NAME}-binary" .
    docker build -t "$IMAGE_REF" .
  )
fi

docker push "$IMAGE_REF"
red_ssh "k3s kubectl -n ${NS} set image deploy/${NAME} ${NAME}=${IMAGE_REF}"
red_ssh "k3s kubectl -n ${NS} rollout status deploy/${NAME} --timeout=180s"
echo "deployed ${IMAGE_REF} in ${NS}"
