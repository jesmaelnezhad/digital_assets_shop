---
name: store4bots-build-deploy
description: Build Store4bots images on BLUE, push to the in-cluster registry, deploy on RED. Use when shipping a service or MFE.
---

# Build and deploy

```bash
. scripts/lib/site-env.sh
# MFE
IMAGE_TAG="$(IMAGE shop-mfe):proto-v10" bash shared/build-mfe.sh shop-mfe
docker push "$(IMAGE shop-mfe):proto-v10"
red_ssh "k3s kubectl -n ${STAGING_NS} set image deploy/shop-mfe shop-mfe=$(IMAGE shop-mfe):proto-v10"
red_ssh "k3s kubectl -n ${STAGING_NS} rollout status deploy/shop-mfe --timeout=180s"

# Go service
cd services/identity-service
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o identity-service-binary .
docker build -t "$(IMAGE identity-service):<tag>" .
docker push "$(IMAGE identity-service):<tag>"
red_ssh "k3s kubectl -n ${STAGING_NS} set image deploy/identity-service identity-service=$(IMAGE identity-service):<tag>"
```

Or `scripts/build-push-deploy.sh identity-service <tag>` / `scripts/build-push-deploy.sh shop-mfe <tag>`.

## Registry

- Login: push user `$REGISTRY_PUSH_USER` at `https://$REGISTRY_HOST` (`/v2`).
- k3s pull: `$REGISTRY_PULL_USER` via `/etc/rancher/k3s/registries.yaml` mirror `https://$STAGING_HOST/registry`.
- Unauthenticated GET `/v2/` must be 401. Pull user cannot PUT.

Stock images (`registry:2`, `nginx`, `postgres`, `mongo`, `rancher/*`) stay on Docker Hub / rancher.

## After deploy

Playwright + `e2e-suite.js` against `https://$STAGING_HOST`. If the UI looks unchanged, force a new tag or delete the pod (`IfNotPresent`).
