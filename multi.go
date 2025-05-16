// Package ultima provides access to Ultima Online multi structures.
package ultima

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"os"
	"sort"

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

func savePng(img image.Image, name string) error {
	file, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	return png.Encode(file, img)
}

// Image renders the multi structure as a full image using art tiles for each MultiItem.
// The image bounds are computed from the offsets of all items. Each item's art is fetched using sdk.ArtTile,
// and composited at the correct position. The method returns an ARGB1555 image.
func (m *Multi) Image() (image.Image, error) {
	if len(m.Items) == 0 {
		return nil, fmt.Errorf("multi has no items")
	}

	// Tile dimensions in pixels
	const tileWidth, tileHeight = 44, 44

	// Compute bounds in tiles (using max/min int16 values)
	minX, minY, maxX, maxY := int16(32767), int16(32767), int16(-32768), int16(-32768)
	for _, item := range m.Items {
		minX = min(minX, item.OffsetX)
		minY = min(minY, item.OffsetY)
		maxX = max(maxX, item.OffsetX)
		maxY = max(maxY, item.OffsetY)
	}

	// Calculate dimensions in pixels
	width := int(maxX-minX+1) * tileWidth
	height := int(maxY-minY+1) * tileHeight
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid multi bounds: width=%d, height=%d", width, height)
	}

	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))

	// Sort items by OffsetZ (and ItemID for stability)
	items := make([]MultiItem, len(m.Items))
	copy(items, m.Items)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].OffsetZ != items[j].OffsetZ {
			return items[i].OffsetZ < items[j].OffsetZ
		}
		return items[i].ItemID < items[j].ItemID
	})

	// Composite each item with bottom-center alignment
	for _, item := range items {
		art, err := m.sdk.StaticArtTile(int(item.ItemID))
		if err != nil || art == nil {
			continue // skip missing art
		}

		tileImg, err := art.Image()
		if err != nil || tileImg == nil {
			continue // skip missing image
		}

		tileBounds := tileImg.Bounds()
		artW := tileBounds.Dx()
		artH := tileBounds.Dy()

		// Calculate position in pixels (bottom-center of the tile)
		drawX := int(item.OffsetX-minX) * tileWidth
		drawY := int(item.OffsetY-minY) * tileHeight

		// Center the art within the tile (bottom-center aligned)
		drawX += (tileWidth - artW) / 2
		drawY += tileHeight - artH

		// Draw the art
		for ty := 0; ty < artH; ty++ {
			for tx := 0; tx < artW; tx++ {
				px := drawX + tx
				py := drawY + ty
				if px < 0 || py < 0 || px >= width || py >= height {
					continue
				}
				img.Set(px, py, tileImg.At(tileBounds.Min.X+tx, tileBounds.Min.Y+ty))
			}
		}
	}

	return img, nil
}

// Multi returns a Multi structure by id, loading from multi.mul/multi.idx
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

	// Assume UOAHS format for now
	const entrySize = 16

	// Parse multi data
	var items []MultiItem
	for i := 0; i+entrySize <= len(data); i += entrySize {
		items = append(items, MultiItem{
			ItemID:  binary.LittleEndian.Uint16(data[i:]),
			OffsetX: int16(binary.LittleEndian.Uint16(data[i+2:])),
			OffsetY: int16(binary.LittleEndian.Uint16(data[i+4:])),
			OffsetZ: int16(binary.LittleEndian.Uint16(data[i+6:])),
			Flags:   binary.LittleEndian.Uint32(data[i+8:]),
			Unk1:    binary.LittleEndian.Uint32(data[i+12:]),
		})
	}
	return &Multi{
		sdk:   s,
		Items: items,
	}, nil
}
