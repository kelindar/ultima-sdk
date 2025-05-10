# Task 5: Implement Unified File Access Logic (internal/file)

## Objective

Create a package that provides a unified interface for accessing both MUL and UOP file formats, handling file indexing, compression, and patching through a consistent API regardless of the underlying file format.

## C# Reference Implementation Analysis

The primary reference is `FileIndex.cs`, which manages file access across different formats. In the C# implementation, a single class handles different file types with conditional logic. The class also handles Verdata patches by overriding file data with patch content when applicable.

Additionally, `Helpers/MythicDecompress.cs` provides the decompression utilities for UOP files using zlib.

## Work Items

1. Create an `internal/file` package directory.
2. Define a `Reader` interface that both MUL and UOP readers will implement:

   ```go
    // Entry3D (offset, length, extra)
    type Entry3D = [3]uint32

    type Reader interface {

        // EntryAt retrieves entry information by its logical index
        EntryAt(index int) (Entry3D, error)

        // ReadAt reads data from a specific offset and length
        ReadAt(offset int64, length int) ([]byte, error)

        // ReadEntry reads the data for a specific entry
        Read(index int) ([]byte, error)

        // Entries returns an iterator over available entries
        Entries() iter.Seq[Entry3D]

        // Close releases resources
        Close() error
   }
   ```

3. Create Entry structures for index files:

   ```go
   type Entry struct {
       Offset int64
       Length int
       Extra  int  // Often used for misc. data like height, compression flags, etc.
   }
   ```

4. Implement a `File` struct that uses the `Reader` interface:

   ```go
   type File struct {
       Path    string
       Index   []Entry
       Reader  Reader
       UopMap  map[int]int64 // For UOP hash-to-index mapping if needed
       mu      sync.RWMutex  // For thread safety
   }
   ```

5. Implement methods for the `File` struct:

   - `NewFile(idxPath, mulPath, uopPath string) (*File, error)` - Create a new File instance
   - `Load() error` - Load index data
   - `ReadData(index int) ([]byte, error)` - Read data for a specific index
   - `ReadDataOffset(index int, offset, length int) ([]byte, error)` - Read partial data
   - `ApplyPatch(patch *verdata.Patch) error` - Apply a verdata patch
   - `Close() error` - Close associated readers

6. Implement compression utilities in `internal/file/compression.go`:

   - `Decompress(data []byte) ([]byte, error)` - Decompress zlib data
   - Other compression-related utilities as needed

7. Write comprehensive unit tests in `file_test.go`:
   - Test loading index files
   - Test reading data from MUL files
   - Test reading data from UOP files
   - Test applying verdata patches
   - Test compression/decompression

## Key Considerations

- The implementation should abstract away the differences between MUL and UOP formats
- Handle lazy initialization of file handles to avoid opening too many files at once
- Support for verdata patch application is crucial for accurate data
- Thread safety is important for concurrent access
- Efficient memory usage for large index files
- Handle edge cases like missing or corrupted files gracefully
- Consider implementing a cache for frequently accessed data
- The File struct should be usable by higher-level components without knowledge of the underlying format

## Expected Output

A robust `file` package that provides uniform access to UO data files regardless of their format, with support for verdata patching and compression handling.

## Verification

- Test with both MUL and UOP files to ensure consistent behavior
- Verify that index loading works correctly for different file types
- Ensure data reading matches the output from C# FileIndex for the same indices
- Verify that verdata patches are correctly applied
- Test compression/decompression against known good examples
- Check memory usage patterns for large files
