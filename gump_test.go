// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGump(t *testing.T) {
	t.Run("NegativeID", func(t *testing.T) {
		runWith(t, func(sdk *SDK) {
			gump, err := sdk.Gump(-1)
			assert.Error(t, err)
			assert.Nil(t, gump)
		})
	})

	t.Run("IteratorExhaustion", func(t *testing.T) {
		runWith(t, func(sdk *SDK) {
			count := 0
			for range sdk.Gumps() {
				count++
			}
			assert.Greater(t, count, 0)
		})
	})

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
			assert.NotNil(t, gump.Image)

			// Image dimensions should match expected values
			assert.Equal(t, gump.Width, gump.Image.Bounds().Dx())
			assert.Equal(t, gump.Height, gump.Image.Bounds().Dy())

			//assert.NoError(t, savePng(img, "gump.png")
		})
	})

	t.Run("GumpInfos", func(t *testing.T) {
		runWith(t, func(sdk *SDK) {
			count := 0
			for info := range sdk.Gumps() {
				assert.NotZero(t, info.Height, fmt.Sprintf("Gump height should be non-zero for ID %d", info.ID))
				assert.NotZero(t, info.Width, fmt.Sprintf("Gump width should be non-zero for ID %d", info.ID))

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
			gump, err := sdk.Gump(0xFFFFF) // Very high ID
			assert.Error(t, err, "Should error on invalid gump ID")
			assert.Nil(t, gump, "Should not return a gump for invalid ID")
		})
	})

	t.Run("Gump40018", func(t *testing.T) {
		runWith(t, func(sdk *SDK) {
			gump, err := sdk.Gump(40018)
			require.NoError(t, err)
			require.NotNil(t, gump)
			assert.NotNil(t, gump.Image)
		})
	})
}
