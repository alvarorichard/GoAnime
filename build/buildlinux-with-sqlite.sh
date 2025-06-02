#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Variables
OUTPUT_DIR="../build"
BINARY_NAME="goanime-linux"
BINARY_PATH="$OUTPUT_DIR/$BINARY_NAME"
TARBALL_NAME="$BINARY_NAME.tar.gz"
TARBALL_PATH="$OUTPUT_DIR/$TARBALL_NAME"
CHECKSUM_FILE="$TARBALL_PATH.sha256"
MAIN_PACKAGE="../cmd/goanime"

# Create the output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

echo "Building the goanime binary for Linux with SQLite support..."
echo "Note: This build enables CGO and requires SQLite development libraries"

# Check if gcc and sqlite3 development libraries are installed
if ! command -v gcc >/dev/null 2>&1; then
    echo "Error: GCC not found. Please install GCC to build with CGO."
    exit 1
fi

if ! pkg-config --exists sqlite3 2>/dev/null; then
    echo "Warning: sqlite3 development libraries not found."
    echo "On Ubuntu/Debian: sudo apt-get install libsqlite3-dev"
    echo "On Fedora: sudo dnf install sqlite-devel"
    echo "On Arch: sudo pacman -S sqlite"
    echo "Continue anyway? (y/n)"
    read -r response
    if [[ "$response" != "y" ]]; then
        exit 1
    fi
fi

# For Linux we don't need go-winio, but we need SQLite with CGO enabled
# Using !windows tag to ensure go-winio dependencies are excluded
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o "$BINARY_PATH" -ldflags="-s -w" -trimpath -tags="!windows" "$MAIN_PACKAGE"

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
echo ""
echo "Note: This build includes SQLite support for tracking anime progress."
echo "The binary is larger than the standard build but includes full tracking features."
