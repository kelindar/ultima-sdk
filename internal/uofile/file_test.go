// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package uofile

import (
	"os"
	"path/filepath"
	"testing"

	uotest "github.com/kelindar/ultima-sdk/internal/testing"
	"github.com/stretchr/testify/assert"
)

// TestFile tests the File type and its options.
func TestFile(t *testing.T) {
	t.Run("WithRealMUL", func(t *testing.T) {
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
			entry, err := file.Entry(i)
			assert.NoError(t, err, "Failed to read entry %d", i)
			assert.NotNil(t, entry, "Entry %d should not be nil", i)
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
	}) // end WithRealMUL
} // end TestFile

// fileExists checks if a file exists and is not a directory
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func TestAnimationNameByBody(t *testing.T) {
	assert.Equal(t, "ogres_ogre (1)", AnimationNameByBody(1), "Body 1 should return correct name")
	assert.Equal(t, "", AnimationNameByBody(99999), "Unknown body should return empty string")
}
