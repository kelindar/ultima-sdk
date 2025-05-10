# Task 3: Implement MUL File Reading Utilities (internal/mul)

## Objective
Create a specialized package for reading MUL (Multi-User Land) file formats, which are the primary data file format used by Ultima Online. This package will provide the foundation for reading binary data from MUL files.

## C# Reference Implementation Analysis
The primary reference is `BinaryExtensions.cs`, which provides extension methods for BinaryReader to facilitate reading UO-specific data types. The C# implementation relies heavily on .NET's BinaryReader and uses extension methods for additional functionality.

## Work Items
1. Create an `internal/mul` package directory.
2. Define a `Reader` struct that wraps `io.Reader` for reading MUL file data:
   ```go
   type Reader struct {
       reader io.Reader
       buffer []byte  // For efficient reading
   }
   
   func NewReader(r io.Reader) *Reader {
       return &Reader{
           reader: r,
           buffer: make([]byte, someSuitableSize),
       }
   }
   ```
3. Implement methods for reading primitive data types:
   - `ReadByte() (byte, error)`
   - `ReadInt16() (int16, error)`
   - `ReadUint16() (uint16, error)`
   - `ReadInt32() (int32, error)`
   - `ReadUint32() (uint32, error)`
   - `ReadBytes(count int) ([]byte, error)`
   - `ReadString(fixedLength int) (string, error)` - For fixed-length and null-terminated strings
   
4. Implement specialized methods for UO-specific data types and formats:
   - Methods for reading coordinates, color values, etc.
   - Support for specialized UO data formats

5. Write comprehensive unit tests in `reader_test.go`:
   - Test all reading methods with known binary input and expected output
   - Test error handling for scenarios like EOF or malformed data

## Key Considerations
- Optimize for performance by avoiding excessive allocations
- Use idiomatic Go error handling instead of exceptions
- Consider providing both raw data methods and convenience methods for common UO data structures
- Ensure byte ordering is correct (UO uses little-endian)
- Consider thread safety if the reader might be used concurrently
- Follow Go conventions for IO handling and error propagation

## Expected Output
A robust `Reader` implementation in the `internal/mul` package that allows reading binary data from MUL files in a type-safe manner, with comprehensive test coverage.

## Verification
- Verify that all reading methods correctly parse binary data according to the specification
- Compare results of reading test files with the C# implementation's output
- Ensure error handling works correctly for edge cases
- Benchmark performance to ensure efficient reading of large files