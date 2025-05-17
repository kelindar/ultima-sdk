// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"encoding/binary"
	"fmt"
	"image"
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
	maxDrawX, maxDrawY := -(1 << 31), -(1 << 31)
	tilePositions := make([]struct {
		drawX, drawY, artW, artH int
		item                     MultiItem
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
			item                     MultiItem
		}{drawX, drawY, artW, artH, item})
	}

	width := maxDrawX - minDrawX
	height := maxDrawY - minDrawY
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid multi bounds: width=%d, height=%d", width, height)
	}

	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))

	// Sort items by OffsetZ, then OffsetY, then OffsetX, then ItemID (matches UO stacking logic)
	sort.SliceStable(tilePositions, func(i, j int) bool {
		a, b := tilePositions[i], tilePositions[j]
		az, bz := int(a.item.OffsetZ), int(b.item.OffsetZ)
		ay, by := int(a.item.OffsetY), int(b.item.OffsetY)
		ax, bx := int(a.item.OffsetX), int(b.item.OffsetX)
		if az != bz {
			return az < bz
		}
		if ay != by {
			return ay < by
		}
		if ax != bx {
			return ax < bx
		}
		return a.item.ItemID < b.item.ItemID
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

		for ty := 0; ty < pos.artH; ty++ {
			for tx := 0; tx < pos.artW; tx++ {
				px := drawX + tx
				py := drawY + ty
				if px < 0 || py < 0 || px >= width || py >= height {
					continue
				}
				col := tileImg.At(tileBounds.Min.X+tx, tileBounds.Min.Y+ty)
				if c, ok := col.(bitmap.ARGB1555Color); ok && c != 0 {
					img.Set(px, py, c)
				}
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
