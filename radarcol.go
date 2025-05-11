package ultima

import (
	"encoding/binary"
	"errors"
	"fmt"
	"iter"
)

var (
	// ErrInvalidRadarColorIndex is returned when an invalid radar color index is requested
	ErrInvalidRadarColorIndex = errors.New("invalid radar color index")
)

const (
	// radarColorEntries is the number of entries for each type (land and static)
	radarColorEntries = 0x4000

	// totalRadarColors is the total number of radar colors (land + static)
	totalRadarColors = 0x8000

	// radarColorLandOffset is the offset for land tile colors (0)
	radarColorLandOffset = 0

	// radarColorStaticOffset is the offset for static tile colors (0x4000)
	radarColorStaticOffset = 0x4000

	// radarChunkSize is the size of chunks to read (512 entries = 1024 bytes)
	radarChunkSize = 512
)

// RadarColor retrieves the radar color for a given tile ID
// If staticTileID is true, the ID is treated as a static tile ID (offsetting by 0x4000)
// Otherwise, it's treated as a land tile ID
func (s *SDK) RadarColor(tileID int) (uint16, error) {
	// Validate the tile ID
	if tileID < 0 || tileID >= radarColorEntries {
		return 0, fmt.Errorf("%w: %d (must be between 0 and 0x3FFF)", ErrInvalidRadarColorIndex, tileID)
	}

	// Load the radar color file
	file, err := s.loadRadarcol()
	if err != nil {
		return 0, fmt.Errorf("failed to load radar colors: %w", err)
	}

	// Calculate chunk and offset
	chunkIndex := tileID / radarChunkSize
	offset := tileID % radarChunkSize

	// Since we're using a chunk size that's smaller than the total file,
	// we need to read the appropriate chunk
	chunkData, err := file.Read(uint64(chunkIndex))
	if err != nil {
		return 0, fmt.Errorf("failed to read radar color chunk: %w", err)
	}

	// Each color is 2 bytes (uint16)
	if offset*2+2 > len(chunkData) {
		return 0, fmt.Errorf("invalid radar color data: chunk %d too small", chunkIndex)
	}

	// Extract the color value (little-endian)
	return binary.LittleEndian.Uint16(chunkData[offset*2:]), nil
}

// RadarColorStatic retrieves the radar color for a given static tile ID
func (s *SDK) RadarColorStatic(tileID int) (uint16, error) {
	// We just use the same index in the second half of the file
	if tileID < 0 || tileID >= radarColorEntries {
		return 0, fmt.Errorf("%w: %d (must be between 0 and 0x3FFF)", ErrInvalidRadarColorIndex, tileID)
	}

	// Calculate the adjusted tile ID (add offset for static tiles)
	adjustedID := tileID + radarColorStaticOffset

	// Calculate chunk and offset
	chunkIndex := adjustedID / radarChunkSize
	offset := adjustedID % radarChunkSize

	// Load the radar color file
	file, err := s.loadRadarcol()
	if err != nil {
		return 0, fmt.Errorf("failed to load radar colors: %w", err)
	}

	// Read the appropriate chunk
	chunkData, err := file.Read(uint64(chunkIndex))
	if err != nil {
		return 0, fmt.Errorf("failed to read radar color chunk: %w", err)
	}

	// Each color is 2 bytes (uint16)
	if offset*2+2 > len(chunkData) {
		return 0, fmt.Errorf("invalid radar color data: chunk %d too small", chunkIndex)
	}

	// Extract the color value (little-endian)
	return binary.LittleEndian.Uint16(chunkData[offset*2:]), nil
}

// RadarColors returns an iterator over all defined radar color mappings
func (s *SDK) RadarColors() iter.Seq2[int, uint16] {
	return func(yield func(int, uint16) bool) {
		// Load the radar color file
		file, err := s.loadRadarcol()
		if err != nil {
			return // Can't iterate if we can't load the file
		}

		// Iterate through all possible indices
		for i := 0; i < totalRadarColors; i++ {
			// Calculate which chunk this index falls into
			chunkIndex := i / radarChunkSize
			offset := i % radarChunkSize

			// Read the chunk
			chunkData, err := file.Read(uint64(chunkIndex))
			if err != nil {
				continue // Skip this entry if we can't read it
			}

			// Make sure we have enough data
			if offset*2+2 > len(chunkData) {
				continue // Skip this entry if the chunk is too small
			}

			// Extract the color value (little-endian)
			color := binary.LittleEndian.Uint16(chunkData[offset*2:])

			// The i value is the global index, but we want to yield the logical tile ID
			// For land tiles (i < 0x4000), the ID is just i
			// For static tiles (i >= 0x4000), the ID is i - 0x4000
			tileID := i
			if tileID >= radarColorStaticOffset {
				tileID -= radarColorStaticOffset
			}

			// Yield the tile ID and color
			if !yield(tileID, color) {
				break
			}
		}
	}
}

// RadarColorsLand returns an iterator over land tile radar color mappings
func (s *SDK) RadarColorsLand() iter.Seq2[int, uint16] {
	return func(yield func(int, uint16) bool) {
		// Load the radar color file
		file, err := s.loadRadarcol()
		if err != nil {
			return // Can't iterate if we can't load the file
		}

		// Iterate through all land tile indices (0 to 0x3FFF)
		for i := 0; i < radarColorEntries; i++ {
			// Calculate which chunk this index falls into
			chunkIndex := i / radarChunkSize
			offset := i % radarChunkSize

			// Read the chunk
			chunkData, err := file.Read(uint64(chunkIndex))
			if err != nil {
				continue // Skip this entry if we can't read it
			}

			// Make sure we have enough data
			if offset*2+2 > len(chunkData) {
				continue // Skip this entry if the chunk is too small
			}

			// Extract the color value (little-endian)
			color := binary.LittleEndian.Uint16(chunkData[offset*2:])

			// Yield the tile ID and color
			if !yield(i, color) {
				break
			}
		}
	}
}

// RadarColorsStatic returns an iterator over static tile radar color mappings
func (s *SDK) RadarColorsStatic() iter.Seq2[int, uint16] {
	return func(yield func(int, uint16) bool) {
		// Load the radar color file
		file, err := s.loadRadarcol()
		if err != nil {
			return // Can't iterate if we can't load the file
		}

		// Iterate through all static tile indices (0x4000 to 0x7FFF)
		for i := 0; i < radarColorEntries; i++ {
			// Calculate the adjusted tile ID (add offset for static tiles)
			adjustedID := i + radarColorStaticOffset

			// Calculate which chunk this index falls into
			chunkIndex := adjustedID / radarChunkSize
			offset := adjustedID % radarChunkSize

			// Read the chunk
			chunkData, err := file.Read(uint64(chunkIndex))
			if err != nil {
				continue // Skip this entry if we can't read it
			}

			// Make sure we have enough data
			if offset*2+2 > len(chunkData) {
				continue // Skip this entry if the chunk is too small
			}

			// Extract the color value (little-endian)
			color := binary.LittleEndian.Uint16(chunkData[offset*2:])

			// Yield the tile ID and color (without the offset)
			if !yield(i, color) {
				break
			}
		}
	}
}
