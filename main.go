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
	InputDir         string
	OutputDir        string
	ScalingRatio     float64
	Width            int
	ThresholdWidth   int
	ThresholdHeight  int
	IgnoreSmartLimit bool
}

var config Config

func init() {
	flag.StringVar(&config.InputDir, "inputdir", "", "Input directory path (required)")
	flag.StringVar(&config.OutputDir, "out", "", "Output directory path (required)")
	flag.Float64Var(&config.ScalingRatio, "size", 0, "Scaling ratio (e.g., 0.5 means scale to 50%)")
	flag.IntVar(&config.Width, "width", 0, "Target width (pixels)")
	flag.IntVar(&config.ThresholdWidth, "threshold-width", 0, "Width threshold (default: 1920 for downscaling, 3840 for upscaling)")
	flag.IntVar(&config.ThresholdHeight, "threshold-height", 0, "Height threshold (default: 1080 for downscaling, 2160 for upscaling)")
	flag.BoolVar(&config.IgnoreSmartLimit, "ignore-smart-limit", false, "Ignore smart default resolution limits")
}

func validateConfig() error {
	if config.InputDir == "" {
		return fmt.Errorf("input directory cannot be empty")
	}

	if config.OutputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	if config.ScalingRatio == 0 && config.Width == 0 {
		return fmt.Errorf("must specify either --size or --width parameter")
	}

	if config.ScalingRatio != 0 && config.Width != 0 {
		return fmt.Errorf("--size and --width parameters cannot be used simultaneously")
	}

	if config.ScalingRatio != 0 && (config.ScalingRatio <= 0 || config.ScalingRatio > 10) {
		return fmt.Errorf("--size parameter must be between 0 and 10")
	}

	if config.Width != 0 && config.Width <= 0 {
		return fmt.Errorf("--width parameter must be greater than 0")
	}

	// Validate threshold parameters
	if config.ThresholdWidth < 0 {
		return fmt.Errorf("--threshold-width parameter must be non-negative")
	}

	if config.ThresholdHeight < 0 {
		return fmt.Errorf("--threshold-height parameter must be non-negative")
	}

	// Apply smart default resolution limits if not ignored
	if !config.IgnoreSmartLimit {
		applySmartDefaults()
	}

	// Check if input directory exists
	if _, err := os.Stat(config.InputDir); os.IsNotExist(err) {
		return fmt.Errorf("input directory does not exist: %s", config.InputDir)
	}

	return nil
}

// applySmartDefaults applies intelligent default resolution limits based on scaling operation
func applySmartDefaults() {
	isDownscaling := false
	isUpscaling := false

	// Determine if operation is downscaling or upscaling
	if config.ScalingRatio > 0 {
		isDownscaling = config.ScalingRatio < 1.0
		isUpscaling = config.ScalingRatio > 1.0
	} else if config.Width > 0 {
		// For width-based scaling, we assume downscaling if target width is common resolution
		// This is a heuristic - in practice, user should specify limits explicitly for width-based scaling
		isDownscaling = config.Width <= 1920
		isUpscaling = config.Width > 1920
	}

	// Apply defaults only if user hasn't specified custom values
	if isDownscaling {
		// For downscaling: set thresholds to avoid processing small images (skip images below threshold)
		if config.ThresholdWidth == 0 {
			config.ThresholdWidth = 1920
			fmt.Printf("Smart default: Setting width threshold to %d (downscaling - skip below)\n", config.ThresholdWidth)
		}
		if config.ThresholdHeight == 0 {
			config.ThresholdHeight = 1080
			fmt.Printf("Smart default: Setting height threshold to %d (downscaling - skip below)\n", config.ThresholdHeight)
		}
	} else if isUpscaling {
		// For upscaling: set thresholds to avoid processing very large images (skip images above threshold)
		if config.ThresholdWidth == 0 {
			config.ThresholdWidth = 3840
			fmt.Printf("Smart default: Setting width threshold to %d (upscaling - skip above)\n", config.ThresholdWidth)
		}
		if config.ThresholdHeight == 0 {
			config.ThresholdHeight = 2160
			fmt.Printf("Smart default: Setting height threshold to %d (upscaling - skip above)\n", config.ThresholdHeight)
		}
	}
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
