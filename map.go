package ultima

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/kelindar/ultima-sdk/internal/uofile"
)

// Tile represents a single map tile (land or static).
type Tile struct {
	ID      uint16           // Tile ID
	Z       int8             // Elevation
	Flags   uint64           // Tile flags (from tiledata)
	Statics []StaticItemData // Statics at this tile (optional, for full fidelity)
}

// TileMap provides access to Ultima Online map data.
type TileMap struct {
	sdk         *SDK
	mapID       int
	width       int
	height      int
	mapFile     *uofile.File // internal: mapX.mul
	staticsFile *uofile.File // internal: staticsX.mul
	staIdxFile  *uofile.File // internal: staidxX.mul
	mutex       sync.Mutex
	loaded      bool
}

// NewTileMap initializes a TileMap for a given map index and files.
func NewTileMap(mapID int, mapFile, staticsFile, staIdxFile *uofile.File, width, height int) *TileMap {
	return &TileMap{
		mapID:       mapID,
		width:       width,
		height:      height,
		mapFile:     mapFile,
		staticsFile: staticsFile,
		staIdxFile:  staIdxFile,
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
func decodeMapBlock(data []byte) (*LandBlock, error) {
	if len(data) != 196 {
		return nil, fmt.Errorf("decodeMapBlock: expected 196 bytes, got %d", len(data))
	}
	var block LandBlock
	for i := 0; i < 64; i++ {
		off := 4 + i*3 // skip 4-byte header
		block.Tiles[i].ID = binary.LittleEndian.Uint16(data[off : off+2])
		block.Tiles[i].Z = int8(data[off+2])
	}
	return &block, nil
}

// TileAt returns the tile at the given x, y coordinate.
func (m *TileMap) TileAt(x, y int) (*Tile, error) {
	if x < 0 || y < 0 || x >= m.width || y >= m.height {
		return nil, fmt.Errorf("TileAt: coordinates out of bounds (%d,%d)", x, y)
	}
	// UO map files are organized into 8x8 blocks
	blockX, blockY := x/8, y/8
	blockIndex := blockY*(m.width/8) + blockX

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// DEBUG: Print blockIndex and max entry index
	entries := m.mapFile.Entries()
	maxEntry := -1
	count := 0
	for entry := range entries {
		if int(entry) > maxEntry {
			maxEntry = int(entry)
		}
		count++
	}
	fmt.Printf("TileAt: mapID=%d x=%d y=%d blockIndex=%d maxEntry=%d count=%d\n", m.mapID, x, y, blockIndex, maxEntry, count)

	// Use the unified Read interface for both UOP and MUL
	blockData, _, err := m.mapFile.Read(uint32(blockIndex))
	if err != nil {
		return nil, fmt.Errorf("TileAt: failed reading map block: %w", err)
	}
	landBlock, err := decodeMapBlock(blockData)
	if err != nil {
		return nil, err
	}
	tileIndex := (y%8)*8 + (x % 8)
	tileID := landBlock.Tiles[tileIndex].ID
	z := landBlock.Tiles[tileIndex].Z

	// TODO: Lookup flags from tiledata if needed
	// --- Statics reading ---
	statics := make([]StaticItemData, 0, 4) // Preallocate for typical count
	blockIndex32 := uint32(blockIndex)
	idxEntry, _, err := m.staIdxFile.Read(blockIndex32)
	if err == nil && len(idxEntry) >= 12 {
		//offset := int64(binary.LittleEndian.Uint32(idxEntry[0:4]))
		//length := int(binary.LittleEndian.Uint32(idxEntry[4:8]))
		//int32 extra := int32(binary.LittleEndian.Uint32(idxEntry[8:12])) // unused
		/*if offset > 0 && length > 0 {
			data, err := m.staticsFile.ReadBlockAt(offset, length)
			if err == nil {
				for i := 0; i+7 <= len(data); i += 7 {
					// Parse static entry
					id := binary.LittleEndian.Uint16(data[i : i+2])
					xoff := data[i+2]
					yoff := data[i+3]
					// zstat := int8(data[i+4]) // statics z, not used in StaticItemData
					hue := byte(data[i+5]) // tiledata uses byte for hue
					if int(xoff) == x%8 && int(yoff) == y%8 {
						staticData, err := m.sdk.StaticTile(int(id))
						if err == nil {
							// Overwrite fields that come from the statics entry
							staticData.Hue = hue
							staticData.Height = staticData.Height // tiledata height
							statics = append(statics, staticData)
						} else {
							// Fallback: fill only hue
							statics = append(statics, StaticItemData{Hue: hue})
						}
					}
				}
			}
		}*/
	}

	tile := &Tile{
		ID:      tileID,
		Z:       z,
		Flags:   0, // TODO: fill from tiledata
		Statics: statics,
	}
	return tile, nil
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
	staIdxFile, err := s.load([]string{
		fmt.Sprintf("staidx%d.mul", mapID),
	}, 0, uofile.WithIndexLength(12))
	if err != nil {
		return nil, fmt.Errorf("loadTileMap: failed to load staidx file: %w", err)
	}
	width, height := detectMapSize(s, mapID)
	return &TileMap{
		sdk:         s,
		mapID:       mapID,
		width:       width,
		height:      height,
		mapFile:     mapFile,
		staticsFile: staticsFile,
		staIdxFile:  staIdxFile,
	}, nil
}

// detectMapSize returns the width and height for a given map ID, checking for extended maps.
func detectMapSize(s *SDK, mapID int) (width, height int) {
	switch mapID {
	case 0: // Felucca
		return 6144, 4096
	case 1: // Trammel
		// Check for extended map file
		if s.fileExists("map1.mul") || s.fileExists("map1legacymul.uop") {
			return 7168, 4096
		}
		return 6144, 4096
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

// --- Helper logic for map sizes, detection, etc. ---
// TODO: Port logic from MapHelper.cs to detect map sizes, legacy vs. extended maps, etc.

// --- Tests ---
// TODO: Write extensive tests for map data reading, verifying against C# tile/static details at specific coordinates.
