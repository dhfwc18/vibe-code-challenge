package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
)

// mapOverlay selects which data field is heat-mapped onto the region polygons.
type mapOverlay int

const (
	overlayFuelPoverty mapOverlay = iota
	overlayPolitical
	overlayInsulation
)

// mapTabState holds selection and overlay state for the map tab.
type mapTabState struct {
	selectedRegion string
	overlay        mapOverlay
}

// mapDetailPanelW is the pixel width of the right-side parliament panel within the map tab.
const mapDetailPanelW = 360

// drawTabMap renders the interactive vector-polygon map tab.
// The right-side panel shows the parliament box (hemicycle + party list).
func drawTabMap(
	screen *ebiten.Image,
	world simulation.WorldState,
	state *mapTabState,
	parliamentState *parliamentTabState,
	pendingActions *[]simulation.Action,
	face font.Face,
	cx, cy, cw, ch, effectiveAP int,
) {
	drawPanel(screen, cx, cy, cw, ch)

	// Overlay selector buttons across the top of the map panel.
	overlayNames := []string{"Fuel Poverty", "Politics", "Insulation"}
	for i, name := range overlayNames {
		bx := cx + 8 + i*110
		by := cy + 8
		bg := ColourButtonNormal
		if state.overlay == mapOverlay(i) {
			bg = colour(0x1A, 0x6B, 0x3A)
		} else if isHovered(bx, by, 108, 20) {
			bg = ColourButtonHover
		}
		solidRect(screen, bx, by, 108, 20, bg)
		drawLabel(screen, bx+6, by+15, name, ColourTextPrimary, face)
	}

	// Map canvas bounds: polygon area fills all but the right-side parliament panel.
	polyW := cw - mapDetailPanelW - 16
	if polyW < 64 {
		polyW = 64
	}
	mx := cx + 8
	my := cy + 36
	mw := polyW
	mh := ch - 44
	if maxH := mw * 14 / 10; mh > maxH {
		mh = maxH
	}

	// Sea background (dark teal).
	solidRect(screen, mx, my, mw, mh, colour(0x0E, 0x1E, 0x2A))

	// Build per-region aggregate values for the overlay colour.
	type regionAgg struct {
		sum   float64
		count int
	}
	agg := make(map[string]*regionAgg, len(world.Regions))
	for _, r := range world.Regions {
		agg[r.ID] = &regionAgg{}
	}
	for _, t := range world.Tiles {
		if a, ok := agg[t.RegionID]; ok {
			switch state.overlay {
			case overlayFuelPoverty:
				a.sum += t.FuelPoverty
			case overlayPolitical:
				a.sum += t.PoliticalOpinion
			case overlayInsulation:
				a.sum += t.InsulationLevel
			}
			a.count++
		}
	}

	omx := float32(mx)
	omy := float32(my)
	omw := float32(mw)
	omh := float32(mh)
	selectedCol := colour(0xF0, 0xE0, 0x60)

	borderCol := colour(0x0A, 0x16, 0x0C)
	selectedBorderCol := colour(0xFF, 0xFF, 0x80)

	// Draw filled polygons in fixed order so rendering is stable.
	for _, regID := range regionDrawOrder {
		pts := regionPolygons[regID]
		var val float64
		if a := agg[regID]; a != nil && a.count > 0 {
			val = a.sum / float64(a.count)
		}
		fill := overlayColour(state.overlay, val)
		if regID == state.selectedRegion {
			fill = selectedCol
		}
		fillMapPolygon(screen, pts, fill, omx, omy, omw, omh)
	}

	// Draw polygon borders on top of all fills so edges are always visible.
	for _, regID := range regionDrawOrder {
		pts := regionPolygons[regID]
		bc := borderCol
		if regID == state.selectedRegion {
			bc = selectedBorderCol
		}
		strokeMapPolygon(screen, pts, bc, omx, omy, omw, omh)
	}

	// Region short-name labels at polygon centroids.
	for _, regID := range regionDrawOrder {
		pts := regionPolygons[regID]
		lbl := regionShortName[regID]
		if lbl == "" {
			continue
		}
		lx, ly := polygonLabelPos(pts, omx, omy, omw, omh)
		drawLabel(screen, int(lx)-len(lbl)*3, int(ly)+4, lbl, colour(0xFF, 0xFF, 0xFF), face)
	}

	// Parliament panel on the right.
	drawParliamentPanel(screen, world, pendingActions, face,
		cx+cw-mapDetailPanelW, cy, mapDetailPanelW, ch, effectiveAP, parliamentState)
}
