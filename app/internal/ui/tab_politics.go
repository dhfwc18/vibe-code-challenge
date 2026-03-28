package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"github.com/vibe-code-challenge/twenty-fifty/internal/stakeholder"
	"golang.org/x/image/font"
)

const politicsColW = 272
const ministerCardH = 88

// drawTabPolitics renders the politics tab.
func drawTabPolitics(
	screen *ebiten.Image,
	world simulation.WorldState,
	pendingActions *[]simulation.Action,
	face font.Face,
	cx, cy, cw, ch int,
) {
	drawPanel(screen, cx, cy, cw, ch)

	parties := []config.Party{
		config.PartyLeft,
		config.PartyRight,
		config.PartyFarLeft,
		config.PartyFarRight,
	}

	for colIdx, party := range parties {
		colX := cx + 8 + colIdx*(politicsColW+8)
		if colX+politicsColW > cx+cw {
			break
		}
		headerCol := partyColour(party)

		// Column border/header.
		solidRect(screen, colX, cy, politicsColW-2, 20, headerCol)
		drawLabel(screen, colX+4, cy+15, config.PartyNames[party], ColourTextPrimary, face)

		// Governing party highlight border.
		if party == world.Government.RulingParty {
			solidRect(screen, colX, cy, politicsColW-2, 1, ColourAccent)
			solidRect(screen, colX, cy, 1, ch, ColourAccent)
			solidRect(screen, colX+politicsColW-3, cy, 1, ch, ColourAccent)
		}

		rowY := cy + 26
		for i := range world.Stakeholders {
			s := world.Stakeholders[i]
			if s.Party != party || !s.IsUnlocked {
				continue
			}
			if rowY+ministerCardH > cy+ch {
				break
			}
			drawMinisterRow(screen, s, world, pendingActions, face, colX+4, rowY)
			rowY += ministerCardH
		}
	}
}

// drawMinisterRow draws one minister card at x, y.
func drawMinisterRow(
	screen *ebiten.Image,
	s stakeholder.Stakeholder,
	world simulation.WorldState,
	pendingActions *[]simulation.Action,
	face font.Face,
	x, y int,
) {
	// Background card.
	solidRect(screen, x-2, y, politicsColW-6, ministerCardH-2, color.RGBA{R: 0x18, G: 0x28, B: 0x1E, A: 0xFF})

	// Name on line y+14.
	drawLabel(screen, x, y+14, s.Name, ColourTextPrimary, face)

	// Role badge at y+20.
	drawBadge(screen, x, y+20, string(s.Role), ColourOrgThinkTank, face)

	// State badge at y+20 offset right.
	stateCol := ministerStateColour(s.State)
	drawBadge(screen, x+110, y+20, string(s.State), stateCol, face)

	// Popularity bar at y+38.
	drawLabel(screen, x, y+38, "Pop", ColourTextMuted, face)
	drawBar(screen, x+30, y+30, 100, 8, s.Popularity, 100, ColourAccent, ColourButtonNormal)

	// Relationship bar at y+52.
	drawLabel(screen, x, y+52, "Rel", ColourTextMuted, face)
	drawBar(screen, x+30, y+44, 100, 8, s.RelationshipScore, 100, ColourOrgThinkTank, ColourButtonNormal)

	// Lobby button at y+62.
	inCabinet := isInCabinet(s, world)
	canLobby := inCabinet && world.Player.APRemaining >= 3
	btnCol := buttonColour(x+145, y+54, 50, 18, canLobby)
	lblCol := ColourTextPrimary
	if !canLobby {
		lblCol = ColourTextMuted
	}
	solidRect(screen, x+145, y+54, 50, 18, btnCol)
	drawLabel(screen, x+148, y+66, "Lobby", lblCol, face)
	_ = pendingActions // button click detection is handled in Update via HitTest
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

