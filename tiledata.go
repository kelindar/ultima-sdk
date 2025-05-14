package ultima

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"iter"
	"os"

	"github.com/kelindar/ultima-sdk/internal/mul"
)

const (
	landOffset = 0xFFFFF
)

// TileFlag represents individual properties of tiles as bit flags.
type TileFlag uint64

// Tile flag constants
const (
	TileFlagNone        TileFlag = 0x00000000
	TileFlagBackground  TileFlag = 0x00000001
	TileFlagWeapon      TileFlag = 0x00000002
	TileFlagTransparent TileFlag = 0x00000004
	TileFlagTranslucent TileFlag = 0x00000008
	TileFlagWall        TileFlag = 0x00000010
	TileFlagDamaging    TileFlag = 0x00000020
	TileFlagImpassable  TileFlag = 0x00000040
	TileFlagWet         TileFlag = 0x00000080
	TileFlagUnknown1    TileFlag = 0x00000100
	TileFlagSurface     TileFlag = 0x00000200
	TileFlagBridge      TileFlag = 0x00000400
	TileFlagGeneric     TileFlag = 0x00000800
	TileFlagWindow      TileFlag = 0x00001000
	TileFlagNoShoot     TileFlag = 0x00002000
	TileFlagArticleA    TileFlag = 0x00004000
	TileFlagArticleAn   TileFlag = 0x00008000
	TileFlagArticleThe  TileFlag = 0x00010000
	TileFlagFoliage     TileFlag = 0x00020000
	TileFlagPartialHue  TileFlag = 0x00040000
	TileFlagNoHouse     TileFlag = 0x00080000
	TileFlagMap         TileFlag = 0x00100000
	TileFlagContainer   TileFlag = 0x00200000
	TileFlagWearable    TileFlag = 0x00400000
	TileFlagLightSource TileFlag = 0x00800000
	TileFlagAnimation   TileFlag = 0x01000000
	TileFlagHoverOver   TileFlag = 0x02000000
	TileFlagNoDiagonal  TileFlag = 0x04000000
	TileFlagArmor       TileFlag = 0x08000000
	TileFlagRoof        TileFlag = 0x10000000
	TileFlagDoor        TileFlag = 0x20000000
	TileFlagStairBack   TileFlag = 0x40000000
	TileFlagStairRight  TileFlag = 0x80000000

	// High dword flags (UO:AHS and newer clients)
	TileFlagAlphaBlend   TileFlag = 0x0100000000
	TileFlagUseNewArt    TileFlag = 0x0200000000
	TileFlagArtUsed      TileFlag = 0x0400000000
	TileFlagUnused8      TileFlag = 0x0800000000
	TileFlagNoShadow     TileFlag = 0x1000000000
	TileFlagPixelBleed   TileFlag = 0x2000000000
	TileFlagPlayAnimOnce TileFlag = 0x4000000000
	TileFlagMultiMovable TileFlag = 0x10000000000
)

// LandTileData represents the data for a single land tile.
type LandTileData struct {
	TextureID uint16   // Texture ID for the land tile
	Flags     TileFlag // Properties of this land tile
	Name      string   // Name of the tile
}

// StaticItemData represents the data for a single static item tile.
type StaticItemData struct {
	Flags          TileFlag // Properties of this static item
	Weight         byte     // Weight of the item
	Quality        byte     // Quality/Layer of the item
	Quantity       byte     // Quantity of the item
	Value          byte     // Value of the item
	Height         byte     // Height of the item
	Animation      int16    // Animation ID of the item
	Hue            byte     // Hue of the item
	StackingOffset byte     // Stacking offset if Generic flag is set
	Name           string   // Name of the item
	MiscData       int16    // Miscellaneous data
	Unk2           byte     // Unknown field 2
	Unk3           byte     // Unknown field 3
}

// Background returns whether the static item has the Background flag set
func (s StaticItemData) Background() bool {
	return s.Flags&TileFlagBackground != 0
}

// Bridge returns whether the static item has the Bridge flag set
func (s StaticItemData) Bridge() bool {
	return s.Flags&TileFlagBridge != 0
}

// Impassable returns whether the static item has the Impassable flag set
func (s StaticItemData) Impassable() bool {
	return s.Flags&TileFlagImpassable != 0
}

// Surface returns whether the static item has the Surface flag set
func (s StaticItemData) Surface() bool {
	return s.Flags&TileFlagSurface != 0
}

// Wearable returns whether the static item has the Wearable flag set
func (s StaticItemData) Wearable() bool {
	return s.Flags&TileFlagWearable != 0
}

// CalcHeight returns the calculated height of the item. For bridges, this is Height/2.
func (s StaticItemData) CalcHeight() int {
	if s.Flags&TileFlagBridge != 0 {
		return int(s.Height) / 2
	}
	return int(s.Height)
}

// readStringFromBytes reads a null-terminated string from a fixed-length byte array
func readStringFromBytes(b []byte) string {
	n := bytes.IndexByte(b, 0)
	if n == -1 {
		n = len(b)
	}
	return string(b[:n])
}

// LandTile returns a specific land tile's data by ID
func (s *SDK) LandTile(id int) (LandTileData, error) {
	if id < 0 || id >= 0x4000 {
		return LandTileData{}, fmt.Errorf("invalid land tile ID: %d", id)
	}

	file, err := s.loadTiledata()
	if err != nil {
		return LandTileData{}, err
	}

	data, _, err := file.Read(uint32(landOffset + id))
	if err != nil {
		return LandTileData{}, fmt.Errorf("error reading land tile data: %w", err)
	}

	return makeLandTileData(data, s.isNewTileDataFormat()), nil
}

// makeLandTileData converts raw bytes into a LandTileData struct
func makeLandTileData(data []byte, isNewFormat bool) LandTileData {
	var result LandTileData

	// Land tile format:
	// - flags: uint32/uint64 (depends on format)
	// - textureID: uint16
	// - name: char[20]
	if isNewFormat {
		result.Flags = TileFlag(binary.LittleEndian.Uint64(data[0:8]))
		result.TextureID = binary.LittleEndian.Uint16(data[8:10])
		result.Name = readStringFromBytes(data[10:30])
	} else {
		result.Flags = TileFlag(binary.LittleEndian.Uint32(data[0:4]))
		result.TextureID = binary.LittleEndian.Uint16(data[4:6])
		result.Name = readStringFromBytes(data[6:26])
	}

	return result
}

// StaticTile returns a specific static tile's data by ID
func (s *SDK) StaticTile(id int) (StaticItemData, error) {
	if id < 0 || id >= s.staticTileCount() {
		return StaticItemData{}, fmt.Errorf("invalid static tile ID: %d", id)
	}

	file, err := s.loadTiledata()
	if err != nil {
		return StaticItemData{}, err
	}

	data, _, err := file.Read(uint32(id))
	if err != nil {
		return StaticItemData{}, fmt.Errorf("error reading static tile data: %w", err)
	}

	return makeStaticItemData(data, s.isNewTileDataFormat()), nil
}

// makeStaticItemData converts raw bytes into a StaticItemData struct
func makeStaticItemData(data []byte, isNewFormat bool) StaticItemData {
	var result StaticItemData
	var offset int

	// Static tile format:
	// - flags: uint32/uint64 (depends on format)
	// - weight: byte
	// - quality: byte
	// - miscData: int16
	// - unk2: byte
	// - quantity: byte
	// - animation: int16
	// - unk3: byte
	// - hue: byte
	// - stackingOffset: byte
	// - value: byte
	// - height: byte
	// - name: char[20]
	if isNewFormat {
		result.Flags = TileFlag(binary.LittleEndian.Uint64(data[0:8]))
		offset = 8
	} else {
		result.Flags = TileFlag(binary.LittleEndian.Uint32(data[0:4]))
		offset = 4
	}

	result.Weight = data[offset]
	result.Quality = data[offset+1]
	result.MiscData = int16(binary.LittleEndian.Uint16(data[offset+2 : offset+4]))
	result.Unk2 = data[offset+4]
	result.Quantity = data[offset+5]
	result.Animation = int16(binary.LittleEndian.Uint16(data[offset+6 : offset+8]))
	result.Unk3 = data[offset+8]
	result.Hue = data[offset+9]
	result.StackingOffset = data[offset+10]
	result.Value = data[offset+11]
	result.Height = data[offset+12]
	result.Name = readStringFromBytes(data[offset+13 : offset+33])

	return result
}

// isNewTileDataFormat returns whether the newer tiledata format should be used
// This would typically be determined by client version, but for now we'll assume
// the newer format is used
func (s *SDK) isNewTileDataFormat() bool {
	// TODO: Implement proper client version checking similar to C#'s Art.IsUOAHS()
	// For now, we'll assume the newer format
	return true
}

// staticTileCount returns the number of static tiles in the tiledata file
// This is determined by the file size and format
func (s *SDK) staticTileCount() int {
	// Most clients have 0x10000 static tiles (65536)
	return 0x10000
}

// LandTiles returns an iterator over all land tiles
func (s *SDK) LandTiles() iter.Seq[LandTileData] {
	file, err := s.loadTiledata()
	if err != nil {
		return func(yield func(LandTileData) bool) {}
	}

	isNewFormat := s.isNewTileDataFormat()

	return func(yield func(LandTileData) bool) {
		for i := 0; i < 0x4000; i++ {
			data, _, err := file.Read(uint32(landOffset + i))
			if err != nil {
				continue
			}

			if !yield(makeLandTileData(data, isNewFormat)) {
				break
			}
		}
	}
}

// StaticTiles returns an iterator over all static tiles
func (s *SDK) StaticTiles() iter.Seq[StaticItemData] {
	file, err := s.loadTiledata()
	if err != nil {
		return func(yield func(StaticItemData) bool) {}
	}

	isNewFormat := s.isNewTileDataFormat()
	count := s.staticTileCount()

	return func(yield func(StaticItemData) bool) {
		for i := 0; i < count; i++ {
			data, _, err := file.Read(uint32(i))
			if err != nil {
				continue
			}

			if !yield(makeStaticItemData(data, isNewFormat)) {
				break
			}
		}
	}
}

// decodeTileDataFile loads the tiledata.mul file and populates the internal
// data structures for land and static tiles
func decodeTileDataFile(file *os.File, add mul.AddFn) error {
	// Calculate file size to help determine format and static entries count
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()

	// Read the entire file into memory for processing
	data := make([]byte, fileSize)
	_, err = file.ReadAt(data, 0)
	if err != nil {
		return err
	}

	// Determine if this is a new format (post-AHS) or old format
	// New format has 64-bit flags, old format has 32-bit flags
	// The easiest way to distinguish is by file size
	isNewFormat := true // Assume new format

	// Land tiles are separated into blocks of 32 entries, each with a 4-byte header
	landBlockCount := 0x4000 / 32 // 512 blocks of land tiles
	currentPos := 0

	// Process land tile blocks
	for block := 0; block < landBlockCount; block++ {
		// Skip the 4-byte header for this block
		currentPos += 4

		// Read 32 land tiles in this block
		for i := 0; i < 32; i++ {
			tileID := (block * 32) + i

			// Read flags (4 or 8 bytes depending on format)
			var flagsSize int
			if isNewFormat {
				flagsSize = 8
			} else {
				flagsSize = 4
			}

			// Read textureID (2 bytes) and name (20 bytes)
			totalSize := flagsSize + 2 + 20

			// Ensure we don't read beyond the file
			if currentPos+totalSize > len(data) {
				return fmt.Errorf("unexpected end of tiledata.mul file at land tile ID %d", tileID)
			}

			// Copy the data for this land tile
			entryData := make([]byte, totalSize)
			copy(entryData, data[currentPos:currentPos+totalSize])
			currentPos += totalSize

			// Add the land tile entry
			add(uint32(landOffset+tileID), uint32(tileID), uint32(len(entryData)), 0, entryData)
		}
	}

	// Calculate how many static tile blocks we have based on remaining file size
	// Each static tile entry is larger than land tiles
	staticEntrySize := 0
	if isNewFormat {
		staticEntrySize = 8 + 1 + 1 + 2 + 1 + 1 + 2 + 1 + 1 + 1 + 1 + 1 + 20
	} else {
		staticEntrySize = 4 + 1 + 1 + 2 + 1 + 1 + 2 + 1 + 1 + 1 + 1 + 1 + 20
	}

	// Process static tiles - each block has a 4-byte header followed by 32 entries
	// We'll use a sequential index for static tiles, starting at 0
	staticIndex := uint32(0)

	// Continue reading while we have enough data for at least a header
	for currentPos+4 <= len(data) {
		// Skip the 4-byte header for this block
		currentPos += 4

		// Read up to 32 static tiles in this block, or until EOF
		for i := 0; i < 32 && currentPos+staticEntrySize <= len(data); i++ {
			// Copy the data for this static tile
			entryData := make([]byte, staticEntrySize)
			copy(entryData, data[currentPos:currentPos+staticEntrySize])
			currentPos += staticEntrySize

			// Add the static tile entry using its sequential index.
			// The actual tile ID (0x4000 + index) is stored within the entry data itself or can be derived.
			add(staticIndex, 0x4000+staticIndex, uint32(len(entryData)), 0, entryData)

			staticIndex++
		}
	}

	return nil
}

// HeightTable returns an array containing the heights of all static tiles
func (s *SDK) HeightTable() ([]int, error) {
	count := s.staticTileCount()
	heights := make([]int, count)

	// Populate the height table by iterating through all static tiles
	for data := range s.StaticTiles() {
		if data.Animation < int16(count) && data.Animation >= 0 {
			heights[data.Animation] = int(data.Height)
		}
	}

	return heights, nil
}
