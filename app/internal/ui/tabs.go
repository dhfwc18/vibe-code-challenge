package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
)

const (
	// tabBarWidth is kept as 0; the left sidebar has been removed.
	// Tabs now render as a horizontal panel selector bar at the bottom of the screen.
	tabBarWidth = 0
	// panelBarH is the height of the bottom panel selector bar.
	panelBarH = 40
	// tabBarTop is kept for backward compatibility with HUD height references.
	tabBarTop = hudHeight
)

// tabNames lists the panel buttons in display order.
// Index 1 ("Map") is the "map-only" state with no overlay panel visible.
var tabNames = []string{
	"Overview",  // 0
	"Map",       // 1 -- clicking this hides all overlays
	"Politics",  // 2
	"Policy",    // 3
	"Energy",    // 4
	"Industry",  // 5
	"Evidence",  // 6
	"Budget",    // 7
}

// TabBar renders and handles input for the bottom panel selector bar.
type TabBar struct {
	activeTab int // -1 = no panel (used internally); 1 = Map = default no-overlay state
}

// newTabBar creates a new TabBar defaulting to Map (no overlay).
func newTabBar() *TabBar {
	return &TabBar{activeTab: 1}
}

// ActiveTab returns the index of the currently selected tab.
func (tb *TabBar) ActiveTab() int {
	return tb.activeTab
}

// Update checks for mouse clicks on the bottom panel buttons and updates the active tab.
func (tb *TabBar) Update() {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	mx, my := ebiten.CursorPosition()
	barY := logicalH - panelBarH
	if my < barY || my >= logicalH {
		return
	}
	n := len(tabNames)
	if n == 0 {
		return
	}
	btnW := logicalW / n
	for i := range tabNames {
		bx := i * btnW
		if mx >= bx && mx < bx+btnW {
			if tb.activeTab == i && i != 1 {
				// Clicking the active non-map tab closes the overlay.
				tb.activeTab = 1
			} else {
				tb.activeTab = i
			}
			return
		}
	}
}

// Draw renders the bottom panel selector bar onto screen.
func (tb *TabBar) Draw(screen *ebiten.Image, face font.Face) {
	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()
	barY := sh - panelBarH
	n := len(tabNames)
	if n == 0 {
		return
	}
	btnW := sw / n

	// Bar background.
	solidRect(screen, 0, barY, sw, panelBarH, ColourPanel)
	// Top border line.
	solidRect(screen, 0, barY, sw, 1, colour(0x2E, 0x45, 0x38))

	for i, name := range tabNames {
		bx := i * btnW
		bw := btnW - 1 // 1px gap between buttons
		bg := ColourButtonNormal
		if i == tb.activeTab {
			bg = ColourButtonHover
		} else if isHovered(bx, barY+1, bw, panelBarH-1) {
			bg = colour(0x2C, 0x48, 0x2E)
		}
		solidRect(screen, bx, barY+1, bw, panelBarH-1, bg)
		// Centre-align text in button.
		textW := len(name) * 7
		labelX := bx + (bw-textW)/2
		drawLabel(screen, labelX, barY+26, name, ColourTextPrimary, face)
		// Active indicator: accent line at top of active button.
		if i == tb.activeTab {
			solidRect(screen, bx, barY+1, bw, 2, ColourAccent)
		}
	}
}
