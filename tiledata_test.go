// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTileData(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		t.Run("LandTile", func(t *testing.T) {
			landTile, err := sdk.landInfo(0)

			assert.NoError(t, err)
			assert.NotEmpty(t, landTile.Name) // ED
			assert.NotZero(t, landTile.Flags)
		})

		t.Run("LandTile_InvalidID", func(t *testing.T) {
			_, err := sdk.landInfo(-1)
			assert.Error(t, err)

			_, err = sdk.landInfo(0xFFFFF)
			assert.Error(t, err)
		})

		t.Run("StaticTile", func(t *testing.T) {
			staticTile, err := sdk.staticInfo(3)

			assert.NoError(t, err)
			assert.NotEmpty(t, staticTile.Name)
		})

		t.Run("StaticTile_InvalidID", func(t *testing.T) {
			_, err := sdk.staticInfo(-1)
			assert.Error(t, err)

			_, err = sdk.staticInfo(sdk.staticTileCount())
			assert.Error(t, err)
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
