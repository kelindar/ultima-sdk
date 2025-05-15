package anim

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoadAnimdata_Empty verifies that LoadAnimdata handles an empty file without error and results in empty Animdata.
func TestLoadAnimdata_Empty(t *testing.T) {
	tmp := filepath.Join(os.TempDir(), "animdata_empty.mul")
	defer os.Remove(tmp)
	err := os.WriteFile(tmp, nil, 0644)
	assert.NoError(t, err)

	err = LoadAnimdata(tmp)
	assert.NoError(t, err)
	assert.Empty(t, Animdata)
}

// TestLoadAnimdata_Simple tests loading a single-chunk animdata file with one valid entry.
func TestLoadAnimdata_Simple(t *testing.T) {
	var buf bytes.Buffer
	// Write chunk header
	assert.NoError(t, binary.Write(&buf, binary.LittleEndian, int32(0)))
	// Write 8 entries
	for i := 0; i < 8; i++ {
		// FrameData
		for j := 0; j < 64; j++ {
			if i == 0 && j == 10 {
				buf.WriteByte(0xFB)
			} else {
				buf.WriteByte(0)
			}
		}
		// Metadata: Unknown, FrameCount, FrameInterval, FrameStart
		if i == 0 {
			buf.Write([]byte{1, 1, 2, 3})
		} else {
			buf.Write([]byte{0, 0, 0, 0})
		}
	}

	tmp := filepath.Join(os.TempDir(), "animdata_simple.mul")
	defer os.Remove(tmp)
	assert.NoError(t, os.WriteFile(tmp, buf.Bytes(), 0644))

	// Load animdata
	err := LoadAnimdata(tmp)
	assert.NoError(t, err)

	// Verify only entry 0 is loaded
	entry, ok := Animdata[0]
	assert.True(t, ok)
	assert.Equal(t, int8(-5), entry.FrameData[10])
	assert.Equal(t, uint8(1), entry.Unknown)
	assert.Equal(t, uint8(1), entry.FrameCount)
	assert.Equal(t, uint8(2), entry.FrameInterval)
	assert.Equal(t, uint8(3), entry.FrameStart)

	// Verify entries 1-7 are not present
	for id := 1; id < 8; id++ {
		_, ok := Animdata[id]
		assert.False(t, ok)
	}
}
