# batchMedia - Batch Media Processing Tool

A powerful command-line tool written in Go for batch processing images and videos. Supports multiple formats, intelligent scaling, HDR video processing, and generates beautiful HTML reports.

## Features

### Image Processing
- 🖼️ **Multi-format Support**: JPEG/JPG, PNG, and HEIC images
- 📏 **Flexible Scaling**: Proportional scaling and width-based scaling
- 🔄 **HEIC Conversion**: Automatic HEIC to JPEG conversion with EXIF preservation
- 📊 **EXIF Metadata**: Preserves and transfers EXIF data between formats

### Video Processing
- 🎬 **Video Support**: MOV, MP4, AVI, MKV formats
- 🌈 **HDR Preservation**: Maintains HDR metadata for high-quality video output
- 🎯 **Smart Encoding**: H.264/H.265 with QuickTime compatibility
- ⚙️ **Flexible Parameters**: Customizable CRF, bitrate, and resolution settings

### Advanced Features
- 📅 **Metadata Preservation**: Maintains original file modification dates
- 📁 **Directory Structure**: Recursive processing with preserved folder hierarchy
- 📋 **HTML Reports**: Beautiful, interactive reports with thumbnails and file links
- 🚀 **High Performance**: Optimized batch processing with minimal dependencies
- 🎛️ **Smart Filtering**: Intelligent resolution-based processing decisions

## Installation

### Prerequisites
- **Go 1.21 or higher** - Required for building the application
- **Git** - For cloning the repository
- **FFmpeg** - Required for video processing (optional for image-only workflows)

#### Installing FFmpeg

**macOS (using Homebrew):**
```bash
brew install ffmpeg
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install ffmpeg
```

**Windows:**
- Download from [FFmpeg official website](https://ffmpeg.org/download.html)
- Add FFmpeg to your system PATH

### Build from Source

1. Clone the repository:
```bash
git clone <repository-url>
cd batchMedia
```

2. Install Go dependencies:
```bash
go mod tidy
```

3. Build the executable:
```bash
go build -o batchMedia
```

4. Verify installation:
```bash
./batchMedia -h
```

## Usage

### Basic Syntax

```bash
./batchMedia -inputdir=<input_directory> -out=<output_directory> [options]
```

### Image Processing Options

- `-size=<ratio>`: Scale by ratio (e.g., 0.5 means scale down to 50%)
- `-width=<pixels>`: Scale by specified width, automatically maintains aspect ratio

**Note: `-size` and `-width` parameters cannot be used simultaneously**

### Video Processing Options

- `-disable-video`: Disable video processing (video processing is enabled by default)
- `-video-codec=<codec>`: Video codec (libx264, libx265) - default: libx265
- `-video-bitrate=<bitrate>`: Video bitrate (e.g., 2M, 1000k)
- `-video-resolution=<resolution>`: Video resolution (e.g., 1920x1080, 1280x720)
- `-video-crf=<value>`: Video CRF quality (0-51, lower is better) - default: 23
- `-video-preset=<preset>`: Encoding preset (ultrafast, fast, medium, slow, veryslow) - default: medium

### Resolution Filtering Options

- `-threshold-width=<pixels>`: Width threshold for filtering (default: 1920 for downscaling, 3840 for upscaling)
- `-threshold-height=<pixels>`: Height threshold for filtering (default: 1080 for downscaling, 2160 for upscaling)
- `-ignore-smart-limit`: Ignore smart default resolution limits

**Smart Threshold Logic:**
- For **downscaling** (scaling ratio < 1.0): Skip images **below** the threshold (too small to downscale effectively)
- For **upscaling** (scaling ratio > 1.0): Skip images **above** the threshold (too large to upscale effectively)
- The scaling ratio is determined by the `-size` parameter or calculated from `-width` parameter
- Images outside the specified resolution range will be copied to output directory without scaling

### Usage Examples

#### Image Processing Examples

##### 1. Scale Images by Ratio
Scale images down to 50% of original size:
```bash
./batchMedia -inputdir=./photos/2019 -out=./photos/2019_resized -size=0.5
```

##### 2. Scale Images by Width
Resize image width to 1920 pixels, height automatically adjusted proportionally:
```bash
./batchMedia -inputdir=./photos/2019 -out=./photos/2019_1920 -width=1920
```

##### 3. Process HEIC Files
Convert HEIC files to JPEG while preserving EXIF data:
```bash
./batchMedia -inputdir=./iphone_photos -out=./converted_photos -size=1.0
```

#### Video Processing Examples

##### 4. Basic Video Processing
Process videos with default H.265 encoding:
```bash
./batchMedia -inputdir=./videos -out=./compressed_videos
```

##### 5. High-Quality Video Encoding
Process videos with custom quality settings:
```bash
./batchMedia -inputdir=./videos -out=./hq_videos -video-crf=18 -video-preset=slow
```

##### 6. Video Resolution Scaling
Scale videos to 1080p resolution:
```bash
./batchMedia -inputdir=./4k_videos -out=./1080p_videos -video-resolution=1920x1080
```

#### Mixed Media Processing Examples

##### 7. Process Images and Videos Together
Process both images and videos in the same directory:
```bash
./batchMedia -inputdir=./mixed_media -out=./processed_media -size=0.8 -video-crf=20
```

##### 8. Advanced Filtering
Downscale images but skip small ones, with custom video settings:
```bash
./batchMedia -inputdir=./media -out=./filtered_media -size=0.5 -threshold-width=1000 -threshold-height=1000 -video-codec=libx264
```


#### 3. Create Test Images
Running the program without any parameters will automatically create test images:
```bash
./batchMedia
```

### Parameter Description

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| **Basic Parameters** |
| `-inputdir` | string | Yes | Input directory path containing media files to process |
| `-out` | string | Yes | Output directory path where processed files will be saved |
| **Image Processing** |
| `-size` | float | No | Scaling ratio, range 0-10 (mutually exclusive with -width) |
| `-width` | int | No | Target width in pixels (mutually exclusive with -size) |
| `-threshold-width` | int | No | Width threshold (default: 1920 for downscaling, 3840 for upscaling) |
| `-threshold-height` | int | No | Height threshold (default: 1080 for downscaling, 2160 for upscaling) |
| `-ignore-smart-limit` | bool | No | Ignore smart default resolution limits |
| **Video Processing** |
| `-disable-video` | bool | No | Disable video processing (video processing is enabled by default) |
| `-video-codec` | string | No | Video codec: libx264, libx265 (default: libx265) |
| `-video-bitrate` | string | No | Video bitrate (e.g., 2M, 1000k) |
| `-video-resolution` | string | No | Video resolution (e.g., 1920x1080, 1280x720) |
| `-video-crf` | int | No | Video CRF quality, 0-51, lower is better (default: 23) |
| `-video-preset` | string | No | Encoding preset: ultrafast, fast, medium, slow, veryslow (default: medium) |
| **Other** |
| `-h` | - | No | Display help information |

## How It Works

1. **File Discovery**: Recursively scans input directory to find all `.jpg`, `.jpeg`, and `.heic` files
2. **Smart Filtering**: 
   - Determines operation type based on scaling ratio (> 1.0 = upscaling, < 1.0 = downscaling)
   - Applies threshold filtering to skip inappropriate images
   - Uses intelligent defaults: 1920x1080 for downscaling, 3840x2160 for upscaling
3. **Image Processing**:
   - Decodes JPEG and HEIC images
   - Calculates new dimensions based on specified parameters
   - Uses Lanczos3 algorithm for high-quality image scaling
   - Re-encodes to JPEG format (90% quality)
4. **File Saving**:
   - Maintains original directory structure
   - Preserves original file modification times
   - Automatically creates necessary output directories

## Technical Features

### Core Processing
- **Multi-format Support**: Supports JPEG, PNG, HEIC images and various video formats (MP4, MOV, AVI, MKV, etc.)
- **EXIF Data Preservation**: Automatically preserves EXIF metadata during image processing
- **HEIC to JPEG Conversion**: Seamlessly converts HEIC files to JPEG format
- **Smart Resolution Filtering**: Intelligent threshold system to avoid unnecessary processing
- **Batch Processing**: Efficiently processes multiple files in parallel
- **Memory Optimization**: Optimized memory usage for large file batches

### Video Processing
- **Modern Codecs**: Supports H.264 and H.265 (HEVC) encoding
- **Quality Control**: CRF-based quality settings for optimal compression
- **Resolution Scaling**: Flexible video resolution adjustment
- **Preset Options**: Multiple encoding speed/quality presets
- **FFmpeg Integration**: Leverages FFmpeg for robust video processing

### Reporting & Analysis
- **HTML Reports**: Generates beautiful, interactive HTML reports
- **Grid Layout**: Modern card-based layout with thumbnails
- **File Statistics**: Detailed processing statistics and file information
- **Clickable Links**: Direct file access from the report
- **Visual Thumbnails**: Preview images and video frames

### System Compatibility
- **Programming Language**: Go 1.21+
- **Image Processing**: Uses Go standard library `image` and `image/jpeg` packages, plus `jdeng/goheif` for HEIC support and `nfnt/resize` for high-quality scaling
- **Smart Logic**: Simplified threshold filtering based on scaling ratio comparison
- **Command Line Parsing**: Uses Go standard library `flag` package
- **File Operations**: Cross-platform file system operations support
- **Algorithm**: Lanczos3 interpolation algorithm for image scaling
- **EXIF Support**: Preserves EXIF metadata for both JPEG and HEIC files using `goexif` and `jdeng/goheif` libraries

## Performance Notes

- Uses in-memory image processing, suitable for batch processing of medium-sized images
- For very large image files, recommend batch processing to avoid memory shortage
- Processing speed depends on image size, quantity, and system performance

## Important Notes

1. **Supported Formats**: Supports JPEG/JPG and HEIC formats
2. **Output Quality**: Output JPEG quality is fixed at 90%
3. **EXIF Metadata**: EXIF data is preserved for both JPEG and HEIC files
4. **Memory Usage**: Large images will consume more memory
5. **File Overwriting**: Existing output files will be overwritten
6. **Directory Structure**: Maintains the relative path structure of input directory

## Error Handling

The program will report errors and exit in the following situations:
- Input directory does not exist
- Missing required parameters
- Both size and width parameters specified simultaneously
- Parameter values outside valid range
- Insufficient file read/write permissions

## Sample Output

### Console Output
```
Batch Media Processing Tool
============================
Input Directory: ./test_media
Output Directory: ./output
Image Scaling: 50% (0.5x)
Video Processing: Enabled (H.265, CRF 23)
Threshold: 1920x1080 (Smart limit enabled)

Processing Files:
✓ IMG_001.jpg (4032x3024 → 2016x1512) - 2.1MB → 0.8MB
✓ IMG_002.HEIC (4032x3024 → 2016x1512) - 3.2MB → 0.9MB (converted to JPEG)
✓ video_001.mp4 (1920x1080 → 1920x1080) - 45.2MB → 12.3MB
✓ IMG_003.png (1920x1080) - Skipped (below threshold)
✓ video_002.mov (4K → 1080p) - 120.5MB → 28.7MB

Processing Summary:
==================
Images processed: 2/3
Videos processed: 2/2
Total size reduction: 171.0MB → 42.7MB (75.0% reduction)
Processing time: 45.67 seconds
HTML Report: ./output/processing_report.html
```

### HTML Report Features
The generated HTML report includes:
- **Interactive Grid Layout**: Visual card-based file display
- **Thumbnail Previews**: Image thumbnails and video frame previews
- **Clickable File Links**: Direct access to processed files
- **Detailed Statistics**: File sizes, dimensions, processing time
- **Responsive Design**: Works on desktop and mobile devices
- **Processing Summary**: Overall statistics and performance metrics

## License

This project is licensed under the MIT License.