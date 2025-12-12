package device

// TapeInfo describes tape width specifications
type TapeInfo struct {
	WidthMM uint8   // Tape width in mm
	WidthPx uint16  // Print area width in pixels (180 DPI)
	Margins float64 // Default tape margins in mm
}

// TapeWidths maps tape width (mm) to print area (px) at 180 DPI
var TapeWidths = []TapeInfo{
	{4, 24, 0.5},   // 3.5 mm tape
	{6, 32, 1.0},   // 6 mm tape
	{9, 52, 1.0},   // 9 mm tape
	{12, 76, 2.0},  // 12 mm tape
	{18, 120, 3.0}, // 18 mm tape
	{21, 124, 3.0}, // 21 mm tape
	{24, 128, 3.0}, // 24 mm tape
	{36, 192, 4.5}, // 36 mm tape
}

// GetTapePixelWidth returns the pixel width for a given tape width in mm
func GetTapePixelWidth(widthMM uint8) uint16 {
	for _, t := range TapeWidths {
		if t.WidthMM == widthMM {
			return t.WidthPx
		}
	}
	return 0 // Unknown tape width
}

// MediaType returns human-readable media type name
func MediaType(code uint8) string {
	switch code {
	case 0x00:
		return "No media"
	case 0x01:
		return "Laminated tape"
	case 0x03:
		return "Non-laminated tape"
	case 0x04:
		return "Fabric tape"
	case 0x11:
		return "Heat-shrink tube"
	case 0x13:
		return "Flexi tape"
	case 0x14:
		return "Flexible ID tape"
	case 0x15:
		return "Satin tape"
	case 0x17:
		return "Heat-shrink tube"
	case 0xff:
		return "Incompatible tape"
	default:
		return "unknown"
	}
}

// TapeColor returns human-readable tape color name
func TapeColor(code uint8) string {
	switch code {
	case 0x01:
		return "White"
	case 0x02:
		return "Other"
	case 0x03:
		return "Clear"
	case 0x04:
		return "Red"
	case 0x05:
		return "Blue"
	case 0x06:
		return "Yellow"
	case 0x07:
		return "Green"
	case 0x08:
		return "Black"
	case 0x09:
		return "Clear"
	case 0x20:
		return "Matte White"
	case 0x21:
		return "Matte Clear"
	case 0x22:
		return "Matte Silver"
	case 0x23:
		return "Satin Gold"
	case 0x24:
		return "Satin Silver"
	case 0x30:
		return "Blue (TZe-5[345]5)"
	case 0x31:
		return "Red (TZe-435)"
	case 0x40:
		return "Fluorescent Orange"
	case 0x41:
		return "Fluorescent Yellow"
	case 0x50:
		return "Berry Pink (TZe-MQP35)"
	case 0x51:
		return "Light Gray (TZe-MQL35)"
	case 0x52:
		return "Lime Green (TZe-MQG35)"
	case 0x60:
		return "Yellow"
	case 0x61:
		return "Pink"
	case 0x62:
		return "Blue"
	case 0x70:
		return "Heat-shrink Tube"
	case 0x71:
		return "Heat-shrink Tube white"
	case 0x90:
		return "White(Flex. ID)"
	case 0x91:
		return "Yellow(Flex. ID)"
	case 0xf0:
		return "Cleaning"
	case 0xf1:
		return "Stencil"
	case 0xff:
		return "Incompatible"
	default:
		return "unknown"
	}
}

// TextColor returns human-readable text color name
func TextColor(code uint8) string {
	switch code {
	case 0x01:
		return "White"
	case 0x02:
		return "Other"
	case 0x04:
		return "Red"
	case 0x05:
		return "Blue"
	case 0x08:
		return "Black"
	case 0x0a:
		return "Gold"
	case 0x62:
		return "Blue(F)"
	case 0xf0:
		return "Cleaning"
	case 0xf1:
		return "Stencil"
	case 0xff:
		return "Incompatible"
	default:
		return "unknown"
	}
}
