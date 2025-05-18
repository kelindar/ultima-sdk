// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"encoding/binary"
	"fmt"
	"image"
	"iter"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
	"github.com/kelindar/ultima-sdk/internal/uofile"
)

// Gump represents a UI element or graphic.
type Gump struct {
	ID     int         // ID of the gump
	Width  int         // Width in pixels
	Height int         // Height in pixels
	Image  image.Image // Image of the gump
}

// Gump retrieves a specific gump graphic by its ID.
// It handles reading from .mul or UOP files.
// The returned Gump object allows for lazy loading of its image.
func (s *SDK) Gump(id int) (*Gump, error) {
	file, err := s.loadGump()
	if err != nil {
		return nil, err
	}

	g, err := uofile.Decode(file, uint32(id), decodeGump)
	if err != nil {
		return nil, err
	}

	g.ID = id
	return g, nil
}

// Gumps returns an iterator over metadata (ID, width, height) for all available gumps.
// This is efficient for listing gumps without loading all their pixel data.
func (s *SDK) Gumps() iter.Seq[*Gump] {
	return func(yield func(*Gump) bool) {
		file, err := s.loadGump()
		if err != nil {
			return
		}

		for id := range file.Entries() {
			g, err := uofile.Decode(file, uint32(id), decodeGump)
			if err != nil {
				continue
			}

			g.ID = int(id)
			if !yield(g) {
				break
			}
		}
	}
}

func decodeGump(data []byte, extra uint64) (*Gump, error) {
	width := int(extra & 0xFFFF)
	height := int((extra >> 32) & 0xFFFF)

	// Sanity check
	if width <= 0 || height <= 0 || width > 2048 || height > 2048 {
		return nil, fmt.Errorf("%w: invalid gump dimensions %dx%d", ErrInvalidArtData, width, height)
	}

	img, err := decodeGumpData(data, width, height)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode gump: %v", ErrInvalidArtData, err)
	}

	return &Gump{
		Width:  width,
		Height: height,
		Image:  img,
	}, nil
}

// decodeGumpData converts raw gump data into an image.Image (RGBA8888).
func decodeGumpData(data []byte, width, height int) (image.Image, error) {
	need := height * 4
	if len(data) < need {
		return nil, fmt.Errorf("data too short for lookup table")
	}

	// Parse lookup table (height * uint32).
	lookup := make([]uint32, height)
	for y := 0; y < height; y++ {
		lookup[y] = binary.LittleEndian.Uint32(data[y*4:])
	}

	// Stage-1: decode straight into 1555 buffer.
	img1555 := bitmap.NewARGB1555(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		pos := int(lookup[y]) * 4 // byte offset from table start
		x := 0
		for x < width {
			if pos+3 >= len(data) {
				return nil, fmt.Errorf("RLE overflow at line %d", y)
			}
			color16 := binary.LittleEndian.Uint16(data[pos:])
			count := int(binary.LittleEndian.Uint16(data[pos+2:]))
			pos += 4

			for i := 0; i < count && x < width; i++ {
				off := y*img1555.Stride + x*2
				img1555.Pix[off] = byte(color16)
				img1555.Pix[off+1] = byte(color16 >> 8)
				x++
			}
		}
		if x != width {
			return nil, fmt.Errorf("scan-line %d decoded %d/%d pixels", y, x, width)
		}
	}

	return img1555, nil
}

// decodeGumpData converts raw gump data to an image.Image.
// Gumps are stored in a run-length encoded format with 16-bit color values.
func decodeGumpData2(data []byte, width, height int) (image.Image, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("%w: gump data too short for lookup table", ErrInvalidArtData)
	}

	// Create a new bitmap to hold the decoded image
	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))

	// Read lookup pointers for each line
	lookupTable := make([]int, height)
	for y := 0; y < height; y++ {
		// Each lookup is a 4-byte offset from the start of the data
		if (y*4)+4 > len(data) {
			return nil, fmt.Errorf("%w: gump data truncated in lookup table at line %d", ErrInvalidArtData, y)
		}
		lookupTable[y] = int(binary.LittleEndian.Uint32(data[y*4 : y*4+4]))
	}

	// Process each line
	for y := 0; y < height; y++ {
		x := 0 // Current x position in the output image

		offset := lookupTable[y]
		if offset < 0 || offset >= len(data) {
			return nil, fmt.Errorf("%w: invalid lookup offset %d for line %d", ErrInvalidArtData, offset, y)
		}

		// Process RLE data for this line
		for x < width {
			// Need at least 4 more bytes for an RLE pair (2 bytes color + 2 bytes run length)
			if offset+4 > len(data) {
				return nil, fmt.Errorf("%w: gump data truncated during RLE decoding at y=%d, x=%d", ErrInvalidArtData, y, x)
			}

			// Read color and run length
			colorValue := binary.LittleEndian.Uint16(data[offset : offset+2])
			offset += 2
			runLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
			offset += 2

			// 0,0 is the terminator for the line
			if colorValue == 0 && runLength == 0 {
				break
			}

			if colorValue == 0 {
				// Transparent run
				x += runLength
			} else {
				// Opaque run - flip the most significant bit to set alpha
				colorValue ^= 0x8000

				// Draw the pixels
				for i := 0; i < runLength; i++ {
					if x+i < width {
						img.Set(x+i, y, bitmap.ARGB1555Color(colorValue))
					}
				}
				x += runLength
			}
		}
	}

	return img, nil
}
