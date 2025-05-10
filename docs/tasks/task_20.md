# Task 20: Port ASCIIFont.cs and UnicodeFont.cs to font.go

## Objective

Implement support for loading and accessing font definitions from `fonts.mul` (ASCII fonts) and `unifont*.mul` files (Unicode fonts). These files define the character sets used for rendering text in the Ultima Online client.

## C# Reference Implementation Analysis

The C# implementation consists of two main classes:

- `ASCIIFont.cs` - Handles loading and managing ASCII fonts from fonts.mul
- `UnicodeFont.cs` - Handles loading and managing Unicode fonts from unifont\*.mul files

Both implementations provide methods for loading font data, measuring text, and rendering characters. ASCII fonts are bitmap-based with variable character widths, while Unicode fonts have additional complexity with separate character tables.

## Work Items

1. Create a new file `font.go` in the root package.

2. Define the `Font` interface that both ASCII and Unicode fonts will implement:

   ```go
   type Font interface {
       // GetCharacter returns the image for a specific character
       GetCharacter(c rune) (image.Image, error)

       // GetWidth returns the width of a character in pixels
       GetWidth(c rune) (int, error)

       // GetHeight returns the height of the font in pixels
       GetHeight() int

       // MeasureText returns the width of the given text in pixels
       MeasureText(text string) (int, error)

       // CreateImage renders text as an image
       CreateImage(text string) (image.Image, error)
   }
   ```

3. Define the `ASCIIFont` struct:

   ```go
   type ASCIIFont struct {
       ID          int
       Height      int
       characters  [256][]byte  // Raw character data
       widths      [256]int     // Character widths
       mutex       sync.RWMutex
   }
   ```

4. Define the `UnicodeFont` struct:

   ```go
   type UnicodeFont struct {
       ID            int
       Height        int
       characterMap  map[rune][]byte  // Raw character data
       widthMap      map[rune]int     // Character widths
       mutex         sync.RWMutex
   }
   ```

5. Add methods to the SDK struct for accessing fonts:

   ```go
   // ASCIIFont retrieves a specific ASCII font by its ID
   func (s *SDK) ASCIIFont(id int) (Font, error) {
       // Implementation for retrieving an ASCII font
   }

   // UnicodeFont retrieves a specific Unicode font by its ID
   func (s *SDK) UnicodeFont(id int) (Font, error) {
       // Implementation for retrieving a Unicode font
   }
   ```

6. Implement the internal loading mechanisms:

   ```go
   // Internal function to load ASCII font data
   func (s *SDK) loadASCIIFont(id int) (*ASCIIFont, error) {
       // Implementation for loading ASCII font data
   }

   // Internal function to load Unicode font data
   func (s *SDK) loadUnicodeFont(id int) (*UnicodeFont, error) {
       // Implementation for loading Unicode font data
   }
   ```

7. Write comprehensive unit tests in `font_test.go`:
   - Test loading ASCII and Unicode fonts
   - Test character rendering
   - Test text measurement
   - Test error handling for invalid font IDs or characters
   - Test image generation for text strings

## Key Considerations

- ASCII fonts (fonts.mul) use a different format than Unicode fonts (unifont\*.mul)
- ASCII fonts cover the standard 256 ASCII characters, while Unicode fonts support a broader range
- Font rendering should be pixel-perfect compared to the C# implementation
- Some characters may have special handling (e.g., color codes, formatting)
- Consider implementing lazy loading of character data to reduce memory usage
- Ensure thread safety for concurrent access
- Character measurements are crucial for proper text layout
- The implementation should work well with the bitmap package (from Task 12)
- Unicode fonts are stored across multiple files and may require special handling

## Expected Output

A complete implementation that allows:

- Loading both ASCII and Unicode fonts
- Retrieving individual character images and widths
- Measuring text strings
- Rendering text as images
- Convenient access through a common Font interface

## Verification

- Compare rendered characters with the C# implementation to ensure visual accuracy
- Test text measurement against known good values
- Verify character widths match the C# implementation
- Test with both ASCII and Unicode fonts to ensure consistent behavior
- Test with special characters and edge cases
