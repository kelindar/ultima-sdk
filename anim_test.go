package ultima

import (
	"fmt"
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
}

func testLoadAnimation(t *testing.T, body, action, direction int) {
	runWith(t, func(sdk *SDK) {
		anim, err := sdk.LoadAnimation(body, action, direction, 0, false, false)
		assert.NoError(t, err, "LoadAnimation should succeed")
		assert.NotNil(t, anim, "Animation should not be nil")
		assert.NotEmpty(t, anim.Frames, "Expected non-empty frames")

		for i, frame := range anim.Frames {
			assert.NotNil(t, frame.Bitmap, "Frame %d bitmap should not be nil", i)
			assert.NotZero(t, frame.Bitmap.Bounds().Dx(), "Frame %d width should not be zero", i)
			assert.NotZero(t, frame.Bitmap.Bounds().Dy(), "Frame %d height should not be zero", i)
			savePng(frame.Bitmap, fmt.Sprintf("frame_%d.png", i))
			break
		}
	})
}
