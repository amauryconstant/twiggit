#!/usr/bin/env bash

#MISE description="Update tool versions in mise configuration"

set -e

CONFIG_FILE=".mise/config.toml"

echo "Updating tool versions in ${CONFIG_FILE}..."
echo ""

get_latest_version() {
    local repo=$1

    local latest=$(curl -s "https://api.github.com/repos/${repo}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')

    if [ -z "$latest" ]; then
        echo "ERROR: Could not fetch latest version for ${repo}"
        exit 1
    fi

    echo "$latest"
}

update_tool() {
    local name=$1
    local repo=$2

    local current=$(grep "${name} = " "${CONFIG_FILE}" | sed "s/${name} = \"//" | sed 's/"//')
    local latest=$(get_latest_version "${repo}")

    echo "Updating ${name}: ${current} -> ${latest}"

    sed -i.tmp "s/${name} = \".*\"/${name} = \"${latest}\"/" "${CONFIG_FILE}"
    rm -f "${CONFIG_FILE}.tmp"
}

update_tool "golangci-lint" "golangci/golangci-lint"
update_tool "goreleaser" "goreleaser/goreleaser"

echo ""
echo "Updated ${CONFIG_FILE}"
echo ""
echo "Next steps:"
echo "1. Review changes: git diff ${CONFIG_FILE}"
echo "2. Install updated tools: mise install --yes"
echo "3. Commit and push"
