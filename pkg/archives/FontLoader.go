package archives

import (
	"github.com/tpetrychyn/osrs-cache-parser/pkg/cachestore"
	"github.com/tpetrychyn/osrs-cache-parser/pkg/cachestore/fs"
	"github.com/tpetrychyn/osrs-cache-parser/pkg/models"
	"github.com/tpetrychyn/osrs-cache-parser/pkg/utils"
	"log"
)

type FontLoader struct {
	store *cachestore.Store
}

func NewFontLoader(store *cachestore.Store) *FontLoader {
	return &FontLoader{store: store}
}

func (f *FontLoader) LoadFonts() map[string]*models.FontDef {
	fontArchive := f.store.FindIndex(models.IndexType.Fonts)
	spriteArchive := f.store.FindIndex(models.IndexType.Sprites)

	fonts := make(map[string]*models.FontDef, len(models.FontNames))
	for _, fontName := range models.FontNames {
		hash := utils.RSHashString(fontName)
		var sg *cachestore.Group
		for _, s := range spriteArchive.Groups {
			if s.NameHash != hash {
				continue
			}
			sg = s
		}
		if sg == nil {
			log.Printf("did not find a spriteGroup with hash %d", hash)
			continue
		}

		spriteLoader := NewSpriteArchive(f.store)
		sprites := spriteLoader.LoadGroupId(sg.GroupId)
		if len(sprites) == 0 {
			log.Printf("bad length of sprites")
			continue
		}

		fg := fontArchive.Groups[sg.GroupId]

		data, err := f.store.DecompressGroup(fg, nil)
		if err != nil {
			log.Printf("bad data")
			continue
		}

		files := &fs.GroupFiles{Files: []*fs.FSFile{
			{FileId: sg.GroupId, NameHash: sg.NameHash},
		}}

		files.LoadContents(data)

		spriteGroup := models.SpriteDefsToSpriteGroup(sprites)
		font := models.NewFontDef(files.Files[0].Contents, spriteGroup.XOffsets, spriteGroup.YOffsets, spriteGroup.SpriteWidths, spriteGroup.SpriteHeights, spriteGroup.Pixels)

		fonts[fontName] = font
	}

	return fonts
}
