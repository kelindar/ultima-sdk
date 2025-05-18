// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"iter"

	"codeberg.org/go-mmap/mmap"
	"github.com/kelindar/ultima-sdk/internal/mul"
)

var (
	// ErrInvalidSpeechID is returned when an invalid speech ID is requested
	ErrInvalidSpeechID = errors.New("invalid speech ID")
)

// Speech represents a single speech entry from speech.mul
type Speech []byte

// ID returns the id of the speecn entry
func (s Speech) ID() int {
	return int(binary.BigEndian.Uint16(s[0:2]))
}

// Text returns the associated text
func (s Speech) Text() string {
	return string(s[2:])
}

// SpeechEntry retrieves a predefined speech entry by its ID
func (s *SDK) SpeechEntry(id int) (Speech, error) {
	file, err := s.loadSpeech()
	if err != nil {
		return Speech{}, err
	}

	data, err := file.ReadFull(uint32(id))
	if err != nil {
		return Speech{}, err
	}

	return Speech(data), nil
}

// SpeechEntries returns an iterator over all defined speech entries
func (s *SDK) SpeechEntries() iter.Seq[Speech] {
	file, err := s.loadSpeech()
	if err != nil {
		return nil
	}

	return func(yield func(Speech) bool) {
		for index := range file.Entries() {
			data, err := file.ReadFull(index)
			if err != nil {
				continue
			}

			if !yield(Speech(data)) {
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
//   - Text (bytes[Length], UTF-8 encoded)
func decodeSpeechFile(reader *mmap.File, add mul.AddFn) error {
	const maxlen = 128
	buffer := make([]byte, maxlen)
	for index := uint32(0); ; index++ {
		head := struct {
			ID  int16
			Len int16
		}{}

		// Read header
		err := binary.Read(reader, binary.BigEndian, &head)
		if err == io.EOF {
			break // End of file, normal termination
		}
		if err != nil {
			return fmt.Errorf("failed to read speech ID: %w", err)
		}

		if head.Len = min(maxlen, head.Len); head.Len > 0 {
			n, err := io.ReadFull(reader, buffer[:head.Len])
			if err != nil && n != int(head.Len) {
				return fmt.Errorf("failed to read text for speech ID %d: %w", head.ID, err)
			}
		}

		// Pack the text into a string
		entry := make([]byte, head.Len+2)
		binary.BigEndian.PutUint16(entry[0:2], uint16(head.ID))
		copy(entry[2:], buffer[:head.Len])

		// Add the entry to the index
		add(index, uint32(head.ID), uint32(head.Len), 0, entry)
	}

	return nil
}
