package graphics

import (
	"fmt"

	polytopiamapmodel "github.com/samuelyuan/polytopiamapmodelgo"
)

const (
	// Tile size constants
	TileSizeWidth  = 1019
	TileSizeHeight = 976

	// Water/Ocean offset constants
	waterOffsetY = 75

	// Terrain offset constants
	mountainOffsetY = -250
	forestOffsetY   = -150
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
		TileSizeWidth:  TileSizeWidth,
		TileSizeHeight: TileSizeHeight,
	}

	spriteSheet, err := LoadSpriteSheet()
	if err != nil {
		return nil, fmt.Errorf("Failed to load sprites: %s", err)
	}

	renderMap.Tiles = make([][]*Tile, saveData.MapHeight)
	for y := 0; y < saveData.MapHeight; y++ {
		renderMap.Tiles[y] = make([]*Tile, saveData.MapWidth)
		for x := 0; x < saveData.MapWidth; x++ {
			tileData := saveData.TileData[y][x]
			renderTile := &Tile{
				Terrain: tileData.Terrain,
			}

			addTerrainSprites(renderTile, tileData.Terrain, tileData.Climate, x, y, spriteSheet)

			if tileData.ResourceExists {
				addResourceSprites(renderTile, tileData.ResourceType, tileData.Climate, spriteSheet)
			}

			if tileData.ImprovementData != nil {
				addImprovementSprites(renderTile, tileData.ImprovementType, spriteSheet)
			}

			renderMap.Tiles[y][x] = renderTile
		}
	}

	return renderMap, nil
}

func getWaterSpriteIndex(x, y int) int {
	if x == 0 && y == 0 {
		return 3 // left wall and right wall
	} else if x == 0 {
		return 1 // right wall
	} else if y == 0 {
		return 2 // left wall
	}
	return 0 // no walls
}

func addTerrainSprites(tile *Tile, terrain, climate int, x, y int, spriteSheet *SpriteSheet) {
	switch terrain {
	case 1: // Water
		idx := getWaterSpriteIndex(x, y)
		tile.AddSprite(SpriteData{Image: spriteSheet.Water[idx], OffsetX: 0, OffsetY: waterOffsetY})
	case 2: // Ocean
		idx := getWaterSpriteIndex(x, y)
		tile.AddSprite(SpriteData{Image: spriteSheet.Ocean[idx], OffsetX: 0, OffsetY: waterOffsetY})
	case 3: // flat land
		tile.AddSprite(SpriteData{Image: spriteSheet.Land[climate], OffsetX: 0, OffsetY: 0})
	case 4: // mountain
		tile.AddSprite(SpriteData{Image: spriteSheet.Land[climate], OffsetX: 0, OffsetY: 0})
		tile.AddSprite(SpriteData{Image: spriteSheet.Mountains[climate], OffsetX: 0, OffsetY: mountainOffsetY})
	case 5: // forest
		tile.AddSprite(SpriteData{Image: spriteSheet.Land[climate], OffsetX: 0, OffsetY: 0})
		tile.AddSprite(SpriteData{Image: spriteSheet.Forests[climate], OffsetX: 0, OffsetY: forestOffsetY})
	case 6: // ice
		tile.AddSprite(SpriteData{Image: spriteSheet.Ice, OffsetX: 0, OffsetY: 0})
	}
}

func addResourceSprites(tile *Tile, resourceType, climate int, spriteSheet *SpriteSheet) {
	switch resourceType {
	case 2: // Crop
		tile.AddSprite(SpriteData{Image: spriteSheet.Crop, OffsetX: 50, OffsetY: -50})
	case 3: // Fish
		tile.AddSprite(SpriteData{Image: spriteSheet.Fish, OffsetX: 0, OffsetY: 0})
	case 6: // Fruit
		tile.AddSprite(SpriteData{Image: spriteSheet.Fruits[climate], OffsetX: 250, OffsetY: 150})
	}
}

func addImprovementSprites(tile *Tile, improvementType int, spriteSheet *SpriteSheet) {
	switch improvementType {
	case 1: // Village
		tile.AddSprite(SpriteData{Image: spriteSheet.Village, OffsetX: 250, OffsetY: 0})
	case 2: // Ruin
		tile.AddSprite(SpriteData{Image: spriteSheet.Ruin, OffsetX: -50, OffsetY: -450})
	case 5: // Farm
		tile.AddSprite(SpriteData{Image: spriteSheet.Farm, OffsetX: 100, OffsetY: 0})
	case 8: // Port
		tile.AddSprite(SpriteData{Image: spriteSheet.Port, OffsetX: 75, OffsetY: 75})
	case 21: // Mine
		tile.AddSprite(SpriteData{Image: spriteSheet.Mine, OffsetX: 250, OffsetY: 50})
	}
}
