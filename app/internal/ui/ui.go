// Package ui provides the front-end rendering and input handling for the 20-50 game.
// All visual output is produced via direct ebiten.Image drawing; ebitenui is used for
// button widget management.
package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/event"
	"github.com/vibe-code-challenge/twenty-fifty/internal/player"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// logicalW and logicalH are the fixed logical screen dimensions returned by game.Layout.
// All click-detection code MUST use these -- never ebiten.WindowSize(), which returns the
// physical window size and breaks on DPI-scaled displays.
const (
	logicalW = 1280
	logicalH = 720
)

// contentX is the left edge of the main content area (sidebar is gone; kept for reference).
const contentX = tabBarWidth

// panelOverlayW is the pixel width of the right-side overlay panel for non-map tabs.
const panelOverlayW = 720

// UI owns the top-level UI state. It is created once before the game loop starts.
type UI struct {
	face font.Face
	hud  *HUD
	tabs *TabBar

	// Tab-specific state.
	mapState        mapTabState
	parliamentState parliamentTabState
	industryState   industryTabState
	evidenceState   evidenceTabState

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

	// newspaper holds pending event news items to show after a week advance.
	newspaper NewspaperQueue
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

// QueueNewsItems enqueues fired events for display in the newspaper modal.
func (u *UI) QueueNewsItems(entries []event.EventEntry) {
	u.newspaper.Enqueue(entries)
}

// Update processes input for this frame and returns any actions queued.
// The returned slice is valid only until the next call to Update.
func (u *UI) Update(world *simulation.WorldState) []simulation.Action {
	if u.feedbackFrames > 0 {
		u.feedbackFrames--
	}

	u.pendingActions = u.pendingActions[:0]

	// Use fixed logical dimensions -- never ebiten.WindowSize() which varies with DPI.
	sw, sh := logicalW, logicalH

	// Newspaper modal blocks all other input until dismissed.
	if u.newspaper.HasPending() {
		if handleNewspaperInput(sw, sh) {
			u.newspaper.Dismiss()
		}
		return u.pendingActions
	}

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
	btnX := sw - hudBtnW - 8
	btnY := (hudHeight - hudBtnH) / 2
	if inRect(mx, my, btnX, btnY, hudBtnW, hudBtnH) {
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
	sw, sh := logicalW, logicalH
	mx, my := ebiten.CursorPosition()

	activeTab := u.tabs.ActiveTab()
	ovX := sw - panelOverlayW

	// When the map tab is active, route parliament panel clicks and map polygon clicks.
	if activeTab == 1 || mx < ovX {
		u.handleMapClick(world, mx, my, sw, sh)
		return
	}

	switch activeTab {
	case 2: // Policy
		u.handlePolicyClick(world, mx, my, ovX)
	case 4: // Industry
		u.handleIndustryClick(mx, my, ovX)
	case 5: // Evidence
		u.handleEvidenceClick(world, mx, my, ovX)
	}
}

// handleIndustryClick detects filter button clicks in the industry tab.
func (u *UI) handleIndustryClick(mx, my, panelX int) {
	// Filter buttons drawn at: x = panelX+12 + i*80, y = hudHeight+4, w=78, h=16
	y := hudHeight + 4
	for i := 0; i < 4; i++ {
		btnX := panelX + 12 + i*80
		if inRect(mx, my, btnX, y, 78, 16) {
			u.industryState.filter = industryFilter(i)
			return
		}
	}
}

// handlePolicyClick detects Submit button clicks in the policy tab.
func (u *UI) handlePolicyClick(world *simulation.WorldState, mx, my, panelX int) {
	cy := hudHeight
	for colIdx, state := range policyColumns {
		colX := panelX + 8 + colIdx*(policyColW+8)
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
func (u *UI) handleEvidenceClick(world *simulation.WorldState, mx, my, panelX int) {
	if world.Cfg == nil {
		return
	}
	cy := hudHeight
	x := panelX + 12
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
	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()
	cy := hudHeight
	ch := sh - hudHeight - panelBarH

	effAP := u.effectiveAP(&world)

	// Map is always rendered at full width as the background layer (includes parliament panel).
	drawTabMap(screen, world, &u.mapState, &u.parliamentState, &u.pendingActions, u.face, 0, cy, sw, ch, effAP)

	// Non-map tabs render as a right-side overlay panel on top of the map.
	activeTab := u.tabs.ActiveTab()
	if activeTab != 1 {
		ovX := sw - panelOverlayW
		switch activeTab {
		case 0:
			drawTabOverview(screen, world, u.face, ovX, cy, panelOverlayW, ch)
		case 2:
			drawTabPolicy(screen, world, &u.pendingActions, u.face, ovX, cy, panelOverlayW, ch, effAP)
		case 3:
			drawTabEnergy(screen, world, u.face, ovX, cy, panelOverlayW, ch)
		case 4:
			drawTabIndustry(screen, world, &u.industryState, u.face, ovX, cy, panelOverlayW, ch)
		case 5:
			drawTabEvidence(screen, world, &u.evidenceState, &u.pendingActions, u.face, ovX, cy, panelOverlayW, ch)
		case 6:
			drawTabBudget(screen, world, u.face, ovX, cy, panelOverlayW, ch)
		}
	}

	// HUD top bar (drawn after panels so it overlaps any content that bleeds upward).
	u.hud.Draw(screen, world, u.face, effAP, u.feedbackMsg)

	// Bottom tab bar.
	u.tabs.Draw(screen, u.face)

	// Modals (top of draw stack).
	if u.newspaper.HasPending() {
		drawNewspaperModal(screen, u.newspaper.Current(), len(u.newspaper.items), u.face)
	} else if world.PendingTickyPressure {
		drawModalTicky(screen, world, &u.pendingActions, u.face)
	} else if len(world.PendingShockResponses) > u.shockHandledCount {
		drawModalShock(screen, world, &u.pendingActions, u.face)
	}
}

// handleMapClick handles clicks on the map tab: parliament right panel, overlay buttons, and region polygons.
func (u *UI) handleMapClick(world *simulation.WorldState, mx, my, sw, sh int) {
	// Route right-panel clicks to parliament handler.
	rightPanelX := sw - mapDetailPanelW
	if mx >= rightPanelX {
		u.handleParliamentClick(world, mx, my)
		return
	}

	cy := hudHeight
	ch := sh - hudHeight - panelBarH

	// Overlay selector buttons.
	for i := 0; i < 3; i++ {
		bx := 8 + i*110
		by := cy + 8
		if inRect(mx, my, bx, by, 108, 20) {
			u.mapState.overlay = mapOverlay(i)
			return
		}
	}

	// Map polygon area (mirrors drawTabMap bounds).
	polyW := sw - mapDetailPanelW - 16
	if polyW < 64 {
		polyW = 64
	}
	pmx := 8
	pmy := cy + 36
	pmw := polyW
	pmh := ch - 44
	if maxH := pmw * 14 / 10; pmh > maxH {
		pmh = maxH
	}
	if !inRect(mx, my, pmx, pmy, pmw, pmh) {
		return
	}
	nx := float32(mx-pmx) / float32(pmw)
	ny := float32(my-pmy) / float32(pmh)
	if id := hitTestMap(nx, ny); id != "" {
		u.mapState.selectedRegion = id
	}
}

// handleParliamentClick routes clicks in the parliament right panel.
func (u *UI) handleParliamentClick(world *simulation.WorldState, mx, my int) {
	px := logicalW - mapDetailPanelW
	py := hudHeight
	pw := mapDetailPanelW
	ph := logicalH - hudHeight - panelBarH

	state := &u.parliamentState

	if state.selectedID != "" {
		// Profile view: Back link (px+16, py+16, 140x28) and Lobby button.
		if inRect(mx, my, px+16, py+16, 140, 28) {
			state.selectedID = ""
			return
		}
		btnX := px + 16
		btnY := py + ph - 44
		if inRect(mx, my, btnX, btnY, 120, 28) {
			for i := range world.Stakeholders {
				s := world.Stakeholders[i]
				if s.ID == state.selectedID {
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
			}
		}
		return
	}

	if state.selectedParty != "" {
		// Party member grid: Back link (px+12, py+12, 140x16) and member cards.
		x := px + 12
		if inRect(mx, my, x, py+12, 140, 16) {
			state.selectedParty = ""
			return
		}
		// Cards start at py+12+28+20 = py+60.
		y := py + 60
		cols := 2
		cardW := (pw - 24 - parliamentCardGap) / cols
		col := 0
		row := 0
		for i := range world.Stakeholders {
			s := world.Stakeholders[i]
			if s.Party != state.selectedParty {
				continue
			}
			cardX := x + col*(cardW+parliamentCardGap)
			cardY := y + row*(parliamentCardH+parliamentCardGap)
			if cardY+parliamentCardH > py+ph-4 {
				break
			}
			if inRect(mx, my, cardX, cardY, cardW, parliamentCardH) {
				state.selectedID = s.ID
				return
			}
			col++
			if col >= cols {
				col = 0
				row++
			}
		}
		return
	}

	// Overview: party table rows start below the region section and parliament header/hemicycle.
	// Layout: py + regionSectionH + 2 (divider) + 12 + 20 (title) + hemicycleH + 8 + 18 (header)
	x := px + 12
	y := py + regionSectionH + 2 + 12 + 20 + hemicycleH + 8 + 18
	rowH := 28
	for _, p := range hemicyclePartyOrder {
		if inRect(mx, my, x, y, pw-24, rowH) {
			state.selectedParty = p
			return
		}
		y += rowH + 2
	}
}

// inRect returns true if (px, py) is inside the rectangle [x, x+w) x [y, y+h).
func inRect(px, py, x, y, w, h int) bool {
	return px >= x && px < x+w && py >= y && py < y+h
}
