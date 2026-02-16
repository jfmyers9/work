#!/bin/bash

# Update Homebrew formula with version and SHA256 hashes from dist/ tarballs.
# Usage: ./scripts/update-formula.sh <version>
# Example: ./scripts/update-formula.sh 0.1.0

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <version>" >&2
    exit 1
fi

VERSION="$1"
FORMULA="homebrew/work.rb"
DIST="dist"

ARM64_ARCHIVE="${DIST}/work-v${VERSION}-darwin-arm64.tar.gz"
AMD64_ARCHIVE="${DIST}/work-v${VERSION}-darwin-amd64.tar.gz"

for f in "$ARM64_ARCHIVE" "$AMD64_ARCHIVE"; do
    if [ ! -f "$f" ]; then
        echo "Missing archive: $f" >&2
        echo "Run ./scripts/build.sh first." >&2
        exit 1
    fi
done

ARM64_SHA256=$(shasum -a 256 "$ARM64_ARCHIVE" | awk '{print $1}')
AMD64_SHA256=$(shasum -a 256 "$AMD64_ARCHIVE" | awk '{print $1}')

sed -i '' "s/version \".*\"/version \"${VERSION}\"/" "$FORMULA"
sed -i '' "s/sha256 \"REPLACE_WITH_ARM64_SHA256\"/sha256 \"${ARM64_SHA256}\"/" "$FORMULA"
sed -i '' "s/sha256 \"REPLACE_WITH_AMD64_SHA256\"/sha256 \"${AMD64_SHA256}\"/" "$FORMULA"

# Also handle updating existing hashes (not just placeholders)
sed -i '' "/darwin-arm64.tar.gz/{n;s/sha256 \"[a-f0-9]\{64\}\"/sha256 \"${ARM64_SHA256}\"/;}" "$FORMULA"
sed -i '' "/darwin-amd64.tar.gz/{n;s/sha256 \"[a-f0-9]\{64\}\"/sha256 \"${AMD64_SHA256}\"/;}" "$FORMULA"

echo "Updated $FORMULA for v${VERSION}"
echo "  arm64: ${ARM64_SHA256}"
echo "  amd64: ${AMD64_SHA256}"
