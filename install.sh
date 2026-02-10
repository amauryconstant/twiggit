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

detect_shell() {
    if [ -n "$ZSH_VERSION" ]; then
        echo "zsh"
    elif [ -n "$BASH_VERSION" ]; then
        echo "bash"
    elif [ -n "$FISH_VERSION" ]; then
        echo "fish"
    else
        case "$(basename "$SHELL")" in
            zsh) echo "zsh" ;;
            bash) echo "bash" ;;
            fish) echo "fish" ;;
            *) echo "" ;;
        esac
    fi
}

detect_config_file() {
    local shell="$1"
    local config_file=""

    case "$shell" in
        zsh)
            for file in "$HOME/.zshrc" "$HOME/.zprofile"; do
                if [ -r "$file" ]; then
                    config_file="$file"
                    break
                fi
            done
            if [ -z "$config_file" ]; then
                config_file="$HOME/.zshrc"
            fi
            ;;
        bash)
            for file in "$HOME/.bashrc" "$HOME/.bash_profile"; do
                if [ -r "$file" ]; then
                    config_file="$file"
                    break
                fi
            done
            if [ -z "$config_file" ]; then
                config_file="$HOME/.bashrc"
            fi
            ;;
        fish)
            config_file="$HOME/.config/fish/config.fish"
            ;;
        *)
            echo "Unsupported shell: $shell"
            return 1
            ;;
    esac

    echo "$config_file"
    return 0
}

install_completions() {
    local shell="$1"
    local completions_dir

    case "$shell" in
        zsh)
            if [ -w "$HOME/.local/share/zsh/site-functions" ]; then
                completions_dir="$HOME/.local/share/zsh/site-functions"
            elif [ -w "$HOME/.config/zsh/.zfunctions" ]; then
                completions_dir="$HOME/.config/zsh/.zfunctions"
            elif [ -w "$(dirname "${fpath[1]}")" ]; then
                completions_dir="$(dirname "${fpath[1]}")"
            else
                echo "⚠️  Could not find writable zsh completion directory"
                echo "   You can manually install with:"
                echo "   twiggit completion zsh > ~/.local/share/zsh/site-functions/_twiggit"
                return 1
            fi

            if "${BIN_DIR}/twiggit" completion zsh > "${completions_dir}/_twiggit"; then
                echo "✓ Zsh completions installed to ${completions_dir}/_twiggit"
                echo "  Restart your shell or run: autoload -Uz compinit && compinit"
                return 0
            else
                echo "✗ Failed to install zsh completions"
                return 1
            fi
            ;;
        bash)
            local bashrc="$HOME/.bashrc"

            if [ -w "$bashrc" ]; then
                if ! grep -q "twiggit completion bash" "$bashrc"; then
                    echo "" >> "$bashrc"
                    echo "# Twiggit completions" >> "$bashrc"
                    echo 'source <(twiggit completion bash)' >> "$bashrc"
                    echo "✓ Bash completions added to $bashrc"
                    echo "  Restart your shell or run: source $bashrc"
                    return 0
                else
                    echo "✓ Bash completions already configured in $bashrc"
                    return 0
                fi
            else
                echo "⚠️  Cannot write to $bashrc"
                echo "   Manually add this line: source <(twiggit completion bash)"
                return 1
            fi
            ;;
        fish)
            local fish_completions="$HOME/.config/fish/completions"
            if [ -w "$fish_completions" ] || mkdir -p "$fish_completions" 2>/dev/null; then
                if "${BIN_DIR}/twiggit" completion fish > "${fish_completions}/twiggit.fish"; then
                    echo "✓ Fish completions installed to ${fish_completions}/twiggit.fish"
                    return 0
                else
                    echo "✗ Failed to install fish completions"
                    return 1
                fi
            else
                echo "⚠️  Cannot write to fish completions directory"
                return 1
            fi
            ;;
        *)
            echo "⚠️  Unknown shell for completions"
            return 1
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
            echo ""
            echo "Possible solutions:"
            echo "  1. Run with sudo: sudo bash install.sh"
            echo "  2. Create directory manually: mkdir -p ~/.local/bin"
            echo "  3. Install to a different directory and add to PATH"
            exit 1
        fi

        mv twiggit "${BIN_DIR}/"
        chmod +x "${BIN_DIR}/twiggit"
        echo "Installation complete! twiggit installed to ${BIN_DIR}"
    fi

    cd -
    rm -rf "${TMP_DIR}"

    echo ""

    SHELL=$(detect_shell)
    if [ -n "$SHELL" ]; then
        echo "Detected shell: $SHELL"
        read -p "Install shell completions? [Y/n] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]] || [ -z "$REPLY" ]; then
            install_completions "$SHELL"
            echo ""
            echo "To install completions for other shells, run:"
            echo "  twiggit completion bash  # for bash"
            echo "  twiggit completion fish  # for fish"
        else
            echo "Skipped completions installation."
        fi
    else
        echo "Could not detect shell. Please specify your shell to install completions:"
        echo "  zsh) twiggit completion zsh > ~/.local/share/zsh/site-functions/_twiggit"
        echo "  bash) echo 'source <(twiggit completion bash)' >> ~/.bashrc"
        echo "  fish) twiggit completion fish > ~/.config/fish/completions/twiggit.fish"
    fi

    echo ""
    read -p "Install shell wrapper for directory navigation (twiggit cd)? [Y/n] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]] || [ -z "$REPLY" ]; then
        config_file=$(detect_config_file "${SHELL:-bash}")

        if [ -z "$config_file" ]; then
            echo "⚠️  Could not detect config file for shell: ${SHELL:-bash}"
            echo "  You can run manually: twiggit init <config-file>"
            return
        fi

        echo "Detected config file: $config_file"

        if grep -q "### BEGIN TWIGGIT WRAPPER" "$config_file" 2>/dev/null; then
            echo ""
            echo "⚠️  Wrapper already installed in $config_file"
            read -p "Reinstall with --force? [y/N] " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                echo "Reinstalling wrapper..."
                if "${BIN_DIR}/twiggit" init "$config_file" --force 2>/dev/null; then
                    echo "✓ Wrapper reinstalled successfully"
                else
                    echo "⚠️  Failed to reinstall shell wrapper"
                    echo "  You can run manually: twiggit init $config_file --force"
                fi
            else
                echo "Skipped reinstall."
            fi
        else
            echo ""
            read -p "Install wrapper to $config_file? [Y/n] " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]] || [ -z "$REPLY" ]; then
                if "${BIN_DIR}/twiggit" init "$config_file" 2>/dev/null; then
                    echo "✓ Wrapper installed successfully"
                else
                    echo "⚠️  Failed to install shell wrapper"
                    echo "  You can run manually: twiggit init $config_file"
                fi
            else
                echo "Skipped installation."
                echo "  You can run manually: twiggit init $config_file"
            fi
        fi
    else
        echo "Skipped shell wrapper installation."
        echo "  You can run manually: twiggit init <config-file>"
    fi

    case ":$PATH:" in
        *:${BIN_DIR}:*) ;;
        *)
            echo ""
            echo "⚠️  WARNING: ${BIN_DIR} may not be in your PATH"
            echo "   Add this line to your shell profile:"
            if [ "$OS" = "windows" ]; then
                echo "   setx PATH \"%PATH%;${BIN_DIR}\""
            else
                echo "   export PATH=\"${BIN_DIR}:\$PATH\""
            fi
            ;;
    esac

    echo ""
    echo "─────────────────────────────────────────────────────────"
    echo "  Installation Complete!"
    echo "─────────────────────────────────────────────────────────"
    echo ""
    echo "Quick start:"
    echo "  twiggit version"
    echo "  twiggit list"
    echo "  twiggit create my-feature"
    echo ""
    echo "For more information: https://gitlab.com/amoconst/twiggit"
    echo "─────────────────────────────────────────────────────────"
}

main
