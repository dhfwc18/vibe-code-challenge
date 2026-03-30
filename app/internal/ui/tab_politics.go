package ui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"github.com/vibe-code-challenge/twenty-fifty/internal/stakeholder"
	"golang.org/x/image/font"
)

// parliamentTabState tracks navigation within the parliament panel embedded in the map tab.
// Empty selectedParty shows the hemicycle overview and party list.
// Non-empty selectedParty shows that party's member grid.
// Non-empty selectedID shows the full politician profile.
type parliamentTabState struct {
	selectedParty config.Party
	selectedID    string
}

const (
	parliamentCardH   = 84
	parliamentCardGap = 6
	hemicycleH        = 100 // pixel height of the hemicycle drawing box
	hemicycleTotal    = 50  // total seat bubbles displayed
)

// hemicycleArc describes one concentric row of the parliament fan chart.
type hemicycleArc struct {
	count  int
	radius int
}

var hemicycleArcs = []hemicycleArc{
	{10, 38},
	{17, 62},
	{23, 86},
}

// hemicyclePartyOrder is the left-to-right political spectrum order for seat assignment.
var hemicyclePartyOrder = []config.Party{
	config.PartyFarLeft,
	config.PartyLeft,
	config.PartyRight,
	config.PartyFarRight,
}

// hemicycleSeats converts a vote-share map into per-party seat counts summing to hemicycleTotal.
func hemicycleSeats(voteShare map[config.Party]float64) map[config.Party]int {
	total := 0.0
	for _, v := range voteShare {
		total += v
	}
	seats := make(map[config.Party]int, len(hemicyclePartyOrder))
	if total == 0 {
		n := hemicycleTotal / len(hemicyclePartyOrder)
		for _, p := range hemicyclePartyOrder {
			seats[p] = n
		}
		seats[config.PartyRight] += hemicycleTotal - n*len(hemicyclePartyOrder)
		return seats
	}
	assigned := 0
	for _, p := range hemicyclePartyOrder[:len(hemicyclePartyOrder)-1] {
		s := int(math.Round(voteShare[p] / total * float64(hemicycleTotal)))
		seats[p] = s
		assigned += s
	}
	last := hemicyclePartyOrder[len(hemicyclePartyOrder)-1]
	rem := hemicycleTotal - assigned
	if rem < 0 {
		rem = 0
	}
	seats[last] = rem
	return seats
}

// buildSeatColors returns a slice of hemicycleTotal colours, ordered left-to-right by party.
func buildSeatColors(voteShare map[config.Party]float64) []color.RGBA {
	seats := hemicycleSeats(voteShare)
	cols := make([]color.RGBA, 0, hemicycleTotal)
	for _, p := range hemicyclePartyOrder {
		for i := 0; i < seats[p]; i++ {
			cols = append(cols, partyColour(p))
		}
	}
	for len(cols) < hemicycleTotal {
		cols = append(cols, ColourPartyNeutral)
	}
	return cols[:hemicycleTotal]
}

// drawParliamentOverview renders the hemicycle fan chart and party list.
func drawParliamentOverview(
	screen *ebiten.Image,
	world simulation.WorldState,
	face font.Face,
	px, py, pw, ph int,
) {
	x := px + 12
	y := py + 12

	drawLabel(screen, x, y+12, "--- Parliament ---", ColourAccent, face)
	y += 20

	drawHemicycle(screen, world, px+4, y, pw-8, hemicycleH)
	y += hemicycleH + 8

	drawLabel(screen, x, y+12, "--- Parties ---", ColourAccent, face)
	y += 18

	rowH := 28
	for _, p := range hemicyclePartyOrder {
		isRuling := p == world.Government.RulingParty
		bg := colour(0x16, 0x28, 0x1C)
		if isRuling {
			bg = colour(0x1A, 0x3E, 0x28)
		}
		if isHovered(x, y, pw-24, rowH) {
			bg = ColourButtonHover
		}
		solidRect(screen, x, y, pw-24, rowH, bg)
		solidRect(screen, x, y, 4, rowH, partyColour(p))
		if isRuling {
			solidRect(screen, x, y+rowH-2, pw-24, 2, colour(0xF5, 0xE0, 0x42))
		}
		pName := config.PartyNames[p]
		if len(pName) > 18 {
			pName = pName[:18]
		}
		drawLabel(screen, x+10, y+18, pName, partyColour(p), face)
		leader := partyLeaderName(world, p)
		if len(leader) > 14 {
			leader = leader[:14]
		}
		drawLabel(screen, x+148, y+18, leader, ColourTextMuted, face)
		if isRuling {
			drawLabel(screen, x+pw-24-34, y+18, "Govt", colour(0xF5, 0xE0, 0x42), face)
		}
		y += rowH + 2
	}
}

// drawHemicycle renders the parliament fan-chart bubble display.
// The semicircle base is at the bottom of the drawing area, centered horizontally.
func drawHemicycle(
	screen *ebiten.Image,
	world simulation.WorldState,
	px, py, pw, ph int,
) {
	seatColors := buildSeatColors(world.Government.LastElectionVoteShare)

	cx := px + pw/2
	cy := py + ph - 4
	bubbleSize := 7

	seatIdx := 0
	for _, arc := range hemicycleArcs {
		r := arc.radius
		n := arc.count
		for j := 0; j < n && seatIdx < hemicycleTotal; j++ {
			theta := math.Pi * float64(n-j) / float64(n+1)
			bx := cx + int(math.Round(float64(r)*math.Cos(theta))) - bubbleSize/2
			by := cy - int(math.Round(float64(r)*math.Sin(theta))) - bubbleSize/2
			solidRect(screen, bx, by, bubbleSize, bubbleSize, seatColors[seatIdx])
			seatIdx++
		}
	}
}

// partyLeaderName returns the name of the first unlocked leader for party p.
func partyLeaderName(world simulation.WorldState, p config.Party) string {
	for _, s := range world.Stakeholders {
		if s.Party == p && s.Role == config.RoleLeader && s.IsUnlocked {
			return s.Name
		}
	}
	return "--"
}

// drawPartyMemberGrid renders a 2-column card grid for all members of a single party.
func drawPartyMemberGrid(
	screen *ebiten.Image,
	world simulation.WorldState,
	face font.Face,
	px, py, pw, ph int,
	party config.Party,
) {
	x := px + 12
	y := py + 12

	drawLabel(screen, x, y+12, "<< Back to parties", ColourTextMuted, face)
	y += 28

	pName := config.PartyNames[party]
	drawLabel(screen, x, y+12, pName, partyColour(party), face)
	y += 20

	cols := 2
	cardW := (pw - 24 - parliamentCardGap) / cols
	mx, my := ebiten.CursorPosition()

	col := 0
	row := 0
	for i := range world.Stakeholders {
		s := world.Stakeholders[i]
		if s.Party != party {
			continue
		}
		cardX := x + col*(cardW+parliamentCardGap)
		cardY := y + row*(parliamentCardH+parliamentCardGap)
		if cardY+parliamentCardH > py+ph-4 {
			break
		}
		hovered := inRect(mx, my, cardX, cardY, cardW, parliamentCardH)
		drawPoliticianCard(screen, s, world, face, cardX, cardY, cardW, parliamentCardH, hovered)
		col++
		if col >= cols {
			col = 0
			row++
		}
	}
}

// drawPoliticianCard draws a single compact politician card with explicit width and height.
func drawPoliticianCard(
	screen *ebiten.Image,
	s stakeholder.Stakeholder,
	world simulation.WorldState,
	face font.Face,
	x, y, w, h int,
	hovered bool,
) {
	bg := colour(0x16, 0x28, 0x1C)
	if hovered {
		bg = colour(0x1E, 0x36, 0x26)
	}
	if !s.IsUnlocked {
		bg = colour(0x12, 0x1C, 0x16)
	}
	solidRect(screen, x, y, w, h, bg)

	solidRect(screen, x, y, 3, h, partyColour(s.Party))

	borderCol := colour(0x2E, 0x45, 0x38)
	if hovered {
		borderCol = ColourAccent
	}
	solidRect(screen, x, y, w, 1, borderCol)
	solidRect(screen, x, y+h-1, w, 1, borderCol)
	solidRect(screen, x+w-1, y, 1, h, borderCol)

	tx := x + 8

	nameCol := ColourTextPrimary
	if !s.IsUnlocked {
		nameCol = ColourTextMuted
	}
	name := s.Name
	maxChars := (w - 12) / 7
	if len(name) > maxChars {
		name = name[:maxChars-2] + ".."
	}
	drawLabel(screen, tx, y+14, name, nameCol, face)

	if s.IsUnlocked {
		drawBadge(screen, tx, y+18, roleAbbrev(s.Role), ColourOrgThinkTank, face)
		stateCol := ministerStateColour(s.State)
		stateStr := stateAbbrev(s.State)
		drawBadge(screen, tx+64, y+18, stateStr, stateCol, face)

		pName := config.PartyNames[s.Party]
		if len(pName) > 8 {
			pName = pName[:8]
		}
		drawLabel(screen, x+w-len(pName)*7-4, y+14, pName, partyColour(s.Party), face)

		drawLabel(screen, tx, y+50, "Pop", ColourTextMuted, face)
		drawBar(screen, tx+26, y+40, w-40, 7, s.Popularity, 100, ColourAccent, ColourButtonNormal)
		drawLabel(screen, tx, y+64, "Rel", ColourTextMuted, face)
		drawBar(screen, tx+26, y+54, w-40, 7, s.RelationshipScore, 100, ColourOrgThinkTank, ColourButtonNormal)

		if isInCabinet(s, world) {
			drawLabel(screen, x+w-12, y+14, "*", ColourAccent, face)
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

	drawLabel(screen, x, y+12, "<< Back to party", ColourTextMuted, face)
	y += 28

	solidRect(screen, cx, y, cw, 24, partyColour(s.Party))
	drawLabel(screen, cx+8, y+17, config.PartyNames[s.Party]+" | "+roleAbbrev(s.Role), ColourTextPrimary, face)
	inCab := isInCabinet(s, world)
	if inCab {
		drawLabel(screen, cx+cw-80, y+17, "* Cabinet", ColourAccent, face)
	}
	y += 30

	drawLabel(screen, x, y+14, s.Name, ColourAccent, face)
	if s.Nickname != "" {
		drawLabel(screen, x+len(s.Name)*7+8, y+14, "(\""+s.Nickname+"\")", ColourTextMuted, face)
	}
	y += 20

	stateCol := ministerStateColour(s.State)
	drawBadge(screen, x, y, string(s.State), stateCol, face)
	y += 24

	solidRect(screen, cx+8, y, cw-16, 1, colour(0x2E, 0x45, 0x38))
	y += 10

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

	solidRect(screen, cx+8, y, cw-16, 1, colour(0x2E, 0x45, 0x38))
	y += 12

	barW := cw - 80
	type attr struct {
		label string
		val   float64
		col   color.RGBA
	}
	attrs := []attr{
		{"Ideology  ", s.IdeologyScore + 100, colour(0xA0, 0x60, 0xE0)},
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
			maxVal = 200
		}
		drawBar(screen, x+80, y+2, barW, 10, a.val, maxVal, a.col, ColourButtonNormal)
		drawLabel(screen, x+80+barW+4, y+10, fmt.Sprintf("%.0f", a.val), ColourTextMuted, face)
		y += 18
	}
	y += 6

	solidRect(screen, cx+8, y, cw-16, 1, colour(0x2E, 0x45, 0x38))
	y += 12

	if len(s.Signals) > 0 {
		drawLabel(screen, x, y+12, "Signals:", ColourTextMuted, face)
		sigX := x + 56
		for _, sig := range s.Signals {
			sw := len(sig)*7 + 12
			if sigX+sw > cx+cw-8 {
				sigX = x + 56
				y += 16
			}
			drawBadge(screen, sigX, y, sig, colour(0x2E, 0x45, 0x38), face)
			sigX += sw + 6
		}
		y += 20
	}

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
	_ = pendingActions // lobby click handled in handleParliamentClick
	_ = y
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

// partyColour returns the display colour for a party.
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
