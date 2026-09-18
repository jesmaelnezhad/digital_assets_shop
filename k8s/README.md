# Kubernetes manifests

Apply **generated** files only: `scripts/render-k8s.sh` then `k3s kubectl apply -f k8s/generated/…` on RED.

Hosts and NodePort come from `config/site.env`. Do not apply leftovers beside this list.

## Apply set (this install)

| File | Why |
|------|-----|
| `k8s/registry.yaml` | in-cluster registry + `/registry` `/v2` Ingress |
| `k8s/ingress-nginx-rbac.yaml` | ingress-nginx RBAC |
| `k8s/ingress-nginx-configmap.yaml` | controller config |
| `k8s/ingress-nginx-deploy.yaml` | controller Deployment |
| `k8s/ingress-nginx-svc.yaml` | NodePort `$INGRESS_HTTP_NODEPORT` |
| `k8s/api-gateway-staging.yaml` | staging `/api/v1/…` |
| `k8s/frontend-ingress-staging.yaml` | staging MFE paths |
| `k8s/frontend-deployments.yaml` | seven MFEs (staging + production) |
| `k8s/manifests/staging-deployments.yaml` | staging Go services |
| `k8s/manifests/production-deployments.yaml` | production Go services |
| `k8s/frontend-ingress.yaml` | production MFE paths |
| `k8s/manifests/production-ingress.yaml` | production API |

`scripts/setup-registry.sh` and `scripts/prepare-red.sh` apply this set after render.

## Do not apply (leftovers)

Single-file copies and old stacks that are **not** in the apply set:

- `k8s/staging/**`
- `k8s/database/**`
- `k8s/postgres.yaml`, `k8s/redis.yaml`
- `k8s/*-mfe-deployment.yaml`, `k8s/shell-mfe-deployment.yaml`, `k8s/admin-mfe-deployment.yaml`
- `k8s/*-service-deployment.yaml` (use `k8s/manifests/*-deployments.yaml`)
- `k8s/nginx-api-gateway*`, `k8s/nginx-frontend.conf`
- `k8s/ingress.yaml`, `k8s/ingress-nginx.yaml`, `k8s/staging-ingress.yaml`, `k8s/staging-ingress-staging-prefix.yaml`, `k8s/production-ingress.yaml`
- `k8s/staging-services.yaml`, `k8s/production-services.yaml`
- `k8s/manifests/staging-services.yaml`, `k8s/manifests/production-services.yaml`, `k8s/manifests/staging-ingress.yaml`
- `k8s/secrets.yaml` (placeholders only — live secrets come from `prepare-red.sh`)
- `k8s/configmap.yaml`, `k8s/ingress/prod/**`

Host DBs are Docker on RED, not these YAMLs.
