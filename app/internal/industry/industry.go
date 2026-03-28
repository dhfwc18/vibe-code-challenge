package industry

import (
	"sort"

	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/mathutil"
	"github.com/vibe-code-challenge/twenty-fifty/internal/technology"
)

// CompanyStatus represents the lifecycle state of an LCT company.
// Transition logic is deferred to the simulation layer (Layer 5).
type CompanyStatus string

const (
	CompanyStatusInactive   CompanyStatus = "INACTIVE"
	CompanyStatusActive     CompanyStatus = "ACTIVE"
	CompanyStatusStruggling CompanyStatus = "STRUGGLING"
	CompanyStatusBankrupt   CompanyStatus = "BANKRUPT"
	CompanyStatusAbsorbed   CompanyStatus = "ABSORBED"
	CompanyStatusStartup    CompanyStatus = "STARTUP"
)

// CompanyState holds the runtime state for one LCT company during a playthrough.
type CompanyState struct {
	DefID              string
	IsActive           bool
	Status             CompanyStatus
	ContractedTech     config.Technology // empty string when inactive
	WeeksOnContract    int
	AccumulatedQuality float64 // resets to zero on DeliverTech
	WorkRate           float64 // 0-100; converges toward def.BaseWorkRate each tick
}

// IndustryState holds all company states for a playthrough.
type IndustryState struct {
	Companies map[string]CompanyState // keyed by CompanyState.DefID
}

// Calibration constants for the industry model.
const (
	// qualityGainPerWeek is the base fraction of BaseQuality added to
	// AccumulatedQuality each week at WorkRate=100 and full capacity.
	qualityGainPerWeek = 0.80

	// workRateDecayRate is the weekly fraction by which WorkRate converges
	// toward the company's BaseWorkRate.
	workRateDecayRate = 0.05

	// capacityDampening controls the minimum effective work rate fraction
	// when InstallerCapacity is zero. 0.70 means a company at zero capacity
	// still operates at 70% of its work rate.
	capacityDampening = 0.70

	// deliveryMaturityBoostBase is the tech maturity points awarded per 100
	// units of AccumulatedQuality on delivery.
	deliveryMaturityBoostBase = 0.50

	// maxDeliveryBoost caps the maturity boost from a single delivery call.
	maxDeliveryBoost = 8.0

	// referenceInstallerCapacity matches region.referenceInstallerCapacity
	// so that capacity fraction is consistent across packages.
	referenceInstallerCapacity = 50.0
)

// SeedIndustry creates an IndustryState with all companies inactive from config defs.
func SeedIndustry(defs []config.CompanyDef) IndustryState {
	m := make(map[string]CompanyState, len(defs))
	for _, d := range defs {
		m[d.ID] = CompanyState{
			DefID:    d.ID,
			IsActive: false,
			Status:   CompanyStatusInactive,
		}
	}
	return IndustryState{Companies: m}
}

// ActivateCompany returns a new IndustryState with the named company set to active,
// contracted to tech, with its WorkRate initialised to baseWorkRate.
// If defID is not found in state, state is returned unchanged.
func ActivateCompany(
	state IndustryState,
	defID string,
	tech config.Technology,
	baseWorkRate float64,
) IndustryState {
	if _, ok := state.Companies[defID]; !ok {
		return state
	}
	m := copyCompanies(state.Companies)
	m[defID] = CompanyState{
		DefID:              defID,
		IsActive:           true,
		Status:             CompanyStatusActive,
		ContractedTech:     tech,
		WeeksOnContract:    0,
		AccumulatedQuality: 0,
		WorkRate:           mathutil.Clamp(baseWorkRate, 0, 100),
	}
	return IndustryState{Companies: m}
}

// DeactivateCompany returns a new IndustryState with the named company cleared
// to an inactive zero state. If defID is not found, state is returned unchanged.
func DeactivateCompany(state IndustryState, defID string) IndustryState {
	if _, ok := state.Companies[defID]; !ok {
		return state
	}
	m := copyCompanies(state.Companies)
	m[defID] = CompanyState{DefID: defID, IsActive: false, Status: CompanyStatusInactive}
	return IndustryState{Companies: m}
}

// TickCompany advances one active company by one week. It:
//   - Accumulates quality proportional to work rate and installer capacity.
//   - Drifts work rate toward def.BaseWorkRate.
//   - Increments WeeksOnContract.
//
// Does nothing if the company is not active or is not found.
func TickCompany(
	state IndustryState,
	defID string,
	def config.CompanyDef,
	installerCapacity float64,
) IndustryState {
	cs, ok := state.Companies[defID]
	if !ok || !cs.IsActive {
		return state
	}

	// Capacity fraction: 0 = no capacity, 1 = full capacity.
	capFrac := mathutil.Clamp(installerCapacity/referenceInstallerCapacity, 0, 1)

	// Effective work rate: dampened so even zero capacity gives partial output.
	effectiveRate := cs.WorkRate * (capacityDampening + (1-capacityDampening)*capFrac)

	// Accumulate quality.
	weeklyQuality := def.BaseQuality * (effectiveRate / 100.0) * qualityGainPerWeek
	cs.AccumulatedQuality += weeklyQuality

	// Drift work rate toward base.
	cs.WorkRate += (def.BaseWorkRate - cs.WorkRate) * workRateDecayRate
	cs.WorkRate = mathutil.Clamp(cs.WorkRate, 0, 100)

	cs.WeeksOnContract++

	m := copyCompanies(state.Companies)
	m[defID] = cs
	return IndustryState{Companies: m}
}

// DeliverTech converts the company's AccumulatedQuality into a technology maturity
// boost and resets the accumulator. Returns updated copies of both state and tracker.
// Does nothing if the company is not active or is not found.
func DeliverTech(
	state IndustryState,
	defID string,
	tracker technology.TechTracker,
	def config.CompanyDef,
) (IndustryState, technology.TechTracker) {
	cs, ok := state.Companies[defID]
	if !ok || !cs.IsActive {
		return state, tracker
	}

	boost := mathutil.Clamp(
		cs.AccumulatedQuality*deliveryMaturityBoostBase/100.0,
		0,
		maxDeliveryBoost,
	)

	bonusMap := map[config.Technology]float64{cs.ContractedTech: boost}
	newTracker := technology.ApplyAccelerationBonus(tracker, bonusMap)

	cs.AccumulatedQuality = 0
	m := copyCompanies(state.Companies)
	m[defID] = cs
	return IndustryState{Companies: m}, newTracker
}

// ActiveCompaniesForTech returns a sorted slice of defIDs for all companies
// currently contracted to tech. Returns a non-nil empty slice when none match.
func ActiveCompaniesForTech(state IndustryState, tech config.Technology) []string {
	ids := []string{}
	for defID, cs := range state.Companies {
		if cs.IsActive && cs.ContractedTech == tech {
			ids = append(ids, defID)
		}
	}
	sort.Strings(ids)
	return ids
}

// copyCompanies returns a shallow copy of the companies map.
// CompanyState contains no pointer fields so a shallow copy is a full copy.
func copyCompanies(src map[string]CompanyState) map[string]CompanyState {
	dst := make(map[string]CompanyState, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
