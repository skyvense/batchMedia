//go:build noheif
// +build noheif

package main

import (
	"fmt"
	"image"
)

// decodeHEIC returns an error when HEIC support is disabled
func decodeHEIC(data []byte) (image.Image, error) {
	return nil, fmt.Errorf("HEIC support is disabled in this build")
}

// extractHEICExifData returns an error when HEIC support is disabled
func extractHEICExifData(data []byte) ([]byte, error) {
	return nil, fmt.Errorf("HEIC support is disabled in this build")
}

// isHEICSupported returns false when HEIC support is disabled
func isHEICSupported() bool {
	return false
}