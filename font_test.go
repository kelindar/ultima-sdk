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
				c := font.Rune('A')
				assert.NotNil(t, c, "Font %d: missing 'A'", i)
				assert.Greater(t, c.Width, int8(0))
				assert.Greater(t, c.Height, int8(0))
				assert.NotNil(t, c.Image)
				assert.NoError(t, savePng(c.Image, "test/font_ascii.png"))
			}
		})

		t.Run("Unicode", func(t *testing.T) {
			font, err := sdk.FontUnicode(1)
			assert.NoError(t, err)
			c := font.Rune('你')
			assert.NotNil(t, c)
			assert.GreaterOrEqual(t, c.Width, int8(0))
			assert.GreaterOrEqual(t, c.Height, int8(0))
			assert.NoError(t, savePng(c.Image, "test/font_utf8.png"))
		})

		t.Run("Space", func(t *testing.T) {
			font, err := sdk.FontUnicode(1)
			assert.NoError(t, err)
			w, h := font.Size(" ")
			assert.Equal(t, 0, h)
			assert.Equal(t, 8, w)
		})

		t.Run("Size", func(t *testing.T) {
			font, err := sdk.FontUnicode(1)
			assert.NoError(t, err)
			w, h := font.Size("Hello, World!")
			assert.Equal(t, 15, h)
			assert.Equal(t, 73, w)
		})

		t.Run("Text", func(t *testing.T) {
			font, err := sdk.FontUnicode(1)
			assert.NoError(t, err)
			img := sdk.Text(font, "Hello, World!", 1)
			assert.NotNil(t, img)
			assert.NoError(t, savePng(img, "test/font_text.png"))
		})
	})
}
