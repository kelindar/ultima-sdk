# Task 10: Port SpeechList.cs to speech.go

## Objective

Implement support for loading and accessing speech entries from the `speech.mul` file, which contains predefined speech phrases used by the UO client for keyword-based responses.

## C# Reference Implementation Analysis

The C# implementation in `SpeechList.cs` is relatively straightforward. It loads speech entries from `speech.mul` into a collection. Each entry consists of an ID and the corresponding text string.

The file format itself is simple, containing serialized speech entries with their associated keyword IDs. These are used by the client for keyword-triggered responses and system messages.

## Work Items

1. Create a new file `speech.go` in the root package.

2. Define the `Speech` struct that represents a single speech entry:

   ```go
   type Speech struct {
       ID   int
       Text string
   }
   ```

3. Add methods to the SDK struct for accessing speech entries:

   ```go
   // SpeechEntry retrieves a predefined speech entry by its ID
   func (s *SDK) SpeechEntry(id int) (*Speech, error) {
       // Implementation for retrieving a specific speech entry
   }

   // SpeechEntries returns an iterator over all defined speech entries
   func (s *SDK) SpeechEntries() iter.Seq[*Speech] {
       // Implementation for iterating over speech entries
   }
   ```

4. Write unit tests in `speech_test.go`:
   - Test accessing speech entries by ID
   - Test iterating over all speech entries
   - Test error handling for invalid or missing entries

## Key Considerations

- The `speech.mul` file has a simple structure but may require careful handling of text encoding
- Character encoding in these files is typically Windows-1252, not UTF-8
- Some speech entries may have special formatting or control characters
- Consider implementing lazy loading to avoid unnecessary file operations
- Ensure proper error handling for missing files or invalid indices
- Speech entries may include placeholders or special formatting codes that should be preserved

## Expected Output

A complete implementation that allows:

- Loading speech entry data from speech.mul
- Retrieving specific speech entries by ID
- Iterating over all available speech entries

## Verification

- Compare loaded speech text with the C# implementation to ensure identical content
- Verify that special characters and formatting are properly preserved
- Test accessing speech entries by common IDs and verify they match expected content
- Verify that the full range of speech entries is correctly loaded
