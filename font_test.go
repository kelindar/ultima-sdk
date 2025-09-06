// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFont_Load(t *testing.T) {

	runWith(t, func(sdk *SDK) {
		t.Run("ASCII", func(t *testing.T) {
			fonts, err := sdk.Font()
			assert.NoError(t, err)
			assert.Len(t, fonts, 10)
			for i, font := range fonts {
				c := font.Character('A')
				assert.NotNil(t, c, "Font %d: missing 'A'", i)
				assert.Greater(t, c.Width, 0)
				assert.Greater(t, c.Height, 0)
				assert.NotNil(t, c.Image)
				assert.NoError(t, savePng(c.Image, "test/font_ascii.png"))
			}
		})

		t.Run("Unicode", func(t *testing.T) {
			_, err := sdk.FontUnicode()
			assert.NoError(t, err)

			font, err := sdk.FontUnicode()
			assert.NoError(t, err)
			c := font.Character('你')
			assert.NotNil(t, c)
			assert.GreaterOrEqual(t, c.Width, 0)
			assert.GreaterOrEqual(t, c.Height, 0)
			assert.NoError(t, savePng(c.Image, "test/font_utf8.png"))
		})
	})
}
