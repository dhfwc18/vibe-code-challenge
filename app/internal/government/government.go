package government

import (
	"twenty-fifty/internal/config"
	"twenty-fifty/internal/mathutil"
	"twenty-fifty/internal/stakeholder"
)

// GovernmentState tracks which party holds power, cabinet composition,
// and the electoral cycle.
type GovernmentState struct {
	RulingParty     config.Party
	CabinetByRole   map[config.Role]string // role -> stakeholder ID; absent key = vacant
	ElectionDueWeek int
	TermNumber      int
}

// MinisterStats is the derived political-economy profile of a minister in post.
// PopularityModifier is an additive weekly delta to government popularity.
// BudgetAllocationBias maps department IDs to multipliers used in
// economy.AllocateBudget's lobbyEffects argument.
type MinisterStats struct {
	StakeholderID        string
	Role                 config.Role
	PopularityModifier   float64
	BudgetAllocationBias map[string]float64
}

// Department IDs used as keys in BudgetAllocationBias and economy.AllocateBudget.
const (
	DeptPower     = "power"
	DeptTransport = "transport"
	DeptBuildings = "buildings"
	DeptIndustry  = "industry"
	DeptCross     = "cross"
)

// Calibration constants for minister stat derivation.
const (
	// lcrPopularityFactor converts an LCR delta into a weekly popularity modifier.
	lcrPopularityFactor = 0.30

	// Popularity bonus/penalty thresholds based on net-zero sympathy.
	nzsHighThreshold = 60.0
	nzsLowThreshold  = 30.0
	nzsBonus         = 0.10
	nzsPenalty       = 0.10

	// budgetBiasStrength maps ideology units to a bias multiplier delta.
	budgetBiasStrength = 0.015

	minBiasMultiplier = 0.50
	maxBiasMultiplier = 2.00
	neutralBias       = 1.00
)

// NewGovernment returns a GovernmentState with an empty cabinet and TermNumber=1.
func NewGovernment(rulingParty config.Party, electionDueWeek int) GovernmentState {
	return GovernmentState{
		RulingParty:     rulingParty,
		CabinetByRole:   make(map[config.Role]string),
		ElectionDueWeek: electionDueWeek,
		TermNumber:      1,
	}
}

// AssignMinister returns a new GovernmentState with role assigned to stakeholderID.
// Any existing occupant of role is silently replaced.
func AssignMinister(g GovernmentState, role config.Role, stakeholderID string) GovernmentState {
	m := copyCabinet(g.CabinetByRole)
	m[role] = stakeholderID
	g.CabinetByRole = m
	return g
}

// RemoveMinister returns a new GovernmentState with role vacated.
// If role is not in the cabinet, the state is returned unchanged.
func RemoveMinister(g GovernmentState, role config.Role) GovernmentState {
	if _, ok := g.CabinetByRole[role]; !ok {
		return g
	}
	m := copyCabinet(g.CabinetByRole)
	delete(m, role)
	g.CabinetByRole = m
	return g
}

// IsElectionDue returns true when currentWeek has reached or passed the due week.
func IsElectionDue(g GovernmentState, currentWeek int) bool {
	return currentWeek >= g.ElectionDueWeek
}

// ComputeMinisterStats derives the popularity modifier and budget allocation
// bias for a minister based on their stakeholder attributes and the current
// weekly LCR delta.
func ComputeMinisterStats(s stakeholder.Stakeholder, lcrDelta float64) MinisterStats {
	// Popularity modifier from LCR movement.
	pop := lcrDelta * lcrPopularityFactor
	if s.NetZeroSympathy > nzsHighThreshold {
		pop += nzsBonus
	} else if s.NetZeroSympathy < nzsLowThreshold {
		pop -= nzsPenalty
	}

	// Budget allocation bias per department.
	// Right ideology (positive) -> more industry, less buildings/transport.
	// Left ideology (negative) -> more buildings/transport, less industry.
	// NetZeroSympathy drives power and cross department bias.
	ideo := s.IdeologyScore
	nzs := s.NetZeroSympathy

	powerBias := mathutil.Clamp(neutralBias+(nzs-50.0)*budgetBiasStrength, minBiasMultiplier, maxBiasMultiplier)
	transportBias := mathutil.Clamp(neutralBias-ideo*budgetBiasStrength, minBiasMultiplier, maxBiasMultiplier)
	buildingsBias := mathutil.Clamp(neutralBias-ideo*budgetBiasStrength*0.8, minBiasMultiplier, maxBiasMultiplier)
	industryBias := mathutil.Clamp(neutralBias+ideo*budgetBiasStrength, minBiasMultiplier, maxBiasMultiplier)
	crossBias := mathutil.Clamp(neutralBias+(nzs-50.0)*budgetBiasStrength*0.5, minBiasMultiplier, maxBiasMultiplier)

	return MinisterStats{
		StakeholderID:      s.ID,
		Role:               s.Role,
		PopularityModifier: pop,
		BudgetAllocationBias: map[string]float64{
			DeptPower:     powerBias,
			DeptTransport: transportBias,
			DeptBuildings: buildingsBias,
			DeptIndustry:  industryBias,
			DeptCross:     crossBias,
		},
	}
}

// TriggerElection returns a new GovernmentState reflecting the election outcome.
// The cabinet is cleared; ministers must be reassigned after an election.
func TriggerElection(g GovernmentState, winner config.Party, newElectionDueWeek int) GovernmentState {
	return GovernmentState{
		RulingParty:     winner,
		CabinetByRole:   make(map[config.Role]string),
		ElectionDueWeek: newElectionDueWeek,
		TermNumber:      g.TermNumber + 1,
	}
}

// copyCabinet returns a shallow copy of a cabinet map.
func copyCabinet(src map[config.Role]string) map[config.Role]string {
	dst := make(map[config.Role]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
