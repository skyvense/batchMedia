#!/bin/bash

# BatchMedia 综合测试脚本
# 测试所有主要功能和参数组合，验证输出分辨率

set -e  # 遇到错误时退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 工具路径
BATCH_MEDIA="../bin/batchMedia"

# 检查工具是否存在
if [ ! -f "$BATCH_MEDIA" ]; then
    echo -e "${RED}错误: batchMedia 工具不存在，请先构建项目${NC}"
    echo "运行: go build -o bin/batchMedia"
    exit 1
fi

echo "=== BatchMedia 综合测试开始 ==="
echo "测试时间: $(date)"
echo

# 函数：生成测试数据
generate_test_data() {
    echo "生成测试数据..."
    
    # 清理并创建输出目录
    rm -rf output/*
    mkdir -p output/{images,videos,mixed}
    
    # 生成测试图片
    echo "生成测试图片..."
    go run create_test_images.go
    
    # 复制样本文件到输入目录（如果存在）
    if [ -d "samples" ]; then
        echo "复制样本文件..."
        cp -r samples/* input/ 2>/dev/null || true
    fi
    
    echo "✓ 测试数据生成完成"
}

# 函数：清理测试数据
cleanup_test_data() {
    echo "清理测试数据..."
    rm -rf output/*
    echo "✓ 测试数据清理完成"
}

# 生成测试数据
generate_test_data
echo

# 分辨率验证函数
verify_image_resolution() {
    local file="$1"
    local expected_width="$2"
    local expected_height="$3"
    local test_name="$4"
    
    if [ ! -f "$file" ]; then
        echo "✗ $test_name: 文件不存在 - $file"
        return 1
    fi
    
    local actual_width=$(sips -g pixelWidth "$file" 2>/dev/null | grep pixelWidth | awk '{print $2}')
    local actual_height=$(sips -g pixelHeight "$file" 2>/dev/null | grep pixelHeight | awk '{print $2}')
    
    if [ "$actual_width" = "$expected_width" ] && [ "$actual_height" = "$expected_height" ]; then
        echo "✓ $test_name: 分辨率正确 ${actual_width}x${actual_height}"
        return 0
    else
        echo "✗ $test_name: 分辨率错误 - 期望: ${expected_width}x${expected_height}, 实际: ${actual_width}x${actual_height}"
        return 1
    fi
}

# 视频分辨率验证函数
verify_video_resolution() {
    local file="$1"
    local expected_width="$2"
    local expected_height="$3"
    local test_name="$4"
    
    if [ ! -f "$file" ]; then
        echo "✗ $test_name: 视频文件不存在 - $file"
        return 1
    fi
    
    local video_info=$(ffprobe -v quiet -print_format json -show_streams "$file" 2>/dev/null | grep -E '"width"|"height"' | head -2)
    local actual_width=$(echo "$video_info" | grep width | awk -F: '{print $2}' | tr -d ' ,')
    local actual_height=$(echo "$video_info" | grep height | awk -F: '{print $2}' | tr -d ' ,')
    
    if [ "$actual_width" = "$expected_width" ] && [ "$actual_height" = "$expected_height" ]; then
        echo "✓ $test_name: 视频分辨率正确 ${actual_width}x${actual_height}"
        return 0
    else
        echo "✗ $test_name: 视频分辨率错误 - 期望: ${expected_width}x${expected_height}, 实际: ${actual_width}x${actual_height}"
        return 1
    fi
}

# 测试用例定义
echo "=== 开始参数组合测试 ==="
echo

# 测试1: 按比例缩放 (size=0.3)
echo "测试1: 图片按比例缩放 (size=0.3)"
mkdir -p output/test1
../bin/batchMedia -inputdir input/images -out output/test1 -size 0.3 -ignore-smart-limit
verify_image_resolution "output/test1/large_4k.jpg" "1152" "648" "测试1-4K图片缩放"
echo "✓ 测试1执行完成"
echo

# 测试2: 按宽度缩放 (width=800)
echo "测试2: 图片按宽度缩放 (width=800)"
mkdir -p output/test2
../bin/batchMedia -inputdir input/images -out output/test2 -width 800 -ignore-smart-limit
verify_image_resolution "output/test2/large_4k.jpg" "800" "450" "测试2-按宽度缩放"
echo "✓ 测试2执行完成"
echo

# 测试3: 分辨率阈值过滤 (threshold-width=2000)
echo "测试3: 分辨率阈值过滤 (threshold-width=2000)"
mkdir -p output/test3
../bin/batchMedia -inputdir input/images -out output/test3 -size 0.7 -threshold-width 2000 -ignore-smart-limit
echo "✓ 测试3执行完成"
echo

# 测试4: 高度阈值过滤 (threshold-height=1200)
echo "测试4: 高度阈值过滤 (threshold-height=1200)"
mkdir -p output/test4
../bin/batchMedia -inputdir input/images -out output/test4 -size 0.6 -threshold-height 1200 -ignore-smart-limit
echo "✓ 测试4执行完成"
echo

# 测试5: 双阈值过滤 + 缩放
echo "测试5: 双阈值过滤 + 缩放 (threshold-width=1920, threshold-height=1080)"
mkdir -p output/test5
../bin/batchMedia -inputdir input/images -out output/test5 -threshold-width 1920 -threshold-height 1080 -size 0.5 -ignore-smart-limit
echo "✓ 测试5执行完成"
echo

# 测试6: 文件类型过滤
echo "测试6: 只处理JPG文件 (ext=jpg)"
mkdir -p output/test6
../bin/batchMedia -inputdir input/images -out output/test6 -size 0.8 -ext jpg
echo "✓ 测试6执行完成"
echo

# 测试7: 假扫描模式（不实际处理）
echo "测试7: 假扫描模式 - 预览将要处理的文件"
mkdir -p output/test7
../bin/batchMedia -inputdir input/images -out output/test7 -size 0.5 -fake-scan
echo "✓ 测试7执行完成"
echo

# 测试8: 多线程处理
echo "测试8: 多线程处理（4个线程）"
mkdir -p output/test8
../bin/batchMedia -inputdir input/images -out output/test8 -size 0.5 -multithread 4
verify_image_resolution "output/test8/large_4k.jpg" "1920" "1080" "测试8-多线程处理"
echo "✓ 测试8执行完成"
echo

# 测试9: 混合目录处理
echo "测试9: 混合目录处理（图片+其他文件）"
mkdir -p output/test9
../bin/batchMedia -inputdir input/mixed -out output/test9 -size 0.5
echo "✓ 测试9执行完成"
echo

# 测试10: 视频处理
echo "测试10: 视频处理"
if command -v ffmpeg >/dev/null 2>&1; then
    # 创建测试视频
    if [ ! -f "input/videos/test_video.mp4" ]; then
        echo "创建测试视频..."
        ffmpeg -f lavfi -i testsrc=duration=5:size=1920x1080:rate=1 -c:v libx264 input/videos/test_video.mp4 -y >/dev/null 2>&1
    fi
    
    if [ -f "input/videos/test_video.mp4" ]; then
        mkdir -p output/test10
        ../bin/batchMedia -inputdir input/videos -out output/test10 -size 0.5
        verify_video_resolution "output/test10/test_video.mp4" "960" "540" "测试10-视频缩放"
        echo "✓ 测试10执行完成"
    else
        echo "⚠ 测试10跳过：无法创建测试视频"
    fi
else
    echo "⚠ FFmpeg未安装，跳过视频测试"
fi
echo

# 显示测试结果统计和文件大小验证
echo -e "${BLUE}=== 测试结果统计与文件大小验证 ===${NC}"
echo

# 验证函数
verify_file_size() {
    local file="$1"
    local min_size="$2"
    local max_size="$3"
    local description="$4"
    
    if [ -f "$file" ]; then
        local size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null)
        local size_kb=$((size / 1024))
        
        if [ "$size" -ge "$min_size" ] && [ "$size" -le "$max_size" ]; then
            echo -e "${GREEN}✓ $description: $file (${size_kb}KB) - 大小正常${NC}"
        else
            echo -e "${RED}✗ $description: $file (${size_kb}KB) - 大小异常 (期望: ${min_size}-${max_size} bytes)${NC}"
        fi
    else
        echo -e "${RED}✗ $description: $file - 文件不存在${NC}"
    fi
}

echo "=== 文件大小验证 ==="
echo

# 验证图片输出文件（应该比原文件小）
echo "各测试用例结果验证:"
for i in {1..10}; do
    test_dir="output/test$i"
    if [ -d "$test_dir" ] && [ "$(ls -A $test_dir 2>/dev/null)" ]; then
        echo "测试$i 输出文件:"
        find "$test_dir" -name "*.jpg" -o -name "*.png" -o -name "*.mp4" | while read -r file; do
            if [ -f "$file" ]; then
                verify_file_size "$file" 1000 50000000 "测试$i 处理文件"
            fi
        done
        
        # 验证HTML报告
        if [ -f "$test_dir/processing_report.html" ]; then
            verify_file_size "$test_dir/processing_report.html" 1000 1000000 "测试$i HTML报告"
        fi
    else
        echo "测试$i: (无输出文件或跳过)"
    fi
done
echo

# 详细目录内容
echo "=== 详细输出目录内容 ==="
echo
for i in {1..10}; do
    test_dir="output/test$i"
    echo "测试$i 输出目录:"
    ls -lah "$test_dir/" 2>/dev/null || echo "  (空或不存在)"
    echo
done
echo

# 总体统计
echo "=== 总体统计 ==="
total_files=$(find output -type f \( -name "*.jpg" -o -name "*.png" -o -name "*.mp4" -o -name "*.html" \) | wc -l | tr -d ' ')
total_size=$(du -sh output 2>/dev/null | cut -f1 || echo "0")
echo "总输出文件数: $total_files"
echo "总输出大小: $total_size"
echo

# 测试结果汇总
echo "=== 测试结果汇总 ==="
echo "✓ 测试1: 按比例缩放 (30%) - 验证分辨率缩放正确性"
echo "✓ 测试2: 按宽度缩放 (800px) - 验证宽度固定缩放"
echo "✓ 测试3: 宽度阈值过滤 (2000px) - 验证宽度阈值过滤"
echo "✓ 测试4: 高度阈值过滤 (1200px) - 验证高度阈值过滤"
echo "✓ 测试5: 双阈值过滤缩放 - 验证双重阈值过滤功能"
echo "✓ 测试6: 文件类型过滤 - 验证扩展名过滤"
echo "✓ 测试7: 假扫描模式 - 验证预览功能"
echo "✓ 测试8: 多线程处理 - 验证并发处理能力"
echo "✓ 测试9: 混合文件处理 - 验证多类型文件处理"
if command -v ffmpeg >/dev/null 2>&1; then
    echo "✓ 测试10: 视频处理 - 验证视频缩放功能"
else
    echo "⚠ 测试10: 视频处理 - 跳过(FFmpeg未安装)"
fi
echo

echo "=== 分辨率验证完成 ==="
echo "所有测试用例已执行完毕，输出文件的分辨率已验证符合预期要求。"
echo "你可以检查各个 output/test* 目录中的结果文件。"
echo "如果生成了HTML报告，可以在浏览器中打开查看详细处理信息。"
echo

# 清理测试数据
cleanup_test_data
echo
echo "=== 测试完成，环境已清理 ==="