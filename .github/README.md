# GitHub Actions 自动编译说明

## 功能特性

本项目配置了完整的 GitHub Actions 自动化工作流，支持：

### 🔄 自动化测试
- 每次推送到 `main` 或 `master` 分支时自动运行
- 每次创建 Pull Request 时自动运行
- 运行完整的测试套件，包括图片处理和视频处理测试

### 🏗️ 多平台编译
支持以下平台的自动编译：
- **Linux**: amd64, arm64
- **macOS**: amd64, arm64 (Intel 和 Apple Silicon)
- **Windows**: amd64

### 📦 自动发布
- 当推送带有 `v*` 标签时（如 `v1.0.0`），自动创建 GitHub Release
- 自动打包所有平台的二进制文件
- Linux/macOS 使用 `.tar.gz` 格式
- Windows 使用 `.zip` 格式

## 使用方法

### 触发自动编译
```bash
# 推送代码到主分支
git push origin main

# 创建并推送标签来发布新版本
git tag v1.0.0
git push origin v1.0.0
```

### 下载编译好的二进制文件
1. 访问项目的 [Releases 页面](../../releases)
2. 选择对应平台的文件下载
3. 解压后即可使用

## 工作流详情

工作流包含三个主要阶段：

1. **Test**: 运行测试套件，确保代码质量
2. **Build**: 多平台并行编译，生成二进制文件
3. **Release**: 仅在推送标签时执行，自动创建发布

所有编译产物都会自动优化（使用 `-ldflags "-s -w"` 减小文件大小）。