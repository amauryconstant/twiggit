#!/bin/bash

set -e

FIXTURES_DIR="test/e2e/fixtures/repos"
SCRIPTS_DIR="scripts/fixtures"

mkdir -p "$FIXTURES_DIR"

generate_fixture() {
	local name=$1
	local setup_script="$SCRIPTS_DIR/$2"

	echo "Generating $name fixture..."
	
	if [ ! -f "$setup_script" ]; then
		echo "Error: Setup script $setup_script not found"
		exit 1
	fi

	tmpdir=$(mktemp -d)
	
	bash "$setup_script" "$tmpdir"
	
	tar -czf "$FIXTURES_DIR/$name.tar.gz" -C "$tmpdir" .
	
	rm -rf "$tmpdir"
	echo "✓ Created $name.tar.gz"
}

generate_fixture "bare-main" "bare-main.sh"
generate_fixture "single-branch" "single-branch.sh"
generate_fixture "multi-branch" "multi-branch.sh"

echo ""
echo "All fixtures generated successfully!"
