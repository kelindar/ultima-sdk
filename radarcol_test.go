// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"fmt"
	"testing"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
	"github.com/stretchr/testify/assert"
)

func TestRadarColor(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Test retrieving a land tile color
		color, err := sdk.RadarColor(16) // Land tile (ID < 0x4000)
		assert.NoError(t, err)
		assert.Equal(t, 16, color.ID())
		assert.False(t, color.IsStatic())
		assert.True(t, color.IsLand())
		assert.NotEqual(t, uint16(0), color.Value()) // The color should be non-zero
		assert.Equal(t, uint16(12549), color.Value())

		// Verify GetColor works correctly
		goColor := color.GetColor()
		assert.NotNil(t, goColor)
		colorARGB, ok := goColor.(bitmap.ARGB1555Color)
		assert.True(t, ok)
		// Alpha bit should be set in the returned color
		assert.Equal(t, color.Value()|0x8000, uint16(colorARGB))

		// Test retrieving a static tile color
		staticID := 0x4001 // Static tile (ID >= 0x4000)
		color, err = sdk.RadarColor(staticID)
		assert.NoError(t, err)
		assert.Equal(t, staticID, color.ID())
		assert.True(t, color.IsStatic())
		assert.False(t, color.IsLand())

		// Test invalid index - negative
		_, err = sdk.RadarColor(-1)
		assert.Error(t, err)

		// Test invalid index - too large
		_, err = sdk.RadarColor(0x8000)
		assert.Error(t, err)
	})
}

func TestRadarColorsIterators(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		t.Run("RadarColors", func(t *testing.T) {
			landCount := 0
			staticCount := 0
			maxToCheck := 20 // Check a few entries

			// Iterate through all radar colors with proper for...range pattern
			count := 0
			for color := range sdk.RadarColors() {
				// Increment the appropriate counter based on the type
				if color.IsLand() {
					landCount++
				} else {
					staticCount++
				}

				count++
				if count >= maxToCheck {
					break
				}
			}

			assert.Greater(t, count, 0, "Should have iterated through at least one radar color")
		})
	})
}

// Test the bit packing and extracting logic
func TestRadarColorBitpacking(t *testing.T) {
	tests := []struct {
		id    int
		value uint16
	}{
		{1, 0x7C00},      // Land tile with red color
		{100, 0x03E0},    // Land tile with green color
		{1000, 0x001F},   // Land tile with blue color
		{0x4001, 0x7FFF}, // Static tile with white color
		{0x4100, 0x7C1F}, // Static tile with magenta color
		{0x5000, 0x03FF}, // Static tile with cyan color
		{0x3FFF, 0x0000}, // Max land tile ID with black color
		{0x7FFF, 0x0000}, // Max static tile ID with black color
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("ID=0x%X,Value=0x%04X", tt.id, tt.value), func(t *testing.T) {
			// Create the RadarColor
			rc := makeRadarColor(tt.id, tt.value)

			// Verify the extracted values
			assert.Equal(t, tt.id, rc.ID(), "ID extraction failed")
			assert.Equal(t, tt.value, rc.Value(), "Value extraction failed")

			// Verify type detection based on ID
			if tt.id < 0x4000 {
				assert.True(t, rc.IsLand(), "Should be detected as land tile")
				assert.False(t, rc.IsStatic(), "Should not be detected as static tile")
			} else {
				assert.True(t, rc.IsStatic(), "Should be detected as static tile")
				assert.False(t, rc.IsLand(), "Should not be detected as land tile")
			}

			// Verify GetColor works correctly
			goColor := rc.GetColor()
			colorARGB, ok := goColor.(bitmap.ARGB1555Color)
			assert.True(t, ok)

			// Alpha bit should be set in the returned color
			assert.Equal(t, tt.value|0x8000, uint16(colorARGB))
		})
	}
}
