#!/bin/bash

set -e

REPO_DIR=$1

if [ -z "$REPO_DIR" ]; then
	echo "Error: Repository directory argument required"
	exit 1
fi

echo "Creating multi-branch fixture in $REPO_DIR"

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

# Create feature branches
git checkout -b feature-1
echo "feature 1 content" > feature1.txt
git add feature1.txt
git commit -m "Feature 1"

git checkout main
git checkout -b feature-2
echo "feature 2 content" > feature2.txt
git add feature2.txt
git commit -m "Feature 2"

git checkout main

echo "Multi-branch repository created with main, feature-1, and feature-2"
