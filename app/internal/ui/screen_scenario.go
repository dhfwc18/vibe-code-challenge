package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// ScenarioScreen renders the campaign start-point selection UI.
type ScenarioScreen struct {
	face font.Face
}

// NewScenarioScreen creates a new ScenarioScreen.
func NewScenarioScreen() *ScenarioScreen {
	return &ScenarioScreen{
		face: basicfont.Face7x13,
	}
}

// Update checks for a card click and returns the selected ScenarioID, or "" if none yet.
func (s *ScenarioScreen) Update(scenarios []config.ScenarioConfig) config.ScenarioID {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return ""
	}
	sw, sh := ebiten.WindowSize()
	mx, my := ebiten.CursorPosition()
	cardX, cardY, cardW, cardH := scenarioCardBounds(sw, sh, len(scenarios))
	for i, sc := range scenarios {
		cx := cardX + i*(cardW+scenarioCardGap)
		btnX := cx + 20
		btnY := cardY + cardH - 40
		btnW := cardW - 40
		btnH := 28
		if inRect(mx, my, btnX, btnY, btnW, btnH) {
			return sc.ID
		}
	}
	return ""
}

// Draw renders the scenario selection screen onto screen.
func (s *ScenarioScreen) Draw(screen *ebiten.Image, scenarios []config.ScenarioConfig) {
	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()

	// Background.
	screen.Fill(colour(0x0A, 0x14, 0x0F))

	// Title.
	title := "20-50  --  Select Your Campaign"
	drawLabel(screen, sw/2-len(title)*3, 28, title, ColourAccent, s.face)

	cardX, cardY, cardW, cardH := scenarioCardBounds(sw, sh, len(scenarios))

	for i, sc := range scenarios {
		cx := cardX + i*(cardW+scenarioCardGap)
		drawScenarioCard(screen, sc, cx, cardY, cardW, cardH, s.face)
	}
}

// scenarioCardGap is the horizontal gap between scenario cards.
const scenarioCardGap = 32

// scenarioCardBounds returns the origin and size of the first card, given the screen
// dimensions and the number of cards.
func scenarioCardBounds(sw, sh, n int) (cardX, cardY, cardW, cardH int) {
	cardW = 360
	cardH = 560
	totalW := n*cardW + (n-1)*scenarioCardGap
	if totalW > sw-40 {
		cardW = (sw - 40 - (n-1)*scenarioCardGap) / n
		totalW = n*cardW + (n-1)*scenarioCardGap
	}
	cardX = (sw - totalW) / 2
	cardY = (sh-cardH)/2 - 10
	if cardY < 50 {
		cardY = 50
	}
	return
}

// drawScenarioCard renders one scenario card at (cx, cy) with size (cw, ch).
func drawScenarioCard(screen *ebiten.Image, sc config.ScenarioConfig, cx, cy, cw, ch int, face font.Face) {
	mx, my := ebiten.CursorPosition()

	hovered := inRect(mx, my, cx, cy, cw, ch)

	// Card background.
	cardBg := colour(0x14, 0x22, 0x1A)
	if hovered {
		cardBg = colour(0x1A, 0x2E, 0x22)
	}
	solidRect(screen, cx, cy, cw, ch, cardBg)

	// Border (accent if hovered).
	borderCol := colour(0x2E, 0x45, 0x38)
	if hovered {
		borderCol = ColourAccent
	}
	// Top border.
	solidRect(screen, cx, cy, cw, 2, borderCol)
	// Bottom border.
	solidRect(screen, cx, cy+ch-2, cw, 2, borderCol)
	// Left border.
	solidRect(screen, cx, cy, 2, ch, borderCol)
	// Right border.
	solidRect(screen, cx+cw-2, cy, 2, ch, borderCol)

	x := cx + 16
	y := cy + 18

	// Year badge.
	yearBadge := sc.ShortName
	drawBadge(screen, cx+cw-50, cy+10, yearBadge, colour(0x1A, 0x6B, 0x3A), face)

	// Scenario name.
	drawLabel(screen, x, y, sc.Name, ColourAccent, face)
	y += 22

	// Party line.
	partyStr := "Party: " + string(sc.InitialParty)
	partyCol := colour(0xA8, 0xC8, 0xFF)
	if sc.InitialParty == config.PartyRight {
		partyCol = colour(0xFF, 0xC8, 0xA8)
	}
	drawLabel(screen, x, y, partyStr, partyCol, face)
	y += 18

	// Divider.
	solidRect(screen, cx+12, y, cw-24, 1, colour(0x2E, 0x45, 0x38))
	y += 10

	// Description -- wrapped at cw-32 pixels (approx 7px per char).
	charsPerLine := (cw - 32) / 7
	if charsPerLine < 20 {
		charsPerLine = 20
	}
	desc := sc.Description
	for len(desc) > 0 {
		line := desc
		if len(line) > charsPerLine {
			cut := charsPerLine
			for cut > 0 && desc[cut] != ' ' {
				cut--
			}
			if cut == 0 {
				cut = charsPerLine
			}
			line = desc[:cut]
			desc = desc[cut+1:]
		} else {
			desc = ""
		}
		drawLabel(screen, x, y, line, ColourTextPrimary, face)
		y += 14
	}
	y += 10

	// Key stats.
	solidRect(screen, cx+12, y, cw-24, 1, colour(0x2E, 0x45, 0x38))
	y += 10
	drawLabel(screen, x, y, "--- Starting Conditions ---", ColourTextMuted, face)
	y += 16
	stats := []string{
		fmt.Sprintf("Carbon output:   %.0f MtCO2e/yr", sc.InitialCarbonMt),
		fmt.Sprintf("Fossil dependence: %.0f%%", sc.InitialFossilDep),
		fmt.Sprintf("Budget:          GBP %.0fm", sc.InitialBudget),
		fmt.Sprintf("Popularity:      %.0f%%", sc.InitialPopularity),
	}
	if sc.ScandalRateMultiplier > 1.0 {
		stats = append(stats, fmt.Sprintf("Scandal risk:    x%.0f", sc.ScandalRateMultiplier))
	}
	electionYears := sc.ElectionDueWeek / 52
	stats = append(stats, fmt.Sprintf("First election:  ~%d yr", electionYears))
	for _, st := range stats {
		drawLabel(screen, x, y, st, ColourTextPrimary, face)
		y += 14
	}

	// "Start Campaign" button at bottom.
	btnX := cx + 20
	btnY := cy + ch - 40
	btnW := cw - 40
	btnH := 28
	btnBg := buttonColour(btnX, btnY, btnW, btnH, true)
	solidRect(screen, btnX, btnY, btnW, btnH, btnBg)
	btnLabel := "Start Campaign"
	labelX := btnX + (btnW-len(btnLabel)*7)/2
	drawLabel(screen, labelX, btnY+19, btnLabel, ColourTextPrimary, face)

}
