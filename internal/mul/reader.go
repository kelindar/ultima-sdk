// Package mul provides utilities for reading Ultima Online MUL files.
package mul

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"sync"
)

// Entry3D represents an entry in a MUL file with offset, length, and extra data
type Entry3D = [3]uint32

// Reader provides access to MUL file data
type Reader struct {
	file       *os.File     // File handle for the MUL file
	idxFile    *os.File     // Optional index file handle
	idxEntries []Entry3D    // Cached index entries
	entrySize  int          // Size of each index entry (typically 12 bytes for 3 uint32s)
	mu         sync.RWMutex // Mutex for thread safety
	closed     bool         // Flag to track if reader is closed
}

// Errors
var (
	ErrReaderClosed  = errors.New("mul reader is closed")
	ErrOutOfBounds   = errors.New("read operation would exceed file bounds")
	ErrInvalidIndex  = errors.New("invalid index")
	ErrInvalidOffset = errors.New("invalid offset")
)

// NewReader creates and initializes a new MUL reader
func NewReader(filename string) (*Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open MUL file: %w", err)
	}

	return &Reader{
		file:      file,
		entrySize: 12, // Default entry size is 12 bytes (3 uint32s)
	}, nil
}

// NewReaderWithIndex creates a new MUL reader with a separate index file
func NewReaderWithIndex(mulFilename, idxFilename string) (*Reader, error) {
	// Open MUL file
	file, err := os.Open(mulFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open MUL file: %w", err)
	}

	// Open IDX file
	idxFile, err := os.Open(idxFilename)
	if err != nil {
		file.Close() // Clean up MUL file handle if IDX file can't be opened
		return nil, fmt.Errorf("failed to open IDX file: %w", err)
	}

	reader := &Reader{
		file:      file,
		idxFile:   idxFile,
		entrySize: 12, // Default entry size is 12 bytes (3 uint32s)
	}

	// Cache index entries
	if err := reader.cacheIndexEntries(); err != nil {
		reader.Close() // Clean up both file handles if caching fails
		return nil, fmt.Errorf("failed to cache index entries: %w", err)
	}

	return reader, nil
}

// cacheIndexEntries loads all index entries from the index file into memory
func (r *Reader) cacheIndexEntries() error {
	if r.idxFile == nil {
		return errors.New("no index file provided")
	}

	// Get file size
	info, err := r.idxFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get index file stats: %w", err)
	}

	fileSize := info.Size()
	entryCount := int(fileSize) / r.entrySize

	// Allocate slice for entries
	r.idxEntries = make([]Entry3D, entryCount)

	// Read all entries at once
	data := make([]byte, fileSize)
	_, err = r.idxFile.ReadAt(data, 0)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to read index file: %w", err)
	}

	// Parse entries
	for i := 0; i < entryCount; i++ {
		offset := i * r.entrySize
		r.idxEntries[i][0] = binary.LittleEndian.Uint32(data[offset : offset+4])    // Offset
		r.idxEntries[i][1] = binary.LittleEndian.Uint32(data[offset+4 : offset+8])  // Length
		r.idxEntries[i][2] = binary.LittleEndian.Uint32(data[offset+8 : offset+12]) // Extra
	}

	return nil
}

// EntryAt retrieves entry information by its logical index
func (r *Reader) EntryAt(index int) (Entry3D, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return Entry3D{}, ErrReaderClosed
	}

	if r.idxEntries != nil {
		// If we have cached index entries, retrieve from cache
		if index < 0 || index >= len(r.idxEntries) {
			return Entry3D{}, ErrInvalidIndex
		}
		return r.idxEntries[index], nil
	}

	// If we don't have an index file, we can't retrieve by index
	return Entry3D{}, errors.New("index file not provided")
}

// ReadAt reads data from a specific offset and length
func (r *Reader) ReadAt(offset int64, length int) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, ErrReaderClosed
	}

	if offset < 0 {
		return nil, ErrInvalidOffset
	}

	// Check file size to ensure the offset is valid
	info, err := r.file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	// Check if the offset is beyond the file size
	fileSize := info.Size()
	if offset >= fileSize {
		return nil, ErrOutOfBounds
	}

	// Adjust length if it would read beyond the end of the file
	if offset+int64(length) > fileSize {
		length = int(fileSize - offset)
	}

	data := make([]byte, length)
	n, err := r.file.ReadAt(data, offset)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	// Adjust slice if we didn't read the full amount (e.g., EOF)
	if n < length {
		data = data[:n]
	}

	return data, nil
}

// Read reads the data for a specific entry
func (r *Reader) Read(index int) ([]byte, error) {
	entry, err := r.EntryAt(index)
	if err != nil {
		return nil, err
	}

	// Skip invalid entries (offset == 0xFFFFFFFF or length == 0)
	if entry[0] == 0xFFFFFFFF || entry[1] == 0 {
		return nil, nil
	}

	return r.ReadAt(int64(entry[0]), int(entry[1]))
}

// Entries returns an iterator over available entries
func (r *Reader) Entries() iter.Seq[Entry3D] {
	return func(yield func(Entry3D) bool) {
		r.mu.RLock()
		defer r.mu.RUnlock()

		if r.closed {
			return
		}

		// Return entries from cache if available
		for _, entry := range r.idxEntries {
			if !yield(entry) {
				return
			}
		}
	}
}

// Close releases resources
func (r *Reader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	var errs []error

	if r.file != nil {
		if err := r.file.Close(); err != nil {
			errs = append(errs, err)
		}
		r.file = nil
	}

	if r.idxFile != nil {
		if err := r.idxFile.Close(); err != nil {
			errs = append(errs, err)
		}
		r.idxFile = nil
	}

	r.idxEntries = nil

	if len(errs) > 0 {
		return fmt.Errorf("failed to close files: %v", errs)
	}

	return nil
}

// ReadByte reads a single byte from data at the specified offset
func ReadByte(data []byte, offset int) (byte, int, error) {
	if offset < 0 || offset >= len(data) {
		return 0, offset, ErrOutOfBounds
	}
	return data[offset], offset + 1, nil
}

// ReadInt16 reads an int16 from data at the specified offset
func ReadInt16(data []byte, offset int) (int16, int, error) {
	if offset < 0 || offset+2 > len(data) {
		return 0, offset, ErrOutOfBounds
	}
	return int16(binary.LittleEndian.Uint16(data[offset:])), offset + 2, nil
}

// ReadUint16 reads a uint16 from data at the specified offset
func ReadUint16(data []byte, offset int) (uint16, int, error) {
	if offset < 0 || offset+2 > len(data) {
		return 0, offset, ErrOutOfBounds
	}
	return binary.LittleEndian.Uint16(data[offset:]), offset + 2, nil
}

// ReadInt32 reads an int32 from data at the specified offset
func ReadInt32(data []byte, offset int) (int32, int, error) {
	if offset < 0 || offset+4 > len(data) {
		return 0, offset, ErrOutOfBounds
	}
	return int32(binary.LittleEndian.Uint32(data[offset:])), offset + 4, nil
}

// ReadUint32 reads a uint32 from data at the specified offset
func ReadUint32(data []byte, offset int) (uint32, int, error) {
	if offset < 0 || offset+4 > len(data) {
		return 0, offset, ErrOutOfBounds
	}
	return binary.LittleEndian.Uint32(data[offset:]), offset + 4, nil
}

// ReadBytes reads a slice of bytes from data at the specified offset
func ReadBytes(data []byte, offset, count int) ([]byte, int, error) {
	if offset < 0 || offset+count > len(data) {
		return nil, offset, ErrOutOfBounds
	}
	return data[offset : offset+count], offset + count, nil
}

// ReadString reads a fixed-length string from data at the specified offset
func ReadString(data []byte, offset, fixedLength int) (string, int, error) {
	if offset < 0 || offset+fixedLength > len(data) {
		return "", offset, ErrOutOfBounds
	}

	// Find null terminator or use the whole length
	end := offset
	for end < offset+fixedLength && data[end] != 0 {
		end++
	}

	return string(data[offset:end]), offset + fixedLength, nil
}
