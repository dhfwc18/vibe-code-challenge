package ui

import (
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/industry"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
)

// industryFilter controls which companies are shown.
type industryFilter int

const (
	filterAll industryFilter = iota
	filterActive
	filterStruggling
	filterBankrupt
)

// industryTabState holds filter state for the industry tab.
type industryTabState struct {
	filter industryFilter
}

// drawTabIndustry renders the industry tab.
func drawTabIndustry(
	screen *ebiten.Image,
	world simulation.WorldState,
	state *industryTabState,
	face font.Face,
	cx, cy, cw, ch int,
) {
	drawPanel(screen, cx, cy, cw, ch)
	x := cx + 12
	y := cy + 16

	// Filter buttons.
	filters := []string{"All", "Active", "Struggling", "Bankrupt"}
	btnW := 80
	for i, label := range filters {
		btnX := x + i*btnW
		bg := ColourButtonNormal
		if industryFilter(i) == state.filter {
			bg = ColourButtonHover
		}
		solidRect(screen, btnX, y-12, btnW-2, 16, bg)
		drawLabel(screen, btnX+4, y, label, ColourTextPrimary, face)
	}
	y += 10

	// Column headers.
	drawLabel(screen, x, y, "Company", ColourTextMuted, face)
	drawLabel(screen, x+200, y, "Category", ColourTextMuted, face)
	drawLabel(screen, x+340, y, "Status", ColourTextMuted, face)
	drawLabel(screen, x+440, y, "Work", ColourTextMuted, face)
	drawLabel(screen, x+520, y, "Quality", ColourTextMuted, face)
	y += 16

	// Sort companies by ID for stable ordering.
	ids := make([]string, 0, len(world.Industry.Companies))
	for id := range world.Industry.Companies {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		cs := world.Industry.Companies[id]

		// Apply filter.
		if !showCompany(cs, state.filter) {
			continue
		}

		// Look up company name.
		name := id
		if world.Cfg != nil {
			for _, def := range world.Cfg.Companies {
				if def.ID == id {
					name = def.Name
					break
				}
			}
		}

		drawLabel(screen, x, y, name, ColourTextPrimary, face)

		// Tech category badge.
		drawBadge(screen, x+200, y-12, string(cs.ContractedTech), techCatColour(cs.ContractedTech), face)

		// Status badge.
		drawBadge(screen, x+340, y-12, string(cs.Status), companyStatusColour(cs.Status), face)

		// Work rate bar.
		drawBar(screen, x+440, y-11, 70, 10, cs.WorkRate, 100, ColourAccent, ColourButtonNormal)

		// Accumulated quality bar (0-200 scale, threshold at 200).
		qualCol := ColourOrgThinkTank
		if cs.AccumulatedQuality >= 200 {
			qualCol = ColourAccent
		}
		drawBar(screen, x+520, y-11, 70, 10, cs.AccumulatedQuality, 200, qualCol, ColourButtonNormal)

		y += 16
		if y > cy+ch-10 {
			break
		}
	}
}

// showCompany returns true if the company should be shown given the current filter.
func showCompany(cs industry.CompanyState, f industryFilter) bool {
	switch f {
	case filterActive:
		return cs.Status == industry.CompanyStatusActive
	case filterStruggling:
		return cs.Status == industry.CompanyStatusStruggling
	case filterBankrupt:
		return cs.Status == industry.CompanyStatusBankrupt
	default:
		return true
	}
}

// techCatColour returns a colour for a tech category badge.
func techCatColour(tc config.Technology) color.RGBA {
	switch tc {
	case config.TechOffshoreWind, config.TechOnshoreWind:
		return ColourAccent
	case config.TechSolarPV:
		return colour(0xF3, 0x9C, 0x12)
	case config.TechNuclear:
		return colour(0x8E, 0x44, 0xAD)
	case config.TechHeatPumps:
		return colour(0x29, 0x80, 0xB9)
	case config.TechEVs:
		return colour(0x16, 0xA0, 0x85)
	default:
		return ColourPartyNeutral
	}
}

// companyStatusColour returns a colour for a company status badge.
func companyStatusColour(s industry.CompanyStatus) color.RGBA {
	switch s {
	case industry.CompanyStatusActive:
		return colour(0x27, 0xAE, 0x60)
	case industry.CompanyStatusStruggling:
		return colour(0xF3, 0x9C, 0x12)
	case industry.CompanyStatusBankrupt:
		return colour(0xE7, 0x4C, 0x3C)
	case industry.CompanyStatusInactive, industry.CompanyStatusStartup:
		return ColourTextMuted
	default:
		return ColourPartyNeutral
	}
}

