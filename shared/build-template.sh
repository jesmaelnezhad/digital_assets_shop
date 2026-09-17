#!/usr/bin/env bash
# Generic MFE build script — copy into each MFE directory and adjust MFE_NAME
set -euo pipefail

MFE_NAME="MFE_NAME_PLACEHOLDER"
DIST="dist"
SRC="src"

rm -rf "$DIST" && mkdir -p "$DIST/assets"

# 1. Copy this MFE's own pages
cp -r "$SRC"/* "$DIST/"

# 2. Shared chrome, theme, vendor, lib
cp ../../shared/chrome/*.html "$DIST/assets/" 2>/dev/null || true
cp ../../shared/chrome/app.js "$DIST/assets/"
cp ../../shared/theme/theme.css "$DIST/assets/"
cp ../../shared/vendor/alpine.min.js "$DIST/assets/"
cp ../../shared/lib/*.js "$DIST/assets/"

# 3. Env template
cp ../../shared/env.template.js "$DIST/env.template.js"

echo "Built $MFE_NAME -> $DIST"
