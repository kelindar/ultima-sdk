package ultima

import (
	"fmt"
	"image"

	"github.com/kelindar/ultima-sdk/internal/anim"
	"github.com/kelindar/ultima-sdk/internal/bitmap"
)

// AnimationFrame holds the center point and bitmap for a single animation frame.
type AnimationFrame struct {
	Center image.Point      // Center point for frame positioning
	Bitmap *bitmap.ARGB1555 // Frame image (ARGB1555), nil if not present
}

// Animation contains all frames for a body/action/direction animation.
type Animation struct {
	Frames []AnimationFrame // All frames in the animation
}

// LoadAnimation loads animation frames for a given body, action, direction, and hue.
// It returns an Animation containing all valid frames, or an error if loading fails.
func (s *SDK) LoadAnimation(body, action, direction, hue int, preserveHue, firstFrame bool) (*Animation, error) {
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

	frameData, _, err := animFile.Read(index)
	if err != nil {
		return nil, fmt.Errorf("LoadAnimation: failed to read anim.mul entry: %w", err)
	}
	if len(frameData) < 520 {
		return nil, fmt.Errorf("invalid frame data length: %d", len(frameData))
	}

	// Palette: first 512 bytes (256 colors, 2 bytes each)
	palette := make([]uint16, 256)
	for i := 0; i < 256; i++ {
		palette[i] = uint16(frameData[i*2]) | uint16(frameData[i*2+1])<<8
	}

	// Frame count and offsets
	frameCount := int(frameData[512])
	if frameCount == 0 {
		return &Animation{Frames: nil}, nil
	}
	frames := make([]AnimationFrame, 0, frameCount)
	for i := 0; i < frameCount; i++ {
		if 516+i*4+4 > len(frameData) {
			break
		}
		offset := int(frameData[516+i*4]) |
			int(frameData[516+i*4+1])<<8 |
			int(frameData[516+i*4+2])<<16 |
			int(frameData[516+i*4+3])<<24
		if offset == 0 || offset >= len(frameData) {
			continue
		}
		frameSlice := frameData[offset:]
		center, img, err := anim.DecodeFrame(palette, frameSlice, false)
		if err != nil || img == nil {
			continue
		}
		frames = append(frames, AnimationFrame{
			Center: center,
			Bitmap: img,
		})
	}
	return &Animation{Frames: frames}, nil
}
