# Task 8: Port RadarCol.cs to radarcol.go

## Objective
Implement support for loading and accessing radar color data from `radarcol.mul`, which defines the colors used to represent different types of map tiles on the game's radar/mini-map.

## C# Reference Implementation Analysis
The C# implementation is relatively simple. `RadarCol.cs` loads color data from the `radarcol.mul` file into a static array. The file contains color mappings for land and static tiles that are used to render the in-game radar.

The file format is straightforward: it contains a series of 16-bit values representing colors in the ARGB1555 format (though some implementations may handle them as simple RGB565 values).

## Work Items
1. Create a new file `radarcol.go` in the root package.

2. Add methods to the SDK struct for radar color access:
   ```go
   // RadarColor retrieves the radar color for a given static tile ID
   func (s *SDK) RadarColor(staticTileID int) (uint16, error) {
       // Implementation for retrieving a specific radar color
   }

   // RadarColors returns an iterator over all defined radar color mappings
   func (s *SDK) RadarColors() iter.Seq2[int, uint16] {
       // Implementation for iterating over radar colors
   }
   ```

3. Implement the internal loading mechanism:
   ```go
   // Internal function to load radar color data from radarcol.mul
   func (s *SDK) loadRadarColors() error {
       // Implementation for loading radar color data
   }
   ```

4. Write comprehensive unit tests in `radarcol_test.go`:
   - Test loading radar color data from the file
   - Test accessing specific colors by tile ID
   - Test iterating over all colors
   - Test error handling for invalid tile IDs

## Key Considerations
- The radar color data is relatively simple compared to other UO file formats
- Colors are stored in a 16-bit format (often ARGB1555), and may need conversion for display
- There are typically 0x4000 land tile colors followed by 0x4000 static tile colors
- Ensure proper error handling for missing files or invalid indices
- Consider memory efficiency since the file is small and could be loaded entirely in memory

## Expected Output
A complete implementation that allows:
- Loading radar color data from radarcol.mul
- Retrieving colors for specific tile IDs
- Iterating over all available color mappings

## Verification
- Compare loaded color data with the C# implementation to ensure identical values
- Test accessing colors for well-known tile IDs and verify they match expected values
- Verify that the full range of colors is correctly loaded