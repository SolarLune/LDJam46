package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/inpututil"

	"github.com/hajimehoshi/ebiten"
)

type Game struct {
	Level         *Level
	Width, Height int
	DebugMode     bool
}

func NewGame() *Game {

	ebiten.SetWindowResizable(true)
	ebiten.SetWindowTitle("LDJam46")

	game := &Game{
		Width:  640,
		Height: 360,
	}

	game.Level = NewLevel(game)

	// Debug FPS printing
	go func() {
		for {
			fmt.Println(ebiten.CurrentFPS())
			fmt.Println(ebiten.CurrentTPS())
			time.Sleep(time.Second)
		}
	}()

	return game

}

func (game *Game) Update(screen *ebiten.Image) error {

	var quit error

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		quit = errors.New("Quit")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		game.Level = NewLevel(game)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		game.DebugMode = !game.DebugMode
	}

	game.Level.Update(screen)

	return quit

}

func (game *Game) Layout(w, h int) (int, int) {
	return game.Width, game.Height
}

func main() {

	game := NewGame()
	ebiten.RunGame(game)

}

func getPath(path string) string {

	root, _ := os.Executable()

	// We're running a debug build, not an executable
	if strings.Contains(root, "/tmp") {

		root, _ = os.Getwd()
	}

	return filepath.Join(root, filepath.FromSlash(path))

}
