# Task 18: Port Textures.cs to texture.go

## Objective
Implement support for loading and accessing texture maps from `texidx.mul` and `texmaps.mul` files. These textures are used for rendering 3D terrain in Ultima Online.

## C# Reference Implementation Analysis
The C# implementation in `Textures.cs` handles loading texture data. Each texture is a 64x64 pixel image used for terrain rendering. The implementation provides:
- Methods to load and decode texture images
- Caching for improved performance
- Support for different texture formats depending on client version

Textures are typically stored in a raw 16-bit color format similar to other UO graphics.

## Work Items
1. Create a new file `texture.go` in the root package.

2. Define the `Texture` struct that represents a single texture:
   ```go
   type Texture struct {
       ID     int
       Width  int // Typically 64
       Height int // Typically 64
       
       // Fields to support lazy loading
       loaded bool
       image  image.Image
       mutex  sync.RWMutex
   }

   // Image returns the decoded texture image
   func (t *Texture) Image() (image.Image, error) {
       // Implementation for lazy-loading and returning the image
   }
   ```

3. Add methods to the SDK struct for accessing textures:
   ```go
   // Texture retrieves a specific texture by its ID
   func (s *SDK) Texture(id int) (*Texture, error) {
       // Implementation for retrieving a specific texture
   }

   // Textures returns an iterator over all available textures
   func (s *SDK) Textures() iter.Seq[*Texture] {
       // Implementation for iterating over textures
   }
   ```

4. Implement internal utility functions:
   ```go
   // Internal function for loading texture data
   func (s *SDK) loadTextureData(id int) ([]byte, error) {
       // Implementation for loading raw texture data
   }
   
   // Decode converts raw texture data to an image
   func decodeTextureData(data []byte) (image.Image, error) {
       // Implementation for decoding texture data into an image
   }
   ```

5. Write comprehensive unit tests in `texture_test.go`:
   - Test loading texture data
   - Test decoding texture images
   - Test handling of invalid texture IDs
   - Test lazy loading behavior
   - Test texture dimensions and pixel data

## Key Considerations
- Textures are typically 64x64 pixels, but the implementation should be flexible
- The texture data uses 16-bit color format (ARGB1555)
- Consider implementing lazy loading to reduce memory usage
- Implement proper caching for frequently accessed textures
- Handle invalid or missing texture IDs gracefully
- Ensure thread safety for concurrent access
- Use the bitmap package (from Task 12) for image handling and conversion
- Textures may be read frequently during terrain rendering, so performance is important

## Expected Output
A complete implementation that allows:
- Loading texture data from texidx.mul and texmaps.mul
- Retrieving individual textures by ID
- Decoding textures into standard Go image.Image instances
- Efficient memory usage through lazy loading and caching

## Verification
- Compare decoded texture images with the C# implementation to ensure visual accuracy
- Test with known texture IDs to verify dimensions and visual appearance
- Verify that invalid texture IDs are handled correctly
- Test textures with different client versions if applicable