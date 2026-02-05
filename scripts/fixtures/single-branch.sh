#!/bin/bash

set -e

REPO_DIR=$1

if [ -z "$REPO_DIR" ]; then
	echo "Error: Repository directory argument required"
	exit 1
fi

echo "Creating single-branch fixture in $REPO_DIR"

git init "$REPO_DIR"
cd "$REPO_DIR"

git config user.email "test@twiggit.dev"
git config user.name "Test User"

# Create main branch with 3 commits
for i in 1 2 3; do
	echo "commit $i" >> file.txt
	git add file.txt
	git commit -m "Commit $i"
done

echo "Single-branch repository created successfully with 3 commits"
