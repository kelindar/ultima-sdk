// Package ultima provides access to Ultima Online texture data.
package ultima

import (
	"image"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
)

// Texture represents a texture entry loaded from texmaps.mul.
type Texture struct {
	Index int    // Texture index
	Size  int    // Texture size (64 or 128)
	image []byte // Texture data
}

// Image returns the image of the texture.
func (t *Texture) Image() image.Image {
	img := bitmap.NewARGB1555(image.Rect(0, 0, t.Size, t.Size))
	img.Pix = t.image
	return img
}

// Texture returns a texture by index.
func (s *SDK) Texture(index int) (*Texture, error) {
	idx := index & 0x3FFF
	file, err := s.loadTextures()
	if err != nil {
		return nil, err
	}

	data, extra, err := file.Read(uint32(idx))
	if err != nil || len(data) == 0 {
		return nil, nil
	}

	size := 64
	if extra == 1 {
		size = 128
	}

	return &Texture{
		Index: idx,
		Size:  size,
		image: data,
	}, nil
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
