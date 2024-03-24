package graphics

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type SpriteSheet struct {
	Land      []*ebiten.Image
	Mountains []*ebiten.Image
	Forests   []*ebiten.Image
	Water     []*ebiten.Image
	Ocean     []*ebiten.Image
	Village   *ebiten.Image
	Ice       *ebiten.Image
}

func loadImage(filename string) *ebiten.Image {
	image, _, err := ebitenutil.NewImageFromFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	return image
}

func LoadSpriteSheet() (*SpriteSheet, error) {
	spriteSheet := &SpriteSheet{}

	spriteSheet.Land = make([]*ebiten.Image, 17)
	for i := 1; i <= 16; i++ {
		spriteSheet.Land[i] = loadImage(fmt.Sprintf("images/Terrain/Tiles/ground_%v.png", i))
	}

	spriteSheet.Mountains = make([]*ebiten.Image, 17)
	for i := 1; i <= 16; i++ {
		spriteSheet.Mountains[i] = loadImage(fmt.Sprintf("images/Terrain/Mountains/mountain_%v.png", i))
	}

	spriteSheet.Forests = make([]*ebiten.Image, 17)
	for i := 1; i <= 16; i++ {
		spriteSheet.Forests[i] = loadImage(fmt.Sprintf("images/Terrain/Forests/Forest_%v.png", i))
	}

	spriteSheet.Water = make([]*ebiten.Image, 4)
	spriteSheet.Water[0] = loadImage("images/Terrain/Water/water.png")
	spriteSheet.Water[1] = loadImage("images/Terrain/Water/water_wall_left.png")
	spriteSheet.Water[2] = loadImage("images/Terrain/Water/water_wall_right.png")
	spriteSheet.Water[3] = loadImage("images/Terrain/Water/water_wall_left_wall_right.png")

	spriteSheet.Ocean = make([]*ebiten.Image, 4)
	spriteSheet.Ocean[0] = loadImage("images/Terrain/Water/ocean.png")
	spriteSheet.Ocean[1] = loadImage("images/Terrain/Water/ocean_wall_left.png")
	spriteSheet.Ocean[2] = loadImage("images/Terrain/Water/ocean_wall_right.png")
	spriteSheet.Ocean[3] = loadImage("images/Terrain/Water/ocean_wall_left_wall_right.png")

	spriteSheet.Village = loadImage("images/village.png")
	spriteSheet.Ice = loadImage("images/Terrain/Tiles/ice.png")
	return spriteSheet, nil
}
