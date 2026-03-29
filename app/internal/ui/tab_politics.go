package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"github.com/vibe-code-challenge/twenty-fifty/internal/stakeholder"
	"golang.org/x/image/font"
)

// politicsTabState tracks selection and scroll offset in the politics tab.
type politicsTabState struct {
	selectedID  string // "" = no politician selected (show grid)
	scrollY     int    // future: scroll offset for large rosters
}

// politicsCardW / politicsCardH are the dimensions of a politician card in the grid.
const (
	politicsCardW   = 210
	politicsCardH   = 84
	politicsCardGap = 8
)

// roleAbbrev returns a short display string for a minister role.
func roleAbbrev(r config.Role) string {
	switch r {
	case config.RoleLeader:
		return "Leader"
	case config.RoleChancellor:
		return "Chanc."
	case config.RoleForeignSecretary:
		return "F.Sec."
	case config.RoleEnergy:
		return "Energy"
	default:
		return string(r)
	}
}

// drawTabPolitics renders the politics panel.
// It shows either the politician grid or a single profile depending on politicsTabState.
func drawTabPolitics(
	screen *ebiten.Image,
	world simulation.WorldState,
	pendingActions *[]simulation.Action,
	face font.Face,
	cx, cy, cw, ch int,
	effectiveAP int,
	state *politicsTabState,
) {
	drawPanel(screen, cx, cy, cw, ch)

	if state.selectedID != "" {
		// Find the selected stakeholder.
		for i := range world.Stakeholders {
			if world.Stakeholders[i].ID == state.selectedID {
				drawPoliticianProfile(screen, world.Stakeholders[i], world, pendingActions, face,
					cx, cy, cw, ch, effectiveAP)
				return
			}
		}
		// If not found (e.g., departed), fall back to grid.
		state.selectedID = ""
	}

	drawPoliticianGrid(screen, world, face, cx, cy, cw, ch, state)
}

// drawPoliticianGrid draws the full grid of all politicians (locked and unlocked).
func drawPoliticianGrid(
	screen *ebiten.Image,
	world simulation.WorldState,
	face font.Face,
	cx, cy, cw, ch int,
	state *politicsTabState,
) {
	// Section header.
	drawLabel(screen, cx+12, cy+18, "Politicians", ColourAccent, face)
	drawLabel(screen, cx+cw-200, cy+18, "Click a card to view profile", ColourTextMuted, face)

	// Party filter strip below header.
	partyOrder := []config.Party{
		config.PartyLeft, config.PartyRight,
		config.PartyFarLeft, config.PartyFarRight,
	}
	headerY := cy + 24

	// Grid layout.
	cols := (cw - politicsCardGap) / (politicsCardW + politicsCardGap)
	if cols < 1 {
		cols = 1
	}
	gridX := cx + politicsCardGap
	gridY := headerY + 12

	col := 0
	row := 0
	mx, my := ebiten.CursorPosition()

	for _, party := range partyOrder {
		for i := range world.Stakeholders {
			s := world.Stakeholders[i]
			if s.Party != party {
				continue
			}
			cardX := gridX + col*(politicsCardW+politicsCardGap)
			cardY := gridY + row*(politicsCardH+politicsCardGap)
			if cardY+politicsCardH > cy+ch-4 {
				break // out of visible area
			}
			hovered := inRect(mx, my, cardX, cardY, politicsCardW, politicsCardH)
			drawPoliticianCard(screen, s, world, face, cardX, cardY, hovered)
			col++
			if col >= cols {
				col = 0
				row++
			}
		}
	}
}

// drawPoliticianCard draws a single compact politician card.
func drawPoliticianCard(
	screen *ebiten.Image,
	s stakeholder.Stakeholder,
	world simulation.WorldState,
	face font.Face,
	x, y int,
	hovered bool,
) {
	// Card background.
	bg := colour(0x16, 0x28, 0x1C)
	if hovered {
		bg = colour(0x1E, 0x36, 0x26)
	}
	if !s.IsUnlocked {
		bg = colour(0x12, 0x1C, 0x16)
	}
	solidRect(screen, x, y, politicsCardW, politicsCardH, bg)

	// Left party stripe (3px).
	solidRect(screen, x, y, 3, politicsCardH, partyColour(s.Party))

	// Top border.
	borderCol := colour(0x2E, 0x45, 0x38)
	if hovered {
		borderCol = ColourAccent
	}
	solidRect(screen, x, y, politicsCardW, 1, borderCol)
	solidRect(screen, x, y+politicsCardH-1, politicsCardW, 1, borderCol)
	solidRect(screen, x+politicsCardW-1, y, 1, politicsCardH, borderCol)

	tx := x + 8

	// Name.
	nameCol := ColourTextPrimary
	if !s.IsUnlocked {
		nameCol = ColourTextMuted
	}
	name := s.Name
	maxChars := (politicsCardW - 12) / 7
	if len(name) > maxChars {
		name = name[:maxChars-2] + ".."
	}
	drawLabel(screen, tx, y+14, name, nameCol, face)

	// Role badge + state badge on row 2.
	if s.IsUnlocked {
		drawBadge(screen, tx, y+18, roleAbbrev(s.Role), ColourOrgThinkTank, face)
		stateCol := ministerStateColour(s.State)
		stateStr := stateAbbrev(s.State)
		drawBadge(screen, tx+64, y+18, stateStr, stateCol, face)

		// Party name right-aligned.
		pName := config.PartyNames[s.Party]
		if len(pName) > 8 {
			pName = pName[:8]
		}
		drawLabel(screen, x+politicsCardW-len(pName)*7-4, y+14, pName, partyColour(s.Party), face)

		// Popularity and relationship bars on rows 3-4.
		drawLabel(screen, tx, y+50, "Pop", ColourTextMuted, face)
		drawBar(screen, tx+26, y+40, politicsCardW-40, 7, s.Popularity, 100, ColourAccent, ColourButtonNormal)
		drawLabel(screen, tx, y+64, "Rel", ColourTextMuted, face)
		drawBar(screen, tx+26, y+54, politicsCardW-40, 7, s.RelationshipScore, 100, ColourOrgThinkTank, ColourButtonNormal)

		// Cabinet star.
		if isInCabinet(s, world) {
			drawLabel(screen, x+politicsCardW-12, y+14, "*", ColourAccent, face)
		}
	} else {
		drawLabel(screen, tx, y+40, "-- locked --", ColourTextMuted, face)
	}
}

// drawPoliticianProfile draws the full-detail profile for the selected politician.
func drawPoliticianProfile(
	screen *ebiten.Image,
	s stakeholder.Stakeholder,
	world simulation.WorldState,
	pendingActions *[]simulation.Action,
	face font.Face,
	cx, cy, cw, ch int,
	effectiveAP int,
) {
	x := cx + 16
	y := cy + 16

	// Back label at top-left.
	drawLabel(screen, x, y+12, "<< Back to grid", ColourTextMuted, face)
	y += 28

	// Party stripe header bar.
	solidRect(screen, cx, y, cw, 24, partyColour(s.Party))
	drawLabel(screen, cx+8, y+17, config.PartyNames[s.Party]+" | "+roleAbbrev(s.Role), ColourTextPrimary, face)
	inCab := isInCabinet(s, world)
	if inCab {
		drawLabel(screen, cx+cw-80, y+17, "* Cabinet", ColourAccent, face)
	}
	y += 30

	// Name + nickname.
	drawLabel(screen, x, y+14, s.Name, ColourAccent, face)
	if s.Nickname != "" {
		drawLabel(screen, x+len(s.Name)*7+8, y+14, "(\""+s.Nickname+"\")", ColourTextMuted, face)
	}
	y += 20

	// State badge.
	stateCol := ministerStateColour(s.State)
	drawBadge(screen, x, y, string(s.State), stateCol, face)
	y += 24

	// Divider.
	solidRect(screen, cx+8, y, cw-16, 1, colour(0x2E, 0x45, 0x38))
	y += 10

	// Biography, wrapped at panel width.
	if s.Biography != "" {
		drawLabel(screen, x, y+12, "Background", ColourTextMuted, face)
		y += 16
		bio := s.Biography
		charsPerLine := (cw - 32) / 7
		if charsPerLine < 20 {
			charsPerLine = 20
		}
		for len(bio) > 0 {
			line := bio
			if len(line) > charsPerLine {
				cut := charsPerLine
				for cut > 0 && bio[cut] != ' ' {
					cut--
				}
				if cut == 0 {
					cut = charsPerLine
				}
				line = bio[:cut]
				bio = bio[cut+1:]
			} else {
				bio = ""
			}
			drawLabel(screen, x, y+12, line, ColourTextPrimary, face)
			y += 14
		}
		y += 6
	}

	// Divider.
	solidRect(screen, cx+8, y, cw-16, 1, colour(0x2E, 0x45, 0x38))
	y += 12

	// Attribute bars.
	barW := cw - 80
	type attr struct {
		label string
		val   float64
		col   color.RGBA
	}
	attrs := []attr{
		{"Ideology  ", s.IdeologyScore + 100, colour(0xA0, 0x60, 0xE0)}, // -100..+100 → 0..200 shown as 0..100
		{"Net Zero  ", s.NetZeroSympathy, colour(0x27, 0xAE, 0x60)},
		{"Risk Tol. ", s.RiskTolerance, colour(0xE6, 0x7E, 0x22)},
		{"Populism  ", s.PopulismScore, colour(0xE7, 0x4C, 0x3C)},
		{"Diplomatic", s.DiplomaticSkill, colour(0x29, 0x80, 0xB9)},
		{"Popularity", s.Popularity, ColourAccent},
		{"Relation  ", s.RelationshipScore, ColourOrgThinkTank},
	}
	for _, a := range attrs {
		drawLabel(screen, x, y+10, a.label, ColourTextMuted, face)
		maxVal := 100.0
		if a.label == "Ideology  " {
			maxVal = 200 // normalized to 0..200 for display
		}
		drawBar(screen, x+80, y+2, barW, 10, a.val, maxVal, a.col, ColourButtonNormal)
		drawLabel(screen, x+80+barW+4, y+10, fmt.Sprintf("%.0f", a.val), ColourTextMuted, face)
		y += 18
	}
	y += 6

	// Divider.
	solidRect(screen, cx+8, y, cw-16, 1, colour(0x2E, 0x45, 0x38))
	y += 12

	// Signals (personality traits).
	if len(s.Signals) > 0 {
		drawLabel(screen, x, y+12, "Signals:", ColourTextMuted, face)
		sigX := x + 56
		for _, sig := range s.Signals {
			w := len(sig)*7 + 12
			if sigX+w > cx+cw-8 {
				sigX = x + 56
				y += 16
			}
			drawBadge(screen, sigX, y, sig, colour(0x2E, 0x45, 0x38), face)
			sigX += w + 6
		}
		y += 20
	}

	// Lobby button (only active if in cabinet and enough AP).
	canLobby := inCab && effectiveAP >= 3
	btnY := cy + ch - 44
	btnX := x
	btnW := 120
	btnH := 28
	btnCol := buttonColour(btnX, btnY, btnW, btnH, canLobby)
	solidRect(screen, btnX, btnY, btnW, btnH, btnCol)
	lblCol := ColourTextPrimary
	if !canLobby {
		lblCol = ColourTextMuted
	}
	drawLabel(screen, btnX+8, btnY+19, "Lobby (3 AP)", lblCol, face)
	if !inCab {
		drawLabel(screen, btnX+btnW+8, btnY+19, "Not in cabinet", ColourTextMuted, face)
	} else if effectiveAP < 3 {
		drawLabel(screen, btnX+btnW+8, btnY+19, "Need 3 AP", colour(0xE7, 0x4C, 0x3C), face)
	}
	_ = pendingActions // lobby click handled in handlePoliticsProfileClick
}

// isInCabinet returns true if s holds a cabinet role in the current government.
func isInCabinet(s stakeholder.Stakeholder, world simulation.WorldState) bool {
	for _, id := range world.Government.CabinetByRole {
		if id == s.ID {
			return true
		}
	}
	return false
}

// stateAbbrev shortens long state strings for card display.
func stateAbbrev(st stakeholder.MinisterState) string {
	switch st {
	case stakeholder.MinisterStateActive:
		return "Active"
	case stakeholder.MinisterStateAppointed:
		return "New"
	case stakeholder.MinisterStateUnderPressure:
		return "Pressure"
	case stakeholder.MinisterStateLeadershipChallenge:
		return "Challenge"
	case stakeholder.MinisterStateDeparted:
		return "Gone"
	case stakeholder.MinisterStateBackbench:
		return "Backbench"
	case stakeholder.MinisterStateOppositionShadow:
		return "Shadow"
	case stakeholder.MinisterStateSacked:
		return "Sacked"
	case stakeholder.MinisterStateResigned:
		return "Resigned"
	case stakeholder.MinisterStateElectionOut:
		return "Lost seat"
	default:
		return string(st)
	}
}

// partyColour returns the header colour for a party column.
func partyColour(p config.Party) color.RGBA {
	switch p {
	case config.PartyLeft:
		return ColourPartyLeft
	case config.PartyRight:
		return ColourPartyRight
	case config.PartyFarLeft:
		return ColourPartyFarLeft
	case config.PartyFarRight:
		return ColourPartyFarRight
	default:
		return ColourPartyNeutral
	}
}

// ministerStateColour maps minister state to a display colour.
func ministerStateColour(state stakeholder.MinisterState) color.RGBA {
	switch state {
	case stakeholder.MinisterStateActive, stakeholder.MinisterStateAppointed:
		return colour(0x27, 0xAE, 0x60)
	case stakeholder.MinisterStateUnderPressure:
		return colour(0xF3, 0x9C, 0x12)
	case stakeholder.MinisterStateLeadershipChallenge:
		return colour(0xE7, 0x4C, 0x3C)
	case stakeholder.MinisterStateDeparted,
		stakeholder.MinisterStateSacked,
		stakeholder.MinisterStateResigned,
		stakeholder.MinisterStateElectionOut:
		return ColourPartyNeutral
	case stakeholder.MinisterStateBackbench,
		stakeholder.MinisterStateOppositionShadow:
		return ColourTextMuted
	default:
		return ColourTextMuted
	}
}
