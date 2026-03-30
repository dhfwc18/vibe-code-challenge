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
	lastEventName string
}

// newHUD creates a new HUD.
func newHUD() *HUD {
	return &HUD{}
}

// setLastEvent stores the most recent event name for the notification strip.
func (h *HUD) setLastEvent(name string) {
	h.lastEventName = name
}

// hudItemW is the fixed width budget per HUD slot (proportional layout).
const (
	hudBtnW = 148
	hudBtnH = 28
)

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

	// "Advance Week" button anchored at the right edge.
	btnX := w - hudBtnW - 8
	btnY := (hudHeight - hudBtnH) / 2
	bg := buttonColour(btnX, btnY, hudBtnW, hudBtnH, true)
	solidRect(screen, btnX, btnY, hudBtnW, hudBtnH, bg)
	lbl := "Advance Week"
	drawLabel(screen, btnX+(hudBtnW-len(lbl)*7)/2, btnY+hudBtnH-8, lbl, ColourTextPrimary, face)

	// Usable content width between left edge and advance-week button.
	usable := btnX - 8

	// Slot layout: divide usable width into 5 equal slots.
	// [0] date/time  [1] AP  [2] Rep+grade  [3] climate badge  [4] boss/event
	slotW := usable / 5
	textY := 30 // baseline for all text labels

	// Slot 0: Month Year Wk Q
	mn := world.Month
	if mn < 1 || mn > 12 {
		mn = 1
	}
	monthStr := [12]string{
		"Jan", "Feb", "Mar", "Apr", "May", "Jun",
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
	}[mn-1]
	timeStr := fmt.Sprintf("%s %d  Wk %d  Q%d", monthStr, world.Year, world.Week, world.Quarter)
	drawLabel(screen, 8, textY, timeStr, ColourTextPrimary, face)

	// Slot 1: AP
	apStr := fmt.Sprintf("AP: %d", effectiveAP)
	drawLabel(screen, slotW+8, textY, apStr, ColourAccent, face)

	// Slot 2: Rep grade
	grade, gradeCol := lcrGrade(world.LCR.LastPollResult)
	gradeStr := fmt.Sprintf("Rep: %s (%.0f)", grade, world.LCR.LastPollResult)
	drawLabel(screen, slotW*2+8, textY, gradeStr, gradeCol, face)

	// Slot 3: Climate badge (vertically centred in bar).
	climateCol := climateColour(world.ClimateState.Level)
	climateLabel := climateLevelName(world.ClimateState.Level)
	badgeX := slotW*3 + 8
	badgeY := (hudHeight - 15) / 2
	drawBadge(screen, badgeX, badgeY, climateLabel, climateCol, face)

	// Slot 4: Boss name OR feedback/event (whichever is active).
	slot4X := slotW*4 + 8
	if feedbackMsg != "" {
		drawLabel(screen, slot4X, textY, feedbackMsg, colour(0xE7, 0x4C, 0x3C), face)
	} else {
		bossName := energyMinisterName(world)
		if bossName != "" {
			drawLabel(screen, slot4X, textY, "Boss: "+bossName, colour(0xF0, 0xC0, 0x40), face)
		} else if h.lastEventName != "" {
			drawLabel(screen, slot4X, textY, "Ev: "+h.lastEventName, ColourClimateMedium, face)
		}
	}
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
