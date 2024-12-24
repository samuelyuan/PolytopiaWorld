package graphics

import (
	"fmt"

	polytopiamapmodel "github.com/samuelyuan/polytopiamapmodelgo"
)

type RenderMap struct {
	MapWidth  int
	MapHeight int

	Tiles          [][]*Tile
	TileSizeWidth  int
	TileSizeHeight int
}

func (renderMap *RenderMap) Tile(x, y int) *Tile {
	if x >= 0 && y >= 0 && x < renderMap.MapWidth && y < renderMap.MapHeight {
		return renderMap.Tiles[y][x]
	}
	return nil
}

func (renderMap *RenderMap) Size() (width, height int) {
	return renderMap.MapWidth, renderMap.MapHeight
}

func NewMap(mapFilename string) (*RenderMap, error) {
	saveData, err := polytopiamapmodel.ReadPolytopiaCompressedFile(mapFilename)
	if err != nil {
		return nil, fmt.Errorf("Failed to load save state: %s", err)
	}

	renderMap := &RenderMap{
		MapWidth:       saveData.MapWidth,
		MapHeight:      saveData.MapHeight,
		TileSizeWidth:  1019,
		TileSizeHeight: 976,
	}

	spriteSheet, err := LoadSpriteSheet()
	if err != nil {
		return nil, fmt.Errorf("Failed to load sprites: %s", err)
	}

	renderMap.Tiles = make([][]*Tile, saveData.MapHeight)
	for y := 0; y < saveData.MapHeight; y++ {
		renderMap.Tiles[y] = make([]*Tile, saveData.MapWidth)
		for x := 0; x < saveData.MapWidth; x++ {
			renderTile := &Tile{}

			tileData := saveData.TileData[y][x]
			terrain := tileData.Terrain
			renderTile.Terrain = terrain
			switch terrain {
			// shift water and ocean to be below land
			// Positive offsetY goes down, negative offsetY goes up
			case 1: // Water
				if x == 0 && y == 0 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Water[3], OffsetX: 0, OffsetY: 75}) // left wall and right wall
				} else if x == 0 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Water[1], OffsetX: 0, OffsetY: 75}) // right wall
				} else if y == 0 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Water[2], OffsetX: 0, OffsetY: 75}) // left wall
				} else {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Water[0], OffsetX: 0, OffsetY: 75})
				}
			case 2: // Ocean
				if x == 0 && y == 0 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Ocean[3], OffsetX: 0, OffsetY: 75}) // left wall and right wall
				} else if x == 0 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Ocean[1], OffsetX: 0, OffsetY: 75}) // right wall
				} else if y == 0 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Ocean[2], OffsetX: 0, OffsetY: 75}) // left wall
				} else {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Ocean[0], OffsetX: 0, OffsetY: 75})
				}
			case 3: // flat land
				renderTile.AddSprite(SpriteData{Image: spriteSheet.Land[tileData.Climate], OffsetX: 0, OffsetY: 0})
			case 4: // mountain
				renderTile.AddSprite(SpriteData{Image: spriteSheet.Land[tileData.Climate], OffsetX: 0, OffsetY: 0})
				renderTile.AddSprite(SpriteData{Image: spriteSheet.Mountains[tileData.Climate], OffsetX: 0, OffsetY: -250})
			case 5: // forest
				renderTile.AddSprite(SpriteData{Image: spriteSheet.Land[tileData.Climate], OffsetX: 0, OffsetY: 0})
				renderTile.AddSprite(SpriteData{Image: spriteSheet.Forests[tileData.Climate], OffsetX: 0, OffsetY: -150})
			case 6: // ice
				renderTile.AddSprite(SpriteData{Image: spriteSheet.Ice, OffsetX: 0, OffsetY: 0})
			}

			if tileData.ResourceExists {
				if tileData.ResourceType == 2 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Crop, OffsetX: 50, OffsetY: -50})
				} else if tileData.ResourceType == 3 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Fish, OffsetX: 0, OffsetY: 0})
				} else if tileData.ResourceType == 6 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Fruits[tileData.Climate], OffsetX: 250, OffsetY: 150})
				}
			}

			if tileData.ImprovementData != nil {
				if tileData.ImprovementType == 1 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Village, OffsetX: 250, OffsetY: 0})
				} else if tileData.ImprovementType == 2 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Ruin, OffsetX: -50, OffsetY: -450})
				} else if tileData.ImprovementType == 5 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Farm, OffsetX: 100, OffsetY: 0})
				} else if tileData.ImprovementType == 8 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Port, OffsetX: 75, OffsetY: 75})
				} else if tileData.ImprovementType == 21 {
					renderTile.AddSprite(SpriteData{Image: spriteSheet.Mine, OffsetX: 250, OffsetY: 50})
				}
			}

			renderMap.Tiles[y][x] = renderTile
		}
	}

	return renderMap, nil
}
