// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSound_Load(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		// Test loading a known sound (index 0)
		snd, err := sdk.Sound(0)
		assert.NoError(t, err)
		assert.NotNil(t, snd)
		assert.Greater(t, snd.Length, 0)
		assert.NotEmpty(t, snd.Data)

		// Test loading an out-of-bounds sound (should return nil, no error)
		snd, err = sdk.Sound(0x1000)
		assert.NoError(t, err)
		assert.Nil(t, snd)
	})
}

func TestSound_Iterator(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		count := 0
		for snd := range sdk.Sounds() {
			assert.NotNil(t, snd)
			count++
			if count >= 10 {
				break
			}
		}
		assert.Equal(t, 10, count)
	})
}
