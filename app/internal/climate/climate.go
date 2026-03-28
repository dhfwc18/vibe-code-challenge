package climate

import (
	"math/rand"

	"twenty-fifty/internal/carbon"
	"twenty-fifty/internal/config"
)

// ClimateState combines the discrete severity level with a continuous position
// within that level's band, giving a smooth severity signal for event scaling.
type ClimateState struct {
	Level    carbon.ClimateLevel
	Severity float64 // 0-1 continuous position within the current level band
}

// ClimateEvent is a fired event instance created when an event def rolls in.
type ClimateEvent struct {
	DefID               string
	Name                string
	EventType           config.EventType
	Severity            config.EventSeverity
	Effects             config.EventEffect
	Week                int
	OffersShockResponse bool
}

// ShockResponseOption is the player's choice when a ShockResponseCard is queued.
type ShockResponseOption string

const (
	OptionAccept   ShockResponseOption = "ACCEPT"
	OptionDecline  ShockResponseOption = "DECLINE"
	OptionMitigate ShockResponseOption = "MITIGATE"
)

// ShockResponseCard is queued for the player after an event with OffersShockResponse.
type ShockResponseCard struct {
	ID                  string
	EventDefID          string
	BackfireProbability float64
}

// ResponseResult is the outcome of a player's shock response choice.
type ResponseResult struct {
	Option          ShockResponseOption
	Succeeded       bool
	Backfired       bool
	LCRDelta        float64
	PopularityDelta float64
}

// DeriveClimateState maps cumulative overshoot stock to a ClimateState.
// Severity is interpolated continuously within the current level band.
func DeriveClimateState(cumulativeStock float64) ClimateState {
	level := carbon.ClimateStockToLevel(cumulativeStock)
	return ClimateState{
		Level:    level,
		Severity: severityWithinLevel(cumulativeStock, level),
	}
}

// severityWithinLevel interpolates a 0-1 position within the current level's band.
func severityWithinLevel(stock float64, level carbon.ClimateLevel) float64 {
	var lo, hi float64
	switch level {
	case carbon.ClimateLevelStable:
		lo, hi = 0, carbon.ThresholdElevated
	case carbon.ClimateLevelElevated:
		lo, hi = carbon.ThresholdElevated, carbon.ThresholdCritical
	case carbon.ClimateLevelCritical:
		lo, hi = carbon.ThresholdCritical, carbon.ThresholdEmergency
	default: // Emergency: open-ended; use a 200-unit band for display purposes
		lo, hi = carbon.ThresholdEmergency, carbon.ThresholdEmergency+200
	}
	if hi <= lo {
		return 1.0
	}
	return clamp((stock-lo)/(hi-lo), 0, 1)
}

// RollClimateEvent rolls the event deck for one week.
// Returns the first event that fires based on adjusted probability, or nil.
//
// For each def:
//   - If climate level >= ELEVATED, probability is multiplied by ClimateMultiplier.
//   - If fossilDependency > 60, probability is multiplied by FossilMultiplier.
func RollClimateEvent(defs []config.EventDef, state ClimateState, fossilDependency float64, week int, rng *rand.Rand) *ClimateEvent {
	for _, def := range defs {
		prob := def.BaseProbability
		if state.Level >= carbon.ClimateLevelElevated {
			prob *= def.ClimateMultiplier
		}
		if fossilDependency > 60 {
			prob *= def.FossilMultiplier
		}
		if rng.Float64() < prob {
			return &ClimateEvent{
				DefID:               def.ID,
				Name:                def.Name,
				EventType:           def.EventType,
				Severity:            def.Severity,
				Effects:             def.BaseEffects,
				Week:                week,
				OffersShockResponse: def.OffersShockResponse,
			}
		}
	}
	return nil
}

// BackfireProbability returns the probability that a shock response backfires.
// Monotone-decreasing in both lcr and playerReputation (both 0-100).
// Base backfire rate of 35% is reduced proportionally by credibility and reputation.
func BackfireProbability(lcr, playerReputation float64) float64 {
	lcrFactor := 1.0 - clamp(lcr, 0, 100)/100.0
	repFactor := 1.0 - clamp(playerReputation, 0, 100)/100.0
	return clamp(0.35*lcrFactor*repFactor, 0, 1)
}

// ShockResponseOutcome resolves a player's choice on a ShockResponseCard.
// Accept: higher gain if succeeds, larger penalty if backfires.
// Mitigate: half backfire risk of Accept, smaller gain/loss.
// Decline: always safe; small LCR and popularity loss.
func ShockResponseOutcome(card ShockResponseCard, option ShockResponseOption, rng *rand.Rand) ResponseResult {
	result := ResponseResult{Option: option}
	switch option {
	case OptionDecline:
		result.Succeeded = true
		result.LCRDelta = -2.0
		result.PopularityDelta = -1.0
	case OptionAccept:
		if rng.Float64() < card.BackfireProbability {
			result.Backfired = true
			result.LCRDelta = -5.0
			result.PopularityDelta = -4.0
		} else {
			result.Succeeded = true
			result.LCRDelta = 4.0
			result.PopularityDelta = 2.0
		}
	case OptionMitigate:
		if rng.Float64() < card.BackfireProbability*0.5 {
			result.Backfired = true
			result.LCRDelta = -2.5
			result.PopularityDelta = -2.0
		} else {
			result.Succeeded = true
			result.LCRDelta = 2.0
			result.PopularityDelta = 1.0
		}
	}
	return result
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
