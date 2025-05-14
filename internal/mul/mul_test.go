package mul

import (
	"path/filepath"
	"testing"

	uotest "github.com/kelindar/ultima-sdk/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewReader tests the basic initialization of a MUL reader
func TestNewReader(t *testing.T) {
	// Test with tiledata.mul which is a standalone MUL file (no idx file)
	filePath := filepath.Join(uotest.Path(), "tiledata.mul")

	reader, err := OpenOne(filePath)
	require.NoError(t, err, "Failed to create reader")
	require.NotNil(t, reader, "Reader should not be nil")

	defer reader.Close()
}

// TestNewReaderWithIndex tests initialization of a MUL reader with an index file
func TestNewReaderWithIndex(t *testing.T) {
	mulPath := filepath.Join(uotest.Path(), "skills.mul")
	idxPath := filepath.Join(uotest.Path(), "skills.idx")

	reader, err := Open(mulPath, idxPath)
	require.NoError(t, err, "Failed to create reader with index")
	require.NotNil(t, reader, "Reader with index should not be nil")

	defer reader.Close()
}

// TestEntryAt tests retrieving an entry by index
func TestEntryAt(t *testing.T) {
	mulPath := filepath.Join(uotest.Path(), "skills.mul")
	idxPath := filepath.Join(uotest.Path(), "skills.idx")

	reader, err := Open(mulPath, idxPath)
	require.NoError(t, err)
	defer reader.Close()

	// Test valid index
	entry, err := reader.entryAt(1) // Index 1 should exist in any art.mul
	assert.NoError(t, err)
	assert.NotNil(t, entry)

	// Test out of bounds index
	_, err = reader.entryAt(99999999) // This index should be way out of bounds
	assert.Error(t, err)
}

// TestRead tests reading data from an entry
func TestRead(t *testing.T) {
	mulPath := filepath.Join(uotest.Path(), "skills.mul")
	idxPath := filepath.Join(uotest.Path(), "skills.idx")

	reader, err := Open(mulPath, idxPath)
	require.NoError(t, err)
	defer reader.Close()

	// Read data for the entry
	data1, _, err := reader.Read(1)
	assert.NoError(t, err)
	assert.NotNil(t, data1)

	// Read data for the entry (cached)
	data2, _, err := reader.Read(1)
	assert.NoError(t, err)
	assert.NotNil(t, data2)

	// Check if the data is the same
	assert.Equal(t, data1, data2)
}

// TestEntries tests the iterator for entries
func TestEntries(t *testing.T) {
	mulPath := filepath.Join(uotest.Path(), "skills.mul")
	idxPath := filepath.Join(uotest.Path(), "skills.idx")

	reader, err := Open(mulPath, idxPath)
	require.NoError(t, err)
	defer reader.Close()

	// Count entries
	var count int
	for e := range reader.Entries() {
		assert.NotNil(t, e)
		count++
		// Just check a few entries to keep the test fast
		if count >= 100 {
			break
		}
	}

	assert.Greater(t, count, 0, "Should have found some entries")
}

// TestClose tests proper resource cleanup
func TestClose(t *testing.T) {
	reader, err := OpenOne(filepath.Join(uotest.Path(), "tiledata.mul"))
	require.NoError(t, err)

	// Close the reader
	err = reader.Close()
	assert.NoError(t, err)

	// Attempt operations after closing
	_, _, err = reader.Read(0)
	assert.Error(t, err)
	assert.Equal(t, ErrReaderClosed, err)
}
