package ultima

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"iter"

	"github.com/kelindar/ultima-sdk/internal/mul"
)

var (
	// ErrInvalidSpeechID is returned when an invalid speech ID is requested
	ErrInvalidSpeechID = errors.New("invalid speech ID")
)

const (
	maxSpeechTextLength = 128
)

// Speech represents a single speech entry from speech.mul
type Speech struct {
	ID   int    // ID of the speech entry (from file)
	Text string // Text content of the speech entry
}

// SpeechEntry retrieves a predefined speech entry by its ID
func (s *SDK) SpeechEntry(id int) (Speech, error) {
	file, err := s.loadSpeech()
	if err != nil {
		return Speech{}, err
	}

	text, err := file.Read(uint64(id))
	if err != nil {
		return Speech{}, err
	}

	return Speech{ID: id, Text: string(text)}, nil
}

// SpeechEntries returns an iterator over all defined speech entries
func (s *SDK) SpeechEntries() iter.Seq[Speech] {
	file, err := s.loadSpeech()
	if err != nil {
		return nil
	}

	return func(yield func(Speech) bool) {
		for id := range file.Entries() {
			text, err := file.Read(uint64(id))
			if err != nil {
				continue
			}

			if !yield(Speech{
				ID:   int(id),
				Text: string(text),
			}) {
				break
			}
		}

	}
}

// decodeSpeechFile loads all speech entries from speech.mul into mul.Entry3D
//
// The speech.mul file format:
// For each entry:
//   - ID (int16, BigEndian)
//   - Length (int16, BigEndian)
//   - Text (bytes[Length], Windows-1252 encoded)
func decodeSpeechFile(reader *os.File) ([]mul.Entry3D, error) {
	entries := make([]mul.Entry3D, 0, 6500)
	buffer := make([]byte, maxSpeechTextLength)

	// Read entries until EOF
	for {
		var id int16
		var length int16

		// Read ID (BigEndian)
		err := binary.Read(reader, binary.BigEndian, &id)
		if err == io.EOF {
			break // End of file, normal termination
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read speech ID: %w", err)
		}

		// Read Length (BigEndian)
		err = binary.Read(reader, binary.BigEndian, &length)
		if err != nil {
			return nil, fmt.Errorf("failed to read length for speech ID %d: %w", id, err)
		}

		textLength := int(length)
		if textLength > maxSpeechTextLength {
			textLength = maxSpeechTextLength
		}

		var text string
		if textLength > 0 {
			n, err := io.ReadFull(reader, buffer[:textLength])
			if err != nil && n != textLength {
				return nil, fmt.Errorf("failed to read text for speech ID %d: %w", id, err)
			}

			text = string(buffer[:n])
		}

		entries = append(entries,
			mul.NewEntry(uint32(id), uint32(textLength), 0, []byte(text)),
		)
	}

	return entries, nil
}
