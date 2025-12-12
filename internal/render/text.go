package render

import (
	"fmt"
	"image"

	"github.com/fogleman/gg"
)

// AlignType defines text alignment options
type AlignType int

const (
	AlignLeft AlignType = iota
	AlignCenter
	AlignRight
)

// MaxLines is the maximum number of text lines supported
const MaxLines = 4

// TextRenderer handles text-to-image rendering
type TextRenderer struct {
	FontPath string
	FontSize float64
	Debug    bool
}

// NewTextRenderer creates a new text renderer with the given font
func NewTextRenderer(fontPath string) *TextRenderer {
	return &TextRenderer{
		FontPath: fontPath,
		FontSize: 0, // auto-detect
	}
}

// findFontSize determines the optimal font size for the given height
func (r *TextRenderer) findFontSize(maxHeight int, text string) (float64, error) {
	var bestSize float64

	for size := 4.0; size < 500; size++ {
		dc := gg.NewContext(1, 1)
		if err := dc.LoadFontFace(r.FontPath, size); err != nil {
			return 0, fmt.Errorf("failed to load font: %w", err)
		}

		_, h := dc.MeasureString(text)
		if h <= float64(maxHeight) {
			bestSize = size
		} else {
			break
		}
	}

	if bestSize == 0 {
		return 0, fmt.Errorf("could not find suitable font size")
	}

	return bestSize, nil
}

// getTextWidth returns the width of text at a given font size
func (r *TextRenderer) getTextWidth(text string, fontSize float64) (float64, error) {
	dc := gg.NewContext(1, 1)
	if err := dc.LoadFontFace(r.FontPath, fontSize); err != nil {
		return 0, err
	}
	w, _ := dc.MeasureString(text)
	return w, nil
}

// RenderText renders multiple lines of text to an image
func (r *TextRenderer) RenderText(lines []string, printWidth int, align AlignType) (image.Image, error) {
	numLines := len(lines)
	if numLines == 0 {
		return nil, fmt.Errorf("no text lines provided")
	}
	if numLines > MaxLines {
		numLines = MaxLines
	}

	// Determine font size (auto or manual)
	fontSize := r.FontSize
	if fontSize == 0 {
		// Find optimal font size that fits all lines
		heightPerLine := printWidth / numLines
		for _, line := range lines[:numLines] {
			if line == "" {
				continue
			}
			size, err := r.findFontSize(heightPerLine, line)
			if err != nil {
				return nil, err
			}
			if fontSize == 0 || size < fontSize {
				fontSize = size
			}
		}
		if r.Debug {
			fmt.Printf("choosing font size=%.0f\n", fontSize)
		}
	} else if r.Debug {
		fmt.Printf("setting font size=%.0f\n", fontSize)
	}

	// Calculate total width needed
	var maxWidth float64
	for _, line := range lines[:numLines] {
		w, err := r.getTextWidth(line, fontSize)
		if err != nil {
			return nil, err
		}
		if w > maxWidth {
			maxWidth = w
		}
	}

	// Create the image context
	imgWidth := int(maxWidth) + 2 // small padding
	dc := gg.NewContext(imgWidth, printWidth)
	dc.SetRGB(1, 1, 1) // white background
	dc.Clear()
	dc.SetRGB(0, 0, 0) // black text

	if err := dc.LoadFontFace(r.FontPath, fontSize); err != nil {
		return nil, fmt.Errorf("failed to load font: %w", err)
	}

	// Calculate line height and spacing
	_, lineHeight := dc.MeasureString("Xy") // Reference height
	totalTextHeight := lineHeight * float64(numLines)
	unusedPx := float64(printWidth) - totalTextHeight
	spacingPerLine := unusedPx / float64(numLines)

	// Render each line
	for i, line := range lines[:numLines] {
		if line == "" {
			continue
		}

		// Calculate vertical position
		y := float64(i)*lineHeight + lineHeight + spacingPerLine/2 + float64(i)*spacingPerLine/float64(numLines)

		// Calculate horizontal position based on alignment
		var x float64
		lineWidth, _ := dc.MeasureString(line)

		switch align {
		case AlignCenter:
			x = (float64(imgWidth) - lineWidth) / 2
		case AlignRight:
			x = float64(imgWidth) - lineWidth
		default: // AlignLeft
			x = 0
		}

		if r.Debug {
			fmt.Printf("debug: line %d pos=%.0f x=%.0f\n", i+1, y, x)
		}

		dc.DrawString(line, x, y)
	}

	return dc.Image(), nil
}
