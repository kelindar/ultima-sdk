package ultima

import (
	"encoding/binary"
	"fmt"
	"image"
	"iter"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
	"github.com/kelindar/ultima-sdk/internal/uofile"
)

// AnimdataEntry holds metadata for a single animation (from animdata.mul)
type AnimdataEntry struct {
	FrameData     [64]int8
	Unknown       uint8
	FrameCount    uint8
	FrameInterval uint8
	FrameStart    uint8
}

// AnimationFrame holds the center point and bitmap for a single animation frame.
type AnimationFrame struct {
	Center image.Point      // Center point for frame positioning
	Bitmap *bitmap.ARGB1555 // Frame image (ARGB1555), nil if not present
}

// Image retrieves and returns the frame's image.
func (af AnimationFrame) Image() (image.Image, error) {
	return af.Bitmap, nil
}

// Animation contains a sequence of frames for a body/action/direction animation.
// Use Frames() to iterate through AnimationFrame instances.
// Metadata returns the animation metadata from animdata.mul.
type Animation struct {
	Name         string
	AnimdataEntry *AnimdataEntry
	frames        []AnimationFrame
}

// Frames returns a sequence (iter.Seq) of AnimationFrame for this animation.
func (a *Animation) Frames() iter.Seq[AnimationFrame] {
	return func(yield func(AnimationFrame) bool) {
		for _, f := range a.frames {
			if !yield(f) {
				break
			}
		}
	}
}

// Animation loads animation frames for a given body, action, direction, and hue.
func (s *SDK) Animation(body, action, direction, hue int, preserveHue, firstFrame bool) (*Animation, error) {
	animdataFile, err := s.loadAnimdata()
	if err != nil {
		return nil, fmt.Errorf("Animation: failed loading animdata: %w", err)
	}

	// Select animX.mul based on body - match the C# implementation
	// FileType in C# ranges from 1-5, with 1 being the default
	fileType := 1 // Default to 1 (which is anim.mul)

	// Note: This is a simplified mapping - C# uses a more complex logic via BodyConverter
	// For now, we'll use a simple mapping based on the body range
	// In C#, both fileType and offset calculation have specific ranges
	animFile, err := s.loadAnim(fileType - 1) // Convert 1-based to 0-based for our loadAnim
	if err != nil {
		return nil, fmt.Errorf("load animation body=%d file=%d: %w", body, fileType, err)
	}

	// Calculate index based on C# logic
	var index uint32
	if body < 200 {
		index = uint32(body * 110)
	} else if body < 400 {
		index = 22000 + uint32((body-200)*65)
	} else {
		index = 35000 + uint32((body-400)*175)
	}

	// Add action and direction offsets
	index += uint32(action * 5)

	// Handle direction offset like in C#
	if direction <= 4 {
		index += uint32(direction)
	} else {
		index += uint32(direction - ((direction - 4) * 2))
	}

	// For animdata.mul, extract the correct entry from the chunk using body ID
	chunkIndex := body / 8
	entryOffset := body % 8
	chunk, _, err := animdataFile.Read(uint32(chunkIndex))
	if err != nil {
		return nil, fmt.Errorf("Animation: failed reading animdata chunk for body %d: %w", body, err)
	}
	if len(chunk) < 4+(entryOffset+1)*68 {
		return nil, fmt.Errorf("Animation: animdata chunk too small for body %d", body)
	}
	entry := chunk[4+entryOffset*68 : 4+(entryOffset+1)*68]
	fmt.Printf("body=%d chunkIndex=%d entryOffset=%d chunkLen=%d entry[0:8]=%v\n",
		body, chunkIndex, entryOffset, len(chunk), entry[:8])
	meta, err := decodeAnimdata(entry)
	if err != nil {
		return nil, fmt.Errorf("Animation: failed decoding animdata entry: %w", err)
	}

	frameData, _, err := animFile.Read(index)
	if err != nil {
		return nil, fmt.Errorf("LoadAnimation: failed to read anim.mul entry: %w", err)
	}

	// Palette: first 512 bytes (256 colors, 2 bytes each)
	const paletteSize = 512
	const frameCountSize = 4
	if len(frameData) < paletteSize+frameCountSize {
		return nil, fmt.Errorf("invalid frame data length: %d", len(frameData))
	}

	palette := make([]uint16, 256)
	for i := 0; i < 256; i++ {
		// C# does: palette[i] = (ushort)(bin.ReadUInt16() ^ 0x8000)
		// This XORs with the high bit which controls transparency
		color := uint16(frameData[i*2]) | uint16(frameData[i*2+1])<<8
		palette[i] = color ^ 0x8000 // XOR with 0x8000 to match C# implementation
	}

	// Frame count and lookup table.
	frameCount := int(int32(binary.LittleEndian.Uint32(frameData[paletteSize : paletteSize+frameCountSize])))
	if frameCount <= 0 {
		return &Animation{
			AnimdataEntry: meta,
			frames:        nil,
		}, nil
	}
	// Lookup table starts immediately after the frame count.
	const lookupStart = paletteSize + frameCountSize
	frames := make([]AnimationFrame, 0, frameCount)
	for i := 0; i < frameCount; i++ {
		entry := lookupStart + i*4
		if entry+4 > len(frameData) {
			break
		}
		rel := int(int32(binary.LittleEndian.Uint32(frameData[entry : entry+4])))
		if rel <= 0 {
			continue
		}
		offset := paletteSize + rel
		if offset < 0 || offset >= len(frameData) {
			continue
		}
		frameSlice := frameData[offset:]
		flip := direction > 4
		center, img, err := decodeFrame(palette, frameSlice, flip)
		if err != nil || img == nil {
			continue
		}
		frames = append(frames, AnimationFrame{Center: center, Bitmap: img})
	}
	// Lookup the animation name using the embedded lookup
	name := "Unknown"
	if n := uofile.AnimationNameByBody(body); n != "" {
		name = n
	}
	return &Animation{
		Name:         name,
		AnimdataEntry: meta,
		frames:        frames,
	}, nil
}

// AnimationNames provides canonical names for humanoid animation actions by index
var AnimationNames = []string{
	"Idle",     // 0
	"Walk",     // 1
	"Run",      // 2
	"Eat",      // 3
	"Attack",   // 4
	"Unknown5", // 5
	"Unknown6", // 6
	"Unknown7", // 7
	"Unknown8", // 8
	"Unknown9", // 9
	// Extend as needed for your use case
}

// AnimationName returns the canonical name for a given animation action index
func AnimationName(action int) string {
	if action >= 0 && action < len(AnimationNames) {
		return AnimationNames[action]
	}
	return "Unknown"
}

// decodeAnimdata parses the animation metadata from the provided binary data.
// The data should be exactly 68 bytes long (64 bytes of frame data + 4 bytes of metadata).
// Format:
//   - FrameData: 64 bytes (signed 8-bit integers)
//   - Unknown: 1 byte (unsigned 8-bit integer)
//   - FrameCount: 1 byte (unsigned 8-bit integer)
//   - FrameInterval: 1 byte (unsigned 8-bit integer)
//   - FrameStart: 1 byte (unsigned 8-bit integer)
func decodeAnimdata(data []byte) (*AnimdataEntry, error) {
	const expectedSize = 68 // 64 (FrameData) + 1 (Unknown) + 1 (FrameCount) + 1 (FrameInterval) + 1 (FrameStart)
	if len(data) < expectedSize {
		return nil, fmt.Errorf("invalid animdata length: expected at least %d bytes, got %d", expectedSize, len(data))
	}

	// Create a new AnimdataEntry
	entry := &AnimdataEntry{
		FrameData:     [64]int8{},
		Unknown:       data[64],
		FrameCount:    data[65],
		FrameInterval: data[66],
		FrameStart:    data[67],
	}

	// Copy the frame data
	for i := 0; i < 64; i++ {
		entry.FrameData[i] = int8(data[i])
	}

	return entry, nil
}
