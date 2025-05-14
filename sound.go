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

// TODO: Translation/removed support via Sound.def (not implemented)

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

// wavHeader returns a standard PCM WAV header for mono, 16-bit, 22050Hz audio.
func wavHeader(dataLen int) []byte {
	const (
		sampleRate    = uint32(22050)
		bitsPerSample = uint16(16)
		channels      = uint16(1)
	)
	blockAlign := channels * bitsPerSample / 8
	byteRate := sampleRate * uint32(blockAlign)
	chunkSize := uint32(36 + dataLen)
	dataLen32 := uint32(dataLen)

	header := make([]byte, 44)
	copy(header[0:], []byte("RIFF"))
	header[4] = byte(chunkSize)
	header[5] = byte(chunkSize >> 8)
	header[6] = byte(chunkSize >> 16)
	header[7] = byte(chunkSize >> 24)
	copy(header[8:], []byte("WAVEfmt "))
	header[16] = 16 // Subchunk1Size for PCM
	header[20] = 1  // AudioFormat PCM
	header[22] = byte(channels)
	// sampleRate (uint32, little-endian)
	// sampleRate (uint32, little-endian)
	header[24] = byte(sampleRate & 0xFF)
	header[25] = byte((sampleRate >> 8) & 0xFF)
	header[26] = byte((sampleRate >> 16) & 0xFF)
	header[27] = byte((sampleRate >> 24) & 0xFF)
	// byteRate (uint32, little-endian)
	header[28] = byte(byteRate)
	header[29] = byte(byteRate >> 8)
	header[30] = byte(byteRate >> 16)
	header[31] = byte(byteRate >> 24)
	// blockAlign (uint16, little-endian)
	header[32] = byte(blockAlign)
	header[33] = byte(blockAlign >> 8)
	// bitsPerSample (uint16, little-endian)
	header[34] = byte(bitsPerSample)
	header[35] = byte(bitsPerSample >> 8)
	copy(header[36:], []byte("data"))
	// dataLen (uint32, little-endian)
	header[40] = byte(dataLen32)
	header[41] = byte(dataLen32 >> 8)
	header[42] = byte(dataLen32 >> 16)
	header[43] = byte(dataLen32 >> 24)
	return header
}
