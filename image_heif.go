//go:build !noheif
// +build !noheif

package main

import (
	"bytes"
	"image"

	"github.com/jdeng/goheif"
)

// decodeHEIC decodes HEIC image using goheif library
func decodeHEIC(data []byte) (image.Image, error) {
	return goheif.Decode(bytes.NewReader(data))
}

// extractHEICExifData extracts EXIF information from HEIC file data
func extractHEICExifData(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)

	// Use goheif.ExtractExif to extract EXIF from HEIC file
	exifData, err := goheif.ExtractExif(reader)
	if err != nil {
		return nil, err
	}

	return exifData, nil
}

// isHEICSupported returns true if HEIC support is available
func isHEICSupported() bool {
	return true
}