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
				assert.NotNil(t, c.Bitmap)

				//savePng(c.Bitmap, "a.png")
			}
		})

		t.Run("Unicode", func(t *testing.T) {
			font, err := sdk.FontUnicode()
			assert.NoError(t, err)
			c := font.Character('你') // Example: Chinese char
			assert.NotNil(t, c)
			assert.GreaterOrEqual(t, c.Width, 0)
			assert.GreaterOrEqual(t, c.Height, 0)

			//savePng(c.Bitmap, "x.png")
		})
	})
}
