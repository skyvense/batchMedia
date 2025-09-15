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

// processImage processes a single image file
func processImage(inputPath, outputPath string, info os.FileInfo) error {
	// Read entire file into memory
	fileData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}

	// Extract EXIF information
	exifData, err := extractEXIF(fileData)
	if err != nil {
		// EXIF extraction failure is not fatal, continue processing
		fmt.Printf("Warning: unable to extract EXIF information: %v\n", err)
	}

	// Decode image
	img, err := jpeg.Decode(bytes.NewReader(fileData))
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	// Get original dimensions
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

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

	fmt.Printf("Processing completed: %s (%dx%d -> %dx%d)\n", inputPath, originalWidth, originalHeight, newWidth, newHeight)
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

	if config.Size > 0 {
		// Scale by ratio
		newWidth := int(float64(originalWidth) * config.Size)
		newHeight := int(float64(originalHeight) * config.Size)
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

// extractEXIF extracts EXIF information from JPEG file data
func extractEXIF(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)
	
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