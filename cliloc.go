package ultima

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"iter"

	"github.com/kelindar/ultima-sdk/internal/mul"
)

var (
	// ErrInvalidStringID is returned when an invalid string ID is requested
	ErrInvalidStringID = errors.New("invalid string ID")
)

// StringEntry represents a single localized string entry from a cliloc file.
// It corresponds to the C# Ultima.StringEntry class.
type StringEntry struct {
	ID   int    // The numeric ID of the string.
	Flag byte   // A flag associated with the string (e.g., custom, modified).
	Text string // The localized text content. Placeholders like ~1_NAME~ are preserved.
}

// makeStringEntry converts raw byte data into a StringEntry
func makeStringEntry(data []byte) StringEntry {
	// First 4 bytes: ID (int32)
	// Next 1 byte: Flag
	// Remaining bytes: Text
	return StringEntry{
		ID:   int(binary.LittleEndian.Uint32(data[0:4])),
		Flag: data[4],
		Text: string(data[5:]),
	}
}

// String retrieves a localized string by its ID using the default language ("enu").
// If the ID is not found, an error is returned.
func (s *SDK) String(id int) (string, error) {
	return s.StringWithLang(id, "enu")
}

// StringWithLang retrieves a localized string by its ID using the specified language.
// If the ID is not found or the language file doesn't exist, an error is returned.
func (s *SDK) StringWithLang(id int, lang string) (string, error) {
	entry, err := s.StringEntry(id, lang)
	if err != nil {
		return "", err
	}
	return entry.Text, nil
}

// StringEntry retrieves a string entry by its ID using the default language ("enu").
func (s *SDK) StringEntry(id int, lang string) (StringEntry, error) {
	file, err := s.loadCliloc(lang)
	if err != nil {
		return StringEntry{}, err
	}

	data, err := file.Read(uint32(id))
	if err != nil {
		return StringEntry{}, fmt.Errorf("%w: %d", ErrInvalidStringID, id)
	}

	return makeStringEntry(data), nil
}

// Strings returns an iterator over all localized strings in the default language ("enu").
func (s *SDK) Strings() iter.Seq2[int, string] {
	return s.StringsWithLang("enu")
}

// StringsWithLang returns an iterator over all localized strings in the specified language.
func (s *SDK) StringsWithLang(lang string) iter.Seq2[int, string] {
	file, err := s.loadCliloc(lang)
	if err != nil {
		return func(yield func(int, string) bool) {} // Empty iterator
	}

	return func(yield func(int, string) bool) {
		for index := range file.Entries() {
			data, err := file.Read(index)
			if err != nil {
				continue
			}

			entry := makeStringEntry(data)
			if !yield(entry.ID, entry.Text) {
				break
			}
		}
	}
}

// decodeClilocFile loads all string entries from a cliloc file into mul.Entry3D
//
// The cliloc file format:
// - Header1 (int32, LittleEndian) - typically 0xFFFFFFFF
// - Header2 (int16, LittleEndian) - typically 0x0000
// For each entry:
//   - ID (int32, LittleEndian)
//   - Flag (byte)
//   - Length (int16, LittleEndian)
//   - Text (bytes[Length], UTF-8 encoded)
func decodeClilocFile(file *os.File, add mul.AddFn) error {
	reader := bufio.NewReader(file)

	// Read file headers
	var header1 int32
	if err := binary.Read(reader, binary.LittleEndian, &header1); err != nil {
		return fmt.Errorf("failed to read cliloc header1: %w", err)
	}

	var header2 int16
	if err := binary.Read(reader, binary.LittleEndian, &header2); err != nil {
		return fmt.Errorf("failed to read cliloc header2: %w", err)
	}

	for {
		// Read ID (4 bytes)
		var id int32
		if err := binary.Read(reader, binary.LittleEndian, &id); err != nil {
			if err == io.EOF {
				break // End of file, normal termination
			}
			return fmt.Errorf("failed to read string entry ID: %w", err)
		}

		// Read Flag (1 byte)
		flag, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break // End of file, normal termination
			}
			return fmt.Errorf("failed to read string entry flag for ID %d: %w", id, err)
		}

		// Read Length (2 bytes)
		var length int16
		if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
			return fmt.Errorf("failed to read string entry length for ID %d: %w", id, err)
		}

		if length < 0 {
			return fmt.Errorf("invalid (negative) string entry length %d for ID %d", length, id)
		}

		// Read the string data
		var text []byte
		if length > 0 {
			text = make([]byte, length)
			if _, err := io.ReadFull(reader, text); err != nil {
				return fmt.Errorf("failed to read string entry text for ID %d (length %d): %w", id, length, err)
			}
		}

		// Create a single byte buffer containing all of the data in our format:
		// ID (4 bytes) + Flag (1 byte) + Text (variable)
		buffer := new(bytes.Buffer)
		binary.Write(buffer, binary.LittleEndian, id)
		buffer.WriteByte(flag)
		buffer.Write(text)

		// Store by both index and ID for retrieval
		data := buffer.Bytes()
		add(uint32(id), uint32(id), uint32(len(data)), 0, data) // Also store indexed by string ID
	}

	return nil
}
