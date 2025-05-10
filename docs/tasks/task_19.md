# Task 19: Port Sound.cs and WaveFormat.cs to sound.go

## Objective

Implement support for loading and accessing sound data from `soundidx.mul` and `sound.mul` files. These files contain audio samples used for various in-game effects in Ultima Online, stored in a custom format based on WAV/PCM.

## C# Reference Implementation Analysis

The C# implementation spans two primary files:

- `Sound.cs` - Handles loading sound entries and providing access to their data
- `WaveFormat.cs` - Contains structures and utilities for working with WAV audio format

The implementation loads sound samples from the MUL files and provides methods to convert them to standard audio formats. The sounds are stored with headers that describe their format (PCM parameters).

## Work Items

1. Create a new file `sound.go` in the root package.

2. Define the `Sound` struct that represents a single sound entry:

   ```go
   type Sound struct {
       ID       int
       Name     string

       // WAV format information
       Channels      uint16
       SamplesPerSec uint32
       BitsPerSample uint16

       // The actual audio data
       Data []byte
   }

   // WavData returns the sound as a properly formatted WAV file
   func (s *Sound) WavData() ([]byte, error) {
       // Implementation for converting raw sound data to WAV format
   }
   ```

3. Add methods to the SDK struct for accessing sounds:

   ```go
   // Sound retrieves a specific sound by its ID
   func (s *SDK) Sound(id int) (*Sound, error) {
       // Implementation for retrieving a specific sound
   }

   // Sounds returns an iterator over all available sounds
   func (s *SDK) Sounds() iter.Seq[*Sound] {
       // Implementation for iterating over sounds
   }
   ```

4. Implement utility functions for WAV handling:

   ```go
   // WriteWaveHeader writes a WAV file header to the provided writer
   func WriteWaveHeader(w io.Writer, dataSize int, channels, samplesPerSec, bitsPerSample uint16) error {
       // Implementation for writing a standard WAV header
   }
   ```

5. Write comprehensive unit tests in `sound_test.go`:
   - Test loading sound data
   - Test converting to WAV format
   - Test handling of invalid sound IDs
   - Test basic audio properties (channels, sample rate, etc.)
   - Test integrity of the sound data

## Key Considerations

- Sound files in UO use a specialized format that includes PCM parameters
- Converting to standard WAV format requires adding proper headers
- Some sounds may be invalid or missing
- Handle error conditions gracefully (missing files, corrupt data)
- Consider implementing lazy loading to reduce memory usage
- Sounds could be large, so avoid unnecessary duplication in memory
- The implementation should not depend on external audio libraries for basic functionality
- Sound file naming conventions may need to be handled for certain utilities

## Expected Output

A complete implementation that allows:

- Loading sound data from soundidx.mul and sound.mul
- Retrieving individual sounds by ID
- Converting sound data to standard WAV format
- Accessing sound properties (channels, sample rate, etc.)
- Iterating over all available sounds

## Verification

- Compare sound data with the C# implementation to ensure accuracy
- Test conversion to WAV format and verify playability
- Verify sound properties match the C# implementation
- Test with a range of sound IDs to ensure proper handling
