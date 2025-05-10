# Task 3: Implement MUL File Reading Utilities (internal/mul)

## Objective

Create a specialized package for reading MUL (Multi-User Land) file formats, which are the primary data file format used by Ultima Online. This package will provide the foundation for reading binary data from MUL files.

## C# Reference Implementation Analysis

The primary reference is `BinaryExtensions.cs`, which provides extension methods for BinaryReader to facilitate reading UO-specific data types. The C# implementation relies heavily on .NET's BinaryReader and uses extension methods for additional functionality.

## Work Items

1. Create an `internal/mul` package directory.
2. Define a `Reader` struct that provides efficient access to MUL file data:

   ```go
   type Reader struct {
        // File handle
   }


   // NewReader creates and initializes a new MUL reader
   func NewReader(filename string) (*Reader, error) {
       // Open the file and initialize the reader
       // Handle any initialization in one step for efficiency
       // Return a fully usable reader
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

4. Implement methods for reading primitive data types from a byte slice:

   ```go
   // Helper methods that operate on byte slices for data parsing
   func ReadByte(data []byte, offset int) (byte, int, error)
   func ReadInt16(data []byte, offset int) (int16, int, error)
   func ReadUint16(data []byte, offset int) (uint16, int, error)
   func ReadInt32(data []byte, offset int) (int32, int, error)
   func ReadUint32(data []byte, offset int) (uint32, int, error)
   func ReadBytes(data []byte, offset, count int) ([]byte, int, error)
   func ReadString(data []byte, offset, fixedLength int) (string, int, error)
   ```

5. Write comprehensive unit tests in `reader_test.go`:
   - Test initialization of the reader with various MUL files
   - Test reading data at specific offsets
   - Test accessing entries by index
   - Test iterator methods
   - Test reading primitive data types
   - Test error handling for various edge cases

## Key Considerations

- Design the Reader to work efficiently with both standalone MUL files and MUL files with separate index files (.idx)
- Optimize for performance by using appropriate buffer sizes and minimizing allocations
- Use idiomatic Go error handling instead of exceptions
- Ensure byte ordering is correct (UO uses little-endian)
- Consider thread safety for concurrent access
- Align the interface with the UOP Reader (Task 4) to facilitate unification in Task 5
- Initialize resources efficiently in a single step rather than requiring multiple method calls
- Handle resources properly to avoid leaks, especially file handles
- Use iterators to provide a more idiomatic Go experience for working with indexed entries

## Expected Output

A robust `Reader` implementation in the `internal/mul` package that allows reading binary data from MUL files in a type-safe manner, with comprehensive test coverage and efficient resource management.

## Verification

- Verify that all reading methods correctly parse binary data according to the specification
- Compare results of reading test files with the C# implementation's output
- Ensure error handling works correctly for edge cases
- Verify proper resource management (opening/closing files)
- Test iterator methods to ensure they properly yield all entries
- Benchmark performance to ensure efficient reading of large files
