package uop

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

// decode decompresses data based on the compression flag
func decode(data []byte, flag CompressionType) ([]byte, error) {
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

// hashFileName calculates a hash for a filename as used in UOP files
// This is a direct port of the C# algorithm
func hashFileName(s string) uint64 {
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
