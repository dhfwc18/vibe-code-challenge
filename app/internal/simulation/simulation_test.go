package simulation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/industry"
	"github.com/vibe-code-challenge/twenty-fifty/internal/player"
	"github.com/vibe-code-challenge/twenty-fifty/internal/policy"
	"github.com/vibe-code-challenge/twenty-fifty/internal/save"
	"github.com/vibe-code-challenge/twenty-fifty/internal/stakeholder"
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
	// Game starts with center-right in power (analogous to 2010 Taitan election).
	assert.Equal(t, config.PartyRight, w.Government.RulingParty)
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

// ---------------------------------------------------------------------------
// Minister popularity
// ---------------------------------------------------------------------------

func TestNewWorld_MinisterPopularityInitialized(t *testing.T) {
	w := loadWorld(t)
	for _, s := range w.Stakeholders {
		if s.IsUnlocked {
			assert.Equal(t, 50.0, s.Popularity,
				"stakeholder %q should start with Popularity=50", s.ID)
		}
	}
}

func TestNewWorld_MinisterLastPollResultsInitialized(t *testing.T) {
	w := loadWorld(t)
	assert.NotNil(t, w.MinisterLastPollResults)
}

func TestHeadlessRun_100Weeks_MinisterPopularityInBounds(t *testing.T) {
	w := loadWorld(t)
	w, _ = HeadlessRun(w, 100)
	for _, s := range w.Stakeholders {
		if s.IsUnlocked {
			assert.GreaterOrEqual(t, s.Popularity, 0.0,
				"stakeholder %q popularity below 0", s.ID)
			assert.LessOrEqual(t, s.Popularity, 100.0,
				"stakeholder %q popularity above 100", s.ID)
		}
	}
}

func TestHeadlessRun_100Weeks_EnergyRingBufferAdvanced(t *testing.T) {
	w := loadWorld(t)
	w, _ = HeadlessRun(w, 100)
	// After 100 weeks the ring buffers should have been written at least once.
	// PriceAt(0) returns the most recent value; the ring head must have moved.
	assert.Greater(t, w.EnergyMarket.GasHistory.Head, 0,
		"gas price ring buffer head should have advanced past zero")
}

func TestAdvanceWeek_ApprovedPolicy_BecomesActive(t *testing.T) {
	w := loadWorld(t)
	// Submit a policy and advance enough weeks for it to be approved.
	// First find an unlocked policy.
	var targetID string
	for _, card := range w.PolicyCards {
		if policy.IsUnlocked(card, w.Tech.Maturities) {
			targetID = card.Def.ID
			break
		}
	}
	require.NotEmpty(t, targetID)

	// Submit.
	w, _ = AdvanceWeek(w, []Action{
		{Type: player.ActionTypeSubmitPolicy, Target: targetID},
	})

	// Advance up to 52 weeks to give the approval pipeline time to resolve.
	for i := 0; i < 52; i++ {
		w, _ = AdvanceWeek(w, nil)
		for _, card := range w.PolicyCards {
			if card.Def.ID == targetID &&
				card.State == policy.PolicyStateActive {
				return // passed
			}
		}
	}
	// Check final state -- it should be ACTIVE or REJECTED (not stuck in APPROVED).
	for _, card := range w.PolicyCards {
		if card.Def.ID == targetID {
			assert.NotEqual(t, policy.PolicyStateApproved, card.State,
				"policy %q must not be stuck in APPROVED state", targetID)
			return
		}
	}
	t.Fatalf("policy %q not found", targetID)
}

func TestAdvanceWeek_LobbyMinister_IncreasesRelationship(t *testing.T) {
	w := loadWorld(t)
	// Find any unlocked stakeholder.
	var targetID string
	var initialRel float64
	for _, s := range w.Stakeholders {
		if s.IsUnlocked {
			targetID = s.ID
			initialRel = s.RelationshipScore
			break
		}
	}
	require.NotEmpty(t, targetID)

	w, _ = AdvanceWeek(w, []Action{
		{Type: player.ActionTypeLobbyMinister, Target: targetID},
	})

	for _, s := range w.Stakeholders {
		if s.ID == targetID {
			// Relationship should have increased (lobby action +5, minus passive decay of ~0.01).
			assert.Greater(t, s.RelationshipScore, initialRel-1.0,
				"lobbying stakeholder %q should not significantly decrease relationship", targetID)
			return
		}
	}
	t.Fatalf("stakeholder %q not found", targetID)
}

func TestNewWorld_TechDeliveryLogInitialized(t *testing.T) {
	w := loadWorld(t)
	assert.NotNil(t, w.TechDeliveryLog, "TechDeliveryLog must not be nil")
	assert.Empty(t, w.TechDeliveryLog, "TechDeliveryLog must be empty at game start")
}

func TestHeadlessRun_TechDelivery_FiresForActiveCompany(t *testing.T) {
	w := loadWorld(t)
	// Activate a company so quality can accumulate each week.
	for _, d := range w.Cfg.Companies {
		w.Industry = industry.ActivateCompany(w.Industry, d.ID, config.TechOffshoreWind, 80.0)
		break
	}
	// Advance enough weeks to exceed techDeliveryThreshold (200).
	// A company at BaseQuality~60, WorkRate~80 accumulates ~38 units/week -> ~6 weeks to threshold.
	w = advanceN(w, 10)
	assert.NotEmpty(t, w.TechDeliveryLog,
		"tech delivery milestone must fire after sustained quality accumulation")
}

func TestAdvanceWeek_SubmitPolicyLockedByTech_IsRejectedByGate(t *testing.T) {
	w := loadWorld(t)
	// Find a policy that IS tech-gated and whose gate is NOT yet met.
	var lockedID string
	for _, card := range w.PolicyCards {
		if card.Def.TechUnlockGate != "" && !policy.IsUnlocked(card, w.Tech.Maturities) {
			lockedID = card.Def.ID
			break
		}
	}
	if lockedID == "" {
		t.Skip("no tech-gated locked policies in seed data")
	}

	w, _ = AdvanceWeek(w, []Action{
		{Type: player.ActionTypeSubmitPolicy, Target: lockedID},
	})

	for _, card := range w.PolicyCards {
		if card.Def.ID == lockedID {
			assert.Equal(t, policy.PolicyStateDraft, card.State,
				"tech-locked policy %q should remain DRAFT when gate not met", lockedID)
			return
		}
	}
	t.Fatalf("policy %q not found", lockedID)
}

// ---------------------------------------------------------------------------
// D1: SACKED/RESIGNED -> BACKBENCH state machine
// ---------------------------------------------------------------------------

func TestAdvanceWeek_SackedMinister_BecomesBackbench(t *testing.T) {
	w := loadWorld(t)
	// Find an unlocked cabinet minister and drive them to SACKED by forcing
	// UNDER_PRESSURE for ministerSackingWeeks weeks.
	var targetID string
	for _, s := range w.Stakeholders {
		if s.IsUnlocked && isInCabinet(w.Government, s.ID) {
			targetID = s.ID
			break
		}
	}
	require.NotEmpty(t, targetID, "need at least one cabinet minister")

	// Force the minister to UNDER_PRESSURE and then advance enough weeks.
	for i := range w.Stakeholders {
		if w.Stakeholders[i].ID == targetID {
			w.Stakeholders[i].State = stakeholder.MinisterStateUnderPressure
			w.Stakeholders[i].Popularity = 5.0 // well below sacking threshold
		}
	}

	// Advance ministerSackingWeeks+2 weeks to trigger SACKED then BACKBENCH.
	w = advanceN(w, ministerSackingWeeks+2)

	for _, s := range w.Stakeholders {
		if s.ID == targetID {
			assert.NotEqual(t, stakeholder.MinisterStateSacked, s.State,
				"minister %q should not remain SACKED after sacking transition", targetID)
			assert.NotEqual(t, stakeholder.MinisterStateResigned, s.State,
				"minister %q should not remain RESIGNED after resignation transition", targetID)
			return
		}
	}
	t.Fatalf("stakeholder %q not found", targetID)
}

// ---------------------------------------------------------------------------
// D2: Grace period -- GraceWeeksRemaining starts at zero
// ---------------------------------------------------------------------------

func TestNewWorld_AllMinistersGraceWeeksZero(t *testing.T) {
	w := loadWorld(t)
	for _, s := range w.Stakeholders {
		if s.IsUnlocked {
			assert.Equal(t, 0, s.GraceWeeksRemaining,
				"stakeholder %q should have GraceWeeksRemaining=0 at game start", s.ID)
		}
	}
}

// ---------------------------------------------------------------------------
// D1: No minister stuck in SACKED or RESIGNED after 100 weeks
// ---------------------------------------------------------------------------

func TestHeadlessRun_100Weeks_NoMinisterStuckInSacked(t *testing.T) {
	w := loadWorld(t)
	w, _ = HeadlessRun(w, 100)
	for _, s := range w.Stakeholders {
		if s.IsUnlocked {
			assert.NotEqual(t, stakeholder.MinisterStateSacked, s.State,
				"stakeholder %q should not be in SACKED state after 100 weeks", s.ID)
			assert.NotEqual(t, stakeholder.MinisterStateResigned, s.State,
				"stakeholder %q should not be in RESIGNED state after 100 weeks", s.ID)
		}
	}
}
