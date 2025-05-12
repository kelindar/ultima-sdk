# Task 11: Port StringList.cs and StringEntry.cs to cliloc.go

## Objective

Implement support for loading and accessing localized strings from the `cliloc.*` files (e.g., `cliloc.enu` for English). These files contain localized text strings used throughout the UO client for various game elements.

## C# Reference Implementation Analysis

The C# implementation uses two main classes:

- `StringList` - Handles loading and managing a collection of string entries from a cliloc file
- `StringEntry` - Represents a single localized string with an ID and text content

The cliloc files use a specific binary format with headers and entries containing the string ID, text length, and text data.

## Work Items

1. Create a new file `cliloc.go` in the root package.

2. Implement methods for the SDK struct to access localized strings:

   ```go
   // String retrieves a localized string by its ID
   func (s *SDK) String(id int) (string, error) {
       // Implementation for retrieving a specific localized string
   }

   // Strings returns an iterator over all loaded localized strings
   func (s *SDK) Strings() iter.Seq2[int, string] {
       // Implementation for iterating over localized strings (ID and text)
   }
   ```

3. Write comprehensive unit tests in `cliloc_test.go`:
   - Test accessing strings by ID
   - Test iterating over all strings
   - Test handling of string formatting and special characters
   - Test error handling for invalid or missing string IDs

## Key Considerations

- The cliloc file format includes a header followed by string entries
- Each entry has an ID, length, and text content
- The strings may contain formatting codes or placeholders (e.g., ~1_NAME~)
- Consider implementing support for multiple languages if needed
- Character encoding in these files is typically UTF-16LE, which will need proper handling
- Some cliloc entries may be very large (e.g., quest text), so consider memory usage
- Consider implementing a cache for frequently accessed strings
- The SDK should handle null/empty strings gracefully
- For many games, English (`cliloc.enu`) is the default language if no other is specified

## Expected Output

A complete implementation that allows:

- Loading localized string data from cliloc files
- Retrieving specific strings by ID
- Iterating over all available strings
- Proper handling of string formatting and special characters

## Verification

- Compare loaded strings with the C# implementation to ensure identical content
- Test accessing commonly used system messages and verify their content
- Verify proper handling of UTF-16 encoding
- Test with strings containing formatting codes to ensure they're preserved
- Verify performance with large cliloc files
