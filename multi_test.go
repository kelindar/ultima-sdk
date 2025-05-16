package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMulti_Load(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		multi, err := sdk.Multi(0)
		assert.NoError(t, err)
		assert.NotNil(t, multi)
		assert.Greater(t, len(multi.Items), 0)
		item := multi.Items[0]
		assert.NotZero(t, item.ItemID)

		img, err := multi.Image()
		assert.NoError(t, err)
		assert.NotNil(t, img)

		savePng(img, "multi.png")
	})
}
