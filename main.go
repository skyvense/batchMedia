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
	flag.StringVar(&config.InputDir, "inputdir", "", "Input directory path (required)")
	flag.StringVar(&config.OutputDir, "out", "", "Output directory path (required)")
	flag.Float64Var(&config.Size, "size", 0, "Scaling ratio (e.g., 0.5 means scale to 50%)")
	flag.IntVar(&config.Width, "width", 0, "Target width (pixels)")
}

func validateConfig() error {
	if config.InputDir == "" {
		return fmt.Errorf("input directory cannot be empty")
	}

	if config.OutputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	if config.Size == 0 && config.Width == 0 {
		return fmt.Errorf("must specify either --size or --width parameter")
	}

	if config.Size != 0 && config.Width != 0 {
		return fmt.Errorf("--size and --width parameters cannot be used simultaneously")
	}

	if config.Size != 0 && (config.Size <= 0 || config.Size > 10) {
		return fmt.Errorf("--size parameter must be between 0 and 10")
	}

	if config.Width != 0 && config.Width <= 0 {
		return fmt.Errorf("--width parameter must be greater than 0")
	}

	// Check if input directory exists
	if _, err := os.Stat(config.InputDir); os.IsNotExist(err) {
		return fmt.Errorf("input directory does not exist: %s", config.InputDir)
	}

	return nil
}

func processImages() error {
	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Walk through JPEG files in input directory
	return filepath.Walk(config.InputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if it's a JPEG file
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".jpg" && ext != ".jpeg" {
			return nil
		}

		fmt.Printf("Processing file: %s\n", path)

		// Calculate relative path
		relPath, err := filepath.Rel(config.InputDir, path)
		if err != nil {
			return err
		}

		// Build output path
		outputPath := filepath.Join(config.OutputDir, relPath)

		// Ensure output directory exists
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}

		// Process image
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

	fmt.Println("Batch processing completed!")
}
