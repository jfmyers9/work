#!/bin/bash

# Build script for work
# Supports cross-compilation for macOS (darwin) on both Intel and Apple Silicon

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

BUILD_DIR="./dist"
BINARY_NAME="work"

LDFLAGS="-X github.com/jfmyers9/work/cmd.version=${VERSION} \
         -X github.com/jfmyers9/work/cmd.commit=${COMMIT} \
         -X github.com/jfmyers9/work/cmd.buildDate=${BUILD_DATE}"

echo -e "${GREEN}Building work${NC}"
echo "Version:    ${VERSION}"
echo "Commit:     ${COMMIT}"
echo "Build Date: ${BUILD_DATE}"
echo ""

if [ -d "$BUILD_DIR" ]; then
    echo -e "${YELLOW}Cleaning previous builds...${NC}"
    rm -rf "$BUILD_DIR"
fi
mkdir -p "$BUILD_DIR"

echo -e "${GREEN}Building for darwin/amd64...${NC}"
GOOS=darwin GOARCH=amd64 go build \
    -ldflags "$LDFLAGS" \
    -o "${BUILD_DIR}/${BINARY_NAME}-darwin-amd64" \
    .

echo -e "${GREEN}Building for darwin/arm64...${NC}"
GOOS=darwin GOARCH=arm64 go build \
    -ldflags "$LDFLAGS" \
    -o "${BUILD_DIR}/${BINARY_NAME}-darwin-arm64" \
    .

echo -e "${GREEN}Creating universal binary...${NC}"
lipo -create \
    "${BUILD_DIR}/${BINARY_NAME}-darwin-amd64" \
    "${BUILD_DIR}/${BINARY_NAME}-darwin-arm64" \
    -output "${BUILD_DIR}/${BINARY_NAME}"

echo -e "${GREEN}Generating checksums...${NC}"
cd "$BUILD_DIR"
shasum -a 256 ${BINARY_NAME}-darwin-amd64 > ${BINARY_NAME}-darwin-amd64.sha256
shasum -a 256 ${BINARY_NAME}-darwin-arm64 > ${BINARY_NAME}-darwin-arm64.sha256
shasum -a 256 ${BINARY_NAME} > ${BINARY_NAME}.sha256
cd - > /dev/null

echo -e "${GREEN}Creating release archives...${NC}"
cd "$BUILD_DIR"
tar -czf "${BINARY_NAME}-${VERSION}-darwin-amd64.tar.gz" ${BINARY_NAME}-darwin-amd64
tar -czf "${BINARY_NAME}-${VERSION}-darwin-arm64.tar.gz" ${BINARY_NAME}-darwin-arm64
tar -czf "${BINARY_NAME}-${VERSION}-darwin-universal.tar.gz" ${BINARY_NAME}
cd - > /dev/null

echo -e "${GREEN}✓ Build complete!${NC}"
echo ""
echo "Artifacts:"
ls -lh "${BUILD_DIR}"
