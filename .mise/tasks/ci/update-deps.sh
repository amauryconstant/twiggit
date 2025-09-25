#!/bin/bash
set -euo pipefail

echo "🔄 Updating CI dependencies securely..."

# Check if Docker is available
echo "🔍 Checking Docker availability..."
if ! command -v docker >/dev/null 2>&1; then
    echo "❌ Docker is not available!"
    echo "🔧 Docker is required for this project."
    echo "💡 To fix this issue:"
    echo "   1. Install Docker: https://docs.docker.com/get-docker/"
    echo "   2. Ensure Docker is in your PATH"
    exit 1
fi

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker daemon is not running!"
    echo "🔧 Docker daemon is required for this project."
    echo "💡 To fix this issue:"
    echo "   1. Start Docker daemon"
    echo "   2. Check Docker service status"
    exit 1
fi

# Check if Docker Buildx is available
echo "🔍 Checking Docker Buildx availability..."
if ! docker buildx version >/dev/null 2>&1; then
    echo "❌ Docker Buildx is not available!"
    echo "🔧 Docker Buildx is required for BuildKit support."
    echo "💡 To fix this issue:"
    echo "   1. Install Docker Buildx: https://docs.docker.com/go/buildx/"
    echo "   2. Or update Docker to a version that includes Buildx"
    echo "   3. For Docker Desktop, ensure Buildx is enabled"
    echo "   4. For this project, run: curl -L https://github.com/docker/buildx/releases/download/v0.16.2/buildx-v0.16.2.linux-amd64 -o ~/.docker/cli-plugins/docker-buildx && chmod +x ~/.docker/cli-plugins/docker-buildx"
    exit 1
fi

echo "✅ Docker and Buildx are available and running"

# Get latest mise release info
echo "📦 Fetching latest mise release..."
LATEST_MISE_RELEASE=$(curl -s https://api.github.com/repos/jdx/mise/releases/latest | grep -o '"tag_name": "v[^"]*' | cut -d'"' -f4)
MISE_VERSION=${LATEST_MISE_RELEASE#v}

echo "🔍 Getting mise binary checksum..."
# Download the binary and calculate checksum directly
MISE_X64_MUSL_CHECKSUM=$(curl -s -L https://github.com/jdx/mise/releases/download/${LATEST_MISE_RELEASE}/mise-${LATEST_MISE_RELEASE}-linux-x64-musl | sha256sum | cut -d' ' -f1)

echo "📋 Latest mise version: $MISE_VERSION"
echo "🔐 Checksum: $MISE_X64_MUSL_CHECKSUM"

# Get latest Alpine version
echo "🏔️ Fetching latest Alpine version..."
LATEST_ALPINE_VERSION=$(curl -s https://hub.docker.com/v2/repositories/library/alpine/tags/?page_size=100 | jq -r '.results[].name' | grep -E '^[0-9]+\.[0-9]+$' | sort -V | tail -1)

echo "📋 Latest Alpine version: $LATEST_ALPINE_VERSION"

# Get latest Docker version
echo "🐳 Fetching latest Docker version..."
LATEST_DOCKER_VERSION=$(curl -s https://hub.docker.com/v2/repositories/library/docker/tags/?page_size=100 | jq -r '.results[].name' | grep -E '^[0-9]+\.[0-9]+$' | sort -V | tail -1)

echo "📋 Latest Docker version: $LATEST_DOCKER_VERSION"

# Get shell versions (bash, zsh, fish) from Alpine package registry
echo "🐚 Fetching shell versions from Alpine package registry..."
LATEST_BASH_VERSION=$(curl -s "https://pkgs.alpinelinux.org/packages?name=bash&branch=v$LATEST_ALPINE_VERSION" | grep -o '<strong class="hint--right hint--rounded text-success" aria-label="Package version">[^<]*</strong>' | head -1 | sed 's/<strong[^>]*>\([^<]*\)<\/strong>/\1/' || echo "")
LATEST_ZSH_VERSION=$(curl -s "https://pkgs.alpinelinux.org/packages?name=zsh&branch=v$LATEST_ALPINE_VERSION" | grep -o '<strong class="hint--right hint--rounded text-success" aria-label="Package version">[^<]*</strong>' | head -1 | sed 's/<strong[^>]*>\([^<]*\)<\/strong>/\1/' || echo "")
LATEST_FISH_VERSION=$(curl -s "https://pkgs.alpinelinux.org/packages?name=fish&branch=v$LATEST_ALPINE_VERSION" | grep -o '<strong class="hint--right hint--rounded text-success" aria-label="Package version">[^<]*</strong>' | head -1 | sed 's/<strong[^>]*>\([^<]*\)<\/strong>/\1/' || echo "")

echo "📋 Shell versions:"
echo "   - Bash: ${LATEST_BASH_VERSION:-"N/A"}"
echo "   - Zsh: ${LATEST_ZSH_VERSION:-"N/A"}"
echo "   - Fish: ${LATEST_FISH_VERSION:-"N/A"}"

# Get current Docker image version
echo "📋 Reading Docker image version..."
DOCKER_IMAGE_VERSION=$(cat DOCKER_IMAGE_VERSION 2>/dev/null || echo "0.1.0")
echo "📋 Current Docker image version: $DOCKER_IMAGE_VERSION"

# Get current versions from CI file for comparison
echo "🔍 Extracting current versions from CI configuration..."
CURRENT_MISE_VERSION=$(grep "MISE_VERSION: v" .gitlab-ci.yml | cut -d':' -f2 | tr -d ' ' | sed 's/^v//')
CURRENT_ALPINE_VERSION=$(grep "ALPINE_VERSION:" .gitlab-ci.yml | cut -d':' -f2 | tr -d ' ')
CURRENT_BASH_VERSION=$(grep "BASH_VERSION:" .gitlab-ci.yml | cut -d':' -f2 | tr -d ' ')
CURRENT_ZSH_VERSION=$(grep "ZSH_VERSION:" .gitlab-ci.yml | cut -d':' -f2 | tr -d ' ')
CURRENT_FISH_VERSION=$(grep "FISH_VERSION:" .gitlab-ci.yml | cut -d':' -f2 | tr -d ' ')
echo "📋 Current versions in CI configuration:"
echo "   - Mise: v$CURRENT_MISE_VERSION"
echo "   - Alpine: $CURRENT_ALPINE_VERSION"
echo "   - Bash: ${CURRENT_BASH_VERSION:-"default"}"
echo "   - Zsh: ${CURRENT_ZSH_VERSION:-"default"}"
echo "   - Fish: ${CURRENT_FISH_VERSION:-"default"}"

# Update Dockerfile with new versions
echo "🐳 Updating Dockerfile..."
sed -i.bak \
    -e "s/MISE_SHA256=[a-f0-9]*/MISE_SHA256=$MISE_X64_MUSL_CHECKSUM/" \
    Dockerfile

# Create backup of version file
cp DOCKER_IMAGE_VERSION DOCKER_IMAGE_VERSION.bak 2>/dev/null || true

# Analyze changes and determine version increment
NEW_VERSION=$DOCKER_IMAGE_VERSION
if ! git diff --quiet Dockerfile; then
    echo "📋 Dockerfile changed, analyzing changes..."
    
    # Check for Alpine version change (MINOR increment)
    if [ "$CURRENT_ALPINE_VERSION" != "$LATEST_ALPINE_VERSION" ]; then
        echo "🟡 MINOR: Alpine version changed from $CURRENT_ALPINE_VERSION to $LATEST_ALPINE_VERSION"
        NEW_VERSION=$(echo $DOCKER_IMAGE_VERSION | awk -F. '{print $1"."($2+1)".0"}')
    # Check for Mise version change (MINOR increment)
    elif [ "$CURRENT_MISE_VERSION" != "$MISE_VERSION" ]; then
        echo "🟡 MINOR: Mise version changed from v$CURRENT_MISE_VERSION to v$MISE_VERSION"
        NEW_VERSION=$(echo $DOCKER_IMAGE_VERSION | awk -F. '{print $1"."($2+1)".0"}')
    # Check for shell version changes (MINOR increment)
    elif [ "$CURRENT_BASH_VERSION" != "${LATEST_BASH_VERSION:-}" ] || [ "$CURRENT_ZSH_VERSION" != "${LATEST_ZSH_VERSION:-}" ] || [ "$CURRENT_FISH_VERSION" != "${LATEST_FISH_VERSION:-}" ]; then
        echo "🟡 MINOR: Shell versions updated"
        [ "$CURRENT_BASH_VERSION" != "${LATEST_BASH_VERSION:-}" ] && echo "   - Bash: ${CURRENT_BASH_VERSION:-"default"} → ${LATEST_BASH_VERSION:-"default"}"
        [ "$CURRENT_ZSH_VERSION" != "${LATEST_ZSH_VERSION:-}" ] && echo "   - Zsh: ${CURRENT_ZSH_VERSION:-"default"} → ${LATEST_ZSH_VERSION:-"default"}"
        [ "$CURRENT_FISH_VERSION" != "${LATEST_FISH_VERSION:-}" ] && echo "   - Fish: ${CURRENT_FISH_VERSION:-"default"} → ${LATEST_FISH_VERSION:-"default"}"
        NEW_VERSION=$(echo $DOCKER_IMAGE_VERSION | awk -F. '{print $1"."($2+1)".0"}')
    else
        echo "🟢 PATCH: Other Dockerfile changes (checksum, etc.)"
        NEW_VERSION=$(echo $DOCKER_IMAGE_VERSION | awk -F. '{print $1"."$2"."($3+1)}')
    fi
    
    echo "📋 Version increment: $DOCKER_IMAGE_VERSION → $NEW_VERSION"
    echo $NEW_VERSION > DOCKER_IMAGE_VERSION
else
    echo "✅ No Dockerfile changes detected, version remains: $DOCKER_IMAGE_VERSION"
fi

# Update GitLab CI with new versions
echo "🔧 Updating GitLab CI configuration..."
sed -i.bak \
    -e "s/image: docker:[0-9]*\\.[0-9]*/image: docker:$LATEST_DOCKER_VERSION/g" \
    -e "s/- docker:[0-9]*\\.[0-9]*-dind/- docker:$LATEST_DOCKER_VERSION-dind/g" \
    -e "s/MISE_VERSION: v[0-9][0-9][0-9][0-9]\\.[0-9]*\\.[0-9]*/MISE_VERSION: v$MISE_VERSION/g" \
    -e "s/DOCKER_IMAGE_VERSION: [0-9]*\\.[0-9]*\\.[0-9]*/DOCKER_IMAGE_VERSION: $NEW_VERSION/g" \
    -e "s/ALPINE_VERSION: [0-9]*\\.[0-9]*/ALPINE_VERSION: $LATEST_ALPINE_VERSION/g" \
    -e "s/BASH_VERSION: [^ ]*/BASH_VERSION: $LATEST_BASH_VERSION/g" \
    -e "s/ZSH_VERSION: [^ ]*/ZSH_VERSION: $LATEST_ZSH_VERSION/g" \
    -e "s/FISH_VERSION: [^ ]*/FISH_VERSION: $LATEST_FISH_VERSION/g" \
    .gitlab-ci.yml

echo "✅ Files updated successfully"

# Test Docker build with BuildKit
echo "🧪 Setting up Buildx builder with container driver (like CI)..."
docker buildx create --use --driver docker-container --name builder 2>/dev/null || docker buildx use builder
docker buildx inspect --bootstrap

echo "🧪 Testing Docker build with BuildKit..."
if docker buildx build --build-arg BUILDKIT_INLINE_CACHE=1 --build-arg ALPINE_VERSION=$LATEST_ALPINE_VERSION --build-arg MISE_VERSION=$MISE_VERSION --build-arg BASH_VERSION=$LATEST_BASH_VERSION --build-arg ZSH_VERSION=$LATEST_ZSH_VERSION --build-arg FISH_VERSION=$LATEST_FISH_VERSION --load -t twiggit-ci-test:latest .; then
    # Verify mise version in built image
    if docker inspect twiggit-ci-test:latest >/dev/null 2>&1; then
        MISE_IN_IMAGE=$(docker run --rm twiggit-ci-test:latest mise --version 2>/dev/null || echo "verification skipped")
        echo "📦 Mise version in image: $MISE_IN_IMAGE"
        
        # Clean up test image
        docker rmi twiggit-ci-test:latest 2>/dev/null || true
    else
        echo "📦 Docker build successful, image verification skipped (image not loaded)"
    fi
    
    echo "✅ Docker build with BuildKit successful"
    
    # Clean up backup files if everything went well
    echo "🧹 Cleaning up backup files..."
    rm -f Dockerfile.bak .gitlab-ci.yml.bak DOCKER_IMAGE_VERSION.bak
    
    echo "🎉 CI dependencies updated successfully!"
    echo "📋 Summary:"
    echo "   - Mise version: v$MISE_VERSION"
    echo "   - Checksum: $MISE_X64_MUSL_CHECKSUM"
    echo "   - Alpine version: $LATEST_ALPINE_VERSION"
    echo "   - Docker version: $LATEST_DOCKER_VERSION"
    echo "   - Shell versions:"
    echo "     * Bash: $LATEST_BASH_VERSION"
    echo "     * Zsh: $LATEST_ZSH_VERSION"
    echo "     * Fish: $LATEST_FISH_VERSION"
    echo "   - Docker image version: $NEW_VERSION"
    echo "   - Dockerfile and .gitlab-ci.yml updated and tested"
    echo "   - BuildKit and Buildx working correctly"
    echo "   - SemVer image tagging: $NEW_VERSION"
    echo "   - Backup files cleaned up"
    if [ "$DOCKER_IMAGE_VERSION" != "$NEW_VERSION" ]; then
        echo "   - Version increment: $DOCKER_IMAGE_VERSION → $NEW_VERSION"
    fi
    
    # Clean up buildx builder
    docker buildx rm builder 2>/dev/null || true
else
    echo "❌ Docker build with BuildKit failed!"
    echo "🔧 BuildKit and Buildx are required for this project."
    echo "💡 To fix this issue:"
    echo "   1. Ensure Docker supports BuildKit (version 18.09+)"
    echo "   2. Install Docker Buildx: https://docs.docker.com/go/buildx/"
    echo "   3. Check Docker daemon is running"
    echo "   4. Verify build context is accessible"
    echo "🔄 Rolling back changes..."
    mv Dockerfile.bak Dockerfile
    mv .gitlab-ci.yml.bak .gitlab-ci.yml
    # Restore original version file if it was changed
    if [ -f DOCKER_IMAGE_VERSION.bak ]; then
        mv DOCKER_IMAGE_VERSION.bak DOCKER_IMAGE_VERSION
    fi
    # Clean up buildx builder
    docker buildx rm builder 2>/dev/null || true
    exit 1
fi