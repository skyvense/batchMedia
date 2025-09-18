# BatchMedia 测试使用指南

## 快速开始

### 1. 生成测试图片
```bash
cd test
go run create_test_images.go
```

### 2. 运行完整测试
```bash
./test_script.sh
```

## 测试结果

测试脚本已成功验证以下功能：

✅ **图片缩放功能**
- 按比例缩放 (`-size 0.5`)
- 按宽度缩放 (`-width 1920`)

✅ **分辨率过滤**
- 阈值设置 (`-threshold-width 1920 -threshold-height 1080`)
- 智能限制忽略 (`-ignore-smart-limit`)

✅ **文件处理**
- 文件类型过滤 (`-ext jpg`)
- 多线程处理 (`-multithread 4`)
- 假扫描模式 (`-fake-scan`)

✅ **视频处理**
- 视频缩放和压缩
- FFmpeg 集成
- HTML 报告生成

## 测试文件说明

### 输入文件
- **大图片**: 4K (3840x2160), 6K (6000x4000), 8K (7680x4320)
- **中等图片**: FHD (1920x1080), 2K (2560x1440), 宽屏 (2048x1152)
- **小图片**: HD (1280x720), VGA (640x480), 缩略图 (320x240)
- **测试视频**: 5秒 1920x1080 测试视频

### 输出结果
- 处理后的图片和视频文件
- HTML 处理报告
- JSON 进度文件

## 手动测试示例

### 基本图片处理
```bash
../bin/batchMedia -inputdir input/images -out output/images -size 0.5
```

### 高质量缩放
```bash
../bin/batchMedia -inputdir input/images -out output/images -width 1920
```

### 只处理大图片
```bash
../bin/batchMedia -inputdir input/images -out output/images -size 0.7 -threshold-width 1920
```

### 预览模式
```bash
../bin/batchMedia -inputdir input/images -out output/images -size 0.5 -fake-scan
```

### 视频处理
```bash
../bin/batchMedia -inputdir input/videos -out output/videos -size 0.5
```

## 验证结果

1. **检查输出文件**：`ls -la output/*/`
2. **查看HTML报告**：打开 `output/videos/processing_report.html`
3. **检查文件大小**：确认压缩效果
4. **验证分辨率**：使用 `file` 命令检查图片尺寸

## 注意事项

- 确保已构建 batchMedia：`go build -o bin/batchMedia`
- 视频测试需要 FFmpeg
- 测试会自动清理输出目录
- 可以修改测试脚本参数进行自定义测试