package region

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"twenty-fifty/internal/carbon"
	"twenty-fifty/internal/config"
)

// ---------------------------------------------------------------------------
// SeedRegions
// ---------------------------------------------------------------------------

func TestSeedRegions_PopulatesFromDefs(t *testing.T) {
	defs := []config.RegionDef{
		{ID: "r1", Name: "Region One", TileIDs: []string{"t1", "t2"},
			InitialSkillsNetwork: 40, InitialInstallerCapacity: 10, InitialSupplyChain: 35},
	}
	regions := SeedRegions(defs)
	assert.Equal(t, 1, len(regions))
	r := regions[0]
	assert.Equal(t, "r1", r.ID)
	assert.Equal(t, "Region One", r.Name)
	assert.Equal(t, 40.0, r.SkillsNetwork)
	assert.Equal(t, 10.0, r.InstallerCapacity)
	assert.Equal(t, 35.0, r.SupplyChain)
	assert.Equal(t, []string{"t1", "t2"}, r.TileIDs)
}

func TestSeedRegions_TileIDsAreCopied(t *testing.T) {
	defs := []config.RegionDef{
		{ID: "r1", TileIDs: []string{"t1"}},
	}
	regions := SeedRegions(defs)
	defs[0].TileIDs[0] = "mutated"
	assert.Equal(t, "t1", regions[0].TileIDs[0], "SeedRegions must copy TileIDs, not alias them")
}

// ---------------------------------------------------------------------------
// SeedTiles
// ---------------------------------------------------------------------------

func TestSeedTiles_PopulatesFromDefs(t *testing.T) {
	defs := []config.TileDef{
		{ID: "t1", RegionID: "r1", Name: "Tile One",
			InitialInsulationLevel: 30, InitialHeatingType: config.HeatingGas,
			InitialLocalIncome: 50, InitialPoliticalOpinion: 55,
			InitialHeatingCapacity: 50, InitialInstallerQuality: 40},
	}
	tiles := SeedTiles(defs)
	assert.Equal(t, 1, len(tiles))
	ti := tiles[0]
	assert.Equal(t, "t1", ti.ID)
	assert.Equal(t, 30.0, ti.InsulationLevel)
	assert.Equal(t, config.HeatingGas, ti.HeatingType)
	assert.Equal(t, 55.0, ti.PoliticalOpinion)
}

func TestSeedTiles_AllAttributesStartUnrevealed(t *testing.T) {
	defs := []config.TileDef{{ID: "t1", RegionID: "r1"}}
	tiles := SeedTiles(defs)
	assert.Empty(t, tiles[0].RevealedAttributes, "all attributes must start unrevealed")
}

func TestSeedTiles_FuelPovertyStartsZero(t *testing.T) {
	defs := []config.TileDef{{ID: "t1", RegionID: "r1"}}
	tiles := SeedTiles(defs)
	assert.Equal(t, 0.0, tiles[0].FuelPoverty)
}

// ---------------------------------------------------------------------------
// RevealAttribute / IsRevealed
// ---------------------------------------------------------------------------

func TestIsRevealed_UnrevealedByDefault(t *testing.T) {
	tile := Tile{RevealedAttributes: make(map[string]bool)}
	assert.False(t, IsRevealed(tile, AttrInsulationLevel))
}

func TestRevealAttribute_RevealsCorrectAttribute(t *testing.T) {
	tile := Tile{RevealedAttributes: make(map[string]bool)}
	updated := RevealAttribute(tile, AttrInsulationLevel)
	assert.True(t, IsRevealed(updated, AttrInsulationLevel))
}

func TestRevealAttribute_DoesNotMutateOriginal(t *testing.T) {
	tile := Tile{RevealedAttributes: make(map[string]bool)}
	RevealAttribute(tile, AttrInsulationLevel)
	assert.False(t, IsRevealed(tile, AttrInsulationLevel), "original tile must not be mutated")
}

func TestRevealAttribute_OtherAttributesRemainUnrevealed(t *testing.T) {
	tile := Tile{RevealedAttributes: make(map[string]bool)}
	updated := RevealAttribute(tile, AttrInsulationLevel)
	assert.False(t, IsRevealed(updated, AttrFuelPoverty))
}

func TestRevealAttribute_MultipleRevealsSameAttribute_Idempotent(t *testing.T) {
	tile := Tile{RevealedAttributes: make(map[string]bool)}
	first := RevealAttribute(tile, AttrInsulationLevel)
	second := RevealAttribute(first, AttrInsulationLevel)
	assert.True(t, IsRevealed(second, AttrInsulationLevel))
}

// ---------------------------------------------------------------------------
// CapacityMultiplier
// ---------------------------------------------------------------------------

func TestCapacityMultiplier_ZeroCapacity_ReturnsZero(t *testing.T) {
	assert.Equal(t, 0.0, CapacityMultiplier(Region{InstallerCapacity: 0}))
}

func TestCapacityMultiplier_ReferenceCapacity_ReturnsOne(t *testing.T) {
	assert.InDelta(t, 1.0, CapacityMultiplier(Region{InstallerCapacity: referenceInstallerCapacity}), 0.001)
}

func TestCapacityMultiplier_AboveMax_ClampsToOne(t *testing.T) {
	assert.Equal(t, 1.0, CapacityMultiplier(Region{InstallerCapacity: referenceInstallerCapacity * 3}))
}

func TestCapacityMultiplier_HalfReference_ReturnsHalf(t *testing.T) {
	assert.InDelta(t, 0.5, CapacityMultiplier(Region{InstallerCapacity: referenceInstallerCapacity / 2}), 0.001)
}

// ---------------------------------------------------------------------------
// ComputeTrueRetrofitRate
// ---------------------------------------------------------------------------

func TestComputeTrueRetrofitRate_ZeroQuality_ReturnsZero(t *testing.T) {
	assert.Equal(t, 0.0, ComputeTrueRetrofitRate(80.0, 0))
}

func TestComputeTrueRetrofitRate_FullQuality_EqualsObserved(t *testing.T) {
	assert.InDelta(t, 80.0, ComputeTrueRetrofitRate(80.0, 100), 0.001)
}

func TestComputeTrueRetrofitRate_HalfQuality_HalvesObserved(t *testing.T) {
	assert.InDelta(t, 40.0, ComputeTrueRetrofitRate(80.0, 50), 0.001)
}

func TestComputeTrueRetrofitRate_AlwaysLessOrEqualObserved(t *testing.T) {
	for quality := 0.0; quality <= 100.0; quality += 10 {
		assert.LessOrEqual(t, ComputeTrueRetrofitRate(60.0, quality), 60.0,
			"true rate must never exceed observed at quality=%.0f", quality)
	}
}

// ---------------------------------------------------------------------------
// ComputeFuelPoverty
// ---------------------------------------------------------------------------

var referenceInput = FuelPovertyInput{
	GasPrice:              25.0, // GBP/MWh (2010 green book anchor)
	ElectricityPrice:      50.0,
	OilPrice:              80.0,
	HeatingType:           config.HeatingGas,
	InsulationLevel:       50.0,
	LocalIncome:           50.0, // median
	HeatingCapacity:       50.0,
	TechMaturityHeatPumps: 0,
	SeasonalMultiplier:    1.0,
}

func TestComputeFuelPoverty_ReturnsNonZeroForReferenceInput(t *testing.T) {
	result := ComputeFuelPoverty(referenceInput)
	assert.Greater(t, result, 0.0)
	assert.Less(t, result, 100.0)
}

func TestComputeFuelPoverty_ZeroIncome_ClampsAt100(t *testing.T) {
	in := referenceInput
	in.LocalIncome = 0
	assert.Equal(t, 100.0, ComputeFuelPoverty(in))
}

func TestComputeFuelPoverty_FullInsulation_LowerThanNoInsulation(t *testing.T) {
	baseline := ComputeFuelPoverty(referenceInput)
	in := referenceInput
	in.InsulationLevel = 100.0
	assert.Less(t, ComputeFuelPoverty(in), baseline)
}

func TestComputeFuelPoverty_HigherIncome_LowerPoverty(t *testing.T) {
	low := referenceInput
	low.LocalIncome = 25.0
	high := referenceInput
	high.LocalIncome = 75.0
	assert.Greater(t, ComputeFuelPoverty(low), ComputeFuelPoverty(high))
}

func TestComputeFuelPoverty_HeatPump_CheaperThanDirectElectric(t *testing.T) {
	electric := referenceInput
	electric.HeatingType = config.HeatingElectric

	hp := referenceInput
	hp.HeatingType = config.HeatingHeatPump
	hp.TechMaturityHeatPumps = 50.0

	assert.Less(t, ComputeFuelPoverty(hp), ComputeFuelPoverty(electric),
		"heat pump (COP > 1) must produce lower fuel poverty than direct electric")
}

func TestComputeFuelPoverty_HigherSeasonalMultiplier_IncreasesPoverty(t *testing.T) {
	baseline := ComputeFuelPoverty(referenceInput)
	in := referenceInput
	in.SeasonalMultiplier = 1.5
	assert.Greater(t, ComputeFuelPoverty(in), baseline)
}

func TestComputeFuelPoverty_AlwaysInBounds(t *testing.T) {
	cases := []FuelPovertyInput{
		referenceInput,
		{GasPrice: 0, ElectricityPrice: 0, OilPrice: 0,
			HeatingType: config.HeatingGas, LocalIncome: 1, SeasonalMultiplier: 1.0},
		{GasPrice: 999, ElectricityPrice: 999, OilPrice: 999,
			HeatingType: config.HeatingOil, LocalIncome: 1, SeasonalMultiplier: 2.0},
		{HeatingType: config.HeatingHeatPump, TechMaturityHeatPumps: 100,
			ElectricityPrice: 200, LocalIncome: 1, SeasonalMultiplier: 1.0},
	}
	for _, in := range cases {
		result := ComputeFuelPoverty(in)
		assert.GreaterOrEqual(t, result, 0.0)
		assert.LessOrEqual(t, result, 100.0)
	}
}

func TestComputeFuelPoverty_MixedHeating_BetweenGasAndElectric(t *testing.T) {
	gas := referenceInput
	gas.HeatingType = config.HeatingGas

	electric := referenceInput
	electric.HeatingType = config.HeatingElectric

	mixed := referenceInput
	mixed.HeatingType = config.HeatingMixed

	gasVal := ComputeFuelPoverty(gas)
	electricVal := ComputeFuelPoverty(electric)
	mixedVal := ComputeFuelPoverty(mixed)

	min, max := gasVal, electricVal
	if electricVal < gasVal {
		min, max = electricVal, gasVal
	}
	assert.GreaterOrEqual(t, mixedVal, min*0.9, "mixed should be close to the lower of gas/electric")
	assert.LessOrEqual(t, mixedVal, max*1.1, "mixed should be close to the higher of gas/electric")
}

// ---------------------------------------------------------------------------
// UpdateLocalPoliticalOpinion
// ---------------------------------------------------------------------------

func TestUpdateLocalPoliticalOpinion_FuelPovertyIncrease_ShiftsRight(t *testing.T) {
	tile := Tile{PoliticalOpinion: 50.0}
	updated := UpdateLocalPoliticalOpinion(tile, 5.0, 0, carbon.ClimateLevelStable)
	assert.Greater(t, updated.PoliticalOpinion, 50.0)
}

func TestUpdateLocalPoliticalOpinion_ClimateEvent_ShiftsLeft(t *testing.T) {
	tile := Tile{PoliticalOpinion: 50.0}
	updated := UpdateLocalPoliticalOpinion(tile, 0, 5.0, carbon.ClimateLevelStable)
	assert.Less(t, updated.PoliticalOpinion, 50.0)
}

func TestUpdateLocalPoliticalOpinion_ClampsAtUpperBound(t *testing.T) {
	tile := Tile{PoliticalOpinion: 99.0}
	pushed := UpdateLocalPoliticalOpinion(tile, 100.0, 0, carbon.ClimateLevelStable)
	assert.Equal(t, 100.0, pushed.PoliticalOpinion)
}

func TestUpdateLocalPoliticalOpinion_ClampsAtLowerBound(t *testing.T) {
	tile := Tile{PoliticalOpinion: 1.0}
	pulled := UpdateLocalPoliticalOpinion(tile, 0, 100.0, carbon.ClimateLevelStable)
	assert.Equal(t, 0.0, pulled.PoliticalOpinion)
}

func TestUpdateLocalPoliticalOpinion_EmergencyAmplifiesClimateShift(t *testing.T) {
	tile := Tile{PoliticalOpinion: 50.0}
	stable := UpdateLocalPoliticalOpinion(tile, 0, 2.0, carbon.ClimateLevelStable)
	emergency := UpdateLocalPoliticalOpinion(tile, 0, 2.0, carbon.ClimateLevelEmergency)
	assert.Less(t, emergency.PoliticalOpinion, stable.PoliticalOpinion,
		"EMERGENCY climate level must amplify the leftward shift")
}

func TestUpdateLocalPoliticalOpinion_DoesNotMutateOriginal(t *testing.T) {
	tile := Tile{PoliticalOpinion: 50.0}
	UpdateLocalPoliticalOpinion(tile, 5.0, 0, carbon.ClimateLevelStable)
	assert.Equal(t, 50.0, tile.PoliticalOpinion, "original tile must not be mutated")
}
