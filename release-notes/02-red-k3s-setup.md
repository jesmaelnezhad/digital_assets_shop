# 02 — RED k3s Setup and Verification

**Date:** 2026-09-04  
**Server:** RED (194.5.206.106)

## What Was Done

### System Hardening (Step 1.3)
- Installed base packages: ca-certificates, curl, gnupg, lsb-release
- Disabled swap: `swapoff -a`, removed swap entries from `/etc/fstab`
- Loaded kernel modules: overlay, br_netfilter
- Created `/etc/sysctl.d/99-k8s.conf` with k8s-required sysctl settings:
  - vm.swappiness=10
  - net.bridge.bridge-nf-call-iptables=1
  - net.bridge.bridge-nf-call-ip6tables=1
  - net.ipv4.ip_forward=1
  - net.ipv6.conf.all.forwarding=1
- Applied all sysctl settings via `sysctl --system`

### containerd Installation (Step 1.4)
- Added Docker apt repository (download.docker.com, jammy stable)
- Installed containerd.io v2.3.4
- Configured containerd: SystemdCgroup=true (required for k3s)
- Enabled and started containerd service

### k3s Installation (Step 1.5)
- Installed k3s v1.36.4+k3s1 from get.k3s.io
- Flags: `--disable traefik --disable local-path-provisioner --write-kubeconfig-mode 644`
- Note: local-path-provisioner was NOT disabled — it was auto-enabled by k3s. Will manage as needed.

### Verification (Step 1.6)
- k3s service: active (running) since 2026-09-04 09:37:20 UTC
- Node: ubuntu---2-vcpu---2-gb-ram, Ready, control-plane, v1.36.4+k3s1
- Internal IP: 194.5.206.106
- Container runtime: containerd://2.3.4-k3s1.36
- kube-system pods (all Running):
  - coredns-54996dc9b4-s59pc (1/1)
  - local-path-provisioner-77b9867795-p9qnw (1/1)
  - metrics-server-6dc596dfb8-jtq8j (1/1)

### Kubeconfig
- Copied `/etc/rancher/k3s/k3s.yaml` to BLUE at `/root/.kube/red-kubeconfig.yaml`
- Updated server address from 127.0.0.1 to 194.5.206.106 for remote access from BLUE

### Helper Tools
- Installed tmux 3.4 (for resilient command execution)
- Installed ufw 0.36.2 (firewall)

## Issues Encountered
- Docker apt repo `/etc/apt/sources.list.d/docker.list` had malformed entry (double space before "stable"). Fixed by rewriting with correct format: `deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu jammy stable`

## Next Steps
- Set up ufw firewall rules on RED
- Create resilient command execution framework (tmux-based)
- Pull application images (postgres already pulled)
- Deploy PostgreSQL, backend, frontend
