# RED Server Image Import Freeze — Operational Notes

## Problem

Piping `docker save` through SSH to `k3s ctr images import` on the 2GB RED server (194.5.206.106) causes the terminal and sometimes the entire server to become unresponsive.

## Root Cause

- RED has ~2GB RAM running k3s + PostgreSQL + 8 pods (backend×2, frontend×2, postgres, coredns, local-path-provisioner, metrics-server)
- `docker save` reads the image from Docker storage on BLUE (can be 10-35MB)
- The pipe transfers data through SSH to RED's `k3s ctr images import`
- Both operations consume RAM simultaneously on an already-loaded server
- When RAM is exhausted, the OOM killer may trigger, or operations simply time out
- Result: terminal hangs, even `echo` or `ls` commands block

## Symptoms

- Terminal appears to hang after starting the docker save | ssh command
- Commands queued after it (like `echo DONE`) don't execute
- SSH session becomes unresponsive
- Other terminal commands (ls, echo, cd) also hang
- Usually recovers spontaneously after 30-120 seconds when the import completes or OOM recovery happens
- In severe cases, the server needs a reboot to recover

## Safe Alternatives (in order of preference)

### 1. Build directly on RED via SSH (BEST)

If Go toolchain is available on RED, build the binary and Docker image there:

```bash
ssh root@194.5.206.106 "
  cd /root/project/backend &&
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o backend-static . &&
  docker build -t registry.local/pawradise-backend:latest -f Dockerfile .
"
```

This avoids transferring the image over SSH entirely.

### 2. Use k3s ctr images pull (for registry-based images)

If the image is in a registry accessible from RED:

```bash
ssh root@194.5.206.106 "k3s ctr images pull docker.io/library/postgres:16-alpine"
```

### 3. Inline the import in a single SSH command (acceptable)

Instead of piping, do the save+import entirely within one SSH session:

```bash
ssh root@194.5.206.106 "
  docker save registry.local/pawradise-backend:latest | k3s ctr images import -
"
```

This keeps the pipe local to RED's containerd, avoiding the SSH transfer overhead. But the image still needs to get to RED's Docker first (e.g., via a prior scp of a tarball).

### 4. docker save | ssh k3s ctr images import (LAST RESORT — can freeze)

```bash
docker save registry.local/pawradise-backend:latest | \
  ssh -o ConnectTimeout=5 -o ServerAliveInterval=10 root@194.5.206.106 \
  "k3s ctr images import - 2>&1 && echo IMPORTED"
```

**Only use this when:**
- The image is already built on BLUE's Docker
- You cannot build on RED (no Go toolchain)
- The image is small (< 15MB ideally)
- You are NOT running other heavy operations simultaneously
- The server has enough free RAM (check with `ssh root@RED free -m` first)
- You accept the risk of terminal freeze

## Recovery if Frozen

1. **Wait** — often recovers on its own in 30-120 seconds
2. **Check from another terminal** if available:
   ```bash
   ssh root@194.5.206.106 "echo test"
   ```
3. **If truly stuck**, the server may need a reboot:
   ```bash
   ssh root@194.5.206.106 "reboot"
   ```
   Then wait ~2 minutes for k3s + PostgreSQL + nginx to come back up.

## Prevention

- Check free RAM before importing: `ssh root@RED "free -m"` — if available is < 500MB, wait or use alternative method
- Build on RED when possible (avoids transfer entirely)
- Keep Docker images small (multi-stage builds, alpine bases)
- Don't run multiple image imports in parallel on the 2GB server
- If the server froze during an import, check pod status after recovery:
  ```bash
  ssh root@RED "k3s kubectl get pods -A"
  ```
  Pods that were running before the freeze should still be running (k3s is resilient).
