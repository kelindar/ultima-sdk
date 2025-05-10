package mul

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadByte tests reading a byte from data
func TestReadByte(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}

	// Test valid read
	b, offset, err := ReadByte(data, 0)
	assert.NoError(t, err)
	assert.Equal(t, byte(0x01), b)
	assert.Equal(t, 1, offset)

	// Test read at specific offset
	b, offset, err = ReadByte(data, 2)
	assert.NoError(t, err)
	assert.Equal(t, byte(0x03), b)
	assert.Equal(t, 3, offset)

	// Test read out of bounds
	_, _, err = ReadByte(data, 5)
	assert.ErrorIs(t, err, ErrOutOfBounds)
}

// TestReadInt16 tests reading an int16 from data
func TestReadInt16(t *testing.T) {
	// 0x0201 (little endian) = 513
	data := []byte{0x01, 0x02, 0x03, 0x04}

	// Test valid read
	i, offset, err := ReadInt16(data, 0)
	assert.NoError(t, err)
	assert.Equal(t, int16(513), i)
	assert.Equal(t, 2, offset)

	// Test read at specific offset
	i, offset, err = ReadInt16(data, 2)
	assert.NoError(t, err)
	assert.Equal(t, int16(1027), i)
	assert.Equal(t, 4, offset)

	// Test read out of bounds
	_, _, err = ReadInt16(data, 3)
	assert.ErrorIs(t, err, ErrOutOfBounds)
}

// TestReadUint16 tests reading a uint16 from data
func TestReadUint16(t *testing.T) {
	// 0x0201 (little endian) = 513
	data := []byte{0x01, 0x02, 0x03, 0x04}

	// Test valid read
	i, offset, err := ReadUint16(data, 0)
	assert.NoError(t, err)
	assert.Equal(t, uint16(513), i)
	assert.Equal(t, 2, offset)

	// Test read at specific offset
	i, offset, err = ReadUint16(data, 2)
	assert.NoError(t, err)
	assert.Equal(t, uint16(1027), i)
	assert.Equal(t, 4, offset)

	// Test read out of bounds
	_, _, err = ReadUint16(data, 3)
	assert.ErrorIs(t, err, ErrOutOfBounds)
}

// TestReadInt32 tests reading an int32 from data
func TestReadInt32(t *testing.T) {
	// 0x04030201 (little endian) = 67305985
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	// Test valid read
	i, offset, err := ReadInt32(data, 0)
	assert.NoError(t, err)
	assert.Equal(t, int32(67305985), i)
	assert.Equal(t, 4, offset)

	// Test read at specific offset
	i, offset, err = ReadInt32(data, 4)
	assert.NoError(t, err)
	assert.Equal(t, int32(134678021), i)
	assert.Equal(t, 8, offset)

	// Test read out of bounds
	_, _, err = ReadInt32(data, 5)
	assert.ErrorIs(t, err, ErrOutOfBounds)
}

// TestReadUint32 tests reading a uint32 from data
func TestReadUint32(t *testing.T) {
	// 0x04030201 (little endian) = 67305985
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	// Test valid read
	i, offset, err := ReadUint32(data, 0)
	assert.NoError(t, err)
	assert.Equal(t, uint32(67305985), i)
	assert.Equal(t, 4, offset)

	// Test read at specific offset
	i, offset, err = ReadUint32(data, 4)
	assert.NoError(t, err)
	assert.Equal(t, uint32(134678021), i)
	assert.Equal(t, 8, offset)

	// Test read out of bounds
	_, _, err = ReadUint32(data, 5)
	assert.ErrorIs(t, err, ErrOutOfBounds)
}

// TestReadBytes tests reading a slice of bytes from data
func TestReadBytes(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	// Test valid read
	b, offset, err := ReadBytes(data, 1, 3)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x02, 0x03, 0x04}, b)
	assert.Equal(t, 4, offset)

	// Test read out of bounds
	_, _, err = ReadBytes(data, 6, 3)
	assert.ErrorIs(t, err, ErrOutOfBounds)
}

// TestReadString tests reading a string from data
func TestReadString(t *testing.T) {
	data := []byte{'H', 'e', 'l', 'l', 'o', 0, 'W', 'o', 'r', 'l', 'd'}

	// Test valid read with null terminator
	s, offset, err := ReadString(data, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, "Hello", s)
	assert.Equal(t, 10, offset)

	// Test valid read with fixed length
	s, offset, err = ReadString(data, 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, "Hello", s)
	assert.Equal(t, 5, offset)

	// Test read out of bounds
	_, _, err = ReadString(data, 9, 5)
	assert.ErrorIs(t, err, ErrOutOfBounds)
}

// createTempFileWithContent creates a temporary file with content and returns its path
func createTempFileWithContent(t *testing.T, content []byte) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "multest-*")
	require.NoError(t, err)
	defer tmpFile.Close()

	_, err = tmpFile.Write(content)
	require.NoError(t, err)

	return tmpFile.Name()
}

// TestNewReader tests creating a new MUL reader
func TestNewReader(t *testing.T) {
	// Create a temporary file
	content := []byte{0x01, 0x02, 0x03, 0x04}
	tmpFile := createTempFileWithContent(t, content)
	defer os.Remove(tmpFile)

	// Test successful creation
	reader, err := NewReader(tmpFile)
	assert.NoError(t, err)
	assert.NotNil(t, reader)
	defer reader.Close()

	// Test file not found
	_, err = NewReader("non-existent-file")
	assert.Error(t, err)
}

// TestNewReaderWithIndex tests creating a new MUL reader with an index file
func TestNewReaderWithIndex(t *testing.T) {
	// Create temporary MUL and IDX files
	mulContent := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	mulFile := createTempFileWithContent(t, mulContent)
	defer os.Remove(mulFile)

	// Create IDX file with two entries
	// Entry 1: offset=0, length=4, extra=0
	// Entry 2: offset=4, length=4, extra=1
	idxContent := []byte{
		0x00, 0x00, 0x00, 0x00, // Offset = 0
		0x04, 0x00, 0x00, 0x00, // Length = 4
		0x00, 0x00, 0x00, 0x00, // Extra = 0
		0x04, 0x00, 0x00, 0x00, // Offset = 4
		0x04, 0x00, 0x00, 0x00, // Length = 4
		0x01, 0x00, 0x00, 0x00, // Extra = 1
	}
	idxFile := createTempFileWithContent(t, idxContent)
	defer os.Remove(idxFile)

	// Test successful creation
	reader, err := NewReaderWithIndex(mulFile, idxFile)
	assert.NoError(t, err)
	assert.NotNil(t, reader)
	defer reader.Close()

	// Verify idxEntries were cached
	assert.Equal(t, 2, len(reader.idxEntries))
	assert.Equal(t, uint32(0), reader.idxEntries[0][0]) // Offset
	assert.Equal(t, uint32(4), reader.idxEntries[0][1]) // Length
	assert.Equal(t, uint32(0), reader.idxEntries[0][2]) // Extra
	assert.Equal(t, uint32(4), reader.idxEntries[1][0]) // Offset
	assert.Equal(t, uint32(4), reader.idxEntries[1][1]) // Length
	assert.Equal(t, uint32(1), reader.idxEntries[1][2]) // Extra

	// Test MUL file not found
	_, err = NewReaderWithIndex("non-existent-file", idxFile)
	assert.Error(t, err)

	// Test IDX file not found
	_, err = NewReaderWithIndex(mulFile, "non-existent-file")
	assert.Error(t, err)
}

// TestEntryAt tests retrieving an entry by index
func TestEntryAt(t *testing.T) {
	// Create temporary MUL and IDX files
	mulFile := createTempFileWithContent(t, []byte{0x01, 0x02, 0x03, 0x04})
	defer os.Remove(mulFile)

	// Create IDX file with one entry
	idxContent := []byte{
		0x00, 0x00, 0x00, 0x00, // Offset = 0
		0x04, 0x00, 0x00, 0x00, // Length = 4
		0x00, 0x00, 0x00, 0x00, // Extra = 0
	}
	idxFile := createTempFileWithContent(t, idxContent)
	defer os.Remove(idxFile)

	// Create reader with index
	reader, err := NewReaderWithIndex(mulFile, idxFile)
	assert.NoError(t, err)
	defer reader.Close()

	// Test valid entry retrieval
	entry, err := reader.EntryAt(0)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), entry[0]) // Offset
	assert.Equal(t, uint32(4), entry[1]) // Length
	assert.Equal(t, uint32(0), entry[2]) // Extra

	// Test invalid index
	_, err = reader.EntryAt(1)
	assert.ErrorIs(t, err, ErrInvalidIndex)

	// Test on closed reader
	reader.Close()
	_, err = reader.EntryAt(0)
	assert.ErrorIs(t, err, ErrReaderClosed)

	// Test without index file
	reader, err = NewReader(mulFile)
	assert.NoError(t, err)
	defer reader.Close()

	_, err = reader.EntryAt(0)
	assert.Error(t, err)
}

// TestReadAt tests reading data at a specific offset
func TestReadAt(t *testing.T) {
	// Create a temporary file with content
	content := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	tmpFile := createTempFileWithContent(t, content)
	defer os.Remove(tmpFile)

	// Create reader
	reader, err := NewReader(tmpFile)
	assert.NoError(t, err)
	defer reader.Close()

	// Test valid read
	data, err := reader.ReadAt(1, 3)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x02, 0x03, 0x04}, data)

	// Test read at end of file with partial read
	data, err = reader.ReadAt(6, 3)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x07, 0x08}, data)

	// Test read beyond file size
	_, err = reader.ReadAt(10, 1)
	assert.ErrorIs(t, err, ErrOutOfBounds)

	// Test read with negative offset
	_, err = reader.ReadAt(-1, 1)
	assert.ErrorIs(t, err, ErrInvalidOffset)

	// Test on closed reader
	reader.Close()
	_, err = reader.ReadAt(0, 1)
	assert.ErrorIs(t, err, ErrReaderClosed)
}

// TestRead tests reading entry data by index
func TestRead(t *testing.T) {
	// Create temporary MUL and IDX files
	mulContent := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	mulFile := createTempFileWithContent(t, mulContent)
	defer os.Remove(mulFile)

	// Create IDX file with two entries
	idxContent := []byte{
		0x00, 0x00, 0x00, 0x00, // Offset = 0
		0x04, 0x00, 0x00, 0x00, // Length = 4
		0x00, 0x00, 0x00, 0x00, // Extra = 0
		0x04, 0x00, 0x00, 0x00, // Offset = 4
		0x04, 0x00, 0x00, 0x00, // Length = 4
		0x01, 0x00, 0x00, 0x00, // Extra = 1
	}
	idxFile := createTempFileWithContent(t, idxContent)
	defer os.Remove(idxFile)

	// Create reader with index
	reader, err := NewReaderWithIndex(mulFile, idxFile)
	assert.NoError(t, err)
	defer reader.Close()

	// Test valid read
	data, err := reader.Read(0)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04}, data)

	data, err = reader.Read(1)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x05, 0x06, 0x07, 0x08}, data)

	// Test invalid index
	_, err = reader.Read(2)
	assert.ErrorIs(t, err, ErrInvalidIndex)

	// Create an IDX entry with invalid marker (0xFFFFFFFF)
	idxContent = []byte{
		0xFF, 0xFF, 0xFF, 0xFF, // Offset = 0xFFFFFFFF (invalid)
		0x04, 0x00, 0x00, 0x00, // Length = 4
		0x00, 0x00, 0x00, 0x00, // Extra = 0
	}
	invalidIdxFile := createTempFileWithContent(t, idxContent)
	defer os.Remove(invalidIdxFile)

	// Create reader with invalid index
	reader, err = NewReaderWithIndex(mulFile, invalidIdxFile)
	assert.NoError(t, err)
	defer reader.Close()

	// Test reading an invalid entry should return nil
	data, err = reader.Read(0)
	assert.NoError(t, err)
	assert.Nil(t, data)
}

// TestEntries tests iterating over available entries
func TestEntries(t *testing.T) {
	// Create temporary MUL and IDX files
	mulFile := createTempFileWithContent(t, []byte{0x01, 0x02, 0x03, 0x04})
	defer os.Remove(mulFile)

	// Create IDX file with two entries
	idxContent := []byte{
		0x00, 0x00, 0x00, 0x00, // Offset = 0
		0x02, 0x00, 0x00, 0x00, // Length = 2
		0x00, 0x00, 0x00, 0x00, // Extra = 0
		0x02, 0x00, 0x00, 0x00, // Offset = 2
		0x02, 0x00, 0x00, 0x00, // Length = 2
		0x01, 0x00, 0x00, 0x00, // Extra = 1
	}
	idxFile := createTempFileWithContent(t, idxContent)
	defer os.Remove(idxFile)

	// Create reader with index
	reader, err := NewReaderWithIndex(mulFile, idxFile)
	assert.NoError(t, err)
	defer reader.Close()

	// Test iterating over entries
	var entries []Entry3D
	for e := range reader.Entries() {
		entries = append(entries, e)
	}

	assert.Equal(t, 2, len(entries))
	assert.Equal(t, Entry3D{0, 2, 0}, entries[0])
	assert.Equal(t, Entry3D{2, 2, 1}, entries[1])

	// Test on closed reader
	reader.Close()
	var closedEntries []Entry3D
	for e := range reader.Entries() {
		closedEntries = append(closedEntries, e)
	}
	assert.Equal(t, 0, len(closedEntries))
}

// TestClose tests closing the reader
func TestClose(t *testing.T) {
	// Create temporary MUL and IDX files
	mulFile := createTempFileWithContent(t, []byte{0x01, 0x02, 0x03, 0x04})
	defer os.Remove(mulFile)

	idxFile := createTempFileWithContent(t, []byte{0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	defer os.Remove(idxFile)

	// Test closing a reader with both MUL and IDX files
	reader, err := NewReaderWithIndex(mulFile, idxFile)
	assert.NoError(t, err)
	assert.NoError(t, reader.Close())
	assert.True(t, reader.closed)
	assert.Nil(t, reader.file)
	assert.Nil(t, reader.idxFile)

	// Test closing already closed reader
	assert.NoError(t, reader.Close())

	// Test closing a reader with only MUL file
	reader, err = NewReader(mulFile)
	assert.NoError(t, err)
	assert.NoError(t, reader.Close())
	assert.True(t, reader.closed)
	assert.Nil(t, reader.file)

	// Test behavior when file close fails
	// This is hard to test directly without mocking or modifying file system
	// For now, we assume file system functions work correctly
}
