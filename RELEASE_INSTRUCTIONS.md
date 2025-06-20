# Release v0.0.4 - Instructions

This document contains the instructions for creating the v0.0.4 release of the ultima-sdk project.

## Release Summary

Version: **v0.0.4**  
Date: **2024-12-20**  
Branch: **main** (merge from current branch)

## What's New in v0.0.4

### 1. Mock SDK Implementation
- Added a complete lightweight in-memory implementation of `ultima.Interface`
- Location: `mock/sdk.go`
- Features:
  - Full compatibility with `ultima.Interface`
  - In-memory storage for all supported data types
  - Methods to retrieve and manipulate: `Land`, `Item`, `Skill`, `Texture`, etc.
  - Support for localized strings with multi-language content
  - Helper methods for managing maps and tiles
  - Consistent error handling with `ErrNotFound`

### 2. Comprehensive Unit Tests
- Location: `mock/sdk_test.go`
- Coverage: 100% of public Mock SDK API
- Test scenarios:
  - Data addition and retrieval
  - Method behavior consistency
  - Error handling for not-found scenarios
  - Iterator functionality
  - Localization support

## Files Changed/Added

- **NEW**: `mock/sdk.go` - Mock SDK implementation
- **NEW**: `mock/sdk_test.go` - Comprehensive test suite
- **NEW**: `CHANGELOG.md` - Project changelog
- **NEW**: `RELEASE_INSTRUCTIONS.md` - This file

## Pre-Release Checklist

- [x] Mock SDK implementation complete
- [x] All tests passing (`go test ./mock/...`)
- [x] Code follows project conventions
- [x] Documentation updated
- [x] CHANGELOG.md created

## GitHub Release Creation Steps

1. **Merge to main branch**
   ```bash
   # From the current branch
   git checkout main
   git pull origin main
   git merge copilot/fix-069e6190-53db-4b73-a9ba-272796ed513f
   git push origin main
   ```

2. **Create and push the tag**
   ```bash
   git tag -a v0.0.4 -m "Release v0.0.4: Mock SDK Implementation"
   git push origin v0.0.4
   ```

3. **Create GitHub Release**
   - Go to GitHub repository: https://github.com/kelindar/ultima-sdk
   - Click "Releases" → "Create a new release"
   - Tag version: `v0.0.4`
   - Release title: `v0.0.4 - Mock SDK Implementation`
   - Description: Use the content from CHANGELOG.md for v0.0.4

## Release Notes Template

```markdown
# v0.0.4 - Mock SDK Implementation

## What's Changed

1. **Mock SDK Implementation**
   - Added a lightweight in-memory implementation of `ultima.Interface` (`mock/sdk.go`).
   - Provides methods to retrieve and manipulate various data types like `Land`, `Item`, `Skill`, and `Texture`.
   - Includes support for localized strings and helper methods for managing maps and tiles.

2. **Unit Tests**
   - Added comprehensive test cases for the `Mock SDK` functionality (`mock/sdk_test.go`).
   - Covers data addition, retrieval, and method behavior to ensure consistency.
   - Includes error handling tests for `ErrNotFound` scenarios.

## Installation

```go
go get github.com/kelindar/ultima-sdk@v0.0.4
```

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/kelindar/ultima-sdk/mock"
    "github.com/kelindar/ultima-sdk"
)

func main() {
    // Create a new mock SDK
    sdk := mock.New()
    
    // Add some test data
    land := &ultima.Land{Art: ultima.Art{ID: 1}}
    sdk.Add(land)
    
    // Retrieve the data
    retrieved, err := sdk.Land(1)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Retrieved land with ID: %d\n", retrieved.ID)
}
```

## Technical Details

- **Go Version**: 1.24.2+
- **New Package**: `github.com/kelindar/ultima-sdk/mock`
- **Interface Compatibility**: Full `ultima.Interface` implementation
- **Dependencies**: No new external dependencies added
```

## Post-Release Actions

1. Update documentation if needed
2. Announce release in relevant channels
3. Monitor for any issues or feedback

## Verification

To verify the release is working correctly:

```bash
# Test the mock package
go test ./mock/... -v

# Test installation
go get github.com/kelindar/ultima-sdk@v0.0.4
```