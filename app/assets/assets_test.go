package assets_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vibe-code-challenge/twenty-fifty/assets"
)

func TestUI_DecodesWithoutPanic(t *testing.T) {
	a := assets.UI()
	assert.NotNil(t, a.Logo)
	assert.NotNil(t, a.IconCarbon)
	assert.NotNil(t, a.IconBudget)
	assert.NotNil(t, a.IconPopularity)
	assert.NotNil(t, a.IconEnergy)
	assert.NotNil(t, a.NewspaperMasthead)
}

func TestUI_LogoDimensions(t *testing.T) {
	logo := assets.UI().Logo
	b := logo.Bounds()
	assert.Equal(t, 320, b.Dx())
	assert.Equal(t, 80, b.Dy())
}

func TestScenarios_DecodesWithoutPanic(t *testing.T) {
	s := assets.Scenarios()
	assert.NotNil(t, s.HumbleBeginnings)
	assert.NotNil(t, s.RisingStorm)
	assert.NotNil(t, s.Crossroads)
}

func TestScenarios_CardDimensions(t *testing.T) {
	s := assets.Scenarios()
	// Each scenario card must be 240x135.
	hb := s.HumbleBeginnings.Bounds()
	assert.Equal(t, 240, hb.Dx())
	assert.Equal(t, 135, hb.Dy())
}

func TestPortraits_DecodesWithoutPanic(t *testing.T) {
	p := assets.Portraits()
	assert.NotNil(t, p.EnergyMinister)
	assert.NotNil(t, p.Chancellor)
	assert.NotNil(t, p.GreenLobby)
	assert.NotNil(t, p.FossilLobby)
	assert.NotNil(t, p.TradeUnions)
	assert.NotNil(t, p.Press)
}

func TestPortraits_Dimensions(t *testing.T) {
	b := assets.Portraits().EnergyMinister.Bounds()
	assert.Equal(t, 64, b.Dx())
	assert.Equal(t, 64, b.Dy())
}

func TestRegionsOverlay_DecodesWithoutPanic(t *testing.T) {
	img := assets.RegionsOverlay()
	assert.NotNil(t, img)
	b := img.Bounds()
	assert.Equal(t, 960, b.Dx())
	assert.Equal(t, 720, b.Dy())
}
