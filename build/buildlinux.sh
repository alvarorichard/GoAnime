#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Variables
OUTPUT_DIR="../build"  # Adjusted to place the binary in the build directory
BINARY_NAME="goanime-linux"
BINARY_PATH="$OUTPUT_DIR/$BINARY_NAME"
TARBALL_NAME="$BINARY_NAME.tar.gz"
TARBALL_PATH="$OUTPUT_DIR/$TARBALL_NAME"
CHECKSUM_FILE="$TARBALL_PATH.sha256"
MAIN_PACKAGE="../cmd/goanime"

# Create the output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

echo "Building the goanime binary for Linux..."
# For Linux we don't need go-winio, using !windows tag to ensure it's excluded
# With CGO_ENABLED=0, SQLite will be compiled without native support
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$BINARY_PATH" -ldflags="-s -w" -trimpath -tags="!windows" "$MAIN_PACKAGE"

echo "Build completed: $BINARY_PATH"

# Check if UPX is installed
if command -v upx >/dev/null 2>&1; then
    echo "Compressing the binary with UPX..."
    upx --best --ultra-brute "$BINARY_PATH"
    echo "Compression completed."
else
    echo "UPX not found. Skipping compression."
fi

# Create tarball
echo "Creating tarball..."
tar -czf "$TARBALL_PATH" -C "$OUTPUT_DIR" "$BINARY_NAME"
echo "Tarball created: $TARBALL_PATH"

# Generate SHA256 checksum for the tarball
echo "Generating SHA256 checksum for the tarball..."
# Check if sha256sum exists, else use shasum
if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$TARBALL_PATH" > "$CHECKSUM_FILE"
elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$TARBALL_PATH" > "$CHECKSUM_FILE"
else
    echo "Neither sha256sum nor shasum is available. Cannot generate checksum."
    exit 1
fi
echo "Checksum generated: $CHECKSUM_FILE"

echo "Build script completed successfully."
