package uop

import (
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

// TestEntryOperations tests entry-related operations (EntryAt, Read, Entries)
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

	assert.Equal(t, 10, count, "Should have found 10 entries in the first iteration")

	// Test EntryAt
	entry, err := reader.EntryAt(0)
	require.NoError(t, err)
	assert.NotNil(t, entry)

	// Check if the entry is valid
	data, err := reader.Read(entry)
	require.NoError(t, err)
	assert.NotNil(t, data)
}
