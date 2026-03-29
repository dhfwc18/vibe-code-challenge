package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vibe-code-challenge/twenty-fifty/internal/event"
	"github.com/vibe-code-challenge/twenty-fifty/internal/ui"
)

func makeEntries(names ...string) []event.EventEntry {
	out := make([]event.EventEntry, len(names))
	for i, n := range names {
		out[i] = event.EventEntry{DefID: n, Name: n, Week: i + 1}
	}
	return out
}

// TestNewspaperQueue_Enqueue_PopulatesQueue verifies that Enqueue adds events with names.
func TestNewspaperQueue_Enqueue_PopulatesQueue(t *testing.T) {
	var nq ui.NewspaperQueue
	nq.Enqueue(makeEntries("event_a", "event_b"))
	assert.True(t, nq.HasPending())
}

// TestNewspaperQueue_Enqueue_SkipsBlankNames verifies that nameless entries are ignored.
func TestNewspaperQueue_Enqueue_SkipsBlankNames(t *testing.T) {
	var nq ui.NewspaperQueue
	nq.Enqueue([]event.EventEntry{{DefID: "no_name", Name: "", Week: 1}})
	assert.False(t, nq.HasPending())
}

// TestNewspaperQueue_Current_ReturnsFront verifies that Current returns the first item.
func TestNewspaperQueue_Current_ReturnsFront(t *testing.T) {
	var nq ui.NewspaperQueue
	nq.Enqueue(makeEntries("first", "second"))
	assert.Equal(t, "first", nq.Current().Name)
}

// TestNewspaperQueue_Dismiss_AdvancesQueue verifies that Dismiss removes the front item.
func TestNewspaperQueue_Dismiss_AdvancesQueue(t *testing.T) {
	var nq ui.NewspaperQueue
	nq.Enqueue(makeEntries("first", "second"))
	nq.Dismiss()
	assert.Equal(t, "second", nq.Current().Name)
}

// TestNewspaperQueue_Dismiss_EmptyQueue_DoesNotPanic verifies safe dismiss on empty queue.
func TestNewspaperQueue_Dismiss_EmptyQueue_DoesNotPanic(t *testing.T) {
	var nq ui.NewspaperQueue
	assert.NotPanics(t, func() { nq.Dismiss() })
}

// TestNewspaperQueue_Clear_EmptiesQueue verifies Clear removes all items.
func TestNewspaperQueue_Clear_EmptiesQueue(t *testing.T) {
	var nq ui.NewspaperQueue
	nq.Enqueue(makeEntries("a", "b", "c"))
	nq.Clear()
	assert.False(t, nq.HasPending())
}

// TestNewspaperQueue_EnqueueRich_UsesHeadline verifies that EnqueueRich sets headline
// from the provided map.
func TestNewspaperQueue_EnqueueRich_UsesHeadline(t *testing.T) {
	var nq ui.NewspaperQueue
	entries := []event.EventEntry{{DefID: "e1", Name: "Event One", Week: 1}}
	eventsMap := map[string][2]string{
		"e1": {"Big News Today", "The full narrative text here."},
	}
	nq.EnqueueRich(entries, eventsMap)
	assert.Equal(t, "Big News Today", nq.Current().Headline)
	assert.Equal(t, "The full narrative text here.", nq.Current().Narrative)
}

// TestNewspaperQueue_EnqueueRich_FallsBackToName verifies fallback when ID not in map.
func TestNewspaperQueue_EnqueueRich_FallsBackToName(t *testing.T) {
	var nq ui.NewspaperQueue
	entries := []event.EventEntry{{DefID: "unknown", Name: "My Event", Week: 1}}
	nq.EnqueueRich(entries, map[string][2]string{})
	assert.Equal(t, "My Event", nq.Current().Headline)
}

// TestNewscenarioScreen_New_DoesNotPanic verifies NewScenarioScreen returns non-nil.
func TestNewScenarioScreen_DoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		s := ui.NewScenarioScreen()
		assert.NotNil(t, s)
	})
}

// TestQueueNewsItems_DoesNotPanic verifies UI.QueueNewsItems handles empty and populated slices.
func TestQueueNewsItems_DoesNotPanic(t *testing.T) {
	world, cfg := newTestWorld(t)
	u := ui.New(&world, cfg)
	assert.NotPanics(t, func() {
		u.QueueNewsItems(nil)
		u.QueueNewsItems(makeEntries("test"))
	})
}
