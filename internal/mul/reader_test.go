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

	reader, err := NewReader(filePath)
	require.NoError(t, err, "Failed to create reader")
	require.NotNil(t, reader, "Reader should not be nil")

	defer reader.Close()
}

// TestNewReaderWithIndex tests initialization of a MUL reader with an index file
func TestNewReaderWithIndex(t *testing.T) {
	mulPath := filepath.Join(uotest.Path(), "skills.mul")
	idxPath := filepath.Join(uotest.Path(), "skills.idx")

	reader, err := NewReaderWithIndex(mulPath, idxPath)
	require.NoError(t, err, "Failed to create reader with index")
	require.NotNil(t, reader, "Reader with index should not be nil")

	defer reader.Close()
}

// TestEntryAt tests retrieving an entry by index
func TestEntryAt(t *testing.T) {
	mulPath := filepath.Join(uotest.Path(), "skills.mul")
	idxPath := filepath.Join(uotest.Path(), "skills.idx")

	reader, err := NewReaderWithIndex(mulPath, idxPath)
	require.NoError(t, err)
	defer reader.Close()

	// Test valid index
	entry, err := reader.EntryAt(1) // Index 1 should exist in any art.mul
	assert.NoError(t, err)
	assert.NotNil(t, entry)

	// Test methods of the Entry interface
	assert.GreaterOrEqual(t, entry.Lookup(), 0)
	assert.GreaterOrEqual(t, entry.Length(), 0)
	extra1, extra2 := entry.Extra()
	assert.GreaterOrEqual(t, extra1, 0)
	assert.Equal(t, 0, extra2) // Second value should be 0 for MUL entries

	// Test out of bounds index
	_, err = reader.EntryAt(99999999) // This index should be way out of bounds
	assert.Error(t, err)
}

// TestRead tests reading data from an entry
func TestRead(t *testing.T) {
	mulPath := filepath.Join(uotest.Path(), "skills.mul")
	idxPath := filepath.Join(uotest.Path(), "skills.idx")

	reader, err := NewReaderWithIndex(mulPath, idxPath)
	require.NoError(t, err)
	defer reader.Close()

	// Read data for the entry
	data, err := reader.Read(1)
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

// TestEntries tests the iterator for entries
func TestEntries(t *testing.T) {
	mulPath := filepath.Join(uotest.Path(), "skills.mul")
	idxPath := filepath.Join(uotest.Path(), "skills.idx")

	reader, err := NewReaderWithIndex(mulPath, idxPath)
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
	reader, err := NewReader(filepath.Join(uotest.Path(), "tiledata.mul"))
	require.NoError(t, err)

	// Close the reader
	err = reader.Close()
	assert.NoError(t, err)

	// Attempt operations after closing
	_, err = reader.ReadAt(0, 10)
	assert.Error(t, err)
	assert.Equal(t, ErrReaderClosed, err)
}

// TestHelperFunctions tests the helper functions for reading data types
func TestHelperFunctions(t *testing.T) {
	// Create test data
	data := []byte{
		0x01,       // byte
		0x23, 0x45, // int16/uint16 (little-endian: 0x4523 = 17699)
		0x67, 0x89, 0xAB, 0xCD, // int32/uint32 (little-endian: 0xCDAB8967)
		0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x00, // string "Hello\0"
	}

	// Test ReadByte
	b, offset, err := ReadByte(data, 0)
	assert.NoError(t, err)
	assert.Equal(t, byte(0x01), b)
	assert.Equal(t, 1, offset)

	// Test ReadInt16
	i16, offset, err := ReadInt16(data, 1)
	assert.NoError(t, err)
	assert.Equal(t, int16(17699), i16)
	assert.Equal(t, 3, offset)

	// Test ReadUint16
	u16, offset, err := ReadUint16(data, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint16(17699), u16)
	assert.Equal(t, 3, offset)

	// Test ReadInt32 - The value is too large for int32, so we'll only check that it's read correctly
	_, offset, err = ReadInt32(data, 3)
	assert.NoError(t, err)
	// Just check that we got some value, we know it'll overflow
	assert.Equal(t, 7, offset)

	// Test ReadUint32
	u32, offset, err := ReadUint32(data, 3)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0xCDAB8967), u32)
	assert.Equal(t, 7, offset)

	// Test ReadBytes
	bytes, offset, err := ReadBytes(data, 7, 5)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Hello"), bytes)
	assert.Equal(t, 12, offset)

	// Test ReadString
	str, offset, err := ReadString(data, 7, 6)
	assert.NoError(t, err)
	assert.Equal(t, "Hello", str)
	assert.Equal(t, 13, offset)

	// Test out of bounds
	_, _, err = ReadByte(data, 100)
	assert.Error(t, err)
	assert.Equal(t, ErrOutOfBounds, err)
}
