package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"

	"github.com/nfnt/resize"
	"github.com/rwcarlsen/goexif/exif"
)

// processImage 处理单个图片文件
func processImage(inputPath, outputPath string, info os.FileInfo) error {
	// 读取整个文件到内存
	fileData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("读取输入文件失败: %v", err)
	}

	// 提取EXIF信息
	exifData, err := extractEXIF(fileData)
	if err != nil {
		// EXIF提取失败不是致命错误，继续处理
		fmt.Printf("警告: 无法提取EXIF信息: %v\n", err)
	}

	// 解码图片
	img, err := jpeg.Decode(bytes.NewReader(fileData))
	if err != nil {
		return fmt.Errorf("解码图片失败: %v", err)
	}

	// 获取原始尺寸
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// 计算新尺寸
	newWidth, newHeight := calculateNewSize(originalWidth, originalHeight)

	// 调整图片大小
	resizedImg := resizeImage(img, newWidth, newHeight)

	// 编码图片到缓冲区
	var buf bytes.Buffer
	options := &jpeg.Options{Quality: 90}
	if err := jpeg.Encode(&buf, resizedImg, options); err != nil {
		return fmt.Errorf("编码图片失败: %v", err)
	}

	// 如果有EXIF数据，尝试将其插入到新图片中
	finalImageData := buf.Bytes()
	if exifData != nil {
		finalImageData = insertEXIF(finalImageData, exifData)
	}

	// 写入输出文件
	if err := os.WriteFile(outputPath, finalImageData, 0644); err != nil {
		return fmt.Errorf("写入输出文件失败: %v", err)
	}

	// 保留原文件的修改时间
	if err := os.Chtimes(outputPath, info.ModTime(), info.ModTime()); err != nil {
		return fmt.Errorf("设置文件时间失败: %v", err)
	}

	fmt.Printf("处理完成: %s (%dx%d -> %dx%d)\n", inputPath, originalWidth, originalHeight, newWidth, newHeight)
	return nil
}

// calculateNewSize 根据配置计算新的图片尺寸
func calculateNewSize(originalWidth, originalHeight int) (int, int) {
	if config.Width > 0 {
		// 按宽度缩放，保持宽高比
		ratio := float64(config.Width) / float64(originalWidth)
		newHeight := int(float64(originalHeight) * ratio)
		return config.Width, newHeight
	}

	if config.Size > 0 {
		// 按比例缩放
		newWidth := int(float64(originalWidth) * config.Size)
		newHeight := int(float64(originalHeight) * config.Size)
		return newWidth, newHeight
	}

	// 默认返回原尺寸
	return originalWidth, originalHeight
}

// resizeImage 使用简单的最近邻算法调整图片大小
func resizeImage(src image.Image, newWidth, newHeight int) image.Image {
	// 使用Lanczos3算法进行高质量缩放
	// Lanczos3提供了最佳的图像质量，特别适合照片缩放
	return resize.Resize(uint(newWidth), uint(newHeight), src, resize.Lanczos3)
}

// extractEXIF 从JPEG文件数据中提取EXIF信息
func extractEXIF(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)
	
	// 查找EXIF数据段
	_, err := exif.Decode(reader)
	if err != nil {
		return nil, err
	}
	
	// 查找APP1段（EXIF数据）
	reader.Seek(0, 0)
	buf := make([]byte, 2)
	
	// 检查JPEG文件头
	if _, err := reader.Read(buf); err != nil {
		return nil, err
	}
	if buf[0] != 0xFF || buf[1] != 0xD8 {
		return nil, fmt.Errorf("不是有效的JPEG文件")
	}
	
	// 查找APP1段
	for {
		if _, err := reader.Read(buf); err != nil {
			return nil, err
		}
		
		if buf[0] != 0xFF {
			continue
		}
		
		// 找到APP1段
		if buf[1] == 0xE1 {
			// 读取段长度
			if _, err := reader.Read(buf); err != nil {
				return nil, err
			}
			length := int(buf[0])<<8 | int(buf[1])
			
			// 读取整个APP1段
			exifSegment := make([]byte, length+2) // +2 for marker
			exifSegment[0] = 0xFF
			exifSegment[1] = 0xE1
			exifSegment[2] = buf[0]
			exifSegment[3] = buf[1]
			
			if _, err := reader.Read(exifSegment[4:]); err != nil {
				return nil, err
			}
			
			return exifSegment, nil
		}
		
		// 如果是其他段，跳过
		if buf[1] >= 0xE0 && buf[1] <= 0xEF {
			if _, err := reader.Read(buf); err != nil {
				return nil, err
			}
			length := int(buf[0])<<8 | int(buf[1])
			reader.Seek(int64(length-2), io.SeekCurrent)
		} else {
			break
		}
	}
	
	return nil, fmt.Errorf("未找到EXIF数据")
}

// insertEXIF 将EXIF数据插入到JPEG文件中
func insertEXIF(jpegData, exifData []byte) []byte {
	if len(jpegData) < 4 || jpegData[0] != 0xFF || jpegData[1] != 0xD8 {
		return jpegData // 不是有效的JPEG文件
	}
	
	// 创建新的JPEG数据
	result := make([]byte, 0, len(jpegData)+len(exifData))
	
	// 添加JPEG文件头
	result = append(result, jpegData[0:2]...)
	
	// 添加EXIF数据
	result = append(result, exifData...)
	
	// 添加剩余的JPEG数据（跳过文件头）
	result = append(result, jpegData[2:]...)
	
	return result
}