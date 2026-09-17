# Separate Pod Deployment Pattern

As of 2026-09-09, Pawradise services run in separate pods (not grouped) on RED (130.185.123.156).

## Memory Budget
- RED is a 4GB VM (3921MB usable)
- Available after OS/k8s: ~2.4GB
- Per-pod: ~20-25MB (Go static binary on Alpine)
- 16 backend pods (8 services × 2 namespaces): ~320-400MB — fits
- ingress-nginx: ~64MB
- PostgreSQL: ~256MB

## Deployment Structure

Per-service Deployment + Service per namespace:
- `identity-service` (port 8081)
- `product-service` (port 8082)
- `commerce-service` (port 8083)
- `community-service` (port 8084)
- `review-service` (port 8085)
- `payment-service` (port 8086)
- `admin-service` (port 8087)
- `media-service` (port 8088)

Each has labels: `app: <service-name>`, `namespace: production|staging`.

## Pitfall: Docker Build Cache Serving Stale Images

Even with `--build-arg BUILD_TIME=$(date +%s)`, the Docker layer cache can serve a stale image if the binary in the build context hasn't changed its file content. This happened when the Go build cache served a stale binary — the Dockerfile's `COPY` layer saw the same file and reused the cached layer.

**Reliable fix:** Use `--no-cache` on the Docker build:
```bash
docker build --no-cache -t <registry>/<svc>:latest -f Dockerfile.<svc> /tmp 2>&1 | tail -3
```

**Also clear Go build cache before building:**
```bash
go clean -cache && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /tmp/<svc>-binary .
```

The `-a` flag forces rebuilding all packages, ignoring the build cache.

**Symptom:** Pod logs show old code behavior even after pushing a new image. The image hash in the registry doesn't change.

## Pitfall: Go Build Cache Serving Stale Binaries

**Symptom:** Pod logs show routes missing even though code registers them. Ingress returns 404 for all paths.

**Fix:** Force new image with:
- `--build-arg BUILD_TIME=$(date +%s)` in Dockerfile
- Or add a LABEL with a unique value

Verify: `k3s crictl images | grep <service>` on RED — the hash should change after push.

**Dockerfile pattern that works:**
```dockerfile
FROM alpine:3.19
ARG BUILD_TIME=unknown
RUN apk add --no-cache ca-certificates
LABEL build_time=${BUILD_TIME}
WORKDIR /
COPY service-binary /server
CMD ["/server"]
```

Build with: `docker build --build-arg BUILD_TIME=$(date +%s) -t tag -f Dockerfile /tmp`

**Pitfall:** Even with build-arg, if the binary file in the build context is the same, Docker may reuse the COPY layer. Always ensure the binary is freshly copied to the build context before building.

## Pitfall: kubectl cp + Process Restart Doesn't Work

Copying a new binary into a running pod and restarting with `pkill` fails because the old process holds the port.

**Correct approach:**
1. Push new image to registry
2. Delete the pod (deployment recreates it): `k3s kubectl delete pod -n <ns> <pod-name>`
3. Or update deployment image: `k3s kubectl set image deployment/<svc> -n <ns> <svc>=registry:tag:latest`

## Pitfall: Cross-namespace Env Vars

Services connect to their own logical database. The DB_NAME env var must match the namespace:
- Production: `appdb_<service>_production`
- Staging: `appdb_<service>_staging`

Use `sed 's/production/staging/g'` to generate staging manifests from production manifests.

## Pitfall: Duplicate Endpoints

When services share the same nodePort (30758 via ingress-nginx), stale endpoints from deleted pods can persist. Clean with:
```bash
k3s kubectl delete endpoints -n <namespace> --all
```
Kubernetes regenerates them automatically from current pods.
