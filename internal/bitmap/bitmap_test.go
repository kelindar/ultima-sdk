package bitmap

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestARGB1555Color_RGBA(t *testing.T) {
	// Test cases for ARGB1555 to RGBA conversion
	tests := []struct {
		name     string
		color    ARGB1555Color
		expected color.RGBA64 // Use RGBA64 for precision
	}{
		{
			name:     "Opaque Black (0x8000)",
			color:    0x8000,
			expected: color.RGBA64{R: 0, G: 0, B: 0, A: 0xFFFF},
		},
		{
			name:     "Transparent Black (0x0000)",
			color:    0x0000,
			expected: color.RGBA64{R: 0, G: 0, B: 0, A: 0x0000},
		},
		{
			name:     "Opaque White (0xFFFF)",
			color:    0xFFFF, // A=1, R=31, G=31, B=31
			expected: color.RGBA64{R: 0xFFFF, G: 0xFFFF, B: 0xFFFF, A: 0xFFFF},
		},
		{
			name:     "Opaque Red (0xFC00)",
			color:    0xFC00, // A=1, R=31, G=0, B=0
			expected: color.RGBA64{R: 0xFFFF, G: 0, B: 0, A: 0xFFFF},
		},
		{
			name:     "Opaque Green (0x83E0)",
			color:    0x83E0, // A=1, R=0, G=31, B=0
			expected: color.RGBA64{R: 0, G: 0xFFFF, B: 0, A: 0xFFFF},
		},
		{
			name:     "Opaque Blue (0x801F)",
			color:    0x801F, // A=1, R=0, G=0, B=31
			expected: color.RGBA64{R: 0, G: 0, B: 0xFFFF, A: 0xFFFF},
		},
		{
			name:     "Transparent White (0x7FFF)",
			color:    0x7FFF, // A=0, R=31, G=31, B=31
			expected: color.RGBA64{R: 0xFFFF, G: 0xFFFF, B: 0xFFFF, A: 0x0000},
		},
		{
			name:  "Opaque Gray (0xC638)", // A=1, R=17, G=17, B=24
			color: 0xC638,
			// R = (17 * 65535) / 31 = 35938 (0x8C62)
			// G = (17 * 65535) / 31 = 35938 (0x8C62)
			// B = (24 * 65535) / 31 = 50736 (0xC630)
			expected: color.RGBA64{R: 0x8C62, G: 0x8C62, B: 0xC630, A: 0xFFFF}, // Corrected expected values based on actual calculation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b, a := tt.color.RGBA()
			actual := color.RGBA64{R: uint16(r), G: uint16(g), B: uint16(b), A: uint16(a)}
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestARGB1555Model(t *testing.T) {
	// Test cases for converting to and from ARGB1555Color using the model
	tests := []struct {
		name         string
		original     color.Color
		expectedARGB ARGB1555Color
	}{
		{
			name:         "Convert Opaque Black RGBA to ARGB1555",
			original:     color.RGBA{R: 0, G: 0, B: 0, A: 255},
			expectedARGB: 0x8000, // A=1, R=0, G=0, B=0
		},
		{
			name:         "Convert Transparent Black RGBA to ARGB1555",
			original:     color.RGBA{R: 0, G: 0, B: 0, A: 0},
			expectedARGB: 0x0000, // A=0, R=0, G=0, B=0
		},
		{
			name:         "Convert Opaque White RGBA to ARGB1555",
			original:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
			expectedARGB: 0xFFFF, // A=1, R=31, G=31, B=31
		},
		{
			name:         "Convert Transparent White RGBA to ARGB1555",
			original:     color.RGBA{R: 255, G: 255, B: 255, A: 0},
			expectedARGB: 0x7FFF, // A=0, R=31, G=31, B=31
		},
		{
			name:         "Convert Opaque Red RGBA to ARGB1555",
			original:     color.RGBA{R: 255, G: 0, B: 0, A: 255},
			expectedARGB: 0xFC00, // A=1, R=31, G=0, B=0
		},
		{
			name:         "Convert Opaque Green RGBA to ARGB1555",
			original:     color.RGBA{R: 0, G: 255, B: 0, A: 255},
			expectedARGB: 0x83E0, // A=1, R=0, G=31, B=0
		},
		{
			name:         "Convert Opaque Blue RGBA to ARGB1555",
			original:     color.RGBA{R: 0, G: 0, B: 255, A: 255},
			expectedARGB: 0x801F, // A=1, R=0, G=0, B=31
		},
		{
			name:         "Convert Semi-Transparent Gray RGBA64 to ARGB1555",
			original:     color.RGBA64{R: 0x8000, G: 0x8000, B: 0x8000, A: 0x7FFF}, // Alpha < 0x8000 -> transparent
			expectedARGB: 0x4210,                                                   // A=0, R=16, G=16, B=16
		},
		{
			name:         "Convert Opaque Gray RGBA64 to ARGB1555",
			original:     color.RGBA64{R: 0x8000, G: 0x8000, B: 0x8000, A: 0xFFFF}, // Alpha >= 0x8000 -> opaque
			expectedARGB: 0xC210,                                                   // A=1, R=16, G=16, B=16
		},
		{
			name:         "Convert ARGB1555 itself (should return same)",
			original:     ARGB1555Color(0xABCD),
			expectedARGB: ARGB1555Color(0xABCD),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converted := ARGB1555Model.Convert(tt.original)
			assert.IsType(t, ARGB1555Color(0), converted)
			assert.Equal(t, tt.expectedARGB, converted.(ARGB1555Color))

			// Optional: Convert back and check (might lose precision)
			// rOrig, gOrig, bOrig, aOrig := tt.original.RGBA()
			// rConv, gConv, bConv, aConv := converted.RGBA()
			// Compare rConv, gConv, bConv, aConv with rOrig, gOrig, bOrig, aOrig
			// Note: Due to 5-bit to 8/16-bit scaling, exact match back isn't guaranteed.
		})
	}
}

func TestARGB1555Image(t *testing.T) {
	rect := image.Rect(0, 0, 10, 5)
	img := NewARGB1555(rect)

	assert.Equal(t, rect, img.Bounds(), "Bounds should match rectangle used for creation")
	assert.Equal(t, ARGB1555Model, img.ColorModel(), "ColorModel should be ARGB1555Model")
	assert.Equal(t, 10*2, img.Stride, "Stride should be width * 2 bytes")
	assert.Len(t, img.Pix, 10*5*2, "Pixel buffer size should be width * height * 2")

	// Test Set and At
	col1 := ARGB1555Color(0x8FFF) // Opaque white
	col2 := ARGB1555Color(0x0ABC) // Transparent color

	img.Set(0, 0, col1)
	img.Set(9, 4, col2)
	img.Set(5, 2, color.RGBA{R: 255, G: 0, B: 0, A: 255}) // Opaque Red

	assert.Equal(t, col1, img.At(0, 0), "At(0,0) should return col1")
	assert.Equal(t, col2, img.At(9, 4), "At(9,4) should return col2")
	assert.Equal(t, ARGB1555Color(0xFC00), img.At(5, 2), "At(5,2) should return converted opaque red")

	// Test At out of bounds
	assert.Equal(t, ARGB1555Color(0), img.At(-1, 0), "At(-1,0) should be transparent black")
	assert.Equal(t, ARGB1555Color(0), img.At(10, 0), "At(10,0) should be transparent black")
	assert.Equal(t, ARGB1555Color(0), img.At(0, 5), "At(0,5) should be transparent black")

	// Test Set out of bounds (should do nothing)
	img.Set(10, 0, col1)
	assert.Equal(t, ARGB1555Color(0), img.At(10, 0), "Set(10,0) should not modify out of bounds")

	// Test PixOffset
	assert.Equal(t, 0, img.PixOffset(0, 0), "PixOffset(0,0) should be 0")
	assert.Equal(t, 2, img.PixOffset(1, 0), "PixOffset(1,0) should be 2")
	assert.Equal(t, 10*2, img.PixOffset(0, 1), "PixOffset(0,1) should be stride")
	assert.Equal(t, 4*img.Stride+(9*2), img.PixOffset(9, 4), "PixOffset(9,4) should be correct")

	// Test Opaque
	assert.False(t, img.Opaque(), "Image should not be opaque because col2 is transparent")

	// Make image opaque
	opaqueColor := ARGB1555Color(0x8000) // Opaque black
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, opaqueColor)
		}
	}
	assert.True(t, img.Opaque(), "Image should now be opaque")

	// Test SubImage
	subRect := image.Rect(2, 1, 8, 4) // 6x3 sub image
	subImg := img.SubImage(subRect).(*ARGB1555)

	assert.Equal(t, subRect, subImg.Bounds(), "SubImage bounds should match intersection")
	assert.Equal(t, img.Stride, subImg.Stride, "SubImage stride should be same as original")
	assert.Equal(t, img.ColorModel(), subImg.ColorModel(), "SubImage color model should be same")

	// Check that subimage shares pixels
	assert.Equal(t, opaqueColor, subImg.At(2, 1), "SubImage pixel should match original")
	newCol := ARGB1555Color(0xFFFF)
	subImg.Set(3, 2, newCol) // Set within subimage bounds
	assert.Equal(t, newCol, img.At(3, 2), "Setting pixel in SubImage should affect original image")

	// Test SubImage with non-overlapping rect
	emptyRect := image.Rect(100, 100, 110, 110)
	emptySub := img.SubImage(emptyRect).(*ARGB1555)
	assert.True(t, emptySub.Bounds().Empty(), "SubImage of non-overlapping rect should be empty")
	assert.Nil(t, emptySub.Pix, "Empty SubImage Pix should be nil")
}
