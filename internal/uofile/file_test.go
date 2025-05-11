package uofile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFile_WithMUL tests the MUL file format support
func TestFile_WithMUL(t *testing.T) {
	// Create a temporary MUL and IDX file for testing
	tempDir := t.TempDir()
	mulPath := filepath.Join(tempDir, "test.mul")
	idxPath := filepath.Join(tempDir, "test.idx")

	// Create mock MUL file
	mulFile, err := os.Create(mulPath)
	require.NoError(t, err)
	defer mulFile.Close()

	// Write some test data to the MUL file
	mulData := []byte("This is test data for MUL file")
	_, err = mulFile.Write(mulData)
	require.NoError(t, err)

	// Create mock IDX file with a single entry
	idxFile, err := os.Create(idxPath)
	require.NoError(t, err)
	defer idxFile.Close()

	// Write a single entry to the IDX file (offset=0, length=len(mulData), extra=0)
	idxEntry := make([]byte, 12)
	// Little-endian: offset=0
	// Little-endian: length=len(mulData)
	for i := 0; i < 4; i++ {
		idxEntry[4+i] = byte(len(mulData) >> (8 * i))
	}
	// Little-endian: extra=0 (already 0s)
	_, err = idxFile.Write(idxEntry)
	require.NoError(t, err)

	// Close the files to ensure data is flushed
	mulFile.Close()
	idxFile.Close()

	// Create a File instance with MUL format
	file := New(WithMUL(mulPath, idxPath))
	defer file.Close()

	// Test reading the entry
	data, err := file.Read(0)
	assert.NoError(t, err)
	assert.Equal(t, mulData, data)

	// Test entries iterator
	var foundEntries []uint64
	file.Entries()(func(idx uint64) bool {
		foundEntries = append(foundEntries, idx)
		return true
	})
	assert.Equal(t, []uint64{0}, foundEntries)

	// Test adding a patch
	patchData := []byte("This is patched data")
	file.AddPatch(0, patchData)

	// Check that the patch is applied
	patchedData, err := file.Read(0)
	assert.NoError(t, err)
	assert.Equal(t, patchData, patchedData)

	// Test removing the patch
	file.RemovePatch(0)
	originalData, err := file.Read(0)
	assert.NoError(t, err)
	assert.Equal(t, mulData, originalData)
}

// TestFile_AutoDetect tests the automatic format detection
func TestFile_AutoDetect(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir := t.TempDir()

	// Create a MUL file and IDX file
	mulPath := filepath.Join(tempDir, "test.mul")
	idxPath := filepath.Join(tempDir, "testidx.mul")

	mulFile, err := os.Create(mulPath)
	require.NoError(t, err)
	mulFile.Close()

	idxFile, err := os.Create(idxPath)
	require.NoError(t, err)
	idxFile.Close()

	// Create a File instance with auto detection
	file := New(AutoDetect(tempDir, "test", 1))
	defer file.Close()

	// Verify that MUL format was selected
	assert.Equal(t, FormatMUL, file.Format())
	assert.Equal(t, mulPath, file.Path())
	assert.Equal(t, idxPath, file.IndexPath())

	// Now let's create a UOP file and see if auto-detection prefers it
	uopPath := filepath.Join(tempDir, "test.uop")
	uopFile, err := os.Create(uopPath)
	require.NoError(t, err)
	uopFile.Close()

	// Create another File instance with auto detection
	file2 := New(AutoDetect(tempDir, "test", 1))
	defer file2.Close()

	// Verify that UOP format was selected
	assert.Equal(t, FormatUOP, file2.Format())
	assert.Equal(t, uopPath, file2.Path())
}

// TestFile_WithPatches tests patch application
func TestFile_WithPatches(t *testing.T) {
	// Create a temporary MUL and IDX file for testing
	tempDir := t.TempDir()
	mulPath := filepath.Join(tempDir, "test.mul")
	idxPath := filepath.Join(tempDir, "test.idx")

	// Create files with minimum content
	mulFile, _ := os.Create(mulPath)
	mulFile.Close()
	idxFile, _ := os.Create(idxPath)
	idxFile.Close()

	// Create patches for indices 5 and 10
	patches := map[uint64][]byte{
		5:  []byte("Patch for index 5"),
		10: []byte("Patch for index 10"),
	}

	// Create a File instance with patches
	file := New(
		WithMUL(mulPath, idxPath),
		WithPatches(patches),
	)
	defer file.Close()

	// Test reading patched entries
	data5, err := file.Read(5)
	assert.NoError(t, err)
	assert.Equal(t, patches[5], data5)

	data10, err := file.Read(10)
	assert.NoError(t, err)
	assert.Equal(t, patches[10], data10)

	// Test Entries includes patched entries
	var foundEntries []uint64
	file.Entries()(func(idx uint64) bool {
		foundEntries = append(foundEntries, idx)
		return true
	})

	// Should find at least our patched entries
	assert.Contains(t, foundEntries, uint64(5))
	assert.Contains(t, foundEntries, uint64(10))
}
