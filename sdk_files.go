package ultima

import (
	"fmt"

	"github.com/kelindar/ultima-sdk/internal/uofile"
)

// cacheKey represents a string key for caching files
type cacheKey string

// Art loads the art files (art.mul, artidx.mul or their UOP equivalent)
func (s *SDK) Art() (*uofile.File, error) {
	return s.load([]string{
		"artLegacyMUL.uop",
		"art.mul",
	}, 0x10000, uofile.WithIndexLength(12)) // Art has 0x10000 (65536) entries
}

// Gump loads the gump files (gumpart.mul, gumpidx.mul or their UOP equivalent)
func (s *SDK) Gump() (*uofile.File, error) {
	return s.load([]string{
		"gumpartLegacyMUL.uop",
		"gumpart.mul",
	}, 0xFFFF, uofile.WithIndexLength(12)) // Gumps have 0xFFFF (65535) entries maximum
}

// Map loads a specific map file (mapX.mul, where X is the map ID)
func (s *SDK) Map(mapID int) (*uofile.File, error) {
	return s.load([]string{
		fmt.Sprintf("map%dLegacyMUL.uop", mapID),
		fmt.Sprintf("map%d.mul", mapID),
	}, 0, uofile.WithIndexLength(12))
}

// Statics loads the statics files for a specific map ID
func (s *SDK) Statics(mapID int) (*uofile.File, error) {
	// Statics files come in pairs: statics*.mul and staidx*.mul
	return s.load([]string{
		fmt.Sprintf("statics%d.mul", mapID),
	}, 0, uofile.WithIndexLength(12))
}

// StaticsIndex loads the statics index files for a specific map ID
func (s *SDK) StaticsIndex(mapID int) (*uofile.File, error) {
	return s.load([]string{
		fmt.Sprintf("staidx%d.mul", mapID),
	}, 0, uofile.WithIndexLength(12))
}

// Hues loads the hues file
func (s *SDK) Hues() (*uofile.File, error) {
	return s.load([]string{
		"hues.mul",
	}, 3000, uofile.WithIndexLength(12)) // 3000 hue entries
}

// Sound loads the sound files
func (s *SDK) Sound() (*uofile.File, error) {
	return s.load([]string{
		"soundLegacyMUL.uop",
		"sound.mul",
	}, 0, uofile.WithIndexLength(12))
}

// SoundIndex loads the sound index file
func (s *SDK) SoundIndex() (*uofile.File, error) {
	return s.load([]string{
		"soundidx.mul",
	}, 0, uofile.WithIndexLength(12))
}

// TileData loads the tiledata file
func (s *SDK) TileData() (*uofile.File, error) {
	return s.load([]string{
		"tiledata.mul",
	}, 0, uofile.WithIndexLength(12))
}

// Texture loads the texture files
func (s *SDK) Texture() (*uofile.File, error) {
	return s.load([]string{
		"texmaps.mul",
	}, 0x4000, uofile.WithIndexLength(12)) // 0x4000 (16384) entries
}

// TextureIndex loads the texture index file
func (s *SDK) TextureIndex() (*uofile.File, error) {
	return s.load([]string{
		"texidx.mul",
	}, 0x4000, uofile.WithIndexLength(12))
}

// Light loads the light files
func (s *SDK) Light() (*uofile.File, error) {
	return s.load([]string{
		"light.mul",
	}, 0, uofile.WithIndexLength(12))
}

// LightIndex loads the light index file
func (s *SDK) LightIndex() (*uofile.File, error) {
	return s.load([]string{
		"lightidx.mul",
	}, 0, uofile.WithIndexLength(12))
}

// Multi loads the multi files
func (s *SDK) Multi() (*uofile.File, error) {
	return s.load([]string{
		"housing.bin", // UOP format
		"multi.mul",   // MUL format
	}, 0, uofile.WithIndexLength(12))
}

// MultiIndex loads the multi index file
func (s *SDK) MultiIndex() (*uofile.File, error) {
	return s.load([]string{
		"multi.idx",
	}, 0, uofile.WithIndexLength(12))
}

// Verdata loads the verdata file which contains patches
func (s *SDK) Verdata() (*uofile.File, error) {
	return s.load([]string{
		"verdata.mul",
	}, 0, uofile.WithIndexLength(12))
}

// Skills loads the skills file
func (s *SDK) Skills() (*uofile.File, error) {
	return s.load([]string{
		"skills.mul",
	}, 0, uofile.WithIndexLength(12))
}

// SkillsIndex loads the skills index file
func (s *SDK) SkillsIndex() (*uofile.File, error) {
	return s.load([]string{
		"skills.idx",
	}, 0, uofile.WithIndexLength(12))
}

// Animation loads the animation files for a specific file type
// fileType can be 1 for anim.mul, 2 for anim2.mul, etc.
func (s *SDK) Animation(fileType int) (*uofile.File, error) {
	return s.load([]string{
		fmt.Sprintf("anim%d.mul", fileType),
	}, 0, uofile.WithIndexLength(12))
}

// AnimationIndex loads the animation index files for a specific file type
// fileType can be 1 for anim.idx, 2 for anim2.idx, etc.
func (s *SDK) AnimationIndex(fileType int) (*uofile.File, error) {
	return s.load([]string{
		fmt.Sprintf("anim%d.idx", fileType),
	}, 0, uofile.WithIndexLength(12))
}

// load loads a file with the given file names and length
// It tries to find the file in cache first, if not found, it creates a new file handle and caches it
// The fileNames parameter should contain possible filenames to look for (e.g., both mul and uop variants)
// length represents the expected number of entries in the file
// options are passed to the underlying uofile.File creation
func (s *SDK) load(fileNames []string, length int, options ...uofile.Option) (*uofile.File, error) {
	// Create a cache key from the first filename (canonical name)
	key := cacheKey(fileNames[0])

	// Try to get from cache first
	if f, ok := s.files.Load(key); ok {
		return f.(*uofile.File), nil
	}

	// Not in cache, create new file
	file := uofile.New(s.basePath, fileNames, length, options...)

	// Store in cache (use LoadOrStore to handle potential race conditions)
	actual, loaded := s.files.LoadOrStore(key, file)
	if loaded {
		// Another goroutine beat us to it, close our file and use the cached one
		file.Close()
		return actual.(*uofile.File), nil
	}

	return file, nil
}

// fileExists checks if a specific file exists in the basePath
func (s *SDK) fileExists(fileName string) bool {
	file := uofile.New(s.basePath, []string{fileName}, 0)
	defer file.Close()

	// Try to read entry 0, if it fails with a specific error that
	// tells us the file doesn't exist, otherwise we assume it exists
	_, err := file.Read(0)
	return err == nil || err != uofile.ErrInvalidFormat
}

// closeAllFiles closes all open file handles
func (s *SDK) closeAllFiles() {
	s.files.Range(func(key, value interface{}) bool {
		if file, ok := value.(*uofile.File); ok {
			file.Close()
		}
		s.files.Delete(key)
		return true
	})
}
