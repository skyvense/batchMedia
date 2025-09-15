#!/bin/bash

# Build script for batchMedia
# Supports building for different platforms with optional HEIF support

set -e

echo "Building batchMedia for multiple platforms..."

# Build for Linux (without HEIF support due to CGO requirements)
echo "Building for Linux x86_64 (without HEIF support)..."
GOOS=linux GOARCH=amd64 go build -tags noheif -o bin/batchMedia-linux-amd64

echo "Building for Linux ARM64 (without HEIF support)..."
GOOS=linux GOARCH=arm64 go build -tags noheif -o bin/batchMedia-linux-arm64

# Build for Windows (without HEIF support)
echo "Building for Windows x86_64 (without HEIF support)..."
GOOS=windows GOARCH=amd64 go build -tags noheif -o bin/batchMedia-windows-amd64.exe

echo "Build completed successfully!"
echo ""
echo "Generated files:"
ls -la bin/batchMedia-*
echo ""
echo "Note: Linux and Windows versions do not support HEIF files due to CGO cross-compilation limitations."
echo "Only JPEG and PNG files are supported in these builds."