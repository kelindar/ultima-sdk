package anim

// AnimdataEntry holds metadata for a single animation (from animdata.mul)
type AnimdataEntry struct {
	FrameData     [64]int8
	Unknown       uint8
	FrameCount    uint8
	FrameInterval uint8
	FrameStart    uint8
}

// Animdata holds all loaded animation metadata.
var Animdata = map[int]*AnimdataEntry{}

// LoadAnimdata loads animdata.mul into memory (stub).
func LoadAnimdata(path string) error {
	// TODO: Implement animdata.mul parsing
	return nil
}
