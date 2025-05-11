package ultima

import (
	"encoding/binary"
	"errors"
	"fmt"
	"iter"
)

var (
	ErrInvalidRadarColorIndex = errors.New("invalid radar color index")
)

const (
	radarColorEntries      = 0x4000
	totalRadarColors       = 0x8000
	radarColorStaticOffset = 0x4000
)

// RadarColor retrieves the radar color for a given land tile ID
func (s *SDK) RadarColor(tileID int) (uint16, error) {
	return s.getRadarColor(tileID, 0)
}

// RadarColorStatic retrieves the radar color for a given static tile ID
func (s *SDK) RadarColorStatic(tileID int) (uint16, error) {
	return s.getRadarColor(tileID, radarColorStaticOffset)
}

// getRadarColor retrieves a radar color with the given offset
func (s *SDK) getRadarColor(tileID int, offset int) (uint16, error) {
	if tileID < 0 || tileID >= radarColorEntries {
		return 0, fmt.Errorf("%w: %d (must be between 0 and 0x3FFF)", ErrInvalidRadarColorIndex, tileID)
	}

	data, err := s.loadRadarData()
	if err != nil {
		return 0, err
	}

	bytePos := (tileID + offset) * 2
	if bytePos+2 > len(data) {
		return 0, fmt.Errorf("invalid radar color data: file too small for tile ID %d", tileID)
	}

	return binary.LittleEndian.Uint16(data[bytePos:]), nil
}

// loadRadarData loads the entire radar color data file
func (s *SDK) loadRadarData() ([]byte, error) {
	file, err := s.loadRadarcol()
	if err != nil {
		return nil, fmt.Errorf("failed to load radar colors: %w", err)
	}

	data, err := file.Read(0)
	if err != nil {
		return nil, fmt.Errorf("failed to read radar color data: %w", err)
	}

	return data, nil
}

// RadarColors returns an iterator over all defined radar color mappings
func (s *SDK) RadarColors() iter.Seq2[int, uint16] {
	return func(yield func(int, uint16) bool) {
		data, err := s.loadRadarData()
		if err != nil {
			return
		}

		entryCount := len(data) / 2
		if entryCount > totalRadarColors {
			entryCount = totalRadarColors
		}

		for i := 0; i < entryCount; i++ {
			color := binary.LittleEndian.Uint16(data[i*2:])
			tileID := i
			if tileID >= radarColorStaticOffset {
				tileID -= radarColorStaticOffset
			}

			if !yield(tileID, color) {
				break
			}
		}
	}
}

// RadarColorsLand returns an iterator over land tile radar color mappings
func (s *SDK) RadarColorsLand() iter.Seq2[int, uint16] {
	return s.radarColorsByType(0, radarColorEntries)
}

// RadarColorsStatic returns an iterator over static tile radar color mappings
func (s *SDK) RadarColorsStatic() iter.Seq2[int, uint16] {
	return s.radarColorsByType(radarColorStaticOffset, radarColorEntries)
}

// radarColorsByType returns an iterator for a specific type of radar colors
func (s *SDK) radarColorsByType(offset, count int) iter.Seq2[int, uint16] {
	return func(yield func(int, uint16) bool) {
		data, err := s.loadRadarData()
		if err != nil {
			return
		}

		entryCount := (len(data) / 2) - offset
		if entryCount <= 0 {
			return
		}

		if entryCount > count {
			entryCount = count
		}

		for i := 0; i < entryCount; i++ {
			pos := (i + offset) * 2
			if pos+2 > len(data) {
				break
			}

			color := binary.LittleEndian.Uint16(data[pos:])
			if !yield(i, color) {
				break
			}
		}
	}
}
