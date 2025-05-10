# Task 23: Port Multis.cs & MultiHelpers.cs to multi.go

## Objective
Implement support for loading and accessing multi-object data from `multi.idx` and `multi.mul` (or UOP equivalents). Multi-objects are complex structures composed of multiple items, such as houses, boats, and other large objects in Ultima Online.

## C# Reference Implementation Analysis
The C# implementation consists of:
- `Multis.cs` - Main class for loading multi data
- `MultiComponentList.cs` - Represents the collection of items that make up a multi
- `MultiMap.cs` - Special visualization of multi structures
- `Helpers/MultiHelpers.cs` - Contains utility functions for multi manipulation

The implementation handles loading multi definitions, which consist of lists of items with relative coordinates that compose the structure. It also provides methods for working with these structures, like calculating bounds and rendering.

## Work Items
1. Create a new file `multi.go` in the root package.

2. Define the `MultiItem` struct that represents a single component of a multi:
   ```go
   type MultiItem struct {
       ItemID   int
       X        int16
       Y        int16
       Z        int16
       Flags    uint32
       // Additional fields from TileData as needed
   }
   ```

3. Define the `Multi` struct that represents a complete multi-object:
   ```go
   type Multi struct {
       ID       int
       Items    []MultiItem
       MinX     int16
       MinY     int16
       MinZ     int16
       MaxX     int16
       MaxY     int16
       MaxZ     int16
       Center   image.Point
   }

   // Methods for the Multi struct
   func (m *Multi) Width() int { /* ... */ }
   func (m *Multi) Height() int { /* ... */ }
   func (m *Multi) GetItemsAt(x, y int16) []MultiItem { /* ... */ }
   ```

4. Add methods to the SDK struct for accessing multi data:
   ```go
   // Multi retrieves a specific multi by its ID
   func (s *SDK) Multi(id int) (*Multi, error) {
       // Implementation for retrieving a specific multi
   }

   // Multis returns an iterator over all defined multis
   func (s *SDK) Multis() iter.Seq[*Multi] {
       // Implementation for iterating over multis
   }
   ```

5. Implement helper functions for multi operations:
   ```go
   // CreateMultiImage generates a 2D representation of the multi object
   func CreateMultiImage(m *Multi, artProvider func(int) (image.Image, error)) image.Image {
       // Implementation for creating a visual representation of the multi
   }
   ```

6. Implement the internal loading mechanism:
   ```go
   // Internal function to load multi data
   func (s *SDK) loadMulti(id int) (*Multi, error) {
       // Implementation for loading multi data
   }

   // Internal function to calculate multi bounds
   func calculateMultiBounds(items []MultiItem) (minX, minY, minZ, maxX, maxY, maxZ int16) {
       // Implementation for calculating multi bounds
   }
   ```

7. Write comprehensive unit tests in `multi_test.go`:
   - Test loading multi data
   - Test calculating bounds
   - Test retrieving items at specific coordinates
   - Test image generation
   - Test handling of invalid multi IDs

## Key Considerations
- Multi data should integrate with TileData for item properties
- Multi bounds calculation is important for proper rendering and collision detection
- The implementation should support both MUL and UOP file formats
- Multi objects can be large and complex, so efficient storage is important
- Some multi IDs may be invalid or missing
- Multi visualization requires integration with the art system
- Some multi items have special flags that affect their behavior
- Consider implementing a spatial index for efficient item lookup by coordinates
- Thread safety is important for concurrent access
- Housing multi objects are particularly important in UO and should be tested thoroughly

## Expected Output
A complete implementation that allows:
- Loading multi object data from multi.idx and multi.mul (or UOP)
- Retrieving individual multi objects by ID
- Accessing the component items of a multi
- Calculating bounds and dimensions
- Finding items at specific coordinates
- Generating visual representations of multi objects

## Verification
- Compare loaded multi data with the C# implementation
- Verify item counts and positions match between implementations
- Test bounds calculations for accuracy
- Test with known multi IDs (common houses, boats) to ensure correct loading
- Verify integration with the art system for visualization