package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/policy"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
)

// policyColumns defines the display order of policy states.
var policyColumns = []policy.PolicyState{
	policy.PolicyStateDraft,
	policy.PolicyStateUnderReview,
	policy.PolicyStateApproved,
	policy.PolicyStateActive,
	policy.PolicyStateArchived,
}

const policyColW = 216
const policyCardH = 72

// drawTabPolicy renders the policy pipeline tab.
func drawTabPolicy(
	screen *ebiten.Image,
	world simulation.WorldState,
	pendingActions *[]simulation.Action,
	face font.Face,
	cx, cy, cw, ch int,
) {
	drawPanel(screen, cx, cy, cw, ch)

	for colIdx, state := range policyColumns {
		colX := cx + 8 + colIdx*(policyColW+8)
		if colX+policyColW > cx+cw {
			break
		}

		// Column header.
		solidRect(screen, colX, cy, policyColW-4, 18, colour(0x2E, 0x45, 0x38))
		drawLabel(screen, colX+6, cy+13, string(state), ColourAccent, face)

		rowY := cy + 24
		for i := range world.PolicyCards {
			pc := world.PolicyCards[i]
			if pc.State != state {
				continue
			}
			if rowY+policyCardH > cy+ch {
				break
			}
			drawPolicyCard(screen, pc, world, pendingActions, face, colX+4, rowY)
			rowY += policyCardH
		}
	}
}

// drawPolicyCard draws one policy card at x, y.
func drawPolicyCard(
	screen *ebiten.Image,
	pc policy.PolicyCard,
	world simulation.WorldState,
	pendingActions *[]simulation.Action,
	face font.Face,
	x, y int,
) {
	solidRect(screen, x, y, policyColW-8, policyCardH-2, color.RGBA{R: 0x18, G: 0x28, B: 0x1E, A: 0xFF})

	if pc.Def == nil {
		return
	}

	// Name on line y+14.
	drawLabel(screen, x+4, y+14, pc.Def.Name, ColourTextPrimary, face)

	// Sector badge at y+20.
	drawBadge(screen, x+4, y+20, string(pc.Def.Sector), sectorColour(pc.Def.Sector), face)

	// Significance badge at y+20.
	drawBadge(screen, x+90, y+20, string(pc.Def.Significance), significanceColour(pc.Def.Significance), face)

	// AP cost at y+38.
	drawLabel(screen, x+4, y+38, fmt.Sprintf("AP: %d", pc.Def.APCost), ColourTextMuted, face)

	// Submit button at y+50 (DRAFT only).
	if pc.State == policy.PolicyStateDraft {
		canSubmit := world.Player.APRemaining >= pc.Def.APCost
		btnCol := buttonColour(x+90, y+50, 50, 16, canSubmit)
		lblCol := ColourTextPrimary
		if !canSubmit {
			lblCol = ColourTextMuted
		}
		solidRect(screen, x+90, y+50, 50, 16, btnCol)
		drawLabel(screen, x+94, y+62, "Submit", lblCol, face)
		_ = pendingActions // click detection in Update
	}
}

// sectorColour returns a colour for a policy sector badge.
func sectorColour(s config.PolicySector) color.RGBA {
	switch s {
	case config.PolicySectorPower:
		return colour(0xF3, 0x9C, 0x12)
	case config.PolicySectorTransport:
		return colour(0x27, 0xAE, 0x60)
	case config.PolicySectorBuildings:
		return colour(0x29, 0x80, 0xB9)
	case config.PolicySectorIndustry:
		return colour(0x8E, 0x44, 0xAD)
	case config.PolicySectorCross:
		return colour(0x16, 0xA0, 0x85)
	default:
		return ColourPartyNeutral
	}
}

// significanceColour returns a colour for a policy significance badge.
func significanceColour(s config.PolicySignificance) color.RGBA {
	switch s {
	case config.PolicySignificanceMinor:
		return ColourPartyNeutral
	case config.PolicySignificanceModerate:
		return ColourOrgThinkTank
	case config.PolicySignificanceMajor:
		return colour(0xD4, 0xAC, 0x0D)
	default:
		return ColourPartyNeutral
	}
}
