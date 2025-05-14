package ultima

// wavHeader returns a standard PCM WAV header for mono, 16-bit, 22050Hz audio.
func wavHeader(dataLen int) []byte {
	const (
		sampleRate   = uint32(22050)
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
