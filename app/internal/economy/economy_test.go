package economy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// NewEconomyState
// ---------------------------------------------------------------------------

func TestNewEconomyState_ValueAboveMedian(t *testing.T) {
	state := NewEconomyState()
	assert.Greater(t, state.Value, 50.0)
	assert.LessOrEqual(t, state.Value, 100.0)
}

func TestNewEconomyState_LobbyEffectsEmpty(t *testing.T) {
	state := NewEconomyState()
	assert.Empty(t, state.LobbyEffects)
}

// ---------------------------------------------------------------------------
// TickEconomy
// ---------------------------------------------------------------------------

func TestTickEconomy_PolicyBonus_IncreasesValue(t *testing.T) {
	state := EconomyState{Value: 50.0, LobbyEffects: make(map[string]float64)}
	updated := TickEconomy(state, 0, 0, 0, 2.0, 0)
	assert.Greater(t, updated.Value, 50.0)
}

func TestTickEconomy_ClimateDamage_DecreasesValue(t *testing.T) {
	state := EconomyState{Value: 50.0, LobbyEffects: make(map[string]float64)}
	updated := TickEconomy(state, 5.0, 0, 0, 0, 0)
	assert.Less(t, updated.Value, 50.0)
}

func TestTickEconomy_ClampsAtZero(t *testing.T) {
	state := EconomyState{Value: 1.0, LobbyEffects: make(map[string]float64)}
	updated := TickEconomy(state, 999, 0, 0, 0, 0)
	assert.Equal(t, 0.0, updated.Value)
}

func TestTickEconomy_ClampsAt100(t *testing.T) {
	state := EconomyState{Value: 99.0, LobbyEffects: make(map[string]float64)}
	updated := TickEconomy(state, 0, 0, 0, 999, 0)
	assert.Equal(t, 100.0, updated.Value)
}

func TestTickEconomy_DoesNotMutateOriginal(t *testing.T) {
	state := EconomyState{Value: 50.0, LobbyEffects: make(map[string]float64)}
	TickEconomy(state, 5.0, 0, 0, 0, 0)
	assert.Equal(t, 50.0, state.Value)
}

func TestTickEconomy_AllInputsZero_ValueUnchanged(t *testing.T) {
	state := EconomyState{Value: 60.0, LobbyEffects: make(map[string]float64)}
	updated := TickEconomy(state, 0, 0, 0, 0, 0)
	assert.InDelta(t, 60.0, updated.Value, 0.001)
}

// ---------------------------------------------------------------------------
// ComputeTaxRevenue
// ---------------------------------------------------------------------------

func TestComputeTaxRevenue_MedianEconomy_ReturnsReferenceRevenue(t *testing.T) {
	state := EconomyState{Value: 50.0}
	rev := ComputeTaxRevenue(state, 1, 2010)
	expected := referenceAnnualRevenueGBPbn / 4.0
	assert.InDelta(t, expected, rev.GBPBillions, 0.01)
}

func TestComputeTaxRevenue_ZeroEconomy_ReturnsZero(t *testing.T) {
	state := EconomyState{Value: 0.0}
	rev := ComputeTaxRevenue(state, 1, 2010)
	assert.Equal(t, 0.0, rev.GBPBillions)
}

func TestComputeTaxRevenue_HighEconomy_HigherThanMedian(t *testing.T) {
	low := ComputeTaxRevenue(EconomyState{Value: 50.0}, 1, 2010)
	high := ComputeTaxRevenue(EconomyState{Value: 80.0}, 1, 2010)
	assert.Greater(t, high.GBPBillions, low.GBPBillions)
}

func TestComputeTaxRevenue_SetsQuarterAndYear(t *testing.T) {
	state := EconomyState{Value: 50.0}
	rev := ComputeTaxRevenue(state, 3, 2025)
	assert.Equal(t, 3, rev.Quarter)
	assert.Equal(t, 2025, rev.Year)
}

func TestComputeTaxRevenue_AlwaysNonNegative(t *testing.T) {
	for v := 0.0; v <= 100.0; v += 10 {
		rev := ComputeTaxRevenue(EconomyState{Value: v}, 1, 2010)
		assert.GreaterOrEqual(t, rev.GBPBillions, 0.0)
	}
}

// ---------------------------------------------------------------------------
// AllocateBudget
// ---------------------------------------------------------------------------

var testBaseFractions = map[string]float64{
	"energy":    0.40,
	"transport": 0.35,
	"housing":   0.25,
}

var testPopWeights = map[string]float64{
	"energy":    60.0,
	"transport": 50.0,
	"housing":   45.0,
}

func TestAllocateBudget_SharesSumToTotal(t *testing.T) {
	revenue := TaxRevenue{GBPBillions: 55.0, Quarter: 1, Year: 2010}
	alloc := AllocateBudget(revenue, testBaseFractions, testPopWeights, 1.0, nil)
	sum := 0.0
	for _, v := range alloc.Departments {
		sum += v
	}
	assert.InDelta(t, alloc.TotalGBPm, sum, 0.01)
}

func TestAllocateBudget_TotalGBPmIsDiscretionaryFraction(t *testing.T) {
	revenue := TaxRevenue{GBPBillions: 55.0}
	alloc := AllocateBudget(revenue, testBaseFractions, testPopWeights, 1.0, nil)
	expected := 55.0 * 1000 * discretionaryFraction
	assert.InDelta(t, expected, alloc.TotalGBPm, 0.01)
}

func TestAllocateBudget_HigherPopWeight_LargerShare(t *testing.T) {
	fractions := map[string]float64{"a": 0.5, "b": 0.5}
	popWeights := map[string]float64{"a": 80.0, "b": 20.0}
	revenue := TaxRevenue{GBPBillions: 55.0}
	alloc := AllocateBudget(revenue, fractions, popWeights, 1.0, nil)
	assert.Greater(t, alloc.Departments["a"], alloc.Departments["b"])
}

func TestAllocateBudget_LobbyEffect_IncreasesShare(t *testing.T) {
	fractions := map[string]float64{"energy": 0.5, "transport": 0.5}
	popWeights := map[string]float64{"energy": 50.0, "transport": 50.0}
	revenue := TaxRevenue{GBPBillions: 55.0}

	noLobby := AllocateBudget(revenue, fractions, popWeights, 1.0, nil)
	withLobby := AllocateBudget(revenue, fractions, popWeights, 1.0,
		map[string]float64{"energy": 2.0})

	assert.Greater(t, withLobby.Departments["energy"], noLobby.Departments["energy"])
}

func TestAllocateBudget_AllDepartmentsPresent(t *testing.T) {
	revenue := TaxRevenue{GBPBillions: 55.0}
	alloc := AllocateBudget(revenue, testBaseFractions, testPopWeights, 1.0, nil)
	for dept := range testBaseFractions {
		_, ok := alloc.Departments[dept]
		assert.True(t, ok, "department %q missing from allocation", dept)
	}
}

// ---------------------------------------------------------------------------
// AccumulateLobbyEffect
// ---------------------------------------------------------------------------

func TestAccumulateLobbyEffect_MultipliesEffect(t *testing.T) {
	state := NewEconomyState()
	updated := AccumulateLobbyEffect(state, "energy", 1.5)
	assert.InDelta(t, 1.5, updated.LobbyEffects["energy"], 0.001)
}

func TestAccumulateLobbyEffect_SecondAccumulation_Multiplies(t *testing.T) {
	state := NewEconomyState()
	state = AccumulateLobbyEffect(state, "energy", 1.5)
	state = AccumulateLobbyEffect(state, "energy", 2.0)
	assert.InDelta(t, 3.0, state.LobbyEffects["energy"], 0.001)
}

func TestAccumulateLobbyEffect_DoesNotMutateOriginal(t *testing.T) {
	state := NewEconomyState()
	AccumulateLobbyEffect(state, "energy", 2.0)
	_, exists := state.LobbyEffects["energy"]
	assert.False(t, exists, "original state must not be mutated")
}

// ---------------------------------------------------------------------------
// ClearLobbyEffectsAtQuarterEnd
// ---------------------------------------------------------------------------

func TestClearLobbyEffectsAtQuarterEnd_RemovesAllEffects(t *testing.T) {
	state := NewEconomyState()
	state = AccumulateLobbyEffect(state, "energy", 2.0)
	state = AccumulateLobbyEffect(state, "transport", 1.5)
	cleared := ClearLobbyEffectsAtQuarterEnd(state)
	assert.Empty(t, cleared.LobbyEffects)
}

func TestClearLobbyEffectsAtQuarterEnd_DoesNotMutateOriginal(t *testing.T) {
	state := NewEconomyState()
	state = AccumulateLobbyEffect(state, "energy", 2.0)
	ClearLobbyEffectsAtQuarterEnd(state)
	assert.NotEmpty(t, state.LobbyEffects, "original state must not be mutated")
}

// ---------------------------------------------------------------------------
// DeriveFossilDependency
// ---------------------------------------------------------------------------

func TestDeriveFossilDependency_AllFossil_Returns100(t *testing.T) {
	assert.InDelta(t, 100.0, DeriveFossilDependency(0.7, 0.3), 0.001)
}

func TestDeriveFossilDependency_AllRenewable_ReturnsZero(t *testing.T) {
	assert.Equal(t, 0.0, DeriveFossilDependency(0, 0))
}

func TestDeriveFossilDependency_HalfFossil_Returns50(t *testing.T) {
	assert.InDelta(t, 50.0, DeriveFossilDependency(0.3, 0.2), 0.001)
}

func TestDeriveFossilDependency_AlwaysInBounds(t *testing.T) {
	for gas := 0.0; gas <= 1.0; gas += 0.2 {
		for oil := 0.0; oil <= 1.0; oil += 0.2 {
			v := DeriveFossilDependency(gas, oil)
			assert.GreaterOrEqual(t, v, 0.0)
			assert.LessOrEqual(t, v, 100.0)
		}
	}
}
