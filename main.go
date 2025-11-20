package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, "PolytopiaWorld - A map viewer for The Battle of Polytopia\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s -input=<filename>\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExample:\n")
	fmt.Fprintf(os.Stderr, "  %s -input=00000000-0000-0000-0000-000000000000.state\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Note: The input file must be a .state file from The Battle of Polytopia save game directory.\n")
	fmt.Fprintf(os.Stderr, "      Make sure to copy the .state file before the game ends, as it will be deleted.\n")
}

func main() {
	inputPtr := flag.String("input", "", "Path to the Polytopia .state file to view")
	helpPtr := flag.Bool("help", false, "Show this help message")
	flag.Usage = printUsage
	flag.Parse()

	if *helpPtr {
		printUsage()
		os.Exit(0)
	}

	inputFilename := *inputPtr
	if inputFilename == "" {
		fmt.Fprintf(os.Stderr, "Error: -input flag is required\n\n")
		printUsage()
		os.Exit(1)
	}

	ebiten.SetWindowTitle("PolytopiaWorld")
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowResizable(true)

	g, err := NewGame(inputFilename)
	if err != nil {
		log.Fatalf("Failed to load game: %v", err)
	}

	if err = ebiten.RunGame(g); err != nil {
		log.Fatalf("Game error: %v", err)
	}
}
