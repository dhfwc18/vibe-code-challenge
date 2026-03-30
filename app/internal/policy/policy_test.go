package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/region"
	"github.com/vibe-code-challenge/twenty-fifty/internal/stakeholder"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeCard(sector config.PolicySector, steps []config.ApprovalRequirement) PolicyCard {
	return PolicyCard{
		Def: &config.PolicyCardDef{
			ID:     "test_card",
			Sector: sector,
			WeeklyEffect: config.WeeklyEffectDef{
				Sector:            sector,
				BaseCarbonDeltaMt: -1.0,
				BudgetCostPerWeek: 5.0,
			},
			ApprovalSteps: steps,
		},
		State: PolicyStateDraft,
	}
}

func makeStakeholder(role config.Role, ideology, relationship float64, unlocked bool) stakeholder.Stakeholder {
	return stakeholder.Stakeholder{
		ID:                "s1",
		Role:              role,
		IdeologyScore:     ideology,
		RelationshipScore: relationship,
		NetZeroSympathy:   50.0,
		IsUnlocked:        unlocked,
		State:             stakeholder.MinisterStateActive,
	}
}

// ---------------------------------------------------------------------------
// SeedPolicyCards
// ---------------------------------------------------------------------------

func TestSeedPolicyCards_AllStartAsDraft(t *testing.T) {
	defs := []config.PolicyCardDef{
		{ID: "a", Sector: config.PolicySectorPower},
		{ID: "b", Sector: config.PolicySectorTransport},
	}
	cards := SeedPolicyCards(defs)
	assert.Len(t, cards, 2)
	for _, c := range cards {
		assert.Equal(t, PolicyStateDraft, c.State)
	}
}

// ---------------------------------------------------------------------------
// IsUnlocked
// ---------------------------------------------------------------------------

func TestIsUnlocked_NoGate_AlwaysTrue(t *testing.T) {
	card := makeCard(config.PolicySectorPower, nil)
	assert.True(t, IsUnlocked(card, nil))
}

func TestIsUnlocked_GateNotMet_ReturnsFalse(t *testing.T) {
	card := makeCard(config.PolicySectorPower, nil)
	card.Def.TechUnlockGate = config.TechOffshoreWind
	card.Def.TechUnlockThreshold = 50.0
	maturity := map[config.Technology]float64{config.TechOffshoreWind: 30.0}
	assert.False(t, IsUnlocked(card, maturity))
}

func TestIsUnlocked_GateMet_ReturnsTrue(t *testing.T) {
	card := makeCard(config.PolicySectorPower, nil)
	card.Def.TechUnlockGate = config.TechOffshoreWind
	card.Def.TechUnlockThreshold = 50.0
	maturity := map[config.Technology]float64{config.TechOffshoreWind: 60.0}
	assert.True(t, IsUnlocked(card, maturity))
}

func TestIsUnlocked_TechAbsent_ReturnsFalse(t *testing.T) {
	card := makeCard(config.PolicySectorPower, nil)
	card.Def.TechUnlockGate = config.TechHydrogen
	card.Def.TechUnlockThreshold = 20.0
	maturity := map[config.Technology]float64{config.TechOffshoreWind: 90.0}
	assert.False(t, IsUnlocked(card, maturity))
}

// ---------------------------------------------------------------------------
// SubmitPolicy
// ---------------------------------------------------------------------------

func TestSubmitPolicy_Draft_MovesToUnderReview(t *testing.T) {
	card := makeCard(config.PolicySectorPower, nil)
	card = SubmitPolicy(card)
	assert.Equal(t, PolicyStateUnderReview, card.State)
}

func TestSubmitPolicy_NonDraft_IsNoop(t *testing.T) {
	card := makeCard(config.PolicySectorPower, nil)
	card.State = PolicyStateActive
	card = SubmitPolicy(card)
	assert.Equal(t, PolicyStateActive, card.State)
}

// ---------------------------------------------------------------------------
// IdeologyConflict
// ---------------------------------------------------------------------------

func TestIdeologyConflict_AlignedStakeholder_LowConflict(t *testing.T) {
	// Industry policy sits at +20; a stakeholder at +20 has zero conflict
	s := makeStakeholder(config.RoleEnergy, 20.0, 50.0, true)
	conflict := IdeologyConflict(config.PolicyCardDef{Sector: config.PolicySectorIndustry}, s)
	assert.InDelta(t, 0.0, conflict, 0.01)
}

func TestIdeologyConflict_OpposedStakeholder_HighConflict(t *testing.T) {
	// Industry policy sits at +20; a far-left stakeholder at -80 has raw conflict 100.
	// makeStakeholder sets NZS=50; effective = 100 * (1 - 50*0.6/100) = 100 * 0.70 = 70.
	s := makeStakeholder(config.RoleEnergy, -80.0, 50.0, true)
	conflict := IdeologyConflict(config.PolicyCardDef{Sector: config.PolicySectorIndustry}, s)
	assert.InDelta(t, 70.0, conflict, 0.01)
}

// ---------------------------------------------------------------------------
// EvaluateApprovalStep
// ---------------------------------------------------------------------------

func TestEvaluateApprovalStep_AllMet_ReturnsApproved(t *testing.T) {
	s := makeStakeholder(config.RoleEnergy, 0.0, 70.0, true)
	req := config.ApprovalRequirement{
		Role:                 config.RoleEnergy,
		MinRelationshipScore: 50.0,
		MaxIdeologyConflict:  60.0,
	}
	card := makeCard(config.PolicySectorPower, nil)
	card.State = PolicyStateUnderReview
	approved, hardReject := EvaluateApprovalStep(card, &config.PolicyCardDef{Sector: config.PolicySectorPower}, s, req)
	assert.True(t, approved)
	assert.False(t, hardReject)
}

func TestEvaluateApprovalStep_HighIdeologyConflict_HardReject(t *testing.T) {
	// Far-left stakeholder, industry policy (ideology position +20): conflict = 100
	s := makeStakeholder(config.RoleEnergy, -80.0, 80.0, true)
	req := config.ApprovalRequirement{
		Role:                 config.RoleEnergy,
		MinRelationshipScore: 50.0,
		MaxIdeologyConflict:  60.0,
	}
	card := makeCard(config.PolicySectorIndustry, nil)
	card.State = PolicyStateUnderReview
	approved, hardReject := EvaluateApprovalStep(card, &config.PolicyCardDef{Sector: config.PolicySectorIndustry}, s, req)
	assert.False(t, approved)
	assert.True(t, hardReject)
}

func TestEvaluateApprovalStep_LowRelationship_PendingNotReject(t *testing.T) {
	s := makeStakeholder(config.RoleEnergy, 0.0, 30.0, true)
	req := config.ApprovalRequirement{
		Role:                 config.RoleEnergy,
		MinRelationshipScore: 60.0,
		MaxIdeologyConflict:  80.0,
	}
	card := makeCard(config.PolicySectorPower, nil)
	card.State = PolicyStateUnderReview
	approved, hardReject := EvaluateApprovalStep(card, &config.PolicyCardDef{Sector: config.PolicySectorPower}, s, req)
	assert.False(t, approved)
	assert.False(t, hardReject)
}

// ---------------------------------------------------------------------------
// EvaluateApprovalStep -- significance-based refusal (D3)
// ---------------------------------------------------------------------------

func TestEvaluateApprovalStep_MajorSignificance_HighConflictAndStalledWeeks_HardRejects(t *testing.T) {
	// Far-left stakeholder (ideology=-80, NZS=0): Cross sector pos=0 -> raw=80.
	// effective = 80 * (1 - 0*0.6/100) = 80 > majorSignificanceRefuseConflict=75.
	// WeeksUnderReview = 8 >= majorSignificanceRefuseWeeks -> hard rejects.
	// NZS=0 is used here to ensure effective conflict equals raw conflict.
	s := stakeholder.Stakeholder{
		ID:                "s1",
		Role:              config.RoleLeader,
		IdeologyScore:     -80.0,
		RelationshipScore: 70.0,
		NetZeroSympathy:   0.0, // zero sympathy: no NZS reduction
		IsUnlocked:        true,
		State:             stakeholder.MinisterStateActive,
	}
	req := config.ApprovalRequirement{
		Role:                 config.RoleLeader,
		MinRelationshipScore: 40.0,
		MaxIdeologyConflict:  200.0, // high threshold so per-step check does not fire
	}
	def := &config.PolicyCardDef{
		Sector:       config.PolicySectorCross,
		Significance: config.PolicySignificanceMajor,
	}
	card := PolicyCard{Def: def, State: PolicyStateUnderReview, WeeksUnderReview: 8}
	approved, hardReject := EvaluateApprovalStep(card, def, s, req)
	assert.False(t, approved)
	assert.True(t, hardReject)
}

func TestEvaluateApprovalStep_MajorSignificance_HighConflictButNotStalled_DoesNotHardReject(t *testing.T) {
	// Same setup as above but WeeksUnderReview = 7 (below threshold)
	s := makeStakeholder(config.RoleLeader, -80.0, 70.0, true)
	req := config.ApprovalRequirement{
		Role:                 config.RoleLeader,
		MinRelationshipScore: 40.0,
		MaxIdeologyConflict:  200.0,
	}
	def := &config.PolicyCardDef{
		Sector:       config.PolicySectorCross,
		Significance: config.PolicySignificanceMajor,
	}
	card := PolicyCard{Def: def, State: PolicyStateUnderReview, WeeksUnderReview: 7}
	_, hardReject := EvaluateApprovalStep(card, def, s, req)
	assert.False(t, hardReject)
}

func TestEvaluateApprovalStep_MinorSignificance_HighConflictAndStalledWeeks_NoSignificanceRefuse(t *testing.T) {
	// MINOR significance: significance-based refusal does not apply regardless of conflict/weeks
	s := makeStakeholder(config.RoleLeader, -80.0, 70.0, true)
	req := config.ApprovalRequirement{
		Role:                 config.RoleLeader,
		MinRelationshipScore: 40.0,
		MaxIdeologyConflict:  200.0,
	}
	def := &config.PolicyCardDef{
		Sector:       config.PolicySectorCross,
		Significance: config.PolicySignificanceMinor,
	}
	card := PolicyCard{Def: def, State: PolicyStateUnderReview, WeeksUnderReview: 20}
	_, hardReject := EvaluateApprovalStep(card, def, s, req)
	assert.False(t, hardReject)
}

// ---------------------------------------------------------------------------
// EvaluateApproval -- pipeline state machine
// ---------------------------------------------------------------------------

func TestEvaluateApproval_AllStepsPass_MovesToApproved(t *testing.T) {
	steps := []config.ApprovalRequirement{
		{Role: config.RoleEnergy, MinRelationshipScore: 40.0, MaxIdeologyConflict: 80.0},
		{Role: config.RoleChancellor, MinRelationshipScore: 40.0, MaxIdeologyConflict: 80.0},
	}
	card := makeCard(config.PolicySectorPower, steps)
	card.State = PolicyStateUnderReview

	stakeholders := []stakeholder.Stakeholder{
		makeStakeholder(config.RoleEnergy, 0.0, 70.0, true),
		makeStakeholder(config.RoleChancellor, 0.0, 70.0, true),
	}
	result := EvaluateApproval(card, stakeholders)
	assert.Equal(t, PolicyStateApproved, result.State)
	assert.Equal(t, 2, result.StepsCleared)
}

func TestEvaluateApproval_HardReject_MovesToRejected(t *testing.T) {
	steps := []config.ApprovalRequirement{
		{Role: config.RoleEnergy, MinRelationshipScore: 40.0, MaxIdeologyConflict: 50.0},
	}
	card := makeCard(config.PolicySectorIndustry, steps)
	card.State = PolicyStateUnderReview

	// Far-left Energy minister; Industry policy conflict = 100 > 50 threshold
	stakeholders := []stakeholder.Stakeholder{
		makeStakeholder(config.RoleEnergy, -80.0, 80.0, true),
	}
	result := EvaluateApproval(card, stakeholders)
	assert.Equal(t, PolicyStateRejected, result.State)
}

func TestEvaluateApproval_RelationshipShortfall_StaysUnderReview(t *testing.T) {
	steps := []config.ApprovalRequirement{
		{Role: config.RoleEnergy, MinRelationshipScore: 70.0, MaxIdeologyConflict: 80.0},
	}
	card := makeCard(config.PolicySectorPower, steps)
	card.State = PolicyStateUnderReview

	stakeholders := []stakeholder.Stakeholder{
		makeStakeholder(config.RoleEnergy, 0.0, 40.0, true), // relationship too low
	}
	result := EvaluateApproval(card, stakeholders)
	assert.Equal(t, PolicyStateUnderReview, result.State)
	assert.Equal(t, 0, result.StepsCleared)
}

func TestEvaluateApproval_LockedStakeholder_StaysUnderReview(t *testing.T) {
	steps := []config.ApprovalRequirement{
		{Role: config.RoleEnergy, MinRelationshipScore: 40.0, MaxIdeologyConflict: 80.0},
	}
	card := makeCard(config.PolicySectorPower, steps)
	card.State = PolicyStateUnderReview

	stakeholders := []stakeholder.Stakeholder{
		makeStakeholder(config.RoleEnergy, 0.0, 70.0, false), // not unlocked
	}
	result := EvaluateApproval(card, stakeholders)
	assert.Equal(t, PolicyStateUnderReview, result.State)
}

func TestEvaluateApproval_NonUnderReview_IsNoop(t *testing.T) {
	card := makeCard(config.PolicySectorPower, nil)
	card.State = PolicyStateActive
	result := EvaluateApproval(card, nil)
	assert.Equal(t, PolicyStateActive, result.State)
}

// ---------------------------------------------------------------------------
// ActivatePolicy / ArchivePolicy
// ---------------------------------------------------------------------------

func TestActivatePolicy_Approved_MovesToActive(t *testing.T) {
	card := makeCard(config.PolicySectorPower, nil)
	card.State = PolicyStateApproved
	card = ActivatePolicy(card)
	assert.Equal(t, PolicyStateActive, card.State)
}

func TestActivatePolicy_NotApproved_IsNoop(t *testing.T) {
	card := makeCard(config.PolicySectorPower, nil)
	card.State = PolicyStateUnderReview
	card = ActivatePolicy(card)
	assert.Equal(t, PolicyStateUnderReview, card.State)
}

func TestArchivePolicy_Active_MovesToArchived(t *testing.T) {
	card := makeCard(config.PolicySectorPower, nil)
	card.State = PolicyStateActive
	card = ArchivePolicy(card, ArchiveReasonCancelled)
	assert.Equal(t, PolicyStateArchived, card.State)
	assert.Equal(t, ArchiveReasonCancelled, card.ArchiveReason)
}

func TestArchivePolicy_AlreadyArchived_IsNoop(t *testing.T) {
	card := makeCard(config.PolicySectorPower, nil)
	card.State = PolicyStateArchived
	card.ArchiveReason = ArchiveReasonCompleted
	card = ArchivePolicy(card, ArchiveReasonCancelled)
	assert.Equal(t, ArchiveReasonCompleted, card.ArchiveReason)
}

// ---------------------------------------------------------------------------
// ResolveWeeklyEffect
// ---------------------------------------------------------------------------

func activeCard(flags struct{ cap, tech, retrofit bool }) PolicyCard {
	card := makeCard(config.PolicySectorPower, nil)
	card.Def.WeeklyEffect = config.WeeklyEffectDef{
		Sector:            config.PolicySectorPower,
		BaseCarbonDeltaMt: -2.0,
		BudgetCostPerWeek: 10.0,
		CapacityDependent: flags.cap,
		TechDependent:     flags.tech,
		RetrofitDependent: flags.retrofit,
	}
	card.State = PolicyStateActive
	return card
}

func makeRegion(capacity float64) region.Region {
	return region.Region{InstallerCapacity: capacity}
}

func TestResolveWeeklyEffect_NotActive_ReturnsZero(t *testing.T) {
	card := makeCard(config.PolicySectorPower, nil)
	card.State = PolicyStateDraft
	delta := ResolveWeeklyEffect(card, makeRegion(50.0), 1.0, 1.0)
	assert.Equal(t, 0.0, delta.DeltaMt)
	assert.Equal(t, 0.0, delta.BudgetCostPerWeek)
}

func TestResolveWeeklyEffect_NoFlags_UsesBase(t *testing.T) {
	card := activeCard(struct{ cap, tech, retrofit bool }{false, false, false})
	delta := ResolveWeeklyEffect(card, makeRegion(50.0), 0.5, 0.5)
	assert.InDelta(t, -2.0, delta.DeltaMt, 0.001)
	assert.InDelta(t, 10.0, delta.BudgetCostPerWeek, 0.001)
}

func TestResolveWeeklyEffect_CapacityDependent_ScalesWithRegion(t *testing.T) {
	card := activeCard(struct{ cap, tech, retrofit bool }{true, false, false})
	// referenceInstallerCapacity = 50; a region with 25 installs => fraction 0.5
	delta := ResolveWeeklyEffect(card, makeRegion(25.0), 1.0, 1.0)
	assert.InDelta(t, -1.0, delta.DeltaMt, 0.001)
}

func TestResolveWeeklyEffect_TechDependent_ScalesWithMaturity(t *testing.T) {
	card := activeCard(struct{ cap, tech, retrofit bool }{false, true, false})
	delta := ResolveWeeklyEffect(card, makeRegion(50.0), 0.25, 1.0)
	assert.InDelta(t, -0.5, delta.DeltaMt, 0.001)
}

func TestResolveWeeklyEffect_RetrofitDependent_UsesToRetrofitRate(t *testing.T) {
	card := activeCard(struct{ cap, tech, retrofit bool }{false, false, true})
	delta := ResolveWeeklyEffect(card, makeRegion(50.0), 1.0, 0.4)
	assert.InDelta(t, -0.8, delta.DeltaMt, 0.001)
}

func TestResolveWeeklyEffect_AllFlags_MultipliersStack(t *testing.T) {
	card := activeCard(struct{ cap, tech, retrofit bool }{true, true, true})
	// capacity 25/50=0.5, tech 0.8, retrofit 0.6 => combined 0.5*0.8*0.6=0.24
	delta := ResolveWeeklyEffect(card, makeRegion(25.0), 0.8, 0.6)
	assert.InDelta(t, -2.0*0.24, delta.DeltaMt, 0.001)
}

func TestResolveWeeklyEffect_CorrectSector(t *testing.T) {
	card := activeCard(struct{ cap, tech, retrofit bool }{false, false, false})
	card.Def.WeeklyEffect.Sector = config.PolicySectorTransport
	delta := ResolveWeeklyEffect(card, makeRegion(50.0), 1.0, 1.0)
	assert.Equal(t, config.PolicySectorTransport, delta.Sector)
}
