package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
)

// drawTabOverview renders the overview tab content area.
func drawTabOverview(screen *ebiten.Image, world simulation.WorldState, face font.Face, cx, cy, cw, ch int) {
	drawPanel(screen, cx, cy, cw, ch)
	y := cy + 16
	x := cx + 12

	// Carbon section header.
	drawLabel(screen, x, y, "--- Carbon ---", ColourAccent, face)
	y += 18
	drawLabel(screen, x, y,
		fmt.Sprintf("Weekly net: %.3f MtCO2e", world.WeeklyNetCarbonMt),
		ColourTextPrimary, face)
	y += 16
	drawLabel(screen, x, y,
		fmt.Sprintf("Cumulative stock: %.1f MtCO2e", world.Carbon.CumulativeStock),
		ColourTextPrimary, face)
	y += 16

	// Annual budget warning.
	if world.Carbon.OvershootAccumulator > 0 {
		warnStr := fmt.Sprintf("! Annual budget exceeded: +%.1f Mt overshoot", world.Carbon.OvershootAccumulator)
		drawLabel(screen, x, y, warnStr, colour(0xE7, 0x4C, 0x3C), face)
	} else {
		drawLabel(screen, x, y, "Annual budget: on track", ColourClimateLow, face)
	}
	y += 24

	// Government section.
	drawLabel(screen, x, y, "--- Government ---", ColourAccent, face)
	y += 18
	drawLabel(screen, x, y,
		fmt.Sprintf("Approval: %.0f%%", world.GovernmentLastPollResult),
		ColourTextPrimary, face)
	y += 4
	drawBar(screen, x, y, 200, 10, world.GovernmentLastPollResult, 100, ColourAccent, ColourButtonNormal)
	y += 20
	drawLabel(screen, x, y,
		fmt.Sprintf("LCR: %.0f", world.LCR.LastPollResult),
		ColourTextPrimary, face)
	y += 4
	drawBar(screen, x, y, 200, 10, world.LCR.LastPollResult, 100, ColourOrgThinkTank, ColourButtonNormal)
	y += 24

	// Budget section.
	drawLabel(screen, x, y, "--- Budget ---", ColourAccent, face)
	y += 18
	drawLabel(screen, x, y,
		fmt.Sprintf("Quarterly discretionary: GBP %.0f m", world.LastBudget.TotalGBPm),
		ColourTextPrimary, face)
	y += 16
	drawLabel(screen, x, y,
		fmt.Sprintf("Tax revenue: GBP %.2f bn", world.LastTaxRevenue.GBPBillions),
		ColourTextPrimary, face)
	y += 24

	// Event log.
	drawLabel(screen, x, y, "--- Recent Events ---", ColourAccent, face)
	y += 18
	entries := world.EventLog.Entries()
	// Show up to last 10.
	start := 0
	if len(entries) > 10 {
		start = len(entries) - 10
	}
	for _, e := range entries[start:] {
		line := fmt.Sprintf("[Wk %4d] %s", e.Week, e.Name)
		drawLabel(screen, x, y, line, ColourTextMuted, face)
		y += 14
		if y > cy+ch-10 {
			break
		}
	}
}

// colour is a small helper to build color.RGBA inline.
func colour(r, g, b uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: 0xFF}
}
