package anim

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// AnimdataEntry holds metadata for a single animation (from animdata.mul)
type AnimdataEntry struct {
	FrameData     [64]int8
	Unknown       uint8
	FrameCount    uint8
	FrameInterval uint8
	FrameStart    uint8
}

// Animdata holds all loaded animation metadata.
var Animdata = map[int]*AnimdataEntry{}

// LoadAnimdata loads animdata.mul into memory.
func LoadAnimdata(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("animdata: failed to open '%s': %w", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("animdata: failed to stat '%s': %w", path, err)
	}
	size := info.Size()

	const entrySize = 64 + 4        // 64 sbytes + 4 metadata bytes
	const chunkSize = 4 + 8*entrySize // 4 bytes header + 8 entries
	count := size / chunkSize

	reader := io.NewSectionReader(file, 0, size)
	id := 0
	Animdata = make(map[int]*AnimdataEntry)
	for c := int64(0); c < count; c++ {
		var hdr int32
		if err := binary.Read(reader, binary.LittleEndian, &hdr); err != nil {
			return fmt.Errorf("animdata: read header: %w", err)
		}
		for i := 0; i < 8; i++ {
			var entry AnimdataEntry
			for j := 0; j < 64; j++ {
				if err := binary.Read(reader, binary.LittleEndian, &entry.FrameData[j]); err != nil {
					return fmt.Errorf("animdata: read frame data: %w", err)
				}
			}
			buf := make([]byte, 4)
			if _, err := io.ReadFull(reader, buf); err != nil {
				return fmt.Errorf("animdata: read metadata: %w", err)
			}
			entry.Unknown = buf[0]
			entry.FrameCount = buf[1]
			entry.FrameInterval = buf[2]
			entry.FrameStart = buf[3]
			if entry.FrameCount > 0 {
				Animdata[id] = &entry
			}
			id++
		}
	}
	return nil
}
