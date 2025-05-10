package uop

import (
	"iter"
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
	reader, err := NewReader(testUOP, 10)
	require.NoError(t, err, "Failed to create reader")
	require.NotNil(t, reader, "Reader should not be nil")

	defer reader.Close()
}

// TestEntryOperations tests entry-related operations (Read and Entries methods)
func TestEntryOperations(t *testing.T) {
	testUOP := filepath.Join(uotest.Path(), "artLegacyMUL.uop")

	// Use the first found UOP file for testing
	patternBase := BuildPatternName(filepath.Base(testUOP))
	pattern := patternBase

	// Adjust pattern if it's a legacy MUL file
	if idx := len(pattern) - len("LegacyMUL"); idx > 0 && pattern[idx:] == "LegacyMUL" {
		pattern = pattern[:idx]
	}

	t.Logf("Using file: %s with pattern: %s", testUOP, pattern)

	reader, err := NewReader(testUOP, 0x14000, WithExtension(".tga"), WithIndexLength(0x13FDC))
	require.NoError(t, err)
	defer reader.Close()

	// Test Entries iterator with the new interface
	var count int
	var indices []uint64

	// Collect up to 10 entries for further testing
	for index := range reader.Entries() {
		indices = append(indices, index)
		count++
		if count >= 10 {
			break
		}
	}

	assert.GreaterOrEqual(t, count, 1, "Should have found at least 1 entry")

	// Test Read method with the first valid index
	if len(indices) > 0 {
		firstIndex := indices[0]
		data, err := reader.Read(firstIndex)
		require.NoError(t, err)
		assert.NotNil(t, data, "Data should not be nil")
		assert.Greater(t, len(data), 0, "Data should not be empty")

		// Test invalid index
		_, err = reader.Read(uint64(0xFFFFFFFF))
		assert.Error(t, err, "Reading invalid index should return error")
	}
}

// TestReaderInterface tests that Reader implements the required interface
func TestReaderInterface(t *testing.T) {
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

	// Create a reader
	reader, err := NewReader(testUOP, 10)
	require.NoError(t, err)
	defer reader.Close()

	// Test that we can assign the reader to a variable of the interface type
	var _ interface {
		Read(uint64) ([]byte, error)
		Entries() iter.Seq[uint64]
		Close() error
	} = reader
}
