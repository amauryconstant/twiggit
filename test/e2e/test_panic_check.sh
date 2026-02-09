#!/bin/bash
cd "$(dirname "$0")"
echo "Testing with panic check..."
unset WORKTREES
mise run build:e2e >/dev/null 2>&1
timeout 5 ginkgo --tags=e2e -v --focus="delete command.*deletes worktree from project context" ./cmd/ 2>&1 | tail -50
