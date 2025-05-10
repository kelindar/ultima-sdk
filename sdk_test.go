package ultima

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenClose_WithValidDirectory(t *testing.T) {
	TestWith(t, func(t *testing.T, sdk *SDK) {
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
	TestWith(t, func(t *testing.T, sdk *SDK) {
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
