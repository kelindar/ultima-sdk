package anim

import (
	"encoding/binary"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDecodeFrame_MinData verifies DecodeFrame returns no image for insufficient data
func TestDecodeFrame_MinData(t *testing.T) {
	center, img, err := DecodeFrame(nil, []byte{1, 2, 3}, false)
	assert.NoError(t, err)
	assert.Nil(t, img)
	assert.Equal(t, image.Point{}, center)
}

// TestDecodeFrame_Terminator verifies DecodeFrame correctly handles a frame with only a terminator
func TestDecodeFrame_Terminator(t *testing.T) {
	xCenter, yCenter, width, height := 5, 7, 10, 12
	data := make([]byte, 8+4)
	binary.LittleEndian.PutUint16(data[0:2], uint16(xCenter))
	binary.LittleEndian.PutUint16(data[2:4], uint16(yCenter))
	binary.LittleEndian.PutUint16(data[4:6], uint16(width))
	binary.LittleEndian.PutUint16(data[6:8], uint16(height))
	// Terminator header
	binary.LittleEndian.PutUint32(data[8:12], 0x7FFF7FFF)
	palette := make([]uint16, 256)

	center, img, err := DecodeFrame(palette, data, false)
	assert.NoError(t, err)
	assert.NotNil(t, img)
	bounds := img.Bounds()
	assert.Equal(t, width, bounds.Dx())
	assert.Equal(t, height, bounds.Dy())
	assert.Equal(t, image.Point{xCenter, yCenter}, center)
}

// TestDecodeFrame_Flip verifies DecodeFrame adjusts center for flipped frames correctly
func TestDecodeFrame_Flip(t *testing.T) {
	xCenter, yCenter, width, height := 2, 3, 6, 8
	data := make([]byte, 8+4)
	binary.LittleEndian.PutUint16(data[0:2], uint16(xCenter))
	binary.LittleEndian.PutUint16(data[2:4], uint16(yCenter))
	binary.LittleEndian.PutUint16(data[4:6], uint16(width))
	binary.LittleEndian.PutUint16(data[6:8], uint16(height))
	binary.LittleEndian.PutUint32(data[8:12], 0x7FFF7FFF)
	palette := make([]uint16, 256)

	center, img, err := DecodeFrame(palette, data, true)
	assert.NoError(t, err)
	assert.NotNil(t, img)
	// For flipped images, X center should be width - xCenter
	assert.Equal(t, width-xCenter, center.X)
	assert.Equal(t, yCenter, center.Y)
}
