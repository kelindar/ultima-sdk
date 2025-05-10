# Task 4: Implement UOP File Handling (internal/uop)

## Objective

Create a package for handling UOP (Ultima Online Package) files, which is a newer format used in more recent UO clients. UOP files contain compressed data and a different indexing structure than traditional MUL files.

## C# Reference Implementation Analysis

The primary reference is `Helpers/UopUtils.cs`, which provides utilities for parsing UOP file format and extracting entries. The C# implementation handles the complex header structure, hash tables, and file record formats specific to UOP files.

Key C# methods to analyze include:

- `ReadUOPHash` - For hash calculations
- `GetFilename` - For resolving hashed filenames
- `ReadUOPFormat` - For main UOP file parsing

## Work Items

1. Create an `internal/uop` package directory.
2. Define the necessary structures for UOP file format:

   ```go
   type FileHeader struct {
       Signature   [4]byte // "MYP\0"
       Version     uint32
       Timestamp   uint32
       NextTable   uint64
       BlockCount  uint32
       // Other header fields as needed
   }

   type FileEntry struct {
       Offset       uint64
       HeaderLength uint32
       CompLength   uint32
       DecompLength uint32
       Hash         uint64
       Checksum     uint32
       Compression  uint16
       // Other fields as needed
   }

   type Reader struct {
       // File handle
   }
   ```

3. Implement methods for accessing MUL file data that complies to the following interface:

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

4. Implement hash calculation functions:

   - `hashFileName(name string) uint64` - Calculate UOP file hash from a name

5. Write comprehensive unit tests in `uop_test.go`:
   - Test initialization and parsing of UOP files
   - Test hash calculation against known values
   - Test extracting specific entries by hash and name
   - Test iterating over entries using the iterator methods
   - Test handling of invalid UOP files

## Key Considerations

- UOP files use a complex structure with hash tables for lookup
- File entries may be compressed (usually zlib), but leave decompression to `internal/file`
- UOP file format changes across different UO client versions
- Ensure correct handling of 64-bit values throughout the implementation
- Be careful with endianness (UOP files use little-endian)
- Optimize memory usage for large UOP files by not loading entire files into memory
- The hashing algorithm is critical to get right - verify it matches the C# implementation exactly
- Use iterators to provide a more idiomatic Go experience for working with collections of entries
- Initialize the reader fully in one step for better performance and simpler API

## Expected Output

A robust `Reader` implementation in the `internal/uop` package that can parse UOP files, locate entries by hash or name, extract raw (compressed) data from those entries, and provide idiomatic iteration over file entries.

## Verification

- Test parsing known UOP files and compare entry counts and metadata with the C# implementation
- Verify hash calculation with known filename-to-hash mappings
- Extract specific entries and compare their compressed data with what the C# implementation extracts
- Test the iterator methods to ensure they properly yield all entries
- Test with different UOP files from various UO client versions
