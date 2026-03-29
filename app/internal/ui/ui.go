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

// logicalW and logicalH are the fixed logical screen dimensions returned by game.Layout.
// All click-detection code must use these -- never ebiten.WindowSize().
const (
	logicalW = 1280
	logicalH = 720
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

	// pendingAPSpend tracks AP committed to queued actions this week.
	pendingAPSpend int
	// feedbackMsg is a short message shown in the HUD notification strip.
	feedbackMsg string
	// feedbackFrames is the number of frames remaining to show feedbackMsg.
	feedbackFrames int

	// shockHandledCount tracks how many shock responses the player has queued
	// this week. The shock modal hides once this reaches len(PendingShockResponses).
	shockHandledCount int
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
		u.pendingAPSpend = 0
		u.feedbackMsg = ""
		u.shockHandledCount = 0
		return true
	}
	return false
}

// effectiveAP returns the player's remaining AP minus any AP committed to queued actions.
func (u *UI) effectiveAP(world *simulation.WorldState) int {
	eff := world.Player.APRemaining - u.pendingAPSpend
	if eff < 0 {
		eff = 0
	}
	return eff
}

// NotifyEvent passes the most recent event name to the HUD notification strip.
func (u *UI) NotifyEvent(name string) {
	u.hud.setLastEvent(name)
}

// Update processes input for this frame and returns any actions queued.
// The returned slice is valid only until the next call to Update.
func (u *UI) Update(world *simulation.WorldState) []simulation.Action {
	if u.feedbackFrames > 0 {
		u.feedbackFrames--
	}

	u.pendingActions = u.pendingActions[:0]

	// Current logical screen dimensions (equals window size because Layout passes
	// through the outside dimensions for native-resolution rendering).
	sw, sh := ebiten.WindowSize()

	// Handle Ticky modal first - it blocks all other input.
	if world.PendingTickyPressure {
		u.handleTickyInput(world, sw, sh)
		return u.pendingActions
	}

	// Handle shock modal until all pending shocks have been responded to.
	if len(world.PendingShockResponses) > u.shockHandledCount {
		u.handleShockInput(world, sw, sh)
		return u.pendingActions
	}

	// Handle evidence modal.
	if u.evidenceState.showModal {
		u.handleEvidenceModalInput(world, sw, sh)
		return u.pendingActions
	}

	// Advance Week button click detection.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		u.handleHUDClick(world, sw)
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
func (u *UI) handleHUDClick(world *simulation.WorldState, sw int) {
	mx, my := ebiten.CursorPosition()
	btnX := sw - 160
	btnY := 6
	if mx >= btnX && mx <= btnX+148 && my >= btnY && my <= btnY+28 {
		u.advanceWeekRequested = true
		u.hud.setLastEvent("") // clear notification on new week
	}
}

// handleTickyInput detects which Ticky response button was clicked.
func (u *UI) handleTickyInput(world *simulation.WorldState, sw, sh int) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
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
func (u *UI) handleShockInput(world *simulation.WorldState, sw, sh int) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	if len(world.PendingShockResponses) <= u.shockHandledCount {
		return
	}
	mw, mh := 460, 180
	bx := (sw - mw) / 2
	by := (sh - mh) / 2
	mx, my := ebiten.CursorPosition()
	shock := world.PendingShockResponses[u.shockHandledCount]

	if inRect(mx, my, bx+20, by+80, 120, 28) {
		u.pendingActions = append(u.pendingActions, simulation.Action{
			Type:   player.ActionTypeShockResponse,
			Target: shock.EventDefID,
			Detail: "ACCEPT",
		})
		u.shockHandledCount++
	} else if inRect(mx, my, bx+160, by+80, 120, 28) {
		u.pendingActions = append(u.pendingActions, simulation.Action{
			Type:   player.ActionTypeShockResponse,
			Target: shock.EventDefID,
			Detail: "DECLINE",
		})
		u.shockHandledCount++
	} else if inRect(mx, my, bx+300, by+80, 120, 28) {
		u.pendingActions = append(u.pendingActions, simulation.Action{
			Type:   player.ActionTypeShockResponse,
			Target: shock.EventDefID,
			Detail: "MITIGATE",
		})
		u.shockHandledCount++
	}
}

// handleEvidenceModalInput detects confirm/cancel clicks on the commission modal.
func (u *UI) handleEvidenceModalInput(world *simulation.WorldState, sw, sh int) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
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
	case 1: // Map
		u.handleMapClick(mx, my)
	case 2: // Politics
		u.handlePoliticsClick(world, mx, my)
	case 3: // Policy
		u.handlePolicyClick(world, mx, my)
	case 5: // Industry
		u.handleIndustryClick(mx, my)
	case 6: // Evidence
		u.handleEvidenceClick(world, mx, my)
	}
}

// handleIndustryClick detects filter button clicks in the industry tab.
func (u *UI) handleIndustryClick(mx, my int) {
	// Filter buttons drawn at: x = contentX+12 + i*80, y = hudHeight+4, w=78, h=16
	// (matches tab_industry.go: x=cx+12, y=cy+16, buttons at y-12=cy+4, w=btnW-2=78)
	y := hudHeight + 4
	for i := 0; i < 4; i++ {
		btnX := contentX + 12 + i*80
		if inRect(mx, my, btnX, y, 78, 16) {
			u.industryState.filter = industryFilter(i)
			return
		}
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
		colX := contentX + 8 + colIdx*(politicsColW+8)
		rowY := cy + 26
		for i := range world.Stakeholders {
			s := world.Stakeholders[i]
			if s.Party != party || !s.IsUnlocked {
				continue
			}
			btnX := colX + 4 + 145
			btnY := rowY + 54
			if inRect(mx, my, btnX, btnY, 50, 18) {
				if isInCabinet(s, *world) && u.effectiveAP(world) >= 3 {
					u.pendingActions = append(u.pendingActions, simulation.Action{
						Type:   player.ActionTypeLobbyMinister,
						Target: s.ID,
					})
					u.pendingAPSpend += 3
				} else {
					u.feedbackMsg = "Not enough AP or not in cabinet"
					u.feedbackFrames = 150
				}
				return
			}
			rowY += ministerCardH
		}
	}
}

// handlePolicyClick detects Submit button clicks in the policy tab.
func (u *UI) handlePolicyClick(world *simulation.WorldState, mx, my int) {
	cy := hudHeight
	for colIdx, state := range policyColumns {
		colX := contentX + 8 + colIdx*(policyColW+8)
		rowY := cy + 24
		for i := range world.PolicyCards {
			pc := world.PolicyCards[i]
			if pc.State != state {
				continue
			}
			if state == "DRAFT" {
				btnX := colX + 4 + 140
				btnY := rowY + 38
				if inRect(mx, my, btnX, btnY, 50, 16) {
					if pc.Def != nil && u.effectiveAP(world) >= pc.Def.APCost {
						u.pendingActions = append(u.pendingActions, simulation.Action{
							Type:   player.ActionTypeSubmitPolicy,
							Target: pc.Def.ID,
						})
						u.pendingAPSpend += pc.Def.APCost
					} else {
						u.feedbackMsg = "Not enough AP"
						u.feedbackFrames = 150
					}
					return
				}
			}
			rowY += policyCardH
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

		btnX := x + 556
		btnY := y - 12
		if canCommission && inRect(mx, my, btnX, btnY, 70, 16) {
			u.evidenceState.selectedOrgID = orgDef.ID
			if len(orgDef.Specialisms) > 0 {
				u.evidenceState.selectedInsight = orgDef.Specialisms[0]
			}
			u.evidenceState.showModal = true
			return
		}
		y += 24
	}
}

// Draw renders the entire UI onto screen.
func (u *UI) Draw(screen *ebiten.Image, world simulation.WorldState) {
	// HUD top bar.
	u.hud.Draw(screen, world, u.face, u.effectiveAP(&world), u.feedbackMsg)

	// Tab sidebar.
	u.tabs.Draw(screen, u.face)

	// Content area dimensions.
	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()
	cx := contentX
	cy := hudHeight
	cw := sw - contentX
	ch := sh - hudHeight

	effAP := u.effectiveAP(&world)

	// Dispatch to active tab renderer.
	switch u.tabs.ActiveTab() {
	case 0:
		drawTabOverview(screen, world, u.face, cx, cy, cw, ch)
	case 1:
		drawTabMap(screen, world, &u.mapState, u.face, cx, cy, cw, ch)
	case 2:
		drawTabPolitics(screen, world, &u.pendingActions, u.face, cx, cy, cw, ch, effAP)
	case 3:
		drawTabPolicy(screen, world, &u.pendingActions, u.face, cx, cy, cw, ch, effAP)
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
	} else if len(world.PendingShockResponses) > u.shockHandledCount {
		drawModalShock(screen, world, &u.pendingActions, u.face)
	}
}

// handleMapClick handles overlay button and region polygon clicks on the map tab.
func (u *UI) handleMapClick(mx, my int) {
	cy := hudHeight
	ch := logicalH - hudHeight

	// Overlay selector buttons.
	for i := 0; i < 3; i++ {
		bx := contentX + 8 + i*110
		by := cy + 8
		if inRect(mx, my, bx, by, 108, 20) {
			u.mapState.overlay = mapOverlay(i)
			return
		}
	}

	// Map polygon area.
	pmx := contentX + 8
	pmy := cy + 36
	pmw := mapPanelW - 16
	pmh := ch - 44
	if !inRect(mx, my, pmx, pmy, pmw, pmh) {
		return
	}
	nx := float32(mx-pmx) / float32(pmw)
	ny := float32(my-pmy) / float32(pmh)
	if id := hitTestMap(nx, ny); id != "" {
		u.mapState.selectedRegion = id
	}
}

// inRect returns true if (px, py) is inside the rectangle [x, x+w) x [y, y+h).
func inRect(px, py, x, y, w, h int) bool {
	return px >= x && px < x+w && py >= y && py < y+h
}
