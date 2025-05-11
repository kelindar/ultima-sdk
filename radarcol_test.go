package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRadarColor(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Test retrieving a land tile color - using index 1 which should exist in most UO clients
		color, err := sdk.RadarColor(1)
		assert.NoError(t, err)
		assert.NotEqual(t, uint16(0), color) // The color should be non-zero

		// Test retrieving a land tile at a higher index
		color, err = sdk.RadarColor(100)
		assert.NoError(t, err)
		// No specific value to assert, just make sure it returns without error

		// Test an index at the upper boundary of land tiles
		color, err = sdk.RadarColor(0x3FFF)
		// This could return an error if the file doesn't have this many entries,
		// so we don't assert the error condition, but if it succeeds, the color should be valid
		if err == nil {
			assert.NotEqual(t, uint16(0), color)
		}

		// Test an invalid index - negative
		_, err = sdk.RadarColor(-1)
		assert.Error(t, err)

		// Test an invalid index - too large
		_, err = sdk.RadarColor(0x4000)
		assert.Error(t, err)
	})
}

func TestRadarColorStatic(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Test retrieving a static tile color - using index 1 which should exist in most UO clients
		color, err := sdk.RadarColorStatic(1)
		assert.NoError(t, err)
		assert.NotEqual(t, uint16(0), color) // The color should be non-zero

		// Test retrieving a static tile at a higher index
		color, err = sdk.RadarColorStatic(100)
		assert.NoError(t, err)
		// No specific value to assert, just make sure it returns without error

		// Test an index at the upper boundary of static tiles
		color, err = sdk.RadarColorStatic(0x3FFF)
		// This could return an error if the file doesn't have this many entries,
		// so we don't assert the error condition, but if it succeeds, the color should be valid
		if err == nil {
			assert.NotEqual(t, uint16(0), color)
		}

		// Test an invalid index - negative
		_, err = sdk.RadarColorStatic(-1)
		assert.Error(t, err)

		// Test an invalid index - too large
		_, err = sdk.RadarColorStatic(0x4000)
		assert.Error(t, err)
	})
}

func TestRadarColorsIterators(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Test RadarColorsLand iterator
		t.Run("RadarColorsLand", func(t *testing.T) {
			count := 0
			maxToCheck := 10 // Just check a few entries

			// Iterate through land radar colors with proper for...range pattern
			for tileID, color := range sdk.RadarColorsLand() {
				// Verify the ID is within valid range for land tiles
				assert.GreaterOrEqual(t, tileID, 0)
				assert.Less(t, tileID, 0x4000)
				assert.NotEqual(t, uint16(0), color, "Color should be non-zero")

				count++
				if count >= maxToCheck {
					break
				}
			}

			// Make sure we iterated through some items
			assert.Greater(t, count, 0, "Should have iterated through at least one land tile color")
		})

		// Test RadarColorsStatic iterator
		t.Run("RadarColorsStatic", func(t *testing.T) {
			count := 0
			maxToCheck := 10 // Just check a few entries

			// Iterate through static radar colors with proper for...range pattern
			for tileID, color := range sdk.RadarColorsStatic() {
				// Verify the ID is within valid range for static tiles
				assert.GreaterOrEqual(t, tileID, 0)
				assert.Less(t, tileID, 0x4000)
				assert.NotEqual(t, uint16(0), color, "Color should be non-zero")

				count++
				if count >= maxToCheck {
					break
				}
			}

			// Make sure we iterated through some items
			assert.Greater(t, count, 0, "Should have iterated through at least one static tile color")
		})

		// Test combined RadarColors iterator
		t.Run("RadarColors", func(t *testing.T) {
			landCount := 0
			staticCount := 0
			maxToCheck := 20 // Check a few entries

			// Iterate through all radar colors with proper for...range pattern
			count := 0
			for _, color := range sdk.RadarColors() {
				assert.NotEqual(t, uint16(0), color, "Color should be non-zero")

				// Keep track of land vs static tiles based on the order they appear
				if count < 10 {
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

// TestSDKLoadRadarcol tests that the loadRadarcol function works correctly
func TestSDKLoadRadarcol(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Load the radar color file
		file, err := sdk.loadRadarcol()
		assert.NoError(t, err, "Should be able to load the radar color file")
		assert.NotNil(t, file, "Radar color file should not be nil")

		// Read a chunk of data to verify we can access the file
		data, err := file.Read(0)
		assert.NoError(t, err, "Should be able to read from the radar color file")
		assert.NotEmpty(t, data, "Radar color data should not be empty")
	})
}

// Test radar colors for specific tile IDs
func TestRadarColorConsistency(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Test that getting the same tile ID through different methods returns the same color
		t.Run("ConsistentColorRetrieval", func(t *testing.T) {
			// Choose a valid tile ID for testing
			tileID := 1

			// Get the color directly with RadarColor
			directColor, err := sdk.RadarColor(tileID)
			assert.NoError(t, err)

			// Get the color via the iterator
			var iteratorColor uint16
			found := false

			// Use the proper for...range pattern for iteration
			for id, color := range sdk.RadarColorsLand() {
				if id == tileID {
					iteratorColor = color
					found = true
					break
				}
			}

			assert.True(t, found, "Should have found the tile ID in the iterator")
			assert.Equal(t, directColor, iteratorColor, "Color should be consistent between direct access and iterator")
		})

		// Same test but for static tiles
		t.Run("ConsistentStaticColorRetrieval", func(t *testing.T) {
			// Choose a valid static tile ID for testing
			tileID := 1

			// Get the color directly with RadarColorStatic
			directColor, err := sdk.RadarColorStatic(tileID)
			assert.NoError(t, err)

			// Get the color via the iterator
			var iteratorColor uint16
			found := false

			// Use the proper for...range pattern for iteration
			for id, color := range sdk.RadarColorsStatic() {
				if id == tileID {
					iteratorColor = color
					found = true
					break
				}
			}

			assert.True(t, found, "Should have found the static tile ID in the iterator")
			assert.Equal(t, directColor, iteratorColor, "Static tile color should be consistent between direct access and iterator")
		})
	})
}
