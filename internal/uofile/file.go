// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package uofile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"codeberg.org/go-mmap/mmap"
	"github.com/kelindar/ultima-sdk/internal/mul"
	"github.com/kelindar/ultima-sdk/internal/uop"
)

// File state constants
const (
	stateNew    int32 = 0 // Initial state
	stateReady  int32 = 1 // Initialized and ready
	stateClosed int32 = 2 // Closed
)

// Common errors
var (
	ErrInvalidIndex  = errors.New("invalid index")
	ErrReaderClosed  = errors.New("file reader is closed")
	ErrEntryNotFound = errors.New("entry not found")
	ErrInvalidFormat = errors.New("invalid file format")
)

// Reader defines the common interface for both MUL and UOP readers
type Reader interface {
	Read(*bytes.Buffer, uint32) ([]byte, uint64, error)
	Entry(key uint32) (Entry, error)
	Entries() iter.Seq[uint32]
	Close() error
}

// Entry defines the common interface for both MUL and UOP entries
type Entry = interface {
	io.ReaderAt
	Len() int
	Extra() uint64
}

// File provides a unified interface for accessing both MUL and UOP files
type File struct {
	reader  Reader
	path    string
	base    string
	idxPath string
	initFn  func() error // Function for lazy initialization
	state   atomic.Int32 // File state (new, ready, closed)
	uopOpts []uop.Option // Options specific to UOP files
	mulOpts []mul.Option // Options specific to MUL files
	length  int          // Length parameter for the file
}

// Option is a function that configures a File instance
type Option func(*File)

// WithCount sets the entry count for UOP files
func WithCount(count int) Option {
	return func(f *File) {
		f.length = count
	}
}

// WithIndexLength sets the index length for UOP files
func WithIndexLength(length int) Option {
	return func(f *File) {
		f.uopOpts = append(f.uopOpts, uop.WithIndexLength(length))
	}
}

// WithExtension sets the file extension for UOP files
func WithExtension(ext string) Option {
	return func(f *File) {
		f.uopOpts = append(f.uopOpts, uop.WithExtension(ext))
	}
}

// WithExtra sets a flag to indicate if extra data is present in UOP files
func WithExtra() Option {
	return func(f *File) {
		f.uopOpts = append(f.uopOpts, uop.WithExtra())
	}
}

// WithChunks configures the reader to handle files with fixed-size chunks
// This is useful for files like hues.mul where data is stored in fixed-size blocks
func WithChunks(chunkSize int) Option {
	return func(f *File) {
		f.mulOpts = append(f.mulOpts, mul.WithChunks(chunkSize))
	}
}

// WithDecodeMUL sets a custom function to read from a MUL file
func WithDecodeMUL(fn func(file *mmap.File, add mul.AddFn) error) Option {
	return func(f *File) {
		f.mulOpts = append(f.mulOpts, mul.WithDecode(fn))
	}
}

// WithStrict sets a flag to indicate if the reader should perform strict entry validation.
func WithStrict() Option {
	return func(f *File) {
		f.uopOpts = append(f.uopOpts, uop.WithStrict())
	}
}

// New creates a new File instance with automatic format detection
// It takes a base path, file names to check for, and options
func New(basePath string, fileNames []string, length int, options ...Option) *File {
	f := &File{
		length: length,
		base:   basePath,
	}

	// Try to detect the format and set up the appropriate reader
	detectFormat(f, basePath, fileNames)

	// Apply any additional options
	for _, option := range options {
		option(f)
	}

	return f
}

// detectFormat tries to determine the file format based on the file names
// and sets up the appropriate reader in the File struct
func detectFormat(f *File, basePath string, fileNames []string) {
	useOne := func(path string) {
		f.path = path
		f.initFn = func() error {
			reader, err := mul.OpenOne(path, f.mulOpts...)
			if err != nil {
				return fmt.Errorf("failed to create reader for %s: %w", path, err)
			}
			f.reader = reader
			return nil
		}
	}

	// 1. Special case for cliloc files (cliloc.*)
	for _, fileName := range fileNames {
		if strings.HasPrefix(fileName, "cliloc.") {
			if path, ok := f.fileExists(fileName); ok {
				useOne(path)
				return
			}
		}
	}

	// 2. Look for UOP files first (preferred format)
	for _, fileName := range fileNames {
		if strings.HasSuffix(fileName, ".uop") {
			if path, ok := f.fileExists(fileName); ok {
				f.path = path
				f.initFn = func() error {
					reader, err := uop.Open(path, f.length, f.uopOpts...)
					if err != nil {
						return fmt.Errorf("failed to create UOP reader: %w", err)
					}
					f.reader = reader
					return nil
				}
				return
			}
		}
	}

	// 3. Look for MUL and IDX files
	var mulPath, idxPath string
	for _, fileName := range fileNames {
		if path, ok := f.fileExists(fileName); ok {
			switch {
			case strings.HasPrefix(fileName, "staidx") || strings.HasSuffix(fileName, "idx.mul") || strings.HasSuffix(fileName, ".idx"):
				idxPath = path
			case strings.HasSuffix(fileName, ".mul") && !strings.HasSuffix(fileName, "idx.mul"):
				mulPath = path
			}
		}
	}

	// If we found both MUL and IDX, use them
	if mulPath != "" && idxPath != "" {
		f.path = mulPath
		f.idxPath = idxPath
		f.initFn = func() error {
			reader, err := mul.Open(mulPath, idxPath, f.mulOpts...)
			if err != nil {
				return fmt.Errorf("failed to create MUL reader: %w", err)
			}
			f.reader = reader
			return nil
		}
		return
	}

	// 4. If we only found a MUL file (no index), use just that
	if mulPath != "" {
		useOne(mulPath)
		return
	}

	// 5. No valid files found, set up a default error handler
	f.path = filepath.Join(basePath, fileNames[0]) // Use first filename as placeholder
	f.initFn = func() error {
		return fmt.Errorf("could not find valid files among %v in %s", fileNames, basePath)
	}
}

// open initializes the reader if it hasn't been already
func (f *File) open() error {
	switch {
	case f.state.Load() == stateReady:
		return nil
	case f.state.Load() == stateClosed:
		return ErrReaderClosed
	case f.initFn == nil:
		return fmt.Errorf("file %s is not initialized", f.path)
	}

	// Try to transition from new to ready
	if f.state.CompareAndSwap(stateNew, stateReady) {
		if err := f.initFn(); err != nil {
			f.state.Store(stateNew)
			return fmt.Errorf("failed to initialize file %s: %w", f.path, err)
		}
	} else {
		runtime.Gosched()
	}

	return nil
}

// Entry returns a specific entry
func (f *File) Entry(key uint32) (Entry, error) {
	if err := f.open(); err != nil {
		return nil, err
	}

	// Double-check reader is not nil after initialization
	if f.reader == nil {
		return nil, fmt.Errorf("reader not initialized for %s", f.path)
	}

	return f.reader.Entry(key)
}

// ReadFull reads the full entry data into a byte slice
func (f *File) ReadFull(key uint32) ([]byte, error) {
	entry, err := f.Entry(key)
	switch {
	case err != nil:
		return nil, err
	case entry == nil:
		return nil, nil
	}

	data := make([]byte, entry.Len())
	if _, err := entry.ReadAt(data, 0); err != nil {
		return nil, err
	}

	return data, nil
}

// Entries returns a sequence of entry indices
func (f *File) Entries() iter.Seq[uint32] {
	if err := f.open(); err != nil {
		panic(err)
	}

	return f.reader.Entries()
}

// Close releases all resources associated with the file
func (f *File) Close() error {
	if prevState := f.state.Swap(stateClosed); prevState == stateClosed {
		return nil
	}

	if f.reader != nil {
		return f.reader.Close()
	}
	return nil
}

func (f *File) fileExists(fileName string) (string, bool) {
	filePath := filepath.Join(f.base, fileName)
	if _, err := os.Stat(filePath); err == nil {
		return filePath, true
	}
	return "", false
}

var bufferPool sync.Pool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 1024))
	},
}

// Borrow allocates a buffer of the specified size from the pool
func Borrow(n int) ([]byte, context.CancelFunc) {
	buffer := bufferPool.Get().(*bytes.Buffer)
	buffer.Grow(n)
	return buffer.Bytes()[:n], func() {
		buffer.Reset()
		bufferPool.Put(buffer)
	}
}

func Decode[T any](f *File, key uint32, fn func([]byte, uint64) (T, error)) (T, error) {
	entry, err := f.Entry(uint32(key))
	switch {
	case err != nil:
		return defaultT[T](), err
	case entry == nil || entry.Len() == 0:
		return defaultT[T](), nil
	}

	data, release := Borrow(entry.Len())
	defer release()

	if _, err := entry.ReadAt(data, 0); err != nil {
		return defaultT[T](), err
	}

	return fn(data, entry.Extra())
}

func defaultT[T any]() T {
	var t T
	return t
}
