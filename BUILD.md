# Build Instructions

## Cross-Platform Compilation

This project supports cross-platform compilation for multiple operating systems and architectures.

### Quick Build

Use the provided build script to generate binaries for all supported platforms:

```bash
./build.sh
```

### Manual Build Commands

#### macOS (with HEIF support)
```bash
# Intel Macs
GOOS=darwin GOARCH=amd64 go build -o batchMedia-macos-amd64

# Apple Silicon Macs
GOOS=darwin GOARCH=arm64 go build -o batchMedia-macos-arm64
```

#### Linux (without HEIF support)
```bash
# x86_64
GOOS=linux GOARCH=amd64 go build -tags noheif -o batchMedia-linux-amd64

# ARM64
GOOS=linux GOARCH=arm64 go build -tags noheif -o batchMedia-linux-arm64
```

#### Windows (without HEIF support)
```bash
# x86_64
GOOS=windows GOARCH=amd64 go build -tags noheif -o batchMedia-windows-amd64.exe
```

## HEIF Support

### macOS
- **Full HEIF support** including `.heic` files
- Uses the `github.com/jdeng/goheif` library with CGO
- Supports HEIF to JPEG conversion

### Linux/Windows
- **No HEIF support** due to CGO cross-compilation limitations
- Only supports JPEG (`.jpg`, `.jpeg`) and PNG (`.png`) files
- HEIC files will be skipped with an error message

## Build Tags

- **Default build**: Includes HEIF support (requires CGO)
- **`-tags noheif`**: Disables HEIF support for cross-compilation

## Dependencies

### With HEIF Support (macOS)
- `github.com/jdeng/goheif` - HEIF image processing
- `github.com/nfnt/resize` - Image resizing
- `github.com/rwcarlsen/goexif` - EXIF data handling
- `github.com/u2takey/ffmpeg-go` - Video processing

### Without HEIF Support (Linux/Windows)
- `github.com/nfnt/resize` - Image resizing
- `github.com/rwcarlsen/goexif` - EXIF data handling
- `github.com/u2takey/ffmpeg-go` - Video processing

## System Requirements

### Target System: Linux fnNate 6.12.18-trim x86_64

For your specific Linux system, use:
```bash
GOOS=linux GOARCH=amd64 go build -tags noheif -o batchMedia-linux
```

The generated `batchMedia-linux` binary is compatible with:
- Linux kernel 6.12.18 and later
- x86_64 architecture
- GNU/Linux systems

## Usage Notes

1. **HEIF Limitation**: Linux and Windows builds cannot process `.heic` files
2. **Static Linking**: All builds are statically linked and don't require external dependencies
3. **Performance**: Cross-compiled binaries have the same performance as native builds
4. **File Support**: 
   - macOS: JPEG, PNG, HEIC, MP4, MOV, AVI, MKV, etc.
   - Linux/Windows: JPEG, PNG, MP4, MOV, AVI, MKV, etc. (no HEIC)

## Troubleshooting

If you encounter issues with HEIF files on Linux/Windows builds:
1. Convert HEIC files to JPEG on macOS first
2. Use the macOS build for HEIC processing
3. Transfer processed files to your target system