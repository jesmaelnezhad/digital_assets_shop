---
name: store4bots-prepare-red
description: Provision RED as the Store4bots serving host (k3s, host Postgres/Mongo, nginx TLS, registry). Use when installing a new RED VM or moving the serving cluster to another server.
---

# Prepare RED

RED serves the site. It runs k3s, host Docker Postgres and Mongo, host nginx (TLS), and an in-cluster image registry.

Moving to a **new VM** is the same procedure: provision new RED, copy data, retarget DNS, then retire the old VM. Do not document old IPs.

## Order (do not skip)

1. SSH keys work as `$RED_SSH_USER@$RED_HOST`. Allow port 22 in the firewall **before** enabling UFW (`scripts/harden-ssh.sh red`).
2. Docker Engine.
3. Host Postgres + Mongo (`scripts/install-host-db.sh`).
4. k3s with Traefik disabled: `curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='--disable traefik' sh -`
5. Namespaces: `staging`, `production`, `registry`, `ingress-nginx` (optional empty `database`).
6. Secret `store4bots-secrets` in `staging` and `production` (`JWT_SECRET`, `ADMIN_TOKEN`).
7. ingress-nginx (canonical YAML in `k8s/`), NodePort **`$INGRESS_HTTP_NODEPORT`** (live default 30758).
8. DNS A records: `$STAGING_HOST` (and `$PRODUCTION_HOST` if used) → `$RED_HOST`.
9. TLS + host nginx (`store4bots-network`). Staging server block proxies **everything** including `/assets/` to `127.0.0.1:$INGRESS_HTTP_NODEPORT`. Unlimited `client_max_body_size` on staging (registry blob uploads).
10. In-cluster registry (`scripts/setup-registry.sh`).
11. `scripts/render-k8s.sh &&` apply canonical manifests.
12. Migrations + seed (`store4bots-databases`).
13. Images from BLUE, `kubectl set image` / apply.

## Host databases (current)

Postgres and Mongo run as **host Docker containers**, not k8s StatefulSets. Pods use `DB_HOST=$RED_HOST` (the node’s address, reachable from CNI via host).

k8s YAML under `k8s/database/` is optional/unused on the current install.

## k3s registries.yaml

After the registry exists, `/etc/rancher/k3s/registries.yaml` mirrors `$STAGING_HOST` to `https://$STAGING_HOST/registry` with the **pull** user only. Then `systemctl restart k3s`.

## Migrate RED to a new server

Use `scripts/migrate-red.sh` (dump Postgres, dump Mongo volume, copy registry PVC or re-push images, install k3s, restore, retarget DNS, confirm TLS). Keep the old server until the new host returns 200 on `https://$STAGING_HOST/`.
