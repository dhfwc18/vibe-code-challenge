package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/save"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 720
	Title        = "20-50"
)

// Game implements ebiten.Game and owns all top-level game state.
type Game struct {
	cfg   *config.Config
	world simulation.WorldState
}

// New returns an initialised Game with a fresh world seeded from masterSeed.
func New(cfg *config.Config, masterSeed save.MasterSeed) *Game {
	return &Game{
		cfg:   cfg,
		world: simulation.NewWorld(cfg, masterSeed),
	}
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
