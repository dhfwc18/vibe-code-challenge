package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/region"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
)

// mapTabState holds selection state for the map tab.
type mapTabState struct {
	selectedRegion string
}

// drawTabMap renders the map placeholder tab.
func drawTabMap(screen *ebiten.Image, world simulation.WorldState, state *mapTabState, face font.Face, cx, cy, cw, ch int) {
	drawPanel(screen, cx, cy, cw, ch)
	x := cx + 12
	y := cy + 16

	drawLabel(screen, x, y, "Map -- placeholder (map geometry asset pending)", ColourAccent, face)
	y += 24

	// Build region summary: group tiles by RegionID.
	type regionRow struct {
		id    string
		name  string
		count int
		avgFP float64
		avgIn float64
	}
	regionMap := make(map[string]*regionRow)
	regionOrder := []string{}
	for _, r := range world.Regions {
		regionMap[r.ID] = &regionRow{id: r.ID, name: r.Name}
		regionOrder = append(regionOrder, r.ID)
	}
	for _, t := range world.Tiles {
		if row, ok := regionMap[t.RegionID]; ok {
			row.count++
			row.avgFP += t.FuelPoverty
			row.avgIn += t.InsulationLevel
		}
	}
	for _, id := range regionOrder {
		row := regionMap[id]
		if row.count > 0 {
			row.avgFP /= float64(row.count)
			row.avgIn /= float64(row.count)
		}
	}

	// Column headers.
	drawLabel(screen, x, y, "Region", ColourTextMuted, face)
	drawLabel(screen, x+160, y, "Tiles", ColourTextMuted, face)
	drawLabel(screen, x+210, y, "Avg FuelPov", ColourTextMuted, face)
	drawLabel(screen, x+330, y, "Avg Insul", ColourTextMuted, face)
	y += 16

	for _, id := range regionOrder {
		row := regionMap[id]
		lineCol := ColourTextPrimary
		if state.selectedRegion == id {
			lineCol = ColourAccent
		}
		drawLabel(screen, x, y, row.name, lineCol, face)
		drawLabel(screen, x+160, y, fmt.Sprintf("%d", row.count), lineCol, face)
		drawLabel(screen, x+210, y, fmt.Sprintf("%.1f", row.avgFP), lineCol, face)
		drawLabel(screen, x+330, y, fmt.Sprintf("%.1f", row.avgIn), lineCol, face)
		y += 14
		if y > cy+ch-60 {
			break
		}
	}

	// Expanded tile list for selected region.
	if state.selectedRegion != "" {
		y += 8
		drawLabel(screen, x, y, "Tiles in region "+state.selectedRegion+":", ColourAccent, face)
		y += 16
		drawLabel(screen, x, y, "Tile ID", ColourTextMuted, face)
		drawLabel(screen, x+160, y, "FuelPov", ColourTextMuted, face)
		drawLabel(screen, x+250, y, "Insul", ColourTextMuted, face)
		drawLabel(screen, x+310, y, "Pol.Opinion", ColourTextMuted, face)
		y += 14
		for _, t := range world.Tiles {
			if t.RegionID != state.selectedRegion {
				continue
			}
			drawTileRow(screen, t, face, x, y)
			y += 13
			if y > cy+ch-10 {
				break
			}
		}
	}
}

func drawTileRow(screen *ebiten.Image, t region.Tile, face font.Face, x, y int) {
	drawLabel(screen, x, y, t.ID, ColourTextPrimary, face)
	drawLabel(screen, x+160, y, fmt.Sprintf("%.1f", t.FuelPoverty), ColourTextPrimary, face)
	drawLabel(screen, x+250, y, fmt.Sprintf("%.1f", t.InsulationLevel), ColourTextPrimary, face)
	drawLabel(screen, x+310, y, fmt.Sprintf("%.1f", t.PoliticalOpinion), ColourTextPrimary, face)
}
