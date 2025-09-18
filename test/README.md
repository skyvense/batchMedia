# BatchMedia 测试目录

这个目录包含了用于测试 BatchMedia 工具各种功能的测试文件和脚本。

## 目录结构

```
test/
├── input/                    # 输入测试文件
│   ├── images/              # 测试图片文件
│   ├── videos/              # 测试视频文件
│   └── mixed/               # 混合文件类型
├── output/                  # 输出目录
│   ├── images/              # 处理后的图片
│   ├── videos/              # 处理后的视频
│   └── mixed/               # 混合处理结果
├── samples/                 # 示例文件
├── create_test_images.go    # 测试图片生成脚本
├── test_script.sh          # 综合测试脚本
└── README.md               # 本文件
```

## 测试图片说明

生成的测试图片包含不同尺寸和格式：

### 大图片（>= 1920x1080）
- `large_4k.jpg` (3840x2160) - 4K分辨率
- `large_6k.png` (6000x4000) - 6K分辨率  
- `large_8k.jpg` (7680x4320) - 8K分辨率

### 中等图片（接近1920x1080）
- `medium_fhd.jpg` (1920x1080) - 全高清
- `medium_2k.png` (2560x1440) - 2K分辨率
- `medium_wide.jpg` (2048x1152) - 宽屏格式

### 小图片（< 1920x1080）
- `small_hd.jpg` (1280x720) - 高清
- `small_vga.png` (640x480) - VGA
- `small_thumb.jpg` (320x240) - 缩略图

## 使用方法

### 1. 生成测试图片

```bash
cd test
go run create_test_images.go
```

### 2. 运行综合测试

```bash
cd test
./test_script.sh
```

### 3. 手动测试特定功能

#### 基本图片缩放
```bash
../bin/batchMedia --input input/images --output output/images --scale 0.5
```

#### 按宽度缩放
```bash
../bin/batchMedia --input input/images --output output/images --width 1920
```

#### 分辨率过滤
```bash
../bin/batchMedia --input input/images --output output/images --scale 0.7 --min-width 1920 --min-height 1080
```

#### 生成HTML报告
```bash
../bin/batchMedia --input input/images --output output/images --scale 0.5 --html-report output/report.html
```

#### 假扫描模式（预览）
```bash
../bin/batchMedia --input input/images --output output/images --scale 0.5 --fake-scan
```

## 测试脚本功能

`test_script.sh` 包含以下测试：

1. **基本图片处理** - 按比例缩放
2. **按宽度缩放** - 固定宽度缩放
3. **分辨率过滤** - 只处理大图片
4. **质量设置** - JPEG压缩质量
5. **递归处理** - 文件类型过滤
6. **文件排除** - 排除特定文件
7. **混合目录处理** - 处理多种文件类型
8. **HTML报告生成** - 生成处理报告
9. **假扫描模式** - 预览功能
10. **多线程处理** - 并发处理
11. **视频处理** - 如果FFmpeg可用

## 注意事项

- 确保已构建 batchMedia 工具：`go build -o bin/batchMedia`
- 视频测试需要安装 FFmpeg
- 测试脚本会自动清理输出目录
- 可以单独运行特定测试或修改脚本参数

## 验证结果

测试完成后，检查：
- `output/` 目录中的处理结果
- `output/report.html` HTML报告（如果生成）
- 控制台输出的处理统计信息