// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
	"iter"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
)

var (
	ErrInvalidRadarColorIndex = errors.New("invalid radar color index")
)

const (
	radarColorEntries      = 0x4000
	totalRadarColors       = 0x8000
	radarColorStaticOffset = 0x4000

	// Bit manipulation constants for RadarColor type
	radarColorIDMask     = 0xFFFFFFFF     // Bits 0-31: Tile ID
	radarColorValueMask  = 0xFFFF00000000 // Bits 32-47: Color value
	radarColorValueShift = 32             // Shift for Color value
)

// RadarColor is a bit-packed uint64 containing a tile ID and color value
// Bits 0-31: Tile ID (uint32)
// Bits 32-47: Color value (uint16)
// Bits 48-63: Unused (reserved for future)
type RadarColor uint64

// ID returns the tile ID component
func (r RadarColor) ID() int {
	return int(uint64(r) & radarColorIDMask)
}

// Value returns the raw 16-bit color value (ARGB1555)
func (r RadarColor) Value() uint16 {
	return uint16((uint64(r) & radarColorValueMask) >> radarColorValueShift)
}

// IsStatic returns true if this represents a static tile color
func (r RadarColor) IsStatic() bool {
	return r.ID() >= radarColorStaticOffset
}

// IsLand returns true if this represents a land tile color
func (r RadarColor) IsLand() bool {
	return r.ID() < radarColorStaticOffset
}

// GetColor returns a standard Go color.Color from the radar color
func (r RadarColor) GetColor() color.Color {
	// Set the alpha bit to 1 for opaque colors
	colorValue := r.Value() | 0x8000
	return bitmap.ARGB1555Color(colorValue)
}

// makeRadarColor creates a bit-packed RadarColor value
func makeRadarColor(id int, value uint16) RadarColor {
	result := uint64(id) & radarColorIDMask
	result |= (uint64(value) << radarColorValueShift) & radarColorValueMask
	return RadarColor(result)
}

// RadarColor retrieves the radar color for a given tile ID
func (s *SDK) RadarColor(tileID int) (RadarColor, error) {
	if tileID < 0 || tileID >= totalRadarColors {
		return 0, fmt.Errorf("%w: %d (must be between 0 and 0x7FFF)", ErrInvalidRadarColorIndex, tileID)
	}

	file, err := s.loadRadarcol()
	if err != nil {
		return 0, fmt.Errorf("failed to load radar colors: %w", err)
	}

	entry, err := file.Entry(0)
	if err != nil {
		return 0, err
	}

	data := make([]byte, 2)
	if _, err := entry.ReadAt(data, int64(tileID)*2); err != nil {
		return 0, err
	}

	return makeRadarColor(tileID, binary.LittleEndian.Uint16(data)), nil
}

// RadarColors returns an iterator over all defined radar color mappings
func (s *SDK) RadarColors() iter.Seq[RadarColor] {
	return func(yield func(RadarColor) bool) {
		file, err := s.loadRadarcol()
		if err != nil {
			return
		}

		data, err := file.ReadFull(0)
		if err != nil {
			return
		}

		entryCount := len(data) / 2
		if entryCount > totalRadarColors {
			entryCount = totalRadarColors
		}

		for i := 0; i < entryCount; i++ {
			color := binary.LittleEndian.Uint16(data[i*2:])
			radarColor := makeRadarColor(i, color)

			if !yield(radarColor) {
				break
			}
		}
	}
}
