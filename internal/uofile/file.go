// Package uofile provides a unified interface for accessing Ultima Online data files,
// supporting both MUL and UOP file formats.
package uofile

import (
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"

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
	Read(index uint64) ([]byte, error)
	Entries() iter.Seq[uint64]
	Close() error
}

// File provides a unified interface for accessing both MUL and UOP files
type File struct {
	reader  Reader
	path    string
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

// WithExtra sets a flag to indicate if extra data is present in UOP files
func WithExtra() Option {
	return func(f *File) {
		f.uopOpts = append(f.uopOpts, uop.WithExtra())
	}
}

// WithChunkSize configures the reader to handle files with fixed-size chunks
// This is useful for files like hues.mul where data is stored in fixed-size blocks
func WithChunkSize(chunkSize int) Option {
	return func(f *File) {
		f.mulOpts = append(f.mulOpts, mul.WithChunkSize(chunkSize))
	}
}

// New creates a new File instance with automatic format detection
// It takes a base path, file names to check for, and options
func New(basePath string, fileNames []string, length int, options ...Option) *File {
	f := &File{
		length: length,
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
	var uopPath string
	var mulPath string
	var idxPath string

	// Look for UOP files first (preferred format)
	for _, fileName := range fileNames {
		if strings.HasSuffix(fileName, ".uop") {
			filePath := filepath.Join(basePath, fileName)
			if _, err := os.Stat(filePath); err == nil {
				uopPath = filePath
				break
			}
		}
	}

	// If UOP file was found, configure for UOP
	if uopPath != "" {
		f.path = uopPath
		f.initFn = func() error {
			reader, err := uop.Open(uopPath, f.length, f.uopOpts...)
			if err != nil {
				return fmt.Errorf("failed to create UOP reader: %w", err)
			}
			f.reader = reader
			return nil
		}
		return
	}

	// Otherwise look for MUL and IDX files
	for _, fileName := range fileNames {
		filePath := filepath.Join(basePath, fileName)

		if strings.HasSuffix(fileName, "idx.mul") || strings.HasSuffix(fileName, ".idx") {
			if _, err := os.Stat(filePath); err == nil {
				idxPath = filePath
			}
		} else if strings.HasSuffix(fileName, ".mul") && !strings.HasSuffix(fileName, "idx.mul") {
			if _, err := os.Stat(filePath); err == nil {
				mulPath = filePath
			}
		}
	}

	// If we found both needed MUL files, configure for MUL
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

	// If we only have a MUL file but no IDX, try to open it without an index
	if mulPath != "" {
		f.path = mulPath
		f.initFn = func() error {
			reader, err := mul.OpenOne(mulPath, f.mulOpts...)
			if err != nil {
				return fmt.Errorf("failed to create MUL reader without index: %w", err)
			}

			f.reader = reader
			return nil
		}
		return
	}

	// If we couldn't find valid files, set up a default that will fail when used
	f.path = filepath.Join(basePath, fileNames[0]) // Use first filename as a placeholder
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

// Read reads data from a specific entry, applying any patches if available
func (f *File) Read(index uint64) ([]byte, error) {
	if err := f.open(); err != nil {
		return nil, err
	}

	// Double-check reader is not nil after initialization
	if f.reader == nil {
		return nil, fmt.Errorf("reader not initialized for %s", f.path)
	}

	return f.reader.Read(index)
}

// Entries returns a sequence of entry indices
func (f *File) Entries() iter.Seq[uint64] {
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
