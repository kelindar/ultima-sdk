package ultima

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"

	"iter"

	"github.com/kelindar/ultima-sdk/internal/mul"
)

// Light represents a light source image.
type Light struct {
	ID     int    // ID of the light
	Width  int    // Width of the light image
	Height int    // Height of the light image
	image  []byte // Raw image data
}

// Image returns the light image as a grayscale image.
func (l *Light) Image() image.Image {

	// The light.mul data contains signed byte values that represent light intensity
	// In C#, each byte is converted to a 16-bit ARGB1555 value where R=G=B = 0x1F + value
	img := image.NewGray16(image.Rect(0, 0, l.Width, l.Height))

	for y := 0; y < l.Height; y++ {
		for x := 0; x < l.Width; x++ {
			offset := y*l.Width + x
			if offset >= len(l.image) {
				break
			}

			// Convert signed byte to intensity
			value := int8(l.image[offset])
			intensity := uint16(0x1F + value)

			// Scale intensity (0-31) to 16-bit grayscale range (0-65535)
			scaledIntensity := uint16((float64(intensity) / 31.0) * 65535.0)
			img.SetGray16(x, y, color.Gray16{Y: scaledIntensity})
		}
	}
	return img
}

// Light retrieves a specific light image by ID.
func (s *SDK) Light(id int) (Light, error) {
	if id < 0 {
		return Light{}, fmt.Errorf("invalid light ID: %d", id)
	}

	file, err := s.loadLights()
	if err != nil {
		return Light{}, err
	}

	// Read the actual data
	data, extra, err := file.Read(uint32(id))
	if err != nil {
		return Light{}, fmt.Errorf("error reading light ID %d: %w", id, err)
	}

	return makeLight(uint32(id), data, uint32(extra))
}

// Lights returns an iterator over all defined light images.
func (s *SDK) Lights() iter.Seq[Light] {
	file, err := s.loadLights()
	if err != nil {
		return func(yield func(Light) bool) {}
	}

	return func(yield func(Light) bool) {
		for index := range file.Entries() {
			data, extra, err := file.Read(index)
			if err != nil {
				continue
			}

			light, err := makeLight(index, data, uint32(extra))
			if err != nil {
				continue
			}

			if !yield(light) {
				return
			}
		}
	}
}

// GetRawLight returns the raw byte data for a light image.
// This is analogous to the C# Light.GetRawLight method.
func (s *SDK) GetRawLight(id int) ([]byte, int, int, error) {
	if id < 0 {
		return nil, 0, 0, fmt.Errorf("invalid light ID: %d", id)
	}

	file, err := s.loadLights()
	if err != nil {
		return nil, 0, 0, err
	}

	// Read the actual data
	data, extra, err := file.Read(uint32(id))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("error reading light ID %d: %w", id, err)
	}

	width, height := lightSize(uint32(extra))
	return data, width, height, nil
}

func lightSize(extra uint32) (int, int) {
	width := int(extra & 0xFFFF)
	height := int((extra >> 16) & 0xFFFF)
	return width, height
}

// makeLight processes the raw byte data from the MUL file into a Light struct.
// The 'extra' value from the index file contains width (lower 16 bits) and height (upper 16 bits).
func makeLight(id uint32, data []byte, extra uint32) (Light, error) {
	width, height := lightSize(extra)
	if width <= 0 || height <= 0 {
		return Light{}, fmt.Errorf("invalid dimensions for light ID %d: width=%d, height=%d", id, width, height)
	}

	if len(data) < width*height {
		return Light{}, fmt.Errorf("data length mismatch for light ID %d: expected at least %d, got %d", id, width*height, len(data))
	}

	return Light{
		ID:     int(id),
		Width:  width,
		Height: height,
		image:  data,
	}, nil
}

// decodeLightFile is a wrapper for the internal light decoder that matches the uofile.DecodeFn signature
func decodeLightFile(file *os.File, add mul.AddFn) error {
	// We're processing just the idx file, which should be the first arg
	// We need to find and open the mul file separately
	idxPath := file.Name()
	if len(idxPath) < 10 || idxPath[len(idxPath)-10:] != "lightidx.mul" {
		return fmt.Errorf("expected lightidx.mul, got %s", idxPath)
	}

	mulPath := idxPath[:len(idxPath)-10] + "light.mul"
	mulFile, err := os.Open(mulPath)
	if err != nil {
		// Can't open the mul file, we'll just process the idx
		return decodeLightFileInternal(file, nil, add)
	}
	defer mulFile.Close()

	return decodeLightFileInternal(file, mulFile, add)
}

// decodeLightFileInternal is the actual implementation that processes both idx and mul files
func decodeLightFileInternal(idxFile *os.File, mulFile *os.File, add mul.AddFn) error {
	idxStat, err := idxFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat lightidx.mul: %w", err)
	}

	numEntries := idxStat.Size() / 12 // Each index entry is 12 bytes
	if numEntries == 0 {
		return nil // No entries
	}

	for i := int64(0); i < numEntries; i++ {
		// Read index entry: lookup, length, extra (width & height)
		var lookup, length, extra uint32

		if err := binary.Read(idxFile, binary.LittleEndian, &lookup); err != nil {
			return fmt.Errorf("failed to read lookup for light index entry %d: %w", i, err)
		}

		if err := binary.Read(idxFile, binary.LittleEndian, &length); err != nil {
			return fmt.Errorf("failed to read length for light index entry %d: %w", i, err)
		}

		if err := binary.Read(idxFile, binary.LittleEndian, &extra); err != nil {
			return fmt.Errorf("failed to read extra data for light index entry %d: %w", i, err)
		}

		// Skip invalid entries (lookup -1 or length 0)
		if lookup == 0xFFFFFFFF || length == 0 {
			add(uint32(i), lookup, length, extra, nil)
			continue
		}

		// Read the actual image data from light.mul
		if mulFile == nil {
			add(uint32(i), lookup, length, extra, nil)
			continue
		}

		// Seek to the position in the mul file
		if _, err := mulFile.Seek(int64(lookup), io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek in light.mul for entry %d: %w", i, err)
		}

		// Read the data
		data := make([]byte, length)
		if _, err := io.ReadFull(mulFile, data); err != nil {
			return fmt.Errorf("failed to read data from light.mul for entry %d: %w", i, err)
		}

		// Add the entry to the file index
		add(uint32(i), lookup, length, extra, data)
	}

	return nil
}
