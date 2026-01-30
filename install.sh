#!/usr/bin/env bash

set -e

PROJECT="amoconst/twiggit"
DOWNLOAD_URL="https://gitlab.com/${PROJECT}/-/releases"

get_latest_version() {
    curl -s "${DOWNLOAD_URL}" | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*' | head -n 1
}

detect_os() {
    case "$(uname -s)" in
        Linux*)
            echo "linux"
            ;;
        Darwin*)
            echo "darwin"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            echo "windows"
            ;;
        *)
            echo "Unsupported OS: $(uname -s)"
            exit 1
            ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            echo "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
}

main() {
    OS=$(detect_os)
    ARCH=$(detect_arch)
    VERSION=${TWIGGIT_VERSION:-$(get_latest_version)}

    if [ -z "$VERSION" ]; then
        echo "ERROR: Could not detect latest version"
        exit 1
    fi

    echo "Installing twiggit ${VERSION} for ${OS}/${ARCH}..."

    if [ "$OS" = "windows" ]; then
        FILENAME="twiggit_${VERSION}_windows_${ARCH}.zip"
    else
        FILENAME="twiggit_${VERSION}_${OS}_${ARCH}.tar.gz"
    fi

    TMP_DIR=$(mktemp -d)
    cd "${TMP_DIR}"

    echo "Downloading ${FILENAME}..."
    curl -fsSL "${DOWNLOAD_URL}/downloads/${FILENAME}" -o "${FILENAME}"

    echo "Extracting..."
    if [ "$OS" = "windows" ]; then
        unzip -o "${FILENAME}"
    else
        tar -xzf "${FILENAME}"
    fi

    if [ "$OS" = "windows" ]; then
        BIN_DIR="${LOCALAPPDATA}\\Programs"
        mkdir -p "${BIN_DIR}"
        mv twiggit.exe "${BIN_DIR}\\"
        echo "Installation complete! twiggit.exe installed to ${BIN_DIR}"
        echo "Add ${BIN_DIR} to your PATH if not already present."
    else
        if [ -w /usr/local/bin ]; then
            BIN_DIR="/usr/local/bin"
        elif [ -w "$HOME/.local/bin" ]; then
            BIN_DIR="$HOME/.local/bin"
            mkdir -p "${BIN_DIR}"
        else
            echo "ERROR: Cannot write to /usr/local/bin or $HOME/.local/bin"
            echo "Please run with sudo or install manually"
            exit 1
        fi

        mv twiggit "${BIN_DIR}/"
        chmod +x "${BIN_DIR}/twiggit"
        echo "Installation complete! twiggit installed to ${BIN_DIR}"
    fi

    cd -
    rm -rf "${TMP_DIR}"

    echo ""
    echo "Verify installation:"
    if [ "$OS" = "windows" ]; then
        echo "  twiggit.exe version"
    else
        echo "  twiggit version"
    fi
}

main
