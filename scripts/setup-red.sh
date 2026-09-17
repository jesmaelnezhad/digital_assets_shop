#!/bin/bash
# setup-red.sh — Complete RED server provisioning for Pawradise (v3)
# Usage: ./setup-red.sh [JWT_SECRET] [ADMIN_TOKEN] [DB_PASSWORD]
# Assumes: Ubuntu 24.04, Docker installed, k3s not yet installed

set -euo pipefail

RED_IP="${1:-$(hostname -I | awk '{print $1}')}"
JWT_SECRET="${2:-$(openssl rand -hex 32)}"
ADMIN_TOKEN="${3:-admin_secret_2026_prod}"
DB_PASSWORD="${4:-CHANGE_ME_IN_PRODUCTION}"

echo "=========================================="
echo " Pawradise RED Server Setup v3"
echo "=========================================="
echo "IP: $RED_IP"
echo "=========================================="

# ============================================
# 1. SECURITY HARDENING
# ============================================
echo ""
echo "=== Phase 1: Security Hardening ==="

export DEBIAN_FRONTEND=noninteractive

apt update && apt upgrade -y
apt install -y \
    ufw fail2ban \
    unattended-upgrades \
    auditd \
    curl wget jq \
    gnupg lsb-release \
    software-properties-common \
    docker.io \
    nginx-core \
    postgresql-client

# UFW
ufw --force reset
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 30099/tcp  # Local registry
ufw allow 5432/tcp   # PostgreSQL
ufw --force enable

# Fail2ban
cat > /etc/fail2ban/jail.local << 'FJAIL'
[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
FJAIL
systemctl enable fail2ban && systemctl restart fail2ban

# Security updates
cat > /etc/apt/apt.conf.d/20auto-upgrades <<'APT'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
APT

# SSH hardening
sed -i 's/#PermitRootLogin yes/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config
sed -i 's/#MaxAuthTries 6/MaxAuthTries 3/' /etc/ssh/sshd_config
sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sed -i 's/X11Forwarding yes/X11Forwarding no/' /etc/ssh/sshd_config
sed -i 's/#ClientAliveInterval 0/ClientAliveInterval 300/' /etc/ssh/sshd_config
sed -i 's/#ClientAliveCountMax 3/ClientAliveCountMax 2/' /etc/ssh/sshd_config
systemctl restart sshd

# Sysctl
cat > /etc/sysctl.d/99-pawradise.conf << 'SYSCTL'
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1
net.ipv4.conf.all.accept_redirects = 0
net.ipv4.conf.default.accept_redirects = 0
net.ipv4.conf.all.send_redirects = 0
net.ipv4.conf.default.send_redirects = 0
net.ipv4.conf.all.accept_source_route = 0
net.ipv4.conf.default.accept_source_route = 0
net.ipv4.conf.all.log_martians = 1
net.ipv4.icmp_echo_ignore_broadcasts = 1
net.ipv4.icmp_ignore_bogus_error_responses = 1
net.ipv4.tcp_syncookies = 1
net.ipv6.conf.all.accept_redirects = 0
net.ipv6.conf.default.accept_redirects = 0
kernel.randomize_va_space = 2
kernel.kptr_restrict = 2
kernel.dmesg_restrict = 1
kernel.yama.ptrace_scope = 1
fs.suid_dumpable = 0
SYSCTL
sysctl --system

systemctl disable --now snapd 2>/dev/null || true
systemctl disable --now ModemManager 2>/dev/null || true

echo "Security hardening complete."

# ============================================
# 2. DOCKER CONFIG
# ============================================
echo ""
echo "=== Phase 2: Docker ==="

systemctl enable docker
systemctl start docker

cat > /etc/docker/daemon.json << DOCKERD
{
    "userns-remap": "default",
    "no-new-privileges": true,
    "icc": false,
    "iptables": true,
    "userland-proxy": false,
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "10m",
        "max-file": "3"
    },
    "insecure-registries": ["${RED_IP}:30099"],
    "default-address-pools": [
        {
            "base": "172.17.0.0/16",
            "size": 24
        }
    ]
}
DOCKERD
systemctl restart docker
echo "Docker ready."

# ============================================
# 3. K3S
# ============================================
echo ""
echo "=== Phase 3: K3s ==="

if ! which k3s >/dev/null 2>&1; then
    curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='--disable traefik' sh -
    sleep 15
fi

export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
mkdir -p ~/.kube
cp /etc/rancher/k3s/k3s.yaml ~/.kube/config

# k3s registry config for insecure local registry
mkdir -p /etc/rancher/k3s
cat > /etc/rancher/k3s/registries.yaml << EOF
mirrors:
  ${RED_IP}:30099:
    endpoint:
      - http://${RED_IP}:30099
configs:
  ${RED_IP}:30099:
    tls:
      insecure_skip_verify: true
EOF
systemctl restart k3s
sleep 15

echo "K3s ready: $(k3s kubectl get nodes 2>&1 | tail -1)"

# ============================================
# 4. LOCAL REGISTRY
# ============================================
echo ""
echo "=== Phase 4: Local Registry ==="

if ! docker ps | grep -q registry; then
    docker run -d \
        --name registry \
        --restart always \
        -p 30099:5000 \
        -v registry-data:/var/lib/registry \
        --memory 64m \
        registry:2
fi

echo "Registry running on port 30099."

# ============================================
# 5. POSTGRESQL
# ============================================
echo ""
echo "=== Phase 5: PostgreSQL ==="

if ! docker ps | grep -q postgres; then
    docker run -d \
        --name postgres \
        --restart always \
        -p 5432:5432 \
        -v pgdata:/var/lib/postgresql/data \
        -e POSTGRES_USER=app \
        -e POSTGRES_PASSWORD=$DB_PASSWORD \
        -e POSTGRES_DB=appdb_identity_production \
        --memory 256m \
        postgres:16-alpine

    sleep 5
fi

# Create logical databases
SERVICES="identity product commerce community review payment admin media"
for svc in $SERVICES; do
    docker exec postgres psql -U app -d appdb_identity_production -c "CREATE DATABASE appdb_${svc}_production;" 2>/dev/null || true
    docker exec postgres psql -U app -d appdb_identity_production -c "CREATE DATABASE appdb_${svc}_staging;" 2>/dev/null || true
done

echo "PostgreSQL ready with 16 logical databases."

# ============================================
# 6. NAMESPACES & SECRETS
# ============================================
echo ""
echo "=== Phase 6: K8s Namespaces & Secrets ==="

k3s kubectl create namespace production 2>/dev/null || true
k3s kubectl create namespace staging 2>/dev/null || true
k3s kubectl create namespace database 2>/dev/null || true

k3s kubectl create secret generic pawradise-secrets \
    --from-literal=JWT_SECRET="$JWT_SECRET" \
    --from-literal=ADMIN_TOKEN="$ADMIN_TOKEN" \
    --namespace production 2>/dev/null || true

k3s kubectl create secret generic pawradise-secrets \
    --from-literal=JWT_SECRET="$JWT_SECRET" \
    --from-literal=ADMIN_TOKEN="$ADMIN_TOKEN" \
    --namespace staging 2>/dev/null || true

k3s kubectl create secret generic postgres-secret \
    --from-literal=POSTGRES_PASSWORD="$DB_PASSWORD" \
    --namespace production 2>/dev/null || true

k3s kubectl create secret generic postgres-secret \
    --from-literal=POSTGRES_PASSWORD="$DB_PASSWORD" \
    --namespace staging 2>/dev/null || true

echo "JWT_SECRET=$JWT_SECRET"
echo "Namespaces and secrets created."

# ============================================
# 7. INGRESS-NGINX
# ============================================
echo ""
echo "=== Phase 7: Ingress-NGINX ==="

# Build from local source (avoids registry.k8s.io 403)
cd /tmp
if [ ! -d ingress-nginx ]; then
    curl -sL https://github.com/kubernetes/ingress-nginx/archive/refs/tags/controller-v1.11.2.tar.gz | tar xz
    mv ingress-nginx-controller-v1.11.2 ingress-nginx || true
fi

cd /tmp/ingress-nginx
if ! docker images | grep -q "pawradise-ingress-nginx.*v1.11.2"; then
    docker build -f Dockerfile.blue -t ${RED_IP}:30099/pawradise-ingress-nginx:v1.11.2 .
    docker push ${RED_IP}:30099/pawradise-ingress-nginx:v1.11.2
fi

# RBAC
k3s kubectl apply -f - << 'RBAC'
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ingress-nginx
rules:
- apiGroups: [""]
  resources: [configmaps, endpoints, nodes, pods, secrets, services, leases]
  verbs: [list, watch, get]
- apiGroups: [coordination.k8s.io]
  resources: [leases]
  verbs: [get, create, update, patch]
- apiGroups: [networking.k8s.io]
  resources: [ingresses, ingressclasses]
  verbs: [get, list, watch]
- apiGroups: [networking.k8s.io]
  resources: [ingresses/status]
  verbs: [update]
- apiGroups: [""]
  resources: [events]
  verbs: [create, patch]
- apiGroups: [discovery.k8s.io]
  resources: [endpointslices]
  verbs: [list, watch, get]
RBAC

k3s kubectl create namespace ingress-nginx 2>/dev/null || true
k3s kubectl create serviceaccount ingress-nginx -n ingress-nginx 2>/dev/null || true

k3s kubectl apply -f - << EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ingress-nginx
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ingress-nginx
subjects:
- kind: ServiceAccount
  name: ingress-nginx
  namespace: ingress-nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ingress-nginx-controller
  namespace: ingress-nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/component: controller
      app.kubernetes.io/instance: ingress-nginx
      app.kubernetes.io/name: ingress-nginx
  template:
    metadata:
      labels:
        app.kubernetes.io/component: controller
        app.kubernetes.io/instance: ingress-nginx
        app.kubernetes.io/name: ingress-nginx
    spec:
      serviceAccountName: ingress-nginx
      containers:
      - name: controller
        image: ${RED_IP}:30099/pawradise-ingress-nginx:v1.11.2
        args:
        - /nginx-ingress-controller
        - --election-id=ingress-nginx-leader
        - --controller-class=k8s.io/ingress-nginx
        - --ingress-class=nginx
        - --configmap=\$(POD_NAMESPACE)/ingress-nginx-controller
        - --watch-ingress-without-class=true
        ports:
        - containerPort: 80
          protocol: TCP
        - containerPort: 443
          protocol: TCP
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            cpu: 200m
            memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: ingress-nginx-controller
  namespace: ingress-nginx
spec:
  type: NodePort
  ports:
  - port: 80
    targetPort: 80
    nodePort: 30758
    protocol: TCP
    name: http
  - port: 443
    targetPort: 443
    nodePort: 30759
    protocol: TCP
    name: https
  selector:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
---
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: nginx
spec:
  controller: k8s.io/ingress-nginx
EOF

# Wait for ingress-nginx (skip webhook)
sleep 15
k3s kubectl wait --namespace ingress-nginx --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller --timeout=120s 2>/dev/null || true

# Delete webhook validation (avoids needing internet for admission images)
k3s kubectl delete validatingwebhookconfiguration ingress-nginx-admission 2>/dev/null || true
k3s kubectl delete job -n ingress-nginx ingress-nginx-admission-create ingress-nginx-admission-patch 2>/dev/null || true

echo "Ingress-nginx ready on NodePort 30758."

# ============================================
# 8. HOST NGINX
# ============================================
echo ""
echo "=== Phase 8: Host Nginx ==="

mkdir -p /var/www/production /var/www/staging /etc/nginx/ssl

cat > /var/www/production/index.html << 'HTML'
<!DOCTYPE html><html><head><title>Pawradise</title></head><body><h1>Welcome to Pawradise</h1></body></html>
HTML
cp /var/www/production/index.html /var/www/staging/index.html

cat > /etc/nginx/sites-available/pawradise << NGINX
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name _;

    proxy_set_header Host \$host;
    proxy_set_header X-Real-IP \$remote_addr;
    proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto \$scheme;
    proxy_set_header Authorization \$http_authorization;

    location / {
        proxy_pass http://127.0.0.1:30758;
    }
}
NGINX

ln -sf /etc/nginx/sites-available/pawradise /etc/nginx/sites-enabled/default
nginx -t && systemctl restart nginx

echo "Host nginx ready on port 80."

# ============================================
# SUMMARY
# ============================================
echo ""
echo "=========================================="
echo " RED Server Setup Complete!"
echo "=========================================="
echo ""
echo "IP: $RED_IP"
echo "JWT_SECRET: $JWT_SECRET"
echo "ADMIN_TOKEN: $ADMIN_TOKEN"
echo ""
echo "Components:"
echo "  - k3s"
echo "  - Docker + local registry (30099)"
echo "  - PostgreSQL (16 logical DBs)"
echo "  - Ingress-nginx (NodePort 30758)"
echo "  - Host nginx (port 80)"
echo "  - Security: UFW, fail2ban, SSH hardening"
echo ""
echo "Next steps:"
echo "  1. Build Go binaries on BLUE"
echo "  2. Build+push Docker images to registry"
echo "  3. Apply K8s manifests"
echo "  4. Run database migrations"
echo "  5. Run e2e tests"
echo ""
echo "Commands from BLUE:"
echo "  ./scripts/deploy-to-red.sh $RED_IP"
echo "=========================================="
