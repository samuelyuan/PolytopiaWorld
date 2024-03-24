package main

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samuelyuan/PolytopiaWorld/graphics"
)

type Game struct {
	w, h       int
	currentMap *graphics.RenderMap

	camX, camY float64
	camScale   float64
	camScaleTo float64

	mousePanX, mousePanY int

	offscreen *ebiten.Image
}

func NewGame(mapFilename string) (*Game, error) {
	level, err := graphics.NewMap(mapFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to create new level: %s", err)
	}

	g := &Game{
		currentMap: level,
		camScale:   1,
		camScaleTo: 1,
		mousePanX:  math.MinInt32,
		mousePanY:  math.MinInt32,
	}
	return g, nil
}

func (g *Game) Update() error {
	var scrollY float64
	if ebiten.IsKeyPressed(ebiten.KeyC) || ebiten.IsKeyPressed(ebiten.KeyPageDown) {
		scrollY = -0.25
	} else if ebiten.IsKeyPressed(ebiten.KeyE) || ebiten.IsKeyPressed(ebiten.KeyPageUp) {
		scrollY = .25
	} else {
		_, scrollY = ebiten.Wheel()
		if scrollY < -1 {
			scrollY = -1
		} else if scrollY > 1 {
			scrollY = 1
		}
	}
	g.camScaleTo += scrollY * (g.camScaleTo / 7)

	if g.camScaleTo < 0.01 {
		g.camScaleTo = 0.01
	} else if g.camScaleTo > 100 {
		g.camScaleTo = 100
	}

	div := 10.0
	if g.camScaleTo > g.camScale {
		g.camScale += (g.camScaleTo - g.camScale) / div
	} else if g.camScaleTo < g.camScale {
		g.camScale -= (g.camScale - g.camScaleTo) / div
	}

	pan := 7.0 / g.camScale
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.camX -= pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.camX += pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.camY -= pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.camY += pan
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		if g.mousePanX == math.MinInt32 && g.mousePanY == math.MinInt32 {
			g.mousePanX, g.mousePanY = ebiten.CursorPosition()
		} else {
			x, y := ebiten.CursorPosition()
			dx, dy := float64(g.mousePanX-x)*(pan/100), float64(g.mousePanY-y)*(pan/100)
			g.camX, g.camY = g.camX-dx, g.camY+dy
		}
	} else if g.mousePanX != math.MinInt32 || g.mousePanY != math.MinInt32 {
		g.mousePanX, g.mousePanY = math.MinInt32, math.MinInt32
	}

	worldWidth := float64(g.currentMap.MapWidth * g.currentMap.TileSizeWidth / 2)
	worldHeight := float64(g.currentMap.MapHeight * g.currentMap.TileSizeHeight / 2)
	if g.camX < -worldWidth {
		g.camX = -worldWidth
	} else if g.camX > worldWidth {
		g.camX = worldWidth
	}
	if g.camY < -worldHeight {
		g.camY = -worldHeight
	} else if g.camY > worldHeight {
		g.camY = worldHeight
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderLevel(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.w, g.h = outsideWidth, outsideHeight
	return g.w, g.h
}

func (g *Game) cartesianToIso(x, y float64) (float64, float64) {
	ix := (x - y) * float64(g.currentMap.TileSizeWidth/2)
	iy := (x + y) * float64(-g.currentMap.TileSizeHeight) / 3.2 // y is inverted
	return ix, iy
}

func (g *Game) renderLevel(screen *ebiten.Image) {
	padding := float64(g.currentMap.TileSizeWidth) * g.camScale
	cx, cy := float64(g.w/2), float64(g.h/2)

	scaleLater := g.camScale > 1
	target := screen
	scale := g.camScale

	// When zooming in, tiles can have slight bleeding edges.
	// To avoid them, render the result on an offscreen first and then scale it later.
	if scaleLater {
		if g.offscreen != nil {
			if g.offscreen.Bounds().Size() != screen.Bounds().Size() {
				// g.offscreen.Deallocate()
				g.offscreen = nil
			}
		}
		if g.offscreen == nil {
			s := screen.Bounds().Size()
			g.offscreen = ebiten.NewImage(s.X, s.Y)
		}
		target = g.offscreen
		target.Clear()
		scale = 1
	}

	for y := g.currentMap.MapHeight - 1; y >= 0; y-- {
		for x := g.currentMap.MapWidth - 1; x >= 0; x-- {
			isoCoordinateX, isoCoordinateY := g.cartesianToIso(float64(x), float64(y))

			drawX, drawY := ((isoCoordinateX-g.camX)*g.camScale)+cx, ((isoCoordinateY+g.camY)*g.camScale)+cy
			if drawX+padding < 0 || drawY+padding < 0 || drawX > float64(g.w) || drawY > float64(g.h) {
				continue
			}

			renderTile := g.currentMap.Tiles[y][x]
			if renderTile == nil {
				continue
			}

			options := make([]*ebiten.DrawImageOptions, len(renderTile.Sprites))
			for i := 0; i < len(renderTile.Sprites); i++ {
				spriteData := renderTile.Sprites[i]
				options[i] = &ebiten.DrawImageOptions{}
				options[i].GeoM.Reset()
				options[i].GeoM.Translate(isoCoordinateX+float64(spriteData.OffsetX), isoCoordinateY+float64(spriteData.OffsetY))
				options[i].GeoM.Translate(-g.camX, g.camY)
				options[i].GeoM.Scale(scale, scale)
				options[i].GeoM.Translate(cx, cy)
			}
			renderTile.Draw(target, options)
		}
	}

	if scaleLater {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-cx, -cy)
		op.GeoM.Scale(float64(g.camScale), float64(g.camScale))
		op.GeoM.Translate(cx, cy)
		screen.DrawImage(target, op)
	}
}
