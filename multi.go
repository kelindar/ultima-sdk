// Package ultima provides access to Ultima Online multi structures.
package ultima

import (
	"encoding/binary"
	"fmt"
)

// MultiItem represents a single item within a multi-structure.
type MultiItem struct {
	ItemID  uint16 // Tile ID of the item.
	OffsetX int16
	OffsetY int16
	OffsetZ int16
	Flags   uint32
	Unk1    uint32 // Only present in UOAHS format (16 bytes per entry)
}

// Multi represents a multi-structure (e.g., house, boat) in Ultima Online.
type Multi struct {
	Items []MultiItem
}

// Multi returns a Multi structure by id, loading from multi.mul/multi.idx via loadMulti().
// This follows the same pattern as other SDK data accessors.
func (s *SDK) Multi(id int) (*Multi, error) {
	file, err := s.loadMulti()
	if err != nil {
		return nil, err
	}
	data, _, err := file.Read(uint32(id))
	if err != nil {
		return nil, err
	}
	if data == nil || len(data) == 0 {
		return nil, fmt.Errorf("multi entry %d not found", id)
	}
	// TODO: Detect UOAHS format properly; for now, assume false
	isUOAHS := false
	entrySize := 12
	if isUOAHS {
		entrySize = 16
	}
	var items []MultiItem
	for i := 0; i+entrySize <= len(data); i += entrySize {
		item := MultiItem{
			ItemID:  binary.LittleEndian.Uint16(data[i:]),
			OffsetX: int16(binary.LittleEndian.Uint16(data[i+2:])),
			OffsetY: int16(binary.LittleEndian.Uint16(data[i+4:])),
			OffsetZ: int16(binary.LittleEndian.Uint16(data[i+6:])),
			Flags:   binary.LittleEndian.Uint32(data[i+8:]),
		}
		if isUOAHS {
			item.Unk1 = binary.LittleEndian.Uint32(data[i+12:])
		}
		items = append(items, item)
	}
	return &Multi{Items: items}, nil
}
