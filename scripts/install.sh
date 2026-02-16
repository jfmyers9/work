#!/bin/bash
#
# Installation script for work
# Usage: curl -fsSL https://raw.githubusercontent.com/jfmyers9/work/main/scripts/install.sh | bash
#

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

REPO="jfmyers9/work"
BINARY_NAME="work"
INSTALL_DIR="/usr/local/bin"

ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        PLATFORM="darwin-amd64"
        ;;
    arm64)
        PLATFORM="darwin-arm64"
        ;;
    *)
        echo -e "${RED}✗ Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo -e "${GREEN}✓${NC} Detected platform: macOS ($ARCH)"

echo -e "${YELLOW}→${NC} Fetching latest release..."
LATEST_VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo -e "${RED}✗ Failed to fetch latest version${NC}"
    exit 1
fi

echo -e "${GREEN}✓${NC} Latest version: $LATEST_VERSION"

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/${BINARY_NAME}-${LATEST_VERSION}-${PLATFORM}.tar.gz"
CHECKSUM_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/${BINARY_NAME}-${PLATFORM}.sha256"

TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

echo -e "${YELLOW}→${NC} Downloading work $LATEST_VERSION for $PLATFORM..."
cd "$TMP_DIR"

if ! curl -fsSL -o "${BINARY_NAME}.tar.gz" "$DOWNLOAD_URL"; then
    echo -e "${RED}✗ Failed to download work${NC}"
    exit 1
fi

echo -e "${GREEN}✓${NC} Downloaded successfully"

if curl -fsSL -o "${BINARY_NAME}.sha256" "$CHECKSUM_URL" 2>/dev/null; then
    echo -e "${YELLOW}→${NC} Verifying checksum..."
    EXPECTED_HASH=$(cat "${BINARY_NAME}.sha256" | awk '{print $1}')
    ACTUAL_HASH=$(shasum -a 256 "${BINARY_NAME}.tar.gz" | awk '{print $1}')

    if [ "$EXPECTED_HASH" = "$ACTUAL_HASH" ]; then
        echo -e "${GREEN}✓${NC} Checksum verified"
    else
        echo -e "${RED}✗ Checksum verification failed${NC}"
        exit 1
    fi
fi

echo -e "${YELLOW}→${NC} Extracting archive..."
tar -xzf "${BINARY_NAME}.tar.gz"

echo -e "${YELLOW}→${NC} Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    cp "${BINARY_NAME}-${PLATFORM}" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    sudo cp "${BINARY_NAME}-${PLATFORM}" "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

echo -e "${GREEN}✓${NC} work installed successfully"
echo ""
echo -e "${GREEN}✓${NC} $($BINARY_NAME version)"
echo ""
echo "Get started:"
echo "  work init"
echo "  work create \"My first issue\""
