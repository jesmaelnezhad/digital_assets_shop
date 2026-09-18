---
name: store4bots-network
description: Store4bots DNS, TLS, host nginx, ingress-nginx, and registry HTTP paths. Use when debugging routing, SSL, 413 uploads, or /registry /v2 access.
---

# Network, domain, TLS

Traffic: client → `$STAGING_HOST` or `$PRODUCTION_HOST`:443 → **host nginx** (Let’s Encrypt) → `127.0.0.1:$INGRESS_HTTP_NODEPORT` → ingress-nginx → Service by **Host + path**.

Do not curl the NodePort with a raw IP and expect staging: production Ingress may match first. Always use the hostname (or `curl -H "Host: $STAGING_HOST"`).

## DNS

A records for `$STAGING_HOST` and `$PRODUCTION_HOST` → `$RED_HOST`.

## Host nginx

File: `/etc/nginx/sites-available/store4bots-ssl`.

- Port 80: ACME + redirect to HTTPS.
- Staging `server_name $STAGING_HOST`: proxy **all** paths (including `/assets/`, `/registry`, `/v2`) to NodePort. `client_max_body_size 0;` `proxy_request_buffering off;` long timeouts (registry pushes).
- Production `server_name $PRODUCTION_HOST`: may serve `/assets/` from `/var/www/production/assets/`; proxy the rest.

Reload: `nginx -t` then `nginx -s reload` (or `kill -HUP` the master if systemd unit is failed). Write the master PID to `/run/nginx.pid`.

## Ingress paths (staging)

API: `k8s/api-gateway-staging.yaml` — `/api/v1/...` to each service port. Put `/api/v1/admin/events` **before** `/api/v1/admin`.

Frontend: `k8s/frontend-ingress-staging.yaml` — `/` shop-mfe; `/login` `/register` `/auth` auth-mfe; `/product` product-mfe; `/account` `/cart` `/wishlist` `/compare` `/recent` `/guest` `/referrals` account-mfe; `/checkout` checkout-mfe; `/community` `/post` `/profile` `/people` community-mfe; `/admin` admin-app.

Registry: `/registry` (rewrite to registry root for k3s) and `/v2` (Docker API). Longer paths win over `/`.

## TLS

certbot (or equivalent) on RED for both hostnames. Cert path used by host nginx: `/etc/ssl/certs/store4bots.crt` + key. k3s pull uses public CA (Let’s Encrypt) — no `insecure_skip_verify`.
