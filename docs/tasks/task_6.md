# Task 6: Implement Top-Level File Accessors (files.go)

## Objective

Create top-level methods for the SDK struct that provide access to specific game files, serving as the interface between high-level SDK functions and the internal file access system. This layer should handle file path resolution, lazy initialization of file resources, and exposing a clean API for internal components to access game files.

## C# Reference Implementation Analysis

In the C# implementation, the `Files` class serves as a centralized registry of file paths and provides static methods to access specific game files. The class maintains a dictionary of file paths (`MulPath`) and a set of methods for resolving and accessing these paths (e.g., `GetFilePath`).

Additionally, the implementation handles:

- Registry lookups for UO installation paths
- Detection of file existence
- Lazy loading of file data
- File path resolution for various game file types

## Work Items

1. Integrate file path resolution and management into the `SDK` struct (in `sdk_files.go` or as part of `sdk.go`):

   ```go
   // Add to SDK struct
   type SDK struct {
       path  string
       files sync.Map // Lazily loaded file handles (string to *File)
   }
   ```

2. Implement method for file path resolution:

   ```go
   func (s *SDK) load(fileNames []string, length int, options ...uofile.Option) (*uofile.File, error) {
    // if not found in cache, open the file and cache it
   }
   ```

3. Write unit tests in `sdk_test.go`:
   - Test resolving file paths for different file types (use testdata as before)
   - Test file existence checks
   - Test accessing different types of game files
   - Test caching behavior

## Key Considerations

- Unlike the C# implementation which uses static methods, follow Go's idiomatic approach with methods on the SDK struct
- Use lazy loading to avoid opening all files at initialization
- Maintain a mapping of canonical file names to their actual paths
- Keep the API clean and focused on providing access to file resources

## Expected Output

A set of methods on the SDK struct that provide seamless access to UO game files regardless of their format or location, with proper resource management and error handling.

## Verification

- Verify that file paths are correctly resolved similar to the C# implementation
- Test access to files in both MUL and UOP formats
- Ensure files are only opened when needed and properly cached
- Verify proper cleanup of file handles when the SDK is closed
