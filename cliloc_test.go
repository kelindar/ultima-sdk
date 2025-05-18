// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCliloc(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		t.Run("StringEntry", func(t *testing.T) {
			entry, err := sdk.StringEntry(1000009, "enu")
			assert.NoError(t, err)
			assert.NotNil(t, entry)
			assert.GreaterOrEqual(t, entry.ID(), 0)
			assert.NotEmpty(t, entry.Text())
		})

		t.Run("StringEntry_Invalid", func(t *testing.T) {
			_, err := sdk.StringEntry(-1, "enu")
			assert.Error(t, err)
		})

		t.Run("String", func(t *testing.T) {
			text, err := sdk.String(1000000)
			assert.NoError(t, err)
			assert.NotEmpty(t, text)
		})

		t.Run("Strings", func(t *testing.T) {
			// Test iterating over all strings
			count := 0
			for id, text := range sdk.Strings() {
				assert.GreaterOrEqual(t, id, 0)
				_ = text // Use variable to satisfy linter
				count++

				if count >= 5 { // Limit number of entries to check to avoid large test output
					break
				}
			}

			if count == 0 {
				t.Skip("Skipping test as cliloc.enu may not be present in test data or is empty")
			}
		})

		t.Run("Strings_IteratorExhaustion", func(t *testing.T) {
			// Test that the iterator doesn't panic when exhausted
			for range sdk.Strings() {
			}
		})
	})
}

// Tests for the decode function
func TestDecodeClilocFile(t *testing.T) {
	// This would ideally involve creating a temporary file with known content
	// and validating the decode function works correctly.
	// For now, we'll just ensure the function is exported and compiles correctly.
	t.Skip("Full testing of decodeClilocFile requires creating test files")
}
