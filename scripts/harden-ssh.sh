#!/usr/bin/env bash
# Write sshd drop-in. Does not reload until --apply, and never if authorized_keys is empty.
# Usage: scripts/harden-ssh.sh blue|red [--apply]
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/site-env.sh
. "$SCRIPT_DIR/lib/site-env.sh"

ROLE="${1:?usage: harden-ssh.sh blue|red [--apply]}"
APPLY="${2:-}"
case "$ROLE" in
  blue) PORT="${BLUE_SSH_PORT}"; SSH=(ssh -p "${BLUE_SSH_PORT}" "${BLUE_SSH_USER}@${BLUE_HOST}") ;;
  red)  PORT="${RED_SSH_PORT}";  SSH=(ssh -p "${RED_SSH_PORT}" "${RED_SSH_USER}@${RED_HOST}") ;;
  *) echo "role must be blue or red" >&2; exit 1 ;;
esac

remote() { "${SSH[@]}" "$@"; }

keys="$(remote "cat ~/.ssh/authorized_keys 2>/dev/null | grep -c '^ssh-' || true")"
if [[ "${keys:-0}" -lt 1 ]]; then
  echo "no SSH public keys in authorized_keys on $ROLE — refusing to disable passwords" >&2
  exit 1
fi

# shellcheck disable=SC2087
remote bash -s <<EOF
set -euo pipefail
install -d -m 0755 /etc/ssh/sshd_config.d
cat > /etc/ssh/sshd_config.d/store4bots.conf <<CONF
Port ${PORT}
PasswordAuthentication no
KbdInteractiveAuthentication no
PermitRootLogin prohibit-password
PubkeyAuthentication yes
CONF
sshd -t
EOF

echo "wrote /etc/ssh/sshd_config.d/store4bots.conf on $ROLE (Port ${PORT}, key-only)."
echo "Keep this session open. Open a NEW ssh -p ${PORT} session. Only then re-run with --apply."

if [[ "$APPLY" == "--apply" ]]; then
  remote "systemctl reload ssh || systemctl reload sshd || true"
  echo "sshd reloaded on $ROLE"
fi
