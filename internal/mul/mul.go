// Package mul provides utilities for reading Ultima Online MUL files.
package mul

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"sync/atomic"
)

// Entry3D represents an entry in MUL index files
type Entry3D struct {
	offset uint32       // Offset where the entry data begins
	length uint32       // Size of the entry data
	extra  uint32       // Extra data (can be split into Extra1/Extra2)
	cache  atomic.Value // Cached data for the entry
}

// Reader provides access to MUL file data
type Reader struct {
	file       *os.File  // File handle for the MUL file
	idxFile    *os.File  // Optional index file handle
	entries    []Entry3D // Cached index entries
	entrySize  int       // Size of each entry in the index file
	entryCount int       // Number of entries per block (for structured files)
	chunkSize  int       // Size of a fixed chunk to divide the file (for files with fixed-size chunks)
	closed     bool      // Flag to track if reader is closed
}

// Errors
var (
	ErrReaderClosed  = errors.New("mul reader is closed")
	ErrOutOfBounds   = errors.New("read operation would exceed file bounds")
	ErrInvalidIndex  = errors.New("invalid index")
	ErrInvalidOffset = errors.New("invalid offset")
	ErrInvalidEntry  = errors.New("invalid entry")
)

// Option represents a configuration option for MulReader
type Option func(*Reader)

// WithChunkSize configures the reader to handle files with fixed-size chunks
// This is useful for files like hues.mul where data is stored in fixed-size blocks
func WithChunkSize(chunkSize int) Option {
	return func(r *Reader) {
		r.chunkSize = chunkSize
	}
}

// WithEntrySize sets the size of each entry in the index file
func WithEntrySize(size int) Option {
	return func(r *Reader) {
		r.entrySize = size
	}
}

// OpenOne creates and initializes a new MUL reader
func OpenOne(filename string, options ...Option) (*Reader, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open MUL file: %w", err)
	}

	reader := &Reader{
		file:      file,
		entrySize: 12,
	}

	// Apply options
	for _, option := range options {
		option(reader)
	}

	// Create virtual entries for chunked files
	switch {
	case reader.chunkSize > 0:
		if err := reader.createChunkEntries(); err != nil {
			reader.Close()
			return nil, err
		}
	default:
		reader.entries = []Entry3D{{
			offset: 0,
			length: uint32(info.Size()),
			extra:  0,
		}}
	}

	return reader, nil
}

// Open creates a new MUL reader with a separate index file
func Open(mulFilename, idxFilename string, options ...Option) (*Reader, error) {
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

	// Apply options
	for _, option := range options {
		option(reader)
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
	r.entries = make([]Entry3D, entryCount)

	// Read all entries at once
	data := make([]byte, fileSize)
	_, err = r.idxFile.ReadAt(data, 0)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to read index file: %w", err)
	}

	// Parse entries
	for i := 0; i < entryCount; i++ {
		offset := i * r.entrySize
		r.entries[i].offset = binary.LittleEndian.Uint32(data[offset : offset+4])
		r.entries[i].length = binary.LittleEndian.Uint32(data[offset+4 : offset+8])
		r.entries[i].extra = binary.LittleEndian.Uint32(data[offset+8 : offset+12])
	}

	return nil
}

// createChunkEntries divides the file into fixed-size chunks and creates virtual index entries
func (r *Reader) createChunkEntries() error {
	info, err := r.file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}

	fileSize := info.Size()
	if r.chunkSize <= 0 {
		return fmt.Errorf("invalid chunk size: %d", r.chunkSize)
	}

	chunkCount := int(fileSize / int64(r.chunkSize))
	if chunkCount == 0 {
		return fmt.Errorf("file too small for chunk format")
	}

	// Create a virtual index entry for each chunk
	r.entries = make([]Entry3D, chunkCount)
	for i := 0; i < chunkCount; i++ {
		r.entries[i] = Entry3D{
			offset: uint32(i * r.chunkSize),
			length: uint32(r.chunkSize),
			extra:  0,
		}
	}

	return nil
}

// Read reads data from the file at the specified index
func (r *Reader) Read(index uint64) (out []byte, err error) {
	entry, err := r.entryAt(index)
	switch {
	case err != nil:
		return nil, err
	case entry == nil:
		return nil, ErrInvalidEntry
	case entry.offset == 0xFFFFFFFF: // Skip invalid entries (offset == 0xFFFFFFFF or length == 0)
		return nil, nil
	case entry.length == 0:
		return nil, nil
	}

	// Check if the entry is cached
	if cached := entry.cache.Load(); cached != nil {
		return cached.([]byte), nil
	}

	// Read data from the file at the specified offset
	out = make([]byte, entry.length)
	err = r.ReadAt(out, index)

	// Write the data to the cache
	if err == nil {
		entry.cache.Store(out)
	}

	return out, err
}

// ReadAt reads data from the file at the specified index
func (r *Reader) ReadAt(p []byte, index uint64) error {
	entry, err := r.entryAt(index)
	switch {
	case err != nil:
		return err
	case entry == nil:
		return ErrInvalidEntry
	case entry.offset == 0xFFFFFFFF: // Skip invalid entries (offset == 0xFFFFFFFF or length == 0)
		return nil
	case entry.length == 0:
		return nil
	}

	// Read data from the file at the specified offset
	_, err = r.file.ReadAt(p, int64(entry.offset))
	if err != nil {
		return fmt.Errorf("failed to read data at index %d: %w", index, err)
	}

	// Check if the read data exceeds the entry length
	if len(p) > int(entry.length) {
		return fmt.Errorf("read data exceeds entry length: %d > %d", len(p), entry.length)
	}

	// If the entry length is less than the requested size, trim the slice
	if len(p) < int(entry.length) {
		p = p[:entry.length]
	}

	return nil
}

// entryAt retrieves entry information by its logical index/hash
func (r *Reader) entryAt(index uint64) (*Entry3D, error) {
	switch {
	case r.closed:
		return nil, ErrReaderClosed
	case r.entries == nil || int(index) < 0 || int(index) >= len(r.entries):
		return nil, ErrInvalidIndex
	default:
		return &r.entries[index], nil
	}
}

// Entries returns an iterator over available entries
func (r *Reader) Entries() iter.Seq[uint64] {
	return func(yield func(uint64) bool) {
		if r.closed {
			return
		}

		// Return entries from cache if available
		if r.entries != nil {
			for i, entry := range r.entries {
				if entry.offset == 0xFFFFFFFF || entry.length == 0 {
					continue // skip invalid entries
				}

				if !yield(uint64(i)) {
					return
				}
			}
		}
		// If no index file, we don't have entries to iterate over
	}
}

// Close releases resources
func (r *Reader) Close() error {
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

	r.entries = nil

	if len(errs) > 0 {
		return fmt.Errorf("failed to close files: %v", errs)
	}

	return nil
}

// ------------------------------- Reader Helper Functions ------------------------------- //

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
