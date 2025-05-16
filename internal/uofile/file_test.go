// Package uofile provides a unified interface for accessing Ultima Online data files,
// supporting both MUL and UOP file formats.
package uofile

import (
	"os"
	"path/filepath"
	"testing"

	uotest "github.com/kelindar/ultima-sdk/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFile_WithRealMUL tests the file reader with real skills.mul/skills.idx files
func TestFile_WithRealMUL(t *testing.T) {
	testdataPath := uotest.Path()

	// Skip if test data directory is not available
	if testdataPath == "" {
		t.Skip("Test data directory not found")
	}

	mulPath := filepath.Join(testdataPath, "skills.mul")
	idxPath := filepath.Join(testdataPath, "skills.idx")

	// Skip if the specific test files aren't available
	if !fileExists(mulPath) || !fileExists(idxPath) {
		t.Skipf("Test files not found: %s, %s", mulPath, idxPath)
	}

	// Create a File instance with the real files
	fileNames := []string{"skills.mul", "skills.idx"}
	file := New(testdataPath, fileNames, 0)
	defer file.Close()

	// Test initialization
	err := file.open()
	assert.NoError(t, err, "Failed to initialize with real files")

	// Test reading entries
	for i := uint32(0); i < 10; i++ {
		data, _, err := file.Read(i)
		if err == nil {
			// If we found an entry, make sure it has some content
			assert.NotEmpty(t, data, "Entry %d should have content", i)
		}
	}

	// Test the Entries iterator
	var count int
	file.Entries()(func(idx uint32) bool {
		count++
		// Just count the first 50 entries max to keep the test quick
		return count < 50
	})

	// skills.mul should have entries
	assert.Greater(t, count, 0, "Expected to find some entries in skills.mul")
}

// fileExists checks if a file exists and is not a directory
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

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

	// Create a File instance with automatic format detection
	fileNames := []string{"test.mul", "test.idx"}
	file := New(tempDir, fileNames, 1)
	defer file.Close()

	// Test reading the entry
	data, _, err := file.Read(0)
	assert.NoError(t, err)
	assert.Equal(t, mulData, data)

	// Test entries iterator
	var foundEntries []uint32
	file.Entries()(func(idx uint32) bool {
		foundEntries = append(foundEntries, idx)
		return true
	})
	assert.Equal(t, []uint32{0}, foundEntries)
}

// TestFile_InitAndClose tests the initialization and closure state transitions
func TestFile_InitAndClose(t *testing.T) {
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

	// Create a File instance with automatic format detection
	fileNames := []string{"test.mul", "testidx.mul"}
	file := New(tempDir, fileNames, 1)

	// Initial state should be stateNew (0)
	assert.Equal(t, int32(stateNew), file.state.Load())

	// Reading should trigger initialization
	_, _, err = file.Read(0)
	// We expect an error since our test files are empty, but that's not what we're testing
	assert.Error(t, err)

	// Close the file
	err = file.Close()
	assert.NoError(t, err)

	// State should now be stateClosed
	assert.Equal(t, int32(stateClosed), file.state.Load())

	// Read after close should return ErrReaderClosed
	_, _, err = file.Read(0)
	assert.ErrorIs(t, err, ErrReaderClosed)

	// Closing again should be a no-op
	err = file.Close()
	assert.NoError(t, err)
}

// TestFile_Options tests the options for file configuration
func TestFile_Options(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir := t.TempDir()

	// Create a UOP file for testing
	uopPath := filepath.Join(tempDir, "test.uop")
	uopFile, err := os.Create(uopPath)
	require.NoError(t, err)
	uopFile.Close()

	// Test WithCount option
	fileNames := []string{"test.uop"}
	file1 := New(tempDir, fileNames, 0, WithCount(1000))
	assert.Equal(t, 1000, file1.length)
	file1.Close()

	// Test WithIndexLength option
	file2 := New(tempDir, fileNames, 0, WithIndexLength(500))
	assert.Len(t, file2.uopOpts, 1) // Should have one option
	file2.Close()

	// Test WithExtra option
	file3 := New(tempDir, fileNames, 0, WithExtra())
	assert.Len(t, file3.uopOpts, 1) // Should have one option
	file3.Close()

	// Test multiple options together
	file4 := New(tempDir, fileNames, 0,
		WithCount(1000),
		WithIndexLength(500),
		WithExtra())
	assert.Equal(t, 1000, file4.length)
	assert.Len(t, file4.uopOpts, 2) // Should have two options
	file4.Close()
}

// TestFile_DetectFormat tests format detection logic
func TestFile_DetectFormat(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir := t.TempDir()

	// Test with non-existent files
	file1 := New(tempDir, []string{"nonexistent.uop"}, 1)
	assert.NotEmpty(t, file1.path) // Should have a path set to the placeholder
	assert.NotNil(t, file1.initFn) // Should have an initialization function

	// Create a MUL file without IDX
	mulPath := filepath.Join(tempDir, "test.mul")
	mulFile, err := os.Create(mulPath)
	require.NoError(t, err)
	mulFile.Close()

	// Test with only MUL file
	file2 := New(tempDir, []string{"test.mul"}, 1)
	assert.Equal(t, mulPath, file2.path)
	assert.Empty(t, file2.idxPath)
	assert.NotNil(t, file2.initFn)

	// Create an IDX file
	idxPath := filepath.Join(tempDir, "test.idx")
	idxFile, err := os.Create(idxPath)
	require.NoError(t, err)
	idxFile.Close()

	// Test with both MUL and IDX files
	file3 := New(tempDir, []string{"test.mul", "test.idx"}, 1)
	assert.Equal(t, mulPath, file3.path)
	assert.Equal(t, idxPath, file3.idxPath)
	assert.NotNil(t, file3.initFn)

	// Create a UOP file
	uopPath := filepath.Join(tempDir, "test.uop")
	uopFile, err := os.Create(uopPath)
	require.NoError(t, err)
	uopFile.Close()

	// Test with UOP file (should prioritize UOP)
	file4 := New(tempDir, []string{"test.mul", "test.idx", "test.uop"}, 1)
	assert.Equal(t, uopPath, file4.path)
	assert.NotNil(t, file4.initFn)
}

// TestFile_Concurrency tests concurrent access to a file
func TestFile_Concurrency(t *testing.T) {
	tempDir := t.TempDir()
	mulPath := filepath.Join(tempDir, "test.mul")
	idxPath := filepath.Join(tempDir, "test.idx")

	mulFile, err := os.Create(mulPath)
	require.NoError(t, err)
	mulFile.Close()

	idxFile, err := os.Create(idxPath)
	require.NoError(t, err)
	idxFile.Close()

	// Create a File instance
	fileNames := []string{"test.mul", "test.idx"}
	file := New(tempDir, fileNames, 1)
	defer file.Close()

	// Run multiple goroutines that try to initialize and read from the file
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			// Try to read (will force initialization)
			_, _, _ = file.Read(0)
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}

	// Now try to close it while also having reads
	closing := make(chan bool)
	reading := make(chan bool)

	// Start a goroutine to close the file
	go func() {
		err := file.Close()
		assert.NoError(t, err)
		closing <- true
	}()

	// Start goroutines that try to read
	for i := 0; i < 5; i++ {
		go func() {
			file.Read(0)
			reading <- true
		}()
	}

	// Wait for all operations to complete
	<-closing
	for i := 0; i < 5; i++ {
		<-reading
	}

	// Final state should be closed
	assert.Equal(t, int32(stateClosed), file.state.Load())
}

func TestAnimationNameByBody(t *testing.T) {
	// Should find a known body (adjust the expected name as per your file_anim.json)
	assert.Equal(t, "ogres_ogre (1)", AnimationNameByBody(1), "Body 1 should return correct name")

	// Should return empty string for unknown body
	assert.Equal(t, "", AnimationNameByBody(99999), "Unknown body should return empty string")
}

