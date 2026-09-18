# Deployment (current)

Build on BLUE, push to the **in-cluster** registry on RED, roll Deployments in `staging` then `production`.

Identity: `config/site.env`. Procedure skill: `store4bots-build-deploy`.

## Registry (current)

Namespace `registry`. Manifest `k8s/registry.yaml`.

- Hub images pulled **once**: `docker.io/library/registry:2`, `docker.io/library/nginx:1.27-alpine`.
- ClusterIP only. Traffic is TLS via host nginx.
- HTTPS on `$STAGING_HOST`: `/v2` (Docker Registry API) and `/registry` (rewrite to registry root for k3s).
- nginx sidecar htpasswd: GET/HEAD = pull **or** push user; writes = **push** only. Unauthenticated `GET /v2/` → 401.
- Secret `registry-auth` on the cluster. Same users in `config/site.secrets.env` and `/root/.registry-auth` (mode 600) on BLUE and RED.

k3s `/etc/rancher/k3s/registries.yaml` (pull user only):

```yaml
mirrors:
  "${STAGING_HOST}":
    endpoint:
      - "https://${STAGING_HOST}/registry"
configs:
  "${STAGING_HOST}":
    auth:
      username: k3s-pull
      password: "(REGISTRY_PULL_PASS)"
```

Then `systemctl restart k3s`. Do not set `insecure_skip_verify`. Staging host nginx must allow large bodies (`client_max_body_size 0`) or blob PUTs return 413.

App image name: `$(IMAGE <name>):<tag>` → `$REGISTRY_HOST/$IMAGE_REPO/<name>:<tag>`.

Stock images (ingress-nginx, postgres, mongo, registry, nginx) stay on public registries. If `registry.k8s.io` is blocked, set `INGRESS_NGINX_IMAGE` in `site.env` to a Hub tag or a copy already in **your** registry *after* the registry exists.

## Build Go

Alpine needs a static binary:

```bash
. scripts/lib/site-env.sh
cd services/<svc>
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test ./handlers -count=1
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o <svc>-binary .
docker build -t "$(IMAGE <svc>):<tag>" .
docker push "$(IMAGE <svc>):<tag>"
```

## Build MFE

```bash
IMAGE_TAG="$(IMAGE shop-mfe):<tag>" bash shared/build-mfe.sh shop-mfe
docker push "$(IMAGE shop-mfe):<tag>"
```

`shared/build-mfe.sh` copies `shared/theme`, `shared/lib/{api,ui,events}.js`, `shared/chrome`. Changing those files means rebuild every MFE that should show the change. Bump `theme.css?v=` when CSS changes.

## Deploy

```bash
red_ssh "k3s kubectl -n ${STAGING_NS} set image deploy/<name> <name>=$(IMAGE <name>):<tag>"
red_ssh "k3s kubectl -n ${STAGING_NS} rollout status deploy/<name> --timeout=180s"
```

Or `scripts/build-push-deploy.sh <name> <tag>`.

`imagePullPolicy: IfNotPresent` + the **same tag**: delete the pod or use a new tag.

Do not kubectl-cp binaries into running pods as the normal path.

## Apply cluster YAML

```bash
scripts/render-k8s.sh
# then kubectl apply -f k8s/generated/...  (see k8s/README.md)
```

Never apply leftovers listed in `k8s/README.md`.

## Secrets

`store4bots-secrets` in `staging` and `production`: `JWT_SECRET`, `ADMIN_TOKEN`, `DB_PASSWORD`, `MONGO_URI`. Created by `scripts/prepare-red.sh` from `site.secrets.env`. Do not commit live values.
