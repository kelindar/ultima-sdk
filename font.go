// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
)

const (
	unicodeFontSize   = 0x10000 // 65536
	unicodeSpaceWidth = 8
	asciiFontsCount   = 10
	asciiGlyphCount   = 224
	asciiFirstRune    = 0x20
	asciiMetricsCount = 96
)

// Rune represents a single character's metadata and bitmap.
type Rune struct {
	Image   image.Image
	Width   int8
	Height  int8
	XOffset int8
	YOffset int8
}

// Font is the interface for font types (ASCII, Unicode)
type Font interface {
	Rune(rune) *Rune
	Size(string) (int, int)
}

// FontUnicode loads a Unicode font from unifont*.mul using the SDK file loader.
func (s *SDK) FontUnicode(n int) (Font, error) {
	file, err := s.loadFontUnicode(n)
	if err != nil {
		return nil, fmt.Errorf("load unifont1.mul: %w", err)
	}

	data, err := file.ReadFull(0)
	if err != nil {
		return nil, err
	}

	font := &unicodeFont{}

	// Read 4-byte little-endian offsets
	offsets := make([]int, unicodeFontSize)
	for i := 0; i < unicodeFontSize; i++ {
		off := i * 4
		if off+4 > len(data) {
			return nil, fmt.Errorf("offset table out of bounds at %d", i)
		}
		offsets[i] = int(binary.LittleEndian.Uint32(data[off : off+4]))
	}

	for i := 0; i < unicodeFontSize; i++ {

		offset := offsets[i]
		if offset <= 0 {
			continue
		}

		if offset+4 > len(data) {
			return nil, fmt.Errorf("char meta out of bounds at %d", i)
		}

		meta := data[offset : offset+4]
		xOff := int(meta[0])
		yOff := int(meta[1])
		width := int(meta[2])
		height := int(meta[3])

		var bmp *bitmap.ARGB1555
		if width > 0 && height > 0 {
			bytesPerRow := (width + 7) / 8
			dataLen := height * bytesPerRow
			if offset+4+dataLen > len(data) {
				return nil, fmt.Errorf("char data out of bounds at %d", i)
			}
			charData := data[offset+4 : offset+4+dataLen]
			bmp = decodeUnicodeBitmap(width, height, charData)
		}

		font.Characters[i] = Rune{
			Width:   int8(width),
			Height:  int8(height),
			XOffset: int8(xOff),
			YOffset: int8(yOff),
			Image:   bmp,
		}
	}
	return font, nil
}

// decodeUnicodeBitmap decodes a bit-packed Unicode font glyph to ARGB1555.
func decodeUnicodeBitmap(width, height int, data []byte) *bitmap.ARGB1555 {
	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))
	bytesPerRow := (width + 7) / 8
	for y := 0; y < height; y++ {
		rowBase := y * bytesPerRow
		for x := 0; x < width; x++ {
			offset := rowBase + (x / 8)
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

	var fonts [asciiFontsCount]*asciiFont

	offset := 0
	for i := 0; i < asciiFontsCount; i++ {
		if offset+1 > len(data) {
			return nil, fmt.Errorf("header out of bounds at font %d", i)
		}
		header := data[offset]
		offset++
		fonts[i] = &asciiFont{Header: header}
		for k := 0; k < asciiGlyphCount; k++ {
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
				if height > fonts[i].Height && k < asciiMetricsCount {
					fonts[i].Height = height
				}
			}

			fonts[i].Characters[k] = Rune{
				Width:  int8(width),
				Height: int8(height),
				Image:  bmp,
			}
		}
	}
	out := make([]Font, asciiFontsCount)
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
	Characters [unicodeFontSize]Rune
}

// Rune returns the FontRune for a given Unicode character.
func (f *unicodeFont) Rune(r rune) *Rune {
	return &f.Characters[int(r)%unicodeFontSize]
}

// Size returns the width and height of the text in pixels.
func (f *unicodeFont) Size(text string) (int, int) {
	if text == "" {
		return 0, 0
	}

	var w, h int
	for _, r := range text {
		switch r {
		case ' ':
			w += unicodeSpaceWidth
			continue
		default:
			c := f.Rune(r)
			if c == nil || c.Width == 0 {
				continue
			}

			w += int(c.Width) + int(c.XOffset)
			if t := int(c.Height) + int(c.YOffset); t > h {
				h = t
			}
		}
	}

	// Add 1 pixel for spacing between characters
	w += len(text) - 1
	return w, h
}

// asciiFont implements Font for ASCII fonts (fonts.mul)
type asciiFont struct {
	Header     byte
	Unk        [asciiGlyphCount]byte
	Characters [asciiGlyphCount]Rune
	Height     int
}

// Rune returns the FontRune for a given ASCII character.
func (f *asciiFont) Rune(r rune) *Rune {
	idx := int(r) - asciiFirstRune
	idx = ((idx % asciiGlyphCount) + asciiGlyphCount) % asciiGlyphCount
	return &f.Characters[idx]
}

// Size returns the width and height of the text in pixels.
func (f *asciiFont) Size(text string) (int, int) {
	if text == "" {
		return 0, 0
	}

	w, h := 0, f.Height

	for _, r := range text {
		switch r {
		case ' ':
			w += unicodeSpaceWidth
			continue
		default:
			w += int(f.Rune(r).Width)
		}
	}

	// Add 1 pixel for spacing between characters
	w += len(text) - 1
	return w, h
}

// Text renders text using the SDK's Unicode font with hue coloring
func (s *SDK) Text(font Font, text string, hue int) image.Image {
	if text == "" {
		return nil
	}

	// Calculate text dimensions with 1 pixel spacing between characters
	width, height := font.Size(text)
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	// Render each character
	x := 0
	runes := []rune(text)
	for i, runeChar := range runes {
		fontRune := font.Rune(runeChar)
		if fontRune == nil || fontRune.Image == nil {
			x += 4
			continue // Skip unsupported characters or characters without images
		}

		// Apply hue coloring to the character image
		charImg := s.applyHueToImage(fontRune.Image, hue)

		// Draw the character at the correct position
		charX := x + int(fontRune.XOffset)
		charY := int(fontRune.YOffset)

		// Ensure we don't draw outside bounds
		if charX >= 0 && charY >= 0 && charX < width && charY < height {
			draw.Draw(img,
				image.Rect(charX, charY, charX+charImg.Bounds().Dx(), charY+charImg.Bounds().Dy()),
				charImg,
				charImg.Bounds().Min,
				draw.Over)
		}

		x += int(fontRune.Width)

		// Add 1 pixel spacing between characters (but not after the last character)
		if i < len(runes)-1 {
			x += 1
		}
	}

	return img
}

// applyHueToImage applies a hue color to an image
func (s *SDK) applyHueToImage(src image.Image, hueIndex int) image.Image {
	if src == nil {
		return nil
	}
	if hueIndex == 0 || s == nil {
		return src
	}

	hue, err := s.Hue(hueIndex)
	if err != nil || hue == nil {
		// Fallback to original if hue not available
		return src
	}

	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)

	// Pick a palette color suited for text. Prefer the end of the table (brightest)
	// falling back to a mid-range if out of bounds.
	paletteIndex := int(hue.TableEnd)
	if paletteIndex < 0 || paletteIndex >= 32 {
		paletteIndex = 31
	}
	hueColor, herr := hue.GetColor(paletteIndex)
	if herr != nil {
		hueColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}
	hr, hg, hb, _ := hueColor.RGBA()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			sr, sg, sb, sa := src.At(x, y).RGBA()
			if sa == 0 {
				// Keep transparent pixels
				dst.Set(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 0})
				continue
			}

			// For Unicode font glyphs, source pixels are bitmask black (opaque) or transparent.
			// Apply the hue color directly and preserve the original alpha.
			na := uint8(sa >> 8)
			nr := uint8(hr >> 8)
			ng := uint8(hg >> 8)
			nb := uint8(hb >> 8)

			// If the glyph image contains grayscale intensity, we could modulate.
			// However, UO unicode fonts are binary; use full hue color for visibility.
			_ = sr
			_ = sg
			_ = sb

			dst.Set(x, y, color.NRGBA{R: nr, G: ng, B: nb, A: na})
		}
	}

	return dst
}
