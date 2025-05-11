# Task 7: Port Hues.cs to hue.go

## Objective

Implement support for loading and processing hue data from `hues.mul`, which defines color palettes used to recolor various game assets. Create a clean, idiomatic Go API for accessing and using hue data.

## C# Reference Implementation Analysis

The C# `Hues` class manages loading of hue information from `hues.mul`. It features:

- A static array (`List`) containing all hues
- Methods to translate 16-bit color values
- Functionality to apply hues to image data
- Support for both indexed and RGB approaches to hue application

The format of `hues.mul` consists of blocks of data for each hue, including color tables and textual information.

## Work Items

1. Create a new file `hue.go` in the root package.

2. Implement the internal loading mechanism:

   ```go
   // Internal function to load hue data from hues.mul
   func (s *SDK) loadHues() (*uofile.File, error ){
       // should use s.load() to load the file
   }
   ```

3. Define the `Hue` struct that represents a single hue entry:

   ```go
   type Hue struct {
       Index      int
       Name       string
       Colors     [32]uint16 // Raw 16-bit color values
       TableStart uint16
       TableEnd   uint16
   }
   ```

4. Implement methods for the `Hue` struct:

   ```go
   // GetColor returns a standard Go color.Color for a specific entry in the hue's palette
   func (h *Hue) GetColor(paletteIndex int) (color.Color, error) {
       // Implementation for converting 16-bit color to standard Go color
   }

   // Image generates a small image.Image representing this hue's palette for visualization
   func (h *Hue) Image(widthPerColor, height int) image.Image {
       // Generate a simple image visualization of the hue's palette
   }
   ```

5. Add methods to the SDK struct for accessing hues:

   ```go
   // HueAt retrieves a specific hue by its index
   func (s *SDK) HueAt(index int) (*Hue, error) {
       // Implementation for retrieving a specific hue
   }

   // Hues returns an iterator over all available hues
   func (s *SDK) Hues() iter.Seq[*Hue] {
       // Implementation for iterating over hues
   }
   ```

6. Write comprehensive unit tests in `hue_test.go`:
   - Test loading hue data from a file
   - Test color conversion
   - Test generating hue palette images
   - Test accessing specific hues by index
   - Test iterating over all hues

## Key Considerations

- Hues in UO use a specific 16-bit color format (ARGB1555) that needs careful handling when converting to Go's color types
- Hue application to images will be implemented in the bitmap package, but the basic color conversions happen here
- Consider using lazy loading for the hue data to avoid unnecessary file operations
- Ensure proper error handling for all file operations and index bounds checking
- The name fields in hues.mul may contain invalid or non-printable characters
- Color values stored in hues.mul need bit manipulation to extract proper RGB values
- Consider thread safety if the SDK might be accessed concurrently

## Expected Output

A complete implementation of hue functionality that allows:

- Loading hue data from hues.mul
- Accessing individual hues by index
- Converting hue colors to standard Go color.Color values
- Generating visualization images of hue palettes
- Iterating over all available hues

## Verification

- Compare loaded hue data with the C# implementation to ensure identical color values
- Test color conversion by comparing rendered colors with known good examples
- Verify that all hues are correctly loaded from the file
- Test with edge cases like the first and last hues in the file
