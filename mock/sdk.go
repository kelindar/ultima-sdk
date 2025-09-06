package mock

import (
	"errors"
	"image"

	"iter"

	"github.com/kelindar/ultima-sdk"
)

var ErrNotFound = errors.New("not found")

// SDK is a lightweight in-memory implementation of the ultima.Interface.
type SDK struct {
	LandsMap       map[int]*ultima.Land
	ItemsMap       map[int]*ultima.Item
	GumpsMap       map[int]*ultima.Gump
	HuesMap        map[int]*ultima.Hue
	LightsMap      map[int]ultima.Light
	RadarColorMap  map[int]ultima.RadarColor
	SkillsMap      map[int]*ultima.Skill
	SkillGroupsMap map[int]*ultima.SkillGroup
	SoundsMap      map[int]*ultima.Sound
	TexturesMap    map[int]*ultima.Texture
	SpeechMap      map[int]ultima.Speech
	StringsMap     map[string]map[int]string
	MultisMap      map[int]*ultima.Multi
	MapsMap        map[int]*TileMap
	UnicodeFont    ultima.Font
	Fonts          []ultima.Font
	nextStringID   int
}

// New creates an empty mock SDK.
func New() *SDK {
	return &SDK{
		LandsMap:       make(map[int]*ultima.Land),
		ItemsMap:       make(map[int]*ultima.Item),
		GumpsMap:       make(map[int]*ultima.Gump),
		HuesMap:        make(map[int]*ultima.Hue),
		LightsMap:      make(map[int]ultima.Light),
		RadarColorMap:  make(map[int]ultima.RadarColor),
		SkillsMap:      make(map[int]*ultima.Skill),
		SkillGroupsMap: make(map[int]*ultima.SkillGroup),
		SoundsMap:      make(map[int]*ultima.Sound),
		TexturesMap:    make(map[int]*ultima.Texture),
		SpeechMap:      make(map[int]ultima.Speech),
		StringsMap:     make(map[string]map[int]string),
		MultisMap:      make(map[int]*ultima.Multi),
		MapsMap:        make(map[int]*TileMap),
	}
}

// Open mirrors ultima.Open but simply returns an empty SDK.
func Open(_ string) (*SDK, error) { return New(), nil }

// Add registers the given value into the mock SDK.
func (s *SDK) Add(v any) {
	switch x := v.(type) {
	case *ultima.Land:
		s.LandsMap[x.ID] = x
	case *ultima.Item:
		s.ItemsMap[x.ID] = x
	case *ultima.Gump:
		s.GumpsMap[x.ID] = x
	case *ultima.Hue:
		s.HuesMap[x.Index] = x
	case ultima.Light:
		s.LightsMap[x.ID] = x
	case *ultima.Light:
		s.LightsMap[x.ID] = *x
	case ultima.RadarColor:
		s.RadarColorMap[x.ID()] = x
	case *ultima.Skill:
		s.SkillsMap[x.ID] = x
	case *ultima.SkillGroup:
		s.SkillGroupsMap[x.ID] = x
	case *ultima.Sound:
		s.SoundsMap[x.Index] = x
	case *ultima.Texture:
		s.TexturesMap[x.Index] = x
	case ultima.Speech:
		s.SpeechMap[x.ID()] = x
	case LocalizedString:
		m, ok := s.StringsMap[x.Lang]
		if !ok {
			m = make(map[int]string)
			s.StringsMap[x.Lang] = m
		}
		m[x.ID] = x.Text
		if x.ID >= s.nextStringID {
			s.nextStringID = x.ID + 1
		}
	case string:
		m, ok := s.StringsMap["enu"]
		if !ok {
			m = make(map[int]string)
			s.StringsMap["enu"] = m
		}
		m[s.nextStringID] = x
		s.nextStringID++
	case *TileMap:
		s.MapsMap[x.ID] = x
	case MultiEntry:
		s.MultisMap[x.ID] = x.Multi
	}
}

// Close is a no-op for the mock SDK.
func (*SDK) Close() error { return nil }

// BasePath returns an empty string.
func (*SDK) BasePath() string { return "" }

// Animation returns a stored animation if present.
func (s *SDK) Animation(body, action, direction, hue int, preserveHue, firstFrame bool) (*ultima.Animation, error) {
	return nil, errors.New("mock: animation not implemented")
}

// Land returns a stored land tile.
func (s *SDK) Land(id int) (*ultima.Land, error) {
	v, ok := s.LandsMap[id]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

// Item returns a stored item tile.
func (s *SDK) Item(id int) (*ultima.Item, error) {
	v, ok := s.ItemsMap[id]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

// Lands iterates over stored lands.
func (s *SDK) Lands() iter.Seq[*ultima.Land] {
	return func(yield func(*ultima.Land) bool) {
		for _, v := range s.LandsMap {
			if !yield(v) {
				break
			}
		}
	}
}

// Items iterates over stored items.
func (s *SDK) Items() iter.Seq[*ultima.Item] {
	return func(yield func(*ultima.Item) bool) {
		for _, v := range s.ItemsMap {
			if !yield(v) {
				break
			}
		}
	}
}

// String retrieves a string in default language "enu".
func (s *SDK) String(id int) (string, error) {
	return s.StringWithLang(id, "enu")
}

// StringWithLang retrieves a localized string.
func (s *SDK) StringWithLang(id int, lang string) (string, error) {
	if m, ok := s.StringsMap[lang]; ok {
		if v, ok := m[id]; ok {
			return v, nil
		}
	}
	return "", ErrNotFound
}

// StringEntry returns a StringEntry for the given id and lang.
func (s *SDK) StringEntry(id int, lang string) (ultima.StringEntry, error) {
	txt, err := s.StringWithLang(id, lang)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 5+len(txt))
	b[4] = 0 // flag
	copy(b[5:], txt)
	return ultima.StringEntry(b), nil
}

// Strings iterates over English strings.
func (s *SDK) Strings() iter.Seq2[int, string] {
	return s.StringsWithLang("enu")
}

// StringsWithLang iterates over strings for a language.
func (s *SDK) StringsWithLang(lang string) iter.Seq2[int, string] {
	m := s.StringsMap[lang]
	return func(yield func(int, string) bool) {
		for id, txt := range m {
			if !yield(id, txt) {
				break
			}
		}
	}
}

func (s *SDK) FontUnicode(int) (ultima.Font, error) {
	if s.UnicodeFont == nil {
		return nil, ErrNotFound
	}
	return s.UnicodeFont, nil
}

func (s *SDK) Font() ([]ultima.Font, error) {
	if len(s.Fonts) == 0 {
		return nil, ErrNotFound
	}
	return s.Fonts, nil
}

func (s *SDK) Gump(id int) (*ultima.Gump, error) {
	v, ok := s.GumpsMap[id]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (s *SDK) Gumps() iter.Seq[*ultima.Gump] {
	return func(yield func(*ultima.Gump) bool) {
		for _, g := range s.GumpsMap {
			if !yield(g) {
				break
			}
		}
	}
}

func (s *SDK) Hue(index int) (*ultima.Hue, error) {
	v, ok := s.HuesMap[index]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (s *SDK) Hues() iter.Seq[*ultima.Hue] {
	return func(yield func(*ultima.Hue) bool) {
		for _, h := range s.HuesMap {
			if !yield(h) {
				break
			}
		}
	}
}

func (s *SDK) Light(id int) (ultima.Light, error) {
	v, ok := s.LightsMap[id]
	if !ok {
		return ultima.Light{}, ErrNotFound
	}
	return v, nil
}

func (s *SDK) Lights() iter.Seq[ultima.Light] {
	return func(yield func(ultima.Light) bool) {
		for _, l := range s.LightsMap {
			if !yield(l) {
				break
			}
		}
	}
}

func (s *SDK) Map(mapID int) (*TileMap, error) {
	v, ok := s.MapsMap[mapID]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (s *SDK) Multi(id int) (*ultima.Multi, error) {
	v, ok := s.MultisMap[id]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (s *SDK) MultiFromCSV(data []byte) (*ultima.Multi, error) {
	return nil, errors.New("mock: MultiFromCSV not implemented")
}

func (s *SDK) RadarColor(tileID int) (ultima.RadarColor, error) {
	v, ok := s.RadarColorMap[tileID]
	if !ok {
		return 0, ErrNotFound
	}
	return v, nil
}

func (s *SDK) RadarColors() iter.Seq[ultima.RadarColor] {
	return func(yield func(ultima.RadarColor) bool) {
		for _, c := range s.RadarColorMap {
			if !yield(c) {
				break
			}
		}
	}
}

func (s *SDK) Skill(id int) (*ultima.Skill, error) {
	v, ok := s.SkillsMap[id]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (s *SDK) Skills() iter.Seq[*ultima.Skill] {
	return func(yield func(*ultima.Skill) bool) {
		for _, sk := range s.SkillsMap {
			if !yield(sk) {
				break
			}
		}
	}
}

func (s *SDK) SkillGroup(id int) (*ultima.SkillGroup, error) {
	v, ok := s.SkillGroupsMap[id]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (s *SDK) SkillGroups() iter.Seq[*ultima.SkillGroup] {
	return func(yield func(*ultima.SkillGroup) bool) {
		for _, sg := range s.SkillGroupsMap {
			if !yield(sg) {
				break
			}
		}
	}
}

func (s *SDK) Sound(index int) (*ultima.Sound, error) {
	v, ok := s.SoundsMap[index]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (s *SDK) Sounds() func(yield func(*ultima.Sound) bool) {
	return func(yield func(*ultima.Sound) bool) {
		for _, snd := range s.SoundsMap {
			if !yield(snd) {
				break
			}
		}
	}
}

func (s *SDK) SpeechEntry(id int) (ultima.Speech, error) {
	v, ok := s.SpeechMap[id]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (s *SDK) SpeechEntries() iter.Seq[ultima.Speech] {
	return func(yield func(ultima.Speech) bool) {
		for _, sp := range s.SpeechMap {
			if !yield(sp) {
				break
			}
		}
	}
}

func (s *SDK) Texture(index int) (*ultima.Texture, error) {
	v, ok := s.TexturesMap[index]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (s *SDK) Textures() func(yield func(*ultima.Texture) bool) {
	return func(yield func(*ultima.Texture) bool) {
		for _, t := range s.TexturesMap {
			if !yield(t) {
				break
			}
		}
	}
}

// ---------------------- TileMap ----------------------

// TileMap is a minimal in-memory implementation of ultima.TileMap.
type TileMap struct {
	ID     int
	Width  int
	Height int
	Tiles  map[[2]int]*ultima.Tile
}

// TileAt retrieves a stored tile.
func (m *TileMap) TileAt(x, y int) (*ultima.Tile, error) {
	if m == nil {
		return nil, errors.New("nil tilemap")
	}
	if t, ok := m.Tiles[[2]int{x, y}]; ok {
		return t, nil
	}
	return nil, ErrNotFound
}

// Image returns a blank image of the correct size.
func (m *TileMap) Image() (image.Image, error) {
	if m == nil {
		return nil, errors.New("nil tilemap")
	}
	return image.NewRGBA(image.Rect(0, 0, m.Width, m.Height)), nil
}

// ---------------------- Helpers ----------------------

type LocalizedString struct {
	Lang string
	ID   int
	Text string
}

type MultiEntry struct {
	ID    int
	Multi *ultima.Multi
}
