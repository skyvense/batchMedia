package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
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
	// File filtering options
	Extensions       string // Comma-separated list of extensions to process
	FakeScan         bool   // Only scan and list files to be processed, don't actually process
	// Video processing options
	VideoDisabled    bool
	VideoCodec       string
	VideoBitrate     string
	VideoResolution  string
	VideoCRF         int
	VideoPreset      string
	// Multithreading options
	Multithread      int    // Number of concurrent threads for processing multiple directories
}

// DirectoryProgress represents the processing progress of a directory
type DirectoryProgress struct {
	Path      string `json:"path"`
	Completed bool   `json:"completed"`
	Timestamp string `json:"timestamp,omitempty"`
}

// ProgressTracker manages the processing progress
type ProgressTracker struct {
	Directories []DirectoryProgress `json:"directories"`
	LastUpdate  string              `json:"last_update"`
}

// loadProgress loads the progress from file
func loadProgress(progressFile string) (*ProgressTracker, error) {
	if _, err := os.Stat(progressFile); os.IsNotExist(err) {
		return &ProgressTracker{Directories: []DirectoryProgress{}}, nil
	}

	data, err := ioutil.ReadFile(progressFile)
	if err != nil {
		return nil, err
	}

	var tracker ProgressTracker
	err = json.Unmarshal(data, &tracker)
	if err != nil {
		return nil, err
	}

	return &tracker, nil
}

// saveProgress saves the progress to file
func (pt *ProgressTracker) saveProgress(progressFile string) error {
	pt.LastUpdate = time.Now().Format(time.RFC3339)
	data, err := json.MarshalIndent(pt, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(progressFile, data, 0644)
}

// scanDirectories recursively scans for all directories to process
func scanDirectories(inputDir string) ([]string, error) {
	var directories []string
	
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip the root input directory itself
		if path == inputDir {
			return nil
		}
		
		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}
		
		// Add all directories (including nested ones)
		if info.IsDir() {
			directories = append(directories, path)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// Sort directories to process from deepest to shallowest
	// This ensures we process leaf directories first
	sort.Slice(directories, func(i, j int) bool {
		depthI := strings.Count(directories[i], string(filepath.Separator))
		depthJ := strings.Count(directories[j], string(filepath.Separator))
		return depthI > depthJ // Deeper directories first
	})
	
	return directories, nil
}

// markDirectoryCompleted marks a directory as completed in the progress tracker
func (pt *ProgressTracker) markDirectoryCompleted(dirPath string) {
	for i := range pt.Directories {
		if pt.Directories[i].Path == dirPath {
			pt.Directories[i].Completed = true
			pt.Directories[i].Timestamp = time.Now().Format(time.RFC3339)
			return
		}
	}
}

// getUncompletedDirectories returns directories that haven't been completed
func (pt *ProgressTracker) getUncompletedDirectories() []string {
	var uncompleted []string
	for _, dir := range pt.Directories {
		if !dir.Completed {
			uncompleted = append(uncompleted, dir.Path)
		}
	}
	return uncompleted
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
var statsMutex sync.Mutex
var progressMutex sync.Mutex

func init() {
	stats.DirectoryStats = make(map[string]*DirectoryStats)
	
	flag.StringVar(&config.InputDir, "inputdir", "", "Input directory path (required)")
	flag.StringVar(&config.OutputDir, "out", "", "Output directory path (required)")
	flag.Float64Var(&config.ScalingRatio, "size", 0, "Scaling ratio (e.g., 0.5 means scale to 50%)")
	flag.IntVar(&config.Width, "width", 0, "Target width (pixels)")
	flag.IntVar(&config.ThresholdWidth, "threshold-width", 0, "Width threshold (default: 1920 for downscaling, 3840 for upscaling)")
	flag.IntVar(&config.ThresholdHeight, "threshold-height", 0, "Height threshold (default: 1080 for downscaling, 2160 for upscaling)")
	flag.BoolVar(&config.IgnoreSmartLimit, "ignore-smart-limit", false, "Ignore smart default resolution limits")
	// File filtering flags
	flag.StringVar(&config.Extensions, "ext", "", "Process only files with specified extensions (comma-separated, e.g., heic,jpg,png)")
	flag.BoolVar(&config.FakeScan, "fake-scan", false, "Only scan and list files to be processed, don't actually process them")
	// Video processing flags
	flag.BoolVar(&config.VideoDisabled, "disable-video", false, "Disable video processing (video processing is enabled by default)")
	flag.StringVar(&config.VideoCodec, "video-codec", "libx265", "Video codec (libx264, libx265, etc.)")
	flag.StringVar(&config.VideoBitrate, "video-bitrate", "", "Video bitrate (e.g., 2M, 1000k)")
	flag.StringVar(&config.VideoResolution, "video-resolution", "", "Video resolution (e.g., 1920x1080, 1280x720)")
	flag.IntVar(&config.VideoCRF, "video-crf", 23, "Video CRF quality (0-51, lower is better quality)")
	flag.StringVar(&config.VideoPreset, "video-preset", "medium", "Video encoding preset (ultrafast, fast, medium, slow, veryslow)")
	// Multithreading flags
	flag.IntVar(&config.Multithread, "multithread", 1, "Number of concurrent threads for processing multiple directories (default: 1)")
}

func validateConfig() error {
	if config.InputDir == "" {
		return fmt.Errorf("input directory cannot be empty")
	}

	if config.OutputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	// Skip size/width validation in fake scan mode
	if !config.FakeScan {
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
// shouldProcessExtension checks if the file extension should be processed based on the -ext filter
func shouldProcessExtension(filePath string) bool {
	// If no extension filter is specified, process all supported files
	if config.Extensions == "" {
		return true
	}
	
	// Parse the extensions list
	allowedExts := strings.Split(strings.ToLower(config.Extensions), ",")
	fileExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))
	
	// Check if the file extension is in the allowed list
	for _, ext := range allowedExts {
		ext = strings.TrimSpace(ext)
		if fileExt == ext {
			return true
		}
	}
	
	return false
}

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

func processImages(targetDir string, threadID int) error {
	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// First pass: count total files to process in the target directory
	totalFilesToProcess := 0
	walkDir := config.InputDir
	if targetDir != "" {
		walkDir = targetDir
	}
	
	// Read directory contents directly (non-recursive)
	entries, err := os.ReadDir(walkDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", walkDir, err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}
		
		filename := entry.Name()
		path := filepath.Join(walkDir, filename)
		
		// Skip hidden files (macOS metadata files starting with ._)
		if strings.HasPrefix(filename, "._") {
			continue
		}
		
		// Check if file extension should be processed based on filter
		if !shouldProcessExtension(path) {
			continue
		}
		
		ext := strings.ToLower(filepath.Ext(path))
		isImageSupported := ext == ".jpg" || ext == ".jpeg" || ext == ".heic" || ext == ".png"
		isVideoSupported := isVideoFile(path)
		if isImageSupported || isVideoSupported {
			totalFilesToProcess++
		}
	}

	// Progress counter
	processedCount := 0

	// Process files in target directory (non-recursive)
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		filename := entry.Name()
		path := filepath.Join(walkDir, filename)

		// Skip hidden files (macOS metadata files starting with ._)
		if strings.HasPrefix(filename, "._") {
			continue
		}

		// Check if file extension should be processed based on filter
		if !shouldProcessExtension(path) {
			continue
		}

		// Get file info
		info, err := entry.Info()
		if err != nil {
			fmt.Printf("Warning: failed to get file info for %s: %v\n", path, err)
			continue
		}
		
		// Check file extension
		ext := strings.ToLower(filepath.Ext(path))
		isImageSupported := ext == ".jpg" || ext == ".jpeg" || ext == ".heic" || ext == ".png"
		isVideoSupported := isVideoFile(path) && !config.VideoDisabled // Video processing enabled by default unless disabled
		
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
		
		if config.FakeScan {
			// Fake scan mode: only list files to be processed
			processedCount++
			percentage := float64(processedCount) / float64(totalFilesToProcess) * 100
			if isVideoSupported {
				fmt.Printf("[thread-%d] [%d/%d] (%.1f%%) Would process video: %s (size: %d bytes) -> %s\n", threadID, processedCount, totalFilesToProcess, percentage, path, info.Size(), outputPath)
			} else if isImageSupported {
				fmt.Printf("[thread-%d] [%d/%d] (%.1f%%) Would process image: %s (size: %d bytes) -> %s\n", threadID, processedCount, totalFilesToProcess, percentage, path, info.Size(), outputPath)
			} else {
				fmt.Printf("[thread-%d] [%d/%d] (%.1f%%) Would copy file: %s (size: %d bytes) -> %s\n", threadID, processedCount, totalFilesToProcess, percentage, path, info.Size(), outputPath)
			}
			stats.TotalInputSize += info.Size()
			dirStats.TotalInputSize += info.Size()
			continue
		}
		
		if isVideoSupported {
			// Process video file
			processedCount++
			percentage := float64(processedCount) / float64(totalFilesToProcess) * 100
			fmt.Printf("[thread-%d] [%d/%d] (%.1f%%) Processing video: %s (size: %d bytes)\n", threadID, processedCount, totalFilesToProcess, percentage, path, info.Size())
			stats.TotalInputSize += info.Size()
			dirStats.TotalInputSize += info.Size()
			err = processVideo(path, outputPath, info, dirStats)
			if err != nil {
				fmt.Printf("Error processing video %s: %v\n", path, err)
			}
		} else if isImageSupported {
			// Process image file
			processedCount++
			percentage := float64(processedCount) / float64(totalFilesToProcess) * 100
			fmt.Printf("[thread-%d] [%d/%d] (%.1f%%) Processing image: %s (size: %d bytes)\n", threadID, processedCount, totalFilesToProcess, percentage, path, info.Size())
			stats.TotalInputSize += info.Size()
			dirStats.TotalInputSize += info.Size()
			err = processImage(path, outputPath, relPath, info, dirStats)
			if err != nil {
				fmt.Printf("Error processing image %s: %v\n", path, err)
			}
		} else {
			// Copy unsupported files directly
			fmt.Printf("[thread-%d] Copying unsupported file: %s (size: %d bytes)\n", threadID, path, info.Size())
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
			
			err = copyFile(path, outputPath, info)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}

func main() {
	flag.Parse()

	if err := validateConfig(); err != nil {
		log.Fatal(err)
	}

	// Handle fake scan mode - skip progress file operations
	// Progress file path - use extension-specific name if filtering by extension
	progressFileName := "progress.json"
	if config.Extensions != "" {
		// Replace commas and spaces with underscores for filename
		extSuffix := strings.ReplaceAll(strings.ReplaceAll(config.Extensions, ",", "_"), " ", "")
		progressFileName = fmt.Sprintf("progress_%s.json", extSuffix)
	}
	progressFile := filepath.Join(config.OutputDir, progressFileName)

	// Load existing progress
	tracker, err := loadProgress(progressFile)
	if err != nil {
		log.Fatalf("Failed to load progress: %v", err)
	}

	if config.FakeScan {
		// Fake scan mode: use progress file but don't save changes or do actual processing
		// Scan directories if progress is empty
		if len(tracker.Directories) == 0 {
			fmt.Println("Scanning directories...")
			directories, err := scanDirectories(config.InputDir)
			if err != nil {
				log.Fatalf("Failed to scan directories: %v", err)
			}

			// If no subdirectories found, process the root directory itself
			if len(directories) == 0 {
				directories = append(directories, config.InputDir)
			}
			
			// Initialize progress tracker (but don't save it)
			for _, dir := range directories {
				tracker.Directories = append(tracker.Directories, DirectoryProgress{
					Path:      dir,
					Completed: false,
				})
			}
			fmt.Printf("Found %d directories to process\n", len(directories))
		}

		// Get uncompleted directories
		uncompletedDirs := tracker.getUncompletedDirectories()
		if len(uncompletedDirs) == 0 {
			fmt.Println("All directories have been processed!")
			return
		}

		fmt.Printf("Processing %d remaining directories...\n", len(uncompletedDirs))

		// Record start time
		startTime := time.Now()

		// Process directories with multithreading support in fake scan mode
		if len(uncompletedDirs) <= 1 || config.Multithread <= 1 {
			// Single-threaded processing for 1 directory or when multithread is disabled
			for i, dirPath := range uncompletedDirs {
				fmt.Printf("[%d/%d] Processing directory: %s\n", i+1, len(uncompletedDirs), dirPath)
				
				// Process this directory
				if err := processImages(dirPath, 0); err != nil {
					fmt.Printf("Error processing directory %s: %v\n", dirPath, err)
					continue
				}
				
				// Skip HTML report generation in fake scan mode
				if config.Extensions != "" {
					fmt.Printf("Skipping HTML report generation (extension filter active: %s)\n", config.Extensions)
				}
				
				fmt.Printf("Completed directory: %s\n", dirPath)
			}
		} else {
			// Multi-threaded processing
			fmt.Printf("Using %d threads for parallel processing\n", config.Multithread)
			
			// Create semaphore to limit concurrent goroutines
			semaphore := make(chan struct{}, config.Multithread)
			var wg sync.WaitGroup
			
			for i, dirPath := range uncompletedDirs {
				wg.Add(1)
				go func(index int, path string) {
					defer wg.Done()
					
					// Acquire semaphore
					semaphore <- struct{}{}
					defer func() { <-semaphore }()
					
					fmt.Printf("[%d/%d] Processing directory: %s\n", index+1, len(uncompletedDirs), path)
					
					// Process this directory
					if err := processImages(path, index+1); err != nil {
						fmt.Printf("Error processing directory %s: %v\n", path, err)
						return
					}
					
					// Skip HTML report generation in fake scan mode
					if config.Extensions != "" {
						fmt.Printf("Skipping HTML report generation (extension filter active: %s)\n", config.Extensions)
					}
					
					fmt.Printf("Completed directory: %s\n", path)
				}(i, dirPath)
			}
			
			// Wait for all goroutines to complete
			wg.Wait()
		}

		// Record processing time
		processingTime := time.Since(startTime).String()

		fmt.Println("Batch processing completed!")
		fmt.Printf("Total processing time: %s\n", processingTime)
		return
	}

	// Normal mode: use progress file tracking

	// Scan directories if progress is empty
	if len(tracker.Directories) == 0 {
		fmt.Println("Scanning directories...")
		directories, err := scanDirectories(config.InputDir)
		if err != nil {
			log.Fatalf("Failed to scan directories: %v", err)
		}

		// If no subdirectories found, process the root directory itself
		if len(directories) == 0 {
			directories = append(directories, config.InputDir)
		}
		
		// Initialize progress tracker
		for _, dir := range directories {
			tracker.Directories = append(tracker.Directories, DirectoryProgress{
				Path:      dir,
				Completed: false,
			})
		}

		// Save initial progress
		if err := tracker.saveProgress(progressFile); err != nil {
			log.Fatalf("Failed to save initial progress: %v", err)
		}
		fmt.Printf("Found %d directories to process\n", len(directories))
	}

	// Get uncompleted directories
	uncompletedDirs := tracker.getUncompletedDirectories()
	if len(uncompletedDirs) == 0 {
		fmt.Println("All directories have been processed!")
		return
	}

	fmt.Printf("Processing %d remaining directories...\n", len(uncompletedDirs))

	// Record start time
	startTime := time.Now()

	// Process directories with multithreading support
	if len(uncompletedDirs) <= 1 || config.Multithread <= 1 {
		// Single-threaded processing for 1 directory or when multithread is disabled
		for i, dirPath := range uncompletedDirs {
			fmt.Printf("[%d/%d] Processing directory: %s\n", i+1, len(uncompletedDirs), dirPath)
			
			// Process this directory
			if err := processImages(dirPath, 0); err != nil {
				fmt.Printf("Error processing directory %s: %v\n", dirPath, err)
				continue
			}
			
			// Mark directory as completed
			tracker.markDirectoryCompleted(dirPath)
			
			// Save progress after each directory
			if err := tracker.saveProgress(progressFile); err != nil {
				fmt.Printf("Warning: failed to save progress: %v\n", err)
			}
			
			// Generate HTML report for this directory only (skip if using extension filter)
			if config.Extensions == "" {
				for dirPath, dirStats := range stats.DirectoryStats {
					if len(dirStats.Files) > 0 {
						if err := generateDirectoryHTMLReport(dirPath, dirStats); err != nil {
							fmt.Printf("Warning: failed to generate HTML report for directory '%s': %v\n", dirPath, err)
						}
					}
				}
			} else {
				fmt.Printf("Skipping HTML report generation (extension filter active: %s)\n", config.Extensions)
			}
			
			// Reset stats for next directory
			stats = ProcessStats{DirectoryStats: make(map[string]*DirectoryStats)}
			
			fmt.Printf("Completed directory: %s\n", dirPath)
		}
	} else {
		// Multi-threaded processing for multiple directories
		fmt.Printf("Using %d threads for parallel processing\n", config.Multithread)
		
		// Create a semaphore to limit concurrent goroutines
		semaphore := make(chan struct{}, config.Multithread)
		var wg sync.WaitGroup
		
		for i, dirPath := range uncompletedDirs {
			wg.Add(1)
			go func(dir string, index int) {
				defer wg.Done()
				
				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				
				fmt.Printf("[%d/%d] Processing directory: %s\n", index+1, len(uncompletedDirs), dir)
				
				// Process this directory
				if err := processImages(dir, index+1); err != nil {
					fmt.Printf("Error processing directory %s: %v\n", dir, err)
					return
				}
				
				// Thread-safe operations with mutex
				progressMutex.Lock()
				tracker.markDirectoryCompleted(dir)
				if err := tracker.saveProgress(progressFile); err != nil {
					fmt.Printf("Warning: failed to save progress: %v\n", err)
				}
				progressMutex.Unlock()
				
				// Generate HTML report (thread-safe)
				statsMutex.Lock()
				if config.Extensions == "" {
					for dirPath, dirStats := range stats.DirectoryStats {
						if len(dirStats.Files) > 0 {
							if err := generateDirectoryHTMLReport(dirPath, dirStats); err != nil {
								fmt.Printf("Warning: failed to generate HTML report for directory '%s': %v\n", dirPath, err)
							}
						}
					}
				} else {
					fmt.Printf("Skipping HTML report generation (extension filter active: %s)\n", config.Extensions)
				}
				// Reset stats for next directory
				stats = ProcessStats{DirectoryStats: make(map[string]*DirectoryStats)}
				statsMutex.Unlock()
				
				fmt.Printf("Completed directory: %s\n", dir)
			}(dirPath, i)
		}
		
		// Wait for all goroutines to complete
		wg.Wait()
		fmt.Println("All directories processed in parallel")
	}

	// Record processing time
	processingTime := time.Since(startTime).String()

	fmt.Println("Batch processing completed!")
	fmt.Printf("Total processing time: %s\n", processingTime)
}

// generateDirectoryHTMLReport generates an HTML report for a specific directory
func generateDirectoryHTMLReport(currentDir string, dirStats *DirectoryStats) error {
	// Generate report in the output directory corresponding to the current directory
	var reportPath string
	if currentDir == "" {
		// Root directory
		reportPath = filepath.Join(config.OutputDir, "processing_report.html")
	} else {
		// Subdirectory - create corresponding path in output directory
		reportPath = filepath.Join(config.OutputDir, currentDir, "processing_report.html")
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
	dirTitle := fmt.Sprintf("Directory: %s", currentDir)
	
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
		
		// Adjust the file path to be relative to the report location
		// Calculate relative path from report location to file
		fileDir := filepath.Dir(actualFilePath)
		fileName := filepath.Base(actualFilePath)
		if fileDir == currentDir {
			// File is in the same directory as the report
			actualFilePath = fileName
		} else {
			// File is in a different directory, use relative path
			relPath, _ := filepath.Rel(currentDir, actualFilePath)
			actualFilePath = relPath
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
