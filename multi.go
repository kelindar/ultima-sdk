// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"image"
	"sort"
	"strconv"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
)

// MultiItem represents a single item within a multi-structure.
type MultiItem struct {
	Item   uint16 // Tile ID of the item.
	X      int16
	Y      int16
	Z      int16
	Flags  uint32
	Cliloc uint32 // Only present in UOAHS format (16 bytes per entry)
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
		art, err := m.sdk.StaticArtTile(int(item.Item))
		if err != nil || art == nil {
			continue
		}

		tileBounds := art.Image.Bounds()
		artW := tileBounds.Dx()
		artH := tileBounds.Dy()

		tileX := int(item.X)
		tileY := int(item.Y)
		drawX := (tileX - tileY) * isoX
		drawY := (tileX + tileY) * isoY
		drawY -= int(item.Z) * 4
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
		az, bz := int(a.item.Z), int(b.item.Z)
		ay, by := int(a.item.Y), int(b.item.Y)
		ax, bx := int(a.item.X), int(b.item.X)
		if az != bz {
			return az < bz
		}
		if ay != by {
			return ay < by
		}
		if ax != bx {
			return ax < bx
		}
		return a.item.Item < b.item.Item
	})

	// Second pass: draw tiles at adjusted positions
	for _, pos := range tilePositions {
		art, err := m.sdk.StaticArtTile(int(pos.item.Item))
		if err != nil || art == nil {
			continue
		}

		tileBounds := art.Image.Bounds()
		drawX := pos.drawX - minDrawX
		drawY := pos.drawY - minDrawY

		for ty := 0; ty < pos.artH; ty++ {
			for tx := 0; tx < pos.artW; tx++ {
				px := drawX + tx
				py := drawY + ty
				if px < 0 || py < 0 || px >= width || py >= height {
					continue
				}
				col := art.Image.At(tileBounds.Min.X+tx, tileBounds.Min.Y+ty)
				if c, ok := col.(bitmap.ARGB1555Color); ok && c != 0 {
					img.Set(px, py, c)
				}
			}
		}
	}

	return img, nil
}

// ToCSV exports all MultiItems to CSV format with headers: item, x, y, z, flags, cliloc.
// Returns the CSV data as bytes following the standard Go marshaling pattern.
func (m *Multi) ToCSV() ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV header
	header := []string{"item", "x", "y", "z", "flags", "cliloc"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("multi: failed to write CSV header: %w", err)
	}

	// Write each MultiItem as a CSV row
	for _, item := range m.Items {
		record := []string{
			strconv.FormatUint(uint64(item.Item), 10),
			strconv.FormatInt(int64(item.X), 10),
			strconv.FormatInt(int64(item.Y), 10),
			strconv.FormatInt(int64(item.Z), 10),
			strconv.FormatUint(uint64(item.Flags), 10),
			strconv.FormatUint(uint64(item.Cliloc), 10),
		}

		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("multi: failed to write CSV record: %w", err)
		}
	}

	// Flush the writer to ensure all data is written to the buffer
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("multi: failed to flush CSV writer: %w", err)
	}
	return buf.Bytes(), nil
}

// Multi returns a Multi structure by id, loading from multi.mul/multi.idx
func (s *SDK) Multi(id int) (*Multi, error) {
	file, err := s.loadMulti()
	if err != nil {
		return nil, err
	}

	data, err := file.ReadFull(uint32(id))
	switch {
	case err != nil:
		return nil, fmt.Errorf("multi entry %d not found: %w", id, err)
	case len(data) == 0:
		return nil, fmt.Errorf("multi entry %d not found", id)
	}

	// Assume UOAHS format for now
	const entrySize = 16

	// Parse multi data
	var items []MultiItem
	for i := 0; i+entrySize <= len(data); i += entrySize {
		items = append(items, MultiItem{
			Item:   binary.LittleEndian.Uint16(data[i:]),
			X:      int16(binary.LittleEndian.Uint16(data[i+2:])),
			Y:      int16(binary.LittleEndian.Uint16(data[i+4:])),
			Z:      int16(binary.LittleEndian.Uint16(data[i+6:])),
			Flags:  binary.LittleEndian.Uint32(data[i+8:]),
			Cliloc: binary.LittleEndian.Uint32(data[i+12:]),
		})
	}

	return &Multi{
		sdk:   s,
		Items: items,
	}, nil
}

// MultiFromCSV parses CSV data and returns a Multi structure.
// The CSV is expected to have columns: item, x, y, z, [flags], [cliloc].
// The first row is assumed to be a header and is skipped.
// The last 2 columns (flags and cliloc) are optional and will default to 0 if not present.
func (s *SDK) MultiFromCSV(data []byte) (*Multi, error) {
	reader := csv.NewReader(bytes.NewReader(data))

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("multi: failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("multi: CSV data is empty")
	}

	// Skip the first row (header) and parse data rows
	var items []MultiItem
	for rowNum, record := range records[1:] { // Skip header
		if len(record) < 4 {
			return nil, fmt.Errorf("multi: invalid CSV row %d, expected at least 4 columns (item,x,y,z), got %d", rowNum+2, len(record))
		}

		// Parse ItemID
		itemID, err := strconv.ParseUint(record[0], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("multi: invalid ItemID in row %d: %w", rowNum+2, err)
		}
		// Parse OffsetX
		offsetX, err := strconv.ParseInt(record[1], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("multi: invalid X in row %d: %w", rowNum+2, err)
		}
		// Parse OffsetY
		offsetY, err := strconv.ParseInt(record[2], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("multi: invalid Y in row %d: %w", rowNum+2, err)
		}		// Parse OffsetZ
		offsetZ, err := strconv.ParseInt(record[3], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("multi: invalid Z in row %d: %w", rowNum+2, err)
		}

		// Parse Flags (optional, defaults to 0)
		var flags uint32
		if len(record) > 4 {
			flagsVal, err := strconv.ParseUint(record[4], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("multi: invalid Flags in row %d: %w", rowNum+2, err)
			}
			flags = uint32(flagsVal)
		}		// Parse Unk1/cliloc (optional, defaults to 0)
		var cliloc uint32
		if len(record) > 5 {
			clilocVal, err := strconv.ParseUint(record[5], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("multi: invalid Unk1 in row %d: %w", rowNum+2, err)
			}
			cliloc = uint32(clilocVal)
		}
		
		items = append(items, MultiItem{
			Item:   uint16(itemID),
			X:      int16(offsetX),
			Y:      int16(offsetY),
			Z:      int16(offsetZ),
			Flags:  flags,
			Cliloc: cliloc,
		})
	}

	return &Multi{
		sdk:   s,
		Items: items,
	}, nil
}
