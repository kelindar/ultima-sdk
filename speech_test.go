// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpeech(t *testing.T) {
	runWith(t, func(sdk *SDK) {

		t.Run("SpeechEntry", func(t *testing.T) {
			// Test retrieval of a specific speech entry by ID
			// Note: using ID 7 as an example ID, adjust based on what's in your speech.mul test data
			entry, err := sdk.SpeechEntry(1000)

			assert.NoError(t, err)
			assert.NotNil(t, entry)
			assert.Equal(t, 168, entry.ID())
			assert.Equal(t, "*wo ist nuje", entry.Text())
		})

		t.Run("SpeechEntry_Invalid", func(t *testing.T) {
			_, err := sdk.SpeechEntry(-1)
			assert.Error(t, err)
		})

		t.Run("SpeechEntries", func(t *testing.T) {
			// Test iterating over all entries
			var entries []Speech
			for entry := range sdk.SpeechEntries() {
				entries = append(entries, entry)
			}

			assert.NotEmpty(t, entries)

			// Speech entries should have valid data
			for _, entry := range entries {
				assert.NotNil(t, entry)
				// Entry ID should be >= 0 if file format is correct
				assert.GreaterOrEqual(t, entry.ID(), 0)
			}
		})
	})
}
