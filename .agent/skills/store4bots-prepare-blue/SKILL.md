---
name: store4bots-prepare-blue
description: Provision BLUE as the Store4bots build host (Go, Docker, git, Node, SSH). Use when installing a new BLUE VM, repairing the build box, or cloning the repo for the first time.
---

# Prepare BLUE

BLUE compiles Go, builds Docker images, runs unit/API/Playwright tests, and pushes images to the registry on `$STAGING_HOST`.

## Required packages

Ubuntu 24.04 recommended. As root:

```bash
. scripts/lib/site-env.sh
apt-get update
apt-get install -y git curl ca-certificates gnupg make python3 python3-venv
# Docker (official repo) + compose plugin
# Go 1.22.x
# Node 20+ (for Playwright and e2e-suite.js)
```

Install Docker so the `docker` group can run without a registry listed as insecure HTTP. Pushes use **HTTPS** to `$REGISTRY_HOST`.

Install Go 1.22+ (`go version` must print go1.22 or newer).

Clone to `$REPO_DIR` (default `/root/digital_assets_shop`).

```bash
git clone <repo-url> "$REPO_DIR"
cd "$REPO_DIR"
cp config/site.env.example config/site.env   # then edit
cp config/site.secrets.env.example config/site.secrets.env
```

`npm ci` in `tests/` and `npx playwright install chromium`.

## Docker login (push)

After the registry exists on RED:

```bash
. scripts/lib/site-env.sh
echo "$REGISTRY_PUSH_PASS" | docker login "$REGISTRY_HOST" -u "$REGISTRY_PUSH_USER" --password-stdin
```

Credentials file on the box: `/root/.registry-auth` (mode 600), same variables as `site.secrets.env`.

## SSH to RED

BLUE must be able to `ssh -p $RED_SSH_PORT $RED_SSH_USER@$RED_HOST` (deploy, kubectl). Install an SSH key; do not disable key auth.

## Hardening

Run `scripts/harden-ssh.sh blue` after SSH keys work. See `store4bots-security`.

## Checks

- `go version`, `docker version`, `node -v`, `git --version`
- `docker login $REGISTRY_HOST` succeeds (once registry is up)
- Playwright: `cd tests && npx playwright test --list`
