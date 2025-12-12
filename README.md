# About:

This is a port of https://github.com/HenrikBengtsson/brother-ptouch-label-printer-on-linux
which was written in C and is now converted to go and support extended to E550W

It is not yet pure go but I hope it will get there eventually.  At the moment is very
experimental but I need to to work well enough to print my XMAS card labels.

ptouch-goprint is a command line tool to print labels on Brother P-Touch
printers on Linux.

There is no need to install the printer via CUPS, the printer is accessed
directly via libusb.

The tool was written for and tested with the PT-2430PC, but meanwhile is also
used with others (see "ptouch-goprint --list-supported")
Maybe others work too (please report USB VID and PID so I can include support
for further models, too).

Further info can be found at:
https://dominic.familie-radermacher.ch/projekte/ptouch-print/


## PT-550W summary

The protocol was summarised from this document [cv_pte550wp750wp710bt_eng_raster_102 downloaded from Brother](https://download.brother.com/welcome/docp100064/cv_pte550wp750wp710bt_eng_raster_102.pdf) 
and a summary in protocol_summary.md is there.

Note:

Dear visitor, currently I have absolutely no time for improvements on this
project (my free time currently is about one or two hours PER MONTH).
Therefore, I can not look at suggestions about improvements.

# Requires task from Taskfile.dev to build

Also need `sudo apt install libusb-1.0-0-dev` to release.

## Using as a Go Module

> **Note:** The core packages are currently in `internal/`, so they cannot be imported by external programs. To use this as a library, you would need to fork the repository and move packages from `internal/` to a public location (e.g., `pkg/`).

Here's an example of printing one or more PNG images:

```go
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/drummonds/ptouch-goprint/pkg/device"  // after moving from internal/
	"github.com/drummonds/ptouch-goprint/pkg/render"  // after moving from internal/
)

func main() {
	images := []string{"label1.png", "label2.png", "label3.png"}

	// Open the printer
	dev, err := device.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening printer: %v\n", err)
		os.Exit(1)
	}
	defer dev.Close()

	// Initialize and get status
	if err := dev.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing: %v\n", err)
		os.Exit(1)
	}

	if err := dev.GetStatus(1); err != nil {
		fmt.Fprintf(os.Stderr, "Error getting status: %v\n", err)
		os.Exit(1)
	}

	// Print each image
	for i, imgPath := range images {
		img, err := loadPNG(imgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", imgPath, err)
			os.Exit(1)
		}

		// Use chain mode for all but the last image (avoids cutting between labels)
		isLast := i == len(images)-1
		if err := printImage(dev, img, !isLast); err != nil {
			fmt.Fprintf(os.Stderr, "Error printing %s: %v\n", imgPath, err)
			os.Exit(1)
		}
		fmt.Printf("Printed %s\n", imgPath)
	}

	fmt.Println("All labels printed!")
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func printImage(dev *device.Device, img image.Image, chain bool) error {
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	tapeWidth := dev.GetTapeWidth()
	maxPixels := dev.GetMaxWidth()

	if imgHeight > tapeWidth {
		return fmt.Errorf("image too tall (%dpx), max tape width is %dpx", imgHeight, tapeWidth)
	}

	// Calculate centering offset
	offset := (maxPixels / 2) - (imgHeight / 2)

	// Start raster mode
	if err := dev.RasterStart(); err != nil {
		return err
	}

	// Send info command for E550W/P750W printers
	if dev.HasFlag(device.FlagUseInfoCmd) {
		if err := dev.InfoCmd(imgWidth); err != nil {
			return err
		}
	}

	// Send raster data column by column
	for x := 0; x < imgWidth; x++ {
		raster := render.ImageToRaster(img, x, maxPixels, offset)
		if err := dev.SendRaster(raster); err != nil {
			return err
		}
	}

	// Finalize - chain=true skips cutting between labels
	return dev.Finalize(chain)
}
```

For now, the easiest way to print from another program is to shell out to the CLI:

```go
// Print a single image
cmd := exec.Command("ptouch-goprint", "-image", "label.png")
if err := cmd.Run(); err != nil {
    log.Fatal(err)
}

// Print multiple images with chain mode (no cut between labels)
cmd := exec.Command("ptouch-goprint",
    "-image", "label1.png",
    "-image", "label2.png",
    "-image", "label3.png",
    "-chain")
if err := cmd.Run(); err != nil {
    log.Fatal(err)
}
```