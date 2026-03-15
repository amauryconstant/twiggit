#!/bin/bash
set -e

# Creates a repository with a git submodule
# This tests handling of nested repositories

echo "Creating submodule fixture in $REPO_DIR"

# First create the submodule repo
SUBMODULE_DIR=$(mktemp -d)
git init "$SUBMODULE_DIR"
cd "$SUBMODULE_DIR"

git config user.email "test@twiggit.dev"
git config user.name "Test User"

echo "submodule content" > submodule.txt
git add submodule.txt
git commit -m "Initial submodule commit"

# Now create the main repo
git init "$REPO_DIR"
cd "$REPO_DIR"

git config user.email "test@twiggit.dev"
git config user.name "Test User"
git config protocol.file.allow always

echo "main repo content" > main.txt
git add main.txt
git commit -m "Initial main commit"

# Add the submodule (requires protocol.file.allow)
git -c protocol.file.allow=always submodule add "$SUBMODULE_DIR" lib/submodule
git commit -m "Add submodule"

# Clean up the temp submodule dir (it's now embedded in the main repo)
rm -rf "$SUBMODULE_DIR"

echo "Submodule repository created successfully"
