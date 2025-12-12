package device

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/google/gousb"
)

// USB endpoints for P-Touch printers
const (
	EndpointOut = 0x02 // Bulk out endpoint
	EndpointIn  = 0x81 // Bulk in endpoint
)

// Status represents the 32-byte printer status response
type Status struct {
	PrintheadMark uint8 // 0x80
	Size          uint8 // 0x20
	BrotherCode   uint8 // 'B'
	SeriesCode    uint8 // '0'
	Model         uint8
	Country       uint8 // '0'
	Reserved1     [2]byte
	Error         uint16 // Error codes (little-endian)
	MediaWidth    uint8  // Tape width in mm
	MediaType     uint8
	NCol          uint8 // 0
	Fonts         uint8 // 0
	JPFonts       uint8 // 0
	Mode          uint8
	Density       uint8 // 0
	MediaLen      uint8 // Table length, always 0
	StatusType    uint8
	PhaseType     uint8
	PhaseNumber   uint16 // (little-endian)
	NotifNumber   uint8
	Exp           uint8 // 0
	TapeColor     uint8
	TextColor     uint8
	HWSettings    [4]byte
	Reserved2     [2]byte
}

// Device represents an open P-Touch printer
type Device struct {
	ctx         *gousb.Context
	usbDev      *gousb.Device
	intf        *gousb.Interface
	intfDone    func()
	inEndpoint  *gousb.InEndpoint
	outEndpoint *gousb.OutEndpoint
	Info        *DeviceInfo
	Status      *Status
	TapeWidthPx uint16
	Debug       bool
}

// Open finds and opens the first supported P-Touch printer
func Open() (*Device, error) {
	ctx := gousb.NewContext()

	// Try each supported device
	for _, devInfo := range SupportedDevices {
		// Skip unsupported modes
		if devInfo.Flags&FlagPLite != 0 {
			continue
		}
		if devInfo.Flags&FlagUnsupRaster != 0 {
			continue
		}

		usbDev, err := ctx.OpenDeviceWithVIDPID(gousb.ID(devInfo.VID), gousb.ID(devInfo.PID))
		if err != nil {
			continue
		}
		if usbDev == nil {
			continue
		}

		fmt.Printf("%s found on USB bus %d, device %d\n",
			devInfo.Name, usbDev.Desc.Bus, usbDev.Desc.Address)

		// Check again for unsupported modes after finding
		if devInfo.Flags&FlagPLite != 0 {
			usbDev.Close()
			return nil, errors.New("printer is in P-Lite Mode, which is unsupported.\n" +
				"Turn off P-Lite mode by changing switch from position EL to position E\n" +
				"or by pressing the PLite button for ~2 seconds (or consult the manual)")
		}
		if devInfo.Flags&FlagUnsupRaster != 0 {
			usbDev.Close()
			return nil, errors.New("that printer currently is unsupported (it has a different raster data transfer)")
		}

		// Set auto-detach for kernel driver
		usbDev.SetAutoDetach(true)

		// Claim interface 0
		intf, done, err := usbDev.DefaultInterface()
		if err != nil {
			usbDev.Close()
			return nil, fmt.Errorf("failed to claim interface: %w", err)
		}

		// Get endpoints
		inEp, err := intf.InEndpoint(EndpointIn)
		if err != nil {
			done()
			usbDev.Close()
			return nil, fmt.Errorf("failed to get input endpoint: %w", err)
		}

		outEp, err := intf.OutEndpoint(EndpointOut)
		if err != nil {
			done()
			usbDev.Close()
			return nil, fmt.Errorf("failed to get output endpoint: %w", err)
		}

		dev := &Device{
			ctx:         ctx,
			usbDev:      usbDev,
			intf:        intf,
			intfDone:    done,
			inEndpoint:  inEp,
			outEndpoint: outEp,
			Info: &DeviceInfo{
				VID:   devInfo.VID,
				PID:   devInfo.PID,
				Name:  devInfo.Name,
				MaxPx: devInfo.MaxPx,
				DPI:   devInfo.DPI,
				Flags: devInfo.Flags,
			},
			Status: &Status{},
		}
		return dev, nil
	}

	ctx.Close()
	return nil, errors.New("no P-Touch printer found on USB (remember to put switch to position E)")
}

// Close releases the device
func (d *Device) Close() error {
	if d.intfDone != nil {
		d.intfDone()
	}
	if d.usbDev != nil {
		d.usbDev.Close()
	}
	if d.ctx != nil {
		d.ctx.Close()
	}
	return nil
}

// Send writes data to the printer
func (d *Device) Send(data []byte) error {
	if d.outEndpoint == nil {
		return errors.New("device not open")
	}
	if len(data) > 128 {
		return errors.New("data too large (max 128 bytes)")
	}

	n, err := d.outEndpoint.Write(data)
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("write error: sent only %d of %d bytes", n, len(data))
	}

	return nil
}

// Receive reads data from the printer
func (d *Device) Receive(buf []byte) (int, error) {
	if d.inEndpoint == nil {
		return 0, errors.New("device not open")
	}

	return d.inEndpoint.Read(buf)
}

// GetStatus queries the printer status
func (d *Device) GetStatus(timeout int) error {
	// Send status request: ESC i S
	cmd := []byte{0x1b, 0x69, 0x53}
	if err := d.Send(cmd); err != nil {
		return err
	}

	buf := make([]byte, 32)
	maxTries := timeout * 10
	if timeout == 0 {
		maxTries = 100000 // essentially infinite
	}

	for tries := 0; tries < maxTries; tries++ {
		time.Sleep(100 * time.Millisecond)

		n, err := d.Receive(buf)
		if err != nil {
			// Timeout is expected while waiting
			continue
		}

		if n == 32 && buf[0] == 0x80 && buf[1] == 0x20 {
			// Parse status
			d.Status.PrintheadMark = buf[0]
			d.Status.Size = buf[1]
			d.Status.BrotherCode = buf[2]
			d.Status.SeriesCode = buf[3]
			d.Status.Model = buf[4]
			d.Status.Country = buf[5]
			d.Status.Error = binary.LittleEndian.Uint16(buf[8:10])
			d.Status.MediaWidth = buf[10]
			d.Status.MediaType = buf[11]
			d.Status.NCol = buf[12]
			d.Status.Fonts = buf[13]
			d.Status.JPFonts = buf[14]
			d.Status.Mode = buf[15]
			d.Status.Density = buf[16]
			d.Status.MediaLen = buf[17]
			d.Status.StatusType = buf[18]
			d.Status.PhaseType = buf[19]
			d.Status.PhaseNumber = binary.LittleEndian.Uint16(buf[20:22])
			d.Status.NotifNumber = buf[22]
			d.Status.Exp = buf[23]
			d.Status.TapeColor = buf[24]
			d.Status.TextColor = buf[25]

			// Lookup tape width in pixels
			d.TapeWidthPx = GetTapePixelWidth(d.Status.MediaWidth)
			if d.TapeWidthPx == 0 {
				fmt.Printf("unknown tape width of %dmm, please report this.\n", d.Status.MediaWidth)
			}

			return nil
		}

		if n == 16 {
			fmt.Println("got only 16 bytes... wondering what they are")
			d.DumpRawStatus(buf[:16])
		}
	}

	return fmt.Errorf("timeout (%d sec) while waiting for status response", timeout)
}

// DumpRawStatus prints raw status bytes for debugging
func (d *Device) DumpRawStatus(raw []byte) {
	fmt.Println("debug: dumping raw status bytes")
	for i, b := range raw {
		fmt.Printf("0x%02x ", b)
		if (i+1)%16 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()
}

// GetMaxWidth returns the maximum pixel width for this printer
func (d *Device) GetMaxWidth() int {
	if d.Info == nil {
		return 0
	}
	return d.Info.MaxPx
}

// GetTapeWidth returns the tape width in pixels
func (d *Device) GetTapeWidth() int {
	return int(d.TapeWidthPx)
}

// HasFlag checks if the device has a specific flag
func (d *Device) HasFlag(flag int) bool {
	return d.Info != nil && (d.Info.Flags&flag) == flag
}
