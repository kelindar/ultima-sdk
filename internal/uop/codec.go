package uop

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

// decode decompresses data based on the compression flag
func decode(data []byte, flag CompressionFlag) ([]byte, error) {
	switch flag {
	case CompressionNone:
		return data, nil
	case CompressionZlib:
		return decodeZlib(data)
	case CompressionMythic:
		return decodeMythic(data)
	default:
		return nil, fmt.Errorf("unknown compression flag: %d", flag)
	}
}

// decodeZlib decompresses zlib data
func decodeZlib(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// decodeMythic decompresses data using the Mythic compression algorithm
// Ported from C# Ultima SDK's Helpers/decodeMythic.cs
func decodeMythic(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short for mythic decompression")
	}

	// Get decompressed size from the first 4 bytes (little-endian)
	decompressedSize := int(binary.LittleEndian.Uint32(data[:4]))
	if decompressedSize <= 0 {
		return nil, fmt.Errorf("invalid decompressed size: %d", decompressedSize)
	}

	// Initialize result buffer
	result := make([]byte, decompressedSize)
	resultPos := 0

	// Start processing from after the size header
	pos := 4

	for pos < len(data) && resultPos < decompressedSize {
		// Read the compression flag
		flag := data[pos]
		pos++

		if flag == 0 {
			// Raw copy
			if pos >= len(data) {
				return nil, fmt.Errorf("incomplete data at position %d", pos)
			}

			copyLen := int(data[pos])
			pos++

			// Copy the raw data
			if pos+copyLen > len(data) || resultPos+copyLen > decompressedSize {
				return nil, fmt.Errorf("data bounds exceeded during raw copy")
			}

			copy(result[resultPos:], data[pos:pos+copyLen])
			resultPos += copyLen
			pos += copyLen
		} else {
			// RLE (Run-Length Encoding) compression
			copyLen := int(flag)

			if pos >= len(data) || resultPos+copyLen > decompressedSize {
				return nil, fmt.Errorf("data bounds exceeded during RLE decompression")
			}

			// Copy the same byte multiple times
			for i := 0; i < copyLen; i++ {
				result[resultPos+i] = data[pos]
			}

			resultPos += copyLen
			pos++
		}
	}

	if resultPos != decompressedSize {
		return nil, fmt.Errorf("decompressed size mismatch: got %d, expected %d", resultPos, decompressedSize)
	}

	return result, nil
}
