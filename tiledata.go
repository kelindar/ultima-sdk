// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"codeberg.org/go-mmap/mmap"
	"github.com/kelindar/ultima-sdk/internal/mul"
	"github.com/kelindar/ultima-sdk/internal/uofile"
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

// LandInfo represents the data for a single land tile.
type LandInfo struct {
	TextureID uint16   // Texture ID for the land tile
	Flags     TileFlag // Properties of this land tile
	Name      string   // Name of the tile
}

// ItemInfo represents the data for a single static item tile in Ultima Online.
// This structure contains all the properties that define how an item behaves
// in the game world, including physical properties, rendering information,
// and game mechanics data.
//
// Context-Sensitive Fields:
// The Quality and Quantity fields contain different data depending on the item's flags.
// Use the typed helper methods for safe access:
//   - IsWearable() (layer byte, ok bool) - equipment layer for wearable items
//   - IsWeapon() (weaponClass byte, ok bool) - weapon class for weapons
//   - IsArmor() (armorClass byte, ok bool) - armor class for armor
//   - IsLightSource() (lightID byte, ok bool) - light pattern ID for light sources
type ItemInfo struct {
	// Name is the display name of the item as it appears in the game.
	Name string

	// Flags contains bitwise flags that define the item's properties and behaviors.
	// Use the helper methods (IsWearable, IsWeapon, etc.) for easier access.
	Flags TileFlag

	// Physical Properties
	Weight byte // Item weight in game units (255 = immovable)
	Height byte // Physical height for collision detection and line-of-sight
	Value  byte // Item's worth for vendor pricing and insurance

	// Visual Properties
	AnimationID    int16 // Animation/Body ID (0 = no animation)
	Hue            byte  // Color/tint value for visual effects
	StackingOffset byte  // Visual offset when stacked (requires Generic flag)

	// Context-Sensitive Properties
	// These fields contain different data depending on the item's flags:
	Quality  byte  // Layer (wearable), LightID (light source), or other quality data
	Quantity byte  // WeaponClass (weapon), ArmorClass (armor), stack quantity, etc.
	MiscData int16 // Legacy data for old weapon templates
}

// Background returns whether the item has the Background flag set
func (i ItemInfo) Background() bool {
	return i.Flags&TileFlagBackground != 0
}

// Bridge returns whether the item has the Bridge flag set
func (i ItemInfo) Bridge() bool {
	return i.Flags&TileFlagBridge != 0
}

// Impassable returns whether the item has the Impassable flag set
func (i ItemInfo) Impassable() bool {
	return i.Flags&TileFlagImpassable != 0
}

// Surface returns whether the item has the Surface flag set
func (i ItemInfo) Surface() bool {
	return i.Flags&TileFlagSurface != 0
}

// IsWearable returns whether the item can be worn/equipped, and if so, the equipment layer
func (i ItemInfo) IsWearable() (layer byte, ok bool) {
	if i.Flags&TileFlagWearable != 0 {
		return i.Quality, true
	}
	return 0, false
}

// IsWeapon returns whether the item is a weapon, and if so, the weapon class
func (i ItemInfo) IsWeapon() (weaponClass byte, ok bool) {
	if i.Flags&TileFlagWeapon != 0 {
		return i.Quantity, true
	}
	return 0, false
}

// IsArmor returns whether the item is armor, and if so, the armor class
func (i ItemInfo) IsArmor() (armorClass byte, ok bool) {
	if i.Flags&TileFlagArmor != 0 {
		return i.Quantity, true
	}
	return 0, false
}

// IsLightSource returns whether the item produces light, and if so, the light pattern ID
func (i ItemInfo) IsLightSource() (lightID byte, ok bool) {
	if i.Flags&TileFlagLightSource != 0 {
		return i.Quality, true
	}
	return 0, false
}

// IsContainer returns whether the item can hold other items
func (i ItemInfo) IsContainer() bool {
	return i.Flags&TileFlagContainer != 0
}

// StackQuantity returns the default stack quantity for stackable items.
// For non-stackable items, this returns the context-specific Quantity value.
func (i ItemInfo) StackQuantity() byte {
	return i.Quantity
}

// Layer returns the equipment layer for wearable items, or 0 if not wearable.
// For better type safety, prefer using IsWearable() which returns (layer, bool).
func (i ItemInfo) Layer() byte {
	if i.Flags&TileFlagWearable != 0 {
		return i.Quality
	}
	return 0
}

// LightID returns the light pattern ID for light sources, or 0 if not a light source.
// For better type safety, prefer using IsLightSource() which returns (lightID, bool).
func (i ItemInfo) LightID() byte {
	if i.Flags&TileFlagLightSource != 0 {
		return i.Quality
	}
	return 0
}

// IsMoveable returns whether the item can be moved/picked up.
// Items with Weight=255 are considered immovable.
func (i ItemInfo) IsMoveable() bool {
	return i.Weight != 255
}

// HasAnimation returns whether this item has an associated animation.
func (i ItemInfo) HasAnimation() bool {
	return i.AnimationID != 0
}

// CalcHeight returns the calculated height of the item. For bridges, this is Height/2.
func (i ItemInfo) CalcHeight() int {
	if i.Flags&TileFlagBridge != 0 {
		return int(i.Height) / 2
	}
	return int(i.Height)
}

// readStringFromBytes reads a null-terminated string from a fixed-length byte array
func readStringFromBytes(b []byte) string {
	n := bytes.IndexByte(b, 0)
	if n == -1 {
		n = len(b)
	}
	return string(b[:n])
}

// landInfo returns a specific land tile's data by ID
func (s *SDK) landInfo(id int) (*LandInfo, error) {
	if id < 0 || id >= 0x4000 {
		return nil, fmt.Errorf("invalid land tile ID: %d", id)
	}

	file, err := s.loadTiledata()
	if err != nil {
		return nil, err
	}

	return uofile.Decode(file, uint32(landOffset+id), decodeLandInfo)
}

// staticInfo returns a specific static tile's data by ID
func (s *SDK) staticInfo(id int) (*ItemInfo, error) {
	if id < 0 || id >= s.staticTileCount() {
		return nil, fmt.Errorf("invalid static tile ID: %d", id)
	}

	file, err := s.loadTiledata()
	if err != nil {
		return nil, err
	}

	return uofile.Decode(file, uint32(id), decodeStaticInfo)
}

// staticTileCount returns the number of static tiles in the tiledata file
// This is determined by the file size and format
func (s *SDK) staticTileCount() int {
	// Most clients have 0x10000 static tiles (65536)
	return 0x10000
}

// decodeTileDataFile loads the tiledata.mul file and populates the internal
// data structures for land and static tiles
func decodeTileDataFile(file *mmap.File, add mul.AddFn) error {
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
			const flagsSize = 8

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
	staticEntrySize = 8 + 1 + 1 + 2 + 1 + 1 + 2 + 1 + 1 + 1 + 1 + 1 + 20

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

func decodeLandInfo(data []byte, _ uint64) (*LandInfo, error) {
	var out LandInfo
	out.Flags = TileFlag(binary.LittleEndian.Uint64(data[0:8]))
	out.TextureID = binary.LittleEndian.Uint16(data[8:10])
	out.Name = readStringFromBytes(data[10:30])
	return &out, nil
}

func decodeStaticInfo(data []byte, _ uint64) (*ItemInfo, error) {
	var out ItemInfo

	// Static tile data format in tiledata.mul:
	// Offset | Size | Field        | Description
	// -------|------|--------------|-------------
	//   0    |  8   | flags        | TileFlag bitfield (64-bit)
	//   8    |  1   | weight       | Item weight (255 = immovable)
	//   9    |  1   | quality      | Layer for wearables, LightID for lights
	//  10    |  2   | miscData     | Miscellaneous data (old weapon templates)
	//  12    |  1   | unk2         | Unknown field (skipped)
	//  13    |  1   | quantity     | Weapon/Armor class, or stack quantity
	//  14    |  2   | animation    | Animation/Body ID (0 = no animation)
	//  16    |  1   | unk3         | Unknown field (skipped)
	//  17    |  1   | hue          | Color/hue value
	//  18    |  1   | stackOffset  | Visual stacking offset (Generic flag)
	//  19    |  1   | value        | Item value/worth
	//  20    |  1   | height       | Physical height in game units
	//  21    | 20   | name         | Null-terminated item name string

	out.Flags = TileFlag(binary.LittleEndian.Uint64(data[0:8]))
	offset := 8

	out.Weight = data[offset]
	out.Quality = data[offset+1] // Context-sensitive: Layer, LightID, etc.
	out.MiscData = int16(binary.LittleEndian.Uint16(data[offset+2 : offset+4]))
	//out.Unk2 = data[offset+4] // Unknown field, skipped
	out.Quantity = data[offset+5] // Context-sensitive: WeaponClass, ArmorClass, etc.
	out.AnimationID = int16(binary.LittleEndian.Uint16(data[offset+6 : offset+8]))
	//out.Unk3 = data[offset+8] // Unknown field, skipped
	out.Hue = data[offset+9]
	out.StackingOffset = data[offset+10]
	out.Value = data[offset+11]
	out.Height = data[offset+12]
	out.Name = readStringFromBytes(data[offset+13 : offset+33])
	return &out, nil
}
