---
name: store4bots-kubernetes
description: Store4bots k3s namespaces, Deployments, Services, and Ingress. Use when applying manifests, debugging pods, or changing cluster topology.
---

# Kubernetes

k3s on RED. Traefik disabled. Canonical YAML is under `k8s/` after `scripts/render-k8s.sh` (replaces `{{STAGING_HOST}}`, `{{PRODUCTION_HOST}}`, `{{RED_HOST}}`, `{{REGISTRY_HOST}}`, `{{INGRESS_HTTP_NODEPORT}}`).

Apply the **generated** files in `k8s/generated/`, not leftovers in `k8s/archive/`.

## Namespaces

| NS | Role |
|----|------|
| `ingress-nginx` | ingress-nginx-controller, Service NodePort `$INGRESS_HTTP_NODEPORT` |
| `staging` | 9 backends + 7 MFEs + API/frontend Ingress (Host=`$STAGING_HOST`) |
| `production` | backends + 7 MFEs + Ingress (Host=`$PRODUCTION_HOST` or `*`) |
| `registry` | registry:2 + nginx auth proxy, ClusterIP, Ingress `/registry` and `/v2` |
| `database` | unused on the current install (DBs are host Docker) |

## Live apply set

- `k8s/registry.yaml`
- `k8s/ingress-nginx-*.yaml` (deploy, svc with nodePort `$INGRESS_HTTP_NODEPORT`, rbac, configmap)
- `k8s/manifests/staging-deployments.yaml` + `production-deployments.yaml`
- `k8s/frontend-deployments.yaml` (fix: staging `API_BASE=/api/v1`)
- `k8s/api-gateway-staging.yaml`
- `k8s/frontend-ingress-staging.yaml`
- `k8s/frontend-ingress.yaml` / production API ingress

## Selectors

Service `selector` must be `app: <name>` only. A `namespace:` key in the selector makes Endpoints empty (503).

## Images

Built: `{{REGISTRY_HOST}}/{{IMAGE_REPO}}/<name>:<tag>`. Stock: `docker.io/library/...`, `rancher/...`.
