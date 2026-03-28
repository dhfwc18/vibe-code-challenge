package technology

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
)

var testCurve = config.TechCurveDef{
	ID:                config.TechOffshoreWind,
	Name:              "Offshore Wind",
	Sector:            config.SectorPower,
	LogisticMidpoint:  520,
	LogisticSteepness: 0.01,
	BaseAdoptionRate:  0.08,
	InitialMaturity:   18.0,
}

// ---------------------------------------------------------------------------
// EvaluateLogistic
// ---------------------------------------------------------------------------

func TestEvaluateLogistic_AtMidpoint_ReturnsHalf(t *testing.T) {
	assert.InDelta(t, 0.5, EvaluateLogistic(50, 50, 0.1), 0.0001)
}

func TestEvaluateLogistic_Monotone_IncreasesWithX(t *testing.T) {
	prev := EvaluateLogistic(0, 50, 0.1)
	for x := 10.0; x <= 100.0; x += 10 {
		curr := EvaluateLogistic(x, 50, 0.1)
		assert.Greater(t, curr, prev, "logistic must be monotone increasing at x=%.0f", x)
		prev = curr
	}
}

func TestEvaluateLogistic_HigherSteepness_SharpensSlope(t *testing.T) {
	gentle := EvaluateLogistic(60, 50, 0.1)
	steep := EvaluateLogistic(60, 50, 0.5)
	assert.Greater(t, steep, gentle, "steeper curve must produce larger value above midpoint")
}

func TestEvaluateLogistic_OutputAlwaysInZeroOne(t *testing.T) {
	for x := -100.0; x <= 200.0; x += 10 {
		v := EvaluateLogistic(x, 50, 0.1)
		assert.GreaterOrEqual(t, v, 0.0)
		assert.LessOrEqual(t, v, 1.0)
	}
}

// ---------------------------------------------------------------------------
// NewTechTracker
// ---------------------------------------------------------------------------

func TestNewTechTracker_SetsInitialMaturities(t *testing.T) {
	tracker := NewTechTracker([]config.TechCurveDef{testCurve})
	assert.InDelta(t, 18.0, tracker.Maturity(config.TechOffshoreWind), 0.001)
}

func TestNewTechTracker_MultipleCurves_AllSeeded(t *testing.T) {
	curves := []config.TechCurveDef{
		testCurve,
		{ID: config.TechHeatPumps, InitialMaturity: 5.0},
	}
	tracker := NewTechTracker(curves)
	assert.InDelta(t, 18.0, tracker.Maturity(config.TechOffshoreWind), 0.001)
	assert.InDelta(t, 5.0, tracker.Maturity(config.TechHeatPumps), 0.001)
}

// ---------------------------------------------------------------------------
// Snapshot
// ---------------------------------------------------------------------------

func TestSnapshot_IndependentCopy(t *testing.T) {
	tracker := NewTechTracker([]config.TechCurveDef{testCurve})
	snap := Snapshot(tracker)
	tracker.Maturities[config.TechOffshoreWind] = 99.0
	assert.InDelta(t, 18.0, snap.Maturity(config.TechOffshoreWind), 0.001,
		"snapshot must not be affected by subsequent tracker mutations")
}

// ---------------------------------------------------------------------------
// AdvanceTick
// ---------------------------------------------------------------------------

func TestAdvanceTick_IncreasesMaturity(t *testing.T) {
	tracker := NewTechTracker([]config.TechCurveDef{testCurve})
	updated := AdvanceTick(tracker, testCurve, 0)
	assert.Greater(t, updated.Maturity(config.TechOffshoreWind), 18.0)
}

func TestAdvanceTick_WithBonus_IncreasesMoreThanWithout(t *testing.T) {
	tracker := NewTechTracker([]config.TechCurveDef{testCurve})
	withoutBonus := AdvanceTick(tracker, testCurve, 0).Maturity(config.TechOffshoreWind)
	withBonus := AdvanceTick(tracker, testCurve, 1.0).Maturity(config.TechOffshoreWind)
	assert.Greater(t, withBonus, withoutBonus)
}

func TestAdvanceTick_ClampsAt100(t *testing.T) {
	tracker := TechTracker{Maturities: map[config.Technology]float64{
		config.TechOffshoreWind: 99.9,
	}}
	updated := AdvanceTick(tracker, testCurve, 10.0)
	assert.Equal(t, 100.0, updated.Maturity(config.TechOffshoreWind))
}

func TestAdvanceTick_DoesNotMutateOriginal(t *testing.T) {
	tracker := NewTechTracker([]config.TechCurveDef{testCurve})
	before := tracker.Maturity(config.TechOffshoreWind)
	AdvanceTick(tracker, testCurve, 1.0)
	assert.Equal(t, before, tracker.Maturity(config.TechOffshoreWind),
		"AdvanceTick must not mutate the input tracker")
}

func TestAdvanceTick_OnlyAdvancesNamedTech(t *testing.T) {
	curves := []config.TechCurveDef{
		testCurve,
		{ID: config.TechHeatPumps, InitialMaturity: 5.0, BaseAdoptionRate: 0.04},
	}
	tracker := NewTechTracker(curves)
	updated := AdvanceTick(tracker, testCurve, 0)
	assert.Equal(t, 5.0, updated.Maturity(config.TechHeatPumps),
		"AdvanceTick must not touch technologies other than the named curve")
}

// ---------------------------------------------------------------------------
// ApplyAccelerationBonus
// ---------------------------------------------------------------------------

func TestApplyAccelerationBonus_AddsBonus(t *testing.T) {
	tracker := NewTechTracker([]config.TechCurveDef{testCurve})
	bonusMap := map[config.Technology]float64{config.TechOffshoreWind: 5.0}
	updated := ApplyAccelerationBonus(tracker, bonusMap)
	assert.InDelta(t, 23.0, updated.Maturity(config.TechOffshoreWind), 0.001)
}

func TestApplyAccelerationBonus_ClampsAt100(t *testing.T) {
	tracker := TechTracker{Maturities: map[config.Technology]float64{
		config.TechOffshoreWind: 98.0,
	}}
	updated := ApplyAccelerationBonus(tracker, map[config.Technology]float64{
		config.TechOffshoreWind: 50.0,
	})
	assert.Equal(t, 100.0, updated.Maturity(config.TechOffshoreWind))
}

func TestApplyAccelerationBonus_UnknownTech_IgnoredSilently(t *testing.T) {
	tracker := TechTracker{Maturities: map[config.Technology]float64{
		config.TechOffshoreWind: 18.0,
	}}
	updated := ApplyAccelerationBonus(tracker, map[config.Technology]float64{
		config.TechHydrogen: 10.0,
	})
	_, exists := updated.Maturities[config.TechHydrogen]
	assert.False(t, exists, "unknown technology must not be added to tracker")
	assert.InDelta(t, 18.0, updated.Maturity(config.TechOffshoreWind), 0.001)
}

func TestApplyAccelerationBonus_DoesNotMutateOriginal(t *testing.T) {
	tracker := NewTechTracker([]config.TechCurveDef{testCurve})
	ApplyAccelerationBonus(tracker, map[config.Technology]float64{
		config.TechOffshoreWind: 5.0,
	})
	assert.InDelta(t, 18.0, tracker.Maturity(config.TechOffshoreWind), 0.001)
}

// ---------------------------------------------------------------------------
// HeatPumpCOP
// ---------------------------------------------------------------------------

func TestHeatPumpCOP_ZeroMaturity_ReturnsMinCOP(t *testing.T) {
	assert.InDelta(t, 2.0, HeatPumpCOP(0), 0.001)
}

func TestHeatPumpCOP_FullMaturity_ReturnsMaxCOP(t *testing.T) {
	assert.InDelta(t, 3.5, HeatPumpCOP(100), 0.001)
}

func TestHeatPumpCOP_MonotoneIncreasing(t *testing.T) {
	prev := HeatPumpCOP(0)
	for m := 10.0; m <= 100.0; m += 10 {
		curr := HeatPumpCOP(m)
		assert.GreaterOrEqual(t, curr, prev,
			"HeatPumpCOP must be monotone increasing at maturity=%.0f", m)
		prev = curr
	}
}

func TestHeatPumpCOP_AboveRange_Clamps(t *testing.T) {
	assert.InDelta(t, 3.5, HeatPumpCOP(200), 0.001)
}
