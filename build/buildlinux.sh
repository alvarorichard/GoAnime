#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Variables
OUTPUT_DIR="../build"  # Adjusted to place the binary in the build directory
BINARY_NAME="goanime"
BINARY_PATH="$OUTPUT_DIR/$BINARY_NAME"
CHECKSUM_FILE="$OUTPUT_DIR/$BINARY_NAME.sha256"
MAIN_PACKAGE="../cmd/goanime"

# Create the output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

echo "Building the goanime binary for Linux..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$BINARY_PATH" -ldflags="-s -w" -trimpath "$MAIN_PACKAGE"

echo "Build completed: $BINARY_PATH"

# Check if UPX is installed
if command -v upx >/dev/null 2>&1; then
    echo "Compressing the binary with UPX..."
    upx --best --ultra-brute "$BINARY_PATH"
    echo "Compression completed."
else
    echo "UPX not found. Skipping compression."
fi

# Generate SHA256 checksum
echo "Generating SHA256 checksum..."
# Check if sha256sum exists, else use shasum
if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$BINARY_PATH" > "$CHECKSUM_FILE"
elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$BINARY_PATH" > "$CHECKSUM_FILE"
else
    echo "Neither sha256sum nor shasum is available. Cannot generate checksum."
    exit 1
fi
echo "Checksum generated: $CHECKSUM_FILE"

echo "Build script completed successfully."
