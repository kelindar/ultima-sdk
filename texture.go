// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"image"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
	"github.com/kelindar/ultima-sdk/internal/uofile"
)

// Texture represents a texture entry loaded from texmaps.mul.
type Texture struct {
	Index int         // Texture index
	Size  int         // Texture size (64 or 128)
	Image image.Image // Texture data
}

// Texture returns a texture by index.
func (s *SDK) Texture(index int) (*Texture, error) {
	idx := index & 0x3FFF
	file, err := s.loadTextures()
	if err != nil {
		return nil, err
	}

	return uofile.Decode(file, uint32(idx), func(data []byte, extra uint64) (*Texture, error) {
		size := 64
		if extra == 1 {
			size = 128
		}
		img := bitmap.NewARGB1555(image.Rect(0, 0, size, size))
		img.Pix = make([]byte, len(data))
		copy(img.Pix, data)
		return &Texture{
			Index: idx,
			Size:  size,
			Image: img,
		}, nil
	})

}

// Textures returns an iterator over all available textures.
func (s *SDK) Textures() func(yield func(*Texture) bool) {
	return func(yield func(*Texture) bool) {
		for i := 0; i < 0x4000; i++ {
			tex, err := s.Texture(i)
			if err != nil || tex == nil {
				continue
			}
			if !yield(tex) {
				break
			}
		}
	}
}
