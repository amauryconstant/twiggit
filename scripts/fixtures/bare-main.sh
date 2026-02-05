#!/bin/bash

set -e

REPO_DIR=$1

if [ -z "$REPO_DIR" ]; then
	echo "Error: Repository directory argument required"
	exit 1
fi

echo "Creating bare-main fixture in $REPO_DIR"

# Initialize bare repo
git init --bare "$REPO_DIR"

# Create initial commit via separate clone
CLONE_DIR=$(mktemp -d)
git clone "$REPO_DIR" "$CLONE_DIR"
cd "$CLONE_DIR"

git config user.email "test@twiggit.dev"
git config user.name "Test User"

echo "initial content" > README.md
git add README.md
git commit -m "Initial commit"
git push origin main

cd -
rm -rf "$CLONE_DIR"

echo "Bare main repository created successfully"
