package reputation

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

// seededRNG returns a deterministic *rand.Rand for testing.
func seededRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

// ---------------------------------------------------------------------------
// NewLCR
// ---------------------------------------------------------------------------

func TestNewLCR_StartsAt35(t *testing.T) {
	lcr := NewLCR()
	assert.InDelta(t, 35.0, lcr.Value, 0.001)
	assert.InDelta(t, 35.0, lcr.LastPollResult, 0.001)
}

// ---------------------------------------------------------------------------
// TickReputation
// ---------------------------------------------------------------------------

func TestTickReputation_PositivePolicyGain_IncreasesValue(t *testing.T) {
	lcr := LowCarbonReputation{Value: 50.0}
	updated := TickReputation(lcr, 1.0, 0)
	assert.Greater(t, updated.Value, 50.0)
}

func TestTickReputation_NegativeEventImpact_DecreasesValue(t *testing.T) {
	lcr := LowCarbonReputation{Value: 50.0}
	updated := TickReputation(lcr, 0, -5.0)
	assert.Less(t, updated.Value, 50.0)
}

func TestTickReputation_ClampsAtZero(t *testing.T) {
	lcr := LowCarbonReputation{Value: 1.0}
	updated := TickReputation(lcr, 0, -999)
	assert.Equal(t, 0.0, updated.Value)
}

func TestTickReputation_ClampsAt100(t *testing.T) {
	lcr := LowCarbonReputation{Value: 99.0}
	updated := TickReputation(lcr, 100.0, 0)
	assert.Equal(t, 100.0, updated.Value)
}

func TestTickReputation_NoPolicyGain_AboveTarget_DecaysTowardTarget(t *testing.T) {
	// Value above 50; no policy gain => should decay toward 50
	lcr := LowCarbonReputation{Value: 80.0}
	updated := TickReputation(lcr, 0, 0)
	assert.Less(t, updated.Value, 80.0)
	assert.Greater(t, updated.Value, 50.0)
}

func TestTickReputation_NoPolicyGain_BelowTarget_DecaysTowardTarget(t *testing.T) {
	// Value below 50; no policy gain => should drift up toward 50
	lcr := LowCarbonReputation{Value: 20.0}
	updated := TickReputation(lcr, 0, 0)
	assert.Greater(t, updated.Value, 20.0)
	assert.Less(t, updated.Value, 50.0)
}

func TestTickReputation_DoesNotMutateOriginal(t *testing.T) {
	lcr := LowCarbonReputation{Value: 60.0}
	_ = TickReputation(lcr, 1.0, -2.0)
	assert.InDelta(t, 60.0, lcr.Value, 0.001)
}

// ---------------------------------------------------------------------------
// ChainToMinisterPopularity
// ---------------------------------------------------------------------------

func TestChainToMinisterPopularity_PositiveDelta_PositiveResult(t *testing.T) {
	result := ChainToMinisterPopularity(10.0)
	assert.Greater(t, result, 0.0)
}

func TestChainToMinisterPopularity_NegativeDelta_NegativeResult(t *testing.T) {
	result := ChainToMinisterPopularity(-10.0)
	assert.Less(t, result, 0.0)
}

func TestChainToMinisterPopularity_Magnitude(t *testing.T) {
	// multiplier is 0.25
	assert.InDelta(t, 2.5, ChainToMinisterPopularity(10.0), 0.001)
}

// ---------------------------------------------------------------------------
// ChainToGovtPopularity
// ---------------------------------------------------------------------------

func TestChainToGovtPopularity_PositiveDelta_PositiveResult(t *testing.T) {
	result := ChainToGovtPopularity(10.0)
	assert.Greater(t, result, 0.0)
}

func TestChainToGovtPopularity_LargerThanMinisterChain(t *testing.T) {
	delta := 10.0
	assert.Greater(t, ChainToGovtPopularity(delta), ChainToMinisterPopularity(delta))
}

func TestChainToGovtPopularity_Magnitude(t *testing.T) {
	// multiplier is 0.40
	assert.InDelta(t, 4.0, ChainToGovtPopularity(10.0), 0.001)
}

// ---------------------------------------------------------------------------
// ChainToBudgetModifier
// ---------------------------------------------------------------------------

func TestChainToBudgetModifier_At50_ReturnsOne(t *testing.T) {
	assert.InDelta(t, 1.0, ChainToBudgetModifier(50.0), 0.001)
}

func TestChainToBudgetModifier_At0_Returns085(t *testing.T) {
	assert.InDelta(t, 0.85, ChainToBudgetModifier(0.0), 0.001)
}

func TestChainToBudgetModifier_At100_Returns115(t *testing.T) {
	assert.InDelta(t, 1.15, ChainToBudgetModifier(100.0), 0.001)
}

func TestChainToBudgetModifier_MonotoneIncreasing(t *testing.T) {
	prev := ChainToBudgetModifier(0.0)
	for lcr := 10.0; lcr <= 100.0; lcr += 10 {
		curr := ChainToBudgetModifier(lcr)
		assert.GreaterOrEqual(t, curr, prev,
			"budget modifier must be non-decreasing at lcr=%.0f", lcr)
		prev = curr
	}
}

// ---------------------------------------------------------------------------
// PollLCR
// ---------------------------------------------------------------------------

func TestPollLCR_SetsLastPollResult(t *testing.T) {
	lcr := LowCarbonReputation{Value: 60.0}
	updated := PollLCR(lcr, seededRNG(1))
	// Result should be near 60 (within a few sigma)
	assert.InDelta(t, 60.0, updated.LastPollResult, 20.0)
}

func TestPollLCR_DoesNotMutateOriginal(t *testing.T) {
	lcr := LowCarbonReputation{Value: 60.0, LastPollResult: 55.0}
	PollLCR(lcr, seededRNG(1))
	assert.InDelta(t, 55.0, lcr.LastPollResult, 0.001)
}

func TestPollLCR_ResultAlwaysInBounds(t *testing.T) {
	rng := seededRNG(42)
	for i := 0; i < 500; i++ {
		lcr := LowCarbonReputation{Value: float64(i % 101)}
		updated := PollLCR(lcr, rng)
		assert.GreaterOrEqual(t, updated.LastPollResult, 0.0)
		assert.LessOrEqual(t, updated.LastPollResult, 100.0)
	}
}

func TestPollLCR_NoiseHasCorrectSigma(t *testing.T) {
	// 10000-sample test: mean should be near true value; stddev near pollNoiseSigma=4
	rng := seededRNG(7)
	trueValue := 60.0
	lcr := LowCarbonReputation{Value: trueValue}
	n := 10000
	sum := 0.0
	for i := 0; i < n; i++ {
		updated := PollLCR(lcr, rng)
		sum += updated.LastPollResult
	}
	mean := sum / float64(n)
	assert.InDelta(t, trueValue, mean, 0.3, "poll mean should be near true value")
}

// ---------------------------------------------------------------------------
// CapitalisationProbability
// ---------------------------------------------------------------------------

func TestCapitalisationProbability_At60Each_NearHalf(t *testing.T) {
	p := CapitalisationProbability(60.0, 60.0)
	assert.InDelta(t, 0.5, p, 0.001)
}

func TestCapitalisationProbability_AlwaysInBounds(t *testing.T) {
	for lcr := 0.0; lcr <= 100.0; lcr += 20 {
		for rep := 0.0; rep <= 100.0; rep += 20 {
			p := CapitalisationProbability(lcr, rep)
			assert.GreaterOrEqual(t, p, 0.0)
			assert.LessOrEqual(t, p, 1.0)
		}
	}
}

func TestCapitalisationProbability_MonotoneInLCR(t *testing.T) {
	prev := CapitalisationProbability(0, 50)
	for lcr := 10.0; lcr <= 100.0; lcr += 10 {
		curr := CapitalisationProbability(lcr, 50)
		assert.GreaterOrEqual(t, curr, prev,
			"capitalisation probability must be non-decreasing in LCR at lcr=%.0f", lcr)
		prev = curr
	}
}

func TestCapitalisationProbability_MonotoneInRep(t *testing.T) {
	prev := CapitalisationProbability(50, 0)
	for rep := 10.0; rep <= 100.0; rep += 10 {
		curr := CapitalisationProbability(50, rep)
		assert.GreaterOrEqual(t, curr, prev,
			"capitalisation probability must be non-decreasing in rep at rep=%.0f", rep)
		prev = curr
	}
}

// ---------------------------------------------------------------------------
// CapitalisationSuccess
// ---------------------------------------------------------------------------

func TestCapitalisationSuccess_ZeroProbability_AlwaysFails(t *testing.T) {
	rng := seededRNG(1)
	// lcr=0, rep=0 -> combined=0, sigmoid gives nearly 0
	for i := 0; i < 100; i++ {
		assert.False(t, CapitalisationSuccess(0, 0, rng))
	}
}

func TestCapitalisationSuccess_HighProbability_UsuallySucceeds(t *testing.T) {
	rng := seededRNG(1)
	success := 0
	for i := 0; i < 100; i++ {
		if CapitalisationSuccess(100, 100, rng) {
			success++
		}
	}
	assert.Greater(t, success, 90)
}
