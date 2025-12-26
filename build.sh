#!/bin/bash
# Local build script for easyGit
# Created for testing version injection and local development

set -e

# Get version from git tag or use default
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev-$(date +%Y%m%d)")
MAIN_PACKAGE="./cmd/easygit"
OUTPUT_NAME="easyGit"

echo "Building easyGit version: $VERSION"

# Build with version injection
go build -ldflags="-s -w -X github.com/KevinYouu/easyGit/internal/version.Version=$VERSION" \
    -o "$OUTPUT_NAME" "$MAIN_PACKAGE"

echo "Build completed: $OUTPUT_NAME"
echo "Run './$OUTPUT_NAME version' to verify version injection"
