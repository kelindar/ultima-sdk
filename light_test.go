// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLight(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		t.Run("Light_Valid", func(t *testing.T) {
			// Test retrieving a valid light (assuming light ID 0 exists in test data)
			light, err := sdk.Light(0)

			assert.NoError(t, err, "Light(0) should not return an error")
			assert.Greater(t, light.Width, 0, "Light width should be positive")
			assert.Greater(t, light.Height, 0, "Light height should be positive")
			assert.Equal(t, 0, light.ID, "Light ID should match requested ID")
			assert.NoError(t, savePng(light.Image(), "test/light.png"))
			assert.NotNil(t, light.Image())
		})

		t.Run("Light_InvalidID", func(t *testing.T) {
			// Test with negative ID - should return error
			_, err := sdk.Light(-1)
			assert.Error(t, err, "Light(-1) should return an error")

			// Test with very large ID that likely doesn't exist
			_, err = sdk.Light(9999)
			assert.Error(t, err, "Light(9999) should return an error for non-existent ID")
		})

		t.Run("Lights_Iterator", func(t *testing.T) {
			// Test iterating through lights
			count := 0
			for light := range sdk.Lights() {
				assert.NotNil(t, light.Image, "Light image should not be nil")
				assert.GreaterOrEqual(t, light.ID, 0, "Light ID should be non-negative")
				assert.Greater(t, light.Width, 0, "Light width should be positive")
				assert.Greater(t, light.Height, 0, "Light height should be positive")

				count++
				if count >= 5 { // Limit to 5 lights for test speed
					break
				}
			}
			assert.Greater(t, count, 0, "Should have found at least one light")
		})
	})
}
