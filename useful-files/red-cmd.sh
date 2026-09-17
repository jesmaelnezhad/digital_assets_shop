#!/bin/bash
# On RED: run a command in a named tmux session
# Usage: red-cmd <session-name> <command>
tmux new-session -d -s "$1" bash 2>/dev/null
tmux send-keys -t "$1" "$2" Enter
