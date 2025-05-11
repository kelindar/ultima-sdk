package ultima

import (
	"fmt"
	"os"
	"sync"
)

// SDK represents the main entry point for accessing Ultima Online game files.
// It holds the necessary state, such as the base path to the game files and
// a cache of opened file handles.
type SDK struct {
	basePath string   // Path to the Ultima Online client directory
	files    sync.Map // Lazily loaded file handles (cacheKey to *uofile.File)
}

// Open initializes a new SDK instance for the specified Ultima Online client directory.
// It verifies that the provided path exists and is a directory.
//
// The 'directory' parameter should be the path to the root of the Ultima Online
// installation directory where files like 'art.mul', 'map0.mul', etc., are located.
func Open(directory string) (*SDK, error) {
	info, err := os.Stat(directory)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("ultima: client directory '%s' does not exist: %w", directory, err)
		}
		return nil, fmt.Errorf("ultima: failed to access client directory '%s': %w", directory, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("ultima: provided path '%s' is not a directory", directory)
	}

	sdk := &SDK{
		basePath: directory,
	}

	// In later tasks, this is where initial file indexing or loading might be triggered,
	// similar to C#'s Files.Initialize(). For now, we just store the path.

	return sdk, nil
}

// Close releases any resources held by the SDK instance.
// This is analogous to C#'s `Files.Dispose()` and should be called when the SDK
// is no longer needed to free up file handles or other resources.
func (s *SDK) Close() error {
	// Close all open file handles
	s.closeAllFiles()

	// Clear the base path to indicate the SDK is "closed"
	s.basePath = ""

	return nil
}

// BasePath returns the base directory path provided when the SDK was opened.
// This is primarily for internal use or testing to confirm the SDK was initialized correctly.
func (s *SDK) BasePath() string {
	return s.basePath
}
