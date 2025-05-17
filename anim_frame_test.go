// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"encoding/binary"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDecodeFrame_MinData verifies DecodeFrame returns no image for insufficient data
func TestDecodeFrame_MinData(t *testing.T) {
	center, img, err := decodeFrame(nil, []byte{1, 2, 3}, false)
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

	center, img, err := decodeFrame(palette, data, false)
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

	center, img, err := decodeFrame(palette, data, true)
	assert.NoError(t, err)
	assert.NotNil(t, img)
	// For flipped images, X center should be width - xCenter
	assert.Equal(t, width-xCenter, center.X)
	assert.Equal(t, yCenter, center.Y)
}

func TestDecodeFrame_EdgeCases(t *testing.T) {
	// Nil palette, valid header
	data := make([]byte, 8+4)
	binary.LittleEndian.PutUint16(data[0:2], 1)
	binary.LittleEndian.PutUint16(data[2:4], 1)
	binary.LittleEndian.PutUint16(data[4:6], 1)
	binary.LittleEndian.PutUint16(data[6:8], 1)
	binary.LittleEndian.PutUint32(data[8:12], 0x7FFF7FFF)
	center, img, err := decodeFrame(nil, data, false)
	assert.NoError(t, err)
	assert.NotNil(t, img)
	assert.Equal(t, image.Point{1, 1}, center)

	// Empty data
	center, img, err = decodeFrame(nil, []byte{}, false)
	assert.NoError(t, err)
	assert.Nil(t, img)
	assert.Equal(t, image.Point{}, center)

	// Short header
	center, img, err = decodeFrame(nil, []byte{1, 2}, false)
	assert.NoError(t, err)
	assert.Nil(t, img)
	assert.Equal(t, image.Point{}, center)

	// Palette wrong size (simulate by passing a slice of wrong length)
	palette := make([]uint16, 10)
	center, img, err = decodeFrame(palette, data, false)
	assert.NoError(t, err)
	assert.NotNil(t, img)
}
