// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"fmt"
	"image"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
)

// FontRune represents a single character's metadata and bitmap.
// FontRune holds metadata and bitmap for a single font character.
type FontRune struct {
	Width   int
	Height  int
	XOffset int
	YOffset int
	Image   image.Image
}

// Font is the interface for font types (ASCII, Unicode)
type Font interface {
	Character(rune) *FontRune
	Size(string) (int, int)
}

// FontUnicode loads a Unicode font from unifont*.mul using the SDK file loader.
func (s *SDK) FontUnicode() (Font, error) {
	file, err := s.loadFontUnicode()
	if err != nil {
		return nil, fmt.Errorf("load unifont1.mul: %w", err)
	}

	data, err := file.ReadFull(0)
	if err != nil {
		return nil, err
	}

	font := &unicodeFont{}

	// Read 0x10000 (65536) 4-byte offsets
	offsets := make([]int32, 0x10000)
	for i := 0; i < 0x10000; i++ {
		off := i * 4
		if off+4 > len(data) {
			return nil, fmt.Errorf("offset table out of bounds at %d", i)
		}
		offsets[i] = int32(data[off]) | int32(data[off+1])<<8 | int32(data[off+2])<<16 | int32(data[off+3])<<24
	}

	for i := 0; i < 0x10000; i++ {
		offset := offsets[i]
		if offset <= 0 {
			font.Characters[i] = &FontRune{}
			continue
		}
		if int(offset)+4 > len(data) {
			return nil, fmt.Errorf("char meta out of bounds at %d", i)
		}
		meta := data[offset : offset+4]
		xOff := int(int8(meta[0]))
		yOff := int(int8(meta[1]))
		width := int(meta[2])
		height := int(meta[3])
		var bmp *bitmap.ARGB1555
		if width > 0 && height > 0 {
			bytesPerRow := (width + 7) / 8
			dataLen := height * bytesPerRow
			if int(offset)+4+dataLen > len(data) {
				return nil, fmt.Errorf("char data out of bounds at %d", i)
			}
			charData := data[offset+4 : offset+4+int32(dataLen)]
			bmp = decodeUnicodeBitmap(width, height, charData)
		}

		font.Characters[i] = &FontRune{
			Width:   width,
			Height:  height,
			XOffset: xOff,
			YOffset: yOff,
			Image:   bmp,
		}
	}
	return font, nil
}

// decodeUnicodeBitmap decodes a bit-packed Unicode font glyph to ARGB1555.
func decodeUnicodeBitmap(width, height int, data []byte) *bitmap.ARGB1555 {
	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := (x / 8) + (y * ((width + 7) / 8))
			if offset >= len(data) {
				continue
			}
			bit := (data[offset] & (1 << (7 - (x % 8)))) != 0
			if bit {
				o := img.PixOffset(x, y)
				img.Pix[o] = 0x00
				img.Pix[o+1] = 0x80 // 0x8000 in little-endian (ARGB1555 opaque black)
			}
		}
	}
	return img
}

// Font loads all ASCII fonts from fonts.mul using the SDK file loader.
func (s *SDK) Font() ([]Font, error) {
	file, err := s.loadFont()
	if err != nil {
		return nil, fmt.Errorf("load fonts.mul: %w", err)
	}

	data, err := file.ReadFull(0)
	if err != nil {
		return nil, err
	}

	var fonts [10]*asciiFont
	offset := 0
	for i := 0; i < 10; i++ {
		if offset+1 > len(data) {
			return nil, fmt.Errorf("header out of bounds at font %d", i)
		}
		header := data[offset]
		offset++
		fonts[i] = &asciiFont{Header: header}
		for k := 0; k < 224; k++ {
			if offset+3 > len(data) {
				return nil, fmt.Errorf("char meta out of bounds at font %d char %d", i, k)
			}
			buf := data[offset : offset+3]
			offset += 3
			width, height, unk := int(buf[0]), int(buf[1]), buf[2]
			fonts[i].Unk[k] = unk
			var bmp *bitmap.ARGB1555
			if width > 0 && height > 0 {
				pixLen := width * height * 2
				if offset+pixLen > len(data) {
					return nil, fmt.Errorf("char pixels out of bounds at font %d char %d", i, k)
				}
				pix := data[offset : offset+pixLen]
				offset += pixLen
				bmp = decodeARGB1555(width, height, pix)
				if height > fonts[i].Height && k < 96 {
					fonts[i].Height = height
				}
			}
			fonts[i].Characters[k] = &FontRune{
				Width:  width,
				Height: height,
				Image:  bmp,
			}
		}
	}
	out := make([]Font, 10)
	for i := range fonts {
		out[i] = fonts[i]
	}
	return out, nil
}

// decodeARGB1555 converts ARGB1555 bytes to a bitmap.ARGB1555 image.
func decodeARGB1555(width, height int, data []byte) *bitmap.ARGB1555 {
	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))
	if len(data) == len(img.Pix) {
		copy(img.Pix, data)
	}
	return img
}

// unicodeFont implements Font for Unicode fonts (unifont*.mul)
type unicodeFont struct {
	Characters [0x10000]*FontRune
}

// Character returns the FontRune for a given Unicode character.
func (f *unicodeFont) Character(r rune) *FontRune {
	return f.Characters[int(r)%0x10000]
}

// Size returns the width and height of the text in pixels.
func (f *unicodeFont) Size(text string) (int, int) {
	w, h := 0, 0
	for _, r := range text {
		c := f.Character(r)
		if c == nil {
			w += 8
			continue
		}

		w += c.Width + c.XOffset
		if (c.Height + c.YOffset) > h {
			h = c.Height + c.YOffset
		}
	}
	return w, h
}

// asciiFont implements Font for ASCII fonts (fonts.mul)
type asciiFont struct {
	Header     byte
	Unk        [224]byte
	Characters [224]*FontRune
	Height     int
}

func (f *asciiFont) Character(r rune) *FontRune {
	idx := int((r - 0x20) & 0x7FFFFFFF % 224)
	return f.Characters[idx]
}

func (f *asciiFont) Size(text string) (int, int) {
	w, h := 0, f.Height
	for _, r := range text {
		if c := f.Character(r); c != nil {
			w += c.Width
		}
	}
	return w, h
}
