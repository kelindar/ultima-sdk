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

1. Integrate file path resolution and management into the `SDK` struct (in `files.go` or as part of `sdk.go`):

   ```go
   // Add to SDK struct
   type SDK struct {
       directory string
       files     map[string]string
       fileCache map[string]*internal.file.File  // Lazily loaded file handles
       mu        sync.RWMutex  // For thread safety
   }
   ```

2. Implement methods for file path resolution:

   ```go
   func (s *SDK) FilePath(filename string) string {
       // Resolve a file path within the UO directory structure
   }

   func (s *SDK) FileExists(filename string) bool {
       // Check if a file exists in the UO directory
   }
   ```

3. Implement methods to access specific file types:

   ```go
   func (s *SDK) GetArtFile() (*internal.file.File, error) {
       // Get access to art.mul/artidx.mul or artLegacyMul.uop
   }

   func (s *SDK) GetMapFile(mapID int) (*internal.file.File, error) {
       // Get access to map{mapID}.mul or map{mapID}LegacyMul.uop
   }

   // Similar methods for other file types
   ```

4. Implement internal caching to avoid repeatedly opening the same files:

   ```go
   func (s *SDK) getFile(key, idxPath, mulPath, uopPath string) (*internal.file.File, error) {
       // Check cache first, create and cache if not found
   }
   ```

5. Write unit tests in `files_test.go`:
   - Test resolving file paths for different file types
   - Test file existence checks
   - Test accessing different types of game files
   - Test caching behavior

## Key Considerations

- Unlike the C# implementation which uses static methods, follow Go's idiomatic approach with methods on the SDK struct
- Use lazy loading to avoid opening all files at initialization
- Implement proper resource management for file handles
- Consider thread safety for concurrent access
- Handle fallback paths (e.g., trying UOP first then MUL)
- Maintain a mapping of canonical file names to their actual paths
- Keep the API clean and focused on providing access to file resources

## Expected Output

A set of methods on the SDK struct that provide seamless access to UO game files regardless of their format or location, with proper resource management and error handling.

## Verification

- Verify that file paths are correctly resolved similar to the C# implementation
- Test access to files in both MUL and UOP formats
- Ensure files are only opened when needed and properly cached
- Verify proper cleanup of file handles when the SDK is closed
