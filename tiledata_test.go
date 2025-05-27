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

	t.Run("ItemInfo_ContextualMethods", func(t *testing.T) {
		// Test a mock weapon item
		weaponItem := ItemInfo{
			Flags:    TileFlagWeapon,
			Quantity: 5, // Weapon class
		}

		// Test IsWeapon returns both value and bool
		weaponClass, isWeapon := weaponItem.IsWeapon()
		assert.True(t, isWeapon)
		assert.Equal(t, byte(5), weaponClass)

		// Test IsArmor returns false for weapon
		_, isArmor := weaponItem.IsArmor()
		assert.False(t, isArmor)

		// Test a mock wearable item
		wearableItem := ItemInfo{
			Flags:   TileFlagWearable,
			Quality: 10, // Layer
		}

		// Test IsWearable returns both value and bool
		layer, isWearable := wearableItem.IsWearable()
		assert.True(t, isWearable)
		assert.Equal(t, byte(10), layer)

		// Test convenience Layer() method
		assert.Equal(t, byte(10), wearableItem.Layer())

		// Test a mock light source
		lightItem := ItemInfo{
			Flags:   TileFlagLightSource,
			Quality: 3, // Light ID
		}

		// Test IsLightSource returns both value and bool
		lightID, isLight := lightItem.IsLightSource()
		assert.True(t, isLight)
		assert.Equal(t, byte(3), lightID)

		// Test convenience LightID() method
		assert.Equal(t, byte(3), lightItem.LightID())

		// Test item with no special flags
		normalItem := ItemInfo{
			Flags:    TileFlagNone,
			Quality:  1,
			Quantity: 2,
		}

		_, isWeapon = normalItem.IsWeapon()
		assert.False(t, isWeapon)

		_, isWearable = normalItem.IsWearable()
		assert.False(t, isWearable)

		_, isLight = normalItem.IsLightSource()
		assert.False(t, isLight)

		// Test StackQuantity always returns Quantity
		assert.Equal(t, byte(2), normalItem.StackQuantity())
	})

}
