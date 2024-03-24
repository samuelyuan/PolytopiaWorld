package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Tile struct {
	Sprites []SpriteData
	Terrain int
}

type SpriteData struct {
	Image   *ebiten.Image
	OffsetX int
	OffsetY int
}

func (t *Tile) AddSprite(spriteData SpriteData) {
	t.Sprites = append(t.Sprites, spriteData)
}

func (t *Tile) Draw(screen *ebiten.Image, options []*ebiten.DrawImageOptions) {
	for i, spriteData := range t.Sprites {
		screen.DrawImage(spriteData.Image, options[i])
	}
}
