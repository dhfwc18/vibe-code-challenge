package carbon

import "twenty-fifty/internal/config"

// ClimateLevel represents the four severity states of the climate system.
// Transitions are driven by cumulative overshoot stock above the 2010 baseline.
type ClimateLevel int

const (
	ClimateLevelStable   ClimateLevel = iota // < ThresholdElevated
	ClimateLevelElevated                     // ThresholdElevated <= stock < ThresholdCritical
	ClimateLevelCritical                     // ThresholdCritical  <= stock < ThresholdEmergency
	ClimateLevelEmergency                    // stock >= ThresholdEmergency
)

// Climate level thresholds in MtCO2e cumulative above the 2010 baseline.
// These are design anchors; they will be tuned during the balance pass.
const (
	ThresholdElevated  = 200.0
	ThresholdCritical  = 400.0
	ThresholdEmergency = 600.0
)

// CarbonBudgetState holds all carbon accounting state for a single playthrough.
// Values are expressed in MtCO2e throughout.
type CarbonBudgetState struct {
	RunningAnnualTotal   float64 // emissions accumulated so far in the current calendar year
	CumulativeStock      float64 // total net emissions above the 2010 baseline since game start
	OvershootAccumulator float64 // total MtCO2e over CCC annual limits (never decreases)
	CurrentBudgetLimit   float64 // annual limit for the current year from the CCC table
}

// BudgetCheckResult is the outcome of a year-end carbon budget check.
type BudgetCheckResult struct {
	Year         int
	Limit        float64
	Actual       float64
	Overshoot    float64 // zero if under budget
	IsOverBudget bool
}

// AccumulateWeekly adds one week of net emissions to the running annual total
// and the cumulative stock. weeklyNetMtCO2e may be negative (net removal).
func AccumulateWeekly(state CarbonBudgetState, weeklyNetMtCO2e float64) CarbonBudgetState {
	state.RunningAnnualTotal += weeklyNetMtCO2e
	state.CumulativeStock += weeklyNetMtCO2e
	return state
}

// CheckAnnualBudget compares the running annual total against the CCC budget
// limit for the given year. It resets RunningAnnualTotal to zero, accumulates
// any overshoot, and sets CurrentBudgetLimit to the next year's limit.
// budgets must be sorted by year ascending (as returned by config.Load).
func CheckAnnualBudget(state CarbonBudgetState, year int, budgets []config.CarbonBudgetEntry) (BudgetCheckResult, CarbonBudgetState) {
	limit := limitForYear(year, budgets)
	actual := state.RunningAnnualTotal
	overshoot := actual - limit
	if overshoot < 0 {
		overshoot = 0
	}
	result := BudgetCheckResult{
		Year:         year,
		Limit:        limit,
		Actual:       actual,
		Overshoot:    overshoot,
		IsOverBudget: actual > limit,
	}
	state.OvershootAccumulator += overshoot
	state.RunningAnnualTotal = 0
	state.CurrentBudgetLimit = limitForYear(year+1, budgets)
	return result, state
}

// ProjectTrajectory returns the extrapolated annual emissions total if the
// current weekly pace continues for the rest of the year.
// weeksElapsedThisYear is clamped to [1, 52].
func ProjectTrajectory(state CarbonBudgetState, weeksElapsedThisYear int) float64 {
	if weeksElapsedThisYear < 1 {
		weeksElapsedThisYear = 1
	}
	if weeksElapsedThisYear > 52 {
		weeksElapsedThisYear = 52
	}
	weeklyRate := state.RunningAnnualTotal / float64(weeksElapsedThisYear)
	return weeklyRate * 52
}

// ClimateStockToLevel maps the cumulative overshoot stock to a ClimateLevel.
// The stock is the total MtCO2e above the CCC annual limits since game start.
func ClimateStockToLevel(cumulativeStock float64) ClimateLevel {
	switch {
	case cumulativeStock >= ThresholdEmergency:
		return ClimateLevelEmergency
	case cumulativeStock >= ThresholdCritical:
		return ClimateLevelCritical
	case cumulativeStock >= ThresholdElevated:
		return ClimateLevelElevated
	default:
		return ClimateLevelStable
	}
}

// LevelLabel returns the display string for a ClimateLevel.
func LevelLabel(level ClimateLevel) string {
	switch level {
	case ClimateLevelElevated:
		return "ELEVATED"
	case ClimateLevelCritical:
		return "CRITICAL"
	case ClimateLevelEmergency:
		return "EMERGENCY"
	default:
		return "STABLE"
	}
}

// limitForYear returns the annual MtCO2e limit for year by linear interpolation
// between the two nearest entries in budgets. budgets must be sorted ascending by Year.
// Returns the first entry's limit for years before the range, the last for years after.
func limitForYear(year int, budgets []config.CarbonBudgetEntry) float64 {
	if len(budgets) == 0 {
		return 0
	}
	if year <= budgets[0].Year {
		return budgets[0].AnnualLimitMtCO2e
	}
	last := budgets[len(budgets)-1]
	if year >= last.Year {
		return last.AnnualLimitMtCO2e
	}
	for i := 1; i < len(budgets); i++ {
		if year <= budgets[i].Year {
			prev := budgets[i-1]
			next := budgets[i]
			t := float64(year-prev.Year) / float64(next.Year-prev.Year)
			return prev.AnnualLimitMtCO2e + t*(next.AnnualLimitMtCO2e-prev.AnnualLimitMtCO2e)
		}
	}
	return 0
}
