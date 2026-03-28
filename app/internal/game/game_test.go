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
	return game.New(cfg, seed)
}

func TestNew_ReturnsNonNilGame(t *testing.T) {
	assert.NotNil(t, newTestGame(t))
}

func TestLayout_ReturnsExpectedDimensions(t *testing.T) {
	g := newTestGame(t)
	w, h := g.Layout(0, 0)
	assert.Equal(t, game.ScreenWidth, w)
	assert.Equal(t, game.ScreenHeight, h)
}

func TestUpdate_ReturnsNoError(t *testing.T) {
	assert.NoError(t, newTestGame(t).Update())
}
