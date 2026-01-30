#!/usr/bin/env bash

#MISE description="Check for available tool updates"

set -e

CONFIG_FILE=".mise/config.toml"

echo "Checking for tool updates..."
echo ""

get_latest_version() {
    local repo=$1
    local current=$2

    local latest=$(curl -s "https://api.github.com/repos/${repo}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')

    if [ -z "$latest" ]; then
        echo "ERROR: Could not fetch latest version for ${repo}"
        exit 1
    fi

    echo "$latest"
}

check_tool() {
    local name=$1
    local repo=$2

    local current=$(grep "${name} = " "${CONFIG_FILE}" | sed "s/${name} = \"//" | sed 's/"//')
    echo -n "${name}: "
    if [ "$current" = "latest" ]; then
        echo "current: latest (always latest)"
        return
    fi

    local latest=$(get_latest_version "${repo}" "${current}")

    if [ "$current" = "$latest" ]; then
        echo "current: ${current} (up to date)"
    else
        echo "current: ${current} -> latest: ${latest} (update available)"
    fi
}

check_tool "golangci-lint" "golangci/golangci-lint"
check_tool "goreleaser" "goreleaser/goreleaser"

echo ""
echo "Run 'mise run tools:update' to update versions in ${CONFIG_FILE}"
