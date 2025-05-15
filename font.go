package ultima

import (
	"image"
	"image/color"

	"fmt"
)

// FontCharacterInfo represents a single character's metadata and bitmap.
type FontCharacterInfo struct {
	Width   int
	Height  int
	XOffset int         // For Unicode
	YOffset int         // For Unicode
	Bitmap  image.Image // ARGB1555, nil if not present
}

// Font is the interface for font types (ASCII, Unicode)
type Font interface {
	Character(rune) *FontCharacterInfo
	Size(string) (int, int)
}

// FontUnicode loads a Unicode font from unifont*.mul using the SDK file loader.
func (s *SDK) FontUnicode() (Font, error) {
	file, err := s.load([]string{"unifont1.mul"}, 0) // TODO: support multiple font files if needed
	if err != nil {
		return nil, fmt.Errorf("load unifont1.mul: %w", err)
	}
	defer file.Close()

	data, _, err := file.Read(0)
	if err != nil {
		return nil, fmt.Errorf("read unifont1.mul: %w", err)
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
			font.Characters[i] = &FontCharacterInfo{}
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
		var bmp image.Image
		if width > 0 && height > 0 {
			bytesPerRow := (width + 7) / 8
			dataLen := height * bytesPerRow
			if int(offset)+4+dataLen > len(data) {
				return nil, fmt.Errorf("char data out of bounds at %d", i)
			}
			charData := data[offset+4 : offset+4+int32(dataLen)]
			bmp = decodeUnicodeBitmap(width, height, charData)
		}
		font.Characters[i] = &FontCharacterInfo{
			Width:   width,
			Height:  height,
			XOffset: xOff,
			YOffset: yOff,
			Bitmap:  bmp,
		}
	}
	return font, nil
}

// decodeUnicodeBitmap decodes Unicode font bitmap (bit-packed)
func decodeUnicodeBitmap(width, height int, data []byte) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := (x / 8) + (y * ((width + 7) / 8))
			if offset >= len(data) {
				continue
			}
			bit := (data[offset] & (1 << (7 - (x % 8)))) != 0
			if bit {
				img.SetNRGBA(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 0xFF})
			} else {
				img.SetNRGBA(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 0x00})
			}
		}
	}
	return img
}

// Font loads all ASCII fonts from fonts.mul using the SDK file loader.
func (s *SDK) Font() ([]Font, error) {
	file, err := s.load([]string{"fonts.mul"}, 0)
	if err != nil {
		return nil, fmt.Errorf("load fonts.mul: %w", err)
	}
	defer file.Close()

	data, _, err := file.Read(0)
	if err != nil {
		return nil, fmt.Errorf("read fonts.mul: %w", err)
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
			var bmp image.Image
			if width > 0 && height > 0 {
				pixLen := width * height * 2
				if offset+pixLen > len(data) {
					return nil, fmt.Errorf("char pixels out of bounds at font %d char %d", i, k)
				}
				pix := data[offset : offset+pixLen]
				offset += pixLen
				// Convert pix (ARGB1555) to image.Image
				bmp = decodeARGB1555(width, height, pix)
				if height > fonts[i].Height && k < 96 {
					fonts[i].Height = height
				}
			}
			fonts[i].Characters[k] = &FontCharacterInfo{
				Width:  width,
				Height: height,
				Bitmap: bmp,
			}
		}
	}
	out := make([]Font, 10)
	for i := range fonts {
		out[i] = fonts[i]
	}
	return out, nil
}

// decodeARGB1555 converts ARGB1555 bytes to image.NRGBA
func decodeARGB1555(width, height int, data []byte) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			off := 2 * (y*width + x)
			if off+1 >= len(data) {
				continue
			}
			pix := uint16(data[off]) | uint16(data[off+1])<<8
			// ARGB1555: A(1), R(5), G(5), B(5)
			a := uint8(0xFF)
			if (pix & 0x8000) == 0 {
				a = 0x00
			}
			r := uint8((pix>>10)&0x1F) << 3
			g := uint8((pix>>5)&0x1F) << 3
			b := uint8(pix&0x1F) << 3
			img.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: a})
		}
	}
	return img
}

// unicodeFont implements Font for Unicode fonts (unifont*.mul)
type unicodeFont struct {
	Characters [0x10000]*FontCharacterInfo
}

func (f *unicodeFont) Character(r rune) *FontCharacterInfo {
	idx := int(r) % 0x10000
	return f.Characters[idx]
}

func (f *unicodeFont) Size(text string) (int, int) {
	w, h := 0, 0
	for _, r := range text {
		c := f.Character(r)
		if c == nil {
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
	Characters [224]*FontCharacterInfo
	Height     int
}

func (f *asciiFont) Character(r rune) *FontCharacterInfo {
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
