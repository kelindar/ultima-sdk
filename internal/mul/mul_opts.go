package mul

import (
	"fmt"
	"os"
)

type AddFn = func(id, offset, length, extra uint32, value []byte)

// Option represents a configuration option for MulReader
type Option func(*Reader)

// WithEntrySize sets the size of each entry in the index file
func WithEntrySize(size int) Option {
	return func(r *Reader) {
		r.entrySize = size
	}
}

// WithDecode sets a custom parser function for the reader
func WithDecode(fn func(file *os.File, add AddFn) error) Option {
	return func(r *Reader) {
		if err := fn(r.file, r.add); err != nil {
			panic(fmt.Sprintf("failed to parse entries: %v", err))
		}
	}
}

// WithChunks configures the reader to handle files with fixed-size chunks
func WithChunks(chunkSize int) Option {
	return WithDecode(func(file *os.File, add AddFn) error {
		info, err := file.Stat()
		if err != nil {
			return fmt.Errorf("failed to get file stats: %w", err)
		}

		fileSize := info.Size()
		if chunkSize <= 0 {
			return fmt.Errorf("invalid chunk size: %d", chunkSize)
		}

		chunkCount := int(fileSize / int64(chunkSize))
		if chunkCount == 0 {
			return fmt.Errorf("file too small for chunk format")
		}

		for i := 0; i < chunkCount; i++ {
			add(uint32(i), uint32(i*chunkSize), uint32(chunkSize), 0, nil)
		}

		return nil
	})
}
