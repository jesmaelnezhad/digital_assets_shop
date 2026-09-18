# Network, DNS, TLS, ingress (current)

Skill: `store4bots-network`. Hosts: `config/site.env`.

## Path of a request

1. DNS A record: `$STAGING_HOST` or `$PRODUCTION_HOST` → `$RED_HOST`.
2. Host nginx `:443` (Let’s Encrypt) on RED.
3. Proxy to `127.0.0.1:$INGRESS_HTTP_NODEPORT` (ingress-nginx Service).
4. Ingress match on **Host** + path → ClusterIP Service → pod.

Do not curl the NodePort by IP and expect staging. Production Ingress may match first. Use the hostname or `curl -H "Host: $STAGING_HOST"`.

## Host nginx

Site file: `/etc/nginx/sites-available/store4bots-ssl` (template: `scripts/templates/store4bots-ssl.conf`).

| `server_name` | Behavior |
|---------------|----------|
| `$STAGING_HOST` | Proxy **all** paths (including `/assets/`, `/registry`, `/v2`) to NodePort. `client_max_body_size 0;` `proxy_request_buffering off;` long timeouts (image pushes). |
| `$PRODUCTION_HOST` | Proxy app traffic; may serve `/assets/` from `/var/www/production/assets/`. |
| `:80` | ACME `/.well-known/acme-challenge/` then redirect to HTTPS. |

Reload: `nginx -t` then `nginx -s reload`. If the systemd unit is failed, `kill -HUP` the master and keep `/run/nginx.pid` accurate.

## Ingress (k3s)

Class `nginx`. Canonical files after render:

- `k8s/frontend-ingress-staging.yaml` — MFE paths, Host=`$STAGING_HOST`
- `k8s/api-gateway-staging.yaml` — `/api/v1/...` to service ports. `/api/v1/admin/events` before `/api/v1/admin`. `/api/v1/recommendations` must hit product-service, not shop `/`.
- `k8s/frontend-ingress.yaml` — production MFEs, Host=`$PRODUCTION_HOST`
- `k8s/manifests/production-ingress.yaml` — production API, Host=`$PRODUCTION_HOST` (not a catch-all)
- Registry Ingress in `k8s/registry.yaml` — `/registry` (rewrite) and `/v2`

Longer prefixes win over `/`.

## TLS

certbot (or equivalent) on RED for the hostnames. Public CA so k3s can pull `https://$STAGING_HOST/registry`.

## Registry HTTP

| Path | Use |
|------|-----|
| `/v2/` | `docker login` / push / pull API (401 without auth) |
| `/registry/...` | k3s mirror endpoint |

Details: `docs/DEPLOYMENT.md`.

## Frontend env

`docker-entrypoint-env.sh` writes `env.js` from `API_BASE`, `ENV_NAME`, `IS_STAGING`. Staging `API_BASE` is `/api/v1`.
