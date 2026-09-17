#!/usr/bin/env bash
# Shared build helper — injects staging-links.js into all HTML files
# Usage: inject_staging_links <dist_dir>
inject_staging_links() {
  local dist_dir="$1"
  for html in "$dist_dir"/*.html; do
    [ -f "$html" ] || continue
    # Add staging-links.js before </body> if not already present
    if ! grep -q "staging-links.js" "$html"; then
      sed -i 's|</body>|<script src="/assets/staging-links.js"></script>\n</body>|' "$html"
    fi
  done
}
