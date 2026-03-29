package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
)

// budgetDepts is the ordered list of department IDs.
var budgetDepts = []string{
	"power",
	"transport",
	"buildings",
	"industry",
	"cross",
}

// budgetDeptNames maps department IDs to display names.
var budgetDeptNames = map[string]string{
	"power":     "Power",
	"transport": "Transport",
	"buildings": "Buildings",
	"industry":  "Industry",
	"cross":     "Cross-Cutting",
}

// drawTabBudget renders the budget tab.
func drawTabBudget(screen *ebiten.Image, world simulation.WorldState, face font.Face, cx, cy, cw, ch int) {
	drawPanel(screen, cx, cy, cw, ch)
	x := cx + 12
	y := cy + 16

	drawLabel(screen, x, y, "--- Department Budget ---", ColourAccent, face)
	y += 18

	// Column headers.
	drawLabel(screen, x, y, "Department", ColourTextMuted, face)
	drawLabel(screen, x+160, y, "Allocated GBPm", ColourTextMuted, face)
	drawLabel(screen, x+320, y, "Lobby Mult", ColourTextMuted, face)
	y += 16

	for _, deptID := range budgetDepts {
		name := budgetDeptNames[deptID]
		if name == "" {
			name = deptID
		}

		allocated := 0.0
		if world.LastBudget.Departments != nil {
			allocated = world.LastBudget.Departments[deptID]
		}

		lobbyMult := 1.0
		if world.Economy.LobbyEffects != nil {
			if v, ok := world.Economy.LobbyEffects[deptID]; ok {
				lobbyMult = v
			}
		}

		drawLabel(screen, x, y, name, ColourTextPrimary, face)
		drawLabel(screen, x+160, y, fmt.Sprintf("%.1f", allocated), ColourTextPrimary, face)
		drawLabel(screen, x+320, y, fmt.Sprintf("%.2f", lobbyMult), lobbyMultColour(lobbyMult), face)
		y += 16
	}

	y += 16
	drawLabel(screen, x, y,
		fmt.Sprintf("Total discretionary: GBP %.1f m", world.LastBudget.TotalGBPm),
		ColourTextPrimary, face)
	y += 16

	// Tax revenue.
	drawLabel(screen, x, y, "--- Tax Revenue ---", ColourAccent, face)
	y += 18
	drawLabel(screen, x, y,
		fmt.Sprintf("Q%d %d:  GBP %.2f bn",
			world.LastTaxRevenue.Quarter,
			world.LastTaxRevenue.Year,
			world.LastTaxRevenue.GBPBillions,
		),
		ColourTextPrimary, face)
}

// lobbyMultColour returns green when a lobby multiplier is above 1.0.
func lobbyMultColour(mult float64) color.RGBA {
	if mult > 1.0 {
		return ColourAccent
	}
	return ColourTextPrimary
}
