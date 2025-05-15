package ultima

import (
	"encoding/binary"
	"fmt"
	"image"
	"iter"
	"path/filepath"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
)

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
	*AnimdataEntry
	frames []AnimationFrame
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
	if len(Animdata) == 0 {
		if err := LoadAnimdata(filepath.Join(s.basePath, "animdata.mul")); err != nil {
			return nil, fmt.Errorf("LoadAnimation: failed loading animdata: %w", err)
		}
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

	// Retrieve metadata for this animation ID, defaulting to empty if missing
	var meta *AnimdataEntry
	if m, ok := Animdata[int(index)]; ok && m != nil {
		meta = m
	} else {
		meta = &AnimdataEntry{}
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
	return &Animation{
		AnimdataEntry: meta,
		frames:        frames,
	}, nil
}
