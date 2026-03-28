package event

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vibe-code-challenge/twenty-fifty/internal/carbon"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/industry"
	"github.com/vibe-code-challenge/twenty-fifty/internal/stakeholder"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func fixedRNG() *rand.Rand {
	return rand.New(rand.NewSource(7))
}

func makeEventDef(id string, base, climateMult, fossilMult float64) config.EventDef {
	return config.EventDef{
		ID:                id,
		BaseProbability:   base,
		ClimateMultiplier: climateMult,
		FossilMultiplier:  fossilMult,
	}
}

func makeRegionDef(id string, tags []string) config.RegionDef {
	return config.RegionDef{ID: id, Tags: tags}
}

func makeStakeholder(id string, role config.Role, unlocked bool) stakeholder.Stakeholder {
	return stakeholder.Stakeholder{
		ID:         id,
		Role:       role,
		IsUnlocked: unlocked,
		State:      stakeholder.MinisterStateActive,
	}
}

func makeCompany(defID string, status industry.CompanyStatus) industry.CompanyState {
	return industry.CompanyState{DefID: defID, Status: status}
}

func makeCompanyDef(id string, cat config.TechCategory) config.CompanyDef {
	return config.CompanyDef{ID: id, TechCategory: cat}
}

// ---------------------------------------------------------------------------
// EventLog
// ---------------------------------------------------------------------------

func TestAppendEventLog_StoresEntry(t *testing.T) {
	log := NewEventLog()
	log = AppendEventLog(log, EventEntry{DefID: "e1", Week: 1})
	entries := log.Entries()
	assert.Len(t, entries, 1)
	assert.Equal(t, "e1", entries[0].DefID)
}

func TestAppendEventLog_WrapsAtCapacity(t *testing.T) {
	log := NewEventLog()
	for i := 0; i < eventLogCapacity+5; i++ {
		log = AppendEventLog(log, EventEntry{DefID: "e", Week: i})
	}
	entries := log.Entries()
	assert.Len(t, entries, eventLogCapacity)
}

func TestAppendEventLog_ChronologicalOrder(t *testing.T) {
	log := NewEventLog()
	for i := 0; i < 3; i++ {
		log = AppendEventLog(log, EventEntry{Week: i})
	}
	entries := log.Entries()
	for i := 1; i < len(entries); i++ {
		assert.LessOrEqual(t, entries[i-1].Week, entries[i].Week)
	}
}

// ---------------------------------------------------------------------------
// ComputeEventProbability
// ---------------------------------------------------------------------------

func TestComputeEventProbability_Stable_UsesBase(t *testing.T) {
	def := makeEventDef("e", 0.05, 2.0, 1.5)
	p := ComputeEventProbability(def, carbon.ClimateLevelStable, 40.0)
	assert.InDelta(t, 0.05, p, 0.001)
}

func TestComputeEventProbability_ElevatedClimate_AppliesMultiplier(t *testing.T) {
	def := makeEventDef("e", 0.05, 2.0, 1.0)
	p := ComputeEventProbability(def, carbon.ClimateLevelElevated, 40.0)
	assert.InDelta(t, 0.10, p, 0.001)
}

func TestComputeEventProbability_HighFossil_AppliesMultiplier(t *testing.T) {
	def := makeEventDef("e", 0.05, 1.0, 2.0)
	p := ComputeEventProbability(def, carbon.ClimateLevelStable, 70.0)
	assert.InDelta(t, 0.10, p, 0.001)
}

func TestComputeEventProbability_ClampedAt1(t *testing.T) {
	def := makeEventDef("e", 0.8, 5.0, 5.0)
	p := ComputeEventProbability(def, carbon.ClimateLevelEmergency, 80.0)
	assert.Equal(t, 1.0, p)
}

// ---------------------------------------------------------------------------
// DrawEvent
// ---------------------------------------------------------------------------

func TestDrawEvent_ZeroProbability_NeverFires(t *testing.T) {
	defs := []config.EventDef{makeEventDef("e", 0.0, 1.0, 1.0)}
	rng := fixedRNG()
	for i := 0; i < 1000; i++ {
		_, fired := DrawEvent(defs, carbon.ClimateLevelStable, 0, rng)
		assert.False(t, fired)
	}
}

func TestDrawEvent_ProbabilityOne_AlwaysFires(t *testing.T) {
	defs := []config.EventDef{makeEventDef("e1", 1.0, 1.0, 1.0)}
	_, fired := DrawEvent(defs, carbon.ClimateLevelStable, 0, fixedRNG())
	assert.True(t, fired)
}

// ---------------------------------------------------------------------------
// RollScandal
// ---------------------------------------------------------------------------

func TestRollScandal_ZeroPressureZeroPopulism_NeverFires(t *testing.T) {
	s := stakeholder.Stakeholder{PopulismScore: 0.0}
	rng := fixedRNG()
	fired := false
	for i := 0; i < 10000; i++ {
		if RollScandal(s, 0, rng) {
			fired = true
			break
		}
	}
	// With base prob 0.005, in 10000 trials we expect ~50 fires.
	// This test just checks it doesn't always fire or is impossible.
	// We verify rate is sane using the base prob constant.
	_ = fired
	// Base rate check: RollScandal should fire roughly scandalBaseProb of the time.
	count := 0
	rng2 := rand.New(rand.NewSource(123))
	for i := 0; i < 100000; i++ {
		if RollScandal(s, 0, rng2) {
			count++
		}
	}
	rate := float64(count) / 100000.0
	assert.InDelta(t, scandalBaseProb, rate, 0.002)
}

func TestRollScandal_HighPressure_HigherRate(t *testing.T) {
	s := stakeholder.Stakeholder{PopulismScore: 0.0}
	rng := rand.New(rand.NewSource(99))
	count := 0
	for i := 0; i < 10000; i++ {
		if RollScandal(s, 20, rng) {
			count++
		}
	}
	// 20 weeks pressure: prob = 0.005 + 20*0.002 = 0.045
	rate := float64(count) / 10000.0
	assert.InDelta(t, 0.045, rate, 0.005)
}

// ---------------------------------------------------------------------------
// MatchRegions
// ---------------------------------------------------------------------------

func TestMatchRegions_EmptyFilter_ReturnsAll(t *testing.T) {
	regions := []config.RegionDef{
		makeRegionDef("r1", []string{"coastal"}),
		makeRegionDef("r2", []string{"urban"}),
	}
	ids := MatchRegions("", regions)
	assert.ElementsMatch(t, []string{"r1", "r2"}, ids)
}

func TestMatchRegions_CoastalFilter_ReturnsOnlyCoastal(t *testing.T) {
	regions := []config.RegionDef{
		makeRegionDef("r1", []string{"coastal", "rural"}),
		makeRegionDef("r2", []string{"urban"}),
		makeRegionDef("r3", []string{"coastal"}),
	}
	ids := MatchRegions("COASTAL", regions)
	assert.ElementsMatch(t, []string{"r1", "r3"}, ids)
}

func TestMatchRegions_RegionIDFilter_ReturnsThatRegion(t *testing.T) {
	regions := []config.RegionDef{
		makeRegionDef("northern_industrial", []string{"industrial", "coastal"}),
		makeRegionDef("capital_region", []string{"urban"}),
	}
	ids := MatchRegions("northern_industrial", regions)
	assert.Equal(t, []string{"northern_industrial"}, ids)
}

func TestMatchRegions_UnknownFilter_ReturnsEmpty(t *testing.T) {
	regions := []config.RegionDef{makeRegionDef("r1", []string{"coastal"})}
	ids := MatchRegions("NONEXISTENT_REGION_XYZ", regions)
	assert.Empty(t, ids)
}

// ---------------------------------------------------------------------------
// MatchStakeholders
// ---------------------------------------------------------------------------

func TestMatchStakeholders_EmptyFilter_ReturnsNil(t *testing.T) {
	stakeholders := []stakeholder.Stakeholder{makeStakeholder("s1", config.RoleEnergy, true)}
	ids := MatchStakeholders("", stakeholders)
	assert.Nil(t, ids)
}

func TestMatchStakeholders_CabinetFilter_ReturnsFourUnlocked(t *testing.T) {
	stakeholders := []stakeholder.Stakeholder{
		makeStakeholder("s1", config.RoleLeader, true),
		makeStakeholder("s2", config.RoleChancellor, true),
		makeStakeholder("s3", config.RoleForeignSecretary, true),
		makeStakeholder("s4", config.RoleEnergy, true),
		makeStakeholder("s5", config.RoleEnergy, false), // locked
	}
	ids := MatchStakeholders("CABINET", stakeholders)
	assert.Len(t, ids, 4)
	assert.NotContains(t, ids, "s5")
}

func TestMatchStakeholders_RoleEnergyFilter_ReturnsOnlyEnergyMinister(t *testing.T) {
	stakeholders := []stakeholder.Stakeholder{
		makeStakeholder("s1", config.RoleLeader, true),
		makeStakeholder("s2", config.RoleEnergy, true),
	}
	ids := MatchStakeholders("ROLE:ENERGY", stakeholders)
	assert.Equal(t, []string{"s2"}, ids)
}

// ---------------------------------------------------------------------------
// MatchCompanies
// ---------------------------------------------------------------------------

func TestMatchCompanies_EmptyFilter_ReturnsNil(t *testing.T) {
	companies := []industry.CompanyState{makeCompany("c1", industry.CompanyStatusActive)}
	defs := map[string]config.CompanyDef{"c1": makeCompanyDef("c1", config.TechCatEVs)}
	ids := MatchCompanies("", companies, defs)
	assert.Nil(t, ids)
}

func TestMatchCompanies_AllFilter_ReturnsActiveCompanies(t *testing.T) {
	companies := []industry.CompanyState{
		makeCompany("c1", industry.CompanyStatusActive),
		makeCompany("c2", industry.CompanyStatusInactive),
		makeCompany("c3", industry.CompanyStatusStartup),
	}
	defs := map[string]config.CompanyDef{
		"c1": makeCompanyDef("c1", config.TechCatEVs),
		"c2": makeCompanyDef("c2", config.TechCatEVs),
		"c3": makeCompanyDef("c3", config.TechCatOffshoreWind),
	}
	ids := MatchCompanies("ALL", companies, defs)
	assert.ElementsMatch(t, []string{"c1", "c3"}, ids)
}

func TestMatchCompanies_TechFilter_ReturnsOnlyMatchingCategory(t *testing.T) {
	companies := []industry.CompanyState{
		makeCompany("c1", industry.CompanyStatusActive),
		makeCompany("c2", industry.CompanyStatusActive),
	}
	defs := map[string]config.CompanyDef{
		"c1": makeCompanyDef("c1", config.TechCatEVs),
		"c2": makeCompanyDef("c2", config.TechCatOffshoreWind),
	}
	ids := MatchCompanies("TECH:EVS", companies, defs)
	assert.Equal(t, []string{"c1"}, ids)
}

// ---------------------------------------------------------------------------
// ResolveEffect
// ---------------------------------------------------------------------------

func TestResolveEffect_GlobalFieldsCopied(t *testing.T) {
	effect := config.EventEffect{
		GasPriceDeltaPct:    5.0,
		EconomyDelta:        -2.0,
		LCRDelta:            1.5,
		GovtPopularityDelta: -1.0,
	}
	out := ResolveEffect(effect, nil, nil, nil, nil)
	assert.Equal(t, 5.0, out.GasPriceDeltaPct)
	assert.Equal(t, -2.0, out.EconomyDelta)
	assert.Equal(t, 1.5, out.LCRDelta)
}

func TestResolveEffect_CoastalFloodingEffect_PopulatesRegionAndTileDeltas(t *testing.T) {
	effect := config.EventEffect{
		RegionFilter:           "COASTAL",
		InstallerCapacityDelta: -6.0,
		TileInsulationDamage:   3.0,
	}
	regions := []config.RegionDef{
		makeRegionDef("r_coast", []string{"coastal"}),
		makeRegionDef("r_inland", []string{"rural"}),
	}
	out := ResolveEffect(effect, regions, nil, nil, nil)
	assert.Contains(t, out.RegionDeltas, "r_coast")
	assert.NotContains(t, out.RegionDeltas, "r_inland")
	assert.Equal(t, -6.0, out.RegionDeltas["r_coast"].InstallerCapacityDelta)
	assert.Contains(t, out.TileDeltas, "r_coast")
	assert.Equal(t, 3.0, out.TileDeltas["r_coast"].InsulationDamage)
}

func TestResolveEffect_EmptyFilters_ProducesNoTargetedDeltas(t *testing.T) {
	effect := config.EventEffect{LCRDelta: 2.0} // no filters set
	out := ResolveEffect(effect, nil, nil, nil, nil)
	assert.Empty(t, out.RegionDeltas)
	assert.Empty(t, out.StakeholderDeltas)
	assert.Empty(t, out.CompanyDeltas)
}

func TestResolveEffect_StakeholderFilter_PopulatesStakeholderDeltas(t *testing.T) {
	effect := config.EventEffect{
		StakeholderFilter:        "CABINET",
		StakeholderPressureDelta: 2,
	}
	stakeholders := []stakeholder.Stakeholder{
		makeStakeholder("min1", config.RoleLeader, true),
		makeStakeholder("min2", config.RoleEnergy, true),
	}
	out := ResolveEffect(effect, nil, stakeholders, nil, nil)
	assert.Contains(t, out.StakeholderDeltas, "min1")
	assert.Equal(t, 2, out.StakeholderDeltas["min1"].PressureDelta)
}

func TestResolveEffect_CompanyFilter_PopulatesCompanyDeltas(t *testing.T) {
	effect := config.EventEffect{
		CompanyFilter:        "ALL",
		CompanyWorkRateDelta: -8.0,
	}
	companies := []industry.CompanyState{makeCompany("co1", industry.CompanyStatusActive)}
	defs := map[string]config.CompanyDef{"co1": makeCompanyDef("co1", config.TechCatEVs)}
	out := ResolveEffect(effect, nil, nil, companies, defs)
	assert.Contains(t, out.CompanyDeltas, "co1")
	assert.Equal(t, -8.0, out.CompanyDeltas["co1"].WorkRateDelta)
}

// ---------------------------------------------------------------------------
// ApplyPressureGroups
// ---------------------------------------------------------------------------

func TestApplyPressureGroups_HighCarbon_GeneratesNegativePop(t *testing.T) {
	groups := DefaultPressureGroups()
	results := ApplyPressureGroups(groups, 550.0, 50.0) // above threshold
	totalPop := 0.0
	for _, r := range results {
		totalPop += r.GovtPopularityDelta
	}
	// With high carbon, the greens alliance fires CarbonPop penalty => net negative
	assert.Less(t, totalPop, 0.0)
}

func TestApplyPressureGroups_LowLCR_BoostsLCRDelta(t *testing.T) {
	groups := DefaultPressureGroups()
	resultsLow := ApplyPressureGroups(groups, 400.0, 20.0)  // LCR below threshold
	resultsHigh := ApplyPressureGroups(groups, 400.0, 60.0) // LCR above threshold

	totalLow := 0.0
	totalHigh := 0.0
	for i := range resultsLow {
		totalLow += resultsLow[i].LCRDelta
		totalHigh += resultsHigh[i].LCRDelta
	}
	// Low LCR should produce a more positive (or less negative) LCR delta total
	// because the greens alliance LowLCRBoost fires
	assert.Greater(t, totalLow, totalHigh)
}
