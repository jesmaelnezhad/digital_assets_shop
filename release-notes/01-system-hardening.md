# Release Note: 01-System Hardening

Date: 2026-09-03

## Objective
Prepare the base Ubuntu 24.04 server for running Kubernetes workloads by applying host-level security and performance tuning.

## Actions Performed

### 1. Basic prerequisites
- Confirmed OS: Ubuntu 24.04 LTS (`noble`), amd64
- Installed base packages: `ca-certificates`, `curl`, `apt-transport-https`, `gpg`
- Verified cgroup/cgroup2 support in kernel

### 2. Kernel and networking sysctls
- Disabled swap temporarily and removed swap entry from `/etc/fstab`
- Applied sysctl tuning via `/etc/sysctl.d/99-k8s.conf`:
  - `vm.swappiness=10`
  - `net.ipv4.ip_forward=1`
  - `net.bridge.bridge-nf-call-iptables=1`
  - `net.bridge.bridge-nf-call-ip6tables=1`
  - `net.ipv6.conf.all.forwarding=1`
- Loaded kernel modules: `overlay`, `br_netfilter`
- Verified forwarding/bridge settings via `sysctl`

### 3. Container runtime preparation
- Added Docker's official apt repo and keyring to `/etc/apt/sources.list.d/docker.list`
- Installed `containerd.io` as the container runtime

## Outcome
The server is now prepared with required kernel parameters and containerd installed for Kubernetes.
