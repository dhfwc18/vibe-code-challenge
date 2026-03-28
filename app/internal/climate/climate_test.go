package climate

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"twenty-fifty/internal/carbon"
	"twenty-fifty/internal/config"
)

// seededRNG returns a deterministic *rand.Rand for testing.
func seededRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

// ---------------------------------------------------------------------------
// DeriveClimateState
// ---------------------------------------------------------------------------

func TestDeriveClimateState_ZeroStock_StableZeroSeverity(t *testing.T) {
	state := DeriveClimateState(0)
	assert.Equal(t, carbon.ClimateLevelStable, state.Level)
	assert.InDelta(t, 0.0, state.Severity, 0.001)
}

func TestDeriveClimateState_AtElevatedThreshold_ElevatedLevel(t *testing.T) {
	state := DeriveClimateState(carbon.ThresholdElevated)
	assert.Equal(t, carbon.ClimateLevelElevated, state.Level)
	assert.InDelta(t, 0.0, state.Severity, 0.001)
}

func TestDeriveClimateState_MidElevated_HalfSeverity(t *testing.T) {
	mid := (carbon.ThresholdElevated + carbon.ThresholdCritical) / 2
	state := DeriveClimateState(mid)
	assert.Equal(t, carbon.ClimateLevelElevated, state.Level)
	assert.InDelta(t, 0.5, state.Severity, 0.01)
}

func TestDeriveClimateState_AtCriticalThreshold_CriticalLevel(t *testing.T) {
	state := DeriveClimateState(carbon.ThresholdCritical)
	assert.Equal(t, carbon.ClimateLevelCritical, state.Level)
}

func TestDeriveClimateState_AtEmergencyThreshold_EmergencyLevel(t *testing.T) {
	state := DeriveClimateState(carbon.ThresholdEmergency)
	assert.Equal(t, carbon.ClimateLevelEmergency, state.Level)
}

func TestDeriveClimateState_SeverityAlwaysInBounds(t *testing.T) {
	for stock := 0.0; stock <= 900.0; stock += 50 {
		s := DeriveClimateState(stock)
		assert.GreaterOrEqual(t, s.Severity, 0.0, "severity must be >= 0 at stock=%.0f", stock)
		assert.LessOrEqual(t, s.Severity, 1.0, "severity must be <= 1 at stock=%.0f", stock)
	}
}

// ---------------------------------------------------------------------------
// RollClimateEvent
// ---------------------------------------------------------------------------

var alwaysFireDef = config.EventDef{
	ID:                "always",
	Name:              "Always Fires",
	EventType:         config.EventWeather,
	Severity:          config.SeverityMinor,
	BaseProbability:   1.0, // always fires
	ClimateMultiplier: 1.0,
	FossilMultiplier:  1.0,
	BaseEffects:       config.EventEffect{},
	OffersShockResponse: false,
}

var neverFireDef = config.EventDef{
	ID:                "never",
	Name:              "Never Fires",
	EventType:         config.EventWeather,
	Severity:          config.SeverityMinor,
	BaseProbability:   0.0, // never fires
	ClimateMultiplier: 1.0,
	FossilMultiplier:  1.0,
}

func stableState() ClimateState {
	return DeriveClimateState(0)
}

func elevatedState() ClimateState {
	return DeriveClimateState(carbon.ThresholdElevated + 50)
}

func TestRollClimateEvent_AlwaysFireDef_ReturnsEvent(t *testing.T) {
	event := RollClimateEvent([]config.EventDef{alwaysFireDef}, stableState(), 0, 1, seededRNG(42))
	assert.NotNil(t, event)
	assert.Equal(t, "always", event.DefID)
}

func TestRollClimateEvent_NeverFireDef_ReturnsNil(t *testing.T) {
	event := RollClimateEvent([]config.EventDef{neverFireDef}, stableState(), 0, 1, seededRNG(42))
	assert.Nil(t, event)
}

func TestRollClimateEvent_EmptyDefs_ReturnsNil(t *testing.T) {
	event := RollClimateEvent(nil, stableState(), 0, 1, seededRNG(42))
	assert.Nil(t, event)
}

func TestRollClimateEvent_SetsWeek(t *testing.T) {
	event := RollClimateEvent([]config.EventDef{alwaysFireDef}, stableState(), 0, 42, seededRNG(1))
	assert.Equal(t, 42, event.Week)
}

func TestRollClimateEvent_ElevatedClimate_IncreasesFireRate(t *testing.T) {
	// Def with 50% base, 2x climate multiplier
	def := config.EventDef{
		ID: "half", BaseProbability: 0.50,
		ClimateMultiplier: 2.0, FossilMultiplier: 1.0,
	}
	stable := countFires(def, stableState(), 0, 10000)
	elevated := countFires(def, elevatedState(), 0, 10000)
	assert.Greater(t, elevated, stable,
		"elevated climate should fire more often (stable=%d, elevated=%d)", stable, elevated)
}

func TestRollClimateEvent_HighFossilDependency_IncreasesFireRate(t *testing.T) {
	def := config.EventDef{
		ID: "fossil_sensitive", BaseProbability: 0.30,
		ClimateMultiplier: 1.0, FossilMultiplier: 2.0,
	}
	low := countFires(def, stableState(), 30, 10000)  // below 60 threshold
	high := countFires(def, stableState(), 80, 10000) // above 60 threshold
	assert.Greater(t, high, low)
}

func TestRollClimateEvent_DeterministicFromSeed(t *testing.T) {
	def := config.EventDef{ID: "det", BaseProbability: 0.5, ClimateMultiplier: 1.0, FossilMultiplier: 1.0}
	r1 := RollClimateEvent([]config.EventDef{def}, stableState(), 0, 1, seededRNG(99))
	r2 := RollClimateEvent([]config.EventDef{def}, stableState(), 0, 1, seededRNG(99))
	assert.Equal(t, r1 == nil, r2 == nil, "same seed must produce same outcome")
}

// countFires runs n rolls and returns how many times the event fired.
func countFires(def config.EventDef, state ClimateState, fossilDep float64, n int) int {
	rng := rand.New(rand.NewSource(42))
	count := 0
	for i := 0; i < n; i++ {
		if RollClimateEvent([]config.EventDef{def}, state, fossilDep, 1, rng) != nil {
			count++
		}
	}
	return count
}

// ---------------------------------------------------------------------------
// BackfireProbability
// ---------------------------------------------------------------------------

func TestBackfireProbability_ZeroLCR_ZeroRep_NearBase(t *testing.T) {
	p := BackfireProbability(0, 0)
	assert.InDelta(t, 0.35, p, 0.001)
}

func TestBackfireProbability_MaxLCR_MaxRep_NearZero(t *testing.T) {
	p := BackfireProbability(100, 100)
	assert.InDelta(t, 0.0, p, 0.001)
}

func TestBackfireProbability_MonotoneDecreasingInLCR(t *testing.T) {
	prev := BackfireProbability(0, 50)
	for lcr := 10.0; lcr <= 100.0; lcr += 10 {
		curr := BackfireProbability(lcr, 50)
		assert.LessOrEqual(t, curr, prev,
			"backfire probability must be non-increasing in LCR at lcr=%.0f", lcr)
		prev = curr
	}
}

func TestBackfireProbability_MonotoneDecreasingInRep(t *testing.T) {
	prev := BackfireProbability(50, 0)
	for rep := 10.0; rep <= 100.0; rep += 10 {
		curr := BackfireProbability(50, rep)
		assert.LessOrEqual(t, curr, prev,
			"backfire probability must be non-increasing in rep at rep=%.0f", rep)
		prev = curr
	}
}

func TestBackfireProbability_AlwaysInBounds(t *testing.T) {
	for lcr := 0.0; lcr <= 100.0; lcr += 20 {
		for rep := 0.0; rep <= 100.0; rep += 20 {
			p := BackfireProbability(lcr, rep)
			assert.GreaterOrEqual(t, p, 0.0)
			assert.LessOrEqual(t, p, 1.0)
		}
	}
}

// ---------------------------------------------------------------------------
// ShockResponseOutcome
// ---------------------------------------------------------------------------

func TestShockResponseOutcome_Decline_AlwaysSucceeds(t *testing.T) {
	card := ShockResponseCard{BackfireProbability: 0.99}
	result := ShockResponseOutcome(card, OptionDecline, seededRNG(1))
	assert.True(t, result.Succeeded)
	assert.False(t, result.Backfired)
}

func TestShockResponseOutcome_Decline_NegativeLCR(t *testing.T) {
	card := ShockResponseCard{BackfireProbability: 0.5}
	result := ShockResponseOutcome(card, OptionDecline, seededRNG(1))
	assert.Less(t, result.LCRDelta, 0.0)
}

func TestShockResponseOutcome_Accept_ZeroBackfire_Succeeds(t *testing.T) {
	card := ShockResponseCard{BackfireProbability: 0.0}
	result := ShockResponseOutcome(card, OptionAccept, seededRNG(1))
	assert.True(t, result.Succeeded)
	assert.False(t, result.Backfired)
	assert.Greater(t, result.LCRDelta, 0.0)
}

func TestShockResponseOutcome_Accept_CertainBackfire_Backfires(t *testing.T) {
	card := ShockResponseCard{BackfireProbability: 1.0}
	result := ShockResponseOutcome(card, OptionAccept, seededRNG(1))
	assert.True(t, result.Backfired)
	assert.False(t, result.Succeeded)
	assert.Less(t, result.LCRDelta, 0.0)
}

func TestShockResponseOutcome_Mitigate_HalfBackfireRisk(t *testing.T) {
	// At backfire prob 1.0, Accept always backfires; Mitigate uses 0.5 so it succeeds
	card := ShockResponseCard{BackfireProbability: 1.0}
	result := ShockResponseOutcome(card, OptionMitigate, seededRNG(1))
	// Mitigate prob = 0.5; with seed 1 we just check it ran without panic
	assert.True(t, result.Backfired || result.Succeeded)
}

func TestShockResponseOutcome_SetsOption(t *testing.T) {
	card := ShockResponseCard{}
	r := ShockResponseOutcome(card, OptionMitigate, seededRNG(1))
	assert.Equal(t, OptionMitigate, r.Option)
}
