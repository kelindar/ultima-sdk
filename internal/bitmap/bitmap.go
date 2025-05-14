package bitmap

import (
	"image"
	"image/color"
)

// ARGB1555Color represents a 16-bit color in ARGB 1-5-5-5 format.
// The highest bit is Alpha (0=transparent, 1=opaque).
// The next 5 bits are Red, then 5 bits Green, then 5 bits Blue.
type ARGB1555Color uint16

// RGBA converts ARGB-1555 to 32-bit RGBA.
//
// UO: bit-15 is unused, colour==0 ⇒ transparent.
func (c ARGB1555Color) RGBA() (r, g, b, a uint32) {
	if c == 0 {
		return 0, 0, 0, 0 // fully transparent
	}
	// colour present → fully opaque
	red5 := uint32(c>>10) & 0x1F
	green5 := uint32(c>>5) & 0x1F
	blue5 := uint32(c) & 0x1F

	// exact 5→8-bit upscale with rounding
	const scale = 255
	const max5 = 31
	r = (red5 * scale / max5) << 8
	g = (green5 * scale / max5) << 8
	b = (blue5 * scale / max5) << 8
	a = 0xFFFF
	return
}

// ARGB1555Model is the color model for ARGB1555 colors.
var ARGB1555Model color.Model = color.ModelFunc(argb1555Model)

func argb1555Model(c color.Color) color.Color {
	if _, ok := c.(ARGB1555Color); ok {
		return c // Already in the correct format
	}

	// Convert standard color.Color to ARGB1555Color
	r, g, b, a := c.RGBA()

	// Scale 16-bit channels (0-65535) down to 5-bit (0-31)
	// Formula: val5 = (val16 * 31) / 65535
	// Simplified: val5 = val16 >> 11
	red5 := uint16(r >> 11)
	green5 := uint16(g >> 11)
	blue5 := uint16(b >> 11)

	// Scale 16-bit alpha (0-65535) down to 1-bit (0 or 1)
	// Consider alpha > 0x8000 (half) as opaque (1)
	alpha1 := uint16(0)
	if a >= 0x8000 {
		alpha1 = 1
	}

	return ARGB1555Color((alpha1 << 15) | (red5 << 10) | (green5 << 5) | blue5)
}

// ARGB1555 is an in-memory image whose pixels are ARGB1555Color values.
type ARGB1555 struct {
	Pix    []byte          // Pix holds the image's pixels, as ARGB1555 (uint16) values stored in big-endian format.
	Stride int             // Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Rect   image.Rectangle // Rect is the image's bounds.
}

// NewARGB1555 returns a new ARGB1555 image with the given bounds.
func NewARGB1555(r image.Rectangle) *ARGB1555 {
	w, h := r.Dx(), r.Dy()
	stride := w * 2 // 2 bytes per pixel
	pix := make([]byte, stride*h)
	return &ARGB1555{Pix: pix, Stride: stride, Rect: r}
}

// ColorModel implements the Image interface.
func (p *ARGB1555) ColorModel() color.Model {
	return ARGB1555Model
}

// Bounds implements the Image interface.
func (p *ARGB1555) Bounds() image.Rectangle {
	return p.Rect
}

// At implements the Image interface.
// It returns the color of the pixel at (x, y).
// At(Bounds().Min.X, Bounds().Min.Y) returns the upper-left pixel of the grid.
// At(Bounds().Max.X-1, Bounds().Max.Y-1) returns the lower-right pixel.
func (p *ARGB1555) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.Rect)) {
		return ARGB1555Color(0) // Transparent black for out-of-bounds
	}
	// Calculate the byte offset for the pixel (x, y)
	// Each pixel is 2 bytes (uint16)
	offset := (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*2

	// Read the 16 bits (2 bytes) in little-endian format
	// UO files use little-endian for 16-bit colors.
	pixelValue := uint16(p.Pix[offset]) | uint16(p.Pix[offset+1])<<8
	return ARGB1555Color(pixelValue)
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *ARGB1555) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*2
}

// Set sets the color of the pixel at (x, y).
func (p *ARGB1555) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return // Ignore out-of-bounds writes
	}
	offset := p.PixOffset(x, y)
	colorARGB1555 := ARGB1555Model.Convert(c).(ARGB1555Color)

	// Write the 16 bits (2 bytes) in little-endian format
	p.Pix[offset] = byte(colorARGB1555)
	p.Pix[offset+1] = byte(colorARGB1555 >> 8)
}

// SubImage returns an image representing the portion of the image p visible
// through r. The returned value shares pixels with the original image.
func (p *ARGB1555) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(p.Rect)
	if r.Empty() {
		return &ARGB1555{}
	}
	offset := p.PixOffset(r.Min.X, r.Min.Y)
	return &ARGB1555{
		Pix:    p.Pix[offset:],
		Stride: p.Stride,
		Rect:   r,
	}
}

// Opaque scans the entire image and reports whether it is fully opaque.
// In ARGB1555, opaque means the highest bit is always 1.
func (p *ARGB1555) Opaque() bool {
	if p.Rect.Empty() {
		return true
	}

	for y := p.Rect.Min.Y; y < p.Rect.Max.Y; y++ {
		offset := p.PixOffset(p.Rect.Min.X, y)
		for x := 0; x < p.Rect.Dx(); x++ {
			if p.Pix[offset+1]&0x80 == 0 {
				return false
			}
			offset += 2
		}
	}
	return true
}
