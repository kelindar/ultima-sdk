// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"iter"
	"strings"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
)

var (
	// ErrInvalidHueIndex is returned when an invalid hue index is requested
	ErrInvalidHueIndex = errors.New("invalid hue index")
	// ErrInvalidPaletteIndex is returned when an invalid palette index is requested
	ErrInvalidPaletteIndex = errors.New("invalid palette index")
)

// Hue defines a color palette used for re-coloring game assets
type Hue struct {
	Index      int        // Index of this hue
	Name       string     // Name of this hue
	Colors     [32]uint16 // Raw 16-bit color values (ARGB1555)
	TableStart uint16     // Start index of the hue table
	TableEnd   uint16     // End index of the hue table
}

// GetColor returns a standard Go color.Color for a specific entry in the hue's palette
func (h *Hue) GetColor(paletteIndex int) (color.Color, error) {
	if paletteIndex < 0 || paletteIndex >= len(h.Colors) {
		return nil, fmt.Errorf("%w: palette index %d out of range", ErrInvalidPaletteIndex, paletteIndex)
	}

	// Convert the 16-bit color value to ARGB1555Color
	// Make sure we add the alpha bit since hue colors should be opaque
	colorValue := h.Colors[paletteIndex] | 0x8000 // Set the alpha bit to 1
	return bitmap.ARGB1555Color(colorValue), nil
}

// Image generates a small image.Image representing this hue's palette for visualization
func (h *Hue) Image(widthPerColor, height int) image.Image {
	width := widthPerColor * len(h.Colors)
	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))

	for i, colorValue := range h.Colors {
		// Set the alpha bit to 1 to make the color opaque
		colorValue |= 0x8000
		hueColor := bitmap.ARGB1555Color(colorValue)

		// Fill a rectangle for this color in the image
		startX := i * widthPerColor
		endX := startX + widthPerColor
		for y := 0; y < height; y++ {
			for x := startX; x < endX; x++ {
				img.Set(x, y, hueColor)
			}
		}
	}

	return img
}

// HueAt retrieves a specific hue by its index
func (s *SDK) HueAt(index int) (*Hue, error) {
	// Check for valid index range
	if index < 0 || index >= 3000 {
		return nil, fmt.Errorf("%w: %d (must be between 0 and 2999)", ErrInvalidHueIndex, index)
	}

	// Load the hues file
	file, err := s.loadHues()
	if err != nil {
		return nil, fmt.Errorf("failed to load hues: %w", err)
	}

	// With the chunk size set to 708 bytes, we need to calculate which block and which
	// entry within the block contains our hue
	blockIndex := index / 8
	entryIndex := index % 8

	// Each block contains 8 hues and starts with a 4-byte header
	// Each hue entry is (708 - 4) / 8 = 88 bytes:
	// - 32 colors * 2 bytes = 64 bytes
	// - TableStart (2 bytes)
	// - TableEnd (2 bytes)
	// - Name (20 bytes)

	// Read the entire block
	blockData, _, err := file.Read(uint32(blockIndex))
	if err != nil {
		return nil, fmt.Errorf("failed to read hue block: %w", err)
	}

	// Skip the 4-byte header and go to the correct entry
	entrySize := 88 // bytes
	entryOffset := 4 + (entryIndex * entrySize)

	// Create a reader for the entry
	if entryOffset+entrySize > len(blockData) {
		return nil, fmt.Errorf("invalid hue data: block %d too small, expected at least %d bytes but got %d",
			blockIndex, entryOffset+entrySize, len(blockData))
	}

	reader := bytes.NewReader(blockData[entryOffset:])

	// Create a new hue and read the data
	hue := &Hue{Index: index}

	// Read the 32 color values
	for i := 0; i < 32; i++ {
		if err := binary.Read(reader, binary.LittleEndian, &hue.Colors[i]); err != nil {
			return nil, fmt.Errorf("failed to read hue color: %w", err)
		}
	}

	// Read TableStart and TableEnd
	if err := binary.Read(reader, binary.LittleEndian, &hue.TableStart); err != nil {
		return nil, fmt.Errorf("failed to read hue TableStart: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &hue.TableEnd); err != nil {
		return nil, fmt.Errorf("failed to read hue TableEnd: %w", err)
	}

	// Read the 20-byte name string, null-terminated ASCII
	nameBytes := make([]byte, 20)
	if _, err := reader.Read(nameBytes); err != nil {
		return nil, fmt.Errorf("failed to read hue name: %w", err)
	}

	// Find the null terminator
	nullTermPos := bytes.IndexByte(nameBytes, 0)
	if nullTermPos != -1 {
		nameBytes = nameBytes[:nullTermPos]
	}

	// Convert to string and clean up any non-printable characters
	hue.Name = strings.Replace(string(nameBytes), "\n", " ", -1)
	hue.Name = strings.TrimSpace(hue.Name)

	// If the name is empty, use a default name
	if hue.Name == "" {
		hue.Name = fmt.Sprintf("Hue %d", index)
	}

	return hue, nil
}

// Hues returns an iterator over all available hues
func (s *SDK) Hues() iter.Seq[*Hue] {
	return func(yield func(*Hue) bool) {
		for i := 0; i < 3000; i++ {
			hue, err := s.HueAt(i)
			if err != nil {
				continue // Skip any hues that can't be loaded
			}

			if !yield(hue) {
				break
			}
		}
	}
}
