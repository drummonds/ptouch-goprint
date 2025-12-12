package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"strings"

	"github.com/drummonds/ptouch-goprint/internal/device"
	"github.com/drummonds/ptouch-goprint/internal/job"
	"github.com/drummonds/ptouch-goprint/internal/render"
	"github.com/google/gousb"
)

const version = "0.1.0-go"

// Command-line flags
var (
	debug          = flag.Bool("debug", false, "Enable debug output")
	fontFile       = flag.String("font", "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", "Font file path")
	fontSize       = flag.Float64("fontsize", 0, "Manually set font size (0=auto)")
	writePNG       = flag.String("writepng", "", "Write output to PNG file instead of printing")
	forceTapeWidth = flag.Int("force-tape-width", 0, "Force tape width in pixels (for PNG preview)")
	copies         = flag.Int("copies", 1, "Number of copies to print")
	timeout        = flag.Int("timeout", 1, "Timeout in seconds for waiting on printer (0=infinite)")
	chain          = flag.Bool("chain", false, "Skip final feed and auto-cut")
	precut         = flag.Bool("precut", false, "Add cut before label")
	showInfo       = flag.Bool("info", false, "Show tape/printer info")
	listSupported  = flag.Bool("list-supported", false, "List supported printers")
	listUSB        = flag.Bool("list-usb", false, "List all USB devices (for debugging)")

	// Job flags - these can be specified multiple times
	textLines  stringSlice
	imageFiles stringSlice
	padPixels  intSlice
	cutmarks   int
	align      = flag.String("align", "l", "Text alignment: l(eft), c(enter), r(ight)")
)

// stringSlice implements flag.Value for multiple string flags
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

// intSlice implements flag.Value for multiple int flags
type intSlice []int

func (s *intSlice) String() string { return fmt.Sprintf("%v", *s) }
func (s *intSlice) Set(v string) error {
	var i int
	if _, err := fmt.Sscanf(v, "%d", &i); err != nil {
		return err
	}
	*s = append(*s, i)
	return nil
}

func init() {
	flag.Var(&textLines, "text", "Print text (can be specified multiple times)")
	flag.Var(&textLines, "t", "Print text (shorthand)")
	flag.Var(&textLines, "newline", "Add text on new line")
	flag.Var(&textLines, "n", "Add text on new line (shorthand)")
	flag.Var(&imageFiles, "image", "Print PNG image")
	flag.Var(&imageFiles, "i", "Print PNG image (shorthand)")
	flag.Var(&padPixels, "pad", "Add padding pixels")
	flag.Var(&padPixels, "p", "Add padding pixels (shorthand)")
	flag.IntVar(&cutmarks, "cutmark", 0, "Add cutmark (use -cutmark=1)")
	flag.IntVar(&cutmarks, "c", 0, "Add cutmark (shorthand)")
}

func main() {
	flag.Parse()

	if *listSupported {
		device.ListSupported()
		os.Exit(0)
	}

	if *listUSB {
		listUSBDevices()
		os.Exit(0)
	}

	// Determine print width
	var printWidth int
	var dev *device.Device
	var err error

	if *forceTapeWidth > 0 {
		if *writePNG == "" {
			fmt.Fprintln(os.Stderr, "Error: --force-tape-width requires --writepng")
			os.Exit(1)
		}
		printWidth = *forceTapeWidth
	} else {
		// Open printer
		dev, err = device.Open()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(5)
		}
		defer dev.Close()

		dev.Debug = *debug

		// Initialize and get status
		if err := dev.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: init failed: %v\n", err)
			os.Exit(1)
		}

		if err := dev.GetStatus(*timeout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: get status failed: %v\n", err)
			os.Exit(1)
		}

		printWidth = dev.GetTapeWidth()
		maxWidth := dev.GetMaxWidth()
		if printWidth > maxWidth {
			printWidth = maxWidth
		}
	}

	// Show info and exit
	if *showInfo {
		if dev == nil {
			fmt.Fprintln(os.Stderr, "Error: --info requires printer connection")
			os.Exit(1)
		}
		fmt.Printf("maximum printing width for this printer is %dpx\n", dev.GetMaxWidth())
		fmt.Printf("maximum printing width for this tape is %dpx\n", dev.GetTapeWidth())
		fmt.Printf("media type = 0x%02x (%s)\n", dev.Status.MediaType, device.MediaType(dev.Status.MediaType))
		fmt.Printf("media width = %d mm\n", dev.Status.MediaWidth)
		fmt.Printf("tape color = 0x%02x (%s)\n", dev.Status.TapeColor, device.TapeColor(dev.Status.TapeColor))
		fmt.Printf("text color = 0x%02x (%s)\n", dev.Status.TextColor, device.TextColor(dev.Status.TextColor))
		fmt.Printf("error = 0x%04x\n", dev.Status.Error)
		if *debug {
			dev.DumpRawStatus([]byte{
				dev.Status.PrintheadMark, dev.Status.Size, dev.Status.BrotherCode, dev.Status.SeriesCode,
				dev.Status.Model, dev.Status.Country, 0, 0,
				byte(dev.Status.Error & 0xff), byte(dev.Status.Error >> 8),
				dev.Status.MediaWidth, dev.Status.MediaType, dev.Status.NCol, dev.Status.Fonts,
				dev.Status.JPFonts, dev.Status.Mode,
			})
		}
		os.Exit(0)
	}

	// Build job queue
	queue := job.NewQueue()

	// Parse alignment
	alignment := render.AlignLeft
	switch strings.ToLower(*align) {
	case "c", "center":
		alignment = render.AlignCenter
	case "r", "right":
		alignment = render.AlignRight
	}

	// Add text jobs
	if len(textLines) > 0 {
		j := queue.AddText(textLines[0])
		for i := 1; i < len(textLines) && j.N < job.MaxLines; i++ {
			j.Lines[j.N] = textLines[i]
			j.N++
		}
	}

	// Add image jobs
	for _, imgFile := range imageFiles {
		queue.AddImage(imgFile)
	}

	// Add padding jobs
	for _, px := range padPixels {
		queue.AddPadding(px)
	}

	// Add cutmarks
	for i := 0; i < cutmarks; i++ {
		queue.AddCutmark()
	}

	// Process jobs into final image
	var finalImage image.Image
	textRenderer := render.NewTextRenderer(*fontFile)
	textRenderer.FontSize = *fontSize
	textRenderer.Debug = *debug

	err = queue.Iterate(func(j *job.Job) error {
		var img image.Image
		var err error

		switch j.Type {
		case job.TypeText:
			lines := make([]string, 0, j.N)
			for i := 0; i < j.N; i++ {
				lines = append(lines, j.Lines[i])
			}
			if *debug {
				fmt.Printf("job: text (%d lines)\n", len(lines))
			}
			img, err = textRenderer.RenderText(lines, printWidth, alignment)
			if err != nil {
				return fmt.Errorf("text render failed: %w", err)
			}

		case job.TypeImage:
			if *debug {
				fmt.Printf("job: image %s\n", j.Lines[0])
			}
			img, err = render.LoadImage(j.Lines[0])
			if err != nil {
				return fmt.Errorf("image load failed: %w", err)
			}

		case job.TypeCutmark:
			if *debug {
				fmt.Println("job: cutmark")
			}
			img = render.CreateCutMark(printWidth)

		case job.TypePad:
			if *debug {
				fmt.Printf("job: pad %d pixels\n", j.N)
			}
			img = render.CreatePadding(printWidth, j.N)
		}

		finalImage = render.AppendImages(finalImage, img)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if finalImage == nil {
		if *debug {
			fmt.Println("nothing to print")
		}
		os.Exit(0)
	}

	// Write PNG or print
	if *writePNG != "" {
		if err := render.SavePNG(finalImage, *writePNG); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Wrote %s\n", *writePNG)
		os.Exit(0)
	}

	// Print to device
	if err := printImage(dev, finalImage, *chain, *precut, *copies); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
}

func printImage(dev *device.Device, img image.Image, chain, precut bool, copies int) error {
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	tapeWidth := dev.GetTapeWidth()
	maxPixels := dev.GetMaxWidth()

	if imgHeight > tapeWidth {
		return fmt.Errorf("image too large (%dpx x %dpx), max tape width is %dpx", imgWidth, imgHeight, tapeWidth)
	}

	fmt.Printf("image size (%dpx x %dpx)\n", imgWidth, imgHeight)

	// Calculate centering offset
	offset := (maxPixels / 2) - (imgHeight / 2)
	fmt.Printf("max_pixels=%d, offset=%d\n", maxPixels, offset)

	for copy := 0; copy < copies; copy++ {
		// Start raster mode (must be first for E550W/P750W protocol)
		if err := dev.RasterStart(); err != nil {
			return fmt.Errorf("raster start failed: %w", err)
		}

		// Send info command if needed (E550W/P750W protocol)
		if dev.HasFlag(device.FlagUseInfoCmd) {
			if err := dev.InfoCmd(imgWidth); err != nil {
				return err
			}
			if *debug {
				fmt.Println("send print information command")
			}
		}

		// D460BT magic
		if dev.HasFlag(device.FlagD460BTMagic) {
			if err := dev.SendD460BTMagic(); err != nil {
				return err
			}
			if *debug {
				fmt.Println("send PT-D460BT magic commands")
			}
		}

		// Mode settings: auto-cut (for E550W protocol)
		// Note: precut uses the same ESC i M command
		if dev.HasFlag(device.FlagHasPrecut) {
			if err := dev.SendPrecut(precut); err != nil {
				return err
			}
			if *debug && precut {
				fmt.Println("send precut command")
			}
		}

		// Advanced mode: chain printing control (E550W/P750W protocol)
		// noChain=true means feed after last label (normal mode)
		// noChain=false means chain printing (don't feed)
		if dev.HasFlag(device.FlagP700Init) && !dev.HasFlag(device.FlagD460BTMagic) {
			if err := dev.SendAdvancedMode(false, !chain, false, false); err != nil {
				return err
			}
			if *debug {
				fmt.Printf("send advanced mode (chain=%v)\n", chain)
			}
		}

		// D460BT chain command
		if dev.HasFlag(device.FlagD460BTMagic) && chain {
			if err := dev.SendD460BTChain(); err != nil {
				return err
			}
			if *debug {
				fmt.Println("send PT-D460BT chain commands")
			}
		}

		// Send margin (E550W/P750W protocol: 14 dots = 2mm at 180dpi)
		if dev.HasFlag(device.FlagP700Init) && !dev.HasFlag(device.FlagD460BTMagic) {
			if err := dev.SendMargin(14); err != nil {
				return err
			}
			if *debug {
				fmt.Println("send margin command (14 dots)")
			}
		}

		// Enable PackBits compression if needed
		if dev.HasFlag(device.FlagRasterPackbits) {
			if *debug {
				fmt.Println("enable PackBits mode")
			}
			if err := dev.EnablePackbits(); err != nil {
				return err
			}
		}

		// Send raster data
		for x := 0; x < imgWidth; x++ {
			raster := render.ImageToRaster(img, x, maxPixels, offset)
			if err := dev.SendRaster(raster); err != nil {
				return fmt.Errorf("send raster failed: %w", err)
			}
		}

		// Finalize (cut or chain mode)
		isLastCopy := copy == copies-1
		if err := dev.Finalize(chain || !isLastCopy); err != nil {
			return fmt.Errorf("finalize failed: %w", err)
		}
	}

	return nil
}

func listUSBDevices() {
	ctx := gousb.NewContext()
	defer ctx.Close()

	fmt.Printf("%-6s %-6s %-5s %-5s %s\n", "VID", "PID", "Bus", "Addr", "Supported")
	fmt.Println(strings.Repeat("-", 60))

	count := 0
	_, _ = ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		count++
		vid := uint16(desc.Vendor)
		pid := uint16(desc.Product)

		// Check if it's a supported P-Touch printer
		supported := ""
		if devInfo := device.FindDevice(vid, pid); devInfo != nil {
			if devInfo.Flags&device.FlagUnsupRaster != 0 {
				supported = fmt.Sprintf("UNSUPPORTED (%s - different raster protocol)", devInfo.Name)
			} else if devInfo.Flags&device.FlagPLite != 0 {
				supported = fmt.Sprintf("UNSUPPORTED (%s - P-Lite mode)", devInfo.Name)
			} else {
				supported = fmt.Sprintf("YES (%s)", devInfo.Name)
			}
		}

		fmt.Printf("0x%04x 0x%04x %-5d %-5d %s\n", vid, pid, desc.Bus, desc.Address, supported)
		return false // Don't actually open the device
	})

	fmt.Printf("\nFound %d USB device(s)\n", count)

	// Also check specifically for Brother devices
	fmt.Println("\nScanning for Brother P-Touch printers...")
	found := false
	for _, devInfo := range device.SupportedDevices {
		if devInfo.Flags&device.FlagPLite != 0 || devInfo.Flags&device.FlagUnsupRaster != 0 {
			continue
		}
		dev, err := ctx.OpenDeviceWithVIDPID(gousb.ID(devInfo.VID), gousb.ID(devInfo.PID))
		if err == nil && dev != nil {
			fmt.Printf("  Found: %s (VID=0x%04x PID=0x%04x)\n", devInfo.Name, devInfo.VID, devInfo.PID)
			dev.Close()
			found = true
		}
	}
	if !found {
		fmt.Println("  No supported P-Touch printers found")
	}
}
