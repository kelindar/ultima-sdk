// Package uop provides utilities for reading Ultima Online UOP (Ultima Online Package) files.
package uop

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"codeberg.org/go-mmap/mmap"
)

// Magic number for UOP file format - "MYP\0" in ASCII
const uopMagic = 0x50594D

// Standard UOP format errors
var (
	ErrInvalidFormat = errors.New("invalid UOP file format")
	ErrInvalidIndex  = errors.New("invalid index")
	ErrReaderClosed  = errors.New("uop reader is closed")
	ErrEntryNotFound = errors.New("entry not found")
	ErrInvalidEntry  = errors.New("invalid entry")
)

// CompressionType represents the compression method used for a UOP entry
type CompressionType int16

// Compression flag constants
const (
	CompressionNone   CompressionType = 0
	CompressionZlib   CompressionType = 1
	CompressionMythic CompressionType = 2
)

// Entry6D represents an entry in UOP files with 6 components including compression info
type Entry6D struct {
	offset uint32       // Offset where the entry data begins
	length uint32       // Size of the entry data (compressed)
	rawLen uint32       // Size after decompression
	extra  [2]uint32    // Extra data
	typ    byte         // Compression flag (0 = none, 1 = zlib, 2 = mythic)
	cache  atomic.Value // Cached data for the entry
}

// Reader implements the interface for reading UOP files
type Reader struct {
	file      *mmap.File  // File handle
	info      os.FileInfo // File information
	entries   []Entry6D   // Map of entries by logical index or hash
	length    int         // Length of the file
	idxLength int         // Length of the index
	ext       string      // File extension
	closed    bool        // Flag to track if reader is closed
	hasextra  bool        // Flag to indicate if extra data is present
}

// Option defines a function that configures a Reader.
type Option func(*Reader)

// WithExtra sets a flag to indicate if extra data is present in the entries.
func WithExtra() Option {
	return func(r *Reader) {
		r.hasextra = true
	}
}

// WithExtension sets the file extension for the pattern.
func WithExtension(ext string) Option {
	return func(r *Reader) {
		r.ext = ext
	}
}

// WithIndexLength sets the length of the index.
func WithIndexLength(length int) Option {
	return func(r *Reader) {
		r.idxLength = length
	}
}

// Open creates a new UOP file reader
func Open(filename string, length int, options ...Option) (*Reader, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	file, err := mmap.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open UOP file: %w", err)
	}

	r := &Reader{
		file:    file,
		info:    info,
		entries: make([]Entry6D, length),
		ext:     ".dat",
		length:  length,
	}

	// Apply any provided options
	for _, option := range options {
		option(r)
	}

	// Parse the UOP file structure
	if err := r.parseFile(); err != nil {
		file.Close() // Close the file if parsing fails
		return nil, err
	}

	return r, nil
}

// parseFile reads the UOP file header and builds the entry tables
func (r *Reader) parseFile() error {
	uopPattern := strings.ToLower(strings.ReplaceAll(filepath.Base(r.info.Name()), filepath.Ext(r.info.Name()), ""))

	// Read and verify the file header
	header := make([]byte, 28)
	if _, err := r.file.ReadAt(header, 0); err != nil {
		return fmt.Errorf("failed to read UOP header: %w", err)
	}

	// Check magic number
	if magic := binary.LittleEndian.Uint32(header[0:4]); magic != uopMagic {
		return ErrInvalidFormat
	}

	// Read header values
	// version := binary.LittleEndian.Uint32(header[4:8])
	// signature := binary.LittleEndian.Uint32(header[8:12])
	nextBlock := int64(binary.LittleEndian.Uint64(header[12:20]))
	// blockCapacity := binary.LittleEndian.Uint32(header[20:24])
	// entryCount := binary.LittleEndian.Uint32(header[24:28])

	// Build the pattern name
	hashes := make(map[uint64]int, r.length)
	for i := 0; i < r.length; i++ {
		name := fmt.Sprintf("build/%s/%08d%s", uopPattern, i, r.ext)
		hash := hashFileName(name)
		hashes[hash] = i
	}

	// Prepare to read block structure
	for nextBlock != 0 {
		// Read block header (filesCount + nextBlock)
		blockHeader := make([]byte, 12)
		if _, err := r.file.ReadAt(blockHeader, nextBlock); err != nil {
			return fmt.Errorf("failed to read block header: %w", err)
		}

		// Get file count and next block
		fileCount := int(binary.LittleEndian.Uint32(blockHeader[0:4]))
		nextBlockOffset := int64(binary.LittleEndian.Uint64(blockHeader[4:12]))

		// Read file entries in this block
		entrySize := 34 // Each entry is 34 bytes
		entryData := make([]byte, fileCount*entrySize)
		if _, err := r.file.ReadAt(entryData, nextBlock+12); err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("failed to read file entries: %w", err)
		}

		// Parse each entry in the block
		for i := 0; i < fileCount; i++ {
			idx := i * entrySize

			offset := int64(binary.LittleEndian.Uint64(entryData[idx : idx+8]))
			headerSize := int32(binary.LittleEndian.Uint32(entryData[idx+8 : idx+12]))
			encodedSize := int32(binary.LittleEndian.Uint32(entryData[idx+12 : idx+16]))
			decodedSize := int32(binary.LittleEndian.Uint32(entryData[idx+16 : idx+20]))
			hash := binary.LittleEndian.Uint64(entryData[idx+20 : idx+28])
			_ = binary.LittleEndian.Uint32(entryData[idx+28 : idx+32]) // data_hash (unused)
			flag := int16(binary.LittleEndian.Uint16(entryData[idx+32 : idx+34]))

			// Skip entries with offset 0 (they're placeholders)
			if offset == 0 {
				continue
			}

			entryIdx, ok := hashes[hash]
			if !ok {
				continue
			}
			if entryIdx < 0 || entryIdx > r.idxLength {
				return fmt.Errorf("hashes dictionary and files collection have different count of entries!")
			}

			offset += int64(headerSize)

			if r.hasextra && flag != 0 {
				extra1 := binary.LittleEndian.Uint32(entryData[idx+34 : idx+38])
				extra2 := binary.LittleEndian.Uint32(entryData[idx+38 : idx+42])
				r.entries[entryIdx] = Entry6D{
					offset: uint32(offset + 8),
					length: uint32(encodedSize - 8),
					rawLen: uint32(decodedSize),
					extra:  [2]uint32{extra1, extra2},
					typ:    byte(flag),
				}

			} else {
				r.entries[entryIdx] = Entry6D{
					offset: uint32(offset),
					length: uint32(encodedSize),
					rawLen: uint32(decodedSize),
					extra:  [2]uint32{0x0FFFFFFF, 0x0FFFFFFF}, // we cant read it right now, but -1 and 0 makes this entry invalid
					typ:    byte(flag),
				}
			}
		}

		// Move to next block
		nextBlock = nextBlockOffset
	}

	return nil
}

// Entries returns an iterator over available entry indices
func (r *Reader) Entries() iter.Seq[uint32] {
	return func(yield func(uint32) bool) {
		if r.closed {
			return
		}

		for i, entry := range r.entries {
			if entry.offset == 0xFFFFFFFF || entry.length == 0 {
				continue // skip invalid entries
			}

			if !yield(uint32(i)) {
				return
			}
		}
	}
}

// Close releases resources
func (r *Reader) Close() error {
	if r.closed {
		return nil
	}

	r.closed = true
	r.entries = nil
	return r.file.Close()
}

// Read reads data from the file at the specified index
func (r *Reader) Read(index uint32) (out []byte, err error) {
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
	if _, err = r.file.ReadAt(out, int64(entry.offset)); err != nil {
		return nil, fmt.Errorf("failed to read data at index %d: %w", index, err)
	}

	// Decompress the data, write it to the cache if needed
	dec, err := decode(out, CompressionType(entry.typ))
	if err == nil {
		entry.cache.Store(out)
	}

	return dec, err
}

// entryAt retrieves entry information by its logical index/hash
func (r *Reader) entryAt(index uint32) (*Entry6D, error) {
	switch {
	case r.closed:
		return nil, ErrReaderClosed
	case r.entries == nil || int(index) < 0 || int(index) >= len(r.entries):
		return nil, ErrInvalidIndex
	default:
		return &r.entries[index], nil
	}
}
