#!/bin/bash
# Quick Go build + test runner for pawradise-backend-dev workflow
# Use: bash /root/project/backend/check.sh
cd /root/project/backend
echo "=== Build ==="
go build ./... 2>&1 | head -5
echo "BUILD:$?"
echo ""
echo "=== Test suite ==="
go test ./handlers/ -count=1 -timeout 120s 2>&1 | tail -20
echo ""
echo "=== Coverage (if pass) ==="
go test ./handlers/ -count=1 -coverprofile=/tmp/cover.out -timeout 120s 2>&1 | tail -5
if [ -f /tmp/cover.out ]; then
    go tool cover -func=/tmp/cover.out 2>&1 | tail -3
fi