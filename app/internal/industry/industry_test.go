package industry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"twenty-fifty/internal/config"
	"twenty-fifty/internal/technology"
)

func makeDef(id string, baseQuality, baseWorkRate float64) config.CompanyDef {
	return config.CompanyDef{
		ID:           id,
		Name:         "Test Co",
		BaseQuality:  baseQuality,
		BaseWorkRate: baseWorkRate,
	}
}

func seedWith(defs ...config.CompanyDef) IndustryState {
	return SeedIndustry(defs)
}

// ---------------------------------------------------------------------------
// SeedIndustry
// ---------------------------------------------------------------------------

func TestSeedIndustry_AllCompaniesInactive(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70), makeDef("co_b", 60, 50))
	for _, cs := range state.Companies {
		assert.False(t, cs.IsActive)
	}
}

func TestSeedIndustry_StatusIsInactive(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	assert.Equal(t, CompanyStatusInactive, state.Companies["co_a"].Status)
}

func TestActivateCompany_StatusIsActive(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state2 := ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	assert.Equal(t, CompanyStatusActive, state2.Companies["co_a"].Status)
}

func TestDeactivateCompany_StatusIsInactive(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	state2 := DeactivateCompany(state, "co_a")
	assert.Equal(t, CompanyStatusInactive, state2.Companies["co_a"].Status)
}

func TestSeedIndustry_AllCompaniesHaveZeroAccumulatedQuality(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	assert.Equal(t, 0.0, state.Companies["co_a"].AccumulatedQuality)
}

func TestSeedIndustry_CountMatchesInput(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70), makeDef("co_b", 60, 50), makeDef("co_c", 40, 30))
	assert.Equal(t, 3, len(state.Companies))
}

func TestSeedIndustry_DoesNotMutateInput(t *testing.T) {
	defs := []config.CompanyDef{makeDef("co_a", 80, 70)}
	SeedIndustry(defs)
	assert.Equal(t, "co_a", defs[0].ID)
}

// ---------------------------------------------------------------------------
// ActivateCompany
// ---------------------------------------------------------------------------

func TestActivateCompany_SetsIsActiveTrue(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state2 := ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	assert.True(t, state2.Companies["co_a"].IsActive)
}

func TestActivateCompany_SetsContractedTech(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state2 := ActivateCompany(state, "co_a", config.TechHeatPumps, 70)
	assert.Equal(t, config.TechHeatPumps, state2.Companies["co_a"].ContractedTech)
}

func TestActivateCompany_ResetsAccumulatedQuality(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	// Manually set some quality to confirm it is reset.
	m := copyCompanies(state.Companies)
	cs := m["co_a"]
	cs.AccumulatedQuality = 999
	m["co_a"] = cs
	state = IndustryState{Companies: m}

	state2 := ActivateCompany(state, "co_a", config.TechSolarPV, 70)
	assert.Equal(t, 0.0, state2.Companies["co_a"].AccumulatedQuality)
}

func TestActivateCompany_UnknownDefID_ReturnsUnchanged(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state2 := ActivateCompany(state, "unknown", config.TechOffshoreWind, 70)
	assert.Equal(t, 1, len(state2.Companies))
	assert.False(t, state2.Companies["co_a"].IsActive)
}

func TestActivateCompany_DoesNotMutateInput(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	assert.False(t, state.Companies["co_a"].IsActive)
}

// ---------------------------------------------------------------------------
// DeactivateCompany
// ---------------------------------------------------------------------------

func TestDeactivateCompany_SetsIsActiveFalse(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	state2 := DeactivateCompany(state, "co_a")
	assert.False(t, state2.Companies["co_a"].IsActive)
}

func TestDeactivateCompany_ClearsContractedTech(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	state2 := DeactivateCompany(state, "co_a")
	assert.Equal(t, config.Technology(""), state2.Companies["co_a"].ContractedTech)
}

func TestDeactivateCompany_ClearsAccumulatedQuality(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	state = TickCompany(state, "co_a", makeDef("co_a", 80, 70), 50)
	state2 := DeactivateCompany(state, "co_a")
	assert.Equal(t, 0.0, state2.Companies["co_a"].AccumulatedQuality)
}

func TestDeactivateCompany_UnknownDefID_ReturnsUnchanged(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state2 := DeactivateCompany(state, "unknown")
	assert.Equal(t, state.Companies["co_a"], state2.Companies["co_a"])
}

func TestDeactivateCompany_DoesNotMutateInput(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	DeactivateCompany(state, "co_a")
	assert.True(t, state.Companies["co_a"].IsActive)
}

// ---------------------------------------------------------------------------
// TickCompany
// ---------------------------------------------------------------------------

func TestTickCompany_InactiveCompany_NoChange(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state2 := TickCompany(state, "co_a", makeDef("co_a", 80, 70), 50)
	assert.Equal(t, state.Companies["co_a"], state2.Companies["co_a"])
}

func TestTickCompany_FullCapacity_AccumulatesQuality(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 100)
	state2 := TickCompany(state, "co_a", makeDef("co_a", 80, 100), 50)
	assert.Greater(t, state2.Companies["co_a"].AccumulatedQuality, 0.0)
}

func TestTickCompany_ZeroCapacity_ReducedButNonZeroQuality(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 100)
	state2 := TickCompany(state, "co_a", makeDef("co_a", 80, 100), 0)
	// capacityDampening = 0.70 so work rate is 70% even at zero capacity
	assert.Greater(t, state2.Companies["co_a"].AccumulatedQuality, 0.0)
}

func TestTickCompany_WorkRateConvergesToBase(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 100) // start at 100
	def := makeDef("co_a", 80, 70)
	state2 := TickCompany(state, "co_a", def, 50)
	// Work rate should move from 100 toward 70
	assert.Less(t, state2.Companies["co_a"].WorkRate, 100.0)
	assert.Greater(t, state2.Companies["co_a"].WorkRate, 70.0)
}

func TestTickCompany_WeeksOnContractIncremented(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	state2 := TickCompany(state, "co_a", makeDef("co_a", 80, 70), 50)
	assert.Equal(t, 1, state2.Companies["co_a"].WeeksOnContract)
}

func TestTickCompany_AccumulatedQualityMonotonicallyIncreases(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	def := makeDef("co_a", 80, 70)
	prev := state.Companies["co_a"].AccumulatedQuality
	for i := 0; i < 10; i++ {
		state = TickCompany(state, "co_a", def, 50)
		curr := state.Companies["co_a"].AccumulatedQuality
		assert.GreaterOrEqual(t, curr, prev)
		prev = curr
	}
}

func TestTickCompany_DoesNotMutateInput(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	before := state.Companies["co_a"].AccumulatedQuality
	TickCompany(state, "co_a", makeDef("co_a", 80, 70), 50)
	assert.Equal(t, before, state.Companies["co_a"].AccumulatedQuality)
}

// ---------------------------------------------------------------------------
// DeliverTech
// ---------------------------------------------------------------------------

func TestDeliverTech_InactiveCompany_NoChange(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	tracker := technology.NewTechTracker(nil)
	state2, tracker2 := DeliverTech(state, "co_a", tracker, makeDef("co_a", 80, 70))
	assert.Equal(t, state.Companies["co_a"], state2.Companies["co_a"])
	assert.Equal(t, tracker.Maturity(config.TechOffshoreWind), tracker2.Maturity(config.TechOffshoreWind))
}

func TestDeliverTech_PositiveAccumulatedQuality_IncreasesMaturity(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	// Tick many weeks to build up quality.
	def := makeDef("co_a", 80, 70)
	for i := 0; i < 20; i++ {
		state = TickCompany(state, "co_a", def, 50)
	}
	curves := []config.TechCurveDef{{ID: config.TechOffshoreWind, InitialMaturity: 18}}
	tracker := technology.NewTechTracker(curves)
	before := tracker.Maturity(config.TechOffshoreWind)
	_, tracker2 := DeliverTech(state, "co_a", tracker, def)
	assert.Greater(t, tracker2.Maturity(config.TechOffshoreWind), before)
}

func TestDeliverTech_ResetsAccumulatedQuality(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	def := makeDef("co_a", 80, 70)
	for i := 0; i < 5; i++ {
		state = TickCompany(state, "co_a", def, 50)
	}
	tracker := technology.NewTechTracker(nil)
	state2, _ := DeliverTech(state, "co_a", tracker, def)
	assert.Equal(t, 0.0, state2.Companies["co_a"].AccumulatedQuality)
}

func TestDeliverTech_MaxBoostCapped(t *testing.T) {
	// Manually set very large accumulated quality and verify boost is capped.
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	m := copyCompanies(state.Companies)
	cs := m["co_a"]
	cs.AccumulatedQuality = 1_000_000 // enormous
	m["co_a"] = cs
	state = IndustryState{Companies: m}

	curves := []config.TechCurveDef{{ID: config.TechOffshoreWind, InitialMaturity: 0}}
	tracker := technology.NewTechTracker(curves)
	_, tracker2 := DeliverTech(state, "co_a", tracker, makeDef("co_a", 80, 70))
	// boost capped at maxDeliveryBoost=8.0
	assert.LessOrEqual(t, tracker2.Maturity(config.TechOffshoreWind), 8.0+0.001)
}

func TestDeliverTech_CorrectTechBoosted(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechHeatPumps, 70)
	def := makeDef("co_a", 80, 70)
	for i := 0; i < 10; i++ {
		state = TickCompany(state, "co_a", def, 50)
	}
	curves := []config.TechCurveDef{
		{ID: config.TechHeatPumps, InitialMaturity: 5},
		{ID: config.TechOffshoreWind, InitialMaturity: 18},
	}
	tracker := technology.NewTechTracker(curves)
	beforeWind := tracker.Maturity(config.TechOffshoreWind)
	_, tracker2 := DeliverTech(state, "co_a", tracker, def)
	// Only heat pumps should increase; wind should be unchanged.
	assert.Greater(t, tracker2.Maturity(config.TechHeatPumps), 5.0)
	assert.InDelta(t, beforeWind, tracker2.Maturity(config.TechOffshoreWind), 0.001)
}

func TestDeliverTech_DoesNotMutateInput(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	before := state.Companies["co_a"].AccumulatedQuality
	tracker := technology.NewTechTracker(nil)
	DeliverTech(state, "co_a", tracker, makeDef("co_a", 80, 70))
	assert.Equal(t, before, state.Companies["co_a"].AccumulatedQuality)
}

// ---------------------------------------------------------------------------
// ActiveCompaniesForTech
// ---------------------------------------------------------------------------

func TestActiveCompaniesForTech_ReturnsOnlyMatchingTech(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70), makeDef("co_b", 80, 70))
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	state = ActivateCompany(state, "co_b", config.TechHeatPumps, 70)
	ids := ActiveCompaniesForTech(state, config.TechOffshoreWind)
	assert.Equal(t, []string{"co_a"}, ids)
}

func TestActiveCompaniesForTech_ExcludesInactive(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	ids := ActiveCompaniesForTech(state, config.TechOffshoreWind)
	assert.Equal(t, 0, len(ids))
}

func TestActiveCompaniesForTech_EmptyResult_IsNonNilSlice(t *testing.T) {
	state := seedWith(makeDef("co_a", 80, 70))
	ids := ActiveCompaniesForTech(state, config.TechOffshoreWind)
	assert.NotNil(t, ids)
}

func TestActiveCompaniesForTech_IsDeterministicallySorted(t *testing.T) {
	state := seedWith(makeDef("co_b", 80, 70), makeDef("co_a", 80, 70))
	state = ActivateCompany(state, "co_b", config.TechOffshoreWind, 70)
	state = ActivateCompany(state, "co_a", config.TechOffshoreWind, 70)
	ids := ActiveCompaniesForTech(state, config.TechOffshoreWind)
	assert.Equal(t, []string{"co_a", "co_b"}, ids)
}
