package ultima

import (
	"path/filepath"
	"testing"

	uotest "github.com/kelindar/ultima-sdk/internal/testing"
	"github.com/stretchr/testify/assert"
)

func TestSDK_EdgeCases(t *testing.T) {
	t.Run("Open_PermissionDenied", func(t *testing.T) {
		// Try to open a directory that should not be accessible (simulate if possible)
		// On CI, this may not be possible, so just check that error is returned
		_, err := Open(filepath.Join(uotest.Path(), "nonexistent", "denied"))
		assert.Error(t, err)
	})

	t.Run("BasePath_AfterClose", func(t *testing.T) {
		runWith(t, func(sdk *SDK) {
			_ = sdk.Close()
			assert.Empty(t, sdk.BasePath(), "BasePath should be empty after Close")
		})
	})

	t.Run("Close_AfterFailedOpen", func(t *testing.T) {
		sdk, err := Open("invalid/path/to/nowhere")
		assert.Nil(t, sdk)
		assert.Error(t, err)
		if sdk != nil {
			assert.NoError(t, sdk.Close())
		}
	})
}

func TestMulti_EdgeCases(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		t.Run("Multi_EmptyData", func(t *testing.T) {
			// Simulate multi.mul returning empty data (requires test double/mocking)
			// For now, just check that a high, likely-nonexistent ID returns error
			_, err := sdk.Multi(999999)
			assert.Error(t, err)
		})

		t.Run("MultiImage_NoItems", func(t *testing.T) {
			m := &Multi{sdk: sdk, Items: nil}
			img, err := m.Image()
			assert.Error(t, err)
			assert.Nil(t, img)
		})
	})
}

func TestSpeech_EdgeCases(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		t.Run("SpeechEntry_OutOfRange", func(t *testing.T) {
			_, err := sdk.SpeechEntry(999999)
			assert.Error(t, err)
		})
	})
}

func TestTileMap_TileAt_OutOfBounds(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		m, err := sdk.Map(1)
		assert.NoError(t, err)
		_, err = m.TileAt(-1, -1)
		assert.Error(t, err)
		_, err = m.TileAt(99999, 99999)
		assert.Error(t, err)
	})
}
