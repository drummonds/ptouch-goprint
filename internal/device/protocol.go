package device

import "fmt"

// Init sends the initialization sequence: 100 null bytes + ESC @
func (d *Device) Init() error {
	cmd := make([]byte, 102)
	// First 100 bytes are already zero
	cmd[100] = 0x1b // ESC
	cmd[101] = 0x40 // @
	return d.Send(cmd)
}

// RasterStart selects graphics transfer mode (raster mode)
func (d *Device) RasterStart() error {
	if d.HasFlag(FlagP700Init) {
		// ESC i a 01 = switch mode (0=esc/p, 1=raster mode)
		cmd := []byte{0x1b, 0x69, 0x61, 0x01}
		return d.Send(cmd)
	}
	// ESC i R 01 = Select graphics transfer mode = Raster
	cmd := []byte{0x1b, 0x69, 0x52, 0x01}
	return d.Send(cmd)
}

// EnablePackbits enables PackBits compression mode
func (d *Device) EnablePackbits() error {
	// M 02 = enable packbits compression mode
	cmd := []byte{0x4d, 0x02}
	return d.Send(cmd)
}

// Print information flags
const (
	InfoFlagMediaType = 0x02 // n2 contains valid media type
	InfoFlagWidth     = 0x04 // n3 contains valid width
	InfoFlagLength    = 0x08 // n4 contains valid length
	InfoFlagQuality   = 0x40 // Quality priority
	InfoFlagRecovery  = 0x80 // Recovery mode
)

// InfoCmd sends the print information command for newer devices
func (d *Device) InfoCmd(sizeX int) error {
	// ESC i z {n1} {n2} {n3} {n4} {n5} {n6} {n7} {n8} {n9} {n10}
	cmd := []byte{0x1b, 0x69, 0x7a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	// n1: Valid flags (media type + width + length + recovery)
	cmd[3] = InfoFlagMediaType | InfoFlagWidth | InfoFlagLength | InfoFlagRecovery

	// n2: Media type (from status)
	cmd[4] = d.Status.MediaType

	// n3: Media width (mm)
	cmd[5] = d.Status.MediaWidth

	// n4: Media length (mm) - usually 0 for continuous tape
	cmd[6] = d.Status.MediaLen

	// n5-n8: Raster number (little-endian)
	cmd[7] = byte(sizeX & 0xff)
	cmd[8] = byte((sizeX >> 8) & 0xff)
	cmd[9] = byte((sizeX >> 16) & 0xff)
	cmd[10] = byte((sizeX >> 24) & 0xff)

	// n9: Page flag (0=first page, 1=subsequent pages)
	// For D460BT series, n9 is set to 2 to feed and stop properly
	if d.HasFlag(FlagD460BTMagic) {
		cmd[11] = 0x02
	}

	// n10: Fixed 0x00 (already set)

	return d.Send(cmd)
}

// SendMargin sends the margin/feed amount command
// margin is in dots (at 180dpi: 14 dots = 2mm minimum)
func (d *Device) SendMargin(margin uint16) error {
	// ESC i d {n1} {n2} - margin = n1 + n2*256
	cmd := []byte{0x1b, 0x69, 0x64, byte(margin & 0xff), byte(margin >> 8)}
	return d.Send(cmd)
}

// Mode settings flags for ESC i M command
const (
	ModeAutocut = 0x40 // Auto-cut enabled
	ModeMirror  = 0x80 // Mirror printing enabled
)

// SendModeSettings sends the various mode settings command
func (d *Device) SendModeSettings(autocut, mirror bool) error {
	// ESC i M {n}
	cmd := []byte{0x1b, 0x69, 0x4d, 0x00}
	if autocut {
		cmd[3] |= ModeAutocut
	}
	if mirror {
		cmd[3] |= ModeMirror
	}
	return d.Send(cmd)
}

// SendPrecut sends the pre-cut command for devices that support it
func (d *Device) SendPrecut(precut bool) error {
	// ESC i M {n}
	cmd := []byte{0x1b, 0x69, 0x4d, 0x00}
	if precut {
		cmd[3] = 0x40
	}
	return d.Send(cmd)
}

// Advanced mode flags for ESC i K command
const (
	AdvHalfCut        = 0x04 // Half-cut enabled (not on PT-P710BT)
	AdvNoChain        = 0x08 // Feed after last label (no chain printing)
	AdvSpecialTape    = 0x10 // Special tape (no cutting)
	AdvHighResolution = 0x40 // 360dpi width (high-resolution mode)
	AdvNoBufferClear  = 0x80 // Don't clear buffer
)

// SendAdvancedMode sends the advanced mode settings command
func (d *Device) SendAdvancedMode(halfCut, noChain, specialTape, highRes bool) error {
	// ESC i K {n}
	cmd := []byte{0x1b, 0x69, 0x4b, 0x00}
	if halfCut {
		cmd[3] |= AdvHalfCut
	}
	if noChain {
		cmd[3] |= AdvNoChain
	}
	if specialTape {
		cmd[3] |= AdvSpecialTape
	}
	if highRes {
		cmd[3] |= AdvHighResolution
	}
	return d.Send(cmd)
}

// SendD460BTMagic sends magic commands for PT-D460BT series
func (d *Device) SendD460BTMagic() error {
	// ESC i d {n1} {n2} {n3} {n4}
	// n1,n2: length margin/spacing (uint16 little-endian)
	// n3: must be 0x4D or print gets corrupted
	// n4: ignored/reserved
	cmd := []byte{0x1b, 0x69, 0x64, 0x01, 0x00, 0x4d, 0x00}
	return d.Send(cmd)
}

// SendD460BTChain sends chain mode command for PT-D460BT series
func (d *Device) SendD460BTChain() error {
	// ESC i K 00
	cmd := []byte{0x1b, 0x69, 0x4b, 0x00}
	return d.Send(cmd)
}

// SendRaster sends a single raster line
func (d *Device) SendRaster(data []byte) error {
	maxBytes := d.Info.MaxPx / 8
	if len(data) > maxBytes {
		return fmt.Errorf("raster data too large: %d bytes (max %d)", len(data), maxBytes)
	}

	var buf []byte

	if d.HasFlag(FlagRasterPackbits) {
		// Fake compression by encoding a single uncompressed run
		// G {len+1} 00 {len-1} {data...}
		buf = make([]byte, len(data)+4)
		buf[0] = 0x47 // G
		buf[1] = byte(len(data) + 1)
		buf[2] = 0x00
		buf[3] = byte(len(data) - 1)
		copy(buf[4:], data)
	} else {
		// G {len} 00 {data...}
		buf = make([]byte, len(data)+3)
		buf[0] = 0x47 // G
		buf[1] = byte(len(data))
		buf[2] = 0x00
		copy(buf[3:], data)
	}

	return d.Send(buf)
}

// SendLF sends an empty line (line feed)
func (d *Device) SendLF() error {
	cmd := []byte{0x5a} // Z
	return d.Send(cmd)
}

// SendFF sends form feed (print and advance, no cut)
func (d *Device) SendFF() error {
	cmd := []byte{0x0c}
	return d.Send(cmd)
}

// Finalize finishes the print job
func (d *Device) Finalize(chain bool) error {
	if chain && !d.HasFlag(FlagD460BTMagic) {
		// Print without cut (chain mode)
		cmd := []byte{0x0c}
		return d.Send(cmd)
	}
	// Print with feeding and cut
	cmd := []byte{0x1a}
	return d.Send(cmd)
}

// SetPixel sets a pixel in a raster line
// rasterline is MSB-first, with the highest pixel index at byte 0
func SetPixel(rasterline []byte, pixel int) {
	if pixel < 0 || pixel >= len(rasterline)*8 {
		return
	}
	byteIdx := len(rasterline) - 1 - (pixel / 8)
	bitIdx := pixel % 8
	rasterline[byteIdx] |= 1 << bitIdx
}
