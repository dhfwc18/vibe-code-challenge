package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
)

// drawTabEnergy renders the energy tab.
func drawTabEnergy(screen *ebiten.Image, world simulation.WorldState, face font.Face, cx, cy, cw, ch int) {
	drawPanel(screen, cx, cy, cw, ch)
	x := cx + 12
	y := cy + 16

	drawLabel(screen, x, y, "--- Energy Market ---", ColourAccent, face)
	y += 20

	// Price displays.
	drawLabel(screen, x, y, fmt.Sprintf("Gas:          GBP %.2f / MWh", world.EnergyMarket.GasPrice), ColourTextPrimary, face)
	y += 16
	drawLabel(screen, x, y, fmt.Sprintf("Electricity:  GBP %.2f / MWh", world.EnergyMarket.ElectricityPrice), ColourTextPrimary, face)
	y += 16
	drawLabel(screen, x, y, fmt.Sprintf("Oil:          GBP %.2f / MWh", world.EnergyMarket.OilPrice), ColourTextPrimary, face)
	y += 24

	// Renewable grid share (value is 0-1; multiply by 100 for display).
	renewShare := world.EnergyMarket.RenewableGridShare * 100
	drawLabel(screen, x, y, fmt.Sprintf("Renewable grid share: %.1f%%", renewShare), ColourTextPrimary, face)
	y += 6
	drawBar(screen, x, y, 300, 12, renewShare, 100, ColourAccent, ColourButtonNormal)
	y += 24

	// Fossil dependency.
	drawLabel(screen, x, y, fmt.Sprintf("Fossil dependency: %.1f%%", world.FossilDependency), ColourTextPrimary, face)
	y += 6
	drawBar(screen, x, y, 300, 12, world.FossilDependency, 100, colour(0xE6, 0x7E, 0x22), ColourButtonNormal)
	y += 24

	// Technology maturity table.
	drawLabel(screen, x, y, "--- Technology Maturity ---", ColourAccent, face)
	y += 18

	techs := []config.Technology{
		config.TechOffshoreWind,
		config.TechOnshoreWind,
		config.TechSolarPV,
		config.TechNuclear,
		config.TechHeatPumps,
		config.TechEVs,
		config.TechHydrogen,
		config.TechIndustrialCCS,
	}

	for _, tech := range techs {
		maturity, ok := world.Tech.Maturities[tech]
		if !ok {
			maturity = 0
		}
		// maturity is 0-1; multiply by 100 for display.
		maturityPct := maturity * 100
		label := fmt.Sprintf("%-20s %.1f%%", string(tech), maturityPct)
		drawLabel(screen, x, y, label, ColourTextPrimary, face)
		drawBar(screen, x+230, y-11, 150, 10, maturityPct, 100, ColourAccent, ColourButtonNormal)
		y += 16
		if y > cy+ch-10 {
			break
		}
	}
}
