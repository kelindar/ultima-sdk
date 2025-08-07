// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"encoding/binary"
	"fmt"
	"image"
	"iter"
	"math"

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

	if extra < math.MaxUint32 {
		width = int(extra & 0xFFFF)
		height = int((extra >> 16) & 0xFFFF)
	}

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
