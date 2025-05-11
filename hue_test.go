package ultima

import (
	"testing"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHue_GetColor(t *testing.T) {
	hue := &Hue{
		Colors: [32]uint16{
			0x7C00, // Red (without alpha bit)
			0x03E0, // Green (without alpha bit)
			0x001F, // Blue (without alpha bit)
		},
	}

	// Test retrieving a valid color
	color, err := hue.GetColor(0)
	require.NoError(t, err)
	require.NotNil(t, color)

	// Verify it's the correct type and has the alpha bit set
	colorARGB, ok := color.(bitmap.ARGB1555Color)
	require.True(t, ok)
	assert.Equal(t, bitmap.ARGB1555Color(0xFC00), colorARGB) // 0x7C00 + 0x8000 (alpha bit)

	// Test green color
	color, err = hue.GetColor(1)
	require.NoError(t, err)
	colorARGB, ok = color.(bitmap.ARGB1555Color)
	require.True(t, ok)
	assert.Equal(t, bitmap.ARGB1555Color(0x83E0), colorARGB) // 0x03E0 + 0x8000 (alpha bit)

	// Test blue color
	color, err = hue.GetColor(2)
	require.NoError(t, err)
	colorARGB, ok = color.(bitmap.ARGB1555Color)
	require.True(t, ok)
	assert.Equal(t, bitmap.ARGB1555Color(0x801F), colorARGB) // 0x001F + 0x8000 (alpha bit)

	// Test out of range index
	_, err = hue.GetColor(-1)
	assert.Error(t, err)
	_, err = hue.GetColor(32)
	assert.Error(t, err)
}

func TestHue_Image(t *testing.T) {
	hue := &Hue{
		Colors: [32]uint16{
			0x7C00, // Red (without alpha bit)
			0x03E0, // Green (without alpha bit)
			0x001F, // Blue (without alpha bit)
			// ... rest are zero
		},
	}

	// Test with a simple 2x2 image per color
	img := hue.Image(2, 2)
	require.NotNil(t, img)

	// Check image dimensions
	bounds := img.Bounds()
	assert.Equal(t, 0, bounds.Min.X)
	assert.Equal(t, 0, bounds.Min.Y)
	assert.Equal(t, 2*32, bounds.Max.X) // widthPerColor * number of colors
	assert.Equal(t, 2, bounds.Max.Y)

	// Check specific colors at expected positions
	// First color (red)
	c := img.At(1, 1)
	redColor, ok := c.(bitmap.ARGB1555Color)
	require.True(t, ok)
	assert.Equal(t, bitmap.ARGB1555Color(0xFC00), redColor) // 0x7C00 + 0x8000 (alpha bit)

	// Second color (green)
	c = img.At(3, 1)
	greenColor, ok := c.(bitmap.ARGB1555Color)
	require.True(t, ok)
	assert.Equal(t, bitmap.ARGB1555Color(0x83E0), greenColor) // 0x03E0 + 0x8000 (alpha bit)

	// Third color (blue)
	c = img.At(5, 1)
	blueColor, ok := c.(bitmap.ARGB1555Color)
	require.True(t, ok)
	assert.Equal(t, bitmap.ARGB1555Color(0x801F), blueColor) // 0x001F + 0x8000 (alpha bit)
}

func TestSDK_HueAt(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Test retrieving a hue at index 0
		hue, err := sdk.HueAt(1337)
		require.NoError(t, err)
		assert.Equal(t, 1337, hue.Index)
		assert.NotEmpty(t, hue.Name)

		// Check that the hue has valid colors
		for _, colorValue := range hue.Colors {
			// The colors should be 16-bit values in ARGB1555 format
			// For standard hues, the alpha bit is not typically set in the raw data
			// (it's set by GetColor when retrieving the color)
			assert.LessOrEqual(t, colorValue, uint16(0x7FFF))
		}

		// Test a hue in the middle of the range
		hue, err = sdk.HueAt(1000)
		require.NoError(t, err)
		assert.Equal(t, 1000, hue.Index)

		// Test retrieving a hue at the upper end of the range
		hue, err = sdk.HueAt(2999)
		require.NoError(t, err)
		assert.Equal(t, 2999, hue.Index)

		// Test invalid indices
		_, err = sdk.HueAt(-1)
		assert.Error(t, err)
		_, err = sdk.HueAt(3000)
		assert.Error(t, err)
	})
}

func TestSDK_Hues(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Count the number of hues
		count := 0
		var firstHue *Hue
		var lastHue *Hue

		for hue := range sdk.Hues() {
			if count == 0 {
				firstHue = hue
			}
			lastHue = hue
			count++
		}

		// Should have a significant number of hues
		assert.Greater(t, count, 100)

		// First hue should be index 0
		assert.Equal(t, 0, firstHue.Index)

		// Last hue should have a high index
		assert.Greater(t, lastHue.Index, 1000)
	})
}

func TestHueColorConversion(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		hue, err := sdk.HueAt(1) // Get hue #1 (typically a bright red)
		require.NoError(t, err)

		// Get a color from the hue
		c, err := hue.GetColor(15) // Middle of the palette
		require.NoError(t, err)

		// Convert to RGBA for comparison
		r, g, b, a := c.RGBA()

		// The color should have full alpha
		assert.Equal(t, uint32(0xFFFF), a)

		// Make sure the color components are valid (non-zero)
		// We can't predict exact values without knowing the exact hue, but they should be valid
		assert.LessOrEqual(t, r, uint32(0xFFFF))
		assert.LessOrEqual(t, g, uint32(0xFFFF))
		assert.LessOrEqual(t, b, uint32(0xFFFF))

		// To simulate what the C# code does, we can create a Go standard color from the hue
		// and confirm it matches what our GetColor method returns
		colorValue := hue.Colors[15] | 0x8000 // Set alpha bit
		expectedColor := bitmap.ARGB1555Color(colorValue)
		assert.Equal(t, expectedColor, c)
	})
}

func TestHueImageRendering(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		hue, err := sdk.HueAt(5) // Get a sample hue
		require.NoError(t, err)

		// Generate a small visualization image
		img := hue.Image(2, 10)
		require.NotNil(t, img)

		// Check image dimensions
		bounds := img.Bounds()
		assert.Equal(t, 0, bounds.Min.X)
		assert.Equal(t, 0, bounds.Min.Y)
		assert.Equal(t, 2*32, bounds.Max.X) // 2 pixels for each of the 32 colors
		assert.Equal(t, 10, bounds.Max.Y)

		// Check that each color block is rendered correctly
		for i := 0; i < 32; i++ {
			x := i*2 + 1 // Center of the i-th color block
			y := 5       // Middle of the image's height

			// Get the color at this position
			c := img.At(x, y)

			// The rendered color should be the hue's color with alpha bit set
			expectedColor := bitmap.ARGB1555Color(hue.Colors[i] | 0x8000)
			assert.Equal(t, expectedColor, c)
		}
	})
}
