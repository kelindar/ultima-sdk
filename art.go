// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"iter"

	"github.com/kelindar/ultima-sdk/internal/bitmap"
	"github.com/kelindar/ultima-sdk/internal/uofile"
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
	ID     int         // ID of the tile
	Name   string      // Name from TileData
	Flags  TileFlag    // Flags from TileData
	Height int8        // Height of the tile, from TileData
	Image  image.Image // Image of the gump
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
	return uofile.Decode(file, uint32(id), func(data []byte, extra uint64) (*ArtTile, error) {
		img, err := decodeLandImage(data)
		if err != nil {
			return nil, err
		}

		return &ArtTile{
			ID:    id,
			Name:  info.Name,
			Flags: info.Flags,
			Image: img,
		}, nil
	})
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
	return uofile.Decode(file, uint32(artID), func(data []byte, extra uint64) (*ArtTile, error) {
		img, err := decodeStaticImage(data)
		if err != nil {
			return nil, err
		}

		return &ArtTile{
			ID:     artID,
			Name:   info.Name,
			Flags:  info.Flags,
			Height: int8(info.Height),
			Image:  img,
		}, nil
	})
}

// LandArtTiles returns an iterator over all available land art tiles.
func (s *SDK) LandArtTiles() iter.Seq[*ArtTile] {
	return func(yield func(*ArtTile) bool) {
		for i := uint32(0); i < landTileMax; i++ {
			tile, err := s.LandArtTile(int(i))
			if tile == nil || err != nil {
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
			if tile == nil || err != nil {
				continue
			}

			if !yield(tile) {
				break
			}
		}
	}
}

// decodeLandImage converts raw land art data into an image.Image.
// Land art is always 44x44 pixels. The format is essentially a run-length
// encoded 44x44 image where each 2-byte value represents a color index.
func decodeLandImage(data []byte) (image.Image, error) {
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

// decodeStaticImage converts raw static art data into an image.Image.
// Static art has a header with dimensions, followed by a lookup table and
// run-length encoded pixel data.
func decodeStaticImage(data []byte) (image.Image, error) {
	if len(data) < 8 { // Header (4) + Width (2) + Height (2)
		return nil, fmt.Errorf("%w: static art data too short for header", ErrInvalidArtData)
	}

	// Skip the 4 byte art entry header
	offset := 4

	// Read dimensions
	width := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	height := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Sanity check on dimensions
	if width <= 0 || height <= 0 || width > 2048 || height > 2048 { // Max typical UO art dim is ~512, 2048 is very safe.
		return nil, fmt.Errorf("%w: invalid dimensions %dx%d", ErrInvalidArtData, width, height)
	}

	// Read lookup table. Each entry is a WORD offset relative to the start of the RLE data block.
	lookupTableValues := make([]int, height)
	lookupTableByteSize := height * 2
	if offset+lookupTableByteSize > len(data) {
		return nil, fmt.Errorf("%w: static art data too short for lookup table (needs %d bytes, has %d remaining from offset %d, total data %d)", ErrInvalidArtData, lookupTableByteSize, len(data)-offset, offset, len(data))
	}
	for i := 0; i < height; i++ {
		lookupTableValues[i] = int(binary.LittleEndian.Uint16(data[offset : offset+2]))
		offset += 2
	}

	// 'offset' is now at the start of the RLE data block.
	// This corresponds to 'start' in the C# reference (UOFiddler Art.cs GetStatic).
	rleDataBlockStartOffset := offset

	img := bitmap.NewARGB1555(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		// Calculate the starting byte offset for this line's RLE data, relative to the beginning of 'data'.
		// lookupTableValues[y] is a WORD offset from rleDataBlockStartOffset.
		lineRleStartOffsetInData := rleDataBlockStartOffset + (lookupTableValues[y] * 2)
		currentReadOffset := lineRleStartOffsetInData

		x := 0 // Current horizontal pixel position in the output image for this line
		for x < width {
			// Ensure we can read xPixelOffset (2 bytes) and runLength (2 bytes) for the RLE pair.
			if currentReadOffset+4 > len(data) {
				if x < width { // If we still expect pixels on this line.
					return nil, fmt.Errorf("%w: static art data truncated before RLE pair header at y=%d, x_cursor=%d. Need 4 bytes from readOffset=%d, dataLen=%d", ErrInvalidArtData, y, x, currentReadOffset, len(data))
				}
				break // Line ends if x >= width or truncated past expected content.
			}

			xPixelOffset := int(binary.LittleEndian.Uint16(data[currentReadOffset : currentReadOffset+2]))
			currentReadOffset += 2
			runLength := int(binary.LittleEndian.Uint16(data[currentReadOffset : currentReadOffset+2]))
			currentReadOffset += 2

			if xPixelOffset == 0 && runLength == 0 {
				break // End of line marker
			}

			x += xPixelOffset // Advance by transparent pixels

			for i := 0; i < runLength; i++ {
				// Ensure we can read 2 bytes for color data.
				if currentReadOffset+2 > len(data) {
					return nil, fmt.Errorf("%w: static art data truncated during pixel data run at y=%d, x_target_pixel=%d (x_cursor_at_run_start=%d, pixel_in_run=%d). Need 2 bytes from readOffset=%d, dataLen=%d. RunLength was %d", ErrInvalidArtData, y, x+i, x, i, runLength, currentReadOffset, len(data))
				}

				colorValue := binary.LittleEndian.Uint16(data[currentReadOffset : currentReadOffset+2])
				colorValue ^= 0x8000 // Flip the alpha bit (UO statics: 0=transparent, 1=opaque for this bit)
				currentReadOffset += 2

				if x+i < width { // Draw only if within image bounds
					img.Set(x+i, y, bitmap.ARGB1555Color(colorValue))
				}
			}
			x += runLength // Advance by opaque pixels drawn/skipped
		}
	}

	return img, nil
}
