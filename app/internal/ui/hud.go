package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/carbon"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
)

const hudHeight = 48

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
func (h *HUD) Draw(screen *ebiten.Image, world simulation.WorldState, face font.Face, effectiveAP int, feedbackMsg string) {
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

	// Left: Month, Year, Week, Quarter.
	mn := world.Month
	if mn < 1 || mn > 12 {
		mn = 1
	}
	monthStr := [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun",
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}[mn-1]
	timeStr := fmt.Sprintf("%s %d  Wk %d  Q%d", monthStr, world.Year, world.Week, world.Quarter)
	drawLabel(screen, 8, 30, timeStr, ColourTextPrimary, face)

	// Centre-left: AP remaining (effective, accounting for queued spend).
	apStr := fmt.Sprintf("AP: %d", effectiveAP)
	drawLabel(screen, 280, 30, apStr, ColourAccent, face)

	// Player reputation grade derived from LCR poll result.
	grade, gradeCol := lcrGrade(world.LCR.LastPollResult)
	gradeStr := fmt.Sprintf("Rep: %s (%.0f)", grade, world.LCR.LastPollResult)
	drawLabel(screen, 360, 30, gradeStr, gradeCol, face)

	// Energy Minister name (player's boss).
	bossName := energyMinisterName(world)
	if bossName != "" {
		bossStr := "Boss: " + bossName
		drawLabel(screen, 530, 30, bossStr, colour(0xF0, 0xC0, 0x40), face)
	}

	// Climate badge.
	climateCol := climateColour(world.ClimateState.Level)
	climateLabel := climateLevelName(world.ClimateState.Level)
	drawBadge(screen, 460, 12, climateLabel, climateCol, face)

	// Event notification strip: feedback message takes priority over event name.
	if feedbackMsg != "" {
		drawLabel(screen, 600, 30, feedbackMsg, colour(0xE7, 0x4C, 0x3C), face)
	} else if h.lastEventName != "" {
		evStr := "Event: " + h.lastEventName
		drawLabel(screen, 600, 30, evStr, ColourClimateMedium, face)
	}

	// "Advance Week" button (drawn directly; click handled in Update).
	btnX := w - 160
	btnY := 6
	btnW, btnH := 148, 28
	bg := buttonColour(btnX, btnY, btnW, btnH, true)
	solidRect(screen, btnX, btnY, btnW, btnH, bg)
	drawLabel(screen, btnX+20, btnY+20, "Advance Week", ColourTextPrimary, face)
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

// lcrGrade converts a LowCarbonReputation poll value to a letter grade and colour.
func lcrGrade(lcr float64) (string, color.RGBA) {
	switch {
	case lcr >= 80:
		return "A", colour(0x2E, 0xCC, 0x71)
	case lcr >= 60:
		return "B", colour(0xA8, 0xD8, 0x60)
	case lcr >= 40:
		return "C", colour(0xF3, 0x9C, 0x12)
	case lcr >= 20:
		return "D", colour(0xE6, 0x7E, 0x22)
	default:
		return "F", colour(0xE7, 0x4C, 0x3C)
	}
}

// energyMinisterName returns the name of the current Energy Secretary, or "" if vacant.
func energyMinisterName(world simulation.WorldState) string {
	energyID := world.Government.CabinetByRole[config.RoleEnergy]
	if energyID == "" {
		return ""
	}
	for _, s := range world.Stakeholders {
		if s.ID == energyID {
			nick := s.Nickname
			if nick != "" {
				return nick
			}
			return s.Name
		}
	}
	return ""
}
