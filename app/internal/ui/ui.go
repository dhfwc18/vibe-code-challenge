// Package ui provides the front-end rendering and input handling for the 20-50 game.
// All visual output is produced via direct ebiten.Image drawing; ebitenui is used for
// button widget management.
package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/player"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// contentX is the left edge of the main content area (after the sidebar).
const contentX = tabBarWidth

// UI owns the top-level UI state. It is created once before the game loop starts.
type UI struct {
	face font.Face
	hud  *HUD
	tabs *TabBar

	// Tab-specific state.
	mapState      mapTabState
	industryState industryTabState
	evidenceState evidenceTabState

	// Pending actions queued this frame to return from Update.
	pendingActions []simulation.Action

	// advanceWeekRequested is set when the player clicks "Advance Week".
	advanceWeekRequested bool
}

// New creates and returns a fully initialised UI.
func New(world *simulation.WorldState, cfg *config.Config) *UI {
	u := &UI{
		face:           basicfont.Face7x13,
		hud:            newHUD(),
		tabs:           newTabBar(),
		pendingActions: []simulation.Action{},
	}
	_ = world
	_ = cfg
	return u
}

// AdvanceWeekRequested returns true if the player signalled Advance Week this frame.
// Calling this resets the flag.
func (u *UI) AdvanceWeekRequested() bool {
	if u.advanceWeekRequested {
		u.advanceWeekRequested = false
		return true
	}
	return false
}

// NotifyEvent passes the most recent event name to the HUD notification strip.
func (u *UI) NotifyEvent(name string) {
	u.hud.setLastEvent(name)
}

// Update processes input for this frame and returns any actions queued.
// The returned slice is valid only until the next call to Update.
func (u *UI) Update(world *simulation.WorldState) []simulation.Action {
	u.pendingActions = u.pendingActions[:0]

	// Handle Ticky modal first - it blocks all other input.
	if world.PendingTickyPressure {
		u.handleTickyInput(world)
		return u.pendingActions
	}

	// Handle shock modal.
	if len(world.PendingShockResponses) > 0 {
		u.handleShockInput(world)
		return u.pendingActions
	}

	// Handle evidence modal.
	if u.evidenceState.showModal {
		u.handleEvidenceModalInput(world)
		return u.pendingActions
	}

	// Advance Week button click detection.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		u.handleHUDClick(world)
	}

	// Tab bar.
	u.tabs.Update()

	// Tab-specific input.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		u.handleTabContentClick(world)
	}

	result := make([]simulation.Action, len(u.pendingActions))
	copy(result, u.pendingActions)
	return result
}

// handleHUDClick detects clicks on the HUD Advance Week button.
func (u *UI) handleHUDClick(world *simulation.WorldState) {
	sw, _ := ebiten.WindowSize()
	if sw == 0 {
		sw = 1280
	}
	mx, my := ebiten.CursorPosition()
	btnX := sw - 160
	btnY := 6
	if mx >= btnX && mx <= btnX+148 && my >= btnY && my <= btnY+28 {
		u.advanceWeekRequested = true
		u.hud.setLastEvent("") // clear notification on new week
	}
}

// handleTickyInput detects which Ticky response button was clicked.
func (u *UI) handleTickyInput(world *simulation.WorldState) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	sw, sh := ebiten.WindowSize()
	if sw == 0 {
		sw = 1280
	}
	if sh == 0 {
		sh = 720
	}
	b := computeTickyBounds(sw, sh)
	mx, my := ebiten.CursorPosition()

	// Accept button: x+20, y+90, w=130, h=28
	if inRect(mx, my, b.mx+20, b.my+90, 130, 28) {
		u.pendingActions = append(u.pendingActions, simulation.Action{
			Type:   player.ActionTypeRespondTickyPressure,
			Target: "ticky_tennison",
			Detail: "ACCEPT",
		})
		return
	}
	// Decline button: x+170, y+90, w=130, h=28
	if inRect(mx, my, b.mx+170, b.my+90, 130, 28) {
		u.pendingActions = append(u.pendingActions, simulation.Action{
			Type:   player.ActionTypeRespondTickyPressure,
			Target: "ticky_tennison",
			Detail: "DECLINE",
		})
		return
	}
	// Negotiate button: x+320, y+90, w=130, h=28
	if inRect(mx, my, b.mx+320, b.my+90, 130, 28) {
		u.pendingActions = append(u.pendingActions, simulation.Action{
			Type:   player.ActionTypeRespondTickyPressure,
			Target: "ticky_tennison",
			Detail: "NEGOTIATE",
		})
	}
}

// handleShockInput detects shock response button clicks.
func (u *UI) handleShockInput(world *simulation.WorldState) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	if len(world.PendingShockResponses) == 0 {
		return
	}
	sw, sh := ebiten.WindowSize()
	if sw == 0 {
		sw = 1280
	}
	if sh == 0 {
		sh = 720
	}
	mw, mh := 440, 180
	bx := (sw - mw) / 2
	by := (sh - mh) / 2
	mx, my := ebiten.CursorPosition()
	shock := world.PendingShockResponses[0]

	if inRect(mx, my, bx+20, by+80, 120, 28) {
		u.pendingActions = append(u.pendingActions, simulation.Action{
			Type:   player.ActionTypeShockResponse,
			Target: shock.EventDefID,
			Detail: "ACCEPT",
		})
	} else if inRect(mx, my, bx+160, by+80, 120, 28) {
		u.pendingActions = append(u.pendingActions, simulation.Action{
			Type:   player.ActionTypeShockResponse,
			Target: shock.EventDefID,
			Detail: "DECLINE",
		})
	}
}

// handleEvidenceModalInput detects confirm/cancel clicks on the commission modal.
func (u *UI) handleEvidenceModalInput(world *simulation.WorldState) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	sw, sh := ebiten.WindowSize()
	if sw == 0 {
		sw = 1280
	}
	if sh == 0 {
		sh = 720
	}
	mw, mh := 400, 200
	mx2 := (sw - mw) / 2
	my2 := (sh - mh) / 2
	x, y := ebiten.CursorPosition()

	// Confirm button: mx+50, my+90, w=100, h=24
	if inRect(x, y, mx2+50, my2+90, 100, 24) {
		queueCommissionAction(&u.pendingActions, u.evidenceState.selectedOrgID, u.evidenceState.selectedInsight)
		u.evidenceState.showModal = false
	}
	// Cancel button: mx+200, my+90, w=100, h=24
	if inRect(x, y, mx2+200, my2+90, 100, 24) {
		u.evidenceState.showModal = false
	}
}

// handleTabContentClick handles clicks within the tab content area.
func (u *UI) handleTabContentClick(world *simulation.WorldState) {
	mx, my := ebiten.CursorPosition()
	if mx < contentX {
		return // not in content area
	}

	switch u.tabs.ActiveTab() {
	case 2: // Politics
		u.handlePoliticsClick(world, mx, my)
	case 3: // Policy
		u.handlePolicyClick(world, mx, my)
	case 6: // Evidence
		u.handleEvidenceClick(world, mx, my)
	}
}

// handlePoliticsClick detects Lobby button clicks in the politics tab.
func (u *UI) handlePoliticsClick(world *simulation.WorldState, mx, my int) {
	parties := []config.Party{
		config.PartyLeft,
		config.PartyRight,
		config.PartyFarLeft,
		config.PartyFarRight,
	}
	cy := hudHeight
	for colIdx, party := range parties {
		colX := contentX + colIdx*politicsColW
		rowY := cy + 26
		for i := range world.Stakeholders {
			s := world.Stakeholders[i]
			if s.Party != party || !s.IsUnlocked {
				continue
			}
			btnX := colX + 4 + 145
			btnY := rowY + 14
			if inRect(mx, my, btnX, btnY, 50, 18) {
				if isInCabinet(s, *world) && world.Player.APRemaining >= 3 {
					u.pendingActions = append(u.pendingActions, simulation.Action{
						Type:   player.ActionTypeLobbyMinister,
						Target: s.ID,
					})
				}
				return
			}
			rowY += 72
		}
	}
}

// handlePolicyClick detects Submit button clicks in the policy tab.
func (u *UI) handlePolicyClick(world *simulation.WorldState, mx, my int) {
	cy := hudHeight
	for colIdx, state := range policyColumns {
		colX := contentX + colIdx*policyColW
		rowY := cy + 24
		for i := range world.PolicyCards {
			pc := world.PolicyCards[i]
			if pc.State != state {
				continue
			}
			if state == "DRAFT" {
				btnX := colX + 4 + 90
				btnY := rowY + 32
				if inRect(mx, my, btnX, btnY, 50, 16) {
					if pc.Def != nil && world.Player.APRemaining >= pc.Def.APCost {
						u.pendingActions = append(u.pendingActions, simulation.Action{
							Type:   player.ActionTypeSubmitPolicy,
							Target: pc.Def.ID,
						})
					}
					return
				}
			}
			rowY += 64
		}
	}
}

// handleEvidenceClick detects Commission button clicks in the evidence tab.
func (u *UI) handleEvidenceClick(world *simulation.WorldState, mx, my int) {
	if world.Cfg == nil {
		return
	}
	cy := hudHeight
	x := contentX + 12
	y := cy + 16 + 18 + 14 // match drawTabEvidence layout
	for _, orgDef := range world.Cfg.Organisations {
		orgState := findOrgState(world.OrgStates, orgDef.ID)
		coolingOff := orgState.CoolingOffUntil > world.Week
		isMurican := orgDef.Origin == config.OrgMurican
		locked := isMurican && !orgState.MuricanUnlocked
		canCommission := !coolingOff && !locked

		btnX := x + 520
		btnY := y - 12
		if canCommission && inRect(mx, my, btnX, btnY, 70, 16) {
			u.evidenceState.selectedOrgID = orgDef.ID
			if len(orgDef.Specialisms) > 0 {
				u.evidenceState.selectedInsight = orgDef.Specialisms[0]
			}
			u.evidenceState.showModal = true
			return
		}
		y += 16
	}
}

// Draw renders the entire UI onto screen.
func (u *UI) Draw(screen *ebiten.Image, world simulation.WorldState) {
	// HUD top bar.
	u.hud.Draw(screen, world, u.face)

	// Tab sidebar.
	u.tabs.Draw(screen, u.face)

	// Content area dimensions.
	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()
	cx := contentX
	cy := hudHeight
	cw := sw - contentX
	ch := sh - hudHeight

	// Dispatch to active tab renderer.
	switch u.tabs.ActiveTab() {
	case 0:
		drawTabOverview(screen, world, u.face, cx, cy, cw, ch)
	case 1:
		drawTabMap(screen, world, &u.mapState, u.face, cx, cy, cw, ch)
	case 2:
		drawTabPolitics(screen, world, &u.pendingActions, u.face, cx, cy, cw, ch)
	case 3:
		drawTabPolicy(screen, world, &u.pendingActions, u.face, cx, cy, cw, ch)
	case 4:
		drawTabEnergy(screen, world, u.face, cx, cy, cw, ch)
	case 5:
		drawTabIndustry(screen, world, &u.industryState, u.face, cx, cy, cw, ch)
	case 6:
		drawTabEvidence(screen, world, &u.evidenceState, &u.pendingActions, u.face, cx, cy, cw, ch)
	case 7:
		drawTabBudget(screen, world, u.face, cx, cy, cw, ch)
	}

	// Modals (drawn last so they appear on top).
	if world.PendingTickyPressure {
		drawModalTicky(screen, world, &u.pendingActions, u.face)
	} else if len(world.PendingShockResponses) > 0 {
		drawModalShock(screen, world, &u.pendingActions, u.face)
	}
}

// inRect returns true if (px, py) is inside the rectangle [x, x+w) x [y, y+h).
func inRect(px, py, x, y, w, h int) bool {
	return px >= x && px < x+w && py >= y && py < y+h
}
