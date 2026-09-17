#!/bin/bash
# deploy-to-red.sh — Build, push, and deploy all 8 services to RED
# Usage: ./deploy-to-red.sh [RED_IP] [tag]
# Defaults: RED_IP=130.185.123.156, tag=latest

set -euo pipefail

RED_IP="${1:-130.185.123.156}"
TAG="${2:-$(date +%s)}"
LOCAL_REGISTRY="${RED_IP}:30099"

echo "=========================================="
echo " Deploy to RED: $RED_IP"
echo " Tag: $TAG"
echo "=========================================="

SERVICES="identity-service product-service commerce-service community-service review-service payment-service admin-service media-service"

# ============================================
# 1. Build Go binaries
# ============================================
echo ""
echo "=== Phase 1: Build Go binaries ==="

for svc in $SERVICES; do
  echo "Building $svc..."
  cd /root/project/services/$svc
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/$svc-binary . 2>&1 | tail -3
  if [ $? -ne 0 ]; then echo "BUILD FAILED: $svc"; exit 1; fi
done

echo "All binaries built."

# ============================================
# 2. Build Docker images
# ============================================
echo ""
echo "=== Phase 2: Build Docker images ==="

for svc in $SERVICES; do
  cat > /tmp/Dockerfile.$svc << DOCKERFILE
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY $svc-binary /server
EXPOSE 8080
CMD ["/server"]
DOCKERFILE

  docker build -t $LOCAL_REGISTRY/pawradise/$svc:$TAG -f /tmp/Dockerfile.$svc /tmp 2>&1 | tail -1
  docker tag $LOCAL_REGISTRY/pawradise/$svc:$TAG $LOCAL_REGISTRY/pawradise/$svc:latest 2>/dev/null || true
done

echo "All images built."

# ============================================
# 3. Push to registry
# ============================================
echo ""
echo "=== Phase 3: Push to registry ==="

for svc in $SERVICES; do
  docker push $LOCAL_REGISTRY/pawradise/$svc:$TAG 2>&1 | tail -1
  docker push $LOCAL_REGISTRY/pawradise/$svc:latest 2>&1 | tail -1
  echo "Pushed $svc:$TAG"
done

echo "All images pushed."

# ============================================
# 4. Deploy to RED
# ============================================
echo ""
echo "=== Phase 4: Deploy to RED ==="

ssh root@$RED_IP "
for ns in production staging; do
  for svc in $SERVICES; do
    k3s kubectl set image deployment/\$svc -n \$ns \$svc=$LOCAL_REGISTRY/pawradise/\$svc:$TAG 2>&1
  done
done
"

echo "Deployments updated."

# ============================================
# 5. Wait for rollout
# ============================================
echo ""
echo "=== Phase 5: Wait for rollout ==="

sleep 15
ssh root@$RED_IP "
echo '=== Production ==='
k3s kubectl get pods -n production 2>&1 | grep -v NAME
echo '=== Staging ==='
k3s kubectl get pods -n staging 2>&1 | grep -v NAME
"

# ============================================
# 6. Run migrations
# ============================================
echo ""
echo "=== Phase 6: Run database migrations ==="

for svc in $SERVICES; do
  dbname="appdb_\${svc%-service}_production"
  dbname_staging="appdb_\${svc%-service}_staging"
  for migration in /root/project/services/\$svc/migrations/*.sql; do
    if [ -f "\$migration" ]; then
      echo "Running \$migration -> \$dbname"
      ssh root@$RED_IP "docker exec -i postgres psql -U app -d \$dbname < \$migration" 2>&1 | tail -1
      ssh root@$RED_IP "docker exec -i postgres psql -U app -d \$dbname_staging < \$migration" 2>&1 | tail -1
    fi
  done
done

echo "All migrations applied."

# ============================================
# 7. Verify
# ============================================
echo ""
echo "=== Phase 7: Verify ==="

echo "Production health:"
curl -s http://$RED_IP/api/v1/health 2>&1 || echo "FAILED"
echo ""
echo "Staging health:"
curl -s http://$RED_IP/staging/api/v1/health 2>&1 || echo "FAILED"
echo ""

echo "=========================================="
echo " Deployment complete!"
echo " Tag: $TAG"
echo "=========================================="
