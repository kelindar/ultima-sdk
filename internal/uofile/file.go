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
	Read(index uint64) ([]byte, error)
	Entries() iter.Seq[uint64]
	Close() error
}

// File provides a unified interface for accessing both MUL and UOP files
type File struct {
	mu          sync.RWMutex
	reader      Reader
	path        string
	idxPath     string
	initialized bool
	closed      bool
	opts        []uop.Option // Options specific to UOP files
	length      int          // Length parameter for the file
	init        func() error // Function for lazy initialization
}

// FileOption is a function that configures a File instance
type FileOption func(*File)

// WithCount sets the entry count for UOP files
func WithCount(count int) FileOption {
	return func(f *File) {
		f.length = count
	}
}

// WithIndexLength sets the index length for UOP files
func WithIndexLength(length int) FileOption {
	return func(f *File) {
		f.opts = append(f.opts, uop.WithIndexLength(length))
	}
}

// WithExtra sets a flag to indicate if extra data is present in UOP files
func WithExtra() FileOption {
	return func(f *File) {
		f.opts = append(f.opts, uop.WithExtra())
	}
}

// New creates a new File instance with automatic format detection
// It takes a base path, file names to check for, and options
func New(basePath string, fileNames []string, length int, options ...FileOption) *File {
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
		f.init = func() error {
			reader, err := uop.Open(uopPath, f.length, f.opts...)
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
		f.init = func() error {
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
		f.path = mulPath
		f.init = func() error {
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
	f.path = filepath.Join(basePath, fileNames[0]) // Use first filename as a placeholder
	f.init = func() error {
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

	if f.init == nil {
		return errors.New("no initialization function provided")
	}

	if err := f.init(); err != nil {
		return err
	}

	f.initialized = true
	return nil
}

// Read reads data from a specific entry, applying any patches if available
func (f *File) Read(index uint64) ([]byte, error) {
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

// Entries returns a sequence of entry indices
func (f *File) Entries() iter.Seq[uint64] {
	if err := f.ensureInitialized(); err != nil {
		panic(err)
	}

	return f.Entries()
}

// Close releases all resources associated with the file
func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}

	f.closed = true

	if f.reader != nil {
		return f.reader.Close()
	}

	return nil
}
