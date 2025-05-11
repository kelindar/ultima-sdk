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
	"sync"
)

// Magic number for UOP file format - "MYP\0" in ASCII
const uopMagic = 0x50594D

// Standard UOP format errors
var (
	ErrInvalidFormat = errors.New("invalid UOP file format")
	ErrInvalidIndex  = errors.New("invalid index")
	ErrReaderClosed  = errors.New("uop reader is closed")
	ErrEntryNotFound = errors.New("entry not found")
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
	offset uint32    // Offset where the entry data begins
	length uint32    // Size of the entry data (compressed)
	rawLen uint32    // Size after decompression
	extra  [2]uint32 // Extra data
	typ    byte      // Compression flag (0 = none, 1 = zlib, 2 = mythic)
}

// Reader implements the interface for reading UOP files
type Reader struct {
	file      *os.File     // File handle
	entries   []Entry6D    // Map of entries by logical index or hash
	mu        sync.RWMutex // Mutex for thread safety
	length    int          // Length of the file
	idxLength int          // Length of the index
	ext       string       // File extension
	closed    bool         // Flag to track if reader is closed
	hasextra  bool         // Flag to indicate if extra data is present
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

// NewReader creates a new UOP file reader
func NewReader(filename string, length int, options ...Option) (*Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open UOP file: %w", err)
	}

	r := &Reader{
		file:    file,
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
	uopPattern := strings.ToLower(strings.ReplaceAll(filepath.Base(r.file.Name()), filepath.Ext(r.file.Name()), ""))

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

// Read reads data from a specific index
func (r *Reader) Read(index uint64) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch {
	case index >= uint64(len(r.entries)):
		return nil, ErrInvalidIndex
	case r.closed:
		return nil, ErrReaderClosed
	}

	entry := &r.entries[index]
	if entry.length == 0 {
		return nil, ErrEntryNotFound
	}

	// Read the raw data
	rawData, err := r.readAt(int64(entry.offset), int(entry.length))
	if err != nil {
		return nil, err
	}

	// Decompress the data
	return decode(rawData, CompressionType(entry.typ))
}

// Entries returns an iterator over available entry indices
func (r *Reader) Entries() iter.Seq[uint64] {
	return func(yield func(uint64) bool) {
		r.mu.RLock()
		defer r.mu.RUnlock()

		if r.closed {
			return
		}

		for i, entry := range r.entries {
			if entry.offset == 0xFFFFFFFF || entry.length == 0 {
				continue // skip invalid entries
			}

			if !yield(uint64(i)) {
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
	r.entries = nil
	return r.file.Close()
}

// readAt reads data from a specific offset and length
func (r *Reader) readAt(offset int64, length int) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, ErrReaderClosed
	}

	// Check if the offset is valid
	fileInfo, err := r.file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	fileSize := fileInfo.Size()
	if offset < 0 || offset >= fileSize {
		return nil, fmt.Errorf("offset %d is out of range for file size %d", offset, fileSize)
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

// hashFileName calculates a hash for a filename as used in UOP files
// This is a direct port of the C# algorithm
func hashFileName(s string) uint64 {
	var eax, ecx, edx, ebx, esi, edi uint32

	eax = 0
	ecx = 0
	edx = 0
	ebx = uint32(len(s)) + 0xDEADBEEF
	esi = ebx
	edi = ebx

	i := 0
	for i+12 <= len(s) {
		edi += (uint32(s[i+7]) << 24) | (uint32(s[i+6]) << 16) | (uint32(s[i+5]) << 8) | uint32(s[i+4])
		esi += (uint32(s[i+11]) << 24) | (uint32(s[i+10]) << 16) | (uint32(s[i+9]) << 8) | uint32(s[i+8])
		edx = (uint32(s[i+3])<<24 | uint32(s[i+2])<<16 | uint32(s[i+1])<<8 | uint32(s[i])) - esi

		edx = (edx + ebx) ^ (esi >> 28) ^ (esi << 4)
		esi += edi
		edi = (edi - edx) ^ (edx >> 26) ^ (edx << 6)
		edx += esi
		esi = (esi - edi) ^ (edi >> 24) ^ (edi << 8)
		edi += edx
		ebx = (edx - esi) ^ (esi >> 16) ^ (esi << 16)
		esi += edi
		edi = (edi - ebx) ^ (ebx >> 13) ^ (ebx << 19)
		ebx += esi
		esi = (esi - edi) ^ (edi >> 28) ^ (edi << 4)
		edi += ebx

		i += 12
	}

	if len(s)-i > 0 {
		remLen := len(s) - i

		// Process remaining bytes
		switch remLen {
		case 12:
			esi += uint32(s[i+11]) << 24
			fallthrough
		case 11:
			esi += uint32(s[i+10]) << 16
			fallthrough
		case 10:
			esi += uint32(s[i+9]) << 8
			fallthrough
		case 9:
			esi += uint32(s[i+8])
			fallthrough
		case 8:
			edi += uint32(s[i+7]) << 24
			fallthrough
		case 7:
			edi += uint32(s[i+6]) << 16
			fallthrough
		case 6:
			edi += uint32(s[i+5]) << 8
			fallthrough
		case 5:
			edi += uint32(s[i+4])
			fallthrough
		case 4:
			ebx += uint32(s[i+3]) << 24
			fallthrough
		case 3:
			ebx += uint32(s[i+2]) << 16
			fallthrough
		case 2:
			ebx += uint32(s[i+1]) << 8
			fallthrough
		case 1:
			ebx += uint32(s[i])
		}

		esi = (esi ^ edi) - ((edi >> 18) ^ (edi << 14))
		ecx = (esi ^ ebx) - ((esi >> 21) ^ (esi << 11))
		edi = (edi ^ ecx) - ((ecx >> 7) ^ (ecx << 25))
		esi = (esi ^ edi) - ((edi >> 16) ^ (edi << 16))
		edx = (esi ^ ecx) - ((esi >> 28) ^ (esi << 4))
		edi = (edi ^ edx) - ((edx >> 18) ^ (edx << 14))
		eax = (esi ^ edi) - ((edi >> 8) ^ (edi << 24))

		return (uint64(edi) << 32) | uint64(eax)
	}

	return (uint64(esi) << 32) | uint64(eax)
}
