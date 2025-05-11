// Package uofile provides a unified interface for accessing Ultima Online data files,
// supporting both MUL and UOP file formats.
package uofile

import (
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"
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
		if patches != nil {
			f.patches = patches
		}
	}
}

// WithUOPOptions passes UOP-specific options if UOP format is detected
func WithUOPOptions(options ...uop.Option) FileOption {
	return func(f *File) {
		// This option only applies if the format is detected as UOP
		if f.format == FormatUOP && f.lazyInit != nil {
			// Store the original initialization function
			original := f.lazyInit

			// Replace it with a new one that includes the UOP options
			f.lazyInit = func() error {
				// Extract path and length from the original function
				reader, err := uop.Open(f.path, 0, options...) // length will be updated by the specific options
				if err != nil {
					return fmt.Errorf("failed to create UOP reader with options: %w", err)
				}
				f.reader = reader
				return nil
			}
		}
	}
}

// WithCount sets the entry count for the file
func WithCount(count int) FileOption {
	return func(f *File) {
		if f.format == FormatUOP && f.lazyInit != nil {
			// Store the original initialization function
			original := f.lazyInit

			// Replace it with a new one that includes the count option
			f.lazyInit = func() error {
				// We need to extract the path from the file structure
				reader, err := uop.Open(f.path, count)
				if err != nil {
					return fmt.Errorf("failed to create UOP reader with count %d: %w", count, err)
				}
				f.reader = reader
				return nil
			}
		}
	}
}

// WithIndexLength sets the index length for UOP files
func WithIndexLength(length int) FileOption {
	return func(f *File) {
		if f.format == FormatUOP && f.lazyInit != nil {
			// Store the original initialization function
			original := f.lazyInit

			// Replace it with a new one that includes the index length option
			f.lazyInit = func() error {
				// We need to extract the path and other options
				reader, err := uop.Open(f.path, 0, uop.WithIndexLength(length))
				if err != nil {
					return fmt.Errorf("failed to create UOP reader with index length %d: %w", length, err)
				}
				f.reader = reader
				return nil
			}
		}
	}
}

// WithExtra sets a flag to indicate if extra data is present in UOP files
func WithExtra() FileOption {
	return func(f *File) {
		if f.format == FormatUOP && f.lazyInit != nil {
			// Store the original initialization function
			original := f.lazyInit

			// Replace it with a new one that includes the extra option
			f.lazyInit = func() error {
				// We need to extract the path and other options
				reader, err := uop.Open(f.path, 0, uop.WithExtra())
				if err != nil {
					return fmt.Errorf("failed to create UOP reader with extra data: %w", err)
				}
				f.reader = reader
				return nil
			}
		}
	}
}

// WithMUL configures the file to use MUL format with the given paths
func WithMUL(mulPath, idxPath string) FileOption {
	return func(f *File) {
		f.format = FormatMUL
		f.path = mulPath
		f.idxPath = idxPath
		f.lazyInit = func() error {
			reader, err := mul.Open(mulPath, idxPath)
			if err != nil {
				return fmt.Errorf("failed to create MUL reader: %w", err)
			}
			f.reader = reader
			return nil
		}
	}
}

// New creates a new File instance with automatic format detection
// It takes a base path, file names to check for, and options
func New(basePath string, fileNames []string, length int, options ...FileOption) *File {
	f := &File{
		patches: make(map[uint64][]byte),
	}

	// Try to detect the format and set up the appropriate reader
	detectFormat(f, basePath, fileNames, length)

	// Apply any additional options
	for _, option := range options {
		option(f)
	}

	return f
}

// detectFormat tries to determine the file format based on the file names
// and sets up the appropriate reader in the File struct
func detectFormat(f *File, basePath string, fileNames []string, length int) {
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
		f.format = FormatUOP
		f.path = uopPath
		f.lazyInit = func() error {
			reader, err := uop.Open(uopPath, length)
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
		f.format = FormatMUL
		f.path = mulPath
		f.idxPath = idxPath
		f.lazyInit = func() error {
			reader, err := mul.Open(mulPath, idxPath)
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
		f.format = FormatMUL
		f.path = mulPath
		f.lazyInit = func() error {
			reader, err := mul.OpenOne(mulPath)
			if err != nil {
				return fmt.Errorf("failed to create MUL reader without index: %w", err)
			}
			f.reader = reader
			return nil
		}
		return
	}

	// If we couldn't find valid files, set up a default that will fail when used
	f.format = FormatMUL                           // Default format
	f.path = filepath.Join(basePath, fileNames[0]) // Use first filename as a placeholder
	f.lazyInit = func() error {
		return fmt.Errorf("could not find valid files among %v in %s", fileNames, basePath)
	}
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
