# Brother PT-E550W/P750W/P710BT Raster Command Protocol Summary

Reference: `cv_pte550wp750wp710bt_eng_raster_102.pdf` (Version 1.02)

## Overview

This protocol allows direct control of Brother P-touch label printers without using the official printer driver. Communication is via USB or network, sending binary raster data with control commands.

## Print Data Structure

Print data consists of 4 sequential parts:

```
┌─────────────────────────┐
│ 1. Initialization       │  (once per job)
├─────────────────────────┤
│ 2. Control Codes        │  ┐
├─────────────────────────┤  │
│ 3. Raster Data          │  ├─ repeat per page
├─────────────────────────┤  │
│ 4. Print Command        │  ┘
└─────────────────────────┘
```

## 1. Initialization Commands

| Sequence | Command    | Hex              | Description                      |
| -------- | ---------- | ---------------- | -------------------------------- |
| 1        | Invalidate | `00` × 100 bytes | Reset printer to receiving state |
| 2        | Initialize | `1B 40`          | Initialize for printing          |

## 2. Control Codes (per page)

| Sequence | Command                  | Hex                   | Description                                      |
| -------- | ------------------------ | --------------------- | ------------------------------------------------ |
| 1        | Switch to raster mode    | `1B 69 61 01`         | Required before sending raster data              |
| 2        | Auto status notification | `1B 69 21 {n}`        | 0=notify, 1=don't notify (not on PT-E550W/P750W) |
| 3        | Print information        | `1B 69 7A` + 10 bytes | Media type, size, raster count                   |
| 4        | Various mode settings    | `1B 69 4D {n}`        | Auto-cut, mirror printing                        |
| 5        | Cut-each-N-labels        | `1B 69 41 {n}`        | Pages before auto-cut (not on PT-P710BT)         |
| 6        | Advanced mode            | `1B 69 4B {n}`        | Half-cut, chain printing, high-res               |
| 7        | Margin amount            | `1B 69 64 {n1} {n2}`  | Feed amount in dots                              |
| 8        | Compression mode         | `4D {n}`              | 0=none, 2=TIFF                                   |

## 3. Raster Data Commands

| Command                  | Hex  | Format                     | Description                        |
| ------------------------ | ---- | -------------------------- | ---------------------------------- |
| Raster graphics transfer | `47` | `47 {n1} {n2} {d1}...{dk}` | Transfer k bytes of raster data    |
| Zero raster graphics     | `5A` | `5A`                       | Blank raster line (TIFF mode only) |

## 4. Print Commands

| Command                     | Hex  | Description          |
| --------------------------- | ---- | -------------------- |
| Print (FF)                  | `0C` | End of non-last page |
| Print with feeding (Ctrl-Z) | `1A` | End of last page     |

---

## Command Details

### ESC i z - Print Information Command

```
1B 69 7A {n1} {n2} {n3} {n4} {n5} {n6} {n7} {n8} {n9} {n10}
```

| Byte  | Description                                                                                   |
| ----- | --------------------------------------------------------------------------------------------- |
| n1    | Valid flags: `0x02`=media type, `0x04`=width, `0x08`=length, `0x40`=quality, `0x80`=recovery  |
| n2    | Media type: `0x00`=none, `0x01`=laminated, `0x03`=non-laminated, `0x11`=HS 2:1, `0x17`=HS 3:1 |
| n3    | Media width (mm)                                                                              |
| n4    | Media length (mm), usually `0x00`                                                             |
| n5-n8 | Raster line count (little-endian: n8×256³ + n7×256² + n6×256 + n5)                            |
| n9    | Page flag: `0x00`=first page, `0x01`=subsequent pages                                         |
| n10   | Fixed `0x00`                                                                                  |

### ESC i M - Various Mode Settings

```
1B 69 4D {n1}
```

| Bit | Mask   | Description                |
| --- | ------ | -------------------------- |
| 6   | `0x40` | Auto-cut: 1=enabled        |
| 7   | `0x80` | Mirror printing: 1=enabled |

### ESC i K - Advanced Mode Settings

```
1B 69 4B {n1}
```

| Bit | Mask   | Description                                |
| --- | ------ | ------------------------------------------ |
| 2   | `0x04` | Half-cut: 1=on (not PT-P710BT)             |
| 3   | `0x08` | No chain printing: 1=feed after last label |
| 4   | `0x10` | Special tape (no cutting): 1=on            |
| 6   | `0x40` | High-resolution: 1=360dpi width            |
| 7   | `0x80` | No buffer clearing                         |

### ESC i d - Margin Amount

```
1B 69 64 {n1} {n2}
```

Margin in dots = n1 + n2×256

| Resolution  | Min Margin    | Max Margin        |
| ----------- | ------------- | ----------------- |
| 180×180 dpi | 14 dots (2mm) | 900 dots (127mm)  |
| 180×360 dpi | 28 dots (2mm) | 1800 dots (127mm) |

---

## Raster Line Format

The print head has **128 pins** total. Each raster line is **16 bytes** (128 bits).

```
Pin 0 (top)                                              Pin 127 (bottom)
    │                                                         │
    ▼                                                         ▼
┌────────┬────────┬────────┬─────────────────────────┬────────┐
│ Byte 0 │ Byte 1 │ Byte 2 │          ...            │Byte 15 │
└────────┴────────┴────────┴─────────────────────────┴────────┘
    MSB→LSB  MSB→LSB  ...
```

### Pin Layout by Tape Width (TZe tape)

| Tape  | Left Margin (pins) | Print Area (pins) | Right Margin (pins) |
| ----- | ------------------ | ----------------- | ------------------- |
| 3.5mm | 52                 | 24                | 52                  |
| 6mm   | 48                 | 32                | 48                  |
| 9mm   | 39                 | 50                | 39                  |
| 12mm  | 29                 | 70                | 29                  |
| 18mm  | 8                  | 112               | 8                   |
| 24mm  | 0                  | 128               | 0                   |

### Heat-Shrink Tube Pin Layout

| Tape      | Left Margin | Print Area | Right Margin |
| --------- | ----------- | ---------- | ------------ |
| HS 5.8mm  | 50          | 28         | 50           |
| HS 8.8mm  | 40          | 48         | 40           |
| HS 11.7mm | 31          | 66         | 31           |
| HS 17.7mm | 11          | 106        | 11           |
| HS 23.6mm | 0           | 128        | 0            |

---

## TIFF Compression (PackBits)

When compression mode is set to TIFF (`4D 02`):

- **Repeated bytes**: `[-(count-1)] [byte]`
  - e.g., 20 zeros → `ED 00` (0xED = -19 in signed byte)
- **Literal run**: `[count-1] [bytes...]`
  - e.g., 6 different bytes → `05 d1 d2 d3 d4 d5 d6`

Example:

```
Uncompressed: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 22 22 23 BA BF A2 22 2B
Compressed:   ED 00 FF 22 05 23 BA BF A2 22 2B
              └─┬─┘ └─┬─┘ └──────────┬──────────┘
         20×0x00  2×0x22    6 literal bytes
```

---

## Status Information (ESC i S)

Request: `1B 69 53`
Response: 32 bytes

| Offset | Size | Field           | Description                                                                     |
| ------ | ---- | --------------- | ------------------------------------------------------------------------------- |
| 0      | 1    | Print head mark | `0x80`                                                                          |
| 1      | 1    | Size            | `0x20` (32)                                                                     |
| 2      | 1    | Brother code    | `0x42` ('B')                                                                    |
| 3      | 1    | Series code     | `0x30` ('0')                                                                    |
| 4      | 1    | Model code      | PT-E550W=`0x66`, PT-P750W=`0x68`, PT-P710BT=`0x70`                              |
| 8      | 1    | Error info 1    | See error flags                                                                 |
| 9      | 1    | Error info 2    | See error flags                                                                 |
| 10     | 1    | Media width     | Width in mm                                                                     |
| 11     | 1    | Media type      | `0x00`=none, `0x01`=laminated, `0x03`=non-laminated, `0x11`=HS2:1, `0x17`=HS3:1 |
| 15     | 1    | Mode            | Current mode settings                                                           |
| 17     | 1    | Media length    | Length in mm (usually 0)                                                        |
| 18     | 1    | Status type     | `0x00`=reply, `0x01`=complete, `0x02`=error, `0x06`=phase change                |
| 19     | 1    | Phase type      | `0x00`=editing, `0x01`=printing                                                 |
| 20-21  | 2    | Phase number    | State within phase                                                              |
| 22     | 1    | Notification    | `0x00`=none, `0x01`=cover open, `0x02`=cover closed                             |
| 24     | 1    | Tape color      | See tape color table                                                            |
| 25     | 1    | Text color      | See text color table                                                            |

### Error Info 1 (offset 8)

| Bit | Mask   | Error                |
| --- | ------ | -------------------- |
| 0   | `0x01` | No media             |
| 2   | `0x04` | Cutter jam           |
| 3   | `0x08` | Weak batteries       |
| 6   | `0x40` | High-voltage adapter |

### Error Info 2 (offset 9)

| Bit | Mask   | Error       |
| --- | ------ | ----------- |
| 0   | `0x01` | Wrong media |
| 4   | `0x10` | Cover open  |
| 5   | `0x20` | Overheating |

### Tape Colors

| ID     | Color               |
| ------ | ------------------- |
| `0x01` | White               |
| `0x02` | Other               |
| `0x03` | Clear               |
| `0x04` | Red                 |
| `0x05` | Blue                |
| `0x06` | Yellow              |
| `0x07` | Green               |
| `0x08` | Black               |
| `0x70` | White (Heat-shrink) |

### Text Colors

| ID     | Color |
| ------ | ----- |
| `0x01` | White |
| `0x04` | Red   |
| `0x05` | Blue  |
| `0x08` | Black |
| `0x0A` | Gold  |

---

## USB Interface

| Property               | Value                                        |
| ---------------------- | -------------------------------------------- |
| Vendor ID              | `0x04F9`                                     |
| Product ID (PT-E550W)  | `0x2060`                                     |
| Product ID (PT-P750W)  | `0x2062`                                     |
| Product ID (PT-P710BT) | `0x20AF`                                     |
| Class                  | Printer                                      |
| Speed                  | Full-speed                                   |
| Endpoint 1             | IN bulk (status from printer), 64 bytes max  |
| Endpoint 2             | OUT bulk (commands to printer), 64 bytes max |

---

## Printing Flow

### Basic Sequence

```
1. Open USB port
2. Send: Invalidate (100× 0x00)
3. Send: Initialize (1B 40)
4. Send: Status request (1B 69 53)
5. Read: Status (32 bytes) - check for errors
6. Send: Control codes (mode, media info, margins, compression)
7. Send: Raster data (G commands or Z for blank lines)
8. Send: Print command (0C or 1A for last page)
9. Read: Status - wait for completion (status type = 0x01)
10. Close USB port
```

### Concurrent vs Buffered Printing

- **Concurrent** (USB, uncompressed): Printing starts immediately when data is received
- **Buffered** (network or compressed): Printing starts after full page received

---

## Page Size Limits

### TZe Tape

| Resolution  | Min Length      | Max Length          |
| ----------- | --------------- | ------------------- |
| 180×180 dpi | 31 dots (4.4mm) | 7086 dots (1000mm)  |
| 180×360 dpi | 60 dots (4.2mm) | 14172 dots (1000mm) |

### Heat-Shrink Tube

| Resolution  | Min Length      | Max Length        |
| ----------- | --------------- | ----------------- |
| 180×180 dpi | 31 dots (4.4mm) | 3543 dots (500mm) |

**Note**: Minimum physical tape output is 24.5mm due to cutter position.

---

## Example: Minimal Print Job

```hex
# Invalidate (100 bytes)
00 00 00 00 00 00 00 00 00 00 ... (100 times)

# Initialize
1B 40

# Switch to raster mode
1B 69 61 01

# Print info: 24mm tape, laminated, 100 raster lines
1B 69 7A 8E 01 18 00 64 00 00 00 00 00
#          │  │  │  │  └─ raster count = 100 (0x64)
#          │  │  │  └──── length = 0
#          │  │  └─────── width = 24mm (0x18)
#          │  └────────── media type = laminated (0x01)
#          └───────────── flags = type+width+length+recover (0x8E)

# Mode: auto-cut enabled
1B 69 4D 40

# Advanced: no chain printing
1B 69 4B 08

# Margin: 14 dots (2mm at 180dpi)
1B 69 64 0E 00

# Compression: TIFF
4D 02

# Raster data (100 lines)
47 10 00 [16 bytes of raster data]  # line 1
47 10 00 [16 bytes of raster data]  # line 2
...
5A                                   # blank line (zero raster)
...

# Print with feeding (last page)
1A
```
