package ultima

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTileMap_TileAt(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		maps := []struct {
			mapID  int
			x, y   int
			wantID uint16
			wantZ  int8
		}{
			//{0, 1323, 1623, 2, 0}, // Example: Felucca, known grass tile
			{1, 12, 14, 0xA8, -5}, // Trammel water
			//{2, 128, 128, 0, 0},   // Ilshenar
		}
		for _, tc := range maps {
			t.Run(fmt.Sprintf("MapID_%d", tc.mapID), func(t *testing.T) {
				m, err := sdk.Map(tc.mapID)
				require.NoError(t, err)
				tile, err := m.TileAt(tc.x, tc.y)
				require.NoError(t, err)
				require.NotNil(t, tile)
				if tc.wantID != 0 {
					require.Equal(t, tc.wantID, tile.ID)
				}
				require.LessOrEqual(t, tile.Z, int8(127))
				require.GreaterOrEqual(t, tile.Z, int8(-128))
				require.NotNil(t, tile.Statics)
			})
		}
	})

}

func TestTileMap_StaticsFidelity(t *testing.T) {
	runWith(t, func(sdk *SDK) {
		m, err := sdk.Map(0)
		require.NoError(t, err)
		tile, err := m.TileAt(1323, 1623) // Example: Felucca
		require.NoError(t, err)
		statics := tile.Statics
		// Statics should be fully populated from tiledata (flags, name, etc)
		for _, s := range statics {
			require.NotEqual(t, s.Flags, TileFlagNone)
			require.NotEmpty(t, s.Name)
		}
	})
}
