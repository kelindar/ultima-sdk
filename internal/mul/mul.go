// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package mul

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"

	"codeberg.org/go-mmap/mmap"
	"github.com/kelindar/intmap"
)

// Entry3D represents an entry in MUL index files
type Entry3D struct {
	offset  uint32 // Offset where the entry data begins
	length  uint32 // Size of the entry data
	extra   uint32 // Extra data (can be split into Extra1/Extra2)
	decoded []byte // Decoded entry data
}

// Reader provides access to MUL file data
type Reader struct {
	file      *mmap.File  // File handle for the MUL file
	index     *mmap.File  // Optional index file handle
	entries   []Entry3D   // Cached index entries
	lookup    *intmap.Map // Lookup table for entry offsets
	entrySize int         // Size of each entry in the index file
	closed    bool        // Flag to track if reader is closed
}

// Errors
var (
	ErrReaderClosed  = errors.New("mul reader is closed")
	ErrOutOfBounds   = errors.New("read operation would exceed file bounds")
	ErrInvalidIndex  = errors.New("invalid index")
	ErrInvalidOffset = errors.New("invalid offset")
	ErrInvalidEntry  = errors.New("invalid entry")
)

// OpenOne creates and initializes a new MUL reader
func OpenOne(filename string, options ...Option) (*Reader, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	// Open the file
	file, err := mmap.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open MUL file: %w", err)
	}

	r := &Reader{
		file:      file,
		lookup:    intmap.New(8000, .95),
		entrySize: 12,
	}

	// Apply options
	for _, option := range options {
		option(r)
	}

	// If no index file is provided, we need to create a default entry
	if len(r.entries) == 0 {
		buffer := make([]byte, info.Size())
		if _, err := file.ReadAt(buffer, 0); err != nil {
			return nil, err
		}

		r.add(0, 0, uint32(info.Size()), 0, buffer)
	}

	return r, nil
}

// Open creates a new MUL reader with a separate index file
func Open(mulFilename, idxFilename string, options ...Option) (*Reader, error) {
	file, err := mmap.Open(mulFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open MUL file: %w", err)
	}

	// Open IDX file
	idxFile, err := mmap.Open(idxFilename)
	if err != nil {
		file.Close() // Clean up MUL file handle if IDX file can't be opened
		return nil, fmt.Errorf("failed to open IDX file: %w", err)
	}

	r := &Reader{
		file:      file,
		index:     idxFile,
		lookup:    intmap.New(128, .95),
		entrySize: 12, // Default entry size is 12 bytes (3 uint32s)
	}

	// Apply options
	for _, option := range options {
		option(r)
	}

	// Cache index entries
	if err := r.loadIndex(); err != nil {
		r.Close() // Clean up both file handles if caching fails
		return nil, fmt.Errorf("failed to cache index entries: %w", err)
	}

	return r, nil
}

// loadIndex loads all index entries from the index file into memory
func (r *Reader) loadIndex() error {
	if r.index == nil {
		return errors.New("no index file provided")
	}

	info, err := r.index.Stat()
	if err != nil {
		return fmt.Errorf("failed to get index file stats: %w", err)
	}

	entryCount := int(info.Size()) / r.entrySize
	r.entries = make([]Entry3D, 0, entryCount)

	// Read all entries at once
	data := make([]byte, info.Size())
	_, err = r.index.ReadAt(data, 0)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to read index file: %w", err)
	}

	// Parse entries
	for i := 0; i < entryCount; i++ {
		offset := i * r.entrySize
		r.add(uint32(i),
			binary.LittleEndian.Uint32(data[offset:offset+4]),
			binary.LittleEndian.Uint32(data[offset+4:offset+8]),
			binary.LittleEndian.Uint32(data[offset+8:offset+12]),
			nil,
		)
	}

	return nil
}

// add creates a new entry and adds it to the reader
func (r *Reader) add(id, offset, length, extra uint32, value []byte) {
	index := uint32(len(r.entries))
	r.entries = append(r.entries, Entry3D{
		offset:  offset,
		length:  length,
		extra:   extra,
		decoded: value,
	})
	r.lookup.Store(id, index)
}

// Entry returns an entry reader
func (r *Reader) Entry(key uint32) (Entry, error) {
	entry, err := r.entryAt(key)
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

	if entry.decoded != nil {
		return reader{
			reader: bytes.NewReader(entry.decoded),
			entry:  entry,
		}, nil
	}

	return reader{
		reader: r.file,
		entry:  entry,
	}, nil
}

// entryAt retrieves entry information by its logical index/hash
func (r *Reader) entryAt(key uint32) (*Entry3D, error) {
	switch {
	case r.closed:
		return nil, ErrReaderClosed
	case r.entries == nil:
		return nil, ErrInvalidIndex
	default:
		index, ok := r.lookup.Load(key)
		if !ok {
			return nil, ErrInvalidIndex
		}

		return &r.entries[index], nil
	}
}

// Entries returns an iterator over available entries
func (r *Reader) Entries() iter.Seq[uint32] {
	return func(yield func(uint32) bool) {
		if r.closed {
			return
		}

		// Return entries from cache if available
		if r.entries != nil {
			for i, entry := range r.entries {
				if entry.offset == 0xFFFFFFFF || entry.length == 0 {
					continue // skip invalid entries
				}

				if !yield(uint32(i)) {
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

	if r.index != nil {
		if err := r.index.Close(); err != nil {
			errs = append(errs, err)
		}
		r.index = nil
	}

	r.entries = nil

	if len(errs) > 0 {
		return fmt.Errorf("failed to close files: %v", errs)
	}

	return nil
}

type Entry = interface {
	io.ReaderAt
	Len() int
	Extra() uint64
}

type reader struct {
	reader io.ReaderAt
	entry  *Entry3D
}

func (r reader) Len() int {
	return int(r.entry.length)
}

func (r reader) Extra() uint64 {
	return uint64(r.entry.extra)
}

func (r reader) ReadAt(p []byte, off int64) (n int, err error) {
	if r.entry.decoded != nil {
		return copy(p, r.entry.decoded), nil
	}

	return r.reader.ReadAt(p, int64(r.entry.offset)+off)
}
