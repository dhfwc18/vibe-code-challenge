package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/player"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
)

// drawModalShock renders the shock response modal when there is a pending shock.
// It shows the first pending shock in the list.
func drawModalShock(
	screen *ebiten.Image,
	world simulation.WorldState,
	pendingActions *[]simulation.Action,
	face font.Face,
) bool {
	if len(world.PendingShockResponses) == 0 {
		return false
	}

	// Look up the event name.
	shock := world.PendingShockResponses[0]
	eventName := shock.EventDefID
	for _, def := range world.Cfg.Events {
		if def.ID == shock.EventDefID {
			eventName = def.Name
			break
		}
	}

	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()

	// Full-screen dark overlay.
	solidRect(screen, 0, 0, sw, sh, ColourOverlay)

	// Modal box (wide enough for 3 buttons).
	mw, mh := 460, 180
	mx := (sw - mw) / 2
	my := (sh - mh) / 2
	drawPanel(screen, mx, my, mw, mh)

	// Title.
	drawLabel(screen, mx+12, my+20, "Shock Response", ColourClimateHigh, face)
	drawLabel(screen, mx+12, my+40, "Event: "+eventName, ColourTextPrimary, face)
	drawLabel(screen, mx+12, my+56, "Choose your response:", ColourTextMuted, face)

	// Accept button.
	solidRect(screen, mx+20, my+80, 120, 28, ColourButtonNormal)
	drawLabel(screen, mx+30, my+98, "Accept", ColourAccent, face)

	// Decline button.
	solidRect(screen, mx+160, my+80, 120, 28, ColourButtonNormal)
	drawLabel(screen, mx+170, my+98, "Decline", ColourClimateCritical, face)

	// Mitigate button.
	solidRect(screen, mx+300, my+80, 120, 28, ColourButtonNormal)
	drawLabel(screen, mx+306, my+98, "Mitigate", ColourClimateMedium, face)

	// Click detection handled in Update.
	_ = pendingActions
	_ = player.ActionTypeShockResponse
	return false
}
