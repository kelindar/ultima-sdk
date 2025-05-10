# Task 17: Port Gumps.cs to gump.go

## Objective

Implement support for loading and accessing gump images from `gumpidx.mul` and `gump.mul` (or their UOP equivalents). Gumps are the UI elements used in Ultima Online's interface, including containers, spellbooks, and various UI components.

## C# Reference Implementation Analysis

The C# implementation in `Gumps.cs` handles loading gump data from MUL files or UOP alternatives. It provides:

- Methods to load and decode gump images
- Support for both MUL and UOP file formats
- Handling of dimensions and image data
- Caching for improved performance

Gumps in UO use a specialized run-length encoding (RLE) format with 16-bit color data.

## Work Items

1. Create a new file `gump.go` in the root package.

2. Define the `Gump` struct to represent a single gump:

   ```go
   type Gump struct {
       ID      int
       Width   int
       Height  int

       // Fields to support lazy loading
       loaded  bool
       rawData []byte
       image   image.Image
       mutex   sync.RWMutex
   }

   // Image returns the decoded gump image
   func (g *Gump) Image() (image.Image, error) {
       // Implementation for lazy-loading and returning the image
   }
   ```

3. Add methods to the SDK struct for accessing gumps:

   ```go
   // Gump retrieves a specific gump by its ID
   func (s *SDK) Gump(id int) (*Gump, error) {
       // Implementation for retrieving a specific gump
   }

   // Gumps returns an iterator over all available gumps
   func (s *SDK) Gumps() iter.Seq[*Gump] {
       // Implementation for iterating over gumps
   }
   ```

4. Implement internal utility functions:

   ```go
   // Internal function for loading raw gump data
   func (s *SDK) loadGumpData(id int) ([]byte, int, int, error) {
       // Implementation for loading raw gump data and dimensions
   }

   // Decode converts raw gump data to an image
   func decodeGumpData(data []byte, width, height int) (image.Image, error) {
       // Implementation of the gump RLE decoding algorithm
   }
   ```

5. Write comprehensive unit tests in `gump_test.go`:
   - Test loading gump data from both MUL and UOP formats
   - Test decoding gump images
   - Test handling of dimensions and pixel data
   - Test lazy loading behavior
   - Test invalid/missing gump handling

## Key Considerations

- Gumps use a specialized RLE format that requires careful decoding
- The implementation should support both MUL and UOP file formats seamlessly
- Consider implementing lazy loading of image data to improve memory usage
- Implement proper caching to avoid repeated decoding of frequently used gumps
- Handle invalid or missing gump IDs gracefully
- Ensure thread safety for concurrent access
- Gump colors use the same 16-bit ARGB1555 format as other UO graphics
- Use the bitmap package (from Task 12) for image handling and conversion
- Common gumps are accessed frequently, so optimization is important

## Expected Output

A complete implementation that allows:

- Loading gump data from both MUL and UOP files
- Retrieving individual gumps by ID
- Decoding gumps into standard Go image.Image instances
- Efficient memory usage through lazy loading and caching

## Verification

- Compare decoded gump images with the C# implementation to ensure visual accuracy
- Test with known gump IDs to verify dimensions and visual appearance
- Verify that invalid gump IDs are handled correctly
- Benchmark performance, especially for frequently accessed gumps like containers
