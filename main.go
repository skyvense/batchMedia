package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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

type ProcessStats struct {
	TotalFiles       int
	ProcessedImages  int
	CopiedFiles      int
	SkippedImages    int
	TotalInputSize   int64
	TotalOutputSize  int64
	ProcessingTime   string
	Files            []FileInfo
}

type FileInfo struct {
	Path         string
	Type         string // "processed", "copied", "skipped"
	InputSize    int64
	OutputSize   int64
	OriginalDim  string
	NewDim       string
	CompressionRatio float64
}

var config Config
var stats ProcessStats

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

		// Check file extension
		ext := strings.ToLower(filepath.Ext(path))
		isSupported := ext == ".jpg" || ext == ".jpeg" || ext == ".heic"
		
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
		
		stats.TotalFiles++
		
		if !isSupported {
			// Copy unsupported files directly
			fmt.Printf("Copying unsupported file: %s (size: %d bytes)\n", path, info.Size())
			stats.CopiedFiles++
			stats.TotalInputSize += info.Size()
			stats.TotalOutputSize += info.Size()
			
			// Record file info
			stats.Files = append(stats.Files, FileInfo{
				Path:         relPath,
				Type:         "copied",
				InputSize:    info.Size(),
				OutputSize:   info.Size(),
				CompressionRatio: 1.0,
			})
			
			return copyFile(path, outputPath, info)
		}

		fmt.Printf("Processing file: %s (size: %d bytes)\n", path, info.Size())
		stats.TotalInputSize += info.Size()

		// Process image
		return processImage(path, outputPath, info)
	})
}

func main() {
	flag.Parse()

	if err := validateConfig(); err != nil {
		log.Fatal(err)
	}

	// Record start time
	startTime := time.Now()

	if err := processImages(); err != nil {
		log.Fatal(err)
	}

	// Record processing time
	stats.ProcessingTime = time.Since(startTime).String()

	// Generate HTML report
	if err := generateHTMLReport(); err != nil {
		fmt.Printf("Warning: failed to generate HTML report: %v\n", err)
	}

	fmt.Println("Batch processing completed!")
	fmt.Printf("Processing summary: %d total files, %d processed, %d copied, %d skipped\n", 
		stats.TotalFiles, stats.ProcessedImages, stats.CopiedFiles, stats.SkippedImages)
}

// generateHTMLReport generates an HTML report of the processing results
func generateHTMLReport() error {
	reportPath := filepath.Join(config.OutputDir, "processing_report.html")
	
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Batch Media Processing Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin: 20px 0; }
        .stat-card { background: #f8f9fa; padding: 15px; border-radius: 5px; text-align: center; }
        .stat-number { font-size: 24px; font-weight: bold; color: #007bff; }
        .stat-label { color: #666; margin-top: 5px; }
        table { width: 100%%; border-collapse: collapse; margin-top: 20px; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f8f9fa; font-weight: bold; }
        .processed { color: #28a745; }
        .copied { color: #ffc107; }
        .skipped { color: #6c757d; }
        .size { text-align: right; }
        .ratio { text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Batch Media Processing Report</h1>
        
        <div class="summary">
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div class="stat-label">Total Files</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div class="stat-label">Processed Images</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div class="stat-label">Copied Files</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div class="stat-label">Skipped Images</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">%.1f MB</div>
                <div class="stat-label">Input Size</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">%.1f MB</div>
                <div class="stat-label">Output Size</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">%.1f%%</div>
                <div class="stat-label">Space Saved</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">%s</div>
                <div class="stat-label">Processing Time</div>
            </div>
        </div>
        
        <h2>File Details</h2>
        <table>
            <thead>
                <tr>
                    <th>File</th>
                    <th>Type</th>
                    <th>Input Size</th>
                    <th>Output Size</th>
                    <th>Compression Ratio</th>
                </tr>
            </thead>
            <tbody>`,
		stats.TotalFiles,
		stats.ProcessedImages,
		stats.CopiedFiles,
		stats.SkippedImages,
		float64(stats.TotalInputSize)/1024/1024,
		float64(stats.TotalOutputSize)/1024/1024,
		(1.0-float64(stats.TotalOutputSize)/float64(stats.TotalInputSize))*100,
		stats.ProcessingTime)
	
	// Add file details
	for _, file := range stats.Files {
		htmlContent += fmt.Sprintf(`
                <tr>
                    <td>%s</td>
                    <td class="%s">%s</td>
                    <td class="size">%.1f KB</td>
                    <td class="size">%.1f KB</td>
                    <td class="ratio">%.2f</td>
                </tr>`,
			file.Path,
			file.Type,
			file.Type,
			float64(file.InputSize)/1024,
			float64(file.OutputSize)/1024,
			file.CompressionRatio)
	}
	
	htmlContent += `
            </tbody>
        </table>
    </div>
</body>
</html>`
	
	return os.WriteFile(reportPath, []byte(htmlContent), 0644)
}
