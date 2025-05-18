// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package ultima

import (
	"runtime"
	"testing"

	uotest "github.com/kelindar/ultima-sdk/internal/testing"
	"github.com/stretchr/testify/require"
)

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkSDK/MapImage-24         	       1	1263445600 ns/op	580756288 B/op	50332215 allocs/op
BenchmarkSDK/MapTiles-24         	      28	  43026179 ns/op	12591062 B/op	  393505 allocs/op
BenchmarkSDK/GumpImage-24        	      12	  96974425 ns/op	106561401 B/op	    4285 allocs/op
BenchmarkSDK/SpeechIter-24       	    9487	    129235 ns/op	  115592 B/op	    6124 allocs/op
BenchmarkSDK/ClilocIter-24       	     820	   1415395 ns/op	     165 B/op	      10 allocs/op
*/
func BenchmarkSDK(b *testing.B) {
	benchWith(b, func(sdk *SDK) {
		b.Run("MapImage", func(b *testing.B) {
			m, err := sdk.Map(0)
			if err != nil {
				b.Fatalf("failed to load map: %v", err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				img, err := m.Image()
				if err != nil || img == nil {
					b.Fatalf("map.Image() error: %v", err)
				}
			}
		})

		b.Run("MapTiles", func(b *testing.B) {
			m, err := sdk.Map(0)
			if err != nil {
				b.Fatalf("failed to load map: %v", err)
			}
			width, height := m.width, m.height
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				count := 0
				for y := 0; y < height; y += 8 {
					for x := 0; x < width; x += 8 {
						tile, err := m.TileAt(x, y)
						if err == nil && tile != nil {
							count++
						}
					}
				}
				runtime.KeepAlive(count)
			}
		})

		b.Run("GumpImage", func(b *testing.B) {
			gumps := []*Gump{}
			for g := range sdk.Gumps() {
				gumps = append(gumps, g)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, g := range gumps {
					img, err := g.Image()
					if err != nil || img == nil {
						b.Fatalf("gump.Image() error: %v", err)
					}
					runtime.KeepAlive(img)
				}
			}
		})

		b.Run("SpeechIter", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				count := 0
				for entry := range sdk.SpeechEntries() {
					if entry.ID >= 0 && len(entry.Text) > 0 {
						count++
					}
				}
				runtime.KeepAlive(count)
			}
		})

		b.Run("ClilocIter", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				count := 0
				for id, text := range sdk.Strings() {
					if id >= 0 && len(text) > 0 {
						count++
					}
				}
				runtime.KeepAlive(count)
			}
		})
	})
}

// Helper for running benchmarks with SDK setup/teardown.
func benchWith(b *testing.B, fn func(sdk *SDK)) {
	b.Helper()
	sdk, err := Open(uotest.Path())
	require.NoError(b, err)
	defer sdk.Close()
	fn(sdk)
}
