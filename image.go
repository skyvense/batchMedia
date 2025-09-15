package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jdeng/goheif"
	"github.com/nfnt/resize"
	"github.com/rwcarlsen/goexif/exif"
)

// processImage processes a single image file
func processImage(inputPath, outputPath string, info os.FileInfo) error {
	// Read entire file into memory
	fileData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}

	// Extract EXIF information
	var exifData []byte
	ext := strings.ToLower(filepath.Ext(inputPath))
	if ext == ".jpg" || ext == ".jpeg" {
		// Extract EXIF from JPEG files
		var err error
		exifData, err = extractEXIF(fileData)
		if err != nil {
			// EXIF extraction failure is not fatal, continue processing
			fmt.Printf("Warning: unable to extract EXIF information from %s: %v\n", inputPath, err)
		}
	} else if ext == ".heic" {
		// Extract EXIF from HEIC files
		var err error
		exifData, err = extractHEICExif(fileData)
		if err != nil {
			// EXIF extraction failure is not fatal, continue processing
			fmt.Printf("Warning: unable to extract EXIF information from %s: %v\n", inputPath, err)
		}
	}

	// Decode image based on file extension
	var img image.Image
	if ext == ".heic" {
		// Decode HEIC image
		img, err = goheif.Decode(bytes.NewReader(fileData))
		if err != nil {
			return fmt.Errorf("failed to decode HEIC image: %v", err)
		}
	} else {
		// Decode JPEG image
		img, err = jpeg.Decode(bytes.NewReader(fileData))
		if err != nil {
			return fmt.Errorf("failed to decode JPEG image: %v", err)
		}
	}

	// Get original dimensions
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Check if image should be skipped based on resolution thresholds
	if shouldSkipImage(originalWidth, originalHeight) {
		fmt.Printf("Skipping %s: resolution %dx%d is outside threshold range (size: %d bytes)\n", inputPath, originalWidth, originalHeight, info.Size())
		
		// Record statistics for skipped image
		stats.SkippedImages++
		stats.TotalOutputSize += info.Size()
		
		// Record file info
		stats.Files = append(stats.Files, FileInfo{
			Path:             filepath.Base(inputPath),
			Type:             "skipped",
			InputSize:        info.Size(),
			OutputSize:       info.Size(),
			CompressionRatio: 1.0,
		})
		
		// Copy original file without processing
		return copyFile(inputPath, outputPath, info)
	}

	// Calculate new dimensions
	newWidth, newHeight := calculateNewSize(originalWidth, originalHeight)

	// Resize image
	resizedImg := resizeImage(img, newWidth, newHeight)

	// Encode image to buffer
	var buf bytes.Buffer
	options := &jpeg.Options{Quality: 90}
	if err := jpeg.Encode(&buf, resizedImg, options); err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	// If EXIF data exists, try to insert it into the new image
	finalImageData := buf.Bytes()
	if exifData != nil {
		finalImageData = insertEXIF(finalImageData, exifData)
	}

	// Write output file
	if err := os.WriteFile(outputPath, finalImageData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	// Preserve original file modification time
	if err := os.Chtimes(outputPath, info.ModTime(), info.ModTime()); err != nil {
		return fmt.Errorf("failed to set file time: %v", err)
	}

	// Record statistics
	outputSize := int64(len(finalImageData))
	stats.ProcessedImages++
	stats.TotalOutputSize += outputSize
	
	// Calculate compression ratio
	compressionRatio := float64(outputSize) / float64(info.Size())
	
	// Record file info
	stats.Files = append(stats.Files, FileInfo{
		Path:             filepath.Base(inputPath),
		Type:             "processed",
		InputSize:        info.Size(),
		OutputSize:       outputSize,
		CompressionRatio: compressionRatio,
	})

	fmt.Printf("Processing completed: %s (%dx%d -> %dx%d, %d bytes -> %d bytes, ratio: %.2f)\n", 
		inputPath, originalWidth, originalHeight, newWidth, newHeight, info.Size(), outputSize, compressionRatio)
	return nil
}

// calculateNewSize calculates new image dimensions based on configuration
func calculateNewSize(originalWidth, originalHeight int) (int, int) {
	if config.Width > 0 {
		// Scale by width, maintain aspect ratio
		ratio := float64(config.Width) / float64(originalWidth)
		newHeight := int(float64(originalHeight) * ratio)
		return config.Width, newHeight
	}

	if config.ScalingRatio > 0 {
		// Scale by ratio
		newWidth := int(float64(originalWidth) * config.ScalingRatio)
		newHeight := int(float64(originalHeight) * config.ScalingRatio)
		return newWidth, newHeight
	}

	// Default return original dimensions
	return originalWidth, originalHeight
}

// resizeImage resizes image using high-quality algorithm
func resizeImage(src image.Image, newWidth, newHeight int) image.Image {
	// Use Lanczos3 algorithm for high-quality scaling
	// Lanczos3 provides the best image quality, especially suitable for photo scaling
	return resize.Resize(uint(newWidth), uint(newHeight), src, resize.Lanczos3)
}

// shouldSkipImage checks if image should be skipped based on resolution thresholds
func shouldSkipImage(width, height int) bool {
	// Apply threshold logic based on scaling type
	if config.ScalingRatio > 1.0 {
		// Upscaling: skip images above threshold (too large to upscale)
		if config.ThresholdWidth > 0 && width > config.ThresholdWidth {
			return true
		}
		if config.ThresholdHeight > 0 && height > config.ThresholdHeight {
			return true
		}
	} else if config.ScalingRatio < 1.0 {
		// Downscaling: skip images below threshold (too small to downscale)
		if config.ThresholdWidth > 0 && width < config.ThresholdWidth {
			return true
		}
		if config.ThresholdHeight > 0 && height < config.ThresholdHeight {
			return true
		}
	}

	return false
}

// copyFile copies a file from source to destination while preserving file info
func copyFile(src, dst string, info os.FileInfo) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	// Preserve file modification time
	return os.Chtimes(dst, info.ModTime(), info.ModTime())
}

// extractEXIF extracts EXIF information from image file data
func extractEXIF(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)
	
	// Check if it's a JPEG file by looking at the header
	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		// Not a JPEG file (likely HEIC), skip EXIF extraction
		return nil, fmt.Errorf("EXIF extraction only supported for JPEG files")
	}
	
	// Find EXIF data segment
	_, err := exif.Decode(reader)
	if err != nil {
		return nil, err
	}
	
	// Find APP1 segment (EXIF data)
	reader.Seek(0, 0)
	buf := make([]byte, 2)
	
	// Check JPEG file header
	if _, err := reader.Read(buf); err != nil {
		return nil, err
	}
	if buf[0] != 0xFF || buf[1] != 0xD8 {
		return nil, fmt.Errorf("not a valid JPEG file")
	}
	
	// Find APP1 segment
	for {
		if _, err := reader.Read(buf); err != nil {
			return nil, err
		}
		
		if buf[0] != 0xFF {
			continue
		}
		
		// Found APP1 segment
		if buf[1] == 0xE1 {
			// Read segment length
			if _, err := reader.Read(buf); err != nil {
				return nil, err
			}
			length := int(buf[0])<<8 | int(buf[1])
			
			// Read entire APP1 segment
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
		
		// If it's another segment, skip it
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
	
	return nil, fmt.Errorf("EXIF data not found")
}

// extractHEICExif extracts EXIF information from HEIC file data
func extractHEICExif(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)
	
	// Use goheif.ExtractExif to extract EXIF from HEIC file
	exifData, err := goheif.ExtractExif(reader)
	if err != nil {
		return nil, err
	}
	
	return exifData, nil
}

// insertEXIF inserts EXIF data into JPEG file
func insertEXIF(jpegData, exifData []byte) []byte {
	if len(jpegData) < 4 || jpegData[0] != 0xFF || jpegData[1] != 0xD8 {
		return jpegData // Not a valid JPEG file
	}
	
	// Create new JPEG data
	result := make([]byte, 0, len(jpegData)+len(exifData))
	
	// Add JPEG file header
	result = append(result, jpegData[0:2]...)
	
	// Add EXIF data
	result = append(result, exifData...)
	
	// Add remaining JPEG data (skip file header)
	result = append(result, jpegData[2:]...)
	
	return result
}