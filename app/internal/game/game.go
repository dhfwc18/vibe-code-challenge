package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/event"
	"github.com/vibe-code-challenge/twenty-fifty/internal/save"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"github.com/vibe-code-challenge/twenty-fifty/internal/ui"
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 720
	Title        = "20-50"
)

// Game implements ebiten.Game and owns all top-level game state.
type Game struct {
	cfg    *config.Config
	world  simulation.WorldState
	ui     *ui.UI
	events []event.EventEntry // fired during the most recent week advance
}

// New returns an initialised Game with a fresh world seeded from masterSeed.
func New(cfg *config.Config, masterSeed save.MasterSeed) *Game {
	world := simulation.NewWorld(cfg, masterSeed)
	u := ui.New(&world, cfg)
	return &Game{
		cfg:    cfg,
		world:  world,
		ui:     u,
		events: []event.EventEntry{},
	}
}

// Update is called once per tick (60 Hz by default) and advances game state.
func (g *Game) Update() error {
	// Collect player actions from the UI.
	actions := g.ui.Update(&g.world)

	// If the player signalled Advance Week, run the simulation for one week.
	if g.ui.AdvanceWeekRequested() {
		newWorld, firedEvents := simulation.AdvanceWeek(g.world, actions)
		g.world = newWorld
		g.events = firedEvents

		// Pass the most recent event name to the HUD notification strip.
		if len(firedEvents) > 0 {
			g.ui.NotifyEvent(firedEvents[len(firedEvents)-1].Name)
		}
	}

	return nil
}

// Draw is called once per frame and renders the current state to screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColour)
	g.ui.Draw(screen, g.world)
}

// Layout returns the logical screen dimensions used by Ebitengine.
// Returning the outside (window) dimensions gives native-resolution rendering
// so the game fills the window without letterboxing.
func (g *Game) Layout(outsideW, outsideH int) (int, int) {
	return outsideW, outsideH
}
