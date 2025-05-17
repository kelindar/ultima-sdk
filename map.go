package ultima

import (
	"encoding/binary"
	"fmt"
	"image"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
	"github.com/kelindar/ultima-sdk/internal/uofile"
)

const blocksPerEntry = 4096

// StaticItem represents a single static placed on the map.
type StaticItem struct {
	ID  uint16 // Static tile ID
	X   uint8  // X offset within the block
	Y   uint8  // Y offset within the block
	Z   int8   // Elevation
	Hue uint16 // Color hue
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

// LandBlock represents a single 8x8 block of land tiles (from UOP/MUL map block).
type LandBlock struct {
	Tiles [64]struct {
		ID uint16
		Z  int8
	}
}

// decodeMapBlock parses a 196-byte map block and returns a LandBlock struct.
func decodeMapBlock(block []byte) (*LandBlock, error) {
	if len(block) != 192 {
		return nil, fmt.Errorf("decodeMapBlock: expected 196 bytes, got %d", len(block))
	}

	var out LandBlock
	for i := 0; i < 64; i++ {
		off := i * 3
		out.Tiles[i].ID = binary.LittleEndian.Uint16(block[off : off+2])
		out.Tiles[i].Z = int8(block[off+2])
	}
	return &out, nil
}

// decodeMapTile parses a single tile from a 196-byte map block, including statics.
func decodeMapTile(block []byte, tileIndex int, statics []StaticItem) (*Tile, error) {
	if len(block) < 196 {
		return nil, fmt.Errorf("decodeMapTile: expected 196 bytes, got %d", len(block))
	}
	tileData := block[tileIndex*3 : tileIndex*3+3]
	// Calculate x, y within the block
	x := tileIndex % 8
	y := tileIndex / 8
	// Filter statics for this tile
	var tileStatics []StaticItem
	for _, s := range statics {
		if int(s.X) == x && int(s.Y) == y {
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

	// Read the entry and check if it's valid
	entry, _, err := m.mapFile.Read(uint32(entryIndex))
	switch {
	case err != nil:
		return nil, fmt.Errorf("TileAt: failed reading UOP entry: %w", err)
	case len(entry) < 4+(blockOffset+1)*196:
		return nil, fmt.Errorf("TileAt: entry too small for block offset (entry len=%d, needed=%d)", len(entry), 4+(blockOffset+1)*196)
	}

	// Get the block data
	blockStart := 4 + blockOffset*196
	blockData := entry[blockStart : blockStart+196]
	tileIndex := (y%8)*8 + (x % 8)

	// Read statics for this block
	statics, err := m.readStatics(blockIndex)
	if err != nil {
		return nil, fmt.Errorf("TileAt: failed to read statics: %w", err)
	}
	return decodeMapTile(blockData, tileIndex, statics)
}

// readStatics reads and parses statics for a given block index.
func (m *TileMap) readStatics(blockIndex int) ([]StaticItem, error) {
	data, _, err := m.staticsFile.Read(uint32(blockIndex))
	if err != nil || len(data) == 0 {
		return nil, nil
	}

	count := len(data) / 7
	statics := make([]StaticItem, 0, count)
	for i := 0; i < count; i++ {
		off := i * 7
		if off+7 > len(data) {
			break
		}
		id := binary.LittleEndian.Uint16(data[off : off+2])
		item := StaticItem{
			ID:  id,
			X:   data[off+2],
			Y:   data[off+3],
			Z:   int8(data[off+4]),
			Hue: binary.LittleEndian.Uint16(data[off+5 : off+7]),
		}
		statics = append(statics, item)
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

	for entry := range m.mapFile.Entries() {
		data, _, err := m.mapFile.Read(uint32(entry))
		switch {
		case err != nil:
			return nil, fmt.Errorf("map.Image: failed reading entry %d: %w", entry, err)
		case len(data)%196 != 0:
			return nil, fmt.Errorf("map.Image: entry %d has invalid length (%d bytes)", entry, len(data))
		}

		length := len(data) / 196
		for blockIndex := 0; blockIndex < length; blockIndex++ {
			blockAbs := int(entry)*length + blockIndex
			blockX := blockAbs / blocksDown
			blockY := blockAbs % blocksDown
			blockData := data[blockIndex*196 : blockIndex*196+196]
			landBlock, err := decodeMapBlock(blockData[4:])
			if err != nil {
				return nil, fmt.Errorf("map.Image: %w", err)
			}

			for i, t := range landBlock.Tiles {
				x0 := (i % 8) + blockX*8
				y0 := (i / 8) + blockY*8
				rc, err := m.sdk.RadarColor(int(t.ID))
				if err != nil {
					continue
				}
				img.Set(x0, y0, rc.GetColor())
			}
		}
	}
	return img, nil
}
