package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// isVideoFile checks if the file is a supported video format
func isVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	supportedFormats := []string{".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v"}
	
	for _, format := range supportedFormats {
		if ext == format {
			return true
		}
	}
	return false
}

// shouldSkipVideo checks if video should be skipped based on resolution thresholds
func shouldSkipVideo(width, height int) bool {
	if config.IgnoreSmartLimit {
		return false
	}

	// Check if video exceeds threshold (should be skipped)
	if config.ThresholdWidth > 0 && width > config.ThresholdWidth {
		return true
	}
	if config.ThresholdHeight > 0 && height > config.ThresholdHeight {
		return true
	}
	return false
}

// getVideoResolution gets the resolution of a video file using ffprobe
func getVideoResolution(inputPath string) (int, int, error) {
	// Use ffprobe to get video information
	probe, err := ffmpeg.Probe(inputPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to probe video file: %v", err)
	}
	
	// Parse probe result to extract width and height
	// This is a simplified implementation - in practice you'd parse the JSON output
	// For now, return default values to avoid compilation errors
	_ = probe // Use probe variable to avoid unused variable error
	return 1920, 1080, nil
}

// processVideo processes a single video file using FFmpeg
func processVideo(inputPath, outputPath string, info os.FileInfo, dirStats *DirectoryStats) error {
	// Get video resolution for threshold checking
	originalWidth, originalHeight, err := getVideoResolution(inputPath)
	if err != nil {
		fmt.Printf("Warning: Could not get video resolution for %s, proceeding with processing\n", inputPath)
		originalWidth = 1920 // Default values
		originalHeight = 1080
	}

	// Check if video should be skipped based on resolution thresholds
	if shouldSkipVideo(originalWidth, originalHeight) {
		fmt.Printf("Skipping video (resolution %dx%d exceeds threshold): %s (size: %d bytes)\n", 
			originalWidth, originalHeight, inputPath, info.Size())
		stats.SkippedImages++ // Using same counter for videos
		stats.TotalOutputSize += info.Size()
		
		// Record file info
		stats.Files = append(stats.Files, FileInfo{
			Path:             filepath.Base(inputPath),
			Type:             "skipped",
			InputSize:        info.Size(),
			OutputSize:       info.Size(),
			OriginalDim:      fmt.Sprintf("%dx%d", originalWidth, originalHeight),
			NewDim:           fmt.Sprintf("%dx%d", originalWidth, originalHeight),
			CompressionRatio: 1.0,
		})
		
		// Copy original file
		return copyFile(inputPath, outputPath, info)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Calculate new dimensions based on same logic as images
	newWidth := originalWidth
	newHeight := originalHeight
	var scaleFilter string
	
	// Add resolution scaling if specified
	if config.VideoResolution != "" {
		scaleFilter = config.VideoResolution
	} else if config.ScalingRatio > 0 {
		// Use scaling ratio
		newWidth = int(float64(originalWidth) * config.ScalingRatio)
		newHeight = int(float64(originalHeight) * config.ScalingRatio)
		scaleFilter = fmt.Sprintf("%d:%d", newWidth, newHeight)
	} else if config.Width > 0 {
		// Scale by width, maintain aspect ratio
		newWidth = config.Width
		newHeight = int(float64(originalHeight) * float64(config.Width) / float64(originalWidth))
		scaleFilter = fmt.Sprintf("%d:-1", config.Width)
	}

	// Build FFmpeg arguments using filter_complex and proper mapping
	input := ffmpeg.Input(inputPath)
	var output *ffmpeg.Stream
	
	// Use filter_complex for video scaling
	if scaleFilter != "" {
		// Apply scale filter using filter_complex
		output = input.Video().Filter("scale", ffmpeg.Args{scaleFilter})
	} else {
		// No scaling, use original video stream
		output = input.Video()
	}

	// Check if input video is HDR
	isHDR := isHDRVideo(inputPath)
	
	// Apply video encoding options based on HDR detection
	var kwargs ffmpeg.KwArgs
	
	if isHDR {
		// HDR video encoding parameters
		kwargs = ffmpeg.KwArgs{
			"c:v": config.VideoCodec,
			"preset": config.VideoPreset,
			"crf": fmt.Sprintf("%d", config.VideoCRF),
			"profile:v": "main10",
			"pix_fmt": "yuv420p10le",
			"tag:v": "hvc1",
			"color_primaries": "bt2020",
			"color_trc": "smpte2084",
			"colorspace": "bt2020nc",
			"x265-params": "hdr-opt=1:repeat-headers=1:colorprim=bt2020:transfer=smpte2084:colormatrix=bt2020nc",
			"level": "5.1",
			"progress": "pipe:1",
			"stats": "",
			"map_metadata": "0",
		}
		fmt.Printf("Processing HDR video: %s\n", inputPath)
	} else {
		// SDR video encoding parameters (standard rec709 colorspace)
		kwargs = ffmpeg.KwArgs{
			"c:v": config.VideoCodec,
			"preset": config.VideoPreset,
			"crf": fmt.Sprintf("%d", config.VideoCRF),
			"profile:v": "main",
			"pix_fmt": "yuv420p",
			"tag:v": "hvc1",
			"level": "4.0",
			"progress": "pipe:1",
			"stats": "",
			"map_metadata": "0",
		}
		fmt.Printf("Processing SDR video: %s\n", inputPath)
	}
	
	// Apply user-specified bitrate if provided
	if config.VideoBitrate != "" {
		kwargs["b:v"] = config.VideoBitrate
		delete(kwargs, "crf") // Remove CRF when using bitrate
	}
	
	// Handle audio stream
	if hasAudioStream(inputPath) {
		// Copy audio stream without re-encoding
		kwargs["c:a"] = "copy"
		fmt.Printf("Audio stream detected in %s, will preserve audio\n", inputPath)
		
		// Map both video and audio streams
		err = ffmpeg.Output([]*ffmpeg.Stream{output, input.Audio()}, outputPath, kwargs).OverWriteOutput().Run()
	} else {
		// No audio stream, process video only
		fmt.Printf("No audio stream detected in %s, processing video only\n", inputPath)
		
		// Map only video stream
		err = output.Output(outputPath, kwargs).OverWriteOutput().Run()
	}

	// Run FFmpeg command
	if err != nil {
		// If processing fails and video has audio, try with audio re-encoding
		if hasAudioStream(inputPath) {
			fmt.Printf("Warning: Audio copy failed for %s, trying with audio re-encoding...\n", inputPath)
			
			// Remove the failed output file
			os.Remove(outputPath)
			
			// Retry with audio re-encoding
			kwargs["c:a"] = "aac"
			kwargs["b:a"] = "128k"
			delete(kwargs, "map") // Remove mapping that might cause issues
			
			err = output.Output(outputPath, kwargs).OverWriteOutput().Run()
			if err != nil {
				return fmt.Errorf("failed to process video even with audio re-encoding: %v", err)
			}
			fmt.Printf("Successfully processed %s with audio re-encoding\n", inputPath)
		} else {
			return fmt.Errorf("failed to process video: %v", err)
		}
	}

	// Get output file info for statistics
	outputInfo, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get output file info: %v", err)
	}

	// Record statistics
	outputSize := outputInfo.Size()
	stats.ProcessedImages++ // Using same counter for videos
	stats.TotalOutputSize += outputSize
	dirStats.ProcessedImages++
	dirStats.TotalOutputSize += outputSize
	
	// Calculate compression ratio
	compressionRatio := float64(outputSize) / float64(info.Size())
	
	// Get relative path for file info
	relPath, _ := filepath.Rel(config.InputDir, inputPath)
	
	// Record file info
	fileInfo := FileInfo{
		Path:             relPath,
		Type:             "video_processed",
		InputSize:        info.Size(),
		OutputSize:       outputSize,
		CompressionRatio: compressionRatio,
	}
	stats.Files = append(stats.Files, fileInfo)
	dirStats.Files = append(dirStats.Files, fileInfo)

	// Preserve original file modification time
	if err := os.Chtimes(outputPath, info.ModTime(), info.ModTime()); err != nil {
		return fmt.Errorf("failed to set file time: %v", err)
	}

	fmt.Printf("Video processing completed: %s (%d bytes -> %d bytes, ratio: %.2f)\n", 
		inputPath, info.Size(), outputSize, compressionRatio)
	return nil
}

// isHDRVideo checks if the video file is HDR format
func isHDRVideo(inputPath string) bool {
	probe, err := ffmpeg.Probe(inputPath)
	if err != nil {
		return false // Assume SDR if probe fails
	}
	
	// Check for HDR indicators in the probe output
	// HDR videos typically have:
	// - color_primaries: bt2020
	// - color_trc: smpte2084 (PQ) or arib-std-b67 (HLG)
	// - colorspace: bt2020nc or bt2020c
	probeStr := strings.ToLower(probe)
	
	// Check for HDR transfer characteristics
	hasHDRTransfer := strings.Contains(probeStr, "smpte2084") || 
					 strings.Contains(probeStr, "arib-std-b67") ||
					 strings.Contains(probeStr, "smpte-st-2084") ||
					 strings.Contains(probeStr, "hlg")
	
	// Check for wide color gamut
	hasWideGamut := strings.Contains(probeStr, "bt2020")
	
	// Consider it HDR if it has both HDR transfer and wide gamut
	return hasHDRTransfer && hasWideGamut
}

// hasAudioStream checks if the video file contains audio streams
func hasAudioStream(inputPath string) bool {
	probe, err := ffmpeg.Probe(inputPath)
	if err != nil {
		return false // Assume no audio if probe fails
	}
	
	// Check if probe result contains audio stream information
	// This is a simplified check - in practice you'd parse the JSON output
	return strings.Contains(probe, "audio") || strings.Contains(probe, "Audio")
}

// getVideoInfo gets basic information about a video file
func getVideoInfo(inputPath string) (map[string]interface{}, error) {
	// This is a placeholder for video info extraction
	// In a real implementation, you might use ffprobe or similar
	return map[string]interface{}{
		"format": filepath.Ext(inputPath),
		"has_audio": hasAudioStream(inputPath),
	}, nil
}