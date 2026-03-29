package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/vibe-code-challenge/twenty-fifty/internal/event"
	"golang.org/x/image/font"
)

// NewsItem holds the display data for one fired event shown in the newspaper modal.
type NewsItem struct {
	Name      string // event display name
	Headline  string // short headline (from EventDef.Headline)
	Narrative string // full narrative text (from EventDef.Narrative)
	Week      int
}

// NewspaperQueue holds pending news items to display after a week advance.
// Items are shown one at a time; the player dismisses each to proceed.
type NewspaperQueue struct {
	items []NewsItem
}

// Enqueue adds news items from a slice of fired event entries.
// Only events that have a non-empty Headline or Narrative are queued.
func (nq *NewspaperQueue) Enqueue(entries []event.EventEntry) {
	for _, e := range entries {
		if e.Name == "" {
			continue
		}
		// Only show items with player-visible content.
		item := NewsItem{
			Name:      e.Name,
			Headline:  e.Name, // default to Name; enriched below if available
			Narrative: "",
			Week:      e.Week,
		}
		nq.items = append(nq.items, item)
	}
}

// EnqueueRich adds news items with headline and narrative text from the config.
// eventsMap maps DefID -> (headline, narrative).
func (nq *NewspaperQueue) EnqueueRich(entries []event.EventEntry, eventsMap map[string][2]string) {
	for _, e := range entries {
		if e.Name == "" {
			continue
		}
		item := NewsItem{
			Name:     e.Name,
			Headline: e.Name,
			Week:     e.Week,
		}
		if kv, ok := eventsMap[e.DefID]; ok {
			if kv[0] != "" {
				item.Headline = kv[0]
			}
			item.Narrative = kv[1]
		}
		nq.items = append(nq.items, item)
	}
}

// HasPending returns true if there are items waiting to be shown.
func (nq *NewspaperQueue) HasPending() bool {
	return len(nq.items) > 0
}

// Current returns the front item without removing it.
func (nq *NewspaperQueue) Current() NewsItem {
	if len(nq.items) == 0 {
		return NewsItem{}
	}
	return nq.items[0]
}

// Dismiss removes the front item.
func (nq *NewspaperQueue) Dismiss() {
	if len(nq.items) > 0 {
		nq.items = nq.items[1:]
	}
}

// Clear removes all pending items.
func (nq *NewspaperQueue) Clear() {
	nq.items = nq.items[:0]
}

// handleNewspaperInput checks for a dismiss click and returns true if dismissed.
func handleNewspaperInput(sw, sh int) bool {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return false
	}
	b := newspaperModalBounds(sw, sh)
	mx, my := ebiten.CursorPosition()
	// Dismiss button at bottom-right of modal.
	btnX := b.x + b.w - 130
	btnY := b.y + b.h - 38
	return inRect(mx, my, btnX, btnY, 120, 28)
}

type modalBounds struct {
	x, y, w, h int
}

func newspaperModalBounds(sw, sh int) modalBounds {
	mw, mh := 560, 340
	return modalBounds{
		x: (sw - mw) / 2,
		y: (sh - mh) / 2,
		w: mw,
		h: mh,
	}
}

// drawNewspaperModal renders the newspaper popup for the current news item.
func drawNewspaperModal(screen *ebiten.Image, item NewsItem, queueLen int, face font.Face) {
	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()
	b := newspaperModalBounds(sw, sh)

	// Dim overlay.
	dimImg := ebiten.NewImage(sw, sh)
	dimImg.Fill(colour(0x00, 0x00, 0x00))
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(0.55)
	screen.DrawImage(dimImg, op)

	// Modal card background.
	solidRect(screen, b.x, b.y, b.w, b.h, colour(0x0E, 0x1E, 0x14))
	// Top accent bar.
	solidRect(screen, b.x, b.y, b.w, 4, ColourAccent)
	// Border.
	solidRect(screen, b.x, b.y, b.w, 1, colour(0x2E, 0x55, 0x38))
	solidRect(screen, b.x, b.y+b.h-1, b.w, 1, colour(0x2E, 0x55, 0x38))
	solidRect(screen, b.x, b.y, 1, b.h, colour(0x2E, 0x55, 0x38))
	solidRect(screen, b.x+b.w-1, b.y, 1, b.h, colour(0x2E, 0x55, 0x38))

	x := b.x + 20
	y := b.y + 20

	// "THE TAITAN TIMES" masthead.
	masthead := "THE TAITAN TIMES"
	drawLabel(screen, x, y, masthead, ColourAccent, face)
	y += 18

	// Thin rule.
	solidRect(screen, x, y, b.w-40, 1, colour(0x2E, 0x55, 0x38))
	y += 10

	// Headline.
	headline := item.Headline
	if len(headline) > 70 {
		headline = headline[:70] + "..."
	}
	drawLabel(screen, x, y, headline, ColourTextPrimary, face)
	y += 22

	// Narrative body (wrapped).
	if item.Narrative != "" {
		text := item.Narrative
		charsPerLine := (b.w - 40) / 7
		if charsPerLine < 20 {
			charsPerLine = 20
		}
		for len(text) > 0 && y < b.y+b.h-60 {
			line := text
			if len(line) > charsPerLine {
				cut := charsPerLine
				for cut > 0 && text[cut] != ' ' {
					cut--
				}
				if cut == 0 {
					cut = charsPerLine
				}
				line = text[:cut]
				text = text[cut+1:]
			} else {
				text = ""
			}
			drawLabel(screen, x, y, line, ColourTextMuted, face)
			y += 14
		}
	}

	// Queue indicator.
	if queueLen > 1 {
		more := "+"
		for i := 1; i < queueLen && i < 4; i++ {
			more += " |"
		}
		drawLabel(screen, b.x+20, b.y+b.h-20, more, ColourTextMuted, face)
	}

	// Dismiss button.
	btnX := b.x + b.w - 130
	btnY := b.y + b.h - 38
	btnBg := buttonColour(btnX, btnY, 120, 28, true)
	solidRect(screen, btnX, btnY, 120, 28, btnBg)
	dismissLabel := "Continue"
	if queueLen > 1 {
		dismissLabel = "Next"
	}
	labelX := btnX + (120-len(dismissLabel)*7)/2
	drawLabel(screen, labelX, btnY+19, dismissLabel, ColourTextPrimary, face)
}
