package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/vibe-code-challenge/twenty-fifty/internal/player"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"golang.org/x/image/font"
)

// drawModalTicky renders the Ticky Pressure modal when PendingTickyPressure is true.
// Returns true if the user has clicked a response (modal should be dismissed).
func drawModalTicky(
	screen *ebiten.Image,
	world simulation.WorldState,
	pendingActions *[]simulation.Action,
	face font.Face,
) bool {
	if !world.PendingTickyPressure {
		return false
	}

	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()

	// Full-screen dark overlay.
	solidRect(screen, 0, 0, sw, sh, ColourOverlay)

	// Modal box.
	mw, mh := 480, 220
	mx := (sw - mw) / 2
	my := (sh - mh) / 2
	drawPanel(screen, mx, my, mw, mh)

	// Title.
	drawLabel(screen, mx+12, my+20, "Ticky Pressure", ColourClimateCritical, face)

	// Description.
	drawLabel(screen, mx+12, my+40,
		"TD Tennison is applying pressure to strengthen Murican ties.", ColourTextPrimary, face)
	drawLabel(screen, mx+12, my+56,
		"Accepting unlocks the Murican Growth Alliance.", ColourTextMuted, face)

	// Accept button.
	solidRect(screen, mx+20, my+90, 130, 28, ColourButtonNormal)
	drawLabel(screen, mx+24, my+108, "Accept (+8 rel)", ColourAccent, face)

	// Decline button.
	solidRect(screen, mx+170, my+90, 130, 28, ColourButtonNormal)
	drawLabel(screen, mx+174, my+108, "Decline (-5 rel)", ColourClimateCritical, face)

	// Negotiate button.
	solidRect(screen, mx+320, my+90, 130, 28, ColourButtonNormal)
	drawLabel(screen, mx+324, my+108, "Negotiate (-2 rel)", ColourClimateMedium, face)

	// Click detection is handled in Update; here we just record references.
	_ = pendingActions
	_ = player.ActionTypeRespondTickyPressure
	return false
}

// tickyButtonBounds returns the screen rectangles for each Ticky response button.
// Used by Update to detect clicks.
type tickyButtonBounds struct {
	mx, my, mw, mh int
}

// computeTickyBounds returns the modal origin given screen dimensions.
func computeTickyBounds(sw, sh int) tickyButtonBounds {
	mw, mh := 480, 220
	return tickyButtonBounds{
		mx: (sw - mw) / 2,
		my: (sh - mh) / 2,
		mw: mw,
		mh: mh,
	}
}
