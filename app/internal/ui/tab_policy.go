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
	effectiveAP int,
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
			drawPolicyCard(screen, pc, world, pendingActions, face, colX+4, rowY, effectiveAP)
			rowY += policyCardH
		}
	}
}

// truncatePol truncates s to 24 chars for display in a policy card.
func truncatePol(s string) string {
	if len(s) > 24 {
		return s[:21] + "..."
	}
	return s
}

// drawPolicyCard draws one policy card at x, y.
func drawPolicyCard(
	screen *ebiten.Image,
	pc policy.PolicyCard,
	world simulation.WorldState,
	pendingActions *[]simulation.Action,
	face font.Face,
	x, y int,
	effectiveAP int,
) {
	solidRect(screen, x, y, policyColW-8, policyCardH-2, color.RGBA{R: 0x18, G: 0x28, B: 0x1E, A: 0xFF})

	if pc.Def == nil {
		return
	}

	// Row 1 (y+14): card name, truncated to 24 chars.
	drawLabel(screen, x+4, y+14, truncatePol(pc.Def.Name), ColourTextPrimary, face)

	// Row 2 (y+28): sector badge at x+4; significance badge at x+120.
	drawBadge(screen, x+4, y+28, string(pc.Def.Sector), sectorColour(pc.Def.Sector), face)
	drawBadge(screen, x+120, y+28, string(pc.Def.Significance), significanceColour(pc.Def.Significance), face)

	// Row 3 (y+46): AP cost label at x+4; submit button at x+140, y+38.
	drawLabel(screen, x+4, y+46, fmt.Sprintf("AP: %d", pc.Def.APCost), ColourTextMuted, face)

	// Submit button at y+38 height 16 (DRAFT only).
	if pc.State == policy.PolicyStateDraft {
		canSubmit := effectiveAP >= pc.Def.APCost
		btnCol := buttonColour(x+140, y+38, 50, 16, canSubmit)
		lblCol := ColourTextPrimary
		if !canSubmit {
			lblCol = ColourTextMuted
		}
		solidRect(screen, x+140, y+38, 50, 16, btnCol)
		drawLabel(screen, x+144, y+50, "Submit", lblCol, face)
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
