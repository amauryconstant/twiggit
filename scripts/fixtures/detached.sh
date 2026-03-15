#!/bin/bash
set -e

# Creates a repository in detached HEAD state
# This tests handling of non-branch refs

echo "Creating detached HEAD fixture in $REPO_DIR"

git init "$REPO_DIR"
cd "$REPO_DIR"

git config user.email "test@twiggit.dev"
git config user.name "Test User"

# Create main branch with commits
echo "commit 1" > file.txt
git add file.txt
git commit -m "Commit 1"

echo "commit 2" >> file.txt
git add file.txt
git commit -m "Commit 2"

# Store the commit hash for detaching
COMMIT_HASH=$(git rev-parse HEAD)

echo "commit 3" >> file.txt
git add file.txt
git commit -m "Commit 3"

# Checkout to detached HEAD state (at commit 2)
git checkout "$COMMIT_HASH"

echo "Detached HEAD repository created successfully"
