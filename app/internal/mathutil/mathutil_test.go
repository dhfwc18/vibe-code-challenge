package mathutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClamp_BelowMin_ReturnsMin(t *testing.T) {
	assert.Equal(t, 0.0, Clamp(-5.0, 0.0, 100.0))
}

func TestClamp_AboveMax_ReturnsMax(t *testing.T) {
	assert.Equal(t, 100.0, Clamp(150.0, 0.0, 100.0))
}

func TestClamp_WithinRange_ReturnsValue(t *testing.T) {
	assert.Equal(t, 50.0, Clamp(50.0, 0.0, 100.0))
}

func TestClamp_AtMin_ReturnsMin(t *testing.T) {
	assert.Equal(t, 0.0, Clamp(0.0, 0.0, 100.0))
}

func TestClamp_AtMax_ReturnsMax(t *testing.T) {
	assert.Equal(t, 100.0, Clamp(100.0, 0.0, 100.0))
}
