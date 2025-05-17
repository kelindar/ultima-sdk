package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLight(t *testing.T) {

	t.Run("GetRawLightNegativeID", func(t *testing.T) {
		runWith(t, func(sdk *SDK) {
			_, _, _, err := sdk.GetRawLight(-1)
			assert.Error(t, err)
		})
	})

	runWith(t, func(sdk *SDK) {
		t.Run("Light_Valid", func(t *testing.T) {
			// Test retrieving a valid light (assuming light ID 0 exists in test data)
			light, err := sdk.Light(0)

			assert.NoError(t, err, "Light(0) should not return an error")
			assert.Greater(t, light.Width, 0, "Light width should be positive")
			assert.Greater(t, light.Height, 0, "Light height should be positive")
			assert.Equal(t, 0, light.ID, "Light ID should match requested ID")
			//assert.NoError(t, savePng(light.Image(), "light.png"), "Failed to save light image")
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

		t.Run("GetRawLight", func(t *testing.T) {
			// Test getting raw light data
			data, width, height, err := sdk.GetRawLight(0)

			assert.NoError(t, err, "GetRawLight(0) should not return an error")
			assert.NotNil(t, data, "Raw light data should not be nil")
			assert.Greater(t, width, 0, "Light width should be positive")
			assert.Greater(t, height, 0, "Light height should be positive")
			assert.Equal(t, len(data), width*height, "Data length should match dimensions")
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
