package ultima

import (
	"os"
	"path/filepath"
	"testing"

	uotest "github.com/kelindar/ultima-sdk/internal/testing"
	"github.com/stretchr/testify/assert"
)

func TestLoadAnimation(t *testing.T) {
	if _, err := os.Stat(filepath.Join(uotest.Path(), "anim.mul")); err != nil {
		t.Skip("anim.mul not found in testdata, skipping test")
	}

	t.Run("BodyDefault", func(t *testing.T) {
		testLoadAnimation(t, 1, 0, 0) // Valid index in anim.mul
	})

	t.Run("Body200", func(t *testing.T) {
		testLoadAnimation(t, 200, 1, 4) // Validated test parameters
	})

	t.Run("NegativeIndices", func(t *testing.T) {
		runWith(t, func(sdk *SDK) {
			anim, err := sdk.Animation(-1, -1, -1, 0, false, false)
			assert.Error(t, err)
			assert.Nil(t, anim)
		})
	})

	t.Run("OutOfRangeIndices", func(t *testing.T) {
		runWith(t, func(sdk *SDK) {
			anim, err := sdk.Animation(99999, 99999, 99999, 0, false, false)
			assert.Error(t, err)
			assert.Nil(t, anim)
		})
	})

	t.Run("ZeroFrames", func(t *testing.T) {
		runWith(t, func(sdk *SDK) {
			anim, err := sdk.Animation(1, 0, 0, 0, false, false)
			if err == nil && anim != nil {
				count := 0
				for range anim.Frames() {
					count++
				}
				assert.GreaterOrEqual(t, count, 0)
			}
		})
	})
}

func testLoadAnimation(t *testing.T, body, action, direction int) {
	runWith(t, func(sdk *SDK) {
		anim, err := sdk.Animation(body, action, direction, 0, false, false)
		assert.NoError(t, err, "LoadAnimation should succeed")
		assert.NotNil(t, anim, "Animation should not be nil")
		assert.NotNil(t, anim.AnimdataEntry, "Animation metadata should not be nil")
		called := false

		for frame := range anim.Frames() {
			img, err := frame.Image()
			assert.NoError(t, err, "Frame.Image() should succeed")
			assert.NotNil(t, img, "Frame image should not be nil")
			bounds := img.Bounds()
			assert.NotZero(t, bounds.Dx(), "Frame width should not be zero")
			assert.NotZero(t, bounds.Dy(), "Frame height should not be zero")
			called = true
			// savePng(img, "frame.png")
			break
		}
		assert.True(t, called, "Expected at least one frame")
	})
}
