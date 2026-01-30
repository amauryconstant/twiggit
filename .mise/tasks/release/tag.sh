#!/usr/bin/env bash

#MISE description="Create and push git tag for release (patch|minor|major)"
#USAGE arg "<bump>" help="The type of bump to make"

set -e

BUMP_TYPE=$1

if [ -z "$BUMP_TYPE" ]; then
  echo "Usage: mise run release:tag <patch|minor|major>"
  exit 1
fi

if [[ ! "$BUMP_TYPE" =~ ^(patch|minor|major)$ ]]; then
  echo "Error: Invalid bump type '$BUMP_TYPE'. Use patch, minor, or major"
  exit 1
fi

if ! git rev-parse --git-dir > /dev/null 2>&1; then
  echo "Error: Not in a git repository"
  exit 1
fi

LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

if [ -z "$LATEST_TAG" ]; then
  echo "Error: No tags found. Create initial tag with: git tag v0.1.0"
  exit 1
fi

if [[ ! "$LATEST_TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: Invalid latest tag format: $LATEST_TAG (expected vX.Y.Z)"
  exit 1
fi

VERSION=${LATEST_TAG#v}

IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

case $BUMP_TYPE in
  patch)
    NEW_PATCH=$((PATCH + 1))
    NEW_VERSION="${MAJOR}.${MINOR}.${NEW_PATCH}"
    ;;
  minor)
    NEW_MINOR=$((MINOR + 1))
    NEW_VERSION="${MAJOR}.${NEW_MINOR}.0"
    ;;
  major)
    NEW_MAJOR=$((MAJOR + 1))
    NEW_VERSION="${NEW_MAJOR}.0.0"
    ;;
esac

NEW_TAG="v${NEW_VERSION}"

echo "Current tag: $LATEST_TAG"
echo "Bumping $BUMP_TYPE version"
echo "New tag: $NEW_TAG"
echo ""

if git tag -l "$NEW_TAG" | grep -q "$NEW_TAG"; then
  echo "Error: Tag $NEW_TAG already exists"
  exit 1
fi

git tag -a "$NEW_TAG" -m "Release $NEW_TAG"
git push origin "$NEW_TAG"

echo "✅ Tag $NEW_TAG created and pushed"
echo "📋 Monitor CI pipeline at $(git remote get-url origin | sed 's|\.git$||')/-/pipelines"
