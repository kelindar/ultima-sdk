// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"fmt"
	"image"
	"image/color"

	"iter"
)

// Light represents a light source image.
type Light struct {
	ID     int    // ID of the light
	Width  int    // Width of the light image
	Height int    // Height of the light image
	image  []byte // Raw image data
}

// Image returns the light image as a grayscale image.
func (l *Light) Image() image.Image {

	// The light.mul data contains signed byte values that represent light intensity
	// In C#, each byte is converted to a 16-bit ARGB1555 value where R=G=B = 0x1F + value
	img := image.NewGray16(image.Rect(0, 0, l.Width, l.Height))

	for y := 0; y < l.Height; y++ {
		for x := 0; x < l.Width; x++ {
			offset := y*l.Width + x
			if offset >= len(l.image) {
				break
			}

			// Convert signed byte to intensity
			value := int8(l.image[offset])
			intensity := uint16(0x1F + value)

			// Scale intensity (0-31) to 16-bit grayscale range (0-65535)
			scaledIntensity := uint16((float64(intensity) / 31.0) * 65535.0)
			img.SetGray16(x, y, color.Gray16{Y: scaledIntensity})
		}
	}
	return img
}

// Light retrieves a specific light image by ID.
func (s *SDK) Light(id int) (Light, error) {
	if id < 0 {
		return Light{}, fmt.Errorf("invalid light ID: %d", id)
	}

	file, err := s.loadLights()
	if err != nil {
		return Light{}, err
	}

	entry, err := file.Entry(uint32(id))
	if err != nil {
		return Light{}, err
	}

	data := make([]byte, entry.Len())
	if _, err := entry.ReadAt(data, 0); err != nil {
		return Light{}, err
	}

	return makeLight(uint32(id), data, uint32(entry.Extra()))
}

// Lights returns an iterator over all defined light images.
func (s *SDK) Lights() iter.Seq[Light] {
	file, err := s.loadLights()
	if err != nil {
		return func(yield func(Light) bool) {}
	}

	return func(yield func(Light) bool) {
		for index := range file.Entries() {
			entry, err := file.Entry(index)
			if err != nil {
				continue
			}

			data := make([]byte, entry.Len())
			if _, err := entry.ReadAt(data, 0); err != nil {
				continue
			}

			light, err := makeLight(index, data, uint32(entry.Extra()))
			if err != nil {
				continue
			}

			if !yield(light) {
				return
			}
		}
	}
}

func lightSize(extra uint32) (int, int) {
	width := int(extra & 0xFFFF)
	height := int((extra >> 16) & 0xFFFF)
	return width, height
}

// makeLight processes the raw byte data from the MUL file into a Light struct.
// The 'extra' value from the index file contains width (lower 16 bits) and height (upper 16 bits).
func makeLight(id uint32, data []byte, extra uint32) (Light, error) {
	width, height := lightSize(extra)
	if width <= 0 || height <= 0 {
		return Light{}, fmt.Errorf("invalid dimensions for light ID %d: width=%d, height=%d", id, width, height)
	}

	if len(data) < width*height {
		return Light{}, fmt.Errorf("data length mismatch for light ID %d: expected at least %d, got %d", id, width*height, len(data))
	}

	return Light{
		ID:     int(id),
		Width:  width,
		Height: height,
		image:  data,
	}, nil
}
