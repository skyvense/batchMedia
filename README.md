# batchMedia - 批量媒体处理工具

一个用 Go 语言编写的强大命令行工具，用于批量处理图片和视频。支持多种格式、智能缩放、HDR 视频处理，并生成精美的 HTML 报告。

## 功能特性

### 图片处理
- 🖼️ **多格式支持**: 支持 JPEG/JPG、PNG 和 HEIC 图片格式
- 📏 **灵活缩放**: 比例缩放和基于宽度的缩放
- 🔄 **HEIC 转换**: 自动将 HEIC 转换为 JPEG，保留 EXIF 信息
- 📊 **EXIF 元数据**: 在格式转换时保留和传输 EXIF 数据

### 视频处理
- 🎬 **视频支持**: 支持 MOV、MP4、AVI、MKV 格式
- 🌈 **HDR 保留**: 保持 HDR 元数据以获得高质量视频输出
- 🎯 **智能编码**: H.264/H.265 编码，兼容 QuickTime
- ⚙️ **灵活参数**: 可自定义 CRF、码率和分辨率设置

### 高级功能
- 📅 **元数据保留**: 保持原始文件修改日期
- 📁 **目录结构**: 递归处理并保留文件夹层次结构
- 📋 **HTML 报告**: 生成精美的交互式报告，包含缩略图和文件链接
- 🚀 **高性能**: 优化的批处理，依赖最少
- 🎛️ **智能过滤**: 基于分辨率的智能处理决策

## 安装

### 前置要求
- **Go 1.21 或更高版本** - 构建应用程序所需
- **Git** - 用于克隆代码仓库
- **FFmpeg** - 视频处理所需（仅图片处理可选）

#### 安装 FFmpeg

**macOS (使用 Homebrew):**
```bash
brew install ffmpeg
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install ffmpeg
```

**Windows:**
- 从 [FFmpeg 官方网站](https://ffmpeg.org/download.html) 下载
- 将 FFmpeg 添加到系统 PATH

### 从源码构建

1. 克隆代码仓库:
```bash
git clone <repository-url>
cd batchMedia
```

2. 安装 Go 依赖:
```bash
go mod tidy
```

3. 构建可执行文件:
```bash
# 使用构建脚本
./build.sh

# 或直接构建
go build -o batchMedia
```

4. 验证安装:
```bash
./batchMedia -h
```

## 使用方法

### 基本语法

```bash
./batchMedia --inputdir=<输入目录> --out=<输出目录> [选项]
```

**注意**: Go的flag包支持单横线和双横线两种格式，例如 `-inputdir` 和 `--inputdir` 都可以使用。

### 图片处理选项

- `--size=<比例>`: 按比例缩放（例如，0.5 表示缩放到 50%）
- `--width=<像素>`: 按指定宽度缩放，自动保持宽高比

**注意：`--size` 和 `--width` 参数不能同时使用**

### 视频处理选项

- `--disable-video`: 禁用视频处理（默认启用视频处理）
- `--video-codec=<编码器>`: 视频编码器（libx264, libx265）- 默认：libx265
- `--video-bitrate=<码率>`: 视频码率（例如：2M, 1000k）
- `--video-resolution=<分辨率>`: 视频分辨率（例如：1920x1080, 1280x720）
- `--video-crf=<值>`: 视频 CRF 质量（0-51，数值越低质量越好）- 默认：23
- `--video-preset=<预设>`: 编码预设（ultrafast, fast, medium, slow, veryslow）- 默认：medium

### 分辨率过滤选项

- `--threshold-width=<像素>`: 宽度过滤阈值（默认：缩小时为 1920，放大时为 3840）
- `--threshold-height=<像素>`: 高度过滤阈值（默认：缩小时为 1080，放大时为 2160）
- `--ignore-smart-limit`: 忽略智能默认分辨率限制

**智能阈值逻辑:**
- **缩小处理**（缩放比例 < 1.0）：跳过**低于**阈值的图片（太小无法有效缩小）
- **放大处理**（缩放比例 > 1.0）：跳过**高于**阈值的图片（太大无法有效放大）
- 缩放比例由 `--size` 参数确定或从 `--width` 参数计算得出
- 超出指定分辨率范围的图片将直接复制到输出目录而不进行缩放

### 使用示例

#### 图片处理示例

##### 1. 按比例缩放图片
将图片缩小到原始尺寸的 50%：
```bash
./batchMedia --inputdir=./photos/2019 --out=./photos/2019_resized --size=0.5
```

##### 2. 按宽度缩放图片
将图片宽度调整为 1920 像素，高度按比例自动调整：
```bash
./batchMedia --inputdir=./photos/2019 --out=./photos/2019_1920 --width=1920
```

##### 3. 处理 HEIC 文件
将 HEIC 文件转换为 JPEG 并保留 EXIF 数据：
```bash
./batchMedia --inputdir=./iphone_photos --out=./converted_photos --size=1.0
```

#### 视频处理示例

##### 4. 基本视频处理
使用默认 H.265 编码处理视频：
```bash
./batchMedia --inputdir=./videos --out=./compressed_videos
```

##### 5. 高质量视频编码
使用自定义质量设置处理视频：
```bash
./batchMedia --inputdir=./videos --out=./hq_videos --video-crf=18 --video-preset=slow
```

##### 6. 视频分辨率缩放
将视频缩放到 1080p 分辨率：
```bash
./batchMedia --inputdir=./4k_videos --out=./1080p_videos --video-resolution=1920x1080
```

#### 混合媒体处理示例

##### 7. 同时处理图片和视频
在同一目录中处理图片和视频：
```bash
./batchMedia --inputdir=./mixed_media --out=./processed_media --size=0.8 --video-crf=20
```

##### 8. 高级过滤
缩小图片但跳过小图片，使用自定义视频设置：
```bash
./batchMedia --inputdir=./media --out=./filtered_media --size=0.5 --threshold-width=1000 --threshold-height=1000 --video-codec=libx264
```

#### 3. 创建测试图片
不带任何参数运行程序将自动创建测试图片：
```bash
./batchMedia
```

### 参数说明

| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| **核心参数（按使用频率排序）** |
| `--inputdir` | string | 是 | 输入目录路径，包含要处理的媒体文件 |
| `--out` | string | 是 | 输出目录路径，处理后的文件将保存在此 |
| `--size` | float | 否 | 缩放比例，范围 0-10（与 --width 互斥） |
| `--multithread` | int | 否 | 并发线程数，用于处理多个目录（默认：1） |
| **图片处理参数** |
| `--width` | int | 否 | 目标宽度（像素）（与 --size 互斥） |
| `--threshold-width` | int | 否 | 宽度阈值（默认：缩小时为 1920，放大时为 3840） |
| `--threshold-height` | int | 否 | 高度阈值（默认：缩小时为 1080，放大时为 2160） |
| `--ignore-smart-limit` | bool | 否 | 忽略智能默认分辨率限制 |
| **文件过滤参数** |
| `--ext` | string | 否 | 仅处理指定扩展名的文件（逗号分隔，如：heic,jpg,png） |
| `--fake-scan` | bool | 否 | 仅扫描和列出待处理文件，不实际处理 |
| **视频处理参数** |
| `--disable-video` | bool | 否 | 禁用视频处理（默认启用视频处理） |
| `--video-codec` | string | 否 | 视频编码器：libx264, libx265（默认：libx265） |
| `--video-bitrate` | string | 否 | 视频码率（例如：2M, 1000k） |
| `--video-resolution` | string | 否 | 视频分辨率（例如：1920x1080, 1280x720） |
| `--video-crf` | int | 否 | 视频 CRF 质量，0-51，数值越低质量越好（默认：23） |
| `--video-preset` | string | 否 | 编码预设：ultrafast, fast, medium, slow, veryslow（默认：medium） |
| **其他** |
| `--help` | - | 否 | 显示帮助信息 |

## 工作原理

1. **文件发现**: 递归扫描输入目录，查找所有 `.jpg`、`.jpeg`、`.png` 和 `.heic` 文件
2. **智能过滤**: 
   - 根据缩放比例确定操作类型（> 1.0 = 放大，< 1.0 = 缩小）
   - 应用阈值过滤跳过不合适的图片
   - 使用智能默认值：缩小时为 1920x1080，放大时为 3840x2160
3. **图片处理**:
   - 解码 JPEG、PNG 和 HEIC 图片
   - 根据指定参数计算新尺寸
   - 使用 Lanczos3 算法进行高质量图片缩放
   - 重新编码为 JPEG 格式（90% 质量）
4. **文件保存**:
   - 保持原始目录结构
   - 保留原始文件修改时间
   - 自动创建必要的输出目录

## 技术特性

### 核心处理
- **多格式支持**: 支持 JPEG、PNG、HEIC 图片和各种视频格式（MP4、MOV、AVI、MKV 等）
- **EXIF 数据保留**: 在图片处理过程中自动保留 EXIF 元数据
- **HEIC 到 JPEG 转换**: 无缝将 HEIC 文件转换为 JPEG 格式
- **智能分辨率过滤**: 智能阈值系统避免不必要的处理
- **批量处理**: 高效并行处理多个文件
- **内存优化**: 针对大文件批次优化内存使用

### 视频处理
- **现代编码器**: 支持 H.264 和 H.265 (HEVC) 编码
- **质量控制**: 基于 CRF 的质量设置，实现最佳压缩
- **分辨率缩放**: 灵活的视频分辨率调整
- **预设选项**: 多种编码速度/质量预设
- **FFmpeg 集成**: 利用 FFmpeg 进行强大的视频处理

### 报告与分析
- **HTML 报告**: 生成精美的交互式 HTML 报告
- **网格布局**: 现代卡片式布局，包含缩略图
- **文件统计**: 详细的处理统计和文件信息
- **可点击链接**: 从报告直接访问文件
- **可视化缩略图**: 预览图片和视频帧

### 系统兼容性
- **跨平台**: 支持 macOS、Linux 和 Windows
- **Go 原生**: 纯 Go 实现，无外部依赖（除 FFmpeg 用于视频处理）
- **高性能**: 优化的并发处理和内存管理

## 重要注意事项

1. **参数互斥**: `--size` 和 `--width` 参数不能同时使用
2. **FFmpeg 依赖**: 视频处理需要安装 FFmpeg
3. **EXIF 元数据**: 为 JPEG 和 HEIC 文件保留 EXIF 数据
4. **内存使用**: 大图片会消耗更多内存
5. **文件覆盖**: 现有输出文件将被覆盖
6. **目录结构**: 保持输入目录的相对路径结构
7. **HEIF 支持**: 现已完全集成 HEIF/HEIC 支持，无需 noheif 标签

## 错误处理

程序在以下情况下会报告错误并退出：
- 输入目录不存在
- 缺少必需参数
- 同时指定 size 和 width 参数
- 参数值超出有效范围
- 文件读写权限不足
- 并发处理错误（已修复）

## 示例输出

### 控制台输出
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

### HTML 报告功能
生成的 HTML 报告包括：
- **交互式网格布局**: 基于卡片的可视化文件显示
- **缩略图预览**: 图片缩略图和视频帧预览
- **可点击文件链接**: 直接访问处理后的文件
- **详细统计**: 文件大小、尺寸、处理时间
- **响应式设计**: 在桌面和移动设备上都能正常工作
- **处理摘要**: 整体统计和性能指标

## 许可证

本项目采用 MIT 许可证。

---

# batchMedia - Batch Media Processing Tool

A powerful command-line tool written in Go for batch processing images and videos. Supports multiple formats, intelligent scaling, HDR video processing, and generates beautiful HTML reports.

## Features

### Image Processing
- 🖼️ **Multi-format Support**: Supports JPEG/JPG, PNG, and HEIC image formats
- 📏 **Flexible Scaling**: Ratio-based and width-based scaling
- 🔄 **HEIC Conversion**: Automatically converts HEIC to JPEG while preserving EXIF information
- 📊 **EXIF Metadata**: Preserves and transfers EXIF data during format conversion

### Video Processing
- 🎬 **Video Support**: Supports MOV, MP4, AVI, MKV formats
- 🌈 **HDR Preservation**: Maintains HDR metadata for high-quality video output
- 🎯 **Smart Encoding**: H.264/H.265 encoding with QuickTime compatibility
- ⚙️ **Flexible Parameters**: Customizable CRF, bitrate, and resolution settings

### Advanced Features
- 📅 **Metadata Preservation**: Maintains original file modification dates
- 📁 **Directory Structure**: Recursive processing while preserving folder hierarchy
- 📋 **HTML Reports**: Generates beautiful interactive reports with thumbnails and file links
- 🚀 **High Performance**: Optimized batch processing with minimal dependencies
- 🎛️ **Smart Filtering**: Intelligent processing decisions based on resolution

## Installation

### Prerequisites
- **Go 1.21 or higher** - Required for building the application
- **Git** - For cloning the repository
- **FFmpeg** - Required for video processing (optional for image-only processing)

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
# Using build script
./build.sh

# Or build directly
go build -o batchMedia
```

4. Verify installation:
```bash
./batchMedia -h
```

## Usage

### Basic Syntax

```bash
./batchMedia --inputdir=<input_directory> --out=<output_directory> [options]
```

**Note**: Go's flag package supports both single and double dash formats, e.g., both `-inputdir` and `--inputdir` work.

### Image Processing Options

- `--size=<ratio>`: Scale by ratio (e.g., 0.5 means scale to 50%)
- `--width=<pixels>`: Scale by specified width, automatically maintains aspect ratio

**Note: `--size` and `--width` parameters cannot be used simultaneously**

### Video Processing Options

- `--disable-video`: Disable video processing (video processing is enabled by default)
- `--video-codec=<codec>`: Video codec (libx264, libx265) - Default: libx265
- `--video-bitrate=<bitrate>`: Video bitrate (e.g., 2M, 1000k)
- `--video-resolution=<resolution>`: Video resolution (e.g., 1920x1080, 1280x720)
- `--video-crf=<value>`: Video CRF quality (0-51, lower is better) - Default: 23
- `--video-preset=<preset>`: Encoding preset (ultrafast, fast, medium, slow, veryslow) - Default: medium

### Resolution Filtering Options

- `--threshold-width=<pixels>`: Width filtering threshold (Default: 1920 for downscaling, 3840 for upscaling)
- `--threshold-height=<pixels>`: Height filtering threshold (Default: 1080 for downscaling, 2160 for upscaling)
- `--ignore-smart-limit`: Ignore smart default resolution limits

**Smart Threshold Logic:**
- **Downscaling** (scale ratio < 1.0): Skip images **below** threshold (too small to effectively downscale)
- **Upscaling** (scale ratio > 1.0): Skip images **above** threshold (too large to effectively upscale)
- Scale ratio is determined by `--size` parameter or calculated from `--width` parameter
- Images outside the specified resolution range will be copied directly to output directory without scaling

### Usage Examples

#### Image Processing Examples

##### 1. Scale Images by Ratio
Scale images down to 50% of original size:
```bash
./batchMedia --inputdir=./photos/2019 --out=./photos/2019_resized --size=0.5
```

##### 2. Scale Images by Width
Resize image width to 1920 pixels, height automatically adjusted proportionally:
```bash
./batchMedia --inputdir=./photos/2019 --out=./photos/2019_1920 --width=1920
```

##### 3. Process HEIC Files
Convert HEIC files to JPEG while preserving EXIF data:
```bash
./batchMedia --inputdir=./iphone_photos --out=./converted_photos --size=1.0
```

#### Video Processing Examples

##### 4. Basic Video Processing
Process videos with default H.265 encoding:
```bash
./batchMedia --inputdir=./videos --out=./compressed_videos
```

##### 5. High-Quality Video Encoding
Process videos with custom quality settings:
```bash
./batchMedia --inputdir=./videos --out=./hq_videos --video-crf=18 --video-preset=slow
```

##### 6. Video Resolution Scaling
Scale videos to 1080p resolution:
```bash
./batchMedia --inputdir=./4k_videos --out=./1080p_videos --video-resolution=1920x1080
```

#### Mixed Media Processing Examples

##### 7. Process Images and Videos Together
Process both images and videos in the same directory:
```bash
./batchMedia --inputdir=./mixed_media --out=./processed_media --size=0.8 --video-crf=20
```

##### 8. Advanced Filtering
Downscale images but skip small ones, with custom video settings:
```bash
./batchMedia --inputdir=./media --out=./filtered_media --size=0.5 --threshold-width=1000 --threshold-height=1000 --video-codec=libx264
```

#### 3. Create Test Images
Running the program without any parameters will automatically create test images:
```bash
./batchMedia
```

### Parameter Description

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| **Core Parameters (Ordered by Usage Frequency)** |
| `--inputdir` | string | Yes | Input directory path containing media files to process |
| `--out` | string | Yes | Output directory path where processed files will be saved |
| `--size` | float | No | Scaling ratio, range 0-10 (mutually exclusive with --width) |
| `--multithread` | int | No | Number of concurrent threads for processing multiple directories (default: 1) |
| **Image Processing Parameters** |
| `--width` | int | No | Target width in pixels (mutually exclusive with --size) |
| `--threshold-width` | int | No | Width threshold (default: 1920 for downscaling, 3840 for upscaling) |
| `--threshold-height` | int | No | Height threshold (default: 1080 for downscaling, 2160 for upscaling) |
| `--ignore-smart-limit` | bool | No | Ignore smart default resolution limits |
| **File Filtering Parameters** |
| `--ext` | string | No | Process only files with specified extensions (comma-separated, e.g., heic,jpg,png) |
| `--fake-scan` | bool | No | Only scan and list files to be processed, don't actually process them |
| **Video Processing Parameters** |
| `--disable-video` | bool | No | Disable video processing (video processing is enabled by default) |
| `--video-codec` | string | No | Video codec: libx264, libx265 (default: libx265) |
| `--video-bitrate` | string | No | Video bitrate (e.g., 2M, 1000k) |
| `--video-resolution` | string | No | Video resolution (e.g., 1920x1080, 1280x720) |
| `--video-crf` | int | No | Video CRF quality, 0-51, lower is better (default: 23) |
| `--video-preset` | string | No | Encoding preset: ultrafast, fast, medium, slow, veryslow (default: medium) |
| **Other** |
| `--help` | - | No | Display help information |

## How It Works

1. **File Discovery**: Recursively scans input directory for all `.jpg`, `.jpeg`, `.png`, and `.heic` files
2. **Smart Filtering**: 
   - Determines operation type based on scale ratio (> 1.0 = upscaling, < 1.0 = downscaling)
   - Applies threshold filtering to skip inappropriate images
   - Uses smart defaults: 1920x1080 for downscaling, 3840x2160 for upscaling
3. **Image Processing**:
   - Decodes JPEG, PNG, and HEIC images
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
- **Batch Processing**: Efficient parallel processing of multiple files
- **Memory Optimization**: Optimized memory usage for large file batches

### Video Processing
- **Modern Codecs**: Supports H.264 and H.265 (HEVC) encoding
- **Quality Control**: CRF-based quality settings for optimal compression
- **Resolution Scaling**: Flexible video resolution adjustment
- **Preset Options**: Multiple encoding speed/quality presets
- **FFmpeg Integration**: Leverages FFmpeg for powerful video processing

### Reporting & Analysis
- **HTML Reports**: Generates beautiful interactive HTML reports
- **Grid Layout**: Modern card-based layout with thumbnails
- **File Statistics**: Detailed processing statistics and file information
- **Clickable Links**: Direct file access from reports
- **Visual Thumbnails**: Preview images and video frames

### System Compatibility
- **Cross-platform**: Supports macOS, Linux, and Windows
- **Go Native**: Pure Go implementation with no external dependencies (except FFmpeg for video processing)
- **High Performance**: Optimized concurrent processing and memory management

## Important Notes

1. **Parameter Exclusivity**: `--size` and `--width` parameters cannot be used simultaneously
2. **FFmpeg Dependency**: Video processing requires FFmpeg installation
3. **EXIF Metadata**: Preserves EXIF data for JPEG and HEIC files
4. **Memory Usage**: Large images will consume more memory
5. **File Overwriting**: Existing output files will be overwritten
6. **Directory Structure**: Maintains relative path structure of input directory
7. **HEIF Support**: Full HEIF/HEIC support is now integrated, no noheif tag needed

## Error Handling

The program will report errors and exit in the following situations:
- Input directory does not exist
- Missing required parameters
- Both size and width parameters specified simultaneously
- Parameter values outside valid ranges
- Insufficient file read/write permissions
- Concurrent processing errors (fixed)

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