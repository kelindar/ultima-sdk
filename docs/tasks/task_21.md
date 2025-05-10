# Task 21: Port Map.cs & MapHelper.cs to map.go

## Objective
Implement support for loading and accessing map data from Ultima Online map files (`map*.mul`, `staidx*.mul`, `statics*.mul`) and their UOP equivalents. This component is one of the most complex in the SDK as it handles the core world data including terrain, static items, and their properties.

## C# Reference Implementation Analysis
The C# implementation spans multiple files:
- `Map.cs` - The main class for loading and accessing map data
- `TileMatrix.cs` - Handles the actual loading and storage of map blocks
- `TileMatrixPatch.cs` - Applies patch data to maps
- `Helpers/MapHelper.cs` - Contains helper functions for map operations

The implementation must handle:
- Multiple map files for different facets (Felucca, Trammel, etc.)
- Static items overlaid on terrain
- Map patching systems
- Efficient lookup of map data by coordinates

## Work Items
1. Create a new file `map.go` in the root package.

2. Define the core map structures:
   ```go
   // LandTile represents a terrain tile
   type LandTile struct {
       ID    int
       Z     int8
       // Additional fields from TileData
   }

   // StaticTile represents a static item on the map
   type StaticTile struct {
       ID     int
       X      int
       Y      int
       Z      int8
       Hue    int
       // Additional fields from TileData
   }

   // MapBlock represents a 8x8 block of map tiles
   type MapBlock struct {
       X       int
       Y       int
       Tiles   [64]LandTile
       Statics []StaticTile
   }

   // Map represents a complete UO map (facet)
   type Map struct {
       ID      int
       Width   int
       Height  int
       // Internal fields for data access and caching
   }
   ```

3. Add methods to the SDK struct for accessing map data:
   ```go
   // Map retrieves a specific map by its ID
   func (s *SDK) Map(id int) (*Map, error) {
       // Implementation for retrieving a specific map
   }

   // Maps returns an iterator over all available maps
   func (s *SDK) Maps() iter.Seq[*Map] {
       // Implementation for iterating over maps
   }
   ```

4. Implement methods for the Map struct:
   ```go
   // GetLandTile retrieves the land tile at the specified coordinates
   func (m *Map) GetLandTile(x, y int) (LandTile, error) {
       // Implementation for retrieving a land tile
   }

   // GetStaticTiles retrieves all static tiles at the specified coordinates
   func (m *Map) GetStaticTiles(x, y int) ([]StaticTile, error) {
       // Implementation for retrieving static tiles
   }

   // GetBlock retrieves a complete map block at the specified block coordinates
   func (m *Map) GetBlock(blockX, blockY int) (*MapBlock, error) {
       // Implementation for retrieving a map block
   }
   ```

5. Implement helper functions for map operations:
   ```go
   // WorldToBlock converts world coordinates to block coordinates
   func WorldToBlock(x, y int) (int, int) {
       // Implementation for coordinate conversion
   }

   // BlockToWorld converts block coordinates to world coordinates
   func BlockToWorld(blockX, blockY int) (int, int) {
       // Implementation for coordinate conversion
   }
   ```

6. Implement the internal loading mechanisms:
   ```go
   // Internal function to load map data
   func (s *SDK) loadMap(id int) (*Map, error) {
       // Implementation for loading map data
   }
   ```

7. Write comprehensive unit tests in `map_test.go`:
   - Test loading map data for different facets
   - Test retrieving land tiles
   - Test retrieving static tiles
   - Test coordinate conversion functions
   - Test edge cases (map boundaries, invalid coordinates)
   - Test patch application

## Key Considerations
- Maps are divided into 8x8 tile blocks for efficient storage and retrieval
- Static items are stored separately from land tiles
- Maps can be large, so efficient caching is important
- Memory usage should be optimized, possibly by using lazy loading for map blocks
- Maps may have different dimensions depending on the facet
- Coordinate systems include world coordinates and block coordinates
- The implementation should support multiple map file formats (MUL and UOP)
- Map patches from various sources need to be correctly applied
- Thread safety is important for concurrent access
- Static items at a single location should be sorted by Z-order
- Consider adding utility functions for common map operations (line of sight, distance calculation, etc.)

## Expected Output
A complete implementation that allows:
- Loading map data for different facets
- Retrieving land tiles and static items by coordinates
- Handling map blocks efficiently with proper caching
- Converting between different coordinate systems
- Accessing tile properties seamlessly through TileData integration

## Verification
- Compare loaded map data with the C# implementation
- Verify land tiles and static items match between implementations
- Test with known coordinates to ensure correct tile retrieval
- Verify patch application works correctly
- Test performance with large areas of map data