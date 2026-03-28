package player

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// NewCivilServant
// ---------------------------------------------------------------------------

func TestNewCivilServant_InitialAPEqualsBase(t *testing.T) {
	cs := NewCivilServant()
	assert.Equal(t, baseAPPool, cs.APRemaining)
}

func TestNewCivilServant_EmptyStaffAndHistory(t *testing.T) {
	cs := NewCivilServant()
	assert.Empty(t, cs.Staff)
	assert.Empty(t, cs.ActionHistory)
}

// ---------------------------------------------------------------------------
// WeeklyAPPool
// ---------------------------------------------------------------------------

func TestWeeklyAPPool_NoStaff_ReturnsBase(t *testing.T) {
	cs := NewCivilServant()
	assert.Equal(t, baseAPPool, WeeklyAPPool(cs))
}

func TestWeeklyAPPool_WithStaff_IncludesBonuses(t *testing.T) {
	cs := NewCivilServant()
	cs = HireStaff(cs, StaffMember{ID: "s1", Role: StaffRoleAnalyst, APBonus: 1})
	cs = HireStaff(cs, StaffMember{ID: "s2", Role: StaffRoleAdvisor, APBonus: 1})
	assert.Equal(t, baseAPPool+2, WeeklyAPPool(cs))
}

// ---------------------------------------------------------------------------
// StartWeekAPPool
// ---------------------------------------------------------------------------

func TestStartWeekAPPool_ResetsToFullPool(t *testing.T) {
	cs := NewCivilServant()
	cs = HireStaff(cs, StaffMember{ID: "s1", Role: StaffRoleChiefOfStaff, APBonus: 2})
	cs.APRemaining = 0
	cs = StartWeekAPPool(cs)
	assert.Equal(t, baseAPPool+2, cs.APRemaining)
}

// ---------------------------------------------------------------------------
// SpendAP
// ---------------------------------------------------------------------------

func TestSpendAP_SufficientAP_DeductsAndReturnsTrue(t *testing.T) {
	cs := NewCivilServant()
	cs.APRemaining = 5
	updated, ok := SpendAP(cs, 3)
	assert.True(t, ok)
	assert.Equal(t, 2, updated.APRemaining)
}

func TestSpendAP_InsufficientAP_ReturnsFalseUnchanged(t *testing.T) {
	cs := NewCivilServant()
	cs.APRemaining = 2
	updated, ok := SpendAP(cs, 5)
	assert.False(t, ok)
	assert.Equal(t, 2, updated.APRemaining)
}

func TestSpendAP_ExactAmount_ZeroRemaining(t *testing.T) {
	cs := NewCivilServant()
	cs.APRemaining = 3
	updated, ok := SpendAP(cs, 3)
	assert.True(t, ok)
	assert.Equal(t, 0, updated.APRemaining)
}

// ---------------------------------------------------------------------------
// RecordAction
// ---------------------------------------------------------------------------

func TestRecordAction_AppendsToHistory(t *testing.T) {
	cs := NewCivilServant()
	record := ActionRecord{ActionType: ActionTypeSubmitPolicy, Week: 5, APCost: 2}
	cs = RecordAction(cs, record)
	assert.Len(t, cs.ActionHistory, 1)
	assert.Equal(t, ActionTypeSubmitPolicy, cs.ActionHistory[0].ActionType)
}

func TestRecordAction_MultipleRecords_PreservesOrder(t *testing.T) {
	cs := NewCivilServant()
	cs = RecordAction(cs, ActionRecord{Week: 1})
	cs = RecordAction(cs, ActionRecord{Week: 2})
	cs = RecordAction(cs, ActionRecord{Week: 3})
	assert.Len(t, cs.ActionHistory, 3)
	assert.Equal(t, 1, cs.ActionHistory[0].Week)
	assert.Equal(t, 3, cs.ActionHistory[2].Week)
}

// ---------------------------------------------------------------------------
// HireStaff / FireStaff
// ---------------------------------------------------------------------------

func TestHireStaff_AddsToRoster(t *testing.T) {
	cs := NewCivilServant()
	cs = HireStaff(cs, StaffMember{ID: "s1", Role: StaffRoleAnalyst, APBonus: 1})
	assert.Len(t, cs.Staff, 1)
	assert.Equal(t, "s1", cs.Staff[0].ID)
}

func TestFireStaff_RemovesFromRoster(t *testing.T) {
	cs := NewCivilServant()
	cs = HireStaff(cs, StaffMember{ID: "s1", Role: StaffRoleAnalyst, APBonus: 1})
	cs = HireStaff(cs, StaffMember{ID: "s2", Role: StaffRoleAdvisor, APBonus: 1})
	cs = FireStaff(cs, "s1")
	assert.Len(t, cs.Staff, 1)
	assert.Equal(t, "s2", cs.Staff[0].ID)
}

func TestFireStaff_UnknownID_IsNoop(t *testing.T) {
	cs := NewCivilServant()
	cs = HireStaff(cs, StaffMember{ID: "s1", Role: StaffRoleAnalyst, APBonus: 1})
	cs = FireStaff(cs, "nonexistent")
	assert.Len(t, cs.Staff, 1)
}

// ---------------------------------------------------------------------------
// TotalWeeklyStaffCost
// ---------------------------------------------------------------------------

func TestTotalWeeklyStaffCost_NoStaff_Zero(t *testing.T) {
	cs := NewCivilServant()
	assert.Equal(t, 0.0, TotalWeeklyStaffCost(cs))
}

func TestTotalWeeklyStaffCost_WithStaff_SumsCosts(t *testing.T) {
	cs := NewCivilServant()
	cs = HireStaff(cs, StaffMember{ID: "s1", Role: StaffRoleAnalyst, APBonus: 1})
	cs = HireStaff(cs, StaffMember{ID: "s2", Role: StaffRoleAdvisor, APBonus: 1})
	// Analyst = 2.0, Advisor = 3.0
	assert.InDelta(t, 5.0, TotalWeeklyStaffCost(cs), 0.001)
}

// ---------------------------------------------------------------------------
// ReputationGrade
// ---------------------------------------------------------------------------

func TestReputationGrade_Boundaries(t *testing.T) {
	cases := []struct {
		rep   float64
		grade string
	}{
		{0.0, "Generalist"},
		{14.9, "Generalist"},
		{15.0, "Higher Executive Officer"},
		{27.9, "Higher Executive Officer"},
		{28.0, "Senior Executive Officer"},
		{40.0, "Grade 7"},
		{52.0, "Deputy Director"},
		{63.0, "Director"},
		{74.0, "Director General"},
		{85.0, "Deputy Secretary"},
		{95.0, "Permanent Secretary"},
		{100.0, "Permanent Secretary"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.grade, ReputationGrade(tc.rep), "rep=%.1f", tc.rep)
	}
}

func TestReputationGrade_ClampedBelow(t *testing.T) {
	assert.Equal(t, "Generalist", ReputationGrade(-10.0))
}

func TestReputationGrade_ClampedAbove(t *testing.T) {
	assert.Equal(t, "Permanent Secretary", ReputationGrade(110.0))
}
