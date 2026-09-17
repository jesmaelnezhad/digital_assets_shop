# Release Note: 04-Publishing via NodePort

Date: 2026-09-03

## Objective
Make the landing page externally reachable without a domain or TLS, suitable for this single-node small server.

## Actions Performed

### 1. NodePort services
- Exposed landing-page service in each namespace with dedicated NodePorts:
  - staging: `http://pawradise.ir:30081`
  - production: `http://pawradise.ir:30082`

### 2. Firewall
- Opened ports 30081 and 30082 in ufw so traffic from your laptop can reach the node

## Outcome
Landing page is accessible from outside the cluster via hostname + NodePort in both staging and production.
