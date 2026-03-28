package simulation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twenty-fifty/internal/config"
	"twenty-fifty/internal/player"
	"twenty-fifty/internal/policy"
	"twenty-fifty/internal/save"
	"twenty-fifty/internal/stakeholder"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

var fixedSeed = save.MasterSeed(42)

func loadWorld(t *testing.T) WorldState {
	t.Helper()
	cfg, err := config.Load()
	require.NoError(t, err, "config.Load failed")
	return NewWorld(cfg, fixedSeed)
}

// advanceN advances the world n weeks with no player input.
func advanceN(w WorldState, n int) WorldState {
	for i := 0; i < n; i++ {
		w, _ = AdvanceWeek(w, nil)
	}
	return w
}

// ---------------------------------------------------------------------------
// NewWorld
// ---------------------------------------------------------------------------

func TestNewWorld_WeekZero(t *testing.T) {
	w := loadWorld(t)
	assert.Equal(t, 0, w.Week)
	assert.Equal(t, 2010, w.Year)
	assert.Equal(t, 1, w.Quarter)
}

func TestNewWorld_StakeholdersSeeded(t *testing.T) {
	w := loadWorld(t)
	assert.NotEmpty(t, w.Stakeholders)
	unlocked := 0
	for _, s := range w.Stakeholders {
		if s.IsUnlocked {
			unlocked++
		}
	}
	assert.Greater(t, unlocked, 0, "at least some stakeholders must be unlocked at start")
}

func TestNewWorld_PoliciesAllDraft(t *testing.T) {
	w := loadWorld(t)
	assert.NotEmpty(t, w.PolicyCards, "config must have at least one policy card")
	for _, card := range w.PolicyCards {
		assert.Equal(t, policy.PolicyStateDraft, card.State,
			"policy %q should start in DRAFT", card.Def.ID)
	}
}

func TestNewWorld_CarbonZero(t *testing.T) {
	w := loadWorld(t)
	assert.Equal(t, 0.0, w.Carbon.CumulativeStock)
	assert.Equal(t, 0.0, w.Carbon.RunningAnnualTotal)
}

func TestNewWorld_RulingPartyAssigned(t *testing.T) {
	w := loadWorld(t)
	assert.Equal(t, config.PartyLeft, w.Government.RulingParty)
	assert.NotEmpty(t, w.Government.CabinetByRole,
		"initial cabinet must have at least one minister assigned")
}

// ---------------------------------------------------------------------------
// AdvanceWeek -- clock
// ---------------------------------------------------------------------------

func TestAdvanceWeek_IncrementsWeek(t *testing.T) {
	w := loadWorld(t)
	w, _ = AdvanceWeek(w, nil)
	assert.Equal(t, 1, w.Week)
}

func TestAdvanceWeek_YearChangeAtWeek53(t *testing.T) {
	w := loadWorld(t)
	w = advanceN(w, 53)
	assert.Equal(t, 2011, w.Year, "year should be 2011 after 53 weeks")
}

func TestAdvanceWeek_QuarterChangeAtWeek14(t *testing.T) {
	w := loadWorld(t)
	w = advanceN(w, 14)
	assert.Equal(t, 2, w.Quarter, "quarter should be 2 after 14 weeks")
}

func TestAdvanceWeek_QuarterResets(t *testing.T) {
	w := loadWorld(t)
	w = advanceN(w, 52)
	assert.Equal(t, 1, w.Quarter, "quarter should reset to 1 at week 52 (start of year 2)")
}

// ---------------------------------------------------------------------------
// AdvanceWeek -- carbon
// ---------------------------------------------------------------------------

func TestAdvanceWeek_CarbonAccumulatesFirstWeek(t *testing.T) {
	w := loadWorld(t)
	w, _ = AdvanceWeek(w, nil)
	assert.Greater(t, w.Carbon.CumulativeStock, 0.0,
		"carbon must accumulate after first week with no policies active")
}

func TestAdvanceWeek_CarbonMonotoneWithNoPolicies(t *testing.T) {
	w := loadWorld(t)
	prev := w.Carbon.CumulativeStock
	for i := 0; i < 10; i++ {
		w, _ = AdvanceWeek(w, nil)
		assert.GreaterOrEqual(t, w.Carbon.CumulativeStock, prev,
			"cumulative carbon must not decrease with no active policies (week %d)", i+1)
		prev = w.Carbon.CumulativeStock
	}
}

// ---------------------------------------------------------------------------
// AdvanceWeek -- stakeholder transitions
// ---------------------------------------------------------------------------

func TestAdvanceWeek_AppointedBecomesActiveAfterFirstWeek(t *testing.T) {
	w := loadWorld(t)
	w, _ = AdvanceWeek(w, nil)
	for _, s := range w.Stakeholders {
		if s.IsUnlocked {
			assert.NotEqual(t, stakeholder.MinisterStateAppointed, s.State,
				"stakeholder %q must not remain APPOINTED after first week", s.ID)
		}
	}
}

// ---------------------------------------------------------------------------
// AdvanceWeek -- player actions
// ---------------------------------------------------------------------------

func TestAdvanceWeek_SubmitPolicy_MovesToUnderReview(t *testing.T) {
	w := loadWorld(t)
	// Find any policy card that is unlocked (no tech gate or gate already met).
	var targetID string
	for _, card := range w.PolicyCards {
		if policy.IsUnlocked(card, w.Tech.Maturities) {
			targetID = card.Def.ID
			break
		}
	}
	require.NotEmpty(t, targetID, "need at least one unlocked policy card to test")

	w, _ = AdvanceWeek(w, []Action{
		{Type: player.ActionTypeSubmitPolicy, Target: targetID},
	})

	for _, card := range w.PolicyCards {
		if card.Def.ID == targetID {
			assert.Equal(t, policy.PolicyStateUnderReview, card.State,
				"submitted policy should be UNDER_REVIEW")
			return
		}
	}
	t.Fatalf("policy %q not found after submit", targetID)
}

func TestAdvanceWeek_HireStaff_IncreasesAPBonus(t *testing.T) {
	w := loadWorld(t)
	basePool := player.WeeklyAPPool(w.Player)

	w, _ = AdvanceWeek(w, []Action{
		{
			Type:   player.ActionTypeHireStaff,
			Target: string(player.StaffRoleAnalyst),
			Detail: "analyst_01",
		},
	})

	assert.Greater(t, player.WeeklyAPPool(w.Player), basePool,
		"hiring staff should increase the weekly AP pool")
}

// ---------------------------------------------------------------------------
// HeadlessRun -- integration gate (100-week run)
// ---------------------------------------------------------------------------

func TestHeadlessRun_100Weeks_NoPanic(t *testing.T) {
	w := loadWorld(t)
	w, _ = HeadlessRun(w, 100)
	assert.Equal(t, 100, w.Week, "world should be at week 100 after HeadlessRun(100)")
}

func TestHeadlessRun_100Weeks_CarbonAccumulated(t *testing.T) {
	w := loadWorld(t)
	_, report := HeadlessRun(w, 100)
	assert.Greater(t, report.FinalCarbon.CumulativeStock, 0.0,
		"cumulative carbon stock must be positive after 100 weeks")
}

func TestHeadlessRun_100Weeks_AtLeastOnePollGenerated(t *testing.T) {
	w := loadWorld(t)
	_, report := HeadlessRun(w, 100)
	assert.Greater(t, report.PollsTaken, 0,
		"at least one poll must be generated over 100 weeks")
}

func TestHeadlessRun_100Weeks_AtLeastOneEventFired(t *testing.T) {
	w := loadWorld(t)
	_, report := HeadlessRun(w, 100)
	assert.Greater(t, report.EventsFired, 0,
		"at least one event must fire over 100 weeks")
}

func TestHeadlessRun_100Weeks_StakeholderStatesValid(t *testing.T) {
	w := loadWorld(t)
	_, report := HeadlessRun(w, 100)
	assert.Empty(t, report.StakeholderIssues,
		"no invalid stakeholder states: %v", report.StakeholderIssues)
}

func TestHeadlessRun_100Weeks_NoBudgetNegative(t *testing.T) {
	w := loadWorld(t)
	_, report := HeadlessRun(w, 100)
	assert.GreaterOrEqual(t, report.MinBudgetValueGBP, 0.0,
		"all department budget allocations must be non-negative")
}

func TestHeadlessRun_100Weeks_GovernmentPopularityInBounds(t *testing.T) {
	w := loadWorld(t)
	_, report := HeadlessRun(w, 100)
	assert.GreaterOrEqual(t, report.FinalGovtPop, 0.0)
	assert.LessOrEqual(t, report.FinalGovtPop, 100.0)
}

func TestHeadlessRun_100Weeks_LCRInBounds(t *testing.T) {
	w := loadWorld(t)
	_, report := HeadlessRun(w, 100)
	assert.GreaterOrEqual(t, report.FinalLCR, 0.0)
	assert.LessOrEqual(t, report.FinalLCR, 100.0)
}
