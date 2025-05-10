# Task 4: Implement UOP File Handling (internal/uop)

## Objective

Create a package for handling UOP (Ultima Online Package) files, which is a newer format used in more recent UO clients. UOP files contain compressed data and a different indexing structure than traditional MUL files.

## C# Reference Implementation Analysis

The primary reference is `Helpers/UopUtils.cs`, which provides utilities for parsing UOP file format and extracting entries. The C# implementation handles the complex header structure, hash tables, and file record formats specific to UOP files. Also relevant is the `UopFileAccessor` class in `FileIndex.cs`, which defines the `Entry6D` structure that has additional fields beyond the standard `Entry3D` to handle compression.

Key C# methods to analyze include:

- `ReadUOPHash` - For hash calculations
- `GetFilename` - For resolving hashed filenames
- `ReadUOPFormat` - For main UOP file parsing

## Work Items

1. Create an `internal/uop` package directory.

2. Define an `Entry6D` type specifically for UOP file entries, which includes additional fields for compression:

   ```go
   // CompressionFlag represents the compression method used for a UOP entry
   type CompressionFlag int16

   // Compression flag constants
   const (
       CompressionNone   CompressionFlag = 0
       CompressionZlib   CompressionFlag = 1
       CompressionMythic CompressionFlag = 3
   )

   // Entry6D represents an entry in UOP files with 6 components including compression info
   type Entry6D struct {
       Lookup            uint32         // Offset where the entry data begins
       Length            uint32         // Size of the entry data (compressed)
       DecompressedLength uint32        // Size after decompression
       Extra             uint32         // Extra data (can be split into Extra1/Extra2)
       Flag              CompressionFlag // Compression flag
   }

   // Convert Entry6D to the simplified Entry3D format when needed
   func (e Entry6D) ToEntry3D() Entry3D {
       return Entry3D{
           e.Lookup,
           e.Length,
           e.Extra,
       }
   }
   ```

3. Implement hash calculation functions:

   ```go
   // HashFileName calculates a hash for a filename as used in UOP files
   func HashFileName(s string) uint64
   ```

4. Define a `Reader` struct that provides efficient access to UOP file data:

   ```go
   type Reader struct {
       // File handle and other state
   }

   // NewReader creates and initializes a new UOP reader
   func NewReader(filename string) (*Reader, error)
   ```

5. Implement methods for accessing UOP file data that complies to the following interface:

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

6. Write comprehensive unit tests in `uop_test.go`:
   - Test initialization and parsing of UOP files
   - Test hash calculation against known values
   - Test extracting specific entries by hash and name
   - Test iterating over entries using the iterator methods
   - Test handling of invalid UOP files

## Key Considerations

- The `Entry6D` structure matches the C# reference implementation for UOP files, providing full compression metadata
- While internally using `Entry6D`, the external interface should be compatible with the common `Entry3D` format used by MUL files for interface uniformity
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
