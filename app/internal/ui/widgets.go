package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

// isHovered returns true if the mouse cursor is currently inside the given rectangle.
// Uses logical screen coordinates (matching ebiten.CursorPosition).
func isHovered(x, y, w, h int) bool {
	mx, my := ebiten.CursorPosition()
	return mx >= x && mx < x+w && my >= y && my < y+h
}

// buttonColour returns the appropriate button background colour based on hover state.
// canAct should be false when the button is disabled.
func buttonColour(x, y, w, h int, canAct bool) color.RGBA {
	if !canAct {
		return ColourButtonDisabled
	}
	if isHovered(x, y, w, h) {
		return ColourButtonHover
	}
	return ColourButtonNormal
}

// drawBar draws a horizontal filled progress bar.
// value and max define the fill fraction (clamped to [0, max]).
func drawBar(screen *ebiten.Image, x, y, w, h int, value, max float64, fill, bg color.RGBA) {
	if max <= 0 {
		max = 1
	}
	frac := value / max
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}

	// Background rectangle.
	bgImg := ebiten.NewImage(w, h)
	bgImg.Fill(bg)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(bgImg, op)

	// Fill rectangle.
	fillW := int(float64(w) * frac)
	if fillW > 0 {
		fillImg := ebiten.NewImage(fillW, h)
		fillImg.Fill(fill)
		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(fillImg, op2)
	}
}

// drawBadge draws a coloured rectangle with centred text.
func drawBadge(screen *ebiten.Image, x, y int, label string, bg color.RGBA, face font.Face) {
	const padX, padY = 6, 2
	w := len(label)*7 + padX*2
	h := 15

	img := ebiten.NewImage(w, h)
	img.Fill(bg)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)

	text.Draw(screen, label, face, x+padX, y+h-padY, ColourTextPrimary)
}

// drawLabel draws a text string at the given position.
func drawLabel(screen *ebiten.Image, x, y int, label string, col color.RGBA, face font.Face) {
	text.Draw(screen, label, face, x, y, col)
}

// drawPanel draws a Panel-coloured rectangle with a subtle border.
func drawPanel(screen *ebiten.Image, x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	panel := ebiten.NewImage(w, h)
	panel.Fill(ColourPanel)

	// Border: 1px darker outline.
	border := color.RGBA{R: 0x2E, G: 0x45, B: 0x38, A: 0xFF}
	// Top row.
	top := ebiten.NewImage(w, 1)
	top.Fill(border)
	panel.DrawImage(top, &ebiten.DrawImageOptions{})
	// Bottom row.
	bop := &ebiten.DrawImageOptions{}
	bop.GeoM.Translate(0, float64(h-1))
	panel.DrawImage(top, bop)
	// Left column.
	left := ebiten.NewImage(1, h)
	left.Fill(border)
	panel.DrawImage(left, &ebiten.DrawImageOptions{})
	// Right column.
	rop := &ebiten.DrawImageOptions{}
	rop.GeoM.Translate(float64(w-1), 0)
	panel.DrawImage(left, rop)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(panel, op)
}

// solidRect draws a filled rectangle of the given colour.
func solidRect(screen *ebiten.Image, x, y, w, h int, col color.RGBA) {
	if w <= 0 || h <= 0 {
		return
	}
	img := ebiten.NewImage(w, h)
	img.Fill(col)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}
