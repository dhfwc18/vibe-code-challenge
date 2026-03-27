package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 720
	Title        = "Net Zero"
)

// Game implements ebiten.Game and owns all top-level game state.
type Game struct{}

// New returns an initialised Game ready to be passed to ebiten.RunGame.
func New() *Game {
	return &Game{}
}

// Update is called once per tick (60 Hz by default) and advances game state.
func (g *Game) Update() error {
	return nil
}

// Draw is called once per frame and renders the current state to screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColour)
}

// Layout returns the logical screen dimensions used by Ebitengine.
func (g *Game) Layout(_, _ int) (int, int) {
	return ScreenWidth, ScreenHeight
}
