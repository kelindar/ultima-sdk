// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"encoding/binary"
	"fmt"
	"image"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
	"github.com/kelindar/ultima-sdk/internal/uofile"
)

const blocksPerEntry = 4096

/*
type StaticItem struct {
	ID  uint16 // Static tile ID
	X   uint8  // X offset within the block
	Y   uint8  // Y offset within the block
	Z   int8   // Elevation
	Hue uint16 // Color hue
}

item := StaticItem{
	ID:  id,
	X:   data[off+2],
	Y:   data[off+3],
	Z:   int8(data[off+4]),
	Hue: binary.LittleEndian.Uint16(data[off+5 : off+7]),
}
*/

// StaticItem represents a single static placed on the map.
type StaticItem []byte

// ID returns the static ID
func (s *StaticItem) ID() uint16 {
	return binary.LittleEndian.Uint16((*s)[:2])
}

// Location returns the static location
func (s *StaticItem) Location() (x, y uint8, z int8) {
	return (*s)[2], (*s)[3], int8((*s)[4])
}

// Hue returns the static hue
func (s *StaticItem) Hue() uint16 {
	return binary.LittleEndian.Uint16((*s)[5:7])
}

// Tile represents a single map tile, including statics.
type Tile struct {
	ID      uint16       // Land tile ID
	Z       int8         // Land tile elevation
	Statics []StaticItem // Statics located at this tile
}

// TileMap provides access to Ultima Online map data.
type TileMap struct {
	sdk           *SDK
	mapID         int
	width, height int
	mapFile       *uofile.File // internal: mapX.mul
	staticsFile   *uofile.File // internal: staticsX.mul + staidxX.mul
}

// NewTileMap initializes a TileMap for a given map index and files.
func NewTileMap(mapID int, mapFile, staticsFile *uofile.File, width, height int) *TileMap {
	return &TileMap{
		mapID:       mapID,
		width:       width,
		height:      height,
		mapFile:     mapFile,
		staticsFile: staticsFile,
	}
}

// decodeMapTile parses a single tile from a 196-byte map block, including statics.
func decodeMapTile(block []byte, tileIndex int, statics []StaticItem) (*Tile, error) {
	if len(block) < 196 {
		return nil, fmt.Errorf("decodeMapTile: expected 196 bytes, got %d", len(block))
	}

	tileData := block[tileIndex*3 : tileIndex*3+3]
	x := tileIndex % 8
	y := tileIndex / 8

	// Filter statics for this tile
	var tileStatics []StaticItem
	for _, s := range statics {
		sx, sy, _ := s.Location()
		if int(sx) == x && int(sy) == y {
			tileStatics = append(tileStatics, s)
		}
	}

	return &Tile{
		ID:      binary.LittleEndian.Uint16(tileData[:2]),
		Z:       int8(tileData[2]),
		Statics: tileStatics,
	}, nil
}

// TileAt returns the tile at the given x, y coordinate.
// TileAt returns the tile at the given x, y coordinate, including statics.
func (m *TileMap) TileAt(x, y int) (*Tile, error) {
	if x < 0 || y < 0 || x >= m.width || y >= m.height {
		return nil, fmt.Errorf("TileAt: coordinates out of bounds (%d,%d)", x, y)
	}

	// Calculate the block index (column-major) and entry index
	blocksDown := m.height / 8
	blockX, blockY := x/8, y/8
	blockIndex := blockX*blocksDown + blockY
	entryIndex := blockIndex / blocksPerEntry
	blockOffset := blockIndex % blocksPerEntry
	blockStart := 4 + blockOffset*196
	tileIndex := (y%8)*8 + (x % 8)

	// Read the entry and check if it's valid
	entry, err := m.mapFile.Entry(uint32(entryIndex))
	switch {
	case err != nil:
		return nil, fmt.Errorf("TileAt: failed reading UOP entry: %w", err)
	case entry.Len() < 4+(blockOffset+1)*196:
		return nil, fmt.Errorf("TileAt: entry too small for block offset (entry len=%d, needed=%d)", entry.Len(), 4+(blockOffset+1)*196)
	}

	// Get the block data
	buffer, release := uofile.Borrow(196)
	defer release()

	n, err := entry.ReadAt(buffer, int64(blockStart))
	switch {
	case err != nil:
		return nil, fmt.Errorf("TileAt: failed reading entry: %w", err)
	case n < 196:
		return nil, fmt.Errorf("TileAt: entry too small for block offset (entry len=%d, needed=%d)", n, 196)
	}

	// Read statics for this block
	statics, err := m.readStatics(blockIndex)
	if err != nil {
		return nil, fmt.Errorf("TileAt: failed to read statics: %w", err)
	}
	return decodeMapTile(buffer, tileIndex, statics)
}

// readStatics reads and parses statics for a given block index.
func (m *TileMap) readStatics(blockIndex int) ([]StaticItem, error) {
	entry, err := m.staticsFile.Entry(uint32(blockIndex))
	switch {
	case err != nil:
		return nil, fmt.Errorf("readStatics: failed reading UOP entry: %w", err)
	case entry == nil:
		return nil, nil
	}

	buffer := make([]byte, entry.Len())
	_, err = entry.ReadAt(buffer, 0)
	if err != nil {
		return nil, fmt.Errorf("readStatics: failed reading entry: %w", err)
	}

	statics := make([]StaticItem, 0, entry.Len()/7)
	for i := 0; i < entry.Len()/7; i++ {
		statics = append(statics, StaticItem(buffer[i*7:i*7+7]))
	}

	return statics, nil
}

// Map returns the TileMap for the given map index, loading if necessary.
func (s *SDK) Map(mapID int) (*TileMap, error) {
	return s.loadTileMap(mapID)
}

// loadTileMap loads and returns a TileMap for the given map ID.
func (s *SDK) loadTileMap(mapID int) (*TileMap, error) {
	mapFile, err := s.loadMap(mapID)
	if err != nil {
		return nil, fmt.Errorf("loadTileMap: failed to load map file: %w", err)
	}
	staticsFile, err := s.loadStatics(mapID)
	if err != nil {
		return nil, fmt.Errorf("loadTileMap: failed to load statics file: %w", err)
	}
	width, height := detectMapSize(mapID)
	return &TileMap{
		sdk:         s,
		mapID:       mapID,
		width:       width,
		height:      height,
		mapFile:     mapFile,
		staticsFile: staticsFile,
	}, nil
}

// detectMapSize returns the width and height for a given map ID, checking for extended maps.
func detectMapSize(mapID int) (width, height int) {
	switch mapID {
	case 0: // Felucca
		return 6144, 4096
	case 1: // Trammel
		return 7168, 4096
	case 2: // Ilshenar
		return 2304, 1600
	case 3: // Malas
		return 2560, 2048
	case 4: // Tokuno
		return 1448, 1448
	case 5: // TerMur
		return 1280, 4096
	default:
		return 6144, 4096 // fallback
	}
}

// Image renders the map as a radar-color overview (1 pixel per tile).
func (m *TileMap) Image() (image.Image, error) {
	img := bitmap.NewARGB1555(image.Rect(0, 0, m.width, m.height))
	blocksDown := m.height / 8

	buffer := make([]byte, 196*blocksPerEntry)
	for key := range m.mapFile.Entries() {
		entry, err := m.mapFile.Entry(uint32(key))
		switch {
		case err != nil:
			return nil, fmt.Errorf("map.Image: failed reading entry %d: %w", key, err)
		case entry.Len()%196 != 0:
			return nil, fmt.Errorf("map.Image: entry %d has invalid length (%d bytes)", key, entry.Len())
		}

		n, err := entry.ReadAt(buffer, 0)
		if err != nil {
			return nil, fmt.Errorf("map.Image: failed reading entry %d: %w", key, err)
		}

		length := n / 196
		for blockIndex := 0; blockIndex < length; blockIndex++ {
			blockAbs := int(key)*length + blockIndex
			blockX := blockAbs / blocksDown
			blockY := blockAbs % blocksDown
			blockData := buffer[blockIndex*196 : blockIndex*196+196]
			if len(blockData) < 4+192 {
				return nil, fmt.Errorf("map.Image: block %d too short (%d bytes)", blockAbs, len(blockData))
			}

			// Tiles start at offset 4, each tile is 3 bytes (id:2, z:1)
			tiles := blockData[4:]
			for i := 0; i < 64; i++ {
				off := i * 3
				tileID := binary.LittleEndian.Uint16(tiles[off : off+2])
				x0 := (i % 8) + blockX*8
				y0 := (i / 8) + blockY*8
				rc, err := m.sdk.RadarColor(int(tileID))
				if err != nil {
					continue
				}

				img.Set(x0, y0, rc.GetColor())
			}
		}
	}
	return img, nil
}
