// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMulti_Load(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		multi, err := sdk.Multi(0)
		assert.NoError(t, err)
		assert.NotNil(t, multi)
		assert.Greater(t, len(multi.Items), 0)
		item := multi.Items[0]
		assert.NotZero(t, item.ItemID)

		img, err := multi.Image()
		assert.NoError(t, err)
		assert.NotNil(t, img)
		//savePng(img, "multi.png")
	})
}

func savePng(img image.Image, name string) error {
	file, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	return png.Encode(file, img)
}
