package game_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/game"
	"github.com/vibe-code-challenge/twenty-fifty/internal/save"
)

func newTestGame(t *testing.T) *game.Game {
	t.Helper()
	cfg, err := config.Load()
	assert.NoError(t, err)
	seed, err := save.NewMasterSeed()
	assert.NoError(t, err)
	return game.New(cfg, seed, t.TempDir()+"\\autosave.json")
}

func TestNew_ReturnsNonNilGame(t *testing.T) {
	assert.NotNil(t, newTestGame(t))
}

func TestLayout_PassesThroughWindowDimensions(t *testing.T) {
	g := newTestGame(t)
	// Layout returns the outside (window) dimensions unchanged so the game
	// renders at native resolution rather than stretching a fixed buffer.
	w, h := g.Layout(1920, 1080)
	assert.Equal(t, 1920, w)
	assert.Equal(t, 1080, h)
}

func TestUpdate_ReturnsNoError(t *testing.T) {
	assert.NoError(t, newTestGame(t).Update())
}
