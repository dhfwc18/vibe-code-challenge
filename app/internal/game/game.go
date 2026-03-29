package game

import (
	"encoding/json"
	"log"

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

type gamePhase int

const (
	phaseScenarioSelect gamePhase = iota
	phasePlay
)

// Game implements ebiten.Game and owns all top-level game state.
type Game struct {
	cfg            *config.Config
	masterSeed     save.MasterSeed
	world          simulation.WorldState
	ui             *ui.UI
	scenarioScreen *ui.ScenarioScreen
	phase          gamePhase
	events         []event.EventEntry // fired during the most recent week advance
	savePath       string             // path for the autosave file
}

// New returns a Game ready to show the scenario selection screen.
// savePath is the path to use for autosave (pass "" to derive automatically).
func New(cfg *config.Config, masterSeed save.MasterSeed, savePath string) *Game {
	if savePath == "" {
		savePath = save.AutoSavePath("")
	}
	ss := ui.NewScenarioScreen()
	// Inform the scenario screen whether a save exists so it can show "Continue".
	ss.SetHasSave(save.Exists(savePath))
	return &Game{
		cfg:            cfg,
		masterSeed:     masterSeed,
		scenarioScreen: ss,
		phase:          phaseScenarioSelect,
		events:         []event.EventEntry{},
		savePath:       savePath,
	}
}

// startScenario initialises the world for the chosen scenario and switches to play phase.
func (g *Game) startScenario(id config.ScenarioID) {
	scenario := config.ScenarioByID(id)
	world := simulation.NewWorldFromScenario(g.cfg, g.masterSeed, scenario)
	g.world = world
	g.ui = ui.New(&world, g.cfg)
	g.phase = phasePlay
}

// loadSave resumes from the autosave file.
func (g *Game) loadSave() {
	ss, err := save.Read(g.savePath)
	if err != nil {
		log.Printf("load save: %v", err)
		return
	}
	if len(ss.WorldData) == 0 {
		log.Printf("load save: no world data in file")
		return
	}
	var wd simulation.WorldSaveData
	if err := json.Unmarshal(ss.WorldData, &wd); err != nil {
		log.Printf("load save: unmarshal world data: %v", err)
		return
	}
	world, err := simulation.RestoreWorld(wd, g.cfg)
	if err != nil {
		log.Printf("load save: restore world: %v", err)
		return
	}
	g.world = world
	g.ui = ui.New(&world, g.cfg)
	g.phase = phasePlay
}

// autoSave writes the current world to disk.
func (g *Game) autoSave() {
	wd := simulation.SaveWorld(g.world)
	raw, err := json.Marshal(wd)
	if err != nil {
		log.Printf("auto-save: marshal world: %v", err)
		return
	}
	ss := &save.SaveState{
		GameWeek:  g.world.Week,
		GameYear:  g.world.Year,
		WorldData: json.RawMessage(raw),
	}
	if saveErr := save.Write(g.savePath, ss); saveErr != nil {
		log.Printf("auto-save: write: %v", saveErr)
	}
}

// Update is called once per tick (60 Hz by default) and advances game state.
func (g *Game) Update() error {
	if g.phase == phaseScenarioSelect {
		result := g.scenarioScreen.Update(config.Scenarios)
		switch result {
		case ui.ScenarioResultContinue:
			g.loadSave()
		case ui.ScenarioResultNone:
			// no action
		default:
			// result is a ScenarioID string
			g.startScenario(config.ScenarioID(result))
		}
		return nil
	}

	// Collect player actions from the UI.
	actions := g.ui.Update(&g.world)

	// If the player signalled Advance Week, run the simulation for one week.
	if g.ui.AdvanceWeekRequested() {
		newWorld, firedEvents := simulation.AdvanceWeek(g.world, actions)
		g.world = newWorld
		g.events = firedEvents

		// Pass events to the newspaper queue and HUD notification strip.
		if len(firedEvents) > 0 {
			g.ui.QueueNewsItems(firedEvents)
			g.ui.NotifyEvent(firedEvents[len(firedEvents)-1].Name)
		}

		// Auto-save after each successful week advance.
		g.autoSave()
	}

	return nil
}

// Draw is called once per frame and renders the current state to screen.
func (g *Game) Draw(screen *ebiten.Image) {
	if g.phase == phaseScenarioSelect {
		g.scenarioScreen.Draw(screen, config.Scenarios)
		return
	}
	screen.Fill(backgroundColour)
	g.ui.Draw(screen, g.world)
}

// Layout returns the logical screen dimensions used by Ebitengine.
// Returning the outside (window) dimensions gives native-resolution rendering
// so the game fills the window without letterboxing.
func (g *Game) Layout(outsideW, outsideH int) (int, int) {
	return outsideW, outsideH
}
