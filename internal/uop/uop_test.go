package uop

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	uotest "github.com/kelindar/ultima-sdk/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildPatternName tests extracting UOP pattern from filename
func TestBuildPatternName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{filepath.Join("dir", "artLegacyMUL.uop"), "artLegacyMUL"},
		{"gumpartLegacyMUL.uop", "gumpartLegacyMUL"},
		{"/path/to/mapLegacyMUL.uop", "mapLegacyMUL"},
		{"file.with.multiple.dots.uop", "file.with.multiple.dots"},
		{"noextension", "noextension"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := BuildPatternName(tt.input)
			assert.Equal(t, tt.want, got, "BuildPatternName(%s)", tt.input)
		})
	}
}

// TestReaderOperations tests the UOP reader operations using actual UOP files
func TestReaderOperations(t *testing.T) {
	testDataPath := uotest.Path()
	require.NotEmpty(t, testDataPath, "Test data path should not be empty")

	// Find UOP files in test data directory
	uopFiles, err := filepath.Glob(filepath.Join(testDataPath, "*.uop"))
	if err != nil || len(uopFiles) == 0 {
		t.Skip("No UOP files found in test data directory")
		return
	}

	// Use the first found UOP file for testing
	testUOP := uopFiles[0]
	t.Logf("Testing with UOP file: %s", testUOP)

	// Test file opening
	reader, err := NewReader(testUOP)
	require.NoError(t, err)
	defer reader.Close()

	// Test basic file info
	_, err = os.Stat(testUOP)
	require.NoError(t, err)

	// Extract UOP pattern from filename
	pattern := BuildPatternName(filepath.Base(testUOP))

	// Test SetupEntries with a reasonable max index
	err = reader.SetupEntries(pattern, 5000) // A reasonable high limit for test
	require.NoError(t, err)

	// Test iterator
	var entriesCount int
	reader.Entries()(func(entry Entry3D) bool {
		// Verify the entry is not just zeros
		if entry[0] > 0 || entry[1] > 0 {
			entriesCount++
		}
		return true
	})

	// A real UOP file should have some entries
	t.Logf("Found %d entries in UOP file", entriesCount)
	assert.Greater(t, entriesCount, 0, "Should have found some entries in the UOP file")

	// Try reading an entry if we found any
	if entriesCount > 0 {
		// Find first valid entry
		var validIndex int = -1
		reader.Entries()(func(entry Entry3D) bool {
			validIndex++
			return entry[0] == 0 && entry[1] == 0
		})

		// If we found a valid entry, test reading it
		if validIndex >= 0 {
			entry, err := reader.EntryAt(validIndex)
			assert.NoError(t, err)

			// Try reading the entry's data
			data, err := reader.Read(validIndex)
			if assert.NoError(t, err) {
				assert.NotEmpty(t, data, "Entry data should not be empty")
				t.Logf("Successfully read entry %d with size %d bytes", validIndex, len(data))
			}

			// Also test direct ReadAt
			directData, err := reader.ReadAt(int64(entry[0]), int(entry[1]))
			if assert.NoError(t, err) {
				assert.Equal(t, data, directData, "Direct read should match entry read")
			}
		}
	}
}

// TestInvalidFiles tests error handling for invalid files
func TestInvalidFiles(t *testing.T) {
	// Test with non-existent file
	_, err := NewReader("non_existent_file.uop")
	assert.Error(t, err)

	// Create a temporary file with invalid content
	tmpFile, err := os.CreateTemp("", "invalid_*.uop")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write some invalid data (not a UOP file)
	_, err = tmpFile.Write([]byte("This is not a valid UOP file"))
	require.NoError(t, err)
	tmpFile.Close()

	// Try to open the invalid file
	_, err = NewReader(tmpFile.Name())
	assert.Error(t, err)
}

// TestGetEntryFromHash tests retrieving entries by hash
func TestGetEntryFromHash(t *testing.T) {
	testDataPath := uotest.Path()
	require.NotEmpty(t, testDataPath, "Test data path should not be empty")

	// Find UOP files in test data directory
	uopFiles, err := filepath.Glob(filepath.Join(testDataPath, "*.uop"))
	if err != nil || len(uopFiles) == 0 {
		t.Skip("No UOP files found in test data directory")
		return
	}

	// Use the first found UOP file for testing
	testUOP := uopFiles[0]

	reader, err := NewReader(testUOP)
	require.NoError(t, err)
	defer reader.Close()

	// Get file pattern and generate a hash
	pattern := BuildPatternName(filepath.Base(testUOP))
	testHash := HashFileName(fmt.Sprintf("build/%s/0000000000.dat", pattern))

	// Try to get entry from hash
	entry, err := reader.GetEntryFromHash(testHash)
	if err == ErrEntryNotFound {
		t.Skip("Test hash not found in UOP file")
	} else {
		assert.NoError(t, err)
		assert.NotNil(t, entry)
	}
}
