package economy

import "twenty-fifty/internal/mathutil"

// EconomyState holds the hidden economy model state for one playthrough.
type EconomyState struct {
	Value        float64            // 0-100, hidden from player
	LobbyEffects map[string]float64 // dept ID -> multiplier; cleared each quarter
}

// TaxRevenue holds the result of a quarterly tax revenue computation.
type TaxRevenue struct {
	GBPBillions float64
	Quarter     int // 1-4
	Year        int
}

// BudgetAllocation holds the per-department discretionary spending for a quarter.
type BudgetAllocation struct {
	Departments map[string]float64 // dept ID -> GBP millions
	TotalGBPm   float64
}

// Calibration constants anchored to Green Book central estimates.
// Economy=50 (median) -> GBP 220bn/year -> GBP 55bn/quarter.
const (
	referenceAnnualRevenueGBPbn = 220.0
	discretionaryFraction       = 0.15 // 15% of tax revenue is discretionary
)

// NewEconomyState creates the starting economy for a new game.
// Value of 55 reflects Taitan's slightly above-median position in 2010.
func NewEconomyState() EconomyState {
	return EconomyState{
		Value:        55.0,
		LobbyEffects: make(map[string]float64),
	}
}

// TickEconomy applies weekly forces to the economy score and returns updated state.
// All delta inputs are on the hidden 0-100 economy scale.
//
//   - climateDamage:    positive value reduces economy (flood, heatwave costs)
//   - fuelPovertyDrag:  aggregate fuel poverty index dragging consumer spending
//   - shockSeverity:    energy shock severity (0-1) reducing economy
//   - policyBonus:      direct economy boost from active green investment policies
//   - fossilDrag:       penalty from high fossil dependency (stranded assets risk)
func TickEconomy(state EconomyState, climateDamage, fuelPovertyDrag, shockSeverity, policyBonus, fossilDrag float64) EconomyState {
	delta := policyBonus - climateDamage - fuelPovertyDrag*0.20 - shockSeverity*0.50 - fossilDrag*0.10
	state.Value = mathutil.Clamp(state.Value+delta, 0, 100)
	return state
}

// ComputeTaxRevenue maps the economy score to quarterly GBP bn tax revenue.
// Linear scaling: economy=0 -> 0, economy=50 -> 55bn/quarter, economy=100 -> 110bn/quarter.
func ComputeTaxRevenue(state EconomyState, quarter, year int) TaxRevenue {
	quarterlyRevenue := (state.Value / 50.0) * (referenceAnnualRevenueGBPbn / 4.0)
	return TaxRevenue{
		GBPBillions: mathutil.Clamp(quarterlyRevenue, 0, 9999),
		Quarter:     quarter,
		Year:        year,
	}
}

// AllocateBudget distributes discretionary spending across departments.
//
//   - baseFractions: dept -> base share fraction (should sum to 1.0)
//   - ministerPopWeights: dept -> governing minister's popularity (0-100); 50=neutral
//   - lcrModifier: budget multiplier from reputation package (0.85-1.15)
//   - lobbyEffects: dept -> lobby multiplier (use state.LobbyEffects or empty map)
//
// Shares are normalised so all departments together receive exactly TotalGBPm.
func AllocateBudget(revenue TaxRevenue, baseFractions, ministerPopWeights map[string]float64, lcrModifier float64, lobbyEffects map[string]float64) BudgetAllocation {
	totalGBPm := revenue.GBPBillions * 1000 * discretionaryFraction

	rawShares := make(map[string]float64, len(baseFractions))
	rawTotal := 0.0
	for dept, base := range baseFractions {
		popWeight := ministerPopWeights[dept]
		if popWeight <= 0 {
			popWeight = 50.0
		}
		lobby := lobbyEffects[dept]
		if lobby <= 0 {
			lobby = 1.0
		}
		raw := base * (popWeight / 50.0) * lcrModifier * lobby
		rawShares[dept] = raw
		rawTotal += raw
	}

	alloc := BudgetAllocation{
		Departments: make(map[string]float64, len(baseFractions)),
		TotalGBPm:   totalGBPm,
	}
	if rawTotal > 0 {
		for dept, raw := range rawShares {
			alloc.Departments[dept] = (raw / rawTotal) * totalGBPm
		}
	}
	return alloc
}

// AccumulateLobbyEffect multiplies a new lobby effect onto the current value
// for a department. Returns a new EconomyState; the input is not mutated.
func AccumulateLobbyEffect(state EconomyState, dept string, effect float64) EconomyState {
	newEffects := copyLobbyEffects(state.LobbyEffects)
	current := newEffects[dept]
	if current == 0 {
		current = 1.0
	}
	newEffects[dept] = current * effect
	state.LobbyEffects = newEffects
	return state
}

// ClearLobbyEffectsAtQuarterEnd resets all lobby effects to neutral.
// Returns a new EconomyState; the input is not mutated.
func ClearLobbyEffectsAtQuarterEnd(state EconomyState) EconomyState {
	state.LobbyEffects = make(map[string]float64)
	return state
}

// DeriveFossilDependency computes a 0-100 fossil dependency score from energy mix.
// gasFraction and oilFraction are each in [0, 1] (fraction of total energy demand).
func DeriveFossilDependency(gasFraction, oilFraction float64) float64 {
	return mathutil.Clamp((mathutil.Clamp(gasFraction, 0, 1)+mathutil.Clamp(oilFraction, 0, 1))*100, 0, 100)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func copyLobbyEffects(src map[string]float64) map[string]float64 {
	dst := make(map[string]float64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

