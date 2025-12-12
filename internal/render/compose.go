package render

import (
	"image"
	"image/color"
	"image/draw"
)

// AppendImages appends two images horizontally
func AppendImages(img1, img2 image.Image) image.Image {
	if img1 == nil && img2 == nil {
		return nil
	}
	if img1 == nil {
		return img2
	}
	if img2 == nil {
		return img1
	}

	b1 := img1.Bounds()
	b2 := img2.Bounds()

	// Use the larger height
	height := b1.Dy()
	if b2.Dy() > height {
		height = b2.Dy()
	}

	// Total width is sum of both
	width := b1.Dx() + b2.Dx()

	// Create new image
	out := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with white
	draw.Draw(out, out.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Copy first image
	draw.Draw(out, image.Rect(0, 0, b1.Dx(), b1.Dy()), img1, b1.Min, draw.Over)

	// Copy second image
	draw.Draw(out, image.Rect(b1.Dx(), 0, width, b2.Dy()), img2, b2.Min, draw.Over)

	return out
}

// CreateCutMark creates a dashed cut line image
func CreateCutMark(printWidth int) image.Image {
	width := 9
	out := image.NewRGBA(image.Rect(0, 0, width, printWidth))

	// Fill with white
	draw.Draw(out, out.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Draw dashed line at x=5
	black := color.RGBA{0, 0, 0, 255}
	dashLen := 3
	gapLen := 3

	for y := 0; y < printWidth; {
		// Skip gap
		y += gapLen
		// Draw dash
		for i := 0; i < dashLen && y < printWidth; i++ {
			out.Set(5, y, black)
			y++
		}
	}

	return out
}

// CreatePadding creates blank padding image
func CreatePadding(printWidth, length int) image.Image {
	if length < 1 {
		length = 1
	}
	if length > 256 {
		length = 256
	}

	out := image.NewRGBA(image.Rect(0, 0, length, printWidth))

	// Fill with white
	draw.Draw(out, out.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	return out
}

// ImageToRaster converts an image column to a raster line for printing
// The image is scanned column by column (x increases), and for each column,
// pixels are read from bottom to top (y decreases)
func ImageToRaster(img image.Image, x int, maxPixels int, offset int) []byte {
	bounds := img.Bounds()
	height := bounds.Dy()

	// Determine which color is "dark" (ink) - check corners
	r0, g0, b0, _ := img.At(bounds.Min.X, bounds.Min.Y).RGBA()
	r1, g1, b1, _ := img.At(bounds.Min.X, bounds.Max.Y-1).RGBA()

	// Assume the more common corner color is background
	dark0 := IsDark(r0, g0, b0)
	dark1 := IsDark(r1, g1, b1)

	// Use simple heuristic: darker pixels are ink
	inkIsDark := true
	if !dark0 && !dark1 {
		inkIsDark = true // Normal case: dark ink on light background
	} else if dark0 && dark1 {
		inkIsDark = false // Inverted: light ink on dark background
	}

	// Create raster line
	rasterBytes := maxPixels / 8
	raster := make([]byte, rasterBytes)

	for y := 0; y < height; y++ {
		r, g, b, _ := img.At(x, height-1-y).RGBA()
		pixelIsDark := IsDark(r, g, b)

		if (inkIsDark && pixelIsDark) || (!inkIsDark && !pixelIsDark) {
			// This pixel should be printed (ink)
			pixelPos := offset + y
			if pixelPos >= 0 && pixelPos < maxPixels {
				byteIdx := rasterBytes - 1 - (pixelPos / 8)
				bitIdx := pixelPos % 8
				if byteIdx >= 0 && byteIdx < rasterBytes {
					raster[byteIdx] |= 1 << bitIdx
				}
			}
		}
	}

	return raster
}
