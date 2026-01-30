#!/usr/bin/env bash

#MISE description="Validate release prerequisites before tagging"

set -e

echo "Validating release prerequisites..."

if ! git rev-parse --git-dir > /dev/null 2>&1; then
  echo "❌ Not in a git repository"
  exit 1
fi

if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
  echo "❌ Working directory has uncommitted changes:"
  git status --short
  echo ""
  echo "Commit your changes first:"
  echo "  git add ."
  echo "  git commit -m 'your commit message'"
  exit 1
fi

CURRENT_BRANCH=$(git branch --show-current)

if [ "$CURRENT_BRANCH" != "main" ]; then
  echo "❌ Not on main branch (current: $CURRENT_BRANCH)"
  echo "Switch to main branch first:"
  echo "  git checkout main"
  exit 1
fi

if command -v goreleaser &> /dev/null; then
  if ! goreleaser check > /dev/null 2>&1; then
    echo "❌ GoReleaser configuration is invalid"
    echo "Run 'goreleaser check' to see errors"
    exit 1
  fi
fi

echo "✅ All checks passed"
echo "   - Working directory is clean"
echo "   - On main branch"
echo "   - GoReleaser configuration is valid"
