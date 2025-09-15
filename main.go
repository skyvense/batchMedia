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
	// Video processing options
	VideoEnabled     bool
	VideoCodec       string
	VideoBitrate     string
	VideoResolution  string
	VideoCRF         int
	VideoPreset      string
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
	DirectoryStats   map[string]*DirectoryStats // 按目录组织的统计信息
}

type DirectoryStats struct {
	TotalFiles      int
	ProcessedImages int
	CopiedFiles     int
	SkippedImages   int
	TotalInputSize  int64
	TotalOutputSize int64
	Files           []FileInfo
	DirectoryPath   string // 相对于输入目录的路径
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
	stats.DirectoryStats = make(map[string]*DirectoryStats)
	
	flag.StringVar(&config.InputDir, "inputdir", "", "Input directory path (required)")
	flag.StringVar(&config.OutputDir, "out", "", "Output directory path (required)")
	flag.Float64Var(&config.ScalingRatio, "size", 0, "Scaling ratio (e.g., 0.5 means scale to 50%)")
	flag.IntVar(&config.Width, "width", 0, "Target width (pixels)")
	flag.IntVar(&config.ThresholdWidth, "threshold-width", 0, "Width threshold (default: 1920 for downscaling, 3840 for upscaling)")
	flag.IntVar(&config.ThresholdHeight, "threshold-height", 0, "Height threshold (default: 1080 for downscaling, 2160 for upscaling)")
	flag.BoolVar(&config.IgnoreSmartLimit, "ignore-smart-limit", false, "Ignore smart default resolution limits")
	// Video processing flags
	flag.BoolVar(&config.VideoEnabled, "video", false, "Enable video processing")
	flag.StringVar(&config.VideoCodec, "video-codec", "libx265", "Video codec (libx264, libx265, etc.)")
	flag.StringVar(&config.VideoBitrate, "video-bitrate", "", "Video bitrate (e.g., 2M, 1000k)")
	flag.StringVar(&config.VideoResolution, "video-resolution", "", "Video resolution (e.g., 1920x1080, 1280x720)")
	flag.IntVar(&config.VideoCRF, "video-crf", 23, "Video CRF quality (0-51, lower is better quality)")
	flag.StringVar(&config.VideoPreset, "video-preset", "medium", "Video encoding preset (ultrafast, fast, medium, slow, veryslow)")
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

	// First pass: count total files to process
	totalFilesToProcess := 0
	filepath.Walk(config.InputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		isImageSupported := ext == ".jpg" || ext == ".jpeg" || ext == ".heic" || ext == ".png"
		isVideoSupported := isVideoFile(path)
		if isImageSupported || isVideoSupported {
			totalFilesToProcess++
		}
		return nil
	})

	// Progress counter
	processedCount := 0

	// Walk through files in input directory
	return filepath.Walk(config.InputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check file extension
		ext := strings.ToLower(filepath.Ext(path))
		isImageSupported := ext == ".jpg" || ext == ".jpeg" || ext == ".heic" || ext == ".png"
		isVideoSupported := isVideoFile(path) // Auto-enable video processing when video files are detected
		
		// Calculate relative path
		relPath, err := filepath.Rel(config.InputDir, path)
		if err != nil {
			return err
		}
		
		// Get directory path for this file
		dirPath := filepath.Dir(relPath)
		if dirPath == "." {
			dirPath = "" // Root directory
		}
		
		// Initialize directory stats if not exists
		if _, exists := stats.DirectoryStats[dirPath]; !exists {
			stats.DirectoryStats[dirPath] = &DirectoryStats{
				DirectoryPath: dirPath,
				Files:         make([]FileInfo, 0),
			}
		}
		dirStats := stats.DirectoryStats[dirPath]
		
		// Build output path
		outputPath := filepath.Join(config.OutputDir, relPath)
		
		// Convert HEIC files to JPEG extension since we encode them as JPEG
		if strings.ToLower(filepath.Ext(path)) == ".heic" {
			outputPath = strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".jpg"
		}
		
		// Ensure output directory exists
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}
		
		stats.TotalFiles++
		dirStats.TotalFiles++
		
		if isVideoSupported {
			// Process video file
			processedCount++
			percentage := float64(processedCount) / float64(totalFilesToProcess) * 100
			fmt.Printf("[%d/%d] (%.1f%%) Processing video: %s (size: %d bytes)\n", processedCount, totalFilesToProcess, percentage, path, info.Size())
			stats.TotalInputSize += info.Size()
			dirStats.TotalInputSize += info.Size()
			return processVideo(path, outputPath, info, dirStats)
		} else if isImageSupported {
			// Process image file
			processedCount++
			percentage := float64(processedCount) / float64(totalFilesToProcess) * 100
			fmt.Printf("[%d/%d] (%.1f%%) Processing image: %s (size: %d bytes)\n", processedCount, totalFilesToProcess, percentage, path, info.Size())
			stats.TotalInputSize += info.Size()
			dirStats.TotalInputSize += info.Size()
			return processImage(path, outputPath, info, dirStats)
		} else {
			// Copy unsupported files directly
			fmt.Printf("Copying unsupported file: %s (size: %d bytes)\n", path, info.Size())
			stats.CopiedFiles++
			dirStats.CopiedFiles++
			stats.TotalInputSize += info.Size()
			stats.TotalOutputSize += info.Size()
			dirStats.TotalInputSize += info.Size()
			dirStats.TotalOutputSize += info.Size()
			
			// Record file info
			fileInfo := FileInfo{
				Path:         relPath,
				Type:         "copied",
				InputSize:    info.Size(),
				OutputSize:   info.Size(),
				CompressionRatio: 1.0,
			}
			stats.Files = append(stats.Files, fileInfo)
			dirStats.Files = append(dirStats.Files, fileInfo)
			
			return copyFile(path, outputPath, info)
		}
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

	// Generate HTML reports for each directory
	for dirPath, dirStats := range stats.DirectoryStats {
		if len(dirStats.Files) > 0 { // Only generate report if directory has processed files
			if err := generateDirectoryHTMLReport(dirPath, dirStats); err != nil {
				fmt.Printf("Warning: failed to generate HTML report for directory '%s': %v\n", dirPath, err)
			}
		}
	}

	// Generate overall HTML report
	if err := generateHTMLReport(); err != nil {
		fmt.Printf("Warning: failed to generate overall HTML report: %v\n", err)
	}

	fmt.Println("Batch processing completed!")
	fmt.Printf("Processing summary: %d total files, %d processed, %d copied, %d skipped\n", 
		stats.TotalFiles, stats.ProcessedImages, stats.CopiedFiles, stats.SkippedImages)
}

// generateDirectoryHTMLReport generates an HTML report for a specific directory
func generateDirectoryHTMLReport(dirPath string, dirStats *DirectoryStats) error {
	// Determine output path for the report
	var reportPath string
	if dirPath == "" {
		// Root directory
		reportPath = filepath.Join(config.OutputDir, "processing_report.html")
	} else {
		// Subdirectory
		reportPath = filepath.Join(config.OutputDir, dirPath, "processing_report.html")
	}
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %v", err)
	}
	
	// Calculate space saved percentage
	spaceSavedPercent := 0.0
	if dirStats.TotalInputSize > 0 {
		spaceSavedPercent = (1.0 - float64(dirStats.TotalOutputSize)/float64(dirStats.TotalInputSize)) * 100
	}
	
	// Generate directory title
	dirTitle := "Root Directory"
	if dirPath != "" {
		dirTitle = fmt.Sprintf("Directory: %s", dirPath)
	}
	
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Processing Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1400px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin: 20px 0; }
        .stat-card { background: #f8f9fa; padding: 15px; border-radius: 5px; text-align: center; }
        .stat-number { font-size: 24px; font-weight: bold; color: #007bff; }
        .stat-label { color: #666; margin-top: 5px; }
        
        /* Grid layout for files */
        .files-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 20px; margin-top: 20px; }
        .file-card { background: #fff; border: 1px solid #ddd; border-radius: 8px; padding: 15px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); transition: transform 0.2s; }
        .file-card:hover { transform: translateY(-2px); box-shadow: 0 4px 10px rgba(0,0,0,0.15); }
        
        .file-header { display: flex; align-items: center; margin-bottom: 10px; }
        .file-name { font-weight: bold; color: #333; text-decoration: none; flex: 1; }
        .file-name:hover { color: #007bff; }
        .file-type { padding: 3px 8px; border-radius: 12px; font-size: 12px; font-weight: bold; text-transform: uppercase; }
        .processed { background: #d4edda; color: #155724; }
        .video_processed { background: #d1ecf1; color: #0c5460; }
        .copied { background: #fff3cd; color: #856404; }
        .skipped { background: #f8d7da; color: #721c24; }
        
        .thumbnail { width: 100%%; height: 200px; object-fit: cover; border-radius: 5px; margin: 10px 0; background: #f8f9fa; display: flex; align-items: center; justify-content: center; color: #666; }
        .video-placeholder { background: #e9ecef; border: 2px dashed #adb5bd; }
        
        .file-details { font-size: 14px; color: #666; }
        .detail-row { display: flex; justify-content: space-between; margin: 5px 0; }
        .detail-label { font-weight: 500; }
        
        .size-info { display: flex; justify-content: space-between; align-items: center; margin-top: 10px; padding-top: 10px; border-top: 1px solid #eee; }
        .compression-ratio { font-weight: bold; color: #28a745; }
        
        h2 { color: #333; margin-top: 30px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>%s</h1>
        
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
                <div class="stat-number">%.1f%%%%</div>
                <div class="stat-label">Space Saved</div>
            </div>
        </div>
        
        <h2>Processed Files</h2>
        <div class="files-grid">`,
		dirTitle, dirTitle,
		dirStats.TotalFiles,
		dirStats.ProcessedImages,
		dirStats.CopiedFiles,
		dirStats.SkippedImages,
		float64(dirStats.TotalInputSize)/1024/1024,
		float64(dirStats.TotalOutputSize)/1024/1024,
		spaceSavedPercent)
	
	// Add file cards for this directory
	for _, file := range dirStats.Files {
		// Determine if it's an image file for thumbnail
		filePath := file.Path
		ext := strings.ToLower(filepath.Ext(filePath))
		isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".heic"
		isVideo := strings.Contains(file.Type, "video") || ext == ".mov" || ext == ".mp4" || ext == ".avi" || ext == ".mkv"
		
		// Handle HEIC files that were converted to JPG
		actualFilePath := filePath
		if ext == ".heic" {
			// HEIC files are converted to JPG, so update the link path
			actualFilePath = strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".jpg"
		}
		
		// Create thumbnail or placeholder
		var thumbnailHTML string
		if isImage {
			thumbnailHTML = fmt.Sprintf(`<img src="%s" alt="%s" class="thumbnail" onerror="this.style.display='none'; this.nextElementSibling.style.display='flex';"><div class="thumbnail" style="display:none;">📷 Image Preview</div>`, actualFilePath, actualFilePath)
		} else if isVideo {
			thumbnailHTML = `<div class="thumbnail video-placeholder">🎬 Video File</div>`
		} else {
			thumbnailHTML = `<div class="thumbnail">📄 File</div>`
		}
		
		htmlContent += fmt.Sprintf(`
            <div class="file-card">
                <div class="file-header">
                    <a href="%s" class="file-name" target="_blank">%s</a>
                    <span class="file-type %s">%s</span>
                </div>
                %s
                <div class="file-details">
                    <div class="detail-row">
                        <span class="detail-label">Original Size:</span>
                        <span>%.1f KB</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Output Size:</span>
                        <span>%.1f KB</span>
                    </div>`,
			actualFilePath,
			filePath,
			file.Type,
			file.Type,
			thumbnailHTML,
			float64(file.InputSize)/1024,
			float64(file.OutputSize)/1024)
		
		// Add dimension info if available
		if file.OriginalDim != "" && file.NewDim != "" {
			htmlContent += fmt.Sprintf(`
                    <div class="detail-row">
                        <span class="detail-label">Dimensions:</span>
                        <span>%s → %s</span>
                    </div>`, file.OriginalDim, file.NewDim)
		}
		
		htmlContent += fmt.Sprintf(`
                </div>
                <div class="size-info">
                    <span>Compression Ratio:</span>
                    <span class="compression-ratio">%.2f</span>
                </div>
            </div>`, file.CompressionRatio)
	}
	
	htmlContent += `
        </div>
    </div>
</body>
</html>`
	
	return os.WriteFile(reportPath, []byte(htmlContent), 0644)
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
        .container { max-width: 1400px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin: 20px 0; }
        .stat-card { background: #f8f9fa; padding: 15px; border-radius: 5px; text-align: center; }
        .stat-number { font-size: 24px; font-weight: bold; color: #007bff; }
        .stat-label { color: #666; margin-top: 5px; }
        
        /* Grid layout for files */
        .files-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 20px; margin-top: 20px; }
        .file-card { background: #fff; border: 1px solid #ddd; border-radius: 8px; padding: 15px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); transition: transform 0.2s; }
        .file-card:hover { transform: translateY(-2px); box-shadow: 0 4px 10px rgba(0,0,0,0.15); }
        
        .file-header { display: flex; align-items: center; margin-bottom: 10px; }
        .file-name { font-weight: bold; color: #333; text-decoration: none; flex: 1; }
        .file-name:hover { color: #007bff; }
        .file-type { padding: 3px 8px; border-radius: 12px; font-size: 12px; font-weight: bold; text-transform: uppercase; }
        .processed { background: #d4edda; color: #155724; }
        .video_processed { background: #d1ecf1; color: #0c5460; }
        .copied { background: #fff3cd; color: #856404; }
        .skipped { background: #f8d7da; color: #721c24; }
        
        .thumbnail { width: 100%%; height: 200px; object-fit: cover; border-radius: 5px; margin: 10px 0; background: #f8f9fa; display: flex; align-items: center; justify-content: center; color: #666; }
        .video-placeholder { background: #e9ecef; border: 2px dashed #adb5bd; }
        
        .file-details { font-size: 14px; color: #666; }
        .detail-row { display: flex; justify-content: space-between; margin: 5px 0; }
        .detail-label { font-weight: 500; }
        
        .size-info { display: flex; justify-content: space-between; align-items: center; margin-top: 10px; padding-top: 10px; border-top: 1px solid #eee; }
        .compression-ratio { font-weight: bold; color: #28a745; }
        
        h2 { color: #333; margin-top: 30px; }
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
        
        <h2>Processed Files</h2>
        <div class="files-grid">`,
		stats.TotalFiles,
		stats.ProcessedImages,
		stats.CopiedFiles,
		stats.SkippedImages,
		float64(stats.TotalInputSize)/1024/1024,
		float64(stats.TotalOutputSize)/1024/1024,
		(1.0-float64(stats.TotalOutputSize)/float64(stats.TotalInputSize))*100,
		stats.ProcessingTime)
	
	// Add file cards
	for _, file := range stats.Files {
		// Determine if it's an image file for thumbnail
		filePath := file.Path
		ext := strings.ToLower(filepath.Ext(filePath))
		isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".heic"
		isVideo := strings.Contains(file.Type, "video") || ext == ".mov" || ext == ".mp4" || ext == ".avi" || ext == ".mkv"
		
		// Handle HEIC files that were converted to JPG
		actualFilePath := filePath
		if ext == ".heic" {
			// HEIC files are converted to JPG, so update the link path
			actualFilePath = strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".jpg"
		}
		
		// Create thumbnail or placeholder
		var thumbnailHTML string
		if isImage {
			thumbnailHTML = fmt.Sprintf(`<img src="%s" alt="%s" class="thumbnail" onerror="this.style.display='none'; this.nextElementSibling.style.display='flex';"><div class="thumbnail" style="display:none;">📷 Image Preview</div>`, actualFilePath, actualFilePath)
		} else if isVideo {
			thumbnailHTML = `<div class="thumbnail video-placeholder">🎬 Video File</div>`
		} else {
			thumbnailHTML = `<div class="thumbnail">📄 File</div>`
		}
		
		htmlContent += fmt.Sprintf(`
            <div class="file-card">
                <div class="file-header">
                    <a href="%s" class="file-name" target="_blank">%s</a>
                    <span class="file-type %s">%s</span>
                </div>
                %s
                <div class="file-details">
                    <div class="detail-row">
                        <span class="detail-label">Original Size:</span>
                        <span>%.1f KB</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Output Size:</span>
                        <span>%.1f KB</span>
                    </div>`,
			actualFilePath,
			filePath,
			file.Type,
			file.Type,
			thumbnailHTML,
			float64(file.InputSize)/1024,
			float64(file.OutputSize)/1024)
		
		// Add dimension info if available
		if file.OriginalDim != "" && file.NewDim != "" {
			htmlContent += fmt.Sprintf(`
                    <div class="detail-row">
                        <span class="detail-label">Dimensions:</span>
                        <span>%s → %s</span>
                    </div>`, file.OriginalDim, file.NewDim)
		}
		
		htmlContent += fmt.Sprintf(`
                </div>
                <div class="size-info">
                    <span>Compression Ratio:</span>
                    <span class="compression-ratio">%.2f</span>
                </div>
            </div>`, file.CompressionRatio)
	}
	
	htmlContent += `
        </div>
    </div>
</body>
</html>`
	
	return os.WriteFile(reportPath, []byte(htmlContent), 0644)
}
