package mock

import (
	"testing"

	"github.com/kelindar/ultima-sdk"
	"github.com/stretchr/testify/assert"
)

func TestMockSDK_AddAndRetrieve(t *testing.T) {
	sdk := New()
	land := &ultima.Land{Art: ultima.Art{ID: 1}}
	sdk.Add(land)

	got, err := sdk.Land(1)
	assert.NoError(t, err)
	assert.Equal(t, land, got)

	sdk.Add(LocalizedString{Lang: "enu", ID: 100, Text: "hello"})
	txt, err := sdk.String(100)
	assert.NoError(t, err)
	assert.Equal(t, "hello", txt)

	id := sdk.nextStringID
	sdk.Add("world")
	txt, err = sdk.String(id)
	assert.NoError(t, err)
	assert.Equal(t, "world", txt)
}

// dummyFont implements ultima.Font for testing purposes.
type dummyFont struct{}

func (dummyFont) Rune(r rune) *ultima.Rune { return &ultima.Rune{} }
func (dummyFont) Size(string) (int, int)   { return 0, 0 }

// setup creates a mock SDK populated with sample data covering the full
// exported surface of the SDK type.
func setup() *SDK {
	sdk := New()

	// Core data types
	sdk.Add(&ultima.Land{Art: ultima.Art{ID: 1}})
	sdk.Add(&ultima.Item{Art: ultima.Art{ID: 2}})
	sdk.Add(&ultima.Gump{ID: 3, Width: 1, Height: 2})
	sdk.Add(&ultima.Hue{Index: 4, Name: "h"})
	sdk.Add(ultima.Light{ID: 5, Width: 1, Height: 1})
	sdk.Add(ultima.RadarColor(6 | (0x1234 << 32)))
	sdk.Add(&ultima.Skill{ID: 7, Name: "test"})
	sdk.Add(&ultima.SkillGroup{ID: 8, Name: "grp", Skills: []int{7}})
	sdk.Add(&ultima.Sound{Index: 9, Length: 1, Name: "s"})
	sdk.Add(&ultima.Texture{Index: 10, Size: 64})
	sdk.Add(ultima.Speech([]byte{0, 11, 'h', 'i'}))
	sdk.Add(LocalizedString{Lang: "enu", ID: 12, Text: "hello"})

	// Map and multi structures
	tm := &TileMap{ID: 13, Width: 2, Height: 2, Tiles: map[[2]int]*ultima.Tile{
		{1, 1}: {ID: 0x100},
	}}
	sdk.Add(tm)
	sdk.Add(MultiEntry{ID: 14, Multi: &ultima.Multi{Items: []ultima.MultiItem{{Item: 1}}}})

	// Fonts
	sdk.UnicodeFont = dummyFont{}
	sdk.Fonts = []ultima.Font{dummyFont{}}

	return sdk
}

// TestMockSDK_Methods exercises all exported methods of the mock SDK
// to ensure they behave consistently with the in-memory data.
func TestMockSDK_Methods(t *testing.T) {
	sdk := setup()

	// Animation is not implemented and should error
	_, err := sdk.Animation(0, 0, 0, 0, false, false)
	assert.Error(t, err)

	// Basic lookup methods
	if land, err := sdk.Land(1); assert.NoError(t, err) {
		assert.Equal(t, 1, land.ID)
	}
	if item, err := sdk.Item(2); assert.NoError(t, err) {
		assert.Equal(t, 2, item.ID)
	}
	if gump, err := sdk.Gump(3); assert.NoError(t, err) {
		assert.Equal(t, 3, gump.ID)
	}
	if hue, err := sdk.Hue(4); assert.NoError(t, err) {
		assert.Equal(t, 4, hue.Index)
	}
	if l, err := sdk.Light(5); assert.NoError(t, err) {
		assert.Equal(t, 5, l.ID)
	}
	if rc, err := sdk.RadarColor(6); assert.NoError(t, err) {
		assert.Equal(t, ultima.RadarColor(6|(0x1234<<32)), rc)
	}
	if sk, err := sdk.Skill(7); assert.NoError(t, err) {
		assert.Equal(t, 7, sk.ID)
	}
	if sg, err := sdk.SkillGroup(8); assert.NoError(t, err) {
		assert.Equal(t, 8, sg.ID)
	}
	if snd, err := sdk.Sound(9); assert.NoError(t, err) {
		assert.Equal(t, 9, snd.Index)
	}
	if tex, err := sdk.Texture(10); assert.NoError(t, err) {
		assert.Equal(t, 10, tex.Index)
	}
	if sp, err := sdk.SpeechEntry(11); assert.NoError(t, err) {
		assert.Equal(t, 11, sp.ID())
	}
	if txt, err := sdk.String(12); assert.NoError(t, err) {
		assert.Equal(t, "hello", txt)
	}

	// Iterators should yield the stored item
	count := 0
	for range sdk.Lands() {
		count++
	}
	assert.Equal(t, 1, count)

	count = 0
	for range sdk.Items() {
		count++
	}
	assert.Equal(t, 1, count)

	count = 0
	for range sdk.Gumps() {
		count++
	}
	assert.Equal(t, 1, count)

	count = 0
	for range sdk.Hues() {
		count++
	}
	assert.Equal(t, 1, count)

	count = 0
	for range sdk.Lights() {
		count++
	}
	assert.Equal(t, 1, count)

	count = 0
	for range sdk.RadarColors() {
		count++
	}
	assert.Equal(t, 1, count)

	count = 0
	for range sdk.Skills() {
		count++
	}
	assert.Equal(t, 1, count)

	count = 0
	for range sdk.SkillGroups() {
		count++
	}
	assert.Equal(t, 1, count)

	count = 0
	for range sdk.Sounds() {
		count++
	}
	assert.Equal(t, 1, count)

	count = 0
	for range sdk.SpeechEntries() {
		count++
	}
	assert.Equal(t, 1, count)

	count = 0
	for range sdk.Textures() {
		count++
	}
	assert.Equal(t, 1, count)

	count = 0
	for range sdk.Strings() {
		count++
	}
	assert.Equal(t, 1, count)

	// Map and tile helpers
	if m, err := sdk.Map(13); assert.NoError(t, err) {
		if tile, err := m.TileAt(1, 1); assert.NoError(t, err) {
			assert.Equal(t, uint16(0x100), tile.ID)
		}
		img, err := m.Image()
		assert.NoError(t, err)
		assert.Equal(t, 2, img.Bounds().Dx())
		assert.Equal(t, 2, img.Bounds().Dy())
	}
	if multi, err := sdk.Multi(14); assert.NoError(t, err) {
		assert.Len(t, multi.Items, 1)
	}
	_, err = sdk.MultiFromCSV(nil)
	assert.Error(t, err)

	// Font helpers
	if f, err := sdk.FontUnicode(); assert.NoError(t, err) {
		assert.NotNil(t, f)
	}
	if fonts, err := sdk.Font(); assert.NoError(t, err) {
		assert.Len(t, fonts, 1)
	}

	// BasePath should be empty and Close should succeed
	assert.Equal(t, "", sdk.BasePath())
	assert.NoError(t, sdk.Close())
}

func TestMockSDK_NotFound(t *testing.T) {
	sdk := New()
	_, err := sdk.Land(999)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.Item(999)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.Gump(999)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.Hue(999)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.Light(999)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.RadarColor(999)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.Skill(999)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.SkillGroup(999)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.Sound(999)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.Texture(999)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.SpeechEntry(999)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.Map(999)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.Multi(999)
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.FontUnicode()
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.Font()
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = sdk.String(999)
	assert.ErrorIs(t, err, ErrNotFound)
}
