package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGump(t *testing.T) {
	t.Run("LoadGump", func(t *testing.T) {
		runWith(t, func(sdk *SDK) {
			// Test loading a specific gump (ID 1, which typically exists in most UO clients)
			gump, err := sdk.Gump(7)
			require.NoError(t, err)
			require.NotNil(t, gump)

			// Check that the gump has valid properties
			assert.Equal(t, 7, gump.ID)
			assert.Greater(t, gump.Width, 0, "Gump width should be greater than 0")
			assert.Greater(t, gump.Height, 0, "Gump height should be greater than 0")

			// Test the image lazy loading
			img, err := gump.Image()
			require.NoError(t, err)
			require.NotNil(t, img)

			// Image dimensions should match expected values
			assert.Equal(t, gump.Width, img.Bounds().Dx())
			assert.Equal(t, gump.Height, img.Bounds().Dy())

			//assert.NoError(t, savePng(img, "gump.png")
		})
	})

	t.Run("GumpInfos", func(t *testing.T) {
		runWith(t, func(sdk *SDK) {
			// Test iterating through gump infos
			count := 0
			for info := range sdk.GumpInfos() {
				// Each info should have a valid ID and dimensions
				assert.GreaterOrEqual(t, info.ID, 0, "Gump ID should be non-negative")
				assert.Greater(t, info.Width, 0, "Gump width should be greater than 0")
				assert.Greater(t, info.Height, 0, "Gump height should be greater than 0")

				// Limit the number of iterations to avoid too long test
				count++
				if count >= 100 {
					break
				}
			}

			// Make sure we found at least some gumps
			assert.Greater(t, count, 0, "Should find at least some gumps")
		})
	})

	t.Run("InvalidGump", func(t *testing.T) {
		runWith(t, func(sdk *SDK) {
			// Test requesting a very high gump ID that shouldn't exist
			gump, err := sdk.Gump(0xFFFFF) // Very high ID
			assert.Error(t, err, "Should error on invalid gump ID")
			assert.Nil(t, gump, "Should not return a gump for invalid ID")
		})
	})
}
