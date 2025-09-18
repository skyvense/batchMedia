# batchMedia - 批量媒体处理工具

一个用 Go 语言编写的强大命令行工具，用于批量处理图片和视频。支持多种格式、智能缩放、HDR 视频处理，并生成精美的 HTML 报告。

A powerful command-line tool written in Go for batch processing images and videos. Supports multiple formats, intelligent scaling, HDR video processing, and generates beautiful HTML reports.

## 功能特性 Features

### 图片处理 Image Processing
- 🖼️ **多格式支持 Multi-format Support**: 支持 JPEG/JPG、PNG 和 HEIC 图片格式
- 📏 **灵活缩放 Flexible Scaling**: 比例缩放和基于宽度的缩放
- 🔄 **HEIC 转换 HEIC Conversion**: 自动将 HEIC 转换为 JPEG，保留 EXIF 信息
- 📊 **EXIF 元数据 EXIF Metadata**: 在格式转换时保留和传输 EXIF 数据

### 视频处理 Video Processing
- 🎬 **视频支持 Video Support**: 支持 MOV、MP4、AVI、MKV 格式
- 🌈 **HDR 保留 HDR Preservation**: 保持 HDR 元数据以获得高质量视频输出
- 🎯 **智能编码 Smart Encoding**: H.264/H.265 编码，兼容 QuickTime
- ⚙️ **灵活参数 Flexible Parameters**: 可自定义 CRF、码率和分辨率设置

### 高级功能 Advanced Features
- 📅 **元数据保留 Metadata Preservation**: 保持原始文件修改日期
- 📁 **目录结构 Directory Structure**: 递归处理并保留文件夹层次结构
- 📋 **HTML 报告 HTML Reports**: 生成精美的交互式报告，包含缩略图和文件链接
- 🚀 **高性能 High Performance**: 优化的批处理，依赖最少
- 🎛️ **智能过滤 Smart Filtering**: 基于分辨率的智能处理决策

## 安装 Installation

### 前置要求 Prerequisites
- **Go 1.21 或更高版本** - 构建应用程序所需
- **Git** - 用于克隆代码仓库
- **FFmpeg** - 视频处理所需（仅图片处理可选）

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

### 从源码构建 Build from Source

1. 克隆代码仓库 Clone the repository:
```bash
git clone <repository-url>
cd batchMedia
```

2. 安装 Go 依赖 Install Go dependencies:
```bash
go mod tidy
```

3. 构建可执行文件 Build the executable:
```bash
# 使用构建脚本 Using build script
./build.sh

# 或直接构建 Or build directly
go build -o batchMedia
```

4. 验证安装 Verify installation:
```bash
./batchMedia -h
```

## 使用方法 Usage

### 基本语法 Basic Syntax

```bash
./batchMedia --inputdir=<输入目录> --out=<输出目录> [选项]
./batchMedia --inputdir=<input_directory> --out=<output_directory> [options]
```

**注意 Note**: Go的flag包支持单横线和双横线两种格式，例如 `-inputdir` 和 `--inputdir` 都可以使用。
**Note**: Go's flag package supports both single and double dash formats, e.g., both `-inputdir` and `--inputdir` work.

### 图片处理选项 Image Processing Options

- `--size=<比例>`: 按比例缩放（例如，0.5 表示缩放到 50%）
- `--width=<像素>`: 按指定宽度缩放，自动保持宽高比

**注意：`--size` 和 `--width` 参数不能同时使用**

### 视频处理选项 Video Processing Options

- `--disable-video`: 禁用视频处理（默认启用视频处理）
- `--video-codec=<编码器>`: 视频编码器（libx264, libx265）- 默认：libx265
- `--video-bitrate=<码率>`: 视频码率（例如：2M, 1000k）
- `--video-resolution=<分辨率>`: 视频分辨率（例如：1920x1080, 1280x720）
- `--video-crf=<值>`: 视频 CRF 质量（0-51，数值越低质量越好）- 默认：23
- `--video-preset=<预设>`: 编码预设（ultrafast, fast, medium, slow, veryslow）- 默认：medium

### 分辨率过滤选项 Resolution Filtering Options

- `--threshold-width=<像素>`: 宽度过滤阈值（默认：缩小时为 1920，放大时为 3840）
- `--threshold-height=<像素>`: 高度过滤阈值（默认：缩小时为 1080，放大时为 2160）
- `--ignore-smart-limit`: 忽略智能默认分辨率限制

**智能阈值逻辑 Smart Threshold Logic:**
- **缩小处理**（缩放比例 < 1.0）：跳过**低于**阈值的图片（太小无法有效缩小）
- **放大处理**（缩放比例 > 1.0）：跳过**高于**阈值的图片（太大无法有效放大）
- 缩放比例由 `--size` 参数确定或从 `--width` 参数计算得出
- 超出指定分辨率范围的图片将直接复制到输出目录而不进行缩放

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


#### 3. 创建测试图片 Create Test Images
不带任何参数运行程序将自动创建测试图片：
```bash
./batchMedia
```

### 参数说明 Parameter Description

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| **核心参数 Core Parameters (按使用频率排序 Ordered by Usage Frequency)** |
| `--inputdir` | string | Yes | 输入目录路径 Input directory path containing media files to process |
| `--out` | string | Yes | 输出目录路径 Output directory path where processed files will be saved |
| `--size` | float | No | 缩放比例，范围 0-10 Scaling ratio, range 0-10 (与 --width 互斥 mutually exclusive with --width) |
| `--multithread` | int | No | 并发线程数 Number of concurrent threads for processing multiple directories (默认 default: 1) |
| **图片处理参数 Image Processing Parameters** |
| `--width` | int | No | 目标宽度（像素）Target width in pixels (与 --size 互斥 mutually exclusive with --size) |
| `--threshold-width` | int | No | 宽度阈值 Width threshold (默认 default: 1920 for downscaling, 3840 for upscaling) |
| `--threshold-height` | int | No | 高度阈值 Height threshold (默认 default: 1080 for downscaling, 2160 for upscaling) |
| `--ignore-smart-limit` | bool | No | 忽略智能默认分辨率限制 Ignore smart default resolution limits |
| **文件过滤参数 File Filtering Parameters** |
| `--ext` | string | No | 仅处理指定扩展名的文件 Process only files with specified extensions (逗号分隔 comma-separated, e.g., heic,jpg,png) |
| `--fake-scan` | bool | No | 仅扫描和列出待处理文件，不实际处理 Only scan and list files to be processed, don't actually process them |
| **视频处理参数 Video Processing Parameters** |
| `--disable-video` | bool | No | 禁用视频处理 Disable video processing (默认启用视频处理 video processing is enabled by default) |
| `--video-codec` | string | No | 视频编码器 Video codec: libx264, libx265 (默认 default: libx265) |
| `--video-bitrate` | string | No | 视频码率 Video bitrate (例如 e.g., 2M, 1000k) |
| `--video-resolution` | string | No | 视频分辨率 Video resolution (例如 e.g., 1920x1080, 1280x720) |
| `--video-crf` | int | No | 视频 CRF 质量 Video CRF quality, 0-51, 数值越低质量越好 lower is better (默认 default: 23) |
| `--video-preset` | string | No | 编码预设 Encoding preset: ultrafast, fast, medium, slow, veryslow (默认 default: medium) |
| **其他 Other** |
| `--help` | - | No | 显示帮助信息 Display help information |

## 工作原理 How It Works

1. **文件发现 File Discovery**: 递归扫描输入目录，查找所有 `.jpg`、`.jpeg`、`.png` 和 `.heic` 文件
2. **智能过滤 Smart Filtering**: 
   - 根据缩放比例确定操作类型（> 1.0 = 放大，< 1.0 = 缩小）
   - 应用阈值过滤跳过不合适的图片
   - 使用智能默认值：缩小时为 1920x1080，放大时为 3840x2160
3. **图片处理 Image Processing**:
   - 解码 JPEG、PNG 和 HEIC 图片
   - 根据指定参数计算新尺寸
   - 使用 Lanczos3 算法进行高质量图片缩放
   - 重新编码为 JPEG 格式（90% 质量）
4. **文件保存 File Saving**:
   - 保持原始目录结构
   - 保留原始文件修改时间
   - 自动创建必要的输出目录

## 技术特性 Technical Features

### 核心处理 Core Processing
- **多格式支持 Multi-format Support**: 支持 JPEG、PNG、HEIC 图片和各种视频格式（MP4、MOV、AVI、MKV 等）
- **EXIF 数据保留 EXIF Data Preservation**: 在图片处理过程中自动保留 EXIF 元数据
- **HEIC 到 JPEG 转换**: 无缝将 HEIC 文件转换为 JPEG 格式
- **智能分辨率过滤 Smart Resolution Filtering**: 智能阈值系统避免不必要的处理
- **批量处理 Batch Processing**: 高效并行处理多个文件
- **内存优化 Memory Optimization**: 针对大文件批次优化内存使用

### 视频处理 Video Processing
- **现代编码器 Modern Codecs**: 支持 H.264 和 H.265 (HEVC) 编码
- **质量控制 Quality Control**: 基于 CRF 的质量设置，实现最佳压缩
- **分辨率缩放 Resolution Scaling**: 灵活的视频分辨率调整
- **预设选项 Preset Options**: 多种编码速度/质量预设
- **FFmpeg 集成**: 利用 FFmpeg 进行强大的视频处理

### 报告与分析 Reporting & Analysis
- **HTML 报告 HTML Reports**: 生成精美的交互式 HTML 报告
- **网格布局 Grid Layout**: 现代卡片式布局，包含缩略图
- **文件统计 File Statistics**: 详细的处理统计和文件信息
- **可点击链接 Clickable Links**: 从报告直接访问文件
- **可视化缩略图 Visual Thumbnails**: 预览图片和视频帧

### 系统兼容性 System Compatibility
- **编程语言 Programming Language**: Go 1.21+
- **图片处理 Image Processing**: 使用 Go 标准库 `image` 和 `image/jpeg` 包，加上 `jdeng/goheif` 提供 HEIC 支持，`nfnt/resize` 提供高质量缩放
- **智能逻辑 Smart Logic**: 基于缩放比例比较的简化阈值过滤
- **命令行解析 Command Line Parsing**: 使用 Go 标准库 `flag` 包
- **文件操作 File Operations**: 跨平台文件系统操作支持
- **算法 Algorithm**: Lanczos3 插值算法用于图片缩放
- **EXIF 支持**: 使用 `goexif` 和 `jdeng/goheif` 库保留 JPEG 和 HEIC 文件的 EXIF 元数据

## 性能说明 Performance Notes

- 使用内存图片处理，适合中等大小图片的批量处理
- 对于非常大的图片文件，建议分批处理以避免内存不足
- 处理速度取决于图片大小、数量和系统性能
- 现已修复并发安全问题，支持多线程处理

## 重要说明 Important Notes

1. **支持格式 Supported Formats**: 支持 JPEG/JPG、PNG 和 HEIC 格式
2. **输出质量 Output Quality**: 输出 JPEG 质量固定为 90%
3. **EXIF 元数据**: 为 JPEG 和 HEIC 文件保留 EXIF 数据
4. **内存使用 Memory Usage**: 大图片会消耗更多内存
5. **文件覆盖 File Overwriting**: 现有输出文件将被覆盖
6. **目录结构 Directory Structure**: 保持输入目录的相对路径结构
7. **HEIF 支持**: 现已完全集成 HEIF/HEIC 支持，无需 noheif 标签

## 错误处理 Error Handling

程序在以下情况下会报告错误并退出：
- 输入目录不存在
- 缺少必需参数
- 同时指定 size 和 width 参数
- 参数值超出有效范围
- 文件读写权限不足
- 并发处理错误（已修复）

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