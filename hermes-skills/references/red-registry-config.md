# Registry Configuration on RED

When setting up the local registry on RED, the k3s registries.yaml configuration can interfere with pulling images from public registries. This documents the correct configuration and common issues.

## The Problem

After configuring `/etc/rancher/k3s/registries.yaml` to add the local registry mirror, pulls from `docker.io` and `registry.k8s.io` may start failing with errors like:
- `pull access denied, repository does not exist or may require 'docker login'`
- `403 Forbidden` from registry.k8s.io
- `401 Unauthorized` from ghcr.io

## Correct Configuration

`/etc/rancher/k3s/registries.yaml`:

```yaml
mirrors:
  docker.io:
    endpoint:
      - https://registry-1.docker.io
  194.5.206.106:30099:
    endpoint:
      - http://194.5.206.106:30099
configs:
  194.5.206.106:30099:
    tls:
      insecure_skip_verify: true
```

**Key points:**
- The `docker.io` mirror entry is REQUIRED — without it, k3s may not be able to pull from Docker Hub
- The `endpoint` for docker.io must be `https://registry-1.docker.io` (not `https://docker.io`)
- The local registry entry must have `tls.insecure_skip_verify: true` for HTTP

## Common Mistakes

1. **Missing docker.io mirror** — if you only add the local registry, docker.io pulls break
2. **Wrong endpoint URL** — `https://docker.io` doesn't work, must be `https://registry-1.docker.io`
3. **Forgetting to restart k3s** — changes require `systemctl restart k3s`
4. **Adding registry.k8s.io mirror** — adding a mirror for registry.k8s.io that doesn't work can make things worse

## Verification

After editing registries.yaml and restarting k3s:

```bash
# From RED:
k3s ctr images pull docker.io/library/nginx:1.25-alpine  # Should work
k3s ctr images pull docker.io/nginx/nginx-ingress:3.4.3    # Should work

# From BLUE:
docker pull docker.io/library/nginx:1.25-alpine           # Should work
docker pull docker.io/nginx/nginx-ingress:3.4.3           # Should work
```

## Registry Status on RED

The local registry is at `194.5.206.106:30099` (NodePort 30099).

Check what's in it:
```bash
curl -s http://194.5.206.106:30099/v2/_catalog | python3 -m json.tool
```

Push from BLUE:
```bash
docker tag myimage 194.5.206.106:30099/myimage
docker push 194.5.206.106:30099/myimage
```

## Ingress-Nginx Controller Image

The standard `registry.k8s.io/ingress-nginx/controller:v1.11.2` is available as a tar file at `/root/ingress-nginx-controller.tar` on BLUE.

To use it:

```bash
# On BLUE
docker load -i /root/ingress-nginx-controller.tar
docker tag registry.k8s.io/ingress-nginx/controller:v1.11.2 194.5.206.106:30099/pawradise-ingress-nginx:v1.11.2
docker push 194.5.206.106:30099/pawradise-ingress-nginx:v1.11.2
```

Then deploy using the standard manifests from `https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.11.2/deploy/static/provider/baremetal/deploy.yaml`, changing only the image reference.

**Note:** `registry.k8s.io` is blocked (403) from this server, so direct pulls will not work. The tar file is the only way to get the standard controller image.
