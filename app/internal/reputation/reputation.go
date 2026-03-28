package reputation

import (
	"math"
	"math/rand"

	"twenty-fifty/internal/mathutil"
)

// LowCarbonReputation tracks the player's green credibility.
// Value is the true hidden score; LastPollResult is what the player sees.
type LowCarbonReputation struct {
	Value          float64 // 0-100, true value (hidden)
	LastPollResult float64 // most recently polled value (player-visible)
}

// Constants for the reputation model.
const (
	pollNoiseSigma     = 4.0  // standard deviation of poll noise
	naturalDecayTarget = 50.0 // LCR drifts toward this when no policies are active
	naturalDecayRate   = 0.005 // weekly drift fraction when policy gain is small
	minPolicyGain      = 0.05  // policy gain below this triggers natural decay

	// Budget modifier: LCR=0 -> 0.85, LCR=50 -> 1.0, LCR=100 -> 1.15
	budgetModifierMid   = 1.0
	budgetModifierRange = 0.15
)

// NewLCR creates a starting LowCarbonReputation for a new game.
// Starting at 35 reflects low initial credibility before the player acts.
func NewLCR() LowCarbonReputation {
	return LowCarbonReputation{Value: 35.0, LastPollResult: 35.0}
}

// TickReputation updates LCR for one week. Returns a new value; input not mutated.
//
//   - weeklyPolicyCarbonDelta: MtCO2e reduction from active policies (positive = good)
//   - eventImpact: direct LCR delta from events (positive or negative)
//
// When policy gain is negligible, LCR drifts slowly back toward 50.
func TickReputation(lcr LowCarbonReputation, weeklyPolicyCarbonDelta, eventImpact float64) LowCarbonReputation {
	policyGain := weeklyPolicyCarbonDelta * 5.5
	decay := 0.0
	if policyGain < minPolicyGain {
		decay = (lcr.Value - naturalDecayTarget) * naturalDecayRate
	}
	lcr.Value = mathutil.Clamp(lcr.Value+policyGain+eventImpact-decay, 0, 100)
	return lcr
}

// ChainToMinisterPopularity returns the weekly popularity delta for the governing
// minister caused by the given LCR delta.
func ChainToMinisterPopularity(lcrDelta float64) float64 {
	return lcrDelta * 0.25
}

// ChainToGovtPopularity returns the weekly government popularity delta caused by
// an LCR change. Larger magnitude than the minister chain.
func ChainToGovtPopularity(lcrDelta float64) float64 {
	return lcrDelta * 0.40
}

// ChainToBudgetModifier maps the current LCR value to a budget multiplier.
//
//	LCR=0   -> 0.85
//	LCR=50  -> 1.00 (neutral)
//	LCR=100 -> 1.15
func ChainToBudgetModifier(lcr float64) float64 {
	normalised := (mathutil.Clamp(lcr, 0, 100) - 50.0) / 50.0 // -1 to +1
	return budgetModifierMid + normalised*budgetModifierRange
}

// PollLCR samples the true LCR value with Gaussian noise (sigma=4).
// Returns an updated LowCarbonReputation with LastPollResult set.
func PollLCR(lcr LowCarbonReputation, rng *rand.Rand) LowCarbonReputation {
	noise := rng.NormFloat64() * pollNoiseSigma
	lcr.LastPollResult = mathutil.Clamp(lcr.Value+noise, 0, 100)
	return lcr
}

// CapitalisationProbability returns the probability that a capitalisation attempt
// succeeds, as a function of LCR and playerReputation (both 0-100).
// Uses a sigmoid centred so lcr=60 + playerRep=60 gives p~0.5.
// Monotone-increasing in both inputs.
func CapitalisationProbability(lcr, playerReputation float64) float64 {
	combined := mathutil.Clamp(lcr, 0, 100) + mathutil.Clamp(playerReputation, 0, 100)
	// sigmoid: p = 1 / (1 + exp(-(combined - 120) / 20))
	return 1.0 / (1.0 + math.Exp(-(combined-120.0)/20.0))
}

// CapitalisationSuccess returns whether the player capitalises on current LCR
// and reputation. Probability is monotone-increasing in both inputs.
func CapitalisationSuccess(lcr, playerReputation float64, rng *rand.Rand) bool {
	return rng.Float64() < CapitalisationProbability(lcr, playerReputation)
}

