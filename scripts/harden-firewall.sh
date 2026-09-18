#!/usr/bin/env bash
# UFW for BLUE (SSH only) or RED (SSH + 80/443; DB ports from BLUE only).
# Usage: scripts/harden-firewall.sh blue|red [--apply]
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/site-env.sh
. "$SCRIPT_DIR/lib/site-env.sh"

ROLE="${1:?usage: harden-firewall.sh blue|red [--apply]}"
APPLY="${2:-}"
case "$ROLE" in
  blue)
    PORT="${BLUE_SSH_PORT}"
    SSH=(ssh -p "${BLUE_SSH_PORT}" "${BLUE_SSH_USER}@${BLUE_HOST}")
    EXTRA=""
    ;;
  red)
    PORT="${RED_SSH_PORT}"
    SSH=(ssh -p "${RED_SSH_PORT}" "${RED_SSH_USER}@${RED_HOST}")
    EXTRA="ufw allow 80/tcp
ufw allow 443/tcp
ufw allow from ${BLUE_HOST} to any port ${DB_PORT}
ufw allow from ${BLUE_HOST} to any port ${MONGO_PORT}
ufw allow from 127.0.0.1 to any port ${DB_PORT}
ufw allow from 127.0.0.1 to any port ${MONGO_PORT}"
    ;;
  *) echo "role must be blue or red" >&2; exit 1 ;;
esac

remote() { "${SSH[@]}" "$@"; }

echo "Planned UFW on $ROLE: allow ${PORT}/tcp${EXTRA:+; plus role extras}."
if [[ "$APPLY" != "--apply" ]]; then
  echo "Re-run with --apply after a second SSH session still works."
  exit 0
fi

# shellcheck disable=SC2087
remote bash -s <<EOF
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y ufw
ufw default deny incoming
ufw default allow outgoing
ufw allow ${PORT}/tcp
${EXTRA}
ufw --force enable
ufw status verbose
EOF
