#!/bin/bash
set -e

# Creates a repository with corrupted .git/objects
# This simulates a repository where git operations will fail

echo "Creating corrupted fixture in $REPO_DIR"

git init "$REPO_DIR"
cd "$REPO_DIR"

git config user.email "test@twiggit.dev"
git config user.name "Test User"

echo "initial content" > README.md
git add README.md
git commit -m "Initial commit"

echo "more content" >> README.md
git add README.md
git commit -m "Second commit"

# Corrupt the objects directory by zeroing object files
# This creates a repo that appears valid but will fail on operations
OBJECTS_DIR=".git/objects"

# Corrupt the repository by making the objects directory unreadable
# This creates a repo that appears valid but will fail on operations
OBJECTS_DIR=".git/objects"

# Remove object files to create a corrupted state
# Git will fail when trying to read missing objects
find "$OBJECTS_DIR" -type f -name "[0-9a-f]*" -delete 2>/dev/null || true

# Also corrupt the HEAD ref to make it point to a non-existent object
echo "0000000000000000000000000000000000000000" > .git/refs/heads/main

echo "Corrupted repository created successfully"
