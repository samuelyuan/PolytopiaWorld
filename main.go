package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	inputPtr := flag.String("input", "", "Input filename")
	flag.Parse()

	inputFilename := *inputPtr
	fmt.Println("Input filename: ", inputFilename)

	ebiten.SetWindowTitle("PolytopiaWorld")
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowResizable(true)

	g, err := NewGame(inputFilename)
	if err != nil {
		log.Fatal(err)
	}

	if err = ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
