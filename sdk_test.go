package ultima

import (
	"os"
	"path/filepath"
	"testing"

	uotest "github.com/kelindar/ultima-sdk/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runWith runs a test with a properly initialized SDK instance using the test data directory.
// The function ensures the SDK is initialized with valid test data and passes it to the test function.
func runWith(t *testing.T, testFn func(*SDK)) {
	sdk, err := Open(uotest.Path())
	require.NoError(t, err, "failed to open SDK with test data directory")
	require.NotNil(t, sdk, "SDK instance should not be nil")
	testFn(sdk)
}

func TestOpenClose_WithValidDirectory(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// SDK is already opened by TestWith
		assert.NotNil(t, sdk, "SDK instance should not be nil from TestWith")
		assert.NotEmpty(t, sdk.BasePath(), "SDK BasePath should be set by TestWith")

		err := sdk.Close()
		assert.NoError(t, err, "Close() should not return an error")
		assert.Empty(t, sdk.BasePath(), "SDK BasePath should be empty after Close()")
	})
}

func TestOpen_InvalidPaths(t *testing.T) {
	t.Run("should return error for non-existent directory", func(t *testing.T) {
		nonExistentDir := "d:\\non\\existent\\directory\\for\\testing" // Adjusted for Windows, ensure cross-platform if needed
		sdk, err := Open(nonExistentDir)
		assert.Error(t, err, "Open() should return an error for a non-existent directory")
		assert.Nil(t, sdk, "Open() should return a nil SDK instance on error")
		assert.Contains(t, err.Error(), nonExistentDir, "Error message should contain the path of the non-existent directory")
	})

	t.Run("should return error if path is a file, not a directory", func(t *testing.T) {
		// Create a temporary file for testing
		tempFile, err := os.CreateTemp("", "ultima-sdk-test-file-*.txt")
		assert.NoError(t, err, "Failed to create temp file")
		filePath := tempFile.Name()
		tempFile.Close()          // Close the file handle
		defer os.Remove(filePath) // Clean up after the test

		sdk, err := Open(filePath)
		assert.Error(t, err, "Open() should return an error if path is a file")
		assert.Nil(t, sdk, "Open() should return a nil SDK instance on error")
		assert.Contains(t, err.Error(), filePath, "Error message should contain the path of the file")
		assert.Contains(t, err.Error(), "is not a directory", "Error message should indicate that the path is not a directory")
	})
}

func TestClose_Idempotent(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// SDK is already opened by TestWith
		assert.NotNil(t, sdk, "SDK instance should not be nil from TestWith")

		err := sdk.Close()
		assert.NoError(t, err, "First Close() should not return an error")
		assert.Empty(t, sdk.BasePath(), "SDK BasePath should be empty after first Close()")

		err = sdk.Close() // Call Close() again
		assert.NoError(t, err, "Second Close() should also not return an error")
		assert.Empty(t, sdk.BasePath(), "SDK BasePath should remain empty after multiple Close() calls")
	})
}

// Test file cleanup on SDK close
func TestFileCleanupOnClose(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Access some files
		_, _ = sdk.loadSkills()

		// Count how many files are in the cache
		count := 0
		sdk.files.Range(func(_, _ interface{}) bool {
			count++
			return true
		})

		assert.Greater(t, count, 0, "Cache should have files after access")

		// Close the SDK
		err := sdk.Close()
		assert.NoError(t, err, "Close() should not return an error")

		// Cache should be empty after close
		count = 0
		sdk.files.Range(func(_, _ interface{}) bool {
			count++
			return true
		})
		assert.Equal(t, 0, count, "Cache should be empty after SDK close")
	})
}

// Test file existence check
func TestFileExists(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Check if a common file exists in the test data
		exists := sdk.fileExists("hues.mul")
		// We can't assert the result because it depends on the test data,
		// but the method shouldn't panic

		// Create a temp file in the base path to test
		tempFileName := "test_file_exists_check.tmp"
		tempFilePath := filepath.Join(sdk.BasePath(), tempFileName)
		f, err := os.Create(tempFilePath)
		if err != nil {
			t.Logf("Could not create temporary file: %v", err)
			return
		}
		f.Close()
		defer os.Remove(tempFilePath)

		// Now check if our temp file exists
		exists = sdk.fileExists(tempFileName)
		assert.True(t, exists, "Temporary file should exist")
	})
}
