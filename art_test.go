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
			tile, err := sdk.LandArtTile(0)

			assert.NoError(t, err)
			assert.NotNil(t, tile)
			assert.Equal(t, 0, tile.ID)
			assert.True(t, tile.isLand)
			assert.NotEmpty(t, tile.Name)
			assert.NotEqual(t, uint64(0), tile.Flags)

			img, err := tile.Image()
			assert.NoError(t, err)
			assert.NotNil(t, img)
			//assert.NoError(t, savePng(img, "land.png"))
		})

		t.Run("StaticArt", func(t *testing.T) {
			tile, err := sdk.StaticArtTile(0x0E3D)

			assert.NoError(t, err)
			assert.NotNil(t, tile)
			assert.Equal(t, 0x0E3D+0x4000, tile.ID)
			assert.False(t, tile.isLand)
			assert.NotEmpty(t, tile.Name)

			img, err := tile.Image()
			assert.NoError(t, err)
			assert.NotNil(t, img)
			//assert.NoError(t, savePng(img, "static.png"))
		})

		t.Run("ArtTile_Land", func(t *testing.T) {
			// Test retrieving a land tile with the general-purpose ArtTile method
			tile, err := sdk.ArtTile(100)

			assert.NoError(t, err)
			assert.Equal(t, 100, tile.ID)
			assert.True(t, tile.isLand)
		})

		t.Run("ArtTile_Static", func(t *testing.T) {
			// Test retrieving a static tile with the general-purpose ArtTile method
			tile, err := sdk.ArtTile(0x4000 + 8000)

			assert.NoError(t, err)
			assert.Equal(t, 0x4000+8000, tile.ID)
			assert.False(t, tile.isLand)
		})

		t.Run("LandArtImage", func(t *testing.T) {
			// Test loading and decoding a land art image
			tile, err := sdk.LandArtTile(100)
			require.NoError(t, err)

			img, err := tile.Image()
			assert.NoError(t, err)
			assert.NotNil(t, img)

			// Land art tiles should always be 44x44
			bounds := img.Bounds()
			assert.Equal(t, 44, bounds.Dx())
			assert.Equal(t, 44, bounds.Dy())
		})

		t.Run("StaticArtImage", func(t *testing.T) {
			// Test loading and decoding a static art image
			tile, err := sdk.StaticArtTile(8000)
			require.NoError(t, err)

			img, err := tile.Image()
			assert.NoError(t, err)
			assert.NotNil(t, img)

			// Static art tiles can have various dimensions, just make sure they're reasonable
			bounds := img.Bounds()
			assert.Greater(t, bounds.Dx(), 0)
			assert.Greater(t, bounds.Dy(), 0)
			assert.Less(t, bounds.Dx(), 1024) // Reasonable size limit
			assert.Less(t, bounds.Dy(), 1024) // Reasonable size limit
		})

		t.Run("InvalidIDs", func(t *testing.T) {
			// Test negative ID
			_, err := sdk.ArtTile(-1)
			assert.Error(t, err)

			// Test invalid land ID
			_, err = sdk.LandArtTile(-1)
			assert.Error(t, err)

			// Test land ID too large
			_, err = sdk.LandArtTile(0x4000)
			assert.Error(t, err)

		})

		t.Run("LandArtTiles_Iterator", func(t *testing.T) {
			// Test the land art iterator
			counter := 0
			for tile := range sdk.LandArtTiles() {
				assert.NotNil(t, tile)
				assert.True(t, tile.isLand)
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
			for tile := range sdk.StaticArtTiles() {
				assert.NotNil(t, tile)
				assert.False(t, tile.isLand)
				assert.GreaterOrEqual(t, tile.ID, 0x4000)

				// Just check the first few to keep test runtime reasonable
				if counter++; counter >= 10 {
					break
				}
			}
			assert.Greater(t, counter, 0, "Expected at least one static art tile")
		})

		t.Run("NoImageData", func(t *testing.T) {
			// Test the behavior when an ArtTile has no image data
			tile := ArtTile{
				ID:        1000,
				imageData: nil, // No image data
			}

			_, err := tile.Image()
			assert.Error(t, err)
		})

		t.Run("RetrieveTwice", func(t *testing.T) {
			// Get the same tile twice, should not reprocess the image
			tile1, err := sdk.ArtTile(100)
			require.NoError(t, err)

			// Load the image
			img1, err := tile1.Image()
			require.NoError(t, err)

			// Get the same tile again
			tile2, err := sdk.ArtTile(100)
			require.NoError(t, err)

			// Load the image again
			img2, err := tile2.Image()
			require.NoError(t, err)

			// Images should not be the same object as they come from different ArtTile instances
			// But they should have the same dimensions
			assert.NotSame(t, img1, img2)
			assert.Equal(t, img1.Bounds(), img2.Bounds())
		})
	})
}

// TestInvalidImageData tests decoding functions with various bad data
func TestInvalidImageData(t *testing.T) {
	t.Run("InvalidLandArt", func(t *testing.T) {
		// Test decoding land art with invalid data

		// Too short
		_, err := decodeLandArt([]byte{1, 2})
		assert.Error(t, err)

		// Valid header but truncated data
		shortData := make([]byte, 100) // Not enough for a 44x44 land tile
		_, err = decodeLandArt(shortData)
		assert.Error(t, err)
	})

	t.Run("InvalidStaticArt", func(t *testing.T) {
		// Test decoding static art with invalid data

		// Too short
		_, err := decodeStaticArt([]byte{1, 2, 3, 4})
		assert.Error(t, err)

		// Invalid dimensions (too large)
		badDimensions := []byte{
			0, 0, 0, 0, // header
			0xFF, 0xFF, // width = 65535 (too large)
			0xFF, 0xFF, // height = 65535 (too large)
		}
		_, err = decodeStaticArt(badDimensions)
		assert.Error(t, err)

		// Valid header but truncated lookup table
		badLookup := []byte{
			0, 0, 0, 0, // header
			10, 0, // width = 10
			10, 0, // height = 10
			0, 0, // First lookup entry
			// Missing rest of lookup table
		}
		_, err = decodeStaticArt(badLookup)
		assert.Error(t, err)
	})
}
