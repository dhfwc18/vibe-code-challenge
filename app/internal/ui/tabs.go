package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
)

const (
	tabBarWidth  = 160
	tabHeight    = 40
	tabBarTop    = hudHeight
)

// tabNames lists the eight tabs in display order.
var tabNames = []string{
	"Overview",
	"Map",
	"Politics",
	"Policy",
	"Energy",
	"Industry",
	"Evidence",
	"Budget",
}

// TabBar renders and handles input for the left sidebar tab buttons.
type TabBar struct {
	activeTab int
}

// newTabBar creates a new TabBar with Overview selected.
func newTabBar() *TabBar {
	return &TabBar{}
}

// ActiveTab returns the index of the currently selected tab.
func (tb *TabBar) ActiveTab() int {
	return tb.activeTab
}

// Update checks for mouse clicks on the tab buttons and updates the active tab.
func (tb *TabBar) Update() {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	mx, my := ebiten.CursorPosition()
	if mx < 0 || mx >= tabBarWidth {
		return
	}
	for i := range tabNames {
		y := tabBarTop + i*tabHeight
		if my >= y && my < y+tabHeight {
			tb.activeTab = i
			return
		}
	}
}

// Draw renders the tab sidebar onto screen.
func (tb *TabBar) Draw(screen *ebiten.Image, face font.Face) {
	screenH := screen.Bounds().Dy()

	// Sidebar background.
	sidebar := ebiten.NewImage(tabBarWidth, screenH-tabBarTop)
	sidebar.Fill(ColourPanel)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, float64(tabBarTop))
	screen.DrawImage(sidebar, op)

	for i, name := range tabNames {
		y := tabBarTop + i*tabHeight
		var bg = ColourButtonNormal
		if i == tb.activeTab {
			bg = ColourButtonHover
		}
		solidRect(screen, 0, y, tabBarWidth, tabHeight, bg)
		// Bottom border between tabs.
		solidRect(screen, 0, y+tabHeight-1, tabBarWidth, 1, ColourBackground)
		drawLabel(screen, 12, y+26, name, ColourTextPrimary, face)
	}
}
