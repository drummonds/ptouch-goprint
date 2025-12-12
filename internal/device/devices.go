package device

import "fmt"

// Device capability flags
const (
	FlagNone           = 0
	FlagUnsupRaster    = 1 << 0 // Unsupported raster mode
	FlagRasterPackbits = 1 << 1 // Use PackBits compression
	FlagPLite          = 1 << 2 // P-Lite mode (unsupported)
	FlagP700Init       = 1 << 3 // PT-P700 series initialization
	FlagUseInfoCmd     = 1 << 4 // Requires print info command
	FlagHasPrecut      = 1 << 5 // Has pre-cut capability
	FlagD460BTMagic    = 1 << 6 // PT-D460BT series magic commands
)

// DeviceInfo describes a supported P-Touch printer model
type DeviceInfo struct {
	VID   uint16 // USB Vendor ID
	PID   uint16 // USB Product ID
	Name  string // Model name
	MaxPx int    // Maximum pixel width
	DPI   int    // Dots per inch
	Flags int    // Capability flags
}

// BrotherVID is the USB vendor ID for Brother devices
const BrotherVID = 0x04f9

// SupportedDevices is the database of all supported P-Touch printers
var SupportedDevices = []DeviceInfo{
	{BrotherVID, 0x2001, "PT-9200DX", 384, 360, FlagRasterPackbits | FlagHasPrecut},
	{BrotherVID, 0x2004, "PT-2300", 112, 180, FlagRasterPackbits | FlagHasPrecut},
	{BrotherVID, 0x2007, "PT-2420PC", 128, 180, FlagRasterPackbits},
	{BrotherVID, 0x2011, "PT-2450PC", 128, 180, FlagRasterPackbits},
	{BrotherVID, 0x2019, "PT-1950", 112, 180, FlagRasterPackbits},
	{BrotherVID, 0x201f, "PT-2700", 128, 180, FlagHasPrecut},
	{BrotherVID, 0x202c, "PT-1230PC", 128, 180, FlagNone},
	{BrotherVID, 0x202d, "PT-2430PC", 128, 180, FlagNone},
	{BrotherVID, 0x2030, "PT-1230PC (PLite Mode)", 128, 180, FlagPLite},
	{BrotherVID, 0x2031, "PT-2430PC (PLite Mode)", 128, 180, FlagPLite},
	{BrotherVID, 0x2041, "PT-2730", 128, 180, FlagNone},
	{BrotherVID, 0x205e, "PT-H500", 128, 180, FlagRasterPackbits | FlagHasPrecut},
	{BrotherVID, 0x205f, "PT-E500", 128, 180, FlagRasterPackbits},
	{BrotherVID, 0x2060, "PT-E550W", 128, 180, FlagRasterPackbits | FlagP700Init | FlagUseInfoCmd | FlagHasPrecut},
	{BrotherVID, 0x2061, "PT-P700", 128, 180, FlagRasterPackbits | FlagP700Init | FlagHasPrecut},
	{BrotherVID, 0x2062, "PT-P750W", 128, 180, FlagRasterPackbits | FlagP700Init},
	{BrotherVID, 0x2064, "PT-P700 (PLite Mode)", 128, 180, FlagPLite},
	{BrotherVID, 0x2065, "PT-P750W (PLite Mode)", 128, 180, FlagPLite},
	{BrotherVID, 0x20df, "PT-D410", 128, 180, FlagUseInfoCmd | FlagHasPrecut | FlagD460BTMagic},
	{BrotherVID, 0x2073, "PT-D450", 128, 180, FlagUseInfoCmd},
	{BrotherVID, 0x20e0, "PT-D460BT", 128, 180, FlagP700Init | FlagUseInfoCmd | FlagHasPrecut | FlagD460BTMagic},
	{BrotherVID, 0x2074, "PT-D600", 128, 180, FlagRasterPackbits},
	{BrotherVID, 0x20e1, "PT-D610BT", 128, 180, FlagP700Init | FlagUseInfoCmd | FlagHasPrecut | FlagD460BTMagic},
	{BrotherVID, 0x20af, "PT-P710BT", 128, 180, FlagRasterPackbits | FlagHasPrecut},
	{BrotherVID, 0x2201, "PT-E310BT", 128, 180, FlagP700Init | FlagUseInfoCmd | FlagD460BTMagic},
	{BrotherVID, 0x2203, "PT-E560BT", 128, 180, FlagP700Init | FlagUseInfoCmd | FlagD460BTMagic},
}

// FindDevice looks up a device by VID/PID
func FindDevice(vid, pid uint16) *DeviceInfo {
	for i := range SupportedDevices {
		if SupportedDevices[i].VID == vid && SupportedDevices[i].PID == pid {
			return &SupportedDevices[i]
		}
	}
	return nil
}

// ListSupported prints all supported printer models
func ListSupported() {
	fmt.Println("Supported printers (some might have quirks)")
	col := 0
	const columns = 5
	for _, dev := range SupportedDevices {
		if dev.Flags&FlagPLite != FlagPLite {
			fmt.Print(dev.Name)
			if len(dev.Name) < 8 {
				fmt.Print("\t")
			}
			if col < columns-1 {
				fmt.Print("\t")
			} else {
				fmt.Println()
			}
			col++
			col = col % columns
		}
	}
	fmt.Println()
}
