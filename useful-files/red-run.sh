#!/bin/bash
# Execute a command on RED via SSH. Survives short disconnects with keepalive.
# Usage: red-run <command> [args...]
ssh -o ServerAliveInterval=30 -o ServerAliveCountMax=3 root@194.5.206.106 "$@"
