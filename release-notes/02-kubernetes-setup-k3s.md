# Release Note: 02-Kubernetes Setup (K3s)

Date: 2026-09-03

## Objective
Install a production-ready Kubernetes distribution on the prepared Ubuntu 24.04 server, optimized for low-resource single-node deployment.

## Decision
Chose **K3s** over upstream `kubeadm` because:
- This server has limited resources (~1.9 GB RAM)
- K3s is a CNCF-certified Kubernetes distribution
- K3s ships with containerd, simplifying setup
- K3s is ideal for single-node / edge / small VM workloads

## Actions Performed

### 1. Install K3s
- Install K3s using official script with Kubernetes v1.32
- Disable traefik ingress (will use nginx ingress later)
- Disable local-path storage (will use Longhorn later)
- Configure kubeconfig at `/etc/rancher/k3s/k3s.yaml`

### 2. Verify installation
- Wait for node to become Ready
- Install kubectl and helm client tools
- Verify cluster health

## Outcome
A single-node K3s Kubernetes cluster is running and ready for application deployment.
