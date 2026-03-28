package carbon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vibe-code-challenge/app/internal/config"
)

// testBudgets is a minimal three-entry table used across all tests.
var testBudgets = []config.CarbonBudgetEntry{
	{Year: 2010, AnnualLimitMtCO2e: 590.0},
	{Year: 2020, AnnualLimitMtCO2e: 300.0},
	{Year: 2050, AnnualLimitMtCO2e: 0.0},
}

// ---------------------------------------------------------------------------
// AccumulateWeekly
// ---------------------------------------------------------------------------

func TestAccumulateWeekly_PositiveEmissions_IncreasesTotals(t *testing.T) {
	state := CarbonBudgetState{}
	result := AccumulateWeekly(state, 10.0)
	assert.Equal(t, 10.0, result.RunningAnnualTotal)
	assert.Equal(t, 10.0, result.CumulativeStock)
}

func TestAccumulateWeekly_NegativeEmissions_DecreasesBoth(t *testing.T) {
	state := CarbonBudgetState{RunningAnnualTotal: 20.0, CumulativeStock: 20.0}
	result := AccumulateWeekly(state, -5.0)
	assert.Equal(t, 15.0, result.RunningAnnualTotal)
	assert.Equal(t, 15.0, result.CumulativeStock)
}

func TestAccumulateWeekly_ZeroEmissions_NoChange(t *testing.T) {
	state := CarbonBudgetState{RunningAnnualTotal: 100.0, CumulativeStock: 200.0}
	result := AccumulateWeekly(state, 0.0)
	assert.Equal(t, 100.0, result.RunningAnnualTotal)
	assert.Equal(t, 200.0, result.CumulativeStock)
}

func TestAccumulateWeekly_DoesNotMutateInput(t *testing.T) {
	state := CarbonBudgetState{RunningAnnualTotal: 50.0}
	AccumulateWeekly(state, 10.0)
	assert.Equal(t, 50.0, state.RunningAnnualTotal)
}

// ---------------------------------------------------------------------------
// CheckAnnualBudget
// ---------------------------------------------------------------------------

func TestCheckAnnualBudget_UnderLimit_NotOverBudget(t *testing.T) {
	state := CarbonBudgetState{RunningAnnualTotal: 280.0}
	result, newState := CheckAnnualBudget(state, 2020, testBudgets)
	assert.False(t, result.IsOverBudget)
	assert.Equal(t, 0.0, result.Overshoot)
	assert.Equal(t, 0.0, newState.OvershootAccumulator)
}

func TestCheckAnnualBudget_OverLimit_AccumulatesOvershoot(t *testing.T) {
	state := CarbonBudgetState{RunningAnnualTotal: 350.0}
	result, newState := CheckAnnualBudget(state, 2020, testBudgets)
	assert.True(t, result.IsOverBudget)
	assert.InDelta(t, 50.0, result.Overshoot, 0.001)
	assert.InDelta(t, 50.0, newState.OvershootAccumulator, 0.001)
}

func TestCheckAnnualBudget_ResetsRunningAnnualTotal(t *testing.T) {
	state := CarbonBudgetState{RunningAnnualTotal: 400.0}
	_, newState := CheckAnnualBudget(state, 2020, testBudgets)
	assert.Equal(t, 0.0, newState.RunningAnnualTotal)
}

func TestCheckAnnualBudget_PreservesExistingOvershoot(t *testing.T) {
	state := CarbonBudgetState{RunningAnnualTotal: 350.0, OvershootAccumulator: 100.0}
	_, newState := CheckAnnualBudget(state, 2020, testBudgets)
	assert.InDelta(t, 150.0, newState.OvershootAccumulator, 0.001)
}

func TestCheckAnnualBudget_SetsNextYearLimit(t *testing.T) {
	state := CarbonBudgetState{RunningAnnualTotal: 100.0}
	_, newState := CheckAnnualBudget(state, 2010, testBudgets)
	// next year = 2011; interpolated between 2010(590) and 2020(300)
	// t = 1/10 = 0.1; limit = 590 + 0.1*(300-590) = 590 - 29 = 561
	assert.InDelta(t, 561.0, newState.CurrentBudgetLimit, 0.1)
}

// ---------------------------------------------------------------------------
// ProjectTrajectory
// ---------------------------------------------------------------------------

func TestProjectTrajectory_HalfYear_DoublesCurrent(t *testing.T) {
	state := CarbonBudgetState{RunningAnnualTotal: 100.0}
	assert.InDelta(t, 200.0, ProjectTrajectory(state, 26), 0.01)
}

func TestProjectTrajectory_FullYear_ReturnsCurrent(t *testing.T) {
	state := CarbonBudgetState{RunningAnnualTotal: 300.0}
	assert.InDelta(t, 300.0, ProjectTrajectory(state, 52), 0.01)
}

func TestProjectTrajectory_ZeroWeeks_ClampsToOne(t *testing.T) {
	state := CarbonBudgetState{RunningAnnualTotal: 52.0}
	// clamped to 1 week: 52.0/1 * 52 = 2704
	assert.InDelta(t, 2704.0, ProjectTrajectory(state, 0), 0.1)
}

func TestProjectTrajectory_OverMaxWeeks_ClampsTo52(t *testing.T) {
	state := CarbonBudgetState{RunningAnnualTotal: 300.0}
	assert.InDelta(t, 300.0, ProjectTrajectory(state, 100), 0.01)
}

func TestProjectTrajectory_Monotone_DecreasingWeeks(t *testing.T) {
	state := CarbonBudgetState{RunningAnnualTotal: 200.0}
	// Earlier in year => higher projected total (same total, fewer weeks elapsed)
	at10 := ProjectTrajectory(state, 10)
	at26 := ProjectTrajectory(state, 26)
	at52 := ProjectTrajectory(state, 52)
	assert.Greater(t, at10, at26)
	assert.Greater(t, at26, at52)
}

// ---------------------------------------------------------------------------
// ClimateStockToLevel
// ---------------------------------------------------------------------------

func TestClimateStockToLevel_BelowElevated_ReturnsStable(t *testing.T) {
	assert.Equal(t, ClimateLevelStable, ClimateStockToLevel(0))
	assert.Equal(t, ClimateLevelStable, ClimateStockToLevel(199.9))
}

func TestClimateStockToLevel_AtThresholds_ReturnsCorrectLevel(t *testing.T) {
	assert.Equal(t, ClimateLevelElevated, ClimateStockToLevel(ThresholdElevated))
	assert.Equal(t, ClimateLevelCritical, ClimateStockToLevel(ThresholdCritical))
	assert.Equal(t, ClimateLevelEmergency, ClimateStockToLevel(ThresholdEmergency))
}

func TestClimateStockToLevel_AboveEmergency_ReturnsEmergency(t *testing.T) {
	assert.Equal(t, ClimateLevelEmergency, ClimateStockToLevel(9999.0))
}

// ---------------------------------------------------------------------------
// LevelLabel
// ---------------------------------------------------------------------------

func TestLevelLabel_AllLevels_ReturnNonEmpty(t *testing.T) {
	levels := []ClimateLevel{ClimateLevelStable, ClimateLevelElevated, ClimateLevelCritical, ClimateLevelEmergency}
	for _, l := range levels {
		assert.NotEmpty(t, LevelLabel(l))
	}
}

func TestLevelLabel_StableLevel_ReturnsStable(t *testing.T) {
	assert.Equal(t, "STABLE", LevelLabel(ClimateLevelStable))
}

// ---------------------------------------------------------------------------
// limitForYear (internal; tested via CheckAnnualBudget and directly)
// ---------------------------------------------------------------------------

func TestLimitForYear_ExactEntries_ReturnExactValues(t *testing.T) {
	assert.InDelta(t, 590.0, limitForYear(2010, testBudgets), 0.001)
	assert.InDelta(t, 300.0, limitForYear(2020, testBudgets), 0.001)
	assert.InDelta(t, 0.0, limitForYear(2050, testBudgets), 0.001)
}

func TestLimitForYear_InterpolatesMidpoint(t *testing.T) {
	// midpoint 2010(590) - 2020(300): year=2015, t=0.5, expected=445
	assert.InDelta(t, 445.0, limitForYear(2015, testBudgets), 0.01)
}

func TestLimitForYear_BeforeFirstEntry_ReturnsFirst(t *testing.T) {
	assert.InDelta(t, 590.0, limitForYear(2000, testBudgets), 0.001)
}

func TestLimitForYear_AfterLastEntry_ReturnsLast(t *testing.T) {
	assert.InDelta(t, 0.0, limitForYear(2060, testBudgets), 0.001)
}
