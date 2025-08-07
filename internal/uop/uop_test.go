// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package uop

import (
	"bytes"
	"compress/zlib"
	"path/filepath"
	"testing"

	uotest "github.com/kelindar/ultima-sdk/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewReader tests creating a new UOP reader
func TestNewReader(t *testing.T) {
	testDataPath := uotest.Path()
	require.NotEmpty(t, testDataPath, "Test data path should not be empty")

	// Find UOP files in test data directory
	uopFiles, err := filepath.Glob(filepath.Join(testDataPath, "*.uop"))
	if err != nil || len(uopFiles) == 0 {
		t.Skip("No UOP files found in test data directory")
		return
	}

	// Use the first found UOP file for testing
	testUOP := uopFiles[0]
	t.Logf("Testing with UOP file: %s", testUOP)

	// Test file opening
	reader, err := Open(testUOP, 10)
	require.NoError(t, err, "Failed to create reader")
	require.NotNil(t, reader, "Reader should not be nil")

	defer reader.Close()
}

// TestEntryOperations tests entry-related operations (Read and Entries methods)
func TestEntryOperations(t *testing.T) {
	testUOP := filepath.Join(uotest.Path(), "artLegacyMUL.uop")

	reader, err := Open(testUOP, 0x14000, WithExtension(".tga"), WithLength(0x13FDC))
	require.NoError(t, err)
	defer reader.Close()

	// Test Entries iterator with the new interface
	var count int
	var indices []uint32

	// Collect up to 10 entries for further testing
	for index := range reader.Entries() {
		indices = append(indices, index)
		count++
		if count >= 10 {
			break
		}
	}

	assert.GreaterOrEqual(t, count, 1, "Should have found at least 1 entry")

	// Test Read method with the first valid index
	if len(indices) > 0 {
		firstIndex := indices[0]
		data, err := reader.Entry(firstIndex)
		require.NoError(t, err)
		assert.NotNil(t, data, "Data should not be nil")

		// Test invalid index
		_, err = reader.Entry(uint32(0xFFFFFFFF))
		assert.Error(t, err, "Reading invalid index should return error")
	}
}

// TestCompression tests the compression/decompression functionality
func TestCompression(t *testing.T) {
	// Test zlib compression
	t.Run("Zlib", func(t *testing.T) {
		originalData := []byte("This is a test for zlib compression")

		// Compress the data using zlib
		var compressedBuf bytes.Buffer
		zlibWriter := zlib.NewWriter(&compressedBuf)
		_, err := zlibWriter.Write(originalData)
		require.NoError(t, err)
		zlibWriter.Close()
		compressedData := compressedBuf.Bytes()

		// Decompress using our function
		decompressedData, err := decodeZlib(compressedData)
		require.NoError(t, err)

		// Verify the decompressed data matches the original
		assert.Equal(t, originalData, decompressedData)
	})

	// Test error cases
	t.Run("Errors", func(t *testing.T) {
		// Test invalid compression flag
		_, err := decode([]byte("test"), CompressionType(99))
		assert.Error(t, err)

		// Test corrupted zlib data
		_, err = decodeZlib([]byte("not zlib data"))
		assert.Error(t, err)

		// Test mythic decompression with invalid data
		_, err = decodeMythic([]byte{0, 0, 0}) // Too short
		assert.Error(t, err)
	})
}
