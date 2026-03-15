#!/usr/bin/env bash

#MISE description="Test GoReleaser configuration with a snapshot release"

set -e

echo "Running GoReleaser dry-run (snapshot)..."
echo ""

goreleaser release --snapshot --clean --skip=archive,sbom,before,homebrew

echo ""
echo "✅ Dry-run completed successfully"
