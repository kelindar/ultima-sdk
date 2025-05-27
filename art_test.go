// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArt(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		t.Run("LandArt", func(t *testing.T) {
			tile, err := sdk.Land(0)

			assert.NoError(t, err)
			assert.NotNil(t, tile)
			assert.Equal(t, 0, tile.ID)
			assert.NotEmpty(t, tile.Name)
			assert.NotEqual(t, uint64(0), tile.Flags)

			assert.NoError(t, err)
			assert.NotNil(t, tile.Image)
			assert.NoError(t, savePng(tile.Image, "test/art_land.png"))
		})

		t.Run("StaticArt", func(t *testing.T) {
			tile, err := sdk.Item(0x0E3D)

			assert.NoError(t, err)
			assert.NotNil(t, tile)
			assert.Equal(t, 0x0E3D+0x4000, tile.ID)
			assert.NotEmpty(t, tile.Name)

			assert.NoError(t, err)
			assert.NotNil(t, tile.Image)
			assert.NoError(t, savePng(tile.Image, "test/art_static.png"))
		})

		t.Run("LandArtImage", func(t *testing.T) {
			// Test loading and decoding a land art image
			tile, err := sdk.Land(100)
			require.NoError(t, err)
			assert.NotNil(t, tile.Image)

			// Land art tiles should always be 44x44
			bounds := tile.Image.Bounds()
			assert.Equal(t, 44, bounds.Dx())
			assert.Equal(t, 44, bounds.Dy())
		})

		t.Run("StaticArtImage", func(t *testing.T) {
			// Test loading and decoding a static art image
			tile, err := sdk.Item(8000)
			assert.NoError(t, err)
			assert.NotNil(t, tile.Image)

			// Static art tiles can have various dimensions, just make sure they're reasonable
			bounds := tile.Image.Bounds()
			assert.Greater(t, bounds.Dx(), 0)
			assert.Greater(t, bounds.Dy(), 0)
			assert.Less(t, bounds.Dx(), 1024) // Reasonable size limit
			assert.Less(t, bounds.Dy(), 1024) // Reasonable size limit
		})

		t.Run("InvalidIDs", func(t *testing.T) {

			// Test invalid land ID
			_, err := sdk.Land(-1)
			assert.Error(t, err)

			// Test land ID too large
			_, err = sdk.Land(0x4000)
			assert.Error(t, err)

		})

		t.Run("LandArtTiles_Iterator", func(t *testing.T) {
			// Test the land art iterator
			counter := 0
			for tile := range sdk.Lands() {
				assert.NotNil(t, tile)
				assert.NotNil(t, tile.Art, "ArtTile should not be nil")
				assert.Less(t, tile.ID, 0x4000)

				// Just check the first few to keep test runtime reasonable
				if counter++; counter >= 10 {
					break
				}
			}
			assert.Greater(t, counter, 0, "Expected at least one land art tile")
		})

		t.Run("StaticArtTiles_Iterator", func(t *testing.T) {
			// Test the static art iterator
			counter := 0
			for tile := range sdk.Items() {
				assert.NotNil(t, tile)
				assert.NotNil(t, tile.Art, "ArtTile should not be nil")
				assert.GreaterOrEqual(t, tile.ID, 0x4000)

				// Just check the first few to keep test runtime reasonable
				if counter++; counter >= 10 {
					break
				}
			}
			assert.Greater(t, counter, 0, "Expected at least one static art tile")
		})
	})
}

// TestInvalidImageData tests decoding functions with various bad data
func TestInvalidImageData(t *testing.T) {
	t.Run("InvalidLandArt", func(t *testing.T) {
		// Test decoding land art with invalid data

		// Too short
		_, err := decodeLandImage([]byte{1, 2})
		assert.Error(t, err)

		// Valid header but truncated data
		shortData := make([]byte, 100) // Not enough for a 44x44 land tile
		_, err = decodeLandImage(shortData)
		assert.Error(t, err)
	})

	t.Run("InvalidStaticArt", func(t *testing.T) {
		// Test decoding static art with invalid data

		// Too short
		_, err := decodeStaticImage([]byte{1, 2, 3, 4})
		assert.Error(t, err)

		// Invalid dimensions (too large)
		badDimensions := []byte{
			0, 0, 0, 0, // header
			0xFF, 0xFF, // width = 65535 (too large)
			0xFF, 0xFF, // height = 65535 (too large)
		}
		_, err = decodeStaticImage(badDimensions)
		assert.Error(t, err)

		// Valid header but truncated lookup table
		badLookup := []byte{
			0, 0, 0, 0, // header
			10, 0, // width = 10
			10, 0, // height = 10
			0, 0, // First lookup entry
			// Missing rest of lookup table
		}
		_, err = decodeStaticImage(badLookup)
		assert.Error(t, err)
	})
}
