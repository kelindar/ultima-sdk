# Task 5: Implement Unified File Access Logic (internal/file)

## Objective

Create a package that provides a unified interface for accessing both MUL and UOP file formats, handling file indexing, compression, and patching through a consistent API regardless of the underlying file format.

## C# Reference Implementation Analysis

The primary reference is `FileIndex.cs`, which manages file access across different formats. In the C# implementation, the `FileIndex` class uses the `IFileAccessor` interface with implementations for both MUL and UOP formats. It also defines an `IEntry` interface that unifies `Entry3D` (MUL) and `Entry6D` (UOP) structures. The class also handles Verdata patches by overriding file data with patch content when applicable.

Additionally, `Helpers/MythicDecompress.cs` provides the decompression utilities for UOP files using zlib.

## Work Items

1. Create an `internal/file` package directory.

2. Define Entry and Reader interfaces that abstract the differences between MUL and UOP formats:

   ```go
    type Entry interface {
      // Lookup returns the offset in the file where the entry data begins
      Lookup() int

      // Length returns the size of the entry data
      Length() int

      // Extra returns additional data associated with the entry (extra1, extra2)
      Extra() (int, int)

      // Zip returns the size after decompression and compression flag (0=none, 1=zlib, 2=mythic)
      Zip() (int, byte)
    }

    type Reader interface {

        // EntryAt retrieves entry information by its hash
        EntryAt(uint64) (Entry, error)

        // Read reads data from a specific offset and length
        Read(entry Entry) ([]byte, error)

        // Entries returns an iterator over available entries
        Entries() iter.Seq[Entry]

        // Close releases resources
        Close() error
   }
   ```

3. Implement a `File` struct that uses the `Reader` interface:

   ```go
   type File struct {
       Path    string
       Reader  Reader
       mu      sync.RWMutex  // For thread safety
   }
   ```

4. Implement methods for the `File` struct:

   ```go
   // NewFile creates a new File instance that automatically selects between MUL and UOP formats
   func NewFile(idxPath, mulPath, uopPath string) (*File, error)

   // Load loads index data and prepares the file for reading
   func (f *File) Load() error

   // ReadData reads data for a specific index
   func (f *File) ReadData(index int) ([]byte, error)

   // ReadCompressedData reads data for a specific index without decompressing
   func (f *File) ReadCompressedData(index int) ([]byte, error)

   // ReadDataOffset reads partial data from an entry
   func (f *File) ReadDataOffset(index int, offset, length int) ([]byte, error)

   // ApplyPatch applies a verdata patch to an entry
   func (f *File) ApplyPatch(patch verdata.Patch) error

   // Close closes associated readers
   func (f *File) Close() error
   ```

5. Implement compression utilities in `internal/file/compression.go`:

   ```go
   // Decompress decompresses data based on the compression flag
   func Decompress(data []byte, flag CompressionFlag) ([]byte, error)

   // ZlibDecompress decompresses zlib data
   func ZlibDecompress(data []byte) ([]byte, error)

   // MythicDecompress decompresses data using the Mythic compression algorithm
   func MythicDecompress(data []byte) ([]byte, error)
   ```

6. Write adapter functions to make the specific readers (MUL's and UOP's) compatible with the unified interface:

   ```go
   // mulReaderAdapter adapts a mul.Reader to the unified Reader interface
   type mulReaderAdapter struct {
       reader *mul.Reader
   }

   // NewMulReader creates a new Reader from a MUL file
   func NewMulReader(filename string) (Reader, error)

   // uopReaderAdapter adapts a uop.Reader to the unified Reader interface
   type uopReaderAdapter struct {
       reader *uop.Reader
   }

   // NewUopReader creates a new Reader from a UOP file
   func NewUopReader(filename string) (Reader, error)
   ```

7. Write comprehensive unit tests in `file_test.go`:
   - Test loading index files for both formats
   - Test reading data from both MUL and UOP files
   - Test applying verdata patches
   - Test compression/decompression
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
