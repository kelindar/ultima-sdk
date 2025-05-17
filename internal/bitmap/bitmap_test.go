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

func TestARGB1555Image_Additional(t *testing.T) {
	rect := image.Rect(0, 0, 3, 2)
	img := NewARGB1555(rect)

	// Test: All pixels initially zero (transparent)
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			assert.Equal(t, ARGB1555Color(0), img.At(x, y), "Initial pixel should be transparent")
		}
	}

	// Test: Set and get with roundtrip for all ARGB1555 values in a small range
	for val := uint16(0); val < 0x20; val++ {
		col := ARGB1555Color(val | 0x8000) // Opaque, varying blue
		img.Set(0, 0, col)
		got := img.At(0, 0)
		assert.Equal(t, col, got, "Roundtrip ARGB1555 should match")
	}

	// Test: Stride correctness for non-square image
	assert.Equal(t, 6, img.Stride)
	img.Set(2, 1, ARGB1555Color(0x801F)) // Opaque blue
	assert.Equal(t, ARGB1555Color(0x801F), img.At(2, 1))

	// Test: Opaque on empty image
	empty := &ARGB1555{}
	assert.True(t, empty.Opaque(), "Empty image should be considered opaque")

	// Test: SubImage chaining
	sub1 := img.SubImage(image.Rect(1, 0, 3, 2)).(*ARGB1555)
	sub2 := sub1.SubImage(image.Rect(2, 0, 3, 2)).(*ARGB1555)
	assert.Equal(t, image.Rect(2, 0, 3, 2), sub2.Bounds())

	// Test: At/Set with negative and out-of-bounds coordinates
	img.Set(-10, -10, ARGB1555Color(0x8000))
	img.Set(100, 100, ARGB1555Color(0x8000))
	assert.Equal(t, ARGB1555Color(0), img.At(-1, -1))
	assert.Equal(t, ARGB1555Color(0), img.At(100, 100))
}

