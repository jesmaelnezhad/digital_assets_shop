#!/usr/bin/env bash
# Build one frontend MFE image. Run from repo: shared/build-mfe.sh <mfe-name>
# IMAGE_TAG overrides the registry tag. PROJECT_ROOT defaults to repo root.
set -euo pipefail
MFE="${1:?mfe name}"
ROOT="${PROJECT_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"
DIR="$ROOT/frontend/$MFE"
cd "$DIR"
rm -rf dist
mkdir -p dist/assets
if [ -d src/assets ]; then cp -r src/assets/. dist/assets/ 2>/dev/null || true; fi
find src -maxdepth 1 -type f \( -name '*.html' -o -name '*.xml' -o -name '*.txt' \) -exec cp {} dist/ \;
cp "$ROOT/shared/theme/theme.css" dist/assets/theme.css
cp "$ROOT/shared/lib/api.js" dist/assets/api.js
cp "$ROOT/shared/lib/ui.js" dist/assets/ui.js
cp "$ROOT/shared/lib/events.js" dist/assets/events.js
cp "$ROOT/shared/chrome/chrome.js" dist/assets/chrome.js
cp "$ROOT/shared/chrome/app.js" dist/assets/app.js 2>/dev/null || true
if [ -f "$ROOT/frontend/shop-mfe/src/assets/favicon.svg" ]; then
  cp "$ROOT/frontend/shop-mfe/src/assets/favicon.svg" dist/assets/favicon.svg
fi
MEDIA="$ROOT/shared/media/catalog"
if [ ! -d "$MEDIA" ] || [ -z "$(ls -A "$MEDIA" 2>/dev/null)" ]; then
  MEDIA="$ROOT/frontend/prototype/shared/media"
fi
if [ -d "$MEDIA" ]; then
  mkdir -p dist/assets/catalog
  cp "$MEDIA"/*.jpg dist/assets/catalog/ 2>/dev/null || true
fi
if [ -f "$ROOT/shared/vendor/alpine.min.js" ]; then
  cp "$ROOT/shared/vendor/alpine.min.js" dist/assets/alpine.min.js
fi
cp "$ROOT/shared/env.template.js" dist/env.template.js
cp "$ROOT/shared/docker-entrypoint-env.sh" dist/docker-entrypoint-env.sh
if [ -f src/nginx-mfe.conf ]; then
  cp src/nginx-mfe.conf dist/nginx-mfe.conf
elif [ -f nginx-mfe.conf ]; then
  cp nginx-mfe.conf dist/nginx-mfe.conf
else
  cp "$ROOT/shared/nginx-mfe.conf" dist/nginx-mfe.conf
fi

case "$MFE" in
  shop-mfe)
    cp dist/index.html dist/home.html 2>/dev/null || true
    ;;
  product-mfe)
    mkdir -p dist/product
    cp dist/product.html dist/product/product.html 2>/dev/null || true
    cp dist/product.html dist/product/index.html 2>/dev/null || true
    cp dist/index.html dist/product/index.html 2>/dev/null || true
    cp dist/bundle.html dist/product/bundle.html 2>/dev/null || true
    cp dist/bundles.html dist/product/bundles.html 2>/dev/null || true
    cp dist/request.html dist/product/request.html 2>/dev/null || true
    ;;
  community-mfe)
    mkdir -p dist/community dist/post dist/profile
    cp dist/index.html dist/community.html 2>/dev/null || true
    cp dist/community.html dist/community/index.html 2>/dev/null || true
    cp dist/post.html dist/post/index.html 2>/dev/null || true
    cp dist/profile.html dist/profile/index.html 2>/dev/null || true
    ;;
  account-mfe)
    mkdir -p dist/account
    cp dist/account.html dist/index.html 2>/dev/null || true
    cp dist/account.html dist/account/index.html 2>/dev/null || true
    cp dist/order.html dist/account/order.html 2>/dev/null || true
    ;;
  checkout-mfe)
    mkdir -p dist/checkout
    cp dist/index.html dist/checkout.html 2>/dev/null || true
    cp dist/checkout.html dist/checkout/checkout.html 2>/dev/null || true
    cp dist/checkout.html dist/checkout/index.html 2>/dev/null || true
    ;;
  admin-app)
    mkdir -p dist/admin
    cp dist/index.html dist/admin.html 2>/dev/null || true
    cp dist/index.html dist/admin/index.html 2>/dev/null || true
    cp dist/index.html dist/admin/admin.html 2>/dev/null || true
    ;;
  auth-mfe)
    cp dist/login.html dist/index.html 2>/dev/null || true
    mkdir -p dist/auth
    cp dist/login.html dist/auth/login.html 2>/dev/null || true
    cp dist/register.html dist/auth/register.html 2>/dev/null || true
    ;;
esac

TAG="${IMAGE_TAG:-}"
if [[ -z "$TAG" ]]; then
  echo "Set IMAGE_TAG (example: IMAGE_TAG=\$(IMAGE ${MFE}):tag bash shared/build-mfe.sh ${MFE})" >&2
  exit 1
fi
docker build -t "$TAG" .
echo "Built $TAG"
