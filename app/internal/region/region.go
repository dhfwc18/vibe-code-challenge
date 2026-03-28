package region

import (
	"twenty-fifty/internal/carbon"
	"twenty-fifty/internal/config"
	"twenty-fifty/internal/mathutil"
)

// Region holds the live state of one of Taitan's 12 administrative regions.
type Region struct {
	ID                string
	Name              string
	TileIDs           []string
	SkillsNetwork     float64 // 0-100: available skilled workforce density
	InstallerCapacity float64 // installs per week
	SupplyChain       float64 // 0-100: component supply chain robustness
}

// Tile holds the live state of one map tile.
type Tile struct {
	ID                 string
	RegionID           string
	Name               string
	InsulationLevel    float64            // 0-100; higher = better insulated
	HeatingType        config.HeatingType
	LocalIncome        float64            // 0-100; 50 = median Taitan household income
	PoliticalOpinion   float64            // 0-100; >50 right-leaning, <50 left-leaning
	HeatingCapacity    float64            // 0-100
	InstallerQuality    float64            // 0-100; drives the true/observed retrofit gap
	FuelPoverty         float64            // 0-100; computed each week by the simulation
	ObservedRetrofitRate float64           // 0-100; survey-reported completion rate (may overstate)
	TrueRetrofitRate    float64            // 0-100; actual completion rate; see ComputeTrueRetrofitRate
	RevealedAttributes  map[string]bool    // fog-of-war: keys are attribute names
}

// Tile attribute name constants used as keys in RevealedAttributes.
const (
	AttrInsulationLevel      = "insulation_level"
	AttrHeatingType          = "heating_type"
	AttrLocalIncome          = "local_income"
	AttrInstallerQuality     = "installer_quality"
	AttrHeatingCapacity      = "heating_capacity"
	AttrFuelPoverty          = "fuel_poverty"
	AttrObservedRetrofitRate = "observed_retrofit_rate"
	AttrTrueRetrofitRate     = "true_retrofit_rate"
)

// referenceInstallerCapacity is the normalisation denominator for CapacityMultiplier.
// Represents the maximum installs per week a fully-developed region could achieve.
const referenceInstallerCapacity = 50.0

// SeedRegions creates live Region values from static config definitions.
func SeedRegions(defs []config.RegionDef) []Region {
	regions := make([]Region, len(defs))
	for i, d := range defs {
		tileIDs := make([]string, len(d.TileIDs))
		copy(tileIDs, d.TileIDs)
		regions[i] = Region{
			ID:                d.ID,
			Name:              d.Name,
			TileIDs:           tileIDs,
			SkillsNetwork:     d.InitialSkillsNetwork,
			InstallerCapacity: d.InitialInstallerCapacity,
			SupplyChain:       d.InitialSupplyChain,
		}
	}
	return regions
}

// SeedTiles creates live Tile values from static config definitions.
// All attributes start unrevealed (fog-of-war); FuelPoverty starts at zero.
func SeedTiles(defs []config.TileDef) []Tile {
	tiles := make([]Tile, len(defs))
	for i, d := range defs {
		tiles[i] = Tile{
			ID:                 d.ID,
			RegionID:           d.RegionID,
			Name:               d.Name,
			InsulationLevel:    d.InitialInsulationLevel,
			HeatingType:        d.InitialHeatingType,
			LocalIncome:        d.InitialLocalIncome,
			PoliticalOpinion:   d.InitialPoliticalOpinion,
			HeatingCapacity:    d.InitialHeatingCapacity,
			InstallerQuality:     d.InitialInstallerQuality,
			FuelPoverty:          0,
			ObservedRetrofitRate: 0,
			TrueRetrofitRate:     0,
			RevealedAttributes:   make(map[string]bool),
		}
	}
	return tiles
}

// RevealAttribute unlocks a tile attribute from fog-of-war.
// Returns a new Tile with the named attribute marked visible; the input is not mutated.
func RevealAttribute(tile Tile, attribute string) Tile {
	revealed := make(map[string]bool, len(tile.RevealedAttributes)+1)
	for k, v := range tile.RevealedAttributes {
		revealed[k] = v
	}
	revealed[attribute] = true
	tile.RevealedAttributes = revealed
	return tile
}

// IsRevealed reports whether a named attribute is currently visible to the player.
func IsRevealed(tile Tile, attribute string) bool {
	return tile.RevealedAttributes[attribute]
}

// CapacityMultiplier returns a [0, 1] fraction of the reference installer
// capacity for this region. Used as a multiplier in policy carbon-effect resolution.
func CapacityMultiplier(r Region) float64 {
	return mathutil.Clamp(r.InstallerCapacity/referenceInstallerCapacity, 0, 1)
}

// ComputeTrueRetrofitRate returns the actual retrofit completion rate after
// accounting for installer quality. Observed rates from surveys are systematically
// higher than reality when installer quality is low.
//
//	trueRate = observedRate * (installerQuality / 100)
//
// Always returns a value in [0, observedRate].
func ComputeTrueRetrofitRate(observedRate, installerQuality float64) float64 {
	q := mathutil.Clamp(installerQuality, 0, 100) / 100.0
	return mathutil.Clamp(observedRate*q, 0, 100)
}

// ApplyClimateEventDamage adds fuelPovertyDelta to a tile's FuelPoverty score,
// clamped to [0, 100]. Used when a climate event directly worsens household
// energy costs (e.g. a cold snap or infrastructure damage event).
// Returns a new Tile; the input is not mutated.
func ApplyClimateEventDamage(tile Tile, fuelPovertyDelta float64) Tile {
	tile.FuelPoverty = mathutil.Clamp(tile.FuelPoverty+fuelPovertyDelta, 0, 100)
	return tile
}

// UpdateLocalPoliticalOpinion updates a tile's political opinion for one week.
//
// Fuel poverty worsening (positive fuelPovertyDelta) shifts opinion right (higher).
// Climate event impact (positive climateEventImpact) shifts opinion left (lower);
// this effect is amplified by the current climate severity level.
//
// Returns a new Tile; the input is not mutated.
func UpdateLocalPoliticalOpinion(tile Tile, fuelPovertyDelta, climateEventImpact float64, level carbon.ClimateLevel) Tile {
	// STABLE=1.0x, ELEVATED=1.25x, CRITICAL=1.5x, EMERGENCY=1.75x
	climateAmplifier := 1.0 + float64(level)*0.25
	opinion := tile.PoliticalOpinion
	opinion += fuelPovertyDelta * 0.30
	opinion -= climateEventImpact * 0.50 * climateAmplifier
	tile.PoliticalOpinion = mathutil.Clamp(opinion, 0, 100)
	return tile
}

