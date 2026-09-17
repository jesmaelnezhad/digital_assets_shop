#!/bin/bash
# Check RED health: tmux sessions, k3s nodes, pods
echo "=== RED tmux sessions ==="
ssh -o ServerAliveInterval=30 -o ServerAliveCountMax=3 root@194.5.206.106 "tmux list-sessions 2>/dev/null || echo 'no sessions'"
echo ""
echo "=== RED k3s nodes ==="
ssh -o ServerAliveInterval=30 -o ServerAliveCountMax=3 root@194.5.206.106 "k3s kubectl get nodes"
echo ""
echo "=== RED pods ==="
ssh -o ServerAliveInterval=30 -o ServerAliveCountMax=3 root@194.5.206.106 "k3s kubectl get pods -A"
