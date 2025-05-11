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
)

// RadarColor retrieves the radar color for a given tile ID
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

	// Since the entire file is now a single entry, we can read it with index 0
	// and the result will be cached automatically
	data, err := file.Read(0)
	if err != nil {
		return 0, fmt.Errorf("failed to read radar color data: %w", err)
	}

	// Check if we have enough data for this tile ID
	// Each color is 2 bytes (uint16)
	bytePos := tileID * 2
	if bytePos+2 > len(data) {
		return 0, fmt.Errorf("invalid radar color data: file too small for tile ID %d", tileID)
	}

	// Extract the color value (little-endian)
	return binary.LittleEndian.Uint16(data[bytePos:]), nil
}

// RadarColorStatic retrieves the radar color for a given static tile ID
func (s *SDK) RadarColorStatic(tileID int) (uint16, error) {
	// Validate the tile ID
	if tileID < 0 || tileID >= radarColorEntries {
		return 0, fmt.Errorf("%w: %d (must be between 0 and 0x3FFF)", ErrInvalidRadarColorIndex, tileID)
	}

	// Calculate the adjusted tile ID (add offset for static tiles)
	adjustedID := tileID + radarColorStaticOffset

	// Load the radar color file
	file, err := s.loadRadarcol()
	if err != nil {
		return 0, fmt.Errorf("failed to load radar colors: %w", err)
	}

	// Since the entire file is now a single entry, we can read it with index 0
	// and the result will be cached automatically
	data, err := file.Read(0)
	if err != nil {
		return 0, fmt.Errorf("failed to read radar color data: %w", err)
	}

	// Check if we have enough data for this tile ID
	// Each color is 2 bytes (uint16)
	bytePos := adjustedID * 2
	if bytePos+2 > len(data) {
		return 0, fmt.Errorf("invalid radar color data: file too small for static tile ID %d", tileID)
	}

	// Extract the color value (little-endian)
	return binary.LittleEndian.Uint16(data[bytePos:]), nil
}

// RadarColors returns an iterator over all defined radar color mappings
func (s *SDK) RadarColors() iter.Seq2[int, uint16] {
	return func(yield func(int, uint16) bool) {
		// Load the radar color file
		file, err := s.loadRadarcol()
		if err != nil {
			return // Can't iterate if we can't load the file
		}

		// Get the entire file data - will be cached automatically
		data, err := file.Read(0)
		if err != nil {
			return // Can't iterate if we can't read the file
		}

		// Calculate how many full color entries we can extract
		// Each entry is a uint16 (2 bytes)
		entryCount := len(data) / 2
		if entryCount > totalRadarColors {
			entryCount = totalRadarColors
		}

		// Iterate over all color entries
		for i := 0; i < entryCount; i++ {
			// Extract color value (little-endian)
			color := binary.LittleEndian.Uint16(data[i*2:])

			// Calculate the logical tile ID
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

		// Get the entire file data - will be cached automatically
		data, err := file.Read(0)
		if err != nil {
			return // Can't iterate if we can't read the file
		}

		// Calculate how many land color entries we can extract
		entryCount := len(data) / 2
		if entryCount > radarColorEntries {
			entryCount = radarColorEntries
		}

		// Iterate over land tile entries (first half of the file)
		for i := 0; i < entryCount; i++ {
			// Extract color value (little-endian)
			color := binary.LittleEndian.Uint16(data[i*2:])

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

		// Get the entire file data - will be cached automatically
		data, err := file.Read(0)
		if err != nil {
			return // Can't iterate if we can't read the file
		}

		// Calculate how many static entries we can extract
		entryCount := len(data) / 2
		if entryCount <= radarColorStaticOffset {
			return // No static entries in this file
		}

		// Cap at maximum number of static entries
		staticEntryCount := entryCount - radarColorStaticOffset
		if staticEntryCount > radarColorEntries {
			staticEntryCount = radarColorEntries
		}

		// Iterate over static tile entries (second half of the file)
		for i := 0; i < staticEntryCount; i++ {
			// Get position in file (offset by radarColorStaticOffset)
			pos := (i + radarColorStaticOffset) * 2

			// Extract color value (little-endian)
			color := binary.LittleEndian.Uint16(data[pos:])

			// Yield the tile ID and color
			if !yield(i, color) {
				break
			}
		}
	}
}
