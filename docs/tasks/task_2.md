# Task 2: Implement SDK Struct and Lifecycle (sdk.go)

## Objective

Implement the core SDK structure that will serve as the main entry point for the library. Define the `SDK` struct and implement the lifecycle methods `Open` and `Close`.

## C# Reference Implementation Analysis

In the C# SDK, this functionality is distributed across several files:

- `Files.cs` - Contains static methods for initializing and accessing UO files
- `Client.cs` - Provides connectivity to the UO client
- Other static manager classes that maintain global state

The Go implementation should be more idiomatic, avoiding global state and providing explicit initialization and cleanup.

## Work Items

1. Create `sdk.go` in the root package.
2. Define the `SDK` struct to hold client configuration and file path information:
   ```go
   type SDK struct {
       directory string        // Root directory of UO files
       files     map[string]string // Map of file paths (similar to C# MulPath)
       // Add other necessary fields for internal state tracking
   }
   ```
3. Implement the `Open` function:
   ```go
   func Open(directory string) (*SDK, error) {
       // Validate the directory
       // Initialize the SDK struct
       // Detect essential files
       return &SDK{}, nil
   }
   ```
4. Implement the `Close` method:
   ```go
   func (s *SDK) Close() error {
       // Clean up any resources (file handles, etc.)
       return nil
   }
   ```
5. Write unit tests in `sdk_test.go`:
   - Test successful initialization with valid directory
   - Test handling of invalid directory
   - Test proper cleanup of resources in Close

## Key Considerations

- Unlike the C# implementation which uses static classes and global state, follow Go's idiomatic approach of explicit instantiation and state management.
- The `SDK` struct should be designed to hold all necessary state without relying on globals.
- File handles should be opened lazily when needed rather than all at once.
- The `Close` method must properly release all resources to prevent leaks.
- Consider thread-safety for potential concurrent use (use mutexes if needed).
- Follow error handling conventions from Go, returning specific errors instead of using exceptions.

## Expected Output

A functional SDK struct with proper lifecycle management that serves as the foundation for all other components of the library.

## Verification

- Verify that the `Open` function correctly identifies UO files similar to the C# `Files.LoadMulPath()` method.
- Test with both valid and invalid paths to ensure proper error handling.
- Ensure resources are properly released when `Close` is called.
