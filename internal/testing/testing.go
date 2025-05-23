// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package testing

import "runtime"

// Path returns the path to the test data directory based on the operating system.
func Path() string {
	var path string
	switch runtime.GOOS {
	case "windows":
		path = `d:\Workspace\Go\src\github.com\kelindar\ultima-sdk-testdata`
	case "linux":
		path = `/mnt/d/Workspace/Go/src/github.com/kelindar/ultima-sdk-testdata`
	}
	return path
}
