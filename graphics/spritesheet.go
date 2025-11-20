package graphics

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type SpriteSheet struct {
	// Terrain
	Land      []*ebiten.Image
	Mountains []*ebiten.Image
	Forests   []*ebiten.Image
	Water     []*ebiten.Image
	Ocean     []*ebiten.Image

	// Resources
	Crop   *ebiten.Image
	Fish   *ebiten.Image
	Fruits []*ebiten.Image

	// Improvements
	Village *ebiten.Image
	Ruin    *ebiten.Image
	Farm    *ebiten.Image
	Port    *ebiten.Image
	Mine    *ebiten.Image
	Ice     *ebiten.Image
}

func loadImage(filename string) (*ebiten.Image, error) {
	image, _, err := ebitenutil.NewImageFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load image %s: %w", filename, err)
	}
	return image, nil
}

func loadImageArray(basePath string, count int, filenameFunc func(int) string) ([]*ebiten.Image, error) {
	images := make([]*ebiten.Image, count)
	for i := 1; i <= count-1; i++ {
		img, err := loadImage(filenameFunc(i))
		if err != nil {
			return nil, err
		}
		images[i] = img
	}
	return images, nil
}

func LoadSpriteSheet() (*SpriteSheet, error) {
	spriteSheet := &SpriteSheet{}

	var err error

	spriteSheet.Land, err = loadImageArray("images/Terrain/Tiles", 17, func(i int) string {
		return fmt.Sprintf("images/Terrain/Tiles/ground_%v.png", i)
	})
	if err != nil {
		return nil, err
	}

	spriteSheet.Mountains, err = loadImageArray("images/Terrain/Mountains", 17, func(i int) string {
		return fmt.Sprintf("images/Terrain/Mountains/mountain_%v.png", i)
	})
	if err != nil {
		return nil, err
	}

	spriteSheet.Forests, err = loadImageArray("images/Terrain/Forests", 17, func(i int) string {
		return fmt.Sprintf("images/Terrain/Forests/Forest_%v.png", i)
	})
	if err != nil {
		return nil, err
	}

	spriteSheet.Fruits, err = loadImageArray("images/Fruits", 17, func(i int) string {
		return fmt.Sprintf("images/Fruits/ResourceGFX_fruit_%v.png", i)
	})
	if err != nil {
		return nil, err
	}

	spriteSheet.Water = make([]*ebiten.Image, 4)
	waterFiles := []string{"water.png", "water_wall_left.png", "water_wall_right.png", "water_wall_left_wall_right.png"}
	for i, filename := range waterFiles {
		spriteSheet.Water[i], err = loadImage("images/Terrain/Water/" + filename)
		if err != nil {
			return nil, err
		}
	}

	spriteSheet.Ocean = make([]*ebiten.Image, 4)
	oceanFiles := []string{"ocean.png", "ocean_wall_left.png", "ocean_wall_right.png", "ocean_wall_left_wall_right.png"}
	for i, filename := range oceanFiles {
		spriteSheet.Ocean[i], err = loadImage("images/Terrain/Water/" + filename)
		if err != nil {
			return nil, err
		}
	}

	spriteSheet.Crop, err = loadImage("images/Common/crop.png")
	if err != nil {
		return nil, err
	}

	spriteSheet.Fish, err = loadImage("images/Common/fish.png")
	if err != nil {
		return nil, err
	}

	spriteSheet.Ice, err = loadImage("images/Terrain/Tiles/ice.png")
	if err != nil {
		return nil, err
	}

	spriteSheet.Village, err = loadImage("images/Common/village.png")
	if err != nil {
		return nil, err
	}

	spriteSheet.Ruin, err = loadImage("images/Common/ruin.png")
	if err != nil {
		return nil, err
	}

	spriteSheet.Farm, err = loadImage("images/Common/farm.png")
	if err != nil {
		return nil, err
	}

	spriteSheet.Port, err = loadImage("images/Common/port.png")
	if err != nil {
		return nil, err
	}

	spriteSheet.Mine, err = loadImage("images/Common/mine.png")
	if err != nil {
		return nil, err
	}

	return spriteSheet, nil
}
