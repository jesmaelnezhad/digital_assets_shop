---
name: store4bots-security
description: Harden BLUE and RED without lockout. Use when preparing a new VM, tightening SSH, firewall, or TLS, or rotating credentials.
---

# Server hardening

Goal: a safe host that **you can still reach**. Never apply a firewall or sshd change that you have not proven on a **second open session**.

## Order (do not skip)

1. Keep an existing SSH session open until the new path is proven.
2. Add your key; confirm `ssh -i … -p $BLUE_SSH_PORT` (or RED) works from a **new** terminal.
3. Only then disable password auth / root password login.
4. Only then restrict firewall to SSH + the ports this role needs.
5. Reboot test only after (2)–(4) work from a fresh client.

Use `scripts/harden-ssh.sh` and `scripts/harden-firewall.sh` rather than ad-hoc `ufw`/`iptables`.

## BLUE (agent host)

Need: `$BLUE_SSH_PORT` (default 2222), outbound HTTPS (git, registries, models). No public 80/443 required.

- SSH key-only, non-default port if you set one.
- Unattended-upgrades or equivalent.
- Docker socket: only the agent user (or root) as designed; do not expose Docker TCP.

## RED (app host)

Need: 22 (or `$RED_SSH_PORT`), 80, 443, and **not** a public Postgres/Mongo unless you explicitly want remote migrate from BLUE.

Current live choice: Postgres `:5432` and Mongo `:27017` are published on `$RED_HOST` so BLUE can migrate. Prefer firewalling those ports to `$BLUE_HOST` only.

k3s NodePort (`$INGRESS_HTTP_NODEPORT`) should **not** be the public HTTP path; host nginx on 443 is.

## Credentials

- Registry htpasswd: pull vs push users (see `scripts/setup-registry.sh`).
- `ADMIN_TOKEN`, DB passwords, JWT secrets: `config/site.secrets.env` only.
- Do not commit live tokens. Do not paste them into docs.

## Lockout recovery

If sshd is misconfigured, use the provider console/VNC. Do not keep iterating firewall rules from a single session that is about to drop.
