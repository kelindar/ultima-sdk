package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTexture_Load(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Test loading a known texture (index 3)
		tex, err := sdk.Texture(3)
		assert.NoError(t, err)
		assert.NotNil(t, tex)
		assert.NotNil(t, tex.Image())
		//savePng(tex.Image(), "test_texture.png")

		// Test loading an out-of-bounds texture (should return nil, no error)
		tex, err = sdk.Texture(0x4000)
		assert.NoError(t, err)
		assert.Nil(t, tex)
	})
}
