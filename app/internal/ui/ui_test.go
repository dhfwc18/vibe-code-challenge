package ui_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/save"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
	"github.com/vibe-code-challenge/twenty-fifty/internal/ui"
)

// newTestWorld creates a minimal WorldState for UI tests.
func newTestWorld(t *testing.T) (simulation.WorldState, *config.Config) {
	t.Helper()
	cfg, err := config.Load()
	assert.NoError(t, err)
	seed, err := save.NewMasterSeed()
	assert.NoError(t, err)
	return simulation.NewWorld(cfg, seed), cfg
}

// TestNew_DoesNotPanic verifies that ui.New returns a non-nil UI without panicking.
func TestNew_DoesNotPanic(t *testing.T) {
	world, cfg := newTestWorld(t)
	u := ui.New(&world, cfg)
	assert.NotNil(t, u)
}

// TestUpdate_NoActions_ReturnsEmpty verifies that Update with no user input
// returns an empty action slice.
func TestUpdate_NoActions_ReturnsEmpty(t *testing.T) {
	world, cfg := newTestWorld(t)
	u := ui.New(&world, cfg)
	actions := u.Update(&world)
	assert.Empty(t, actions)
}

// TestDraw_DoesNotPanic verifies that Draw does not panic with a real world state.
func TestDraw_DoesNotPanic(t *testing.T) {
	world, cfg := newTestWorld(t)
	u := ui.New(&world, cfg)

	screen := ebiten.NewImage(1280, 720)
	assert.NotPanics(t, func() {
		u.Draw(screen, world)
	})
}

// TestAdvanceWeekRequested_FalseByDefault verifies that AdvanceWeekRequested
// returns false when no click has occurred.
func TestAdvanceWeekRequested_FalseByDefault(t *testing.T) {
	world, cfg := newTestWorld(t)
	u := ui.New(&world, cfg)
	assert.False(t, u.AdvanceWeekRequested())
}

// TestNotifyEvent_DoesNotPanic verifies that NotifyEvent handles arbitrary strings.
func TestNotifyEvent_DoesNotPanic(t *testing.T) {
	world, cfg := newTestWorld(t)
	u := ui.New(&world, cfg)
	assert.NotPanics(t, func() {
		u.NotifyEvent("test_event")
	})
}
