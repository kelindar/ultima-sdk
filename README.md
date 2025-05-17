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

## API Highlights

- `Open(dir string) (*SDK, error)` – Open a UO client directory
- `(*SDK).Animation(body, action, direction, hue int, preserveHue, firstFrame bool) (*Animation, error)` – Load animation frames
- `(*SDK).String(id int) (string, error)` – Retrieve localized string
- `(*SDK).Font() ([]*Font, error)` – Load UO fonts
- `(*SDK).Hue(id int) (*Hue, error)` – Get hue/color data
- `(*SDK).Gump(id int) (*Gump, error)` – Load gump images
- `(*SDK).Map(id int) (*Map, error)` – Load map data
- `(*SDK).LandArtTile(id int) (*ArtTile, error)` – Load land art tiles
- `(*SDK).StaticArtTile(id int) (*ArtTile, error)` – Load static art tiles

## Contributing

PRs are welcome! Please:
- Follow the established code style and architectural patterns
- Add doc comments for all exported functions
- Avoid package-level mutable state

## License

MIT License. See [LICENSE](LICENSE) for details.
