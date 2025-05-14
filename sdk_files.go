package ultima

import (
	"fmt"

	"github.com/kelindar/ultima-sdk/internal/uofile"
)

// cacheKey represents a string key for caching files
type cacheKey string

// loadHues loads the hues file
func (s *SDK) loadHues() (*uofile.File, error) {
	return s.load([]string{"hues.mul"}, 3000, uofile.WithChunks(708))
}

// loadRadarcol loads the radar colors file
func (s *SDK) loadRadarcol() (*uofile.File, error) {
	return s.load([]string{"radarcol.mul"}, totalRadarColors)
}

// loadSkills loads the skills file
func (s *SDK) loadSkills() (*uofile.File, error) {
	return s.load([]string{"skills.mul", "skills.idx"}, 0, uofile.WithIndexLength(16))
}

// loadSkillGroups loads the skill groups file
func (s *SDK) loadSkillGroups() (*uofile.File, error) {
	return s.load([]string{"skillgrp.mul"}, 0)
}

// loadTiledata loads the tiledata file
func (s *SDK) loadTiledata() (*uofile.File, error) {
	return s.load([]string{
		"tiledata.mul",
	}, 0, uofile.WithDecodeMUL(decodeTileDataFile))
}

// loadLights loads the light files
func (s *SDK) loadLights() (*uofile.File, error) {
	return s.load([]string{
		"light.mul",
		"lightidx.mul",
	}, 0, uofile.WithIndexLength(12))
}

// loadArt loads the art.mul/artidx.mul file
func (s *SDK) loadArt() (*uofile.File, error) {
	return s.load([]string{
		"artLegacyMUL.uop",
		"art.mul",
		"artidx.mul",
	}, 0x14000, uofile.WithExtension(".tga"), uofile.WithIndexLength(0x13FDC))
}

// loadGumpart loads the gump files (gumpart.mul or UOP equivalent)
func (s *SDK) loadGump() (*uofile.File, error) {
	return s.load([]string{
		"gumpartLegacyMUL.uop",
		"gumpart.mul",
		"gumpidx.mul",
	}, 0xFFFF, uofile.WithExtension(".tga"), uofile.WithExtra())
}

// loadMap loads a specific map file (mapX.mul, where X is the map ID)
func (s *SDK) loadMap(mapID int) (*uofile.File, error) {
	return s.load([]string{
		fmt.Sprintf("map%dLegacyMUL.uop", mapID),
		fmt.Sprintf("map%d.mul", mapID),
	}, 0, uofile.WithIndexLength(12))
}

// loadStatics loads the statics files for a specific map ID
func (s *SDK) loadStatics(mapID int) (*uofile.File, error) {
	return s.load([]string{
		fmt.Sprintf("statics%d.mul", mapID),
		fmt.Sprintf("staidx%d.mul", mapID),
	}, 0, uofile.WithIndexLength(12))
}

// loadSound loads the sound files
func (s *SDK) loadSound() (*uofile.File, error) {
	return s.load([]string{
		"soundLegacyMUL.uop",
		"sound.mul",
		"soundidx.mul",
	}, 0, uofile.WithIndexLength(12))
}

// loadTextures loads the texture files
func (s *SDK) loadTextures() (*uofile.File, error) {
	return s.load([]string{
		"texmaps.mul",
		"texidx.mul",
	}, 0x4000, uofile.WithIndexLength(12))
}

// loadMulti loads the multi files
func (s *SDK) loadMulti() (*uofile.File, error) {
	return s.load([]string{
		"housing.bin", // UOP format
		"multi.mul",   // MUL format
		"multi.idx",
	}, 0, uofile.WithIndexLength(12))
}

// loadVerdata loads the verdata file which contains patches
func (s *SDK) loadVerdata() (*uofile.File, error) {
	return s.load([]string{
		"verdata.mul",
	}, 0, uofile.WithIndexLength(12))
}

// loadAnim loads the animation files for a specific file type
// fileType can be 1 for anim.mul, 2 for anim2.mul, etc.
func (s *SDK) loadAnim(fileType int) (*uofile.File, error) {
	return s.load([]string{
		fmt.Sprintf("anim%d.mul", fileType),
		fmt.Sprintf("anim%d.idx", fileType),
	}, 0, uofile.WithIndexLength(12))
}

// loadUnicodeFonts loads the Unicode fonts file
func (s *SDK) loadUnicodeFonts() (*uofile.File, error) {
	return s.load([]string{
		"fonts.mul",
	}, 0, uofile.WithIndexLength(12))
}

// loadCliloc loads the client localization file for a specific language
func (s *SDK) loadCliloc(language string) (*uofile.File, error) {
	return s.load([]string{
		fmt.Sprintf("cliloc.%s", language),
	}, 0, uofile.WithDecodeMUL(decodeClilocFile))
}

// loadSpeech loads the speech.mul file
func (s *SDK) loadSpeech() (*uofile.File, error) {
	return s.load([]string{"speech.mul"}, 0, uofile.WithDecodeMUL(decodeSpeechFile))
}

// load loads a file with the given file names and length
// It tries to find the file in cache first, if not found, it creates a new file handle and caches it
// The fileNames parameter should contain possible filenames to look for (e.g., both mul and uop variants)
// length represents the expected number of entries in the file
// options are passed to the underlying uofile.File creation
func (s *SDK) load(fileNames []string, length int, options ...uofile.Option) (*uofile.File, error) {
	key := cacheKey(fileNames[0])
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
	_, _, err := file.Read(0)
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
