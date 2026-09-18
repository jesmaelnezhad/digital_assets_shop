# Bring-up and deploy scripts

Run from the repo root on BLUE after `. scripts/lib/site-env.sh` (each script sources that itself). Hosts live in `config/site.env`; passwords in `config/site.secrets.env`.

| Script | What |
|--------|------|
| `prepare-blue.sh` | Git, Docker, Go 1.22+, Node 20+, Playwright on this machine |
| `prepare-red.sh` | Docker, DBs, k3s, ingress-nginx, host nginx, registry, app YAML on RED |
| `harden-ssh.sh` | Key-only sshd drop-in (`blue` or `red`). Pass `--apply` only after a second session works |
| `harden-firewall.sh` | UFW. `--apply` to enable |
| `install-host-db.sh` | Host Docker Postgres 16 + Mongo 7 and the 16 `appdb_*` databases |
| `install-host-nginx.sh` | `/etc/nginx/sites-available/store4bots-ssl` + certbot |
| `setup-registry.sh` | In-cluster registry, `registry-auth`, k3s `registries.yaml`, prove `/v2/` is 401 |
| `render-k8s.sh` | Write `k8s/generated/` from placeholders in `k8s/` |
| `migrate.sh` | Apply `services/<svc>/migrations/` |
| `seed-staging.sh` | Staging SQL under `scripts/seed-staging-*.sql` |
| `build-push-deploy.sh` | Build, push, `kubectl set image` in staging |
| `migrate-red.sh` | Dump or restore Postgres/Mongo when moving RED |

`mirror-host-registry.py` copies blobs from an old host `:30099` registry into the in-cluster registry. Use only for that cutover.

Template: `scripts/templates/store4bots-ssl.conf`.
