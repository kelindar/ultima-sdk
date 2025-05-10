# Task 13: Port TileData.cs & TileDataHelpers.cs to tiledata.go

## Objective

Implement support for loading and accessing tile data from the `tiledata.mul` file, which contains essential information about land and static item tiles in the game, including their properties, names, and flags.

## C# Reference Implementation Analysis

The C# implementation consists of:

- `TileData.cs` - Main class that loads and provides access to land and item tile data
- `TileDataHelpers.cs` - Contains helper methods for working with tile flags and properties

The `tiledata.mul` file contains two distinct sections:

1. Land tile data - Describes terrain tiles (texture IDs, flags, names)
2. Static item data - Describes static objects (flags, weight, quality, name, etc.)

The implementation must handle both old and new tiledata formats, as the structure changed between UO clients.

## Work Items

1. Create a new file `tiledata.go` in the root package.

2. Define the `LandTileData` struct for land tiles:

   ```go
   type LandTileData struct {
       ID    int
       Flags uint64        // Tile flags (Wet, Impassable, etc.)
       Name  string        // Name of the tile
       // Additional fields as needed
   }
   ```

3. Define the `StaticItemData` struct for static item tiles:

   ```go
   type StaticItemData struct {
       ID          int
       Flags       uint64  // Tile flags
       Weight      uint8
       Quality     uint8
       Quantity    uint8
       Hue         uint8
       StackingOffset uint8
       AnimationID uint16  // For animation linking
       Height      uint8   // Tile height/depth
       Name        string  // Name of the item
       // Additional fields as needed
   }
   ```

4. Implement methods for the SDK struct to access tile data:

   ```go
   // LandTileData retrieves the details for a specific land tile ID
   func (s *SDK) LandTileData(id int) (*LandTileData, error) {
       // Implementation for retrieving land tile data
   }

   // AllLandTileData returns an iterator over all land tile data
   func (s *SDK) AllLandTileData() iter.Seq[*LandTileData] {
       // Implementation for iterating over land tiles
   }

   // StaticItemData retrieves the details for a specific static item tile ID
   func (s *SDK) StaticItemData(id int) (*StaticItemData, error) {
       // Implementation for retrieving static item data
   }

   // AllStaticItemData returns an iterator over all static item data
   func (s *SDK) AllStaticItemData() iter.Seq[*StaticItemData] {
       // Implementation for iterating over static items
   }
   ```

5. Implement helper functions for working with tile flags:

   ```go
   // Flag checking functions for land tiles
   func IsWet(flags uint64) bool { /* ... */ }
   func IsImpassable(flags uint64) bool { /* ... */ }
   // Additional flag helpers as needed

   // Flag checking functions for static items
   func IsWearable(flags uint64) bool { /* ... */ }
   func IsContainer(flags uint64) bool { /* ... */ }
   // Additional flag helpers as needed
   ```

6. Implement the internal loading mechanism:

   ```go
   // Internal function to load tile data
   func (s *SDK) loadTileData() error {
       // Implementation for loading land and static item data
       // Handle both old and new tiledata formats
   }
   ```

7. Write comprehensive unit tests in `tiledata_test.go`:
   - Test loading land tile data
   - Test loading static item data
   - Test accessing specific tile properties
   - Test flag checking functions
   - Test handling of different tiledata formats

## Key Considerations

- The tiledata.mul format changed between UO clients - the implementation needs to handle both old and new formats
- The old format uses 32-bit flags, while the new format uses 64-bit flags
- Character encoding is typically Windows-1252, not UTF-8
- Land and static data have different structures and are stored in separate sections
- Flag bits have specific meanings that need to be preserved in the flag helper functions
- The file is relatively large but can be loaded entirely in memory
- Names are stored as fixed-length null-padded strings

## Expected Output

A complete implementation that allows:

- Loading land tile and static item data from tiledata.mul
- Accessing individual tile entries by ID
- Checking tile flags and properties
- Iterating over all land tiles and static items
- Properly handling both old and new tiledata formats

## Verification

- Compare loaded tile names and properties with the C# implementation
- Verify flag values match between implementations
- Test with known tile IDs to ensure properties are correctly loaded
- Verify correct handling of both old and new tiledata formats
