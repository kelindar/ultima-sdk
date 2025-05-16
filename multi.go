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

	// Isometric tile placement constants
	const isoX, isoY = 22, 22

	// First pass: compute min/max drawX/drawY for canvas size
	minDrawX, minDrawY := 1<<31-1, 1<<31-1
	maxDrawX, maxDrawY := -(1<<31), -(1<<31)
	tilePositions := make([]struct {
		drawX, drawY, artW, artH int
		item MultiItem
	}, 0, len(m.Items))

	for _, item := range m.Items {
		art, err := m.sdk.StaticArtTile(int(item.ItemID))
		if err != nil || art == nil {
			continue
		}
		tileImg, err := art.Image()
		if err != nil || tileImg == nil {
			continue
		}
		tileBounds := tileImg.Bounds()
		artW := tileBounds.Dx()
		artH := tileBounds.Dy()

		tileX := int(item.OffsetX)
		tileY := int(item.OffsetY)
		drawX := (tileX - tileY) * isoX
		drawY := (tileX + tileY) * isoY
		drawY -= int(item.OffsetZ) * 4
		drawX -= artW / 2
		drawY -= artH

		if drawX < minDrawX {
			minDrawX = drawX
		}
		if drawY < minDrawY {
			minDrawY = drawY
		}
		if drawX+artW > maxDrawX {
			maxDrawX = drawX + artW
		}
		if drawY+artH > maxDrawY {
			maxDrawY = drawY + artH
		}
		tilePositions = append(tilePositions, struct {
			drawX, drawY, artW, artH int
			item MultiItem
		}{drawX, drawY, artW, artH, item})
	}

	width := maxDrawX - minDrawX
	height := maxDrawY - minDrawY
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid multi bounds: width=%d, height=%d", width, height)
	}

	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))

	// Sort items by OffsetZ (and ItemID for stability)
	sort.SliceStable(tilePositions, func(i, j int) bool {
		if tilePositions[i].item.OffsetZ != tilePositions[j].item.OffsetZ {
			return tilePositions[i].item.OffsetZ < tilePositions[j].item.OffsetZ
		}
		return tilePositions[i].item.ItemID < tilePositions[j].item.ItemID
	})

	// Second pass: draw tiles at adjusted positions
	for _, pos := range tilePositions {
		art, err := m.sdk.StaticArtTile(int(pos.item.ItemID))
		if err != nil || art == nil {
			continue
		}
		tileImg, err := art.Image()
		if err != nil || tileImg == nil {
			continue
		}
		tileBounds := tileImg.Bounds()
		drawX := pos.drawX - minDrawX
		drawY := pos.drawY - minDrawY
		// Debug output
		fmt.Printf("Draw ItemID=%d at iso (%d,%d) px (%d,%d)\n", pos.item.ItemID, pos.item.OffsetX, pos.item.OffsetY, drawX, drawY)
		for ty := 0; ty < pos.artH; ty++ {
			for tx := 0; tx < pos.artW; tx++ {
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
