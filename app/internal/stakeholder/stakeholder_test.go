package stakeholder

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"twenty-fifty/internal/carbon"
	"twenty-fifty/internal/config"
)

func seededRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

func makeSeed(id string, timing config.EntryTiming, entryMin int) config.StakeholderSeed {
	return config.StakeholderSeed{
		ID:              id,
		Party:           config.PartyLeft,
		Role:            config.RoleEnergy,
		EntryTiming:     timing,
		EntryWeekMin:    entryMin,
		EntryWeekMax:    entryMin + 52,
		Name:            "Test Person",
		IdeologyScore:   0,
		NetZeroSympathy: 50,
		RiskTolerance:   50,
		PopulismScore:   50,
		SpecialMechanic: config.MechanicNone,
	}
}

// ---------------------------------------------------------------------------
// SeedStakeholders
// ---------------------------------------------------------------------------

func TestSeedStakeholders_AllSeeds_StartAtRelationship50(t *testing.T) {
	defs := []config.StakeholderSeed{
		makeSeed("a", config.TimingStart, 0),
		makeSeed("b", config.TimingMid, 260),
	}
	out := SeedStakeholders(defs)
	for _, s := range out {
		assert.InDelta(t, 50.0, s.RelationshipScore, 0.001)
	}
}

func TestSeedStakeholders_StartTimingSeeds_AreUnlocked(t *testing.T) {
	defs := []config.StakeholderSeed{makeSeed("a", config.TimingStart, 0)}
	out := SeedStakeholders(defs)
	assert.True(t, out[0].IsUnlocked)
}

func TestSeedStakeholders_MidTimingSeeds_AreLockedAtStart(t *testing.T) {
	defs := []config.StakeholderSeed{makeSeed("a", config.TimingMid, 260)}
	out := SeedStakeholders(defs)
	assert.False(t, out[0].IsUnlocked)
}

func TestSeedStakeholders_PendingSignals_IsNonNilEmptySlice(t *testing.T) {
	defs := []config.StakeholderSeed{makeSeed("a", config.TimingStart, 0)}
	out := SeedStakeholders(defs)
	assert.NotNil(t, out[0].PendingSignals)
	assert.Equal(t, 0, len(out[0].PendingSignals))
}

func TestSeedStakeholders_CountMatchesDefs(t *testing.T) {
	defs := []config.StakeholderSeed{
		makeSeed("a", config.TimingStart, 0),
		makeSeed("b", config.TimingMid, 260),
		makeSeed("c", config.TimingLate, 520),
	}
	out := SeedStakeholders(defs)
	assert.Equal(t, 3, len(out))
}

func TestSeedStakeholders_StateIsActive(t *testing.T) {
	defs := []config.StakeholderSeed{makeSeed("a", config.TimingStart, 0)}
	out := SeedStakeholders(defs)
	assert.Equal(t, MinisterStateActive, out[0].State)
}

func TestSeedStakeholders_PopularityStartsAt50(t *testing.T) {
	defs := []config.StakeholderSeed{makeSeed("a", config.TimingStart, 0)}
	out := SeedStakeholders(defs)
	assert.Equal(t, 50.0, out[0].Popularity)
}

func TestSeedStakeholders_WeeksUnderPressureStartsAtZero(t *testing.T) {
	defs := []config.StakeholderSeed{makeSeed("a", config.TimingStart, 0)}
	out := SeedStakeholders(defs)
	assert.Equal(t, 0, out[0].WeeksUnderPressure)
}

func TestSeedStakeholders_IdeologyConflictScoreStartsAtZero(t *testing.T) {
	defs := []config.StakeholderSeed{makeSeed("a", config.TimingStart, 0)}
	out := SeedStakeholders(defs)
	assert.Equal(t, 0.0, out[0].IdeologyConflictScore)
}

func TestSeedStakeholders_IdentityFieldsCopied(t *testing.T) {
	d := makeSeed("my_id", config.TimingStart, 0)
	d.IdeologyScore = 42.0
	d.NetZeroSympathy = 75.0
	out := SeedStakeholders([]config.StakeholderSeed{d})
	assert.Equal(t, "my_id", out[0].ID)
	assert.InDelta(t, 42.0, out[0].IdeologyScore, 0.001)
	assert.InDelta(t, 75.0, out[0].NetZeroSympathy, 0.001)
}

// ---------------------------------------------------------------------------
// UnlockStakeholder
// ---------------------------------------------------------------------------

func TestUnlockStakeholder_WeekAtMin_SetsUnlocked(t *testing.T) {
	s := Stakeholder{IsUnlocked: false, EntryWeekMin: 100}
	s2 := UnlockStakeholder(s, 100)
	assert.True(t, s2.IsUnlocked)
}

func TestUnlockStakeholder_WeekBelowMin_RemainsLocked(t *testing.T) {
	s := Stakeholder{IsUnlocked: false, EntryWeekMin: 100}
	s2 := UnlockStakeholder(s, 99)
	assert.False(t, s2.IsUnlocked)
}

func TestUnlockStakeholder_AlreadyUnlocked_NoChange(t *testing.T) {
	s := Stakeholder{IsUnlocked: true, EntryWeekMin: 100}
	s2 := UnlockStakeholder(s, 50)
	assert.True(t, s2.IsUnlocked)
}

func TestUnlockStakeholder_DoesNotMutateInput(t *testing.T) {
	s := Stakeholder{IsUnlocked: false, EntryWeekMin: 10}
	UnlockStakeholder(s, 50)
	assert.False(t, s.IsUnlocked)
}

// ---------------------------------------------------------------------------
// TickRelationship
// ---------------------------------------------------------------------------

func TestTickRelationship_NeutralDecay_ConvergesTo50FromAbove(t *testing.T) {
	s := Stakeholder{RelationshipScore: 80.0}
	s2 := TickRelationship(s, 0, 0)
	assert.Less(t, s2.RelationshipScore, 80.0)
	assert.Greater(t, s2.RelationshipScore, 50.0)
}

func TestTickRelationship_NeutralDecay_ConvergesTo50FromBelow(t *testing.T) {
	s := Stakeholder{RelationshipScore: 20.0}
	s2 := TickRelationship(s, 0, 0)
	assert.Greater(t, s2.RelationshipScore, 20.0)
	assert.Less(t, s2.RelationshipScore, 50.0)
}

func TestTickRelationship_PositiveAction_IncreasesScore(t *testing.T) {
	s := Stakeholder{RelationshipScore: 50.0}
	s2 := TickRelationship(s, 3.0, 0)
	assert.Greater(t, s2.RelationshipScore, 50.0)
}

func TestTickRelationship_NegativeEvent_DecreasesScore(t *testing.T) {
	s := Stakeholder{RelationshipScore: 50.0}
	s2 := TickRelationship(s, 0, -3.0)
	assert.Less(t, s2.RelationshipScore, 50.0)
}

func TestTickRelationship_LargePositiveAction_ClampedAt100(t *testing.T) {
	s := Stakeholder{RelationshipScore: 98.0}
	s2 := TickRelationship(s, 999.0, 999.0)
	assert.Equal(t, 100.0, s2.RelationshipScore)
}

func TestTickRelationship_LargeNegativeAction_ClampedAt0(t *testing.T) {
	s := Stakeholder{RelationshipScore: 2.0}
	s2 := TickRelationship(s, -999.0, -999.0)
	assert.Equal(t, 0.0, s2.RelationshipScore)
}

func TestTickRelationship_DoesNotMutateInput(t *testing.T) {
	s := Stakeholder{RelationshipScore: 70.0}
	TickRelationship(s, 5.0, 0)
	assert.InDelta(t, 70.0, s.RelationshipScore, 0.001)
}

// ---------------------------------------------------------------------------
// ComputeInfluence
// ---------------------------------------------------------------------------

func TestComputeInfluence_HighPolling_HighInfluence(t *testing.T) {
	s := Stakeholder{Party: config.PartyLeft, Role: config.RoleLeader, RelationshipScore: 80}
	polling := map[config.Party]float64{config.PartyLeft: 90}
	s2 := ComputeInfluence(s, polling, DefaultRoleWeights)
	assert.Greater(t, s2.InfluenceScore, 70.0)
}

func TestComputeInfluence_MissingParty_UsesDefault25(t *testing.T) {
	s := Stakeholder{Party: config.PartyFarRight, Role: config.RoleEnergy, RelationshipScore: 50}
	polling := map[config.Party]float64{} // party absent
	s2 := ComputeInfluence(s, polling, DefaultRoleWeights)
	// (25*0.60 + 50*0.40) * 1.0 = 35.0
	assert.InDelta(t, 35.0, s2.InfluenceScore, 0.5)
}

func TestComputeInfluence_InfluenceClamped0To100(t *testing.T) {
	s := Stakeholder{Party: config.PartyLeft, Role: config.RoleLeader, RelationshipScore: 100}
	polling := map[config.Party]float64{config.PartyLeft: 100}
	roleWeights := map[config.Role]float64{config.RoleLeader: 5.0} // huge weight
	s2 := ComputeInfluence(s, polling, roleWeights)
	assert.Equal(t, 100.0, s2.InfluenceScore)
}

func TestComputeInfluence_DoesNotMutateInput(t *testing.T) {
	s := Stakeholder{Party: config.PartyLeft, Role: config.RoleEnergy, RelationshipScore: 50, InfluenceScore: 0}
	polling := map[config.Party]float64{config.PartyLeft: 40}
	ComputeInfluence(s, polling, DefaultRoleWeights)
	assert.InDelta(t, 0.0, s.InfluenceScore, 0.001)
}

// ---------------------------------------------------------------------------
// TickSpecialMechanic
// ---------------------------------------------------------------------------

func TestTickSpecialMechanic_TickyPressure_ElevatedIncrementsCounter(t *testing.T) {
	s := Stakeholder{SpecialMechanic: config.MechanicTickyPressure, TickyPressureCounter: 3}
	s2 := TickSpecialMechanic(s, carbon.ClimateLevelElevated, nil)
	assert.Equal(t, 4, s2.TickyPressureCounter)
}

func TestTickSpecialMechanic_TickyPressure_StableResetsCounter(t *testing.T) {
	s := Stakeholder{SpecialMechanic: config.MechanicTickyPressure, TickyPressureCounter: 10}
	s2 := TickSpecialMechanic(s, carbon.ClimateLevelStable, nil)
	assert.Equal(t, 0, s2.TickyPressureCounter)
}

func TestTickSpecialMechanic_TickyPressure_OtherMechanicIgnored(t *testing.T) {
	s := Stakeholder{SpecialMechanic: config.MechanicDizzySurge, TickyPressureCounter: 0}
	s2 := TickSpecialMechanic(s, carbon.ClimateLevelElevated, nil)
	assert.Equal(t, 0, s2.TickyPressureCounter)
}

func TestTickSpecialMechanic_DizzySurge_RandomToggle_ReturnsValidState(t *testing.T) {
	s := Stakeholder{SpecialMechanic: config.MechanicDizzySurge}
	// Run many ticks and verify state is always bool (always valid).
	rng := seededRNG(42)
	for i := 0; i < 200; i++ {
		s = TickSpecialMechanic(s, carbon.ClimateLevelStable, rng)
	}
	// DizzySurgeActive is a bool; test just that it doesn't panic and stays valid.
	_ = s.DizzySurgeActive
}

func TestTickSpecialMechanic_ElectoralFatigue_AlwaysIncrements(t *testing.T) {
	s := Stakeholder{SpecialMechanic: config.MechanicElectoralFatigue, ElectoralFatigueCount: 5}
	s2 := TickSpecialMechanic(s, carbon.ClimateLevelStable, nil)
	assert.Equal(t, 6, s2.ElectoralFatigueCount)
}

func TestTickSpecialMechanic_NoMechanic_StateUnchanged(t *testing.T) {
	s := Stakeholder{
		SpecialMechanic:      config.MechanicNone,
		TickyPressureCounter: 3,
		ElectoralFatigueCount: 7,
	}
	s2 := TickSpecialMechanic(s, carbon.ClimateLevelEmergency, nil)
	assert.Equal(t, 3, s2.TickyPressureCounter)
	assert.Equal(t, 7, s2.ElectoralFatigueCount)
}

func TestTickSpecialMechanic_DoesNotMutateInput(t *testing.T) {
	s := Stakeholder{SpecialMechanic: config.MechanicTickyPressure, TickyPressureCounter: 0}
	TickSpecialMechanic(s, carbon.ClimateLevelElevated, nil)
	assert.Equal(t, 0, s.TickyPressureCounter)
}

// ---------------------------------------------------------------------------
// ApprovalChance
// ---------------------------------------------------------------------------

func makePolicy(sector config.PolicySector) config.PolicyCardDef {
	return config.PolicyCardDef{ID: "test_policy", Sector: sector}
}

func TestApprovalChance_HighSympathyHighRelation_HighProbability(t *testing.T) {
	s := Stakeholder{
		IdeologyScore:     0,
		NetZeroSympathy:   90,
		RelationshipScore: 90,
	}
	p := ApprovalChance(s, makePolicy(config.PolicySectorPower))
	assert.Greater(t, p, 0.70)
}

func TestApprovalChance_LowSympathyLowRelationIdeologyConflict_LowProbability(t *testing.T) {
	// Far-left stakeholder with low sympathy/relationship approving a right-leaning
	// industry policy: ideology misalignment + low sympathy + low relationship.
	s := Stakeholder{
		IdeologyScore:     -80,
		NetZeroSympathy:   10,
		RelationshipScore: 10,
	}
	p := ApprovalChance(s, makePolicy(config.PolicySectorIndustry))
	assert.Less(t, p, 0.35)
}

func TestApprovalChance_AlwaysInRange(t *testing.T) {
	sectors := []config.PolicySector{
		config.PolicySectorPower,
		config.PolicySectorTransport,
		config.PolicySectorBuildings,
		config.PolicySectorIndustry,
		config.PolicySectorCross,
	}
	for ideo := -100.0; ideo <= 100.0; ideo += 50 {
		for nzs := 0.0; nzs <= 100.0; nzs += 25 {
			for rel := 0.0; rel <= 100.0; rel += 25 {
				s := Stakeholder{IdeologyScore: ideo, NetZeroSympathy: nzs, RelationshipScore: rel}
				for _, sec := range sectors {
					p := ApprovalChance(s, makePolicy(sec))
					assert.GreaterOrEqual(t, p, 0.05)
					assert.LessOrEqual(t, p, 0.95)
				}
			}
		}
	}
}

func TestApprovalChance_FarRightSeed_IndustryHigherThanBuildings(t *testing.T) {
	s := Stakeholder{IdeologyScore: 80, NetZeroSympathy: 30, RelationshipScore: 50}
	pIndustry := ApprovalChance(s, makePolicy(config.PolicySectorIndustry))
	pBuildings := ApprovalChance(s, makePolicy(config.PolicySectorBuildings))
	assert.Greater(t, pIndustry, pBuildings)
}

func TestApprovalChance_DoesNotMutateInput(t *testing.T) {
	s := Stakeholder{IdeologyScore: 20, NetZeroSympathy: 60, RelationshipScore: 70}
	ApprovalChance(s, makePolicy(config.PolicySectorPower))
	assert.InDelta(t, 70.0, s.RelationshipScore, 0.001)
}
