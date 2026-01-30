#!/bin/bash
#MISE description="Calculate coverage excluding cmd/, test/, and main.go packages"

set -euo pipefail

if [ -n "${MISE_PROJECT_ROOT:-}" ]; then
  cd "$MISE_PROJECT_ROOT"
elif [ -n "${MISE_ORIGINAL_CWD:-}" ]; then
  cd "$MISE_ORIGINAL_CWD"
else
  # Fallback: navigate to project root from script location
  cd "$(dirname "$0")/../.."
fi

echo "🧪 Running unit tests with coverage..."
go test -short -tags=!integration -coverprofile=coverage_unit.out -covermode=atomic ./...

echo "🧪 Running integration tests with coverage..."
go test -tags=integration -coverprofile=coverage_integration.out -covermode=atomic ./test/integration/...

echo "📊 Merging coverage profiles..."
if [ -s coverage_integration.out ] && [ $(grep -c "^[^m]" coverage_integration.out) -gt 1 ]; then
  go run github.com/wadey/gocovmerge@latest coverage_unit.out coverage_integration.out > coverage.out
else
  echo "Integration coverage is empty, using unit coverage only"
  cp coverage_unit.out coverage.out
fi

echo "🔍 Calculating coverage for core packages only (excluding cmd/, test/, main.go)..."
# Extract coverage for internal packages only
go tool cover -func=coverage.out | grep "^github.com/amaury/twiggit/internal/" > coverage_filtered.txt

if [ ! -s coverage_filtered.txt ]; then
  echo "❌ No internal packages found in coverage report"
  exit 1
fi

# Calculate total coverage from filtered packages using awk
# We need to parse the actual statement counts from the coverage format
TOTAL_COVERAGE=$(awk '
BEGIN {
  total_statements = 0
  covered_statements = 0
}
{
  # Each line represents a function with its coverage percentage
  # Format: "github.com/amaury/twiggit/internal/package/file.go:line:	function_name	coverage%"
  coverage = $NF
  gsub(/%/, "", coverage)
  
  # Since we cannot get actual statement counts from -func output,
  # we use the coverage percentage as a weighted average
  # This is an approximation but better than counting functions
  if (coverage != "") {
    # Weight each function equally (this is an approximation)
    total_statements += 100  # Base weight
    covered_statements += coverage  # Actual coverage contribution
  }
}
END {
  if (total_statements > 0) {
    printf "%.1f", (covered_statements / total_statements) * 100
  } else {
    print "0"
  }
}
' coverage_filtered.txt)

echo "📊 Coverage Analysis:"
echo "   - Filtered coverage: ${TOTAL_COVERAGE}%"
echo "   - Excluded packages: cmd/, test/, main.go"

echo "📋 Coverage Summary:"
echo "   - Total coverage (filtered): ${TOTAL_COVERAGE}%"
echo "   - Required threshold: 80.0%"

# Use awk for floating point comparison
if awk "BEGIN {exit !($TOTAL_COVERAGE < 80.0)}"; then
  echo "❌ Coverage threshold not met: ${TOTAL_COVERAGE}% (minimum required: 80.0%)"
  exit 1
else
  echo "✅ Coverage threshold met: ${TOTAL_COVERAGE}% (minimum required: 80.0%)"
fi

# Clean up temporary file
rm -f coverage_filtered.txt