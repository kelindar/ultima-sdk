package ultima

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"iter"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
)

// Land art tile ID range constants
const (
	landTileMax       = 0x4000 // Maximum ID for land tiles
	staticTileMinID   = 0x4000 // First ID for static tiles
	maxValidArtIndex  = 0xFFFF // Maximum possible art index
	landTileSize      = 44     // Land tiles are always 44x44 pixels
	landTileRawLength = 2048   // Raw data length for land tiles
)

var (
	ErrInvalidTileID  = errors.New("invalid tile ID")
	ErrNoArtData      = errors.New("no art data available for tile")
	ErrInvalidArtData = errors.New("invalid art data")
)

// ArtTile represents a piece of art (land or static item).
// Information is combined from art.mul/artidx.mul and tiledata.mul.
// The image is loaded lazily.
type ArtTile struct {
	ID        int      // ID of the tile
	Name      string   // Name from TileData
	Flags     TileFlag // Flags from TileData
	Height    int8     // Height of the tile, from TileData
	isLand    bool     // Whether this is a land tile (true) or static tile (false)
	imageData []byte   // Raw image data, cleared after Image() is called
}

// Image retrieves and decodes the art tile's graphical representation.
// The image is loaded on the first call and cached for subsequent calls.
func (a *ArtTile) Image() (image.Image, error) {
	if a.imageData == nil || len(a.imageData) == 0 {
		return nil, ErrNoArtData
	}

	if a.isLand {
		return decodeLandArt(a.imageData)
	} else {
		return decodeStaticArt(a.imageData)
	}
}

// ArtTile returns an art tile by ID.
// Use LandArtTile or StaticArtTile for clearer semantics.
func (s *SDK) ArtTile(id int) (*ArtTile, error) {
	if id < 0 || id > maxValidArtIndex {
		return nil, ErrInvalidTileID
	}

	if id < landTileMax {
		return s.LandArtTile(id)
	}
	return s.StaticArtTile(id - staticTileMinID)
}

// LandArtTile retrieves a land art tile by its ID.
func (s *SDK) LandArtTile(id int) (*ArtTile, error) {
	if id < 0 || id >= landTileMax {
		return nil, fmt.Errorf("%w: land tile ID %d out of range [0-%d]",
			ErrInvalidTileID, id, landTileMax-1)
	}

	// Load the art file
	file, err := s.loadArt()
	if err != nil {
		return nil, err
	}

	// Read the land tile data
	info, _ := s.LandTile(id)
	data, _, err := file.Read(uint32(id))
	if err != nil {
		return nil, err
	}

	return &ArtTile{
		ID:        id,
		Name:      info.Name,
		Flags:     info.Flags,
		isLand:    true,
		imageData: data,
	}, nil
}

// StaticArtTile retrieves a static art tile by its ID.
func (s *SDK) StaticArtTile(id int) (*ArtTile, error) {
	if id < 0 || id > maxValidArtIndex-staticTileMinID {
		return nil, fmt.Errorf("%w: static tile ID %d out of range [0-%d]",
			ErrInvalidTileID, id, maxValidArtIndex-staticTileMinID)
	}

	// Calculate the actual ID in the art file
	artID := id + staticTileMinID

	// Load the art file
	file, err := s.loadArt()
	if err != nil {
		return nil, err
	}

	// Read the static tile data
	info, _ := s.StaticTile(id)
	data, _, err := file.Read(uint32(artID))
	if err != nil {
		return nil, err
	}

	// Create and return the ArtTile
	return &ArtTile{
		ID:        artID,
		Name:      info.Name,
		Flags:     info.Flags,
		Height:    int8(info.Height),
		isLand:    false,
		imageData: data,
	}, nil
}

// LandArtTiles returns an iterator over all available land art tiles.
func (s *SDK) LandArtTiles() iter.Seq[*ArtTile] {
	return func(yield func(*ArtTile) bool) {
		for i := uint32(0); i < landTileMax; i++ {
			tile, err := s.LandArtTile(int(i))
			if err != nil {
				continue
			}

			if !yield(tile) {
				break
			}
		}
	}
}

// StaticArtTiles returns an iterator over all available static art tiles.
func (s *SDK) StaticArtTiles() iter.Seq[*ArtTile] {
	return func(yield func(*ArtTile) bool) {
		for i := uint32(staticTileMinID); i <= maxValidArtIndex; i++ {
			tile, err := s.StaticArtTile(int(i - staticTileMinID))
			if err != nil {
				continue
			}

			if !yield(tile) {
				break
			}
		}
	}
}

// decodeLandArt converts raw land art data into an image.Image.
// Land art is always 44x44 pixels. The format is essentially a run-length
// encoded 44x44 image where each 2-byte value represents a color index.
func decodeLandArt(data []byte) (image.Image, error) {
	if len(data) < landTileRawLength {
		return nil, fmt.Errorf("%w: land art data too short, expected %d bytes, got %d",
			ErrInvalidArtData, landTileRawLength, len(data))
	}

	img := bitmap.NewARGB1555(image.Rect(0, 0, landTileSize, landTileSize))
	offset := 0
	for y := 0; y < 22; y++ {
		// Start at the center-top of the tile and work outward
		// For the first 22 rows, each row gets 2 more pixels
		startX := 22 - y - 1
		pixelsInRow := (y * 2) + 2 // Number of pixels in this row
		for x := 0; x < pixelsInRow; x++ {
			if offset+1 >= len(data) {
				return nil, fmt.Errorf("%w: land art data truncated in first half", ErrInvalidArtData)
			}

			// Read 16-bit color value (little-endian)
			colorValue := binary.LittleEndian.Uint16(data[offset : offset+2])
			colorValue |= 0x8000 // Set the alpha bit (make it opaque)
			offset += 2

			// Set the pixel in the bitmap
			bitmapX := startX + x
			bitmapY := y
			img.Set(bitmapX, bitmapY, bitmap.ARGB1555Color(colorValue))
		}
	}

	for y := 0; y < 22; y++ {
		// For the last 22 rows, each row gets 2 fewer pixels
		// C# xOffset for this part effectively starts at 0 and increments with y_loop.
		// C# xRun for this part effectively starts at 44 and decrements by 2 with y_loop.
		startX := y                 // Corrected: C#'s xOffset for this part of the diamond
		pixelsInRow := 44 - (2 * y) // Corrected: Number of pixels for this row
		for x := 0; x < pixelsInRow; x++ {
			if offset+1 >= len(data) {
				return nil, fmt.Errorf("%w: land art data truncated in second half", ErrInvalidArtData)
			}

			// Read 16-bit color value (little-endian)
			colorValue := binary.LittleEndian.Uint16(data[offset : offset+2])
			colorValue |= 0x8000 // Set the alpha bit (make it opaque)
			offset += 2

			// Set the pixel in the bitmap
			bitmapX := startX + x
			bitmapY := y + 22
			img.Set(bitmapX, bitmapY, bitmap.ARGB1555Color(colorValue))
		}
	}

	return img, nil
}

// decodeStaticArt converts raw static art data into an image.Image.
// Static art has a header with dimensions, followed by a lookup table and
// run-length encoded pixel data.
func decodeStaticArt(data []byte) (image.Image, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("%w: static art data too short for header", ErrInvalidArtData)
	}

	// Skip the 4 byte header
	offset := 4

	// Read dimensions
	width := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	height := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Sanity check on dimensions
	if width <= 0 || height <= 0 || width > 1024 || height > 1024 {
		return nil, fmt.Errorf("%w: invalid dimensions %dx%d", ErrInvalidArtData, width, height)
	}

	// Read lookup table (array of uint16 offsets, one per line)
	lookupTable := make([]int, height)
	for i := 0; i < height; i++ {
		if offset+1 >= len(data) {
			return nil, fmt.Errorf("%w: static art lookup table truncated", ErrInvalidArtData)
		}
		lookupTable[i] = int(binary.LittleEndian.Uint16(data[offset : offset+2]))
		offset += 2
	}

	// Create image
	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))

	// Process each line using the lookup table
	for y := 0; y < height; y++ {
		// Get the offset for this line
		lineOffset := 8 + (height * 2) + lookupTable[y]*2 // Multiply by 2 because offsets are in words

		if lineOffset >= len(data) {
			return nil, fmt.Errorf("%w: invalid line offset at y=%d", ErrInvalidArtData, y)
		}

		// Process the run-length encoded line data
		x := 0
		for x < width {
			if lineOffset+1 >= len(data) {
				return nil, fmt.Errorf("%w: static art data truncated at y=%d, x=%d", ErrInvalidArtData, y, x)
			}

			// Read the run header value
			xOffset := int(binary.LittleEndian.Uint16(data[lineOffset : lineOffset+2]))
			lineOffset += 2

			if xOffset+x >= width {
				break // Safety check
			}

			x += xOffset // Skip transparent pixels

			if lineOffset+1 >= len(data) {
				return nil, fmt.Errorf("%w: static art data truncated at y=%d, x=%d", ErrInvalidArtData, y, x)
			}

			// Read the run length
			runLength := int(binary.LittleEndian.Uint16(data[lineOffset : lineOffset+2]))
			lineOffset += 2

			if runLength > width-x {
				runLength = width - x // Safety check
			}

			// Read the pixel data for this run
			for i := 0; i < runLength; i++ {
				if lineOffset+1 >= len(data) {
					return nil, fmt.Errorf("%w: static art data truncated at y=%d, x=%d", ErrInvalidArtData, y, x+i)
				}

				// Read the color value and set the pixel
				colorValue := binary.LittleEndian.Uint16(data[lineOffset : lineOffset+2])
				colorValue ^= 0x8000 // Flip the alpha bit (from 0=transparent to 1=opaque)

				lineOffset += 2
				img.Set(x+i, y, bitmap.ARGB1555Color(colorValue))
			}

			x += runLength
		}
	}

	return img, nil
}
