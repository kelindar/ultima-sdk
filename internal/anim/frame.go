package anim

import (
	"encoding/binary"
	"image"
	"github.com/kelindar/ultima-sdk/internal/bitmap"
)

// DecodeFrame decodes a single animation frame from a binary slice using the provided palette.
// Returns center (image.Point), ARGB1555 image, and error if any.
func DecodeFrame(palette []uint16, data []byte, flip bool) (image.Point, *bitmap.ARGB1555, error) {
	if len(data) < 8 {
		return image.Point{}, nil, nil // Not enough data for header
	}
	// Read center and dimensions
	xCenter := int(int16(binary.LittleEndian.Uint16(data[0:2])))
	yCenter := int(int16(binary.LittleEndian.Uint16(data[2:4])))
	width := int(binary.LittleEndian.Uint16(data[4:6]))
	height := int(binary.LittleEndian.Uint16(data[6:8]))
	if width == 0 || height == 0 {
		return image.Point{xCenter, yCenter}, nil, nil
	}

	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))
	offset := 8
	const doubleXor = (0x200 << 22) | (0x200 << 12)
	for offset+4 <= len(data) {
		header := int(int32(binary.LittleEndian.Uint32(data[offset : offset+4])))
		offset += 4
		if header == 0x7FFF7FFF {
			break
		}
		header ^= doubleXor
		// Decode run
		line := 0
		if !flip {
			line = ((header >> 12) & 0x3FF)
		} else {
			line = ((header >> 12) & 0x3FF)
		}
		col := ((header >> 22) & 0x3FF)
		count := header & 0xFFF
		if !flip {
			for i := 0; i < count && offset < len(data); i++ {
				paletteIdx := int(data[offset])
				offset++
				if line < height && col+i < width {
					pix := bitmap.ARGB1555Color(palette[paletteIdx])
					img.Set(col+i, line, pix)
				}
			}
		} else {
			for i := 0; i < count && offset < len(data); i++ {
				paletteIdx := int(data[offset])
				offset++
				if line < height && col-i >= 0 {
					pix := bitmap.ARGB1555Color(palette[paletteIdx])
					img.Set(col-i, line, pix)
				}
			}
		}
	}
	return image.Point{xCenter, yCenter}, img, nil
}
