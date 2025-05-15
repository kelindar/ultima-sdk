package anim

import (
	"encoding/binary"
	"fmt"
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
	
	// Debug output
	fmt.Printf("Frame header: center=(%d,%d), size=%dx%d\n", xCenter, yCenter, width, height)

	if height == 0 || width == 0 {
		return image.Point{xCenter, yCenter}, nil, nil
	}

	// Create image with the right dimensions
	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))
	
	// Constants used for bit manipulation, same as C#
	const doubleXor = (0x200 << 22) | (0x200 << 12)
	
	// Calculate base X position like C# does
	xBase := xCenter - 0x200
	// C# calculates yBase = (yCenter + height) - 0x200
	yBase := (yCenter + height) - 0x200
	
	// Start after the frame header
	offset := 8
	
	// Process pixel runs until we find the terminator (0x7FFF7FFF)
	for offset+4 <= len(data) {
		// Read header for this run
		header := int(binary.LittleEndian.Uint32(data[offset:offset+4]))
		offset += 4
		
		if header == 0x7FFF7FFF {
			break // End of frame marker
		}
		
		// XOR with magic value to get the real coordinates
		header ^= doubleXor
		
		// The coordinates and run length are packed in the header
		runY := (header >> 12) & 0x3FF // Y position of this run
		runX := (header >> 22) & 0x3FF // X position of this run
		runLen := header & 0xFFF       // Length of this run
		
		if !flip {
			// Normal direction (left to right)
			pixelY := yBase + runY
			pixelX := runX + xBase
			
			// End position for this run
			endX := pixelX + runLen
			
			// Read and plot each pixel in the run
			for x := pixelX; x < endX && offset < len(data); x++ {
				paletteIdx := int(data[offset])
				offset++
				
				// Plot only if within the image bounds
				if x >= 0 && x < width && pixelY >= 0 && pixelY < height {
					pix := bitmap.ARGB1555Color(palette[paletteIdx])
					img.Set(x, pixelY, pix)
				}
			}
		} else {
			// Flipped direction (right to left)
			// In C# flipped images use a different coordinate calculation
			pixelY := yBase + runY
			// Notice how width is used here to flip coordinates
			pixelX := width - 1 - (runX + xBase)
			
			// End position (going left)
			endX := pixelX - runLen
			
			// Read and plot each pixel in the run (right to left)
			for x := pixelX; x > endX && offset < len(data); x-- {
				paletteIdx := int(data[offset])
				offset++
				
				// Plot only if within bounds
				if x >= 0 && x < width && pixelY >= 0 && pixelY < height {
					pix := bitmap.ARGB1555Color(palette[paletteIdx])
					img.Set(x, pixelY, pix)
				}
			}
		}
	}
	
	// For flipped images, adjust the center X position
	if flip {
		xCenter = width - xCenter
	}
	
	return image.Point{xCenter, yCenter}, img, nil
}
