# Task 22: Port Animations Code to anim.go and internal/anim

## Objective

Implement support for loading and accessing animation data from Ultima Online animation files. This is one of the most complex parts of the SDK, requiring implementation of body conversion, animation frame decoding, and handling of different animation file formats.

## C# Reference Implementation Analysis

The C# implementation spans multiple files:

- `Animations.cs` - Main class for loading and accessing animation frames
- `AnimationEdit.cs` - Contains utilities for animation editing and manipulation
- `Animdata.cs` - Handles loading animation metadata from animdata.mul
- `BodyConverter.cs` - Handles conversion between body IDs and file IDs
- `BodyTable.cs` - Maps body types to file IDs and animation groups

The animation system is highly complex, with several lookup tables and conversion systems to map creature body types to their corresponding animation files. It also needs to handle multiple animation file formats and frame decoding algorithms.

## Work Items

1. Create a new file `anim.go` in the root package for the public API.

2. Create an `internal/anim` package for implementation details.

3. Define core animation structures in `anim.go`:

   ```go
   // Animation represents a complete animation sequence
   type Animation struct {
       BodyID   int
       Action   int
       Direction int
       Frames   []*AnimationFrame
   }

   // AnimationFrame represents a single frame of an animation
   type AnimationFrame struct {
       ID       int
       Width    int
       Height   int
       CenterX  int
       CenterY  int

       // Fields to support lazy loading
       loaded   bool
       rawData  []byte
       image    image.Image
       mutex    sync.RWMutex
   }

   // Image returns the decoded frame image
   func (af *AnimationFrame) Image() (image.Image, error) {
       // Implementation for lazy-loading and returning the image
   }
   ```

4. Implement necessary internal structures in `internal/anim`:

   ```go
   // Define structures for animation metadata, body conversion tables, etc.
   type AnimData struct {
       // Implementation details
   }

   type BodyTable struct {
       // Implementation details
   }

   type BodyConverter struct {
       // Implementation details
   }
   ```

5. Add methods to the SDK struct for accessing animations:

   ```go
   // GetAnimation retrieves a specific animation by body ID, action, and direction
   func (s *SDK) GetAnimation(bodyID, action, direction int) (*Animation, error) {
       // Implementation for retrieving a specific animation
   }

   // GetAnimationFrame retrieves a single animation frame
   func (s *SDK) GetAnimationFrame(bodyID, action, direction, frame int) (*AnimationFrame, error) {
       // Implementation for retrieving a specific frame
   }
   ```

6. Implement internal loading mechanisms:

   ```go
   // Internal function to load animation data
   func (s *SDK) loadAnimationData() error {
       // Implementation for loading animation metadata and lookup tables
   }

   // Internal function to load a specific animation
   func (s *SDK) loadAnimation(bodyID, action, direction int) (*Animation, error) {
       // Implementation for loading a specific animation sequence
   }
   ```

7. Write comprehensive unit tests:
   - Test loading animation metadata
   - Test body conversion functions
   - Test retrieving animations for various body types
   - Test frame decoding and image generation
   - Test handling of different file formats

## Key Considerations

- Multiple animation file formats exist (MUL, UOP, and various versions)
- Body IDs must be converted to file IDs using complex lookup tables
- Animation frames use specialized encoding formats that must be carefully decoded
- Animation data is large, so efficient memory usage through lazy loading is crucial
- Frame centering information is important for proper rendering
- Different creatures and characters have different animation sets
- Some animations may be missing or invalid
- Animation metadata from animdata.mul provides crucial movement timing information
- Lookup tables from files like body.def and bodyconv.def need to be properly parsed
- Thread safety is important for concurrent access

## Expected Output

A complete implementation that allows:

- Loading animation metadata and lookup tables
- Converting between body IDs and file IDs
- Retrieving animations for specific body types, actions, and directions
- Decoding individual frames into standard Go image.Image instances
- Accessing animation timing and metadata
- Efficient memory usage through lazy loading and caching

## Verification

- Compare decoded animation frames with the C# implementation
- Verify body conversion logic works correctly for various body types
- Test with known animations to ensure frames are decoded accurately
- Verify frame dimensions and center points match the C# implementation
- Test with both MUL and UOP file formats if applicable
