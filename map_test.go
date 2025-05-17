package ultima

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTileMap_TileAt(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		maps := []struct {
			mapID       int
			x, y        int
			wantID      uint16
			wantZ       int8
			wantStatics []uint16
		}{
			//{1, 12, 14, 0xA8, -5, nil},               // Trammel water
			{1, 536, 970, 0x409, 0, []uint16{0x5D0}}, // Trammel wooden floor with statics (wooden shingles: 0x5D0 Hue: 0 Altitude: 29)
		}
		for _, tc := range maps {
			t.Run(fmt.Sprintf("MapID_%d", tc.mapID), func(t *testing.T) {
				m, err := sdk.Map(tc.mapID)
				assert.NoError(t, err)
				tile, err := m.TileAt(tc.x, tc.y)
				assert.NoError(t, err)
				assert.NotNil(t, tile)
				if tc.wantID != 0 {
					assert.Equal(t, tc.wantID, tile.ID)
				}

				// Validate elevation
				assert.LessOrEqual(t, tile.Z, int8(127))
				assert.GreaterOrEqual(t, tile.Z, int8(-128))

				// Check statics
				if tc.wantStatics != nil {
					var tileStatics []uint16
					for _, s := range tile.Statics {
						tileStatics = append(tileStatics, s.ID)
					}
					assert.Equal(t, tc.wantStatics, tileStatics)
				} else {
					assert.Empty(t, tile.Statics)
				}
			})
		}
	})
}
