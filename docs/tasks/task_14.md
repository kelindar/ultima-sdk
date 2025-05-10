# Task 14: Port Light.cs to light.go

## Objective
Implement support for loading and accessing light source definitions from the `light.idx` and `light.mul` files, which define various light sources and their properties in Ultima Online.

## C# Reference Implementation Analysis
The C# implementation in `Light.cs` handles loading light information from `lightidx.mul` and `light.mul`. Light sources in UO define illumination patterns used for various game effects like torches, magic, and day/night cycles.

The files contain binary data describing the light patterns, dimensions, and intensity values.

## Work Items
1. Create a new file `light.go` in the root package.

2. Define the `Light` struct that represents a single light source:
   ```go
   type Light struct {
       ID     int
       Width  int    // Width of the light effect
       Height int    // Height of the light effect
       Data   []byte // Raw light data
   }
   ```

3. Add methods to the SDK struct for accessing light definitions:
   ```go
   // Light retrieves a specific light definition by its ID
   func (s *SDK) Light(id int) (*Light, error) {
       // Implementation for retrieving a specific light source
   }

   // Lights returns an iterator over all defined light sources
   func (s *SDK) Lights() iter.Seq[*Light] {
       // Implementation for iterating over light sources
   }
   ```

4. Implement the internal loading mechanism:
   ```go
   // Internal function to load light data
   func (s *SDK) loadLights() error {
       // Implementation for loading light data from lightidx.mul and light.mul
   }
   ```

5. Write unit tests in `light_test.go`:
   - Test loading light data from the files
   - Test accessing specific light sources by ID
   - Test iterating over all light sources
   - Test error handling for invalid light IDs

## Key Considerations
- The `lightidx.mul` file contains index information mapping to entries in `light.mul`
- The light data itself is structured as a pattern of intensity values
- Some light IDs may not be defined or may have zero dimensions
- Light effects often have specific dimensions and properties depending on their intended use
- Consider implementing lazy loading to avoid unnecessary file operations
- Ensure proper error handling for all file operations and index bounds checking
- Light data may be useful for rendering lighting effects in applications using the SDK

## Expected Output
A complete implementation that allows:
- Loading light source data from light.idx and light.mul
- Accessing individual light sources by ID
- Iterating over all available light sources
- Retrieving dimensions and pattern data for rendering

## Verification
- Compare loaded light dimensions with the C# implementation
- Verify that pattern data matches between implementations
- Test with common light IDs (e.g., torch, magic effects) to ensure correct loading
- Verify proper handling of undefined or invalid light entries