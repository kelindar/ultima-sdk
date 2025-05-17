// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package uofile

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed file_anim.json
var fileAnimJSON []byte

// AnimationEntry represents a single animation entry from file_anim.json
type AnimationEntry struct {
	Name string `json:"name"`
	Body int    `json:"body"`
	Type int    `json:"type"`
}

// AnimationList is the root structure for file_anim.json
type AnimationList struct {
	Mobs []AnimationEntry `json:"Mobs"`
}

var animNameByBody map[int]string

func init() {
	var animList AnimationList
	if err := json.Unmarshal(fileAnimJSON, &animList); err != nil {
		panic(fmt.Errorf("failed to parse embedded file_anim.json: %w", err))
	}
	animNameByBody = make(map[int]string, len(animList.Mobs))
	for _, mob := range animList.Mobs {
		animNameByBody[mob.Body] = mob.Name
	}
}

// AnimationNameByBody returns the animation name for a body ID, or "" if not found.
func AnimationNameByBody(body int) string {
	return animNameByBody[body]
}
