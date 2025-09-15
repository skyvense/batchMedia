# batchMedia - Batch Media Processing Tool

A command-line tool written in Go for batch processing JPEG image files. Supports resizing images by ratio or specified width while preserving original file modification dates.

## Features

- 🖼️ Batch processing of JPEG/JPG format images
- 📏 Two scaling modes: proportional scaling and width-based scaling
- 📅 Preserves original file modification dates
- 📁 Recursive processing of subdirectories
- 🚀 Uses Go standard library with minimal external dependencies
- ⚡ High-performance batch processing

## Installation

### Build from Source

```bash
git clone <repository-url>
cd batchMedia
go build -o batchMedia
```

## Usage

### Basic Syntax

```bash
./batchMedia -inputdir=<input_directory> -out=<output_directory> [scaling_options]
```

### Scaling Options

- `-size=<ratio>`: Scale by ratio (e.g., 0.5 means scale down to 50%)
- `-width=<pixels>`: Scale by specified width, automatically maintains aspect ratio

**Note: `-size` and `-width` parameters cannot be used simultaneously**

### Usage Examples

#### 1. Scale by Ratio
Scale images down to 50% of original size:
```bash
./batchMedia -inputdir=./photos/2019 -out=./photos/2019_resized -size=0.5
```

#### 2. Scale by Width
Resize image width to 1920 pixels, height automatically adjusted proportionally:
```bash
./batchMedia -inputdir=./photos/2019 -out=./photos/2019_1920 -width=1920
```


#### 3. Create Test Images
Running the program without any parameters will automatically create test images:
```bash
./batchMedia
```

### Parameter Description

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `-inputdir` | string | Yes | Input directory path containing JPEG files to process |
| `-out` | string | Yes | Output directory path where processed files will be saved |
| `-size` | float | No | Scaling ratio, range 0-10 (mutually exclusive with -width) |
| `-width` | int | No | Target width in pixels (mutually exclusive with -size) |
| `-h` | - | No | Display help information |

## How It Works

1. **File Discovery**: Recursively scans input directory to find all `.jpg` and `.jpeg` files
2. **Image Processing**:
   - Decodes JPEG images
   - Calculates new dimensions based on specified parameters
   - Uses Lanczos3 algorithm for high-quality image scaling
   - Re-encodes to JPEG format (90% quality)
3. **File Saving**:
   - Maintains original directory structure
   - Preserves original file modification times
   - Automatically creates necessary output directories

## Technical Features

- **Programming Language**: Go 1.21+
- **Image Processing**: Uses Go standard library `image` and `image/jpeg` packages, plus `nfnt/resize` for high-quality scaling
- **Command Line Parsing**: Uses Go standard library `flag` package
- **File Operations**: Cross-platform file system operations support
- **Algorithm**: Lanczos3 interpolation algorithm for image scaling
- **EXIF Support**: Preserves EXIF metadata using `goexif` library

## Performance Notes

- Uses in-memory image processing, suitable for batch processing of medium-sized images
- For very large image files, recommend batch processing to avoid memory shortage
- Processing speed depends on image size, quantity, and system performance

## Important Notes

1. **Supported Formats**: Currently only supports JPEG/JPG format
2. **Output Quality**: Output JPEG quality is fixed at 90%
3. **Memory Usage**: Large images will consume more memory
4. **File Overwriting**: Existing output files will be overwritten
5. **Directory Structure**: Maintains the relative path structure of input directory

## Error Handling

The program will report errors and exit in the following situations:
- Input directory does not exist
- Missing required parameters
- Both size and width parameters specified simultaneously
- Parameter values outside valid range
- Insufficient file read/write permissions

## Sample Output

```
Processing file: photos/2019/IMG_001.jpg
Processing completed: photos/2019/IMG_001.jpg (4032x3024 -> 2016x1512)
Processing file: photos/2019/IMG_002.jpg
Processing completed: photos/2019/IMG_002.jpg (3840x2160 -> 1920x1080)
Batch processing completed!
```

## License

This project is licensed under the MIT License.