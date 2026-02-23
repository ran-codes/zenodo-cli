#!/bin/bash
#
# Install script for zenodo-cli
# Usage: curl -fsSL https://raw.githubusercontent.com/ran-codes/zenodo-cli/main/scripts/install.sh | bash
#

set -e

REPO="ran-codes/zenodo-cli"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Map OS names
case "$OS" in
    linux)
        OS="linux"
        ;;
    darwin)
        OS="darwin"
        ;;
    mingw*|msys*|cygwin*)
        OS="windows"
        EXT=".exe"
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

echo "Detected: ${OS}/${ARCH}"

# Get latest release tag
LATEST_TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d '"' -f 4)

if [ -z "$LATEST_TAG" ]; then
    echo "Could not determine latest release"
    exit 1
fi

echo "Latest release: ${LATEST_TAG}"

# Download both binaries
for BINARY in zenodo zenodo-mcp; do
    BINARY_NAME="${BINARY}-${OS}-${ARCH}${EXT}"
    URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${BINARY_NAME}"

    echo "Downloading ${BINARY_NAME}..."

    TMP_FILE=$(mktemp)
    if ! curl -fsSL "$URL" -o "$TMP_FILE"; then
        echo "Could not download ${BINARY_NAME}"
        echo "Please check https://github.com/${REPO}/releases for available downloads"
        rm -f "$TMP_FILE"
        exit 1
    fi

    chmod +x "$TMP_FILE"

    INSTALL_PATH="${INSTALL_DIR}/${BINARY}${EXT}"

    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_FILE" "$INSTALL_PATH"
    else
        echo "Installing to ${INSTALL_PATH} (requires sudo)..."
        sudo mv "$TMP_FILE" "$INSTALL_PATH"
    fi

    echo "Installed ${BINARY} to ${INSTALL_PATH}"
done

echo ""
echo "Get started:"
echo "  zenodo config set token <your-zenodo-token>"
echo "  zenodo records list"
echo ""
echo "Run 'zenodo --help' for more information."
