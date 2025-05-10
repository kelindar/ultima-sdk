# Task 15: Port Verdata.cs to verdata.go

## Objective
Implement support for loading and processing `verdata.mul`, which contains patches for various game files. The verdata system allows the UO client to update file content without replacing the original files.

## C# Reference Implementation Analysis
The C# implementation in `Verdata.cs` handles loading patch information from `verdata.mul` and provides this data to other components. Each patch entry specifies a file ID, index, and offset for patch data within verdata.mul itself.

The patch system is crucial for correctly displaying game content that has been updated from its original state in the MUL files.

## Work Items
1. Create a new file `verdata.go` in the root package.

2. Define the `VerdataPatch` struct that represents a single patch entry:
   ```go
   type VerdataPatch struct {
       FileID int32 // Identifier for the file being patched
       Index  int32 // Block number or entry ID within the file
       Lookup int32 // Offset in verdata.mul where the patch data begins
       Length int32 // Length of the patch data
       Extra  int32 // Extra data, often 0, but can be used for specific patch types
   }
   ```

3. Implement the SDK method for accessing verdata information:
   ```go
   // VerdataPatches returns an iterator over all verdata patch entries
   func (s *SDK) VerdataPatches() iter.Seq[VerdataPatch] {
       // Implementation for iterating over verdata patches
   }
   ```

4. Implement the internal loading mechanism:
   ```go
   // Internal function to load verdata information
   func (s *SDK) loadVerdata() error {
       // Implementation for loading patch entries
   }
   ```

5. Write unit tests in `verdata_test.go`:
   - Test loading verdata patch entries
   - Test iterating over patches
   - Test handling missing verdata.mul file (should not cause errors)
   - Test patch application mechanism (if implemented)

## Key Considerations
- The `verdata.mul` file contains a header with patch count followed by patch entries
- Each patch entry points to a location in the same file where the actual patch data resides
- Not all UO clients/installations have a verdata.mul file - the SDK should handle this gracefully
- Patches are typically integrated into the file reading process rather than applied explicitly
- The implementation should coordinate with the `internal/file` package to ensure patches are applied when data is read
- Consider creating a lookup mechanism to efficiently find patches for a specific file and index
- Ensure thread safety if patches might be accessed concurrently

## Expected Output
A complete implementation that allows:
- Loading patch entries from verdata.mul
- Providing these patches to other components (especially the file access system)
- Iterating over all available patches
- Handling the absence of verdata.mul gracefully

## Verification
- Verify that patch entries match those in the C# implementation
- Test accessing patched data in various files to ensure patches are correctly applied
- Check handling of edge cases like patches for nonexistent files
- Validate that the system works correctly when verdata.mul is missing