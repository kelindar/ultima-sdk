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

// CompressionFlag represents the compression method used for a UOP entry
type CompressionFlag int16

// Compression flag constants
const (
	CompressionNone   CompressionFlag = 0
	CompressionZlib   CompressionFlag = 1
	CompressionMythic CompressionFlag = 2
)

// Entry interface defines methods to access entry information in a UOP file
type Entry interface {
	// Lookup returns the offset in the file where the entry data begins
	Lookup() int

	// Length returns the size of the entry data
	Length() int

	// Extra returns additional data associated with the entry (extra1, extra2)
	Extra() (int, int)

	// Zip returns the size after decompression and compression flag (0=none, 1=zlib, 2=mythic)
	Zip() (int, byte)
}

// Entry6D represents an entry in UOP files with 6 components including compression info
type Entry6D struct {
	LookupOffset       uint32 // Offset where the entry data begins
	DataLength         uint32 // Size of the entry data (compressed)
	DecompressedLength uint32 // Size after decompression
	ExtraInfo          uint32 // Extra data (can be split into Extra1/Extra2)
	CompressionType    byte   // Compression flag (0 = none, 1 = zlib, 2 = mythic)
}

// Implementation of the Entry interface for Entry6D
func (e *Entry6D) Lookup() int {
	return int(e.LookupOffset)
}

func (e *Entry6D) Length() int {
	return int(e.DataLength)
}

func (e *Entry6D) Extra() (int, int) {
	// For UOP entries, ExtraInfo is typically treated as a single value
	return int(e.ExtraInfo), 0
}

func (e *Entry6D) Zip() (int, byte) {
	return int(e.DecompressedLength), e.CompressionType
}

// ToEntry3D converts an Entry6D to a simplified Entry3D format
func (e *Entry6D) ToEntry3D() [3]uint32 {
	return [3]uint32{
		e.LookupOffset,
		e.DataLength,
		e.ExtraInfo,
	}
}

// FileEntry represents a single entry in a UOP file's structure
type FileEntry struct {
	Offset     int64  // Offset where the entry data begins in the file
	HeaderSize int32  // Size of the entry header
	Size       int32  // Size of the entry data (compressed)
	SizeDecomp int32  // Size of the entry data when decompressed
	Hash       uint64 // Hash of the entry filename
	Adler32    uint32 // Adler32 checksum
	Flag       int16  // Compression flag (0 = none, 1 = zlib)
}

// Reader implements the interface for reading UOP files
type Reader struct {
	file        *os.File              // File handle
	entries     map[uint64]*Entry6D   // Map of entries by logical index or hash
	fileEntries map[uint64]*FileEntry // Map of raw file entries by hash from the file
	mu          sync.RWMutex          // Mutex for thread safety
	closed      bool                  // Flag to track if reader is closed
}

// NewReader creates a new UOP file reader
func NewReader(filename string) (*Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open UOP file: %w", err)
	}

	r := &Reader{
		file:        file,
		entries:     make(map[uint64]*Entry6D),
		fileEntries: make(map[uint64]*FileEntry),
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
	// Read and verify the file header
	header := make([]byte, 28)
	if _, err := r.file.ReadAt(header, 0); err != nil {
		return fmt.Errorf("failed to read UOP header: %w", err)
	}

	// Check magic number
	magic := binary.LittleEndian.Uint32(header[0:4])
	if magic != uopMagic {
		return ErrInvalidFormat
	}

	// Read header values
	// version := binary.LittleEndian.Uint32(header[4:8])
	// signature := binary.LittleEndian.Uint32(header[8:12])
	nextBlock := int64(binary.LittleEndian.Uint64(header[12:20]))
	// blockCapacity := binary.LittleEndian.Uint32(header[20:24])
	// entryCount := binary.LittleEndian.Uint32(header[24:28])

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
			offset := i * entrySize
			fileEntry := &FileEntry{
				Offset:     int64(binary.LittleEndian.Uint64(entryData[offset : offset+8])),
				HeaderSize: int32(binary.LittleEndian.Uint32(entryData[offset+8 : offset+12])),
				Size:       int32(binary.LittleEndian.Uint32(entryData[offset+12 : offset+16])),
				SizeDecomp: int32(binary.LittleEndian.Uint32(entryData[offset+16 : offset+20])),
				Hash:       binary.LittleEndian.Uint64(entryData[offset+20 : offset+28]),
				Adler32:    binary.LittleEndian.Uint32(entryData[offset+28 : offset+32]),
				Flag:       int16(binary.LittleEndian.Uint16(entryData[offset+32 : offset+34])),
			}

			// Skip entries with offset 0 (they're placeholders)
			if fileEntry.Offset == 0 {
				continue
			}

			// Store the entry in fileEntries map
			r.fileEntries[fileEntry.Hash] = fileEntry

			// Create and store the corresponding Entry6D
			entry6D := &Entry6D{
				LookupOffset:       uint32(fileEntry.Offset + int64(fileEntry.HeaderSize)),
				DataLength:         uint32(fileEntry.Size),
				DecompressedLength: uint32(fileEntry.SizeDecomp),
				ExtraInfo:          0, // Could be used for extra info if needed
				CompressionType:    byte(fileEntry.Flag),
			}
			r.entries[fileEntry.Hash] = entry6D
		}

		// Move to next block
		nextBlock = nextBlockOffset
	}

	return nil
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
	r.fileEntries = nil
	return r.file.Close()
}

// EntryAt retrieves entry information by its hash
func (r *Reader) EntryAt(hash uint64) (Entry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, ErrReaderClosed
	}

	entry, exists := r.entries[hash]
	if !exists {
		return nil, ErrInvalidIndex
	}

	return entry, nil
}

// ReadAt reads data from a specific offset and length
func (r *Reader) ReadAt(offset int64, length int) ([]byte, error) {
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

// Read reads the data for a specific entry
func (r *Reader) Read(entry Entry) ([]byte, error) {
	if entry == nil {
		return nil, errors.New("nil entry")
	}

	return r.ReadAt(int64(entry.Lookup()), entry.Length())
}

// Entries returns an iterator over available entries
func (r *Reader) Entries() iter.Seq[Entry] {
	return func(yield func(Entry) bool) {
		r.mu.RLock()
		defer r.mu.RUnlock()

		if r.closed {
			return
		}

		for _, entry := range r.entries {
			if !yield(entry) {
				return
			}
		}
	}
}

// SetupEntries builds the logical index mapping based on a pattern
func (r *Reader) SetupEntries(pattern string, maxIndex int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrReaderClosed
	}

	// Clear existing entries
	r.entries = make(map[uint64]*Entry6D)

	// Generate file names based on pattern and calculate hashes
	for i := 0; i < maxIndex; i++ {
		filename := fmt.Sprintf("build/%s/%08d.dat", pattern, i)
		hash := HashFileName(filename)

		// Find entry with this hash
		if fileEntry, exists := r.fileEntries[hash]; exists {
			// Create a new Entry6D
			entry := &Entry6D{
				LookupOffset:       uint32(fileEntry.Offset + int64(fileEntry.HeaderSize)),
				DataLength:         uint32(fileEntry.Size),
				DecompressedLength: uint32(fileEntry.SizeDecomp),
				ExtraInfo:          uint32(i), // Store index as extra info
				CompressionType:    byte(fileEntry.Flag),
			}

			// Store by hash
			r.entries[hash] = entry
		}
	}

	return nil
}

// GetEntryByName retrieves an entry by its file path/name
func (r *Reader) GetEntryByName(filePath string) (Entry, error) {
	hash := HashFileName(filePath)
	return r.EntryAt(hash)
}

// HashFileName calculates a hash for a filename as used in UOP files
// This is a direct port of the C# algorithm
func HashFileName(s string) uint64 {
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

// BuildPatternName constructs the UOP naming pattern from a filename
func BuildPatternName(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	if ext != "" {
		return base[:len(base)-len(ext)]
	}
	return base
}
