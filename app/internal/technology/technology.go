package technology

import (
	"math"

	"twenty-fifty/internal/config"
	"twenty-fifty/internal/mathutil"
)

// TechTracker holds the live maturity value for each of the 8 tracked
// decarbonisation technologies. Maturity is in [0, 100]; 100 = fully deployed.
// TechTracker is owned by the simulation; all functions return new values
// rather than mutating in place.
type TechTracker struct {
	Maturities map[config.Technology]float64
}

// TechSnapshot is an immutable read-only copy of all maturities at one point
// in time. Other packages receive snapshots; only the simulation holds a
// TechTracker and may advance it.
type TechSnapshot struct {
	Maturities map[config.Technology]float64
}

// NewTechTracker initialises a TechTracker from the technology curve definitions,
// seeding each technology at its configured InitialMaturity.
func NewTechTracker(curves []config.TechCurveDef) TechTracker {
	m := make(map[config.Technology]float64, len(curves))
	for _, c := range curves {
		m[c.ID] = c.InitialMaturity
	}
	return TechTracker{Maturities: m}
}

// Snapshot returns an immutable copy of the tracker safe to pass to other packages.
func Snapshot(tracker TechTracker) TechSnapshot {
	m := make(map[config.Technology]float64, len(tracker.Maturities))
	for k, v := range tracker.Maturities {
		m[k] = v
	}
	return TechSnapshot{Maturities: m}
}

// EvaluateLogistic computes the standard logistic (sigmoid) function at x.
// Returns a value in (0, 1). Output = 0.5 when x == midpoint.
// steepness controls how sharply the curve rises around the midpoint.
func EvaluateLogistic(x, midpoint, steepness float64) float64 {
	return 1.0 / (1.0 + math.Exp(-steepness*(x-midpoint)))
}

// AdvanceTick advances one technology's maturity by one game week.
//
// The weekly gain is shaped by a logistic function in maturity-space so that
// adoption is slow at low and high maturities and fastest in the middle.
// accelerationBonus is an additive weekly bonus from active policies (>= 0).
//
// The returned TechTracker is a new value; the input is not mutated.
func AdvanceTick(tracker TechTracker, curve config.TechCurveDef, accelerationBonus float64) TechTracker {
	current := tracker.Maturities[curve.ID] // zero if not present
	// logistic shaping in maturity-space: peak gain at maturity=50
	shape := EvaluateLogistic(current, 50.0, 0.08)
	weeklyGain := curve.BaseAdoptionRate*shape + accelerationBonus
	result := copyTracker(tracker)
	result.Maturities[curve.ID] = mathutil.Clamp(current+weeklyGain, 0, 100)
	return result
}

// ApplyAccelerationBonus adds per-technology bonuses from active policies directly
// to the tracker maturities. bonusMap maps Technology IDs to their additional
// weekly maturity gain. Technologies not present in the tracker are skipped.
//
// The returned TechTracker is a new value; the input is not mutated.
func ApplyAccelerationBonus(tracker TechTracker, bonusMap map[config.Technology]float64) TechTracker {
	result := copyTracker(tracker)
	for tech, bonus := range bonusMap {
		if current, ok := result.Maturities[tech]; ok {
			result.Maturities[tech] = mathutil.Clamp(current+bonus, 0, 100)
		}
	}
	return result
}

// HeatPumpCOP returns the coefficient of performance for heat pumps at the
// given technology maturity. COP ranges from 2.0 (zero maturity) to 3.5
// (fully mature). Used by the region package's fuel poverty formula.
func HeatPumpCOP(maturity float64) float64 {
	return 2.0 + (mathutil.Clamp(maturity, 0, 100)/100.0)*1.5
}

// Maturity returns the current maturity for a specific technology.
// Returns 0 if the technology is not present in the tracker.
func (t TechTracker) Maturity(tech config.Technology) float64 {
	return t.Maturities[tech]
}

// Maturity returns the maturity for a specific technology in this snapshot.
// Returns 0 if the technology is not present in the snapshot.
func (s TechSnapshot) Maturity(tech config.Technology) float64 {
	return s.Maturities[tech]
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------


func copyTracker(t TechTracker) TechTracker {
	m := make(map[config.Technology]float64, len(t.Maturities))
	for k, v := range t.Maturities {
		m[k] = v
	}
	return TechTracker{Maturities: m}
}
