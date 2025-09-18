package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

// createTestImage creates a test image with specified dimensions and color
func createTestImage(width, height int, bgColor color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)
	
	// Add some pattern to make it visually distinct
	for y := 0; y < height; y += 50 {
		for x := 0; x < width; x += 50 {
			if (x/50+y/50)%2 == 0 {
				img.Set(x, y, color.RGBA{255, 255, 255, 255}) // White
			} else {
				img.Set(x, y, color.RGBA{0, 0, 0, 255}) // Black
			}
		}
	}
	
	return img
}

// saveJPEG saves image as JPEG
func saveJPEG(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
}

// savePNG saves image as PNG
func savePNG(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	return png.Encode(file, img)
}

func main() {
	// Create test directories
	dirs := []string{
		"input/images",
		"input/videos", 
		"input/mixed",
	}
	
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}
	
	// Test images with different sizes and formats
	testImages := []struct {
		name   string
		width  int
		height int
		color  color.RGBA
		format string
	}{
		// Large images (above threshold)
		{"large_4k", 3840, 2160, color.RGBA{255, 0, 0, 255}, "jpg"},
		{"large_6k", 6000, 4000, color.RGBA{0, 255, 0, 255}, "png"},
		{"large_8k", 7680, 4320, color.RGBA{0, 0, 255, 255}, "jpg"},
		
		// Medium images (around threshold)
		{"medium_fhd", 1920, 1080, color.RGBA{255, 255, 0, 255}, "jpg"},
		{"medium_2k", 2560, 1440, color.RGBA{255, 0, 255, 255}, "png"},
		{"medium_wide", 2048, 1152, color.RGBA{0, 255, 255, 255}, "jpg"},
		
		// Small images (below threshold)
		{"small_hd", 1280, 720, color.RGBA{128, 128, 128, 255}, "jpg"},
		{"small_vga", 640, 480, color.RGBA{64, 64, 64, 255}, "png"},
		{"small_thumb", 320, 240, color.RGBA{192, 192, 192, 255}, "jpg"},
	}
	
	// Create images in different directories
	for _, testImg := range testImages {
		img := createTestImage(testImg.width, testImg.height, testImg.color)
		
		// Save in images directory
		imagesPath := filepath.Join("input/images", testImg.name+"."+testImg.format)
		if testImg.format == "jpg" {
			saveJPEG(img, imagesPath)
		} else {
			savePNG(img, imagesPath)
		}
		
		// Also save some in mixed directory
		if testImg.name == "large_4k" || testImg.name == "medium_fhd" || testImg.name == "small_hd" {
			mixedPath := filepath.Join("input/mixed", testImg.name+"."+testImg.format)
			if testImg.format == "jpg" {
				saveJPEG(img, mixedPath)
			} else {
				savePNG(img, mixedPath)
			}
		}
	}
	
	println("Test images created successfully!")
	println("Large images (>= 1920x1080):")
	println("  - large_4k.jpg (3840x2160)")
	println("  - large_6k.png (6000x4000)")
	println("  - large_8k.jpg (7680x4320)")
	println("")
	println("Medium images (around 1920x1080):")
	println("  - medium_fhd.jpg (1920x1080)")
	println("  - medium_2k.png (2560x1440)")
	println("  - medium_wide.jpg (2048x1152)")
	println("")
	println("Small images (< 1920x1080):")
	println("  - small_hd.jpg (1280x720)")
	println("  - small_vga.png (640x480)")
	println("  - small_thumb.jpg (320x240)")
}