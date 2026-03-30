package ui

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	textv2 "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font"
)

// goXFaceCache caches GoXFace wrappers keyed by the underlying font.Face pointer.
// GoXFace wraps an x/image/font.Face for use with the text/v2 API.
var (
	goXFaceCache   = map[font.Face]*textv2.GoXFace{}
	goXFaceCacheMu sync.Mutex
)

// toGoXFace returns a *textv2.GoXFace wrapping the given font.Face.
// Results are cached so the same face always returns the same wrapper.
func toGoXFace(f font.Face) *textv2.GoXFace {
	goXFaceCacheMu.Lock()
	defer goXFaceCacheMu.Unlock()
	if gx, ok := goXFaceCache[f]; ok {
		return gx
	}
	gx := textv2.NewGoXFace(f)
	goXFaceCache[f] = gx
	return gx
}

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
// value and maxVal define the fill fraction (clamped to [0, maxVal]).
func drawBar(screen *ebiten.Image, x, y, w, h int, value, maxVal float64, fill, bg color.RGBA) {
	if maxVal <= 0 {
		maxVal = 1
	}
	frac := value / maxVal
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

	drawLabel(screen, x+padX, y+h-padY, label, ColourTextPrimary, face)
}

// drawLabel draws a text string at the given position.
func drawLabel(screen *ebiten.Image, x, y int, label string, col color.RGBA, face font.Face) {
	var opts textv2.DrawOptions
	opts.GeoM.Translate(float64(x), float64(y))
	opts.ColorScale.SetR(float32(col.R) / 255)
	opts.ColorScale.SetG(float32(col.G) / 255)
	opts.ColorScale.SetB(float32(col.B) / 255)
	opts.ColorScale.SetA(float32(col.A) / 255)
	textv2.Draw(screen, label, toGoXFace(face), &opts)
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
