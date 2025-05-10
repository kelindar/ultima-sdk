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

// TestNewReader tests creating a new UOP reader
func TestNewReader(t *testing.T) {
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
	require.NoError(t, err, "Failed to create reader")
	require.NotNil(t, reader, "Reader should not be nil")

	defer reader.Close()
}

// TestInvalidFiles tests error handling for invalid files
func TestInvalidFiles(t *testing.T) {
	// Test with non-existent file
	_, err := NewReader("non_existent_file.uop")
	assert.Error(t, err)

	// Try to open a non-UOP file as UOP
	filePath := filepath.Join(uotest.Path(), "tiledata.mul") // MUL file, not UOP
	if _, err := os.Stat(filePath); err == nil {
		_, err = NewReader(filePath)
		assert.Error(t, err, "Should fail to open non-UOP file as UOP")
	}

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

// TestEntryOperations tests entry-related operations (EntryAt, Read, Entries)
func TestEntryOperations(t *testing.T) {
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
	patternBase := BuildPatternName(filepath.Base(testUOP))
	pattern := patternBase

	// Adjust pattern if it's a legacy MUL file
	if idx := len(pattern) - len("LegacyMUL"); idx > 0 && pattern[idx:] == "LegacyMUL" {
		pattern = pattern[:idx]
	}

	t.Logf("Using file: %s with pattern: %s", testUOP, pattern)

	reader, err := NewReader(testUOP)
	require.NoError(t, err)
	defer reader.Close()

	// Setup entries mapping
	err = reader.SetupEntries(pattern, 10000) // A reasonable upper limit for test
	require.NoError(t, err)

	// Test Entries iterator
	var count int
	for entry := range reader.Entries() {
		assert.NotNil(t, entry)
		count++

		// Just check a few entries to keep the test fast
		if count >= 10 {
			break
		}
	}

	t.Logf("Found %d entries in first 10 iterations", count)

	// If we found entries, test EntryAt and Read
	if count > 0 {
		// Try to retrieve a specific entry by hash
		testFilename := fmt.Sprintf("build/%s/00000000.dat", pattern)
		hash := HashFileName(testFilename)
		t.Logf("Testing with hash for %s: 0x%X", testFilename, hash)

		entry, err := reader.EntryAt(hash)
		if err == nil {
			assert.NotNil(t, entry)
			t.Logf("Found entry with Lookup: %d, Length: %d", entry.Lookup(), entry.Length())

			// Test methods of the Entry interface
			assert.GreaterOrEqual(t, entry.Lookup(), 0)
			assert.GreaterOrEqual(t, entry.Length(), 0)

			// Test Zip() method for compression info
			decompSize, compFlag := entry.Zip()
			t.Logf("Decompressed size: %d, Compression flag: %d", decompSize, compFlag)
			assert.GreaterOrEqual(t, decompSize, 0)
			assert.GreaterOrEqual(t, int(compFlag), 0)

			// Test Read method
			data, err := reader.Read(entry)
			if assert.NoError(t, err) {
				assert.NotNil(t, data)
				assert.Len(t, data, entry.Length())
				t.Logf("Read %d bytes of data", len(data))
			}
		} else {
			t.Logf("Entry not found by hash: %v", err)
		}

		// Also test GetEntryByName
		entry, err = reader.GetEntryByName(testFilename)
		if err == nil {
			t.Logf("Found entry by name with Lookup: %d, Length: %d",
				entry.Lookup(), entry.Length())
		} else {
			t.Logf("Entry not found by name: %v", err)
		}
	} else {
		t.Skip("No entries found to test EntryAt and Read")
	}
}
