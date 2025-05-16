// Package ultima provides access to Ultima Online multi structures.
package ultima

import (
	"encoding/binary"
	"fmt"
	"image"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
)

// MultiItem represents a single item within a multi-structure.
type MultiItem struct {
	ItemID  uint16 // Tile ID of the item.
	OffsetX int16
	OffsetY int16
	OffsetZ int16
	Flags   uint32
	Unk1    uint32 // Only present in UOAHS format (16 bytes per entry)
}

// Multi represents a multi-structure (e.g., house, boat) in Ultima Online.
type Multi struct {
	sdk   *SDK
	Items []MultiItem
}

// Image renders the multi structure as a full image using art tiles for each MultiItem.
// The image bounds are computed from the offsets of all items. Each item's art is fetched using sdk.ArtTile,
// and composited at the correct position. The method returns an ARGB1555 image.
func (m *Multi) Image() (image.Image, error) {
	if len(m.Items) == 0 {
		return nil, fmt.Errorf("multi has no items")
	}

	// Compute bounds
	minX, minY, maxX, maxY := int(1<<15), int(1<<15), int(-1<<15), int(-1<<15)
	for _, item := range m.Items {
		x := int(item.OffsetX)
		y := int(item.OffsetY)
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}
	width := (maxX - minX) + 44  // 44: max art tile width
	height := (maxY - minY) + 44 // 44: max art tile height
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid multi bounds: width=%d, height=%d", width, height)
	}

	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))

	// Composite each item
	for _, item := range m.Items {
		art, err := m.sdk.ArtTile(int(item.ItemID))
		if err != nil || art == nil {
			continue // skip missing art
		}

		tileImg, err := art.Image()
		if err != nil || tileImg == nil {
			continue // skip missing image
		}

		// Compute top-left position for this item in the output image
		x := int(item.OffsetX) - minX
		y := int(item.OffsetY) - minY
		tileBounds := tileImg.Bounds()
		for ty := 0; ty < tileBounds.Dy(); ty++ {
			for tx := 0; tx < tileBounds.Dx(); tx++ {
				px := x + tx
				py := y + ty
				if px < 0 || py < 0 || px >= width || py >= height {
					continue
				}
				img.Set(px, py, tileImg.At(tileBounds.Min.X+tx, tileBounds.Min.Y+ty))
			}
		}
	}

	return img, nil
}

// Multi returns a Multi structure by id, loading from multi.mul/multi.idx via loadMulti().
// This follows the same pattern as other SDK data accessors.
func (s *SDK) Multi(id int) (*Multi, error) {
	file, err := s.loadMulti()
	if err != nil {
		return nil, err
	}
	data, _, err := file.Read(uint32(id))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("multi entry %d not found", id)
	}
	// TODO: Detect UOAHS format properly; for now, assume false
	isUOAHS := false
	entrySize := 12
	if isUOAHS {
		entrySize = 16
	}
	var items []MultiItem
	for i := 0; i+entrySize <= len(data); i += entrySize {
		item := MultiItem{
			ItemID:  binary.LittleEndian.Uint16(data[i:]),
			OffsetX: int16(binary.LittleEndian.Uint16(data[i+2:])),
			OffsetY: int16(binary.LittleEndian.Uint16(data[i+4:])),
			OffsetZ: int16(binary.LittleEndian.Uint16(data[i+6:])),
			Flags:   binary.LittleEndian.Uint32(data[i+8:]),
		}
		if isUOAHS {
			item.Unk1 = binary.LittleEndian.Uint32(data[i+12:])
		}
		items = append(items, item)
	}
	return &Multi{
		sdk:   s,
		Items: items,
	}, nil
}
