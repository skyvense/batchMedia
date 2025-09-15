package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jdeng/goheif"
	"github.com/nfnt/resize"
	"github.com/rwcarlsen/goexif/exif"
)

// processImage processes a single image file
func processImage(inputPath, outputPath, relPath string, info os.FileInfo, dirStats *DirectoryStats) error {
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
	// Note: PNG files typically don't contain EXIF data, so no extraction needed

	// Decode image based on file extension
	var img image.Image
	if ext == ".heic" {
		// Decode HEIC image
		img, err = goheif.Decode(bytes.NewReader(fileData))
		if err != nil {
			return fmt.Errorf("failed to decode HEIC image: %v", err)
		}
	} else if ext == ".png" {
		// Decode PNG image
		img, err = png.Decode(bytes.NewReader(fileData))
		if err != nil {
			return fmt.Errorf("failed to decode PNG image: %v", err)
		}
	} else {
		// Decode JPEG image
		img, err = jpeg.Decode(bytes.NewReader(fileData))
		if err != nil {
			return fmt.Errorf("failed to decode JPEG image: %v", err)
		}
	}

	// Apply EXIF orientation correction if needed
	img = applyEXIFOrientation(img, fileData)

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
		dirStats.SkippedImages++
		dirStats.TotalOutputSize += info.Size()

		// Record file info
		fileInfo := FileInfo{
			Path:             relPath,
			Type:             "skipped",
			InputSize:        info.Size(),
			OutputSize:       info.Size(),
			CompressionRatio: 1.0,
		}
		stats.Files = append(stats.Files, fileInfo)
		dirStats.Files = append(dirStats.Files, fileInfo)

		// Copy original file without processing
		return copyFile(inputPath, outputPath, info)
	}

	// Calculate new dimensions
	newWidth, newHeight := calculateNewSize(originalWidth, originalHeight)

	// Resize image
	resizedImg := resizeImage(img, newWidth, newHeight)

	// Encode image to buffer
	// Note: Currently all images are encoded as JPEG for compatibility
	// HEIC encoding is not supported by the goheif library
	var buf bytes.Buffer
	options := &jpeg.Options{Quality: 85} // Higher quality for better compatibility
	if err := jpeg.Encode(&buf, resizedImg, options); err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	// Get final image data and insert EXIF if available
	finalImageData := buf.Bytes()
	if exifData != nil {
		// Clear orientation tag from EXIF data since we've already applied the correction
		cleanedExifData := clearOrientationTag(exifData)
		finalImageData = insertEXIFCorrectly(finalImageData, cleanedExifData)
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
	dirStats.ProcessedImages++
	dirStats.TotalOutputSize += outputSize

	// Calculate compression ratio
	compressionRatio := float64(outputSize) / float64(info.Size())

	// Record file info
	fileInfo := FileInfo{
		Path:             relPath,
		Type:             "processed",
		InputSize:        info.Size(),
		OutputSize:       outputSize,
		CompressionRatio: compressionRatio,
	}
	stats.Files = append(stats.Files, fileInfo)
	dirStats.Files = append(dirStats.Files, fileInfo)

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

// insertEXIFCorrectly inserts EXIF data into JPEG file with proper APP1 segment structure
// applyEXIFOrientation applies EXIF orientation correction to the image
func applyEXIFOrientation(img image.Image, fileData []byte) image.Image {
	// Try to extract EXIF orientation
	reader := bytes.NewReader(fileData)
	x, err := exif.Decode(reader)
	if err != nil {
		// No EXIF data or unable to decode, return original image
		return img
	}

	// Get orientation tag
	orientationTag, err := x.Get(exif.Orientation)
	if err != nil {
		// No orientation tag, return original image
		return img
	}

	// Get orientation value
	orientation, err := orientationTag.Int(0)
	if err != nil {
		return img
	}

	// Apply transformation based on orientation value
	switch orientation {
	case 1:
		// Normal orientation, no transformation needed
		return img
	case 2:
		// Flip horizontal
		return flipHorizontal(img)
	case 3:
		// Rotate 180 degrees
		return rotate180(img)
	case 4:
		// Flip vertical
		return flipVertical(img)
	case 5:
		// Rotate 90 degrees clockwise and flip horizontal
		return flipHorizontal(rotate90CW(img))
	case 6:
		// Rotate 90 degrees clockwise
		return rotate90CW(img)
	case 7:
		// Rotate 90 degrees counter-clockwise and flip horizontal
		return flipHorizontal(rotate90CCW(img))
	case 8:
		// Rotate 90 degrees counter-clockwise
		return rotate90CCW(img)
	default:
		// Unknown orientation, return original
		return img
	}
}

// rotate90CW rotates image 90 degrees clockwise
func rotate90CW(src image.Image) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(h-1-y, x, src.At(x, y))
		}
	}
	return dst
}

// rotate90CCW rotates image 90 degrees counter-clockwise
func rotate90CCW(src image.Image) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(y, w-1-x, src.At(x, y))
		}
	}
	return dst
}

// rotate180 rotates image 180 degrees
func rotate180(src image.Image) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(w-1-x, h-1-y, src.At(x, y))
		}
	}
	return dst
}

// flipHorizontal flips image horizontally
func flipHorizontal(src image.Image) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(w-1-x, y, src.At(x, y))
		}
	}
	return dst
}

// flipVertical flips image vertically
func flipVertical(src image.Image) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(x, h-1-y, src.At(x, y))
		}
	}
	return dst
}

// clearOrientationTag removes the orientation tag from EXIF data
func clearOrientationTag(exifData []byte) []byte {
	// For simplicity, we'll create a new EXIF segment with orientation set to 1 (normal)
	// This is a basic implementation that works for most cases
	if len(exifData) < 10 {
		return exifData
	}

	// Make a copy of the EXIF data
	cleanedData := make([]byte, len(exifData))
	copy(cleanedData, exifData)

	// Look for orientation tag (0x0112) in the EXIF data
	// This is a simplified approach - in a full implementation, you'd parse the TIFF structure
	for i := 0; i < len(cleanedData)-4; i++ {
		// Look for orientation tag (0x0112 in big-endian or 0x1201 in little-endian)
		if (cleanedData[i] == 0x01 && cleanedData[i+1] == 0x12) || 
		   (cleanedData[i] == 0x12 && cleanedData[i+1] == 0x01) {
			// Found potential orientation tag, set value to 1 (normal orientation)
			if i+8 < len(cleanedData) {
				// Set the value to 1 (normal orientation)
				cleanedData[i+6] = 0x00
				cleanedData[i+7] = 0x01
				break
			}
		}
	}

	return cleanedData
}

func insertEXIFCorrectly(jpegData, exifData []byte) []byte {
	if len(jpegData) < 4 || jpegData[0] != 0xFF || jpegData[1] != 0xD8 {
		return jpegData // Not a valid JPEG file
	}

	// Create APP1 segment with EXIF data
	// APP1 marker (0xFFE1) + length (2 bytes) + "Exif\x00\x00" + EXIF data
	exifHeader := []byte{0xFF, 0xE1} // APP1 marker
	exifIdentifier := []byte("Exif\x00\x00")
	segmentLength := uint16(2 + len(exifIdentifier) + len(exifData))

	// Build complete APP1 segment
	app1Segment := make([]byte, 0, 4+len(exifIdentifier)+len(exifData))
	app1Segment = append(app1Segment, exifHeader...)
	app1Segment = append(app1Segment, byte(segmentLength>>8), byte(segmentLength&0xFF)) // Big-endian length
	app1Segment = append(app1Segment, exifIdentifier...)
	app1Segment = append(app1Segment, exifData...)

	// Insert APP1 segment after SOI marker (0xFFD8)
	result := make([]byte, 0, len(jpegData)+len(app1Segment))
	result = append(result, jpegData[0:2]...) // SOI marker
	result = append(result, app1Segment...)   // APP1 segment with EXIF
	result = append(result, jpegData[2:]...)  // Rest of JPEG data

	return result
}