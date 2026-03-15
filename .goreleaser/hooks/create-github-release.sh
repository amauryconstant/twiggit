#!/usr/bin/env bash

# Hook to create GitHub release for discoverability
# This runs after GitLab release succeeds

set -e

# Check if running in snapshot mode (skip GitHub release in snapshot)
if [ "${GORELEASER_CURRENT_TAG}" == "" ] || [[ "${GORELEASER_CURRENT_TAG}" == *"-next"* ]]; then
  echo "Snapshot mode detected, skipping GitHub release creation"
  exit 0
fi

# Check if GitHub token is available
if [ -z "${GITHUB_TOKEN}" ]; then
  echo "GITHUB_TOKEN not set, skipping GitHub release creation"
  exit 0
fi

echo "Creating GitHub release for discoverability..."

# Get release notes from git
TAG="${GORELEASER_CURRENT_TAG}"
NOTES=$(git log -1 --pretty=format:'%s' ${TAG})

# Create GitHub release
gh release create "${TAG}" \
  --repo amauryconstant/twiggit \
  --title "v${TAG}" \
  --notes "${NOTES}" \
  --notes "Download artifacts from GitLab: https://gitlab.com/amoconst/twiggit/-/releases/${TAG}" || echo "Failed to create GitHub release (may already exist)"

echo "GitHub release created or already exists"
