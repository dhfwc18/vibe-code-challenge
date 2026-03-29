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

	// Objectives section.
	drawLabel(screen, x, y, "--- Objectives ---", ColourAccent, face)
	y += 16

	// 1. Net Zero by 2050.
	yearsLeft := 2050 - world.Year
	if yearsLeft < 0 {
		yearsLeft = 0
	}
	annualMt := world.WeeklyNetCarbonMt * 52
	nzStatus := "On track"
	nzCol := ColourClimateLow
	if annualMt > 0 && yearsLeft > 0 {
		// Reduction needed per year to reach zero linearly.
		needed := annualMt / float64(yearsLeft)
		annualReduction := world.WeeklyPolicyReductionMt * 52
		if annualReduction < needed*0.5 {
			nzStatus = "Off track"
			nzCol = colour(0xE7, 0x4C, 0x3C)
		} else if annualReduction < needed {
			nzStatus = "Behind"
			nzCol = colour(0xF3, 0x9C, 0x12)
		}
	}
	if annualMt <= 0 {
		nzStatus = "ACHIEVED"
		nzCol = ColourAccent
	}
	drawLabel(screen, x, y, fmt.Sprintf("Net Zero 2050 [%d yr left]: %s", yearsLeft, nzStatus), nzCol, face)
	y += 14
	// Progress bar: how much carbon has been cut from baseline.
	if world.BaseWeeklyMt > 0 {
		cut := 1.0 - world.WeeklyNetCarbonMt/world.BaseWeeklyMt
		if cut < 0 {
			cut = 0
		}
		drawBar(screen, x, y, cw-24, 8, cut, 1.0, nzCol, ColourButtonNormal)
		y += 12
	}
	y += 4

	// 2. Stay in power.
	pop := world.GovernmentLastPollResult
	powStatus := "Comfortable"
	powCol := ColourClimateLow
	if pop < 30 {
		powStatus = "Critical"
		powCol = colour(0xE7, 0x4C, 0x3C)
	} else if pop < 40 {
		powStatus = "Vulnerable"
		powCol = colour(0xF3, 0x9C, 0x12)
	}
	electionWeek := world.Government.ElectionDueWeek
	weeksToElection := electionWeek - world.Week
	elecStr := fmt.Sprintf("Stay in Power [pop %.0f%%]: %s", pop, powStatus)
	if weeksToElection > 0 {
		elecStr += fmt.Sprintf(" -- election in %d wk", weeksToElection)
	} else if electionWeek > 0 {
		elecStr += " -- election overdue"
	}
	drawLabel(screen, x, y, elecStr, powCol, face)
	y += 14
	drawBar(screen, x, y, cw-24, 8, pop, 100, powCol, ColourButtonNormal)
	y += 20

	// 3. Fossil dependence.
	fd := world.FossilDependency
	fdStatus := "Good"
	fdCol := ColourClimateLow
	if fd > 50 {
		fdStatus = "High"
		fdCol = colour(0xF3, 0x9C, 0x12)
	} else if fd > 20 {
		fdStatus = "Moderate"
		fdCol = colour(0xA8, 0xD8, 0x60)
	}
	drawLabel(screen, x, y, fmt.Sprintf("Fossil Dependence [%.0f%%]: %s", fd, fdStatus), fdCol, face)
	y += 14
	drawBar(screen, x, y, cw-24, 8, fd, 100, fdCol, ColourButtonNormal)
	y += 24

	// Carbon section header.
	drawLabel(screen, x, y, "--- Carbon ---", ColourAccent, face)
	y += 18
	drawLabel(screen, x, y,
		fmt.Sprintf("Weekly net: %.3f MtCO2e", world.WeeklyNetCarbonMt),
		ColourTextPrimary, face)
	y += 18
	drawLabel(screen, x, y,
		fmt.Sprintf("Cumulative stock: %.1f MtCO2e", world.Carbon.CumulativeStock),
		ColourTextPrimary, face)
	y += 18

	// Annual budget warning.
	if world.Carbon.OvershootAccumulator > 0 {
		warnStr := fmt.Sprintf("! Annual budget exceeded: +%.1f Mt overshoot", world.Carbon.OvershootAccumulator)
		drawLabel(screen, x, y, warnStr, colour(0xE7, 0x4C, 0x3C), face)
	} else {
		drawLabel(screen, x, y, "Annual budget: on track", ColourClimateLow, face)
	}
	y += 32

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
	y += 32

	// Budget section.
	drawLabel(screen, x, y, "--- Budget ---", ColourAccent, face)
	y += 18
	drawLabel(screen, x, y,
		fmt.Sprintf("Quarterly discretionary: GBP %.0f m", world.LastBudget.TotalGBPm),
		ColourTextPrimary, face)
	y += 18
	drawLabel(screen, x, y,
		fmt.Sprintf("Tax revenue: GBP %.2f bn", world.LastTaxRevenue.GBPBillions),
		ColourTextPrimary, face)
	y += 32

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
		y += 18
		if y > cy+ch-10 {
			break
		}
	}
}

// colour is a small helper to build color.RGBA inline.
func colour(r, g, b uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: 0xFF}
}
