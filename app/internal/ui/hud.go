package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/carbon"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
)

const hudHeight = 40

// HUD renders the top bar of the game screen.
type HUD struct {
	advanceWeekSignal bool
	lastEventName     string
}

// newHUD creates a new HUD.
func newHUD() *HUD {
	return &HUD{}
}

// signalAdvanceWeek marks that the "Advance Week" button was pressed this frame.
func (h *HUD) signalAdvanceWeek() {
	h.advanceWeekSignal = true
}

// consumeAdvanceWeek returns true if "Advance Week" was pressed, then resets.
func (h *HUD) consumeAdvanceWeek() bool {
	if h.advanceWeekSignal {
		h.advanceWeekSignal = false
		return true
	}
	return false
}

// setLastEvent stores the most recent event name for the notification strip.
func (h *HUD) setLastEvent(name string) {
	h.lastEventName = name
}

// Draw renders the HUD top bar onto screen.
func (h *HUD) Draw(screen *ebiten.Image, world simulation.WorldState, face font.Face) {
	w := screen.Bounds().Dx()

	// Background bar.
	bar := ebiten.NewImage(w, hudHeight)
	bar.Fill(ColourPanel)
	screen.DrawImage(bar, &ebiten.DrawImageOptions{})

	// Border bottom.
	borderImg := ebiten.NewImage(w, 1)
	borderImg.Fill(color.RGBA{R: 0x2E, G: 0x45, B: 0x38, A: 0xFF})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, float64(hudHeight-1))
	screen.DrawImage(borderImg, op)

	// Left: Year, Week, Quarter.
	timeStr := fmt.Sprintf("Year %d  Wk %d  Q%d", world.Year, world.Week, world.Quarter)
	drawLabel(screen, 8, 26, timeStr, ColourTextPrimary, face)

	// Centre-left: AP remaining.
	apStr := fmt.Sprintf("AP: %d", world.Player.APRemaining)
	drawLabel(screen, 260, 26, apStr, ColourAccent, face)

	// Centre: LCR value.
	lcrStr := fmt.Sprintf("LCR: %.0f", world.LCR.LastPollResult)
	drawLabel(screen, 360, 26, lcrStr, ColourTextPrimary, face)

	// Climate badge.
	climateCol := climateColour(world.ClimateState.Level)
	climateLabel := climateLevelName(world.ClimateState.Level)
	drawBadge(screen, 460, 12, climateLabel, climateCol, face)

	// Event notification strip.
	if h.lastEventName != "" {
		evStr := "Event: " + h.lastEventName
		drawLabel(screen, 600, 26, evStr, ColourClimateMedium, face)
	}

	// "Advance Week" button placeholder (drawn directly; click handled in Update).
	btnX := w - 160
	btnY := 6
	btnW := 148
	btnH := 28
	solidRect(screen, btnX, btnY, btnW, btnH, ColourButtonNormal)
	drawLabel(screen, btnX+20, btnY+19, "Advance Week", ColourTextPrimary, face)
}

// climateColour returns the palette colour for a given climate level.
func climateColour(level carbon.ClimateLevel) color.RGBA {
	switch level {
	case carbon.ClimateLevelStable:
		return ColourClimateLow
	case carbon.ClimateLevelElevated:
		return ColourClimateMedium
	case carbon.ClimateLevelCritical:
		return ColourClimateCritical
	case carbon.ClimateLevelEmergency:
		return ColourClimateEmergency
	default:
		return ColourClimateLow
	}
}

// climateLevelName returns the display name for a climate level.
func climateLevelName(level carbon.ClimateLevel) string {
	switch level {
	case carbon.ClimateLevelStable:
		return "STABLE"
	case carbon.ClimateLevelElevated:
		return "ELEVATED"
	case carbon.ClimateLevelCritical:
		return "CRITICAL"
	case carbon.ClimateLevelEmergency:
		return "EMERGENCY"
	default:
		return "STABLE"
	}
}
