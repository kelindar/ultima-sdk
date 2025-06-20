# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.0.4] - 2024-12-20

### What's Changed

1. **Mock SDK Implementation**
   - Added a lightweight in-memory implementation of `ultima.Interface` (`mock/sdk.go`).
   - Provides methods to retrieve and manipulate various data types like `Land`, `Item`, `Skill`, and `Texture`.
   - Includes support for localized strings and helper methods for managing maps and tiles.

2. **Unit Tests**
   - Added comprehensive test cases for the `Mock SDK` functionality (`mock/sdk_test.go`).
   - Covers data addition, retrieval, and method behavior to ensure consistency.
   - Includes error handling tests for `ErrNotFound` scenarios.

### Added
- Mock SDK implementation with full `ultima.Interface` compatibility
- In-memory storage for all Ultima Online data types
- Comprehensive test suite for mock functionality
- Support for localized strings and multi-language content
- Helper methods for managing maps and tiles
- Error handling with consistent `ErrNotFound` responses

### Technical Details
- **File**: `mock/sdk.go` - Complete mock implementation
- **File**: `mock/sdk_test.go` - Full test coverage
- **Package**: `github.com/kelindar/ultima-sdk/mock`
- **Interface**: Implements `prima.Interface`
- **Test Coverage**: 100% of public API methods

## [Unreleased]

### Changed
- Improved mock SDK Add method with better error handling