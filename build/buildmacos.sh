#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Variables
OUTPUT_DIR="../build"  # Adjusted to place the binaries in the build directory
BINARY_NAME_AMD64="goanime-darwin-amd64"
BINARY_NAME_ARM64="goanime-darwin-arm64"
BINARY_NAME_UNIVERSAL="goanime-darwin-universal"
BINARY_NAME_UNIVERSAL_GENERIC="goanime-darwin"
BINARY_PATH_AMD64="$OUTPUT_DIR/$BINARY_NAME_AMD64"
BINARY_PATH_ARM64="$OUTPUT_DIR/$BINARY_NAME_ARM64"
BINARY_PATH_UNIVERSAL="$OUTPUT_DIR/$BINARY_NAME_UNIVERSAL"
BINARY_PATH_UNIVERSAL_GENERIC="$OUTPUT_DIR/$BINARY_NAME_UNIVERSAL_GENERIC"
MAIN_PACKAGE="../cmd/goanime"

# Create the output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

echo "Building goanime binaries for macOS..."

# Build for Intel (amd64)
echo "Building for Intel (amd64)..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o "$BINARY_PATH_AMD64" -ldflags="-s -w" -trimpath "$MAIN_PACKAGE"
echo "Intel build completed: $BINARY_PATH_AMD64"

# Build for Apple Silicon (arm64)
echo "Building for Apple Silicon (arm64)..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o "$BINARY_PATH_ARM64" -ldflags="-s -w" -trimpath "$MAIN_PACKAGE"
echo "Apple Silicon build completed: $BINARY_PATH_ARM64"

# Create universal binary using lipo
echo "Creating universal binary..."
if command -v lipo >/dev/null 2>&1; then
    lipo -create -output "$BINARY_PATH_UNIVERSAL" "$BINARY_PATH_AMD64" "$BINARY_PATH_ARM64"
    echo "Universal binary created: $BINARY_PATH_UNIVERSAL"
    
    # Create a copy with generic name for updater compatibility
    cp "$BINARY_PATH_UNIVERSAL" "$BINARY_PATH_UNIVERSAL_GENERIC"
    echo "Generic universal binary created: $BINARY_PATH_UNIVERSAL_GENERIC"
else
    echo "Warning: lipo command not found. Cannot create universal binary."
fi

# Function to compress binary with UPX
compress_binary() {
    local binary_path="$1"
    local binary_name=$(basename "$binary_path")
    
    if command -v upx >/dev/null 2>&1; then
        echo "Compressing $binary_name with UPX..."
        if upx --best --ultra-brute --force-macos "$binary_path" 2>/dev/null; then
            echo "Compression completed for $binary_name."
        else
            echo "UPX compression failed for $binary_name. Continuing without compression."
        fi
    else
        echo "UPX not found. Skipping compression for $binary_name."
    fi
}

# Compress binaries
compress_binary "$BINARY_PATH_AMD64"
compress_binary "$BINARY_PATH_ARM64"
if [ -f "$BINARY_PATH_UNIVERSAL" ]; then
    compress_binary "$BINARY_PATH_UNIVERSAL"
fi
if [ -f "$BINARY_PATH_UNIVERSAL_GENERIC" ]; then
    compress_binary "$BINARY_PATH_UNIVERSAL_GENERIC"
fi

# Check if binaries were built successfully
for binary in "$BINARY_PATH_AMD64" "$BINARY_PATH_ARM64"; do
    if [ ! -f "$binary" ]; then
        echo "Error: Binary not found at $binary. Build may have failed."
        exit 1
    fi
done

# Function to create tarball and checksum
create_tarball_and_checksum() {
    local binary_path="$1"
    local binary_name=$(basename "$binary_path")
    local tarball_name="$binary_name.tar.gz"
    local tarball_path="$OUTPUT_DIR/$tarball_name"
    local checksum_file="$tarball_path.sha256"
    
    # Create tarball
    echo "Creating tarball for $binary_name..."
    tar -czf "$tarball_path" -C "$OUTPUT_DIR" "$binary_name"
    echo "Tarball created: $tarball_path"
    
    # Generate SHA256 checksum for the tarball
    echo "Generating SHA256 checksum for $tarball_name..."
    if command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$tarball_path" > "$checksum_file"
    elif command -v openssl >/dev/null 2>&1; then
        openssl dgst -sha256 "$tarball_path" | awk '{print $2}' > "$checksum_file"
    else
        echo "Neither shasum nor openssl is available. Cannot generate checksum for $tarball_name."
        return 1
    fi
    echo "Checksum generated: $checksum_file"
}

# Create tarballs and checksums for all binaries
create_tarball_and_checksum "$BINARY_PATH_AMD64"
create_tarball_and_checksum "$BINARY_PATH_ARM64"
if [ -f "$BINARY_PATH_UNIVERSAL" ]; then
    create_tarball_and_checksum "$BINARY_PATH_UNIVERSAL"
fi
if [ -f "$BINARY_PATH_UNIVERSAL_GENERIC" ]; then
    create_tarball_and_checksum "$BINARY_PATH_UNIVERSAL_GENERIC"
fi

echo "Build script completed successfully. Generated binaries:"
echo "- Intel (amd64): $BINARY_PATH_AMD64"
echo "- Apple Silicon (arm64): $BINARY_PATH_ARM64"
if [ -f "$BINARY_PATH_UNIVERSAL" ]; then
    echo "- Universal (explicit): $BINARY_PATH_UNIVERSAL"
fi
if [ -f "$BINARY_PATH_UNIVERSAL_GENERIC" ]; then
    echo "- Universal (generic): $BINARY_PATH_UNIVERSAL_GENERIC"
fi

echo ""
echo "GitHub Release Assets:"
echo "- goanime-darwin-amd64 (Intel macOS)"
echo "- goanime-darwin-arm64 (Apple Silicon macOS)"
if [ -f "$BINARY_PATH_UNIVERSAL" ]; then
    echo "- goanime-darwin-universal (Universal macOS - explicit)"
fi
if [ -f "$BINARY_PATH_UNIVERSAL_GENERIC" ]; then
    echo "- goanime-darwin (Universal macOS - fallback for updater)"
fi
