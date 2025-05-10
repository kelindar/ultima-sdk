# Task 5: Implement Unified File Access Logic (internal/file)

## Objective

Create a package that provides a unified interface for accessing both MUL and UOP file formats, handling file indexing, compression, and patching through a consistent API regardless of the underlying file format.

## C# Reference Implementation Analysis

The primary reference is `FileIndex.cs`, which manages file access across different formats. In the C# implementation, the `FileIndex` class uses the `IFileAccessor` interface with implementations for both MUL and UOP formats. It also defines an `IEntry` interface that unifies `Entry3D` (MUL) and `Entry6D` (UOP) structures. The class also handles Verdata patches by overriding file data with patch content when applicable.

Additionally, `Helpers/MythicDecompress.cs` provides the decompression utilities for UOP files using zlib.

## Work Items

1. Create an `internal/uofile` package directory.

2. Implement a `File` struct that uses the `Reader` interface:

   ```go
   type File struct {
       Path    string
       Reader  Reader
       mu      sync.RWMutex  // For thread safety
   }
   ```

3. Implement methods for accessing both MUL and UOP file data that complies to the following interface:

   ```go
   type Reader interface {

   	// Read reads data from a specific entry
   	Read(index uint64) ([]byte, error)

   	// Entries returns an iterator over available entries
   	Entries() iter.Seq[uint64]

   	// Close releases resources
   	Close() error
   }

   ```

4. Write comprehensive unit tests in `file_test.go`:
   - Test loading index files for both formats
   - Test reading data from both MUL and UOP files
   - Test automatic format detection

## Key Considerations

- The implementation should abstract away the differences between MUL and UOP formats
- Handle lazy initialization of file handles to avoid opening too many files at once
- Support for verdata patch application is crucial for accurate data
- Thread safety is important for concurrent access
- Efficient memory usage for large index files
- Handle edge cases like missing or corrupted files gracefully
- Consider implementing a cache for frequently accessed data
- The File struct should be usable by higher-level components without knowledge of the underlying format
- Ensure the adaptation between specific readers and the unified interface preserves all necessary metadata
- Decompression should happen automatically when reading data unless specifically requested otherwise

## Expected Output

A robust `file` package that provides uniform access to UO data files regardless of their format, with support for verdata patching and compression handling. The package should present a consistent interface that hides the underlying format differences while preserving all the necessary functionality of both formats.

## Verification

- Test with both MUL and UOP files to ensure consistent behavior
- Verify that index loading works correctly for different file types
- Ensure data reading matches the output from C# FileIndex for the same indices
- Verify that verdata patches are correctly applied
- Test compression/decompression against known good examples
- Check memory usage patterns for large files
- Ensure thread safety with concurrent access
