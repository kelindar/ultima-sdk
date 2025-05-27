# Ultima SDK

<p align="center">
  <img width="330" height="110" src=".github/logo.png" border="0">
  <br>
  <img src="https://img.shields.io/github/go-mod/go-version/kelindar/ultima-sdk" alt="Go Version">
  <a href="https://pkg.go.dev/github.com/kelindar/ultima-sdk"><img src="https://pkg.go.dev/badge/github.com/kelindar/ultima-sdk" alt="PkgGoDev"></a>
  <a href="https://goreportcard.com/report/github.com/kelindar/ultima-sdk"><img src="https://goreportcard.com/badge/github.com/kelindar/ultima-sdk" alt="Go Report Card"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
</p>

A modern, idiomatic Go SDK for reading and manipulating Ultima Online client files (MUL/UOP), with robust error handling and high performance. Read-only at the moment.

## Features

- Read art, animations, maps, gumps, hues, fonts, and localization (cliloc)
- Supports MUL and UOP formats (where applicable)
- Idiomatic Go iterators for collections
- No global mutable state, thread-safe design

## Installation

```sh
go get github.com/kelindar/ultima-sdk
```

## Quick Example

```go
package main

import (
	"fmt"

	"github.com/kelindar/ultima-sdk"
)

func main() {
	sdk, err := ultima.Open("/path/to/UO/client")
	if err != nil {
		panic(err)
	}

	str, _ := sdk.String(3000001)
	fmt.Println(str)
}
```

## API Reference

### Core SDK Operations

- `Open(dir string) (*SDK, error)` – Open a UO client directory
- `(*SDK).Close() error` – Close SDK and release resources
- `(*SDK).BasePath() string` – Get the base directory path

### Animation

- `(*SDK).Animation(body, action, direction, hue int, preserveHue, firstFrame bool) (*Animation, error)` – Load animation frames

### Localization (Cliloc)

- `(*SDK).String(id int) (string, error)` – Retrieve localized string
- `(*SDK).StringWithLang(id int, lang string) (string, error)` – Retrieve string in specific language
- `(*SDK).StringEntry(id int, lang string) (StringEntry, error)` – Get string entry with metadata
- `(*SDK).Strings() iter.Seq2[int, string]` – Iterate over all strings
- `(*SDK).StringsWithLang(lang string) iter.Seq2[int, string]` – Iterate over strings in specific language

### Fonts

- `(*SDK).Font() ([]Font, error)` – Load ASCII fonts
- `(*SDK).FontUnicode() (Font, error)` – Load Unicode font

### Hues/Colors

- `(*SDK).Hue(index int) (*Hue, error)` – Get hue/color data
- `(*SDK).Hues() iter.Seq[*Hue]` – Iterate over all hues

### Gumps (UI Graphics)

- `(*SDK).Gump(id int) (*Gump, error)` – Load gump images
- `(*SDK).Gumps() iter.Seq[*Gump]` – Iterate over all gumps

### Maps & Tiles

- `(*SDK).Map(mapID int) (*TileMap, error)` – Load map data
- `(*SDK).Land(id int) (*Land, error)` – Load land art tiles
- `(*SDK).Lands() iter.Seq[*Land]` – Iterate over all land tiles
- `(*SDK).Item(id int) (*Item, error)` – Load static art tiles
- `(*SDK).Items() iter.Seq[*Item]` – Iterate over all static items

### Multi-Tile Objects

- `(*SDK).Multi(id int) (*Multi, error)` – Load multi-tile object
- `(*SDK).MultiFromCSV(id int) (*Multi, error)` – Load multi from CSV data

### Radar Colors

- `(*SDK).RadarColor(id int) (RadarColor, error)` – Get radar color
- `(*SDK).RadarColors() iter.Seq[RadarColor]` – Iterate over all radar colors

### Audio

- `(*SDK).Sound(id int) (Sound, error)` – Load sound data
- `(*SDK).Sounds() iter.Seq[Sound]` – Iterate over all sounds
- `(*SDK).SpeechEntry(id int) (Speech, error)` – Get speech entry
- `(*SDK).SpeechEntries() iter.Seq[Speech]` – Iterate over all speech entries

### Skills

- `(*SDK).Skill(id int) (*Skill, error)` – Get skill information
- `(*SDK).Skills() iter.Seq[*Skill]` – Iterate over all skills
- `(*SDK).SkillGroup(id int) (*SkillGroup, error)` – Get skill group
- `(*SDK).SkillGroups() iter.Seq[*SkillGroup]` – Iterate over all skill groups

### Lighting & Textures

- `(*SDK).Light(id int) (Light, error)` – Load light data
- `(*SDK).Lights() iter.Seq[Light]` – Iterate over all lights
- `(*SDK).Texture(id int) (Texture, error)` – Load texture data
- `(*SDK).Textures() iter.Seq[Texture]` – Iterate over all textures

## Contributing

PRs are welcome! Please:

- Follow the established code style and architectural patterns
- Add doc comments for all exported functions
- Avoid package-level mutable state

## License

MIT License. See [LICENSE](LICENSE) for details.
