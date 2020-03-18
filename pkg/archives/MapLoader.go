package archives

import (
	"bytes"
	"fmt"
	"github.com/tpetrychyn/osrs-cache-parser/pkg/cachestore"
	"github.com/tpetrychyn/osrs-cache-parser/pkg/models"
	"github.com/tpetrychyn/osrs-cache-parser/pkg/utils"
)

const X = 64
const Y = 64
const Height = 4

const BlockedTile = 0x1
const BridgeTile = 0x2
const RoofTile = 0x4

type MapLoader struct {
	store *cachestore.Store
}

func NewMapLoader(store *cachestore.Store) *MapLoader {
	return &MapLoader{store: store}
}

// returns two maps - blocked and bridges
// the key is x-y-z offset from 0,0 of region
// to get true world coord will need to add regionBase
func (m *MapLoader) LoadBlockedTiles(regionId int) ([]*models.Tile, []*models.Tile) {
	blockedTiles := make([]*models.Tile, 0)
	bridgeTiles := make([]*models.Tile, 0)

	index := m.store.FindIndex(models.IndexType.Maps)

	x := regionId >> 8
	z := regionId & 0xFF
	var mapArchive *cachestore.Group
	for _, v := range index.Groups {
		nameHash := utils.Djb2(fmt.Sprintf("m%d_%d", x, z))
		if nameHash == v.NameHash {
			mapArchive = v
			continue
		}
	}
	if mapArchive == nil {
		return blockedTiles, bridgeTiles
	}

	data, err := m.store.DecompressGroup(mapArchive, nil)
	if err != nil {
		return blockedTiles, bridgeTiles
	}
	buf := bytes.NewBuffer(data)

	for z := 0; z < Height; z++ {
		for lx := 0; lx < X; lx++ {
			for ly := 0; ly < Y; ly++ {
				tile := &InternalTile{}
				for {
					attribute, _ := buf.ReadByte()
					if attribute == 0 {
						break
					} else if attribute == 1 {
						height, _ := buf.ReadByte()
						tile.Height = height
						break
					} else if attribute <= 49 {
						tile.AttrOpcode = attribute
						tile.OverlayId, _ = buf.ReadByte()
						tile.OverlayPath = (attribute - 2) / 4
						tile.OverlayRotation = (attribute - 2) & 3
					} else if attribute <= 81 {
						tile.Settings = attribute - 49
					} else {
						tile.UnderlayId = attribute - 82
					}
				}

				baseX, baseY := ((regionId>>8)&0xFF)<<6, (regionId&0xFF)<<6
				x, y := baseX+lx, baseY + ly

				if tile.Settings&BlockedTile == BlockedTile {
					blockedTiles = append(blockedTiles, &models.Tile{
						X:      x,
						Y:      y,
						Height: z,
					})
				}

				if tile.Settings&BridgeTile == BridgeTile {
					bridgeTiles = append(bridgeTiles, &models.Tile{
						X:      x,
						Y:      y,
						Height: z,
					})
				}
			}
		}
	}

	return blockedTiles, bridgeTiles
}

// FIXME: leaving internal tile stuff here incase it is needed in future
type InternalTile struct {
	Height          byte
	AttrOpcode      byte
	Settings        byte
	OverlayId       byte
	OverlayPath     byte
	OverlayRotation byte
	UnderlayId      byte
}
