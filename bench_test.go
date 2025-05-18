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
BenchmarkSDK/MapImage-24         	       5	 239619060 ns/op	101894747 B/op	25165935 allocs/op
BenchmarkSDK/MapTiles-24         	      20	  52252060 ns/op	28328275 B/op	 1179852 allocs/op
BenchmarkSDK/GumpImage-24        	 3051325	       397.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkSDK/SpeechIter-24       	    4653	    255869 ns/op	  266456 B/op	   12322 allocs/op
BenchmarkSDK/ClilocIter-24       	     854	   1392080 ns/op	    1219 B/op	      11 allocs/op
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
					runtime.KeepAlive(g.Image)
				}
			}
		})

		b.Run("SpeechIter", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				count := 0
				for entry := range sdk.SpeechEntries() {
					if entry.ID() >= 0 && len(entry.Text()) > 0 {
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

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkRadarcol-24    	43445518	        25.18 ns/op	      16 B/op	       1 allocs/op
*/
func BenchmarkRadarcol(b *testing.B) {
	benchWith(b, func(sdk *SDK) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sdk.RadarColor(123)
		}
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
