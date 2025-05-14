package ultima

import (
	"encoding/binary"
	"fmt"
	"image"
	"iter"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
)

// GumpInfo contains metadata about a gump, like its ID, width, and height.
type GumpInfo struct {
	ID     int // ID of the gump
	Width  int // Width in pixels
	Height int // Height in pixels
}

// Gump represents a UI element or graphic.
// The Image is loaded lazily.
type Gump struct {
	GumpInfo
	imageData []byte // Raw gump data
}

// Image retrieves and decodes the gump's graphical representation.
func (g *Gump) Image() (image.Image, error) {
	if g.imageData == nil || len(g.imageData) == 0 {
		return nil, fmt.Errorf("%w: no gump data available", ErrInvalidArtData)
	}

	// Decode the image
	return decodeGumpData(g.imageData, g.Width, g.Height)
}

// Gump retrieves a specific gump graphic by its ID.
// It handles reading from .mul or UOP files.
// The returned Gump object allows for lazy loading of its image.
func (s *SDK) Gump(id int) (*Gump, error) {
	// Load the gump file
	file, err := s.loadGump()
	if err != nil {
		return nil, err
	}

	// Read the raw data and info
	data, extra, err := file.Read(uint32(id))
	if err != nil {
		return nil, err
	}

	// The extra data contains width and height information (lower 32 bits)
	width := int((extra >> 16) & 0xFFFF)
	height := int(extra & 0xFFFF)

	// If dimensions are not valid, try to extract from the gump data itself (UOP fallback)
	if width <= 0 || height <= 0 || width > 2048 || height > 2048 {
		if len(data) >= 8 {
			width = int(binary.LittleEndian.Uint32(data[0:4]))
			height = int(binary.LittleEndian.Uint32(data[4:8]))
		}
	}

	// Sanity check again
	if width <= 0 || height <= 0 || width > 2048 || height > 2048 {
		return nil, fmt.Errorf("%w: invalid gump dimensions %dx%d", ErrInvalidArtData, width, height)
	}

	return &Gump{
		GumpInfo: GumpInfo{
			ID:     id,
			Width:  width,
			Height: height,
		},
		imageData: data,
	}, nil
}

// GumpInfos returns an iterator over metadata (ID, width, height) for all available gumps.
// This is efficient for listing gumps without loading all their pixel data.
func (s *SDK) GumpInfos() iter.Seq[GumpInfo] {
	return func(yield func(GumpInfo) bool) {
		// Load the gump file
		file, err := s.loadGump()
		if err != nil {
			return
		}

		// Iterate through all entries in the file
		for id := uint32(0); ; id++ {
			_, extra, err := file.Read(id)
			if err != nil {
				break // End of file or invalid entry
			}

			// Skip entries with invalid dimensions
			width := int((extra >> 16) & 0xFFFF)
			height := int(extra & 0xFFFF)
			if width <= 0 || height <= 0 {
				continue
			}

			// Create a GumpInfo with just the metadata
			info := GumpInfo{
				ID:     int(id),
				Width:  width,
				Height: height,
			}

			// Yield the info to the iterator
			if !yield(info) {
				break
			}
		}
	}
}

// decodeGumpData converts raw gump data to an image.Image.
// Gumps are stored in a run-length encoded format with 16-bit color values.
func decodeGumpData(data []byte, width, height int) (image.Image, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("%w: gump data too short for lookup table", ErrInvalidArtData)
	}

	// Create a new bitmap to hold the decoded image
	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))

	// Read lookup pointers for each line
	lookupTable := make([]int, height)
	for y := 0; y < height; y++ {
		// Each lookup is a 4-byte offset from the start of the data
		if (y*4)+4 > len(data) {
			return nil, fmt.Errorf("%w: gump data truncated in lookup table at line %d", ErrInvalidArtData, y)
		}
		lookupTable[y] = int(binary.LittleEndian.Uint32(data[y*4 : y*4+4]))
	}

	// Process each line
	for y := 0; y < height; y++ {
		x := 0 // Current x position in the output image

		offset := lookupTable[y]
		if offset < 0 || offset >= len(data) {
			return nil, fmt.Errorf("%w: invalid lookup offset %d for line %d", ErrInvalidArtData, offset, y)
		}

		// Process RLE data for this line
		for x < width {
			// Need at least 4 more bytes for an RLE pair (2 bytes color + 2 bytes run length)
			if offset+4 > len(data) {
				return nil, fmt.Errorf("%w: gump data truncated during RLE decoding at y=%d, x=%d", ErrInvalidArtData, y, x)
			}

			// Read color and run length
			colorValue := binary.LittleEndian.Uint16(data[offset : offset+2])
			offset += 2
			runLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
			offset += 2

			// 0,0 is the terminator for the line
			if colorValue == 0 && runLength == 0 {
				break
			}

			if colorValue == 0 {
				// Transparent run
				x += runLength
			} else {
				// Opaque run - flip the most significant bit to set alpha
				colorValue ^= 0x8000

				// Draw the pixels
				for i := 0; i < runLength; i++ {
					if x+i < width {
						img.Set(x+i, y, bitmap.ARGB1555Color(colorValue))
					}
				}
				x += runLength
			}
		}
	}

	return img, nil
}
