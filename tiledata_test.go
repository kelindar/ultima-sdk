// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTileData(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		t.Run("LandTile", func(t *testing.T) {
			landTile, err := sdk.LandTile(0)

			assert.NoError(t, err)
			assert.NotEmpty(t, landTile.Name) // ED
			assert.NotZero(t, landTile.Flags)
		})

		t.Run("LandTile_InvalidID", func(t *testing.T) {
			_, err := sdk.LandTile(-1)
			assert.Error(t, err)

			_, err = sdk.LandTile(0xFFFFF)
			assert.Error(t, err)
		})

		t.Run("StaticTile", func(t *testing.T) {
			staticTile, err := sdk.StaticTile(3)

			assert.NoError(t, err)
			assert.NotEmpty(t, staticTile.Name)
		})

		t.Run("StaticTile_InvalidID", func(t *testing.T) {
			_, err := sdk.StaticTile(-1)
			assert.Error(t, err)

			_, err = sdk.StaticTile(sdk.staticTileCount())
			assert.Error(t, err)
		})

		t.Run("LandTiles_Iterator", func(t *testing.T) {
			count := 0
			for tile := range sdk.LandTiles() {
				if tile.Name != "" {
					count++
				}
				if count >= 5 {
					break
				}
			}
			assert.Equal(t, 5, count)
		})

		t.Run("StaticTiles_Iterator", func(t *testing.T) {
			count := 0
			for tile := range sdk.StaticTiles() {
				if tile.Name != "" {
					count++
				}
				if count >= 5 {
					break
				}
			}
			assert.Equal(t, 5, count)
		})

		t.Run("StaticItemData_Properties", func(t *testing.T) {
			// Test some properties that are commonly used

			// Find a known bridge tile to test CalcHeight
			var bridgeTile StaticItemData
			bridgeFound := false

			for tile := range sdk.StaticTiles() {
				if tile.Flags&TileFlagBridge != 0 {
					bridgeTile = tile
					bridgeFound = true
					break
				}
			}

			if bridgeFound {
				assert.Equal(t, int(bridgeTile.Height)/2, bridgeTile.CalcHeight())
			}

			// Test flag helper methods on some tile
			staticTile, err := sdk.StaticTile(1)
			require.NoError(t, err)

			assert.Equal(t, staticTile.Flags&TileFlagBackground != 0, staticTile.Background())
			assert.Equal(t, staticTile.Flags&TileFlagBridge != 0, staticTile.Bridge())
			assert.Equal(t, staticTile.Flags&TileFlagImpassable != 0, staticTile.Impassable())
			assert.Equal(t, staticTile.Flags&TileFlagSurface != 0, staticTile.Surface())
			assert.Equal(t, staticTile.Flags&TileFlagWearable != 0, staticTile.Wearable())
		})

		t.Run("HeightTable", func(t *testing.T) {
			heights, err := sdk.HeightTable()

			assert.NoError(t, err)
			assert.NotNil(t, heights)
			assert.Len(t, heights, sdk.staticTileCount())
		})
	})
}

// Test for the helper functions
func TestTileData_Helpers(t *testing.T) {
	t.Run("readStringFromBytes", func(t *testing.T) {
		// Test with null terminator
		input := []byte{'T', 'e', 's', 't', 0, 'X', 'Y', 'Z'}
		result := readStringFromBytes(input)
		assert.Equal(t, "Test", result)

		// Test without null terminator (uses full slice)
		input = []byte{'T', 'e', 's', 't'}
		result = readStringFromBytes(input)
		assert.Equal(t, "Test", result)

		// Test with empty input
		input = []byte{}
		result = readStringFromBytes(input)
		assert.Equal(t, "", result)
	})

}
