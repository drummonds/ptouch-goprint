package render

import (
	"fmt"
	"image"
	"image/png"
	"os"
)

// LoadImage loads a PNG image from a file
func LoadImage(path string) (image.Image, error) {
	var f *os.File
	var err error

	if path == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open image: %w", err)
		}
		defer f.Close()
	}

	img, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	return img, nil
}

// SavePNG saves an image to a PNG file
func SavePNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// ImageSize returns the width and height of an image
func ImageSize(img image.Image) (width, height int) {
	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy()
}

// IsDark determines if a color is dark (closer to black than white)
func IsDark(r, g, b uint32) bool {
	// Convert from 16-bit to 8-bit color
	r8 := r >> 8
	g8 := g >> 8
	b8 := b >> 8
	// Simple luminance check
	return (r8 + g8 + b8) < (128 * 3)
}
