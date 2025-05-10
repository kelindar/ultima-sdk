// Package ultima provides access to Ultima Online data files
package ultima

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestWith runs a test with a properly initialized SDK instance using the test data directory.
// The function ensures the SDK is initialized with valid test data and passes it to the test function.
func TestWith(t *testing.T, testFn func(*testing.T, *SDK)) {
	var path string
	switch runtime.GOOS {
	case "windows":
		path = `d:\Workspace\Go\src\github.com\kelindar\ultima-sdk-testdata`
	case "linux":
		path = `/mnt/d/Workspace/Go/src/github.com/kelindar/ultima-sdk-testdata`
	}
	// Open the SDK with the test data directory
	sdk, err := Open(path)
	require.NoError(t, err, "failed to open SDK with test data directory")
	require.NotNil(t, sdk, "SDK instance should not be nil")

	// Run the test with the SDK instance
	testFn(t, sdk)
}
