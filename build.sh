#!/bin/bash

# Build script for batchMedia
# Builds for current platform with full HEIF support

set -e

# Create bin directory if it doesn't exist
mkdir -p bin

echo "Building batchMedia for current platform..."

# Build for current platform (with HEIF support)
echo "Building with HEIF/HEIC support..."
go build -o bin/batchMedia

echo "Build completed successfully!"
echo ""
echo "Generated files:"
ls -la bin/batchMedia
echo ""
echo "Build includes full HEIF/HEIC support along with JPEG and PNG processing."
echo ""
echo "For cross-platform builds:"
echo "- Use 'go build' on target platforms for native builds with HEIF support"
echo "- Or use Docker with appropriate base images for each target platform"
echo "- CGO dependencies make cross-compilation complex but native builds work perfectly"