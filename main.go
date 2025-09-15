package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	InputDir  string
	OutputDir string
	Size      float64
	Width     int
}

var config Config

func init() {
	flag.StringVar(&config.InputDir, "inputdir", "", "输入目录路径 (必需)")
	flag.StringVar(&config.OutputDir, "out", "", "输出目录路径 (必需)")
	flag.Float64Var(&config.Size, "size", 0, "缩放比例 (例如: 0.5表示缩小到50%)")
	flag.IntVar(&config.Width, "width", 0, "目标宽度 (像素)")
}

func validateConfig() error {
	if config.InputDir == "" {
		return fmt.Errorf("输入目录不能为空")
	}

	if config.OutputDir == "" {
		return fmt.Errorf("输出目录不能为空")
	}

	if config.Size == 0 && config.Width == 0 {
		return fmt.Errorf("必须指定 --size 或 --width 参数")
	}

	if config.Size != 0 && config.Width != 0 {
		return fmt.Errorf("--size 和 --width 参数不能同时使用")
	}

	if config.Size != 0 && (config.Size <= 0 || config.Size > 10) {
		return fmt.Errorf("--size 参数必须在 0 到 10 之间")
	}

	if config.Width != 0 && config.Width <= 0 {
		return fmt.Errorf("--width 参数必须大于 0")
	}

	// 检查输入目录是否存在
	if _, err := os.Stat(config.InputDir); os.IsNotExist(err) {
		return fmt.Errorf("输入目录不存在: %s", config.InputDir)
	}

	return nil
}

func processImages() error {
	// 创建输出目录
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 遍历输入目录中的JPEG文件
	return filepath.Walk(config.InputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 检查是否为JPEG文件
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".jpg" && ext != ".jpeg" {
			return nil
		}

		fmt.Printf("处理文件: %s\n", path)

		// 计算相对路径
		relPath, err := filepath.Rel(config.InputDir, path)
		if err != nil {
			return err
		}

		// 构建输出路径
		outputPath := filepath.Join(config.OutputDir, relPath)

		// 确保输出目录存在
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}

		// 处理图片
		return processImage(path, outputPath, info)
	})
}

func main() {
	flag.Parse()

	if err := validateConfig(); err != nil {
		log.Fatal(err)
	}

	if err := processImages(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("批量处理完成！")
}
