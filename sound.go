package ultima

// No extra imports needed

// Sound represents a sound entry loaded from sound.mul.
type Sound struct {
	Index  int    // Sound index
	Length int    // Length of the sound data (bytes)
	Name   string // Name from MUL header
	Data   []byte // Raw PCM/WAV data (with WAV header)
}

// Sound returns a sound by index.
func (s *SDK) Sound(index int) (*Sound, error) {
	idx := index & 0x3FFF
	file, err := s.loadSound()
	if err != nil {
		return nil, err
	}
	data, _, err := file.Read(uint32(idx))
	if err != nil || len(data) <= 32 {
		return nil, nil
	}

	// Extract name from first 32 bytes (null-terminated ASCII)
	nameBytes := data[:32]
	name := string(nameBytes)
	if i := indexOfNull(nameBytes); i >= 0 {
		name = string(nameBytes[:i])
	}

	// Skip 32-byte MUL header, prepend WAV header
	pcm := data[32:]
	wav := wavHeader(len(pcm))
	wav = append(wav, pcm...)

	return &Sound{
		Index:  idx,
		Length: len(pcm),
		Name:   name,
		Data:   wav,
	}, nil
}

// indexOfNull returns the index of the first null byte, or -1 if not found
func indexOfNull(b []byte) int {
	for i, v := range b {
		if v == 0 {
			return i
		}
	}
	return -1
}

// TODO: Translation/removed support via Sound.def (not implemented, file not found)


// Sounds returns an iterator over all available sounds.
func (s *SDK) Sounds() func(yield func(*Sound) bool) {
	return func(yield func(*Sound) bool) {
		for i := 0; i < 0x1000; i++ {
			snd, err := s.Sound(i)
			if err != nil || snd == nil {
				continue
			}
			if !yield(snd) {
				break
			}
		}
	}
}
