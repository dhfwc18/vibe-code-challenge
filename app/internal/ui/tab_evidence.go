package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/evidence"
	"github.com/vibe-code-challenge/twenty-fifty/internal/player"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
)

// evidenceTabState holds selection state and modal for the evidence tab.
type evidenceTabState struct {
	showModal       bool
	selectedOrgID   string
	selectedInsight config.InsightType
}

// truncate returns s truncated to at most n characters.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// drawTabEvidence renders the evidence tab.
func drawTabEvidence(
	screen *ebiten.Image,
	world simulation.WorldState,
	state *evidenceTabState,
	pendingActions *[]simulation.Action,
	face font.Face,
	cx, cy, cw, ch int,
) {
	drawPanel(screen, cx, cy, cw, ch)
	x := cx + 12
	y := cy + 16

	// 1. Org list.
	drawLabel(screen, x, y, "--- Advisory Organisations ---", ColourAccent, face)
	y += 18

	drawLabel(screen, x, y, "Name", ColourTextMuted, face)
	drawLabel(screen, x+224, y, "Type", ColourTextMuted, face)
	drawLabel(screen, x+320, y, "Origin", ColourTextMuted, face)
	drawLabel(screen, x+380, y, "Rel", ColourTextMuted, face)
	y += 14

	for _, orgDef := range world.Cfg.Organisations {
		orgState := findOrgState(world.OrgStates, orgDef.ID)
		coolingOff := orgState.CoolingOffUntil > world.Week
		isMurican := orgDef.Origin == config.OrgMurican
		locked := isMurican && !orgState.MuricanUnlocked

		nameCol := ColourTextPrimary
		if coolingOff || locked {
			nameCol = ColourTextMuted
		}

		drawLabel(screen, x, y, truncate(orgDef.Name, 28), nameCol, face)
		drawBadge(screen, x+224, y-12, string(orgDef.OrgType), orgTypeColour(orgDef.OrgType), face)
		drawLabel(screen, x+320, y, string(orgDef.Origin), ColourTextMuted, face)
		drawBar(screen, x+380, y-11, 80, 10, orgState.RelationshipScore, 100, ColourAccent, ColourButtonNormal)

		// Cooling-off indicator.
		if coolingOff {
			drawLabel(screen, x+468, y, fmt.Sprintf("cool %d", orgState.CoolingOffUntil-world.Week), ColourClimateCritical, face)
		}

		// Commission button.
		canCommission := !coolingOff && !locked
		btnCol := buttonColour(x+556, y-12, 70, 16, canCommission)
		lblCol := ColourTextPrimary
		if !canCommission {
			lblCol = ColourTextMuted
		}
		solidRect(screen, x+556, y-12, 70, 16, btnCol)
		drawLabel(screen, x+560, y, "Commission", lblCol, face)

		_ = pendingActions // click detection in Update
		y += 24
		if y > cy+ch/2 {
			break
		}
	}

	y += 8

	// 2. Active commissions.
	drawLabel(screen, x, y, "--- Active Commissions ---", ColourAccent, face)
	y += 18
	for _, c := range world.Commissions {
		if c.Delivered {
			continue
		}
		drawLabel(screen, x, y,
			fmt.Sprintf("[%s] %s  due wk %d", c.OrgID, string(c.InsightType), c.DeliveryWeek),
			ColourTextPrimary, face)
		y += 14
		if y > cy+ch*3/4 {
			break
		}
	}

	y += 8

	// 3. Report inbox (newest first).
	drawLabel(screen, x, y, "--- Report Inbox ---", ColourAccent, face)
	y += 18
	for i := len(world.Reports) - 1; i >= 0; i-- {
		r := world.Reports[i]
		drawLabel(screen, x, y,
			fmt.Sprintf("[Wk %d] %s | %s  val=%.1f", r.DeliveryWeek, r.OrgID, string(r.InsightType), r.ReportedValue),
			ColourTextPrimary, face)
		y += 14
		if y > cy+ch-10 {
			break
		}
	}

	// Commission modal.
	if state.showModal && state.selectedOrgID != "" {
		drawCommissionModal(screen, world, state, pendingActions, face)
	}
}

// drawCommissionModal draws the commission modal overlay.
func drawCommissionModal(
	screen *ebiten.Image,
	world simulation.WorldState,
	state *evidenceTabState,
	pendingActions *[]simulation.Action,
	face font.Face,
) {
	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()

	// Dark overlay.
	solidRect(screen, 0, 0, sw, sh, ColourOverlay)

	// Modal box.
	mw, mh := 400, 200
	mx := (sw - mw) / 2
	my := (sh - mh) / 2
	drawPanel(screen, mx, my, mw, mh)
	drawLabel(screen, mx+12, my+20, "Commission Report", ColourAccent, face)
	drawLabel(screen, mx+12, my+40, "Org: "+state.selectedOrgID, ColourTextPrimary, face)
	drawLabel(screen, mx+12, my+56, "Insight: "+string(state.selectedInsight), ColourTextPrimary, face)

	// Confirm button.
	solidRect(screen, mx+50, my+90, 100, 24, buttonColour(mx+50, my+90, 100, 24, true))
	drawLabel(screen, mx+70, my+106, "Confirm", ColourTextPrimary, face)

	// Cancel button.
	solidRect(screen, mx+200, my+90, 100, 24, buttonColour(mx+200, my+90, 100, 24, true))
	drawLabel(screen, mx+220, my+106, "Cancel", ColourTextPrimary, face)

	_ = pendingActions
}

// findOrgState returns the OrgState for the given org ID, or a default with RelationshipScore=50.
func findOrgState(states []evidence.OrgState, orgID string) evidence.OrgState {
	for _, s := range states {
		if s.OrgID == orgID {
			return s
		}
	}
	return evidence.OrgState{OrgID: orgID, RelationshipScore: 50}
}

// orgTypeColour returns the colour for an org type badge.
func orgTypeColour(t config.OrgType) color.RGBA {
	switch t {
	case config.OrgConsultancy:
		return ColourOrgConsultancy
	case config.OrgThinkTank:
		return ColourOrgThinkTank
	case config.OrgAcademic:
		return ColourOrgAcademic
	default:
		return ColourPartyNeutral
	}
}

// queueCommissionAction adds a commission action to the pending actions slice.
func queueCommissionAction(pendingActions *[]simulation.Action, orgID string, insight config.InsightType) {
	*pendingActions = append(*pendingActions, simulation.Action{
		Type:   player.ActionTypeCommissionReport,
		Target: orgID,
		Detail: string(insight),
	})
}
