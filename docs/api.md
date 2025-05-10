# Go Ultima SDK API Reference

## Overview

This document describes the public API for the Go Ultima SDK. The SDK provides access to various Ultima Online data files (e.g., `.mul`, `.uop`) in a manner idiomatic to Go development. The primary goal is to offer a clean, easy-to-use interface for reading UO game data, while the internal implementation will closely mirror the reference C# SDK for correctness.

## SDK Lifecycle and Initialization

This section describes how to initialize, manage, and close an instance of the Ultima SDK.

### `type SDK struct{}`

The `SDK` struct is an opaque type that holds the necessary state and file references for accessing UO data. It is not meant to be instantiated directly; use the `Open` function instead.

```go
// SDK manages access to Ultima Online game files.
// Its fields are unexported; interaction is through its methods.
type SDK struct {
    // internal state, including uoPath and lazily opened file handles
}
```

### `func Open(uoPath string) (*SDK, error)`

Initializes and returns a new `SDK` instance. It requires the path to the root directory of an Ultima Online installation (where files like `map0.mul`, `artidx.mul`, etc., are located).

This function will:

1. Validate the provided `uoPath`.
2. Locate and parse essential index files (e.g., `artidx.mul`, `gumpidx.mul`). These are generally small and necessary to know what data exists and where to find it.
3. Prepare the SDK for data retrieval.

Actual UO data files (e.g., `art.mul`, `gumpart.mul`) are **opened lazily** on the first access to the corresponding data type (e.g., when `SDK.ArtTile()` or `SDK.Gump()` is first called). This optimizes resource usage and startup time.

```go
// Open initializes the Ultima SDK with the game client's directory path.
// It loads essential index files and prepares for data access.
// All file operations are rooted within the provided uoPath.
// Data files are opened lazily on first access.
func Open(uoPath string) (*SDK, error)
```

### `func (s *SDK) Close() error`

Releases any resources held by the `SDK` instance, such as open file handles that were lazily opened. It's important to call `Close` when the SDK is no longer needed to prevent resource leaks.

```go
// Close releases any resources held by the SDK, such as open file handles.
func (s *SDK) Close() error
```

## Core Data Structures

This section details the primary data structures used throughout the SDK to represent various game elements. Many of these structures feature lazy-loading for their graphical or audio data to improve performance.

### `type Hue struct { ... }`

Represents a color palette definition from `hues.mul`, used for re-coloring game assets.

```go
// Hue defines a color palette used for re-coloring game assets.
type Hue struct {
    Index      int
    Name       string
    Colors     [32]uint16 // Raw 16-bit color values (e.g., ARGB1555 or similar format used by UO)
    TableStart uint16
    TableEnd   uint16
}

// GetColor returns a standard Go color.Color for a specific entry in the Hue's palette.
// The implementation will convert the UO's 16-bit color to a standard Go color type.
func (h *Hue) GetColor(paletteIndex int) (color.Color, error)

// Image generates a small image.Image representing this hue's palette for visualization.
func (h *Hue) Image(widthPerColor, height int) image.Image
```

### `type VerdataPatch struct { ... }`

Represents a patch entry from `verdata.mul`, used to override standard game file data.

```go
// VerdataPatch describes a single patch entry from verdata.mul.
type VerdataPatch struct {
    FileID int32 // Identifier for the .mul file being patched (e.g., map0.mul, art.mul)
    Index  int32 // Block number or entry ID within the file
    Lookup int32 // Offset in verdata.mul where the patch data begins
    Length int32 // Length of the patch data
    Extra  int32 // Extra data, often 0, but can be used for specific patch types
}
```

### `type AnimationFrame struct { ... }`

Represents a single frame within an animation sequence. Its graphical image is loaded lazily.

```go
// AnimationFrame represents a single frame of an animation.
// The Image is loaded lazily.
type AnimationFrame struct {
    Center image.Point // Center offset for rendering the frame relative to a central point.
    // internal fields for lazy loading image
}

// Image retrieves and decodes the animation frame's graphical representation.
// The image is loaded on the first call and cached for subsequent calls.
func (af *AnimationFrame) Image() (image.Image, error)
```

### `type Animation struct { ... }`

Represents a complete animation sequence for a character or creature.

```go
// Animation represents a full animation sequence (e.g., walk, attack)
// for a specific entity, action, and direction.
type Animation struct {
    // frames are not directly exposed; use Frames().
    palette color.Palette  // The color palette used for this animation's frames.
                           // This is specific to the animation data itself.
}

// Frames returns an iterator over the animation frames. Each AnimationFrame loads its image lazily.
func (a *Animation) Frames() iter.Seq[AnimationFrame]
```

### `type ArtTile struct { ... }`

Represents a static art tile (e.g., items, terrain features). Its graphical image is loaded lazily.

```go
// ArtTile represents a piece of static art (item or terrain).
// Information may be combined from art.mul/artidx.mul and tiledata.mul.
// The image is loaded lazily.
type ArtTile struct {
    ID     int
    Name   string        // From TileData
    Flags  uint64        // From TileData (e.g., Impassable, Surface, etc.)
    Height int8          // Height of the tile, from TileData.
    // internal fields for lazy loading image
}

// Image retrieves and decodes the art tile's graphical representation.
// The image is loaded on the first call and cached for subsequent calls.
func (at *ArtTile) Image() (image.Image, error)
```

### `type LandTileData struct { ... }`

Represents the detailed properties of a land tile as defined in `tiledata.mul`.

```go
// LandTileData contains detailed properties for a land tile.
type LandTileData struct {
    ID    int
    Name  string // Typically 20 characters max
    Flags uint64 // Tile flags (e.g., Impassable, Wet, Surface, etc.)
    // TextureID int16 // If applicable, though textures are usually separate
    // ... other fields from land tile data structure in tiledata.mul
}
```

### `type StaticItemData struct { ... }`

Represents the detailed properties of a static item tile as defined in `tiledata.mul`.

```go
// StaticItemData contains detailed properties for a static item tile.
type StaticItemData struct {
    ID          int
    Name        string // Typically 20 characters max
    Flags       uint64 // Tile flags (e.g., Impassable, Wearable, LightSource, etc.)
    Weight      uint8
    Quality     uint8
    Quantity    uint8
    Hue         uint8  // Default hue, if any
    StackingOffset uint8 // Value for stacking items
    AnimationID uint16 // Body ID for animations or equipable items
    Height      uint8  // Tile height/depth
    LightID     uint8  // Light source ID if it's a light source
    SubType     uint8  // e.g., for armor/weapon types
    // ... other fields from item tile data structure in tiledata.mul
}
```

### `type GumpInfo struct { ... }`

Represents metadata for a gump graphic, such as its dimensions.

```go
// GumpInfo contains metadata about a gump, like its ID, width, and height.
type GumpInfo struct {
    ID     int
    Width  int
    Height int
}
```

### `type Gump struct { ... }`

Represents a single gump graphic (UI element). Its graphical image is loaded lazily.

```go
// Gump represents a UI element or graphic.
// The Image is loaded lazily.
type Gump struct {
    GumpInfo
    // internal fields for lazy loading image
}

// Image retrieves and decodes the gump's graphical representation.
// The image is loaded on the first call and cached for subsequent calls.
func (g *Gump) Image() (image.Image, error)
```

### `type MultiItem struct { ... }`

Represents a single component within a multi-item structure (e.g., a wall in a house).

```go
// MultiItem defines a single item within a multi-structure.
type MultiItem struct {
    ItemID    uint16 // Tile ID of the item.
    OffsetX   int16  // Relative X offset.
    OffsetY   int16  // Relative Y offset.
    OffsetZ   int8   // Relative Z offset.
    Flags     uint64 // Flags associated with this item (e.g., 0x1 for "visible").
    ClilocID  uint32 // Optional Cliloc ID for item description (if applicable, often 0).
}
```

### `type Multi struct { ... }`

Represents a complete multi-item structure, like a house or a boat.

```go
// Multi represents a multi-item structure like a house or boat.
type Multi struct {
    ID          int
    // items are not directly exposed; use Items().
    // width, height, minZ, maxZ are not stored directly; use Dimensions().
}

// Items returns an iterator over all items in this multi-structure.
func (m *Multi) Items() iter.Seq[MultiItem]

// Dimensions returns the calculated width, height, and z-range of the multi.
// These are typically calculated based on the extents of its items.
func (m *Multi) Dimensions() (width, height int, minZ, maxZ int8, err error)
```

### `type Skill struct { ... }`

Represents a single character skill.

```go
// Skill defines a character skill.
type Skill struct {
    ID        int
    IsAction  bool   // True if the skill is an action, false if a passive skill.
    Name      string // Name of the skill.
    // GroupID is typically resolved by iterating through SkillGroups and their contained skills.
}
```

### `type SkillGroup struct { ... }`

Represents a group of related skills (e.g., "Combat Skills").

```go
// SkillGroup defines a category of skills.
type SkillGroup struct {
    ID   int
    Name string
    // internal list of skill IDs or direct skill references for this group
}

// Skills returns an iterator over the skills belonging to this skill group.
// Requires an SDK instance to resolve skill details from their IDs.
func (sg *SkillGroup) Skills(sdk *SDK) iter.Seq[*Skill]
```

### `type Sound struct { ... }`

Represents a single sound effect. Its audio data is loaded lazily.

```go
// Sound represents a sound effect.
// The audio Data is loaded lazily.
type Sound struct {
    ID   int
    Name string // Name from sound.def, if available.
    // internal fields for lazy loading data
}

// Data retrieves and returns the raw sound data (typically WAV format, but could be just PCM).
// The data is loaded on the first call and cached for subsequent calls.
func (s *Sound) Data() ([]byte, error)
```

### `type FontCharacterInfo struct { ... }`

Represents information about a single character in a font. Its graphical image is loaded lazily.

```go
// FontCharacterInfo holds the metrics for a single character and allows lazy loading of its bitmap.
type FontCharacterInfo struct {
    Width  int         // Width of the character bitmap.
    Height int         // Height of the character bitmap.
    // Additional metrics like advance width, offsets might be needed depending on font type.
    // internal fields for lazy loading image
}

// Image retrieves and decodes the character's graphical representation.
// The image is loaded on the first call and cached for subsequent calls.
func (fci *FontCharacterInfo) Image() (image.Image, error)
```

### `type Font interface { ... }`

Defines an interface for interacting with a specific game font (ASCII or Unicode).

```go
// Font defines an interface for interacting with a specific game font.
type Font interface {
    // GetCharacter retrieves the graphical information for a given character rune.
    // The returned FontCharacterInfo loads its image lazily.
    GetCharacter(char rune) (*FontCharacterInfo, error)

    // LineHeight returns the default height for a line of text with this font.
    LineHeight() int

    // Name returns the identifier or name of the font.
    Name() string
}
```

### `type Light struct { ... }`

Represents a light source definition from `light.mul`.

```go
// Light describes a light source, including its ID and potentially its graphic/radius.
// The exact contents of light.mul entries need to be mapped here.
type Light struct {
    ID     int
    Width  int    // Width of the light effect (if applicable, from light.mul)
    Height int    // Height of the light effect (if applicable, from light.mul)
    Data   []byte // Raw light data, interpretation depends on client usage.
    // Image image.Image // Optional: if lights are directly renderable as images.
}
```

### `type Speech struct { ... }`

Represents a single predefined speech entry from `speech.mul`.

```go
// Speech holds the ID and text for a predefined speech entry.
type Speech struct {
    ID   int
    Text string
}
```

### `type Texture struct { ... }`

Represents a land texture graphic from `texmaps.mul`. Its graphical image is loaded lazily.

```go
// Texture represents a land texture graphic.
// The Image is loaded lazily.
type Texture struct {
    ID     int
    // internal fields for lazy loading image
}

// Image retrieves and decodes the texture's graphical representation.
// The image is loaded on the first call and cached for subsequent calls.
func (t *Texture) Image() (image.Image, error)
```

### Map-Related Data Structures

These structures are specifically related to the game world map.

#### `type LandTile struct { ... }`

Represents the base land part of a map tile. Its graphical image is loaded lazily.

```go
// LandTile represents the base land part of a map tile.
// The image is loaded lazily.
type LandTile struct {
    ID     int         // Tile ID for the land.
    Z      int8        // Z-coordinate (height).
    // internal fields for lazy loading image
}

// Image retrieves and decodes the land tile's graphical representation.
// The image is loaded on the first call and cached for subsequent calls.
func (lt *LandTile) Image() (image.Image, error)
```

#### `type StaticTile struct { ... }`

Represents a static item present on a map tile. Its graphical image is loaded lazily.

```go
// StaticTile represents a static item present on a map tile.
// The image is loaded lazily.
type StaticTile struct {
    ID     int         // Tile ID for the static item.
    X, Y   int16       // Coordinates (redundant if part of MapTileInfo for x,y)
    Z      int8        // Z-coordinate (height).
    Hue    *Hue        // Optional hue applied to this static item.
    // internal fields for lazy loading image
}

// Image retrieves and decodes the static tile's graphical representation.
// The image is loaded on the first call and cached for subsequent calls.
func (st *StaticTile) Image() (image.Image, error)
```

#### `type MapTileInfo struct { ... }`

Contains all land and static items at a specific map coordinate.

```go
// MapTileInfo contains all land and static items at a specific map coordinate.
// It is retrieved via methods on a GameMap object.
type MapTileInfo struct {
    LandTile // Embedded LandTile
    // staticTiles are not directly exposed; use StaticTiles().
}

// StaticTiles returns an iterator over the static tiles at this map location.
// Each StaticTile loads its image lazily.
func (mti *MapTileInfo) StaticTiles() iter.Seq[StaticTile]
```

#### `type GameMap struct { ... }`

Represents a specific game world map (e.g., Felucca, Trammel) and provides methods to access its tile data.

```go
// GameMap provides access to tile data for a specific map (e.g., Felucca, Trammel).
// Its fields are unexported; interaction is through its methods.
// An instance is obtained by calling SDK.Map().
type GameMap struct {
    // internal state, e.g., mapID, file references
}

// Tile retrieves all tile information (land and statics) for a specific coordinate on this map.
// x and y are the tile coordinates.
// This method handles reading from the appropriate map, statics, and art files,
// applying hues, and resolving verdata patches.
// The images for LandTile and StaticTile components are loaded lazily.
func (gm *GameMap) Tile(x, y int) (*MapTileInfo, error)

// Dimensions returns the width and height of this map in tiles.
func (gm *GameMap) Dimensions() (width, height int, err error)

// Tiles returns an iterator over tiles within the given rectangle on this map.
// The iterator yields an image.Point (for x, y coordinates) and the *MapTileInfo.
func (gm *GameMap) Tiles(bounds image.Rectangle) iter.Seq2[image.Point, *MapTileInfo]
```

## SDK Data Accessor Methods

This section details the methods available on an initialized `*SDK` instance, grouped by the type of game data they provide access to.

### Hues

Methods for accessing hue data from `hues.mul`.

```go
// HueAt retrieves a specific hue by its index (typically 0-2999).
func (s *SDK) HueAt(index int) (*Hue, error)

// Hues returns an iterator over all available hues defined in hues.mul.
func (s *SDK) Hues() iter.Seq[*Hue]
```

### Animations

Methods for accessing character, creature, and effect animations from `anim.idx`/`mul`, `anim2.idx`/`mul`, etc.

```go
// Animation retrieves a decoded animation sequence.
// fileType indicates which animation set to use (e.g., 1 for anim.mul, 2 for anim2.mul).
// body is the creature or character ID.
// action defines the animation type (e.g., walk, attack, cast spell).
// direction indicates the facing of the animation (0-7, where 0 is typically North).
// The returned Animation object allows for lazy loading of its frames' images.
func (s *SDK) Animation(fileType, body, action, direction int) (*Animation, error)

// ActionDefined checks if a specific animation (body, action) is defined in the specified fileType.
// This is useful for determining if an animation exists before attempting to load it.
func (s *SDK) ActionDefined(fileType, body, action int) (bool, error)
```

### Art and TileData

Methods for accessing static art tiles (from `artidx.mul`, `art.mul`) and detailed tile properties (from `tiledata.mul`).

```go
// ArtTile retrieves static art data by its tile ID.
// This method will internally handle UOP file formats if present and configured.
// It also incorporates relevant details from tiledata.mul.
// The returned ArtTile object allows for lazy loading of its image.
func (s *SDK) ArtTile(tileID int) (*ArtTile, error)

// LandTileData retrieves the detailed properties for a specific land tile ID from tiledata.mul.
func (s *SDK) LandTileData(id int) (*LandTileData, error)

// AllLandTileData returns an iterator over all land tile data entries from tiledata.mul.
func (s *SDK) AllLandTileData() iter.Seq[*LandTileData]

// StaticItemData retrieves the detailed properties for a specific static item tile ID from tiledata.mul.
func (s *SDK) StaticItemData(id int) (*StaticItemData, error)

// AllStaticItemData returns an iterator over all static item data entries from tiledata.mul.
func (s *SDK) AllStaticItemData() iter.Seq[*StaticItemData]
```

### Maps

Methods for accessing map data from `mapX.mul`, `staidxX.mul`, `staticsX.mul`, and related files.

```go
// Map retrieves a handler for a specific game map (e.g., map 0 for Felucca, 1 for Trammel).
// The returned GameMap object can then be used to query tile data for that map.
func (s *SDK) Map(mapID int) (*GameMap, error)
```

### Gumps

Methods for accessing gump graphics (UI elements) from `gumpidx.mul`/`gumpart.mul` or UOP equivalent.

```go
// Gump retrieves a specific gump graphic by its ID.
// It handles reading from .mul or UOP files.
// The returned Gump object allows for lazy loading of its image.
func (s *SDK) Gump(id int) (*Gump, error)

// GumpInfos returns an iterator over metadata (ID, width, height) for all available gumps.
// This is efficient for listing gumps without loading all their pixel data.
func (s *SDK) GumpInfos() iter.Seq[*GumpInfo]
```

### Multis

Methods for accessing multi-item structures (e.g., houses, boats) from `multi.idx`/`multi.mul`.

```go
// Multi retrieves a specific multi-structure by its ID.
func (s *SDK) Multi(id int) (*Multi, error)

// MultiIDs returns an iterator over all available multi IDs.
func (s *SDK) MultiIDs() iter.Seq[int]
```

### Radar Colors

Methods for accessing radar color mappings from `radarcol.mul`.

```go
// RadarColor retrieves the radar color for a given static tile ID.
// The returned color is typically a 16-bit value (e.g., ARGB1555) specific to UO's radar.
func (s *SDK) RadarColor(staticTileID int) (uint16, error)

// RadarColors returns an iterator over all defined radar color mappings.
// It yields the static tile ID and its corresponding 16-bit radar color.
func (s *SDK) RadarColors() iter.Seq2[int, uint16]
```

### Skills and Skill Groups

Methods for accessing skill definitions from `skills.idx`/`skills.mul` and `skillgrp.mul`.

```go
// Skill retrieves a specific skill by its ID.
func (s *SDK) Skill(id int) (*Skill, error)

// Skills returns an iterator over all defined skills.
// Note: This iterates over individual skills; for grouped skills, see SkillGroups() and SkillGroup.Skills().
func (s *SDK) Skills() iter.Seq[*Skill]

// SkillGroup retrieves a specific skill group by its ID.
func (s *SDK) SkillGroup(id int) (*SkillGroup, error)

// SkillGroups returns an iterator over all defined skill groups.
func (s *SDK) SkillGroups() iter.Seq[*SkillGroup]
```

### Sound

Methods for accessing sound effects from `sound.def`, `soundidx.mul`, and `sound.mul`.

```go
// Sound retrieves a specific sound effect by its ID.
// The returned Sound object allows for lazy loading of its audio data.
func (s *SDK) Sound(id int) (*Sound, error)

// Sounds returns an iterator over all available sound effects (ID and Name).
// The audio Data for each sound is loaded lazily when its Data() method is called.
func (s *SDK) Sounds() iter.Seq[*Sound]
```

### Client Localized Strings (Cliloc)

Methods for accessing localized strings from `cliloc.*` files (e.g., `cliloc.enu`). The SDK should handle loading the appropriate language file based on configuration or auto-detection.

```go
// String retrieves a localized string by its ID.
// The SDK is responsible for loading and querying the correct language cliloc file.
func (s *SDK) String(id int) (string, error)

// Strings returns an iterator over all loaded localized strings from the active language file.
// It yields the ID and the corresponding string.
// Note: This could be a very large sequence.
func (s *SDK) Strings() iter.Seq2[int, string]
```

### Fonts

Methods for accessing font data from `fonts.mul` (ASCII) and `unicode.flt` (Unicode), along with their definition files.

```go
// Font retrieves the specified ASCII font (0-9) as a Font interface.
// Note: UO typically has multiple ASCII fonts.
func (s *SDK) Font(fontID int) (Font, error)

// FontUnicode retrieves the Unicode font as a Font interface.
func (s *SDK) FontUnicode() (Font, error)

// TODO: Consider if a generic Font(nameOrID string) (Font, error) is better if font types are numerous or configurable.
```

### Light Effects

Methods for accessing light source definitions from `lightidx.mul` and `light.mul`.

```go
// Light retrieves a specific light definition by its ID.
func (s *SDK) Light(id int) (*Light, error)

// Lights returns an iterator over all defined light definitions.
func (s *SDK) Lights() iter.Seq[*Light]
```

### Speech Entries

Methods for accessing predefined speech entries from `speech.mul`. These are typically keyword-triggered responses or system messages.

```go
// SpeechEntry retrieves a predefined speech entry by its ID.
func (s *SDK) SpeechEntry(id int) (*Speech, error)

// SpeechEntries returns an iterator over all defined speech entries.
func (s *SDK) SpeechEntries() iter.Seq[*Speech]
```

### Textures

Methods for accessing land textures from `texidx.mul` and `texmaps.mul`.

```go
// Texture retrieves a specific land texture graphic by its ID.
// The returned Texture object allows for lazy loading of its image.
func (s *SDK) Texture(id int) (*Texture, error)

// Textures returns an iterator over all available textures.
// The Image for each texture is loaded lazily when its Image() method is called.
func (s *SDK) Textures() iter.Seq[*Texture]
```

### Verdata (Patches)

Methods for accessing patch information from `verdata.mul`.

```go
// VerdataPatches returns an iterator over all verdata patch entries.
// These patches indicate modifications to other .mul files.
// Note: Applying patches will typically be handled internally by other methods
// (e.g., ArtTile, MapTile.Tile) rather than requiring manual patch application by the user.
func (s *SDK) VerdataPatches() iter.Seq[VerdataPatch]
```

## Image Handling

This section outlines general conventions for how image data is handled by the SDK.

- Game assets, such as art tiles, gumps, and animation frames, will be returned as standard Go `image.Image` interface types where an image is directly part of a struct, or via an `Image()` method that returns `(image.Image, error)` for lazy-loaded images.
- The SDK will handle the conversion from UO's native pixel formats (often 16-bit ARGB1555 or similar) to a common Go image type (e.g., `image.NRGBA` or `image.RGBA`).
- Palettes associated with specific assets (like animations or some gumps) will be provided as `color.Palette`.
- Hues, when applied, will modify the colors of the resulting `image.Image`.

## Error Handling

This section describes the error handling strategy for the SDK.

- All functions and methods that can encounter issues (e.g., file not found, invalid data, index out of bounds) will return an `error` as their last return value.
- Callers should always check the returned error. A non-nil error indicates failure.
- Error messages will be descriptive to aid in debugging. Error wrapping (using `fmt.Errorf` with `%w`) will be used where appropriate to provide context while preserving underlying error types.
- The SDK will not use `panic` for recoverable errors.
