package game_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"twenty-fifty/internal/game"
)

func TestNew_ReturnsNonNilGame(t *testing.T) {
	g := game.New()
	assert.NotNil(t, g)
}

func TestLayout_ReturnsExpectedDimensions(t *testing.T) {
	g := game.New()
	w, h := g.Layout(0, 0)
	assert.Equal(t, game.ScreenWidth, w)
	assert.Equal(t, game.ScreenHeight, h)
}

func TestUpdate_ReturnsNoError(t *testing.T) {
	g := game.New()
	assert.NoError(t, g.Update())
}
