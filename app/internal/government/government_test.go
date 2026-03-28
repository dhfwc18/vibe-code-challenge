package government

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/stakeholder"
)

func makeStakeholder(id string, ideo, nzs float64, role config.Role) stakeholder.Stakeholder {
	return stakeholder.Stakeholder{
		ID:              id,
		Role:            role,
		IdeologyScore:   ideo,
		NetZeroSympathy: nzs,
		RelationshipScore: 50,
	}
}

// ---------------------------------------------------------------------------
// NewGovernment
// ---------------------------------------------------------------------------

func TestNewGovernment_RulingPartySet(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	assert.Equal(t, config.PartyLeft, g.RulingParty)
}

func TestNewGovernment_CabinetEmpty(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	assert.Equal(t, 0, len(g.CabinetByRole))
}

func TestNewGovernment_TermNumberIsOne(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	assert.Equal(t, 1, g.TermNumber)
}

func TestNewGovernment_ElectionDueWeekSet(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	assert.Equal(t, 260, g.ElectionDueWeek)
}

func TestNewGovernment_PhaseIsStable(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	assert.Equal(t, GovernmentPhaseStable, g.Phase)
}

func TestTriggerElection_PhaseResetsToStable(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	g.Phase = GovernmentPhaseElectionCampaign
	g2 := TriggerElection(g, config.PartyRight, 520)
	assert.Equal(t, GovernmentPhaseStable, g2.Phase)
}

// ---------------------------------------------------------------------------
// AssignMinister
// ---------------------------------------------------------------------------

func TestAssignMinister_AssignsID(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	g2 := AssignMinister(g, config.RoleEnergy, "alice")
	assert.Equal(t, "alice", g2.CabinetByRole[config.RoleEnergy])
}

func TestAssignMinister_OverwritesExistingRole(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	g = AssignMinister(g, config.RoleEnergy, "alice")
	g2 := AssignMinister(g, config.RoleEnergy, "bob")
	assert.Equal(t, "bob", g2.CabinetByRole[config.RoleEnergy])
}

func TestAssignMinister_DoesNotMutateInput(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	AssignMinister(g, config.RoleEnergy, "alice")
	_, ok := g.CabinetByRole[config.RoleEnergy]
	assert.False(t, ok)
}

func TestAssignMinister_OtherRolesUnchanged(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	g = AssignMinister(g, config.RoleLeader, "alice")
	g2 := AssignMinister(g, config.RoleEnergy, "bob")
	assert.Equal(t, "alice", g2.CabinetByRole[config.RoleLeader])
}

// ---------------------------------------------------------------------------
// RemoveMinister
// ---------------------------------------------------------------------------

func TestRemoveMinister_RemovesRole(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	g = AssignMinister(g, config.RoleEnergy, "alice")
	g2 := RemoveMinister(g, config.RoleEnergy)
	_, ok := g2.CabinetByRole[config.RoleEnergy]
	assert.False(t, ok)
}

func TestRemoveMinister_MissingRole_NoOp(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	g2 := RemoveMinister(g, config.RoleEnergy)
	assert.Equal(t, 0, len(g2.CabinetByRole))
}

func TestRemoveMinister_DoesNotMutateInput(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	g = AssignMinister(g, config.RoleEnergy, "alice")
	RemoveMinister(g, config.RoleEnergy)
	assert.Equal(t, "alice", g.CabinetByRole[config.RoleEnergy])
}

func TestRemoveMinister_OtherRolesPreserved(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	g = AssignMinister(g, config.RoleLeader, "alice")
	g = AssignMinister(g, config.RoleEnergy, "bob")
	g2 := RemoveMinister(g, config.RoleEnergy)
	assert.Equal(t, "alice", g2.CabinetByRole[config.RoleLeader])
}

// ---------------------------------------------------------------------------
// IsElectionDue
// ---------------------------------------------------------------------------

func TestIsElectionDue_WeekAtDue_ReturnsTrue(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	assert.True(t, IsElectionDue(g, 260))
}

func TestIsElectionDue_WeekBeforeDue_ReturnsFalse(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	assert.False(t, IsElectionDue(g, 259))
}

func TestIsElectionDue_WeekAfterDue_ReturnsTrue(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	assert.True(t, IsElectionDue(g, 300))
}

// ---------------------------------------------------------------------------
// ComputeMinisterStats
// ---------------------------------------------------------------------------

func TestComputeMinisterStats_PositiveLCRDelta_PositivePopularity(t *testing.T) {
	s := makeStakeholder("a", 0, 50, config.RoleEnergy)
	stats := ComputeMinisterStats(s, 5.0)
	assert.Greater(t, stats.PopularityModifier, 0.0)
}

func TestComputeMinisterStats_NegativeLCRDelta_NegativePopularity(t *testing.T) {
	s := makeStakeholder("a", 0, 50, config.RoleEnergy)
	stats := ComputeMinisterStats(s, -5.0)
	assert.Less(t, stats.PopularityModifier, 0.0)
}

func TestComputeMinisterStats_HighNetZeroSympathy_BonusPopularity(t *testing.T) {
	sHigh := makeStakeholder("a", 0, 80, config.RoleEnergy)
	sNeutral := makeStakeholder("b", 0, 50, config.RoleEnergy)
	assert.Greater(t,
		ComputeMinisterStats(sHigh, 0).PopularityModifier,
		ComputeMinisterStats(sNeutral, 0).PopularityModifier,
	)
}

func TestComputeMinisterStats_LowNetZeroSympathy_PenaltyPopularity(t *testing.T) {
	sLow := makeStakeholder("a", 0, 20, config.RoleEnergy)
	sNeutral := makeStakeholder("b", 0, 50, config.RoleEnergy)
	assert.Less(t,
		ComputeMinisterStats(sLow, 0).PopularityModifier,
		ComputeMinisterStats(sNeutral, 0).PopularityModifier,
	)
}

func TestComputeMinisterStats_RightIdeology_IndustryBiasAboveNeutral(t *testing.T) {
	s := makeStakeholder("a", 80, 50, config.RoleEnergy) // right-leaning
	stats := ComputeMinisterStats(s, 0)
	assert.Greater(t, stats.BudgetAllocationBias[DeptIndustry], neutralBias)
}

func TestComputeMinisterStats_LeftIdeology_BuildingsBiasAboveNeutral(t *testing.T) {
	s := makeStakeholder("a", -80, 50, config.RoleEnergy) // left-leaning
	stats := ComputeMinisterStats(s, 0)
	assert.Greater(t, stats.BudgetAllocationBias[DeptBuildings], neutralBias)
}

func TestComputeMinisterStats_AllBiasesInRange(t *testing.T) {
	depts := []string{DeptPower, DeptTransport, DeptBuildings, DeptIndustry, DeptCross}
	for ideo := -100.0; ideo <= 100.0; ideo += 50 {
		for nzs := 0.0; nzs <= 100.0; nzs += 50 {
			s := makeStakeholder("a", ideo, nzs, config.RoleEnergy)
			stats := ComputeMinisterStats(s, 0)
			for _, dept := range depts {
				bias := stats.BudgetAllocationBias[dept]
				assert.GreaterOrEqual(t, bias, minBiasMultiplier,
					"dept=%s ideo=%.0f nzs=%.0f", dept, ideo, nzs)
				assert.LessOrEqual(t, bias, maxBiasMultiplier,
					"dept=%s ideo=%.0f nzs=%.0f", dept, ideo, nzs)
			}
		}
	}
}

func TestComputeMinisterStats_DoesNotMutateInput(t *testing.T) {
	s := makeStakeholder("a", 0, 50, config.RoleEnergy)
	ComputeMinisterStats(s, 5.0)
	assert.InDelta(t, 50.0, s.RelationshipScore, 0.001)
}

// ---------------------------------------------------------------------------
// TriggerElection
// ---------------------------------------------------------------------------

func TestTriggerElection_NewRulingParty(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	g2 := TriggerElection(g, config.PartyRight, 520)
	assert.Equal(t, config.PartyRight, g2.RulingParty)
}

func TestTriggerElection_TermNumberIncremented(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	g2 := TriggerElection(g, config.PartyRight, 520)
	assert.Equal(t, 2, g2.TermNumber)
}

func TestTriggerElection_CabinetCleared(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	g = AssignMinister(g, config.RoleEnergy, "alice")
	g2 := TriggerElection(g, config.PartyRight, 520)
	assert.Equal(t, 0, len(g2.CabinetByRole))
}

func TestTriggerElection_NewElectionDueWeekSet(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	g2 := TriggerElection(g, config.PartyRight, 520)
	assert.Equal(t, 520, g2.ElectionDueWeek)
}

func TestTriggerElection_DoesNotMutateInput(t *testing.T) {
	g := NewGovernment(config.PartyLeft, 260)
	TriggerElection(g, config.PartyRight, 520)
	assert.Equal(t, config.PartyLeft, g.RulingParty)
	assert.Equal(t, 1, g.TermNumber)
}
