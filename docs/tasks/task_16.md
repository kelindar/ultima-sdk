# Task 16: Port Art.cs to art.go

## Objective
Implement support for loading and accessing static art tiles and land tiles from `artidx.mul` and `art.mul` (or their UOP equivalents). These files contain the graphical representation of all land and static item tiles used in Ultima Online.

## C# Reference Implementation Analysis
The C# implementation in `Art.cs` handles loading art data from the MUL files or UOP alternatives. It includes:
- Methods to load and decode both land tiles and static items
- Support for both MUL and UOP file formats
- Caching mechanisms to improve performance
- Handling of removed/invalid tiles
- Integration with TileData for determining item properties

The art system is fundamental to UO, as it provides the visual representation of all map elements and items.

## Work Items
1. Create a new file `art.go` in the root package.

2. Define the `ArtTile` struct that will represent a single art tile:
   ```go
   type ArtTile struct {
       ID      int
       Name    string       // From TileData
       Flags   uint64       // From TileData
       Height  int8         // From TileData

       // Fields to support lazy loading
       width   int
       height  int
       loaded  bool
       rawData []byte
       image   image.Image
       mutex   sync.RWMutex
   }

   // Image retrieves and decodes the art tile's graphical representation
   func (at *ArtTile) Image() (image.Image, error) {
       // Implementation for lazy-loading and returning the image
   }
   ```

3. Add methods to the SDK struct for accessing art tiles:
   ```go
   // ArtTile retrieves static art data by its tile ID
   func (s *SDK) ArtTile(tileID int) (*ArtTile, error) {
       // Implementation for retrieving a specific art tile
   }
   ```

4. Implement internal utility functions:
   ```go
   // Internal function for loading and decoding art data
   func (s *SDK) loadArtData(id int, isStatic bool) ([]byte, int, int, error) {
       // Implementation for loading raw art data and dimensions
   }

   // Utility function for normalizing item IDs
   func getLegalItemID(id int) int {
       // Implementation based on C# GetLegalItemID
   }
   ```

5. Write comprehensive unit tests in `art_test.go`:
   - Test loading land tiles
   - Test loading static items
   - Test handling of invalid/removed tiles
   - Test image dimensions and pixel data
   - Test lazy loading behavior
   - Test UOP file support (if applicable)

## Key Considerations
- Art data is stored in a specialized format that requires careful decoding
- Land tiles and static items have different storage formats and decoding methods
- The implementation should support both MUL and UOP file formats seamlessly
- Art tiles should integrate with TileData for property information
- Consider implementing lazy loading of image data to improve memory usage
- Implement proper caching to avoid repeated decoding of frequently used tiles
- Handle invalid or removed tiles gracefully
- Ensure thread safety for concurrent access to art data
- Use the bitmap package (implemented in Task 12) for image handling

## Expected Output
A complete implementation that allows:
- Loading art data from both MUL and UOP files
- Retrieving individual art tiles by ID
- Decoding tiles into standard Go image.Image instances
- Accessing tile properties via TileData integration
- Efficient memory usage through lazy loading and caching

## Verification
- Compare decoded images with the C# implementation to ensure visual accuracy
- Test with known tile IDs to verify dimensions and visual appearance
- Verify that invalid/removed tiles are handled correctly
- Benchmark performance, especially for frequently accessed tiles