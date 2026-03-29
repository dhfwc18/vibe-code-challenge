// Package assets embeds all game image assets and exposes named accessors.
// Every file under images/ is a placeholder PNG generated at the correct
// target dimensions. Replace each file with production art before shipping;
// the dimensions documented below are the contract the UI code depends on.
//
// Placeholder colours:
//   - ui/logo.png                320x80   dark-navy fill
//   - ui/icon_*.png               32x32   colour-coded per resource type
//   - ui/newspaper_masthead.png  640x80   off-white fill
//   - scenarios/*.png            240x135  muted scene-colour fill
//   - stakeholders/*.png          64x64   grey fill
//   - map/regions_overlay.png    960x720  black (transparent areas added in prod)
package assets

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
)

// --- UI ---

//go:embed images/ui/logo.png
var logoBytes []byte

//go:embed images/ui/icon_carbon.png
var iconCarbonBytes []byte

//go:embed images/ui/icon_budget.png
var iconBudgetBytes []byte

//go:embed images/ui/icon_popularity.png
var iconPopularityBytes []byte

//go:embed images/ui/icon_energy.png
var iconEnergyBytes []byte

//go:embed images/ui/newspaper_masthead.png
var newspaperMastheadBytes []byte

// --- Scenario cards ---

//go:embed images/scenarios/humble_beginnings.png
var scenarioHumbleBeginningsBytes []byte

//go:embed images/scenarios/rising_storm.png
var scenarioRisingStormBytes []byte

//go:embed images/scenarios/crossroads.png
var scenarioCrossroadsBytes []byte

// --- Stakeholder portraits ---

//go:embed images/stakeholders/energy_minister.png
var portraitEnergyMinisterBytes []byte

//go:embed images/stakeholders/chancellor.png
var portraitChancellorBytes []byte

//go:embed images/stakeholders/green_lobby.png
var portraitGreenLobbyBytes []byte

//go:embed images/stakeholders/fossil_lobby.png
var portraitFossilLobbyBytes []byte

//go:embed images/stakeholders/trade_unions.png
var portraitTradeUnionsBytes []byte

//go:embed images/stakeholders/press.png
var portraitPressBytes []byte

// --- Map ---

//go:embed images/map/regions_overlay.png
var regionsOverlayBytes []byte

// decode decodes a PNG from a byte slice. Panics on corrupt data (only
// embedded assets are passed, so a panic indicates a bad build).
func decode(b []byte) image.Image {
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		panic("assets: failed to decode embedded image: " + err.Error())
	}
	return img
}

// UI returns decoded UI image assets.
func UI() UIAssets {
	return UIAssets{
		Logo:              decode(logoBytes),
		IconCarbon:        decode(iconCarbonBytes),
		IconBudget:        decode(iconBudgetBytes),
		IconPopularity:    decode(iconPopularityBytes),
		IconEnergy:        decode(iconEnergyBytes),
		NewspaperMasthead: decode(newspaperMastheadBytes),
	}
}

// UIAssets holds decoded UI images.
type UIAssets struct {
	Logo              image.Image
	IconCarbon        image.Image
	IconBudget        image.Image
	IconPopularity    image.Image
	IconEnergy        image.Image
	NewspaperMasthead image.Image
}

// Scenarios returns decoded scenario card images keyed by scenario ID.
func Scenarios() ScenarioAssets {
	return ScenarioAssets{
		HumbleBeginnings: decode(scenarioHumbleBeginningsBytes),
		RisingStorm:      decode(scenarioRisingStormBytes),
		Crossroads:       decode(scenarioCrossroadsBytes),
	}
}

// ScenarioAssets holds decoded scenario card images.
type ScenarioAssets struct {
	HumbleBeginnings image.Image
	RisingStorm      image.Image
	Crossroads       image.Image
}

// Portraits returns decoded stakeholder portrait images.
func Portraits() PortraitAssets {
	return PortraitAssets{
		EnergyMinister: decode(portraitEnergyMinisterBytes),
		Chancellor:     decode(portraitChancellorBytes),
		GreenLobby:     decode(portraitGreenLobbyBytes),
		FossilLobby:    decode(portraitFossilLobbyBytes),
		TradeUnions:    decode(portraitTradeUnionsBytes),
		Press:          decode(portraitPressBytes),
	}
}

// PortraitAssets holds decoded stakeholder portrait images.
type PortraitAssets struct {
	EnergyMinister image.Image
	Chancellor     image.Image
	GreenLobby     image.Image
	FossilLobby    image.Image
	TradeUnions    image.Image
	Press          image.Image
}

// RegionsOverlay returns the decoded map regions overlay image (960x720).
// In production this will have transparent regions that sit above the tile map.
func RegionsOverlay() image.Image {
	return decode(regionsOverlayBytes)
}
