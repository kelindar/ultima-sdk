// Package uofile provides a unified interface for accessing Ultima Online data files,
// supporting both MUL and UOP file formats.
package uofile

import (
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"sync"

	"github.com/kelindar/ultima-sdk/internal/mul"
	"github.com/kelindar/ultima-sdk/internal/uop"
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
	// Read reads data from a specific entry
	Read(index uint64) ([]byte, error)

	// Entries returns an iterator over available entries
	Entries() iter.Seq[uint64]

	// Close releases resources
	Close() error
}

// FormatType represents the file format type
type FormatType int

const (
	// FormatMUL represents the MUL file format
	FormatMUL FormatType = iota

	// FormatUOP represents the UOP file format
	FormatUOP
)

// File provides a unified interface for accessing both MUL and UOP files
type File struct {
	mu          sync.RWMutex
	reader      Reader
	format      FormatType
	path        string
	idxPath     string
	initialized bool
	closed      bool
	patches     map[uint64][]byte // Map of patches applied to file entries
	lazyInit    func() error      // Function for lazy initialization
}

// FileOption is a function that configures a File instance
type FileOption func(*File)

// WithPatches adds a map of patches to be applied to the file entries
func WithPatches(patches map[uint64][]byte) FileOption {
	return func(f *File) {
		f.patches = patches
	}
}

// WithMUL configures the file to use MUL format with the given paths
func WithMUL(mulPath, idxPath string) FileOption {
	return func(f *File) {
		f.format = FormatMUL
		f.path = mulPath
		f.idxPath = idxPath
		f.lazyInit = func() error {
			reader, err := mul.OpenWithIndex(mulPath, idxPath)
			if err != nil {
				return fmt.Errorf("failed to create MUL reader: %w", err)
			}
			f.reader = reader
			return nil
		}
	}
}

// WithUOP configures the file to use UOP format with the given path
func WithUOP(uopPath string, length int, options ...uop.Option) FileOption {
	return func(f *File) {
		f.format = FormatUOP
		f.path = uopPath
		f.lazyInit = func() error {
			reader, err := uop.Open(uopPath, length, options...)
			if err != nil {
				return fmt.Errorf("failed to create UOP reader: %w", err)
			}
			f.reader = reader
			return nil
		}
	}
}

// AutoDetect tries to automatically select the appropriate file format (MUL or UOP)
// based on file existence and naming conventions
func AutoDetect(basePath, baseName string, length int, uopOptions ...uop.Option) FileOption {
	// Try UOP format first (newer clients)
	uopPath := filepath.Join(basePath, baseName+".uop")
	if _, err := os.Stat(uopPath); err == nil {
		// UOP file exists
		return WithUOP(uopPath, length, uopOptions...)
	}

	// Fall back to MUL format
	mulPath := filepath.Join(basePath, baseName+".mul")
	idxPath := filepath.Join(basePath, baseName+"idx.mul")

	// For some files, the idx has a different naming pattern
	if _, err := os.Stat(idxPath); os.IsNotExist(err) {
		idxPath = filepath.Join(basePath, baseName+".idx")
	}

	return WithMUL(mulPath, idxPath)
}

// New creates a new File instance with the given options
func New(options ...FileOption) *File {
	f := &File{
		patches: make(map[uint64][]byte),
	}

	for _, option := range options {
		option(f)
	}

	return f
}

// ensureInitialized initializes the reader if it hasn't been already
func (f *File) ensureInitialized() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.initialized {
		return nil
	}

	if f.lazyInit == nil {
		return errors.New("no initialization function provided")
	}

	if err := f.lazyInit(); err != nil {
		return err
	}

	f.initialized = true
	return nil
}

// Read reads data from a specific entry, applying any patches if available
func (f *File) Read(index uint64) ([]byte, error) {
	// Check if we have a patch first
	f.mu.RLock()
	patch, hasPatch := f.patches[index]
	f.mu.RUnlock()

	if hasPatch {
		return patch, nil
	}

	// Ensure reader is initialized
	if err := f.ensureInitialized(); err != nil {
		return nil, err
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return nil, ErrReaderClosed
	}

	return f.reader.Read(index)
}

// Entries returns an iterator over available entries, including patched entries
func (f *File) Entries() iter.Seq[uint64] {
	return func(yield func(uint64) bool) {
		// Ensure reader is initialized
		if err := f.ensureInitialized(); err != nil {
			return
		}

		f.mu.RLock()
		defer f.mu.RUnlock()

		if f.closed {
			return
		}

		// First collect all entries from the underlying reader
		entries := make(map[uint64]struct{})

		// Add entries from reader
		f.reader.Entries()(func(idx uint64) bool {
			entries[idx] = struct{}{}
			return true
		})

		// Add entries from patches
		for idx := range f.patches {
			entries[idx] = struct{}{}
		}

		// Yield all unique entries
		for idx := range entries {
			if !yield(idx) {
				return
			}
		}
	}
}

// Close releases all resources associated with the file
func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}

	f.closed = true
	f.patches = nil

	if f.reader != nil {
		return f.reader.Close()
	}

	return nil
}

// AddPatch adds or updates a patch for a specific entry
func (f *File) AddPatch(index uint64, data []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Make a copy of the data to avoid external modifications
	patch := make([]byte, len(data))
	copy(patch, data)

	f.patches[index] = patch
}

// RemovePatch removes a patch for a specific entry if it exists
func (f *File) RemovePatch(index uint64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.patches, index)
}

// Format returns the file format (MUL or UOP)
func (f *File) Format() FormatType {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.format
}

// Path returns the file path
func (f *File) Path() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.path
}

// IndexPath returns the index file path for MUL files
func (f *File) IndexPath() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.idxPath
}

// Open tries to open and initialize the file reader
func (f *File) Open() error {
	return f.ensureInitialized()
}
