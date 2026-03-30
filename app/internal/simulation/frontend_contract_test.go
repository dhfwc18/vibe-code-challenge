package simulation

// frontend_contract_test.go verifies every synchronisation point between the
// frontend (ui package) and the simulation backend, as documented in
// docs/frontend_contract.md.
//
// Tests are grouped by contract section:
//   A. WorldState field validity -- all fields the UI reads are in range
//   B. Action round-trips       -- every ActionType the UI queues is processed
//   C. AP economy               -- replenishment and deduction cycle
//
// Tests in this file are intentionally additive: they cover gaps not already
// addressed by simulation_test.go, integration_test.go, or
// counterintuitive_test.go.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/event"
	"github.com/vibe-code-challenge/twenty-fifty/internal/government"
	"github.com/vibe-code-challenge/twenty-fifty/internal/player"
	"github.com/vibe-code-challenge/twenty-fifty/internal/policy"
)

// ---------------------------------------------------------------------------
// A. WorldState field validity (contract section 2)
// ---------------------------------------------------------------------------

// TestContractEnergyTab_RenewableShareInPercentRange verifies the field the
// energy tab reads is always in [0, 100].
// NOTE: the contract doc incorrectly states "0-1 fraction"; energy.EnergyMarket
// stores RenewableGridShare as 0-100 percent (comment: "// 0-100 percent").
// The UI must NOT multiply this by 100. This test catches a recurrence of
// the prior *100 display bug.
func TestContractEnergyTab_RenewableShareInPercentRange(t *testing.T) {
	w := loadWorld(t)
	w, _ = AdvanceWeek(w, nil)
	assert.GreaterOrEqual(t, w.EnergyMarket.RenewableGridShare, 0.0,
		"RenewableGridShare must be >= 0")
	assert.LessOrEqual(t, w.EnergyMarket.RenewableGridShare, 100.0,
		"RenewableGridShare must be <= 100 (stored as percent, not unit fraction)")
}

// TestContractEnergyTab_AllTechMaturitiesInPercentRange verifies that
// Tech.Maturity() returns a value in [0, 100] for each of the 8 technologies.
// NOTE: the contract doc incorrectly states "0-1 fraction"; technology.TechTracker
// clamps maturities to [0, 100] (percent). The UI must NOT multiply by 100.
// This test prevents recurrence of the prior *100 display bug.
func TestContractEnergyTab_AllTechMaturitiesInPercentRange(t *testing.T) {
	w := loadWorld(t)
	w, _ = AdvanceWeek(w, nil)

	techs := []config.Technology{
		config.TechOffshoreWind,
		config.TechOnshoreWind,
		config.TechSolarPV,
		config.TechNuclear,
		config.TechHeatPumps,
		config.TechEVs,
		config.TechHydrogen,
		config.TechIndustrialCCS,
	}
	for _, tech := range techs {
		m := w.Tech.Maturity(tech)
		assert.GreaterOrEqual(t, m, 0.0,
			"Tech.Maturity(%s) must be >= 0", tech)
		assert.LessOrEqual(t, m, 100.0,
			"Tech.Maturity(%s) must be <= 100 (stored as percent, not unit fraction)", tech)
	}
}

// TestContractPolicyTab_AllCardsHaveNonNilDef verifies that every PolicyCard
// returned by NewWorld has a non-nil Def pointer. A nil Def would panic in
// drawPolicyCard when accessing Def.ID, Def.Name, Def.APCost, etc.
// Ref: contract "Def.ID | string | ...", "Def.APCost | int | ...".
func TestContractPolicyTab_AllCardsHaveNonNilDef(t *testing.T) {
	w := loadWorld(t)
	require.NotEmpty(t, w.PolicyCards, "PolicyCards must not be empty")
	for i, pc := range w.PolicyCards {
		assert.NotNilf(t, pc.Def, "PolicyCard at index %d has nil Def -- "+
			"ui.drawPolicyCard will panic when rendering the policy tab", i)
	}
}

// TestContractPolicyTab_AtLeastOneDraftCardExists verifies that at least one
// policy card starts in DRAFT state. Without a DRAFT card the Submit button
// in the policy tab is never rendered and the player has no initial action.
// Ref: contract "State | policy.PolicyState | ...".
func TestContractPolicyTab_AtLeastOneDraftCardExists(t *testing.T) {
	w := loadWorld(t)
	for _, pc := range w.PolicyCards {
		if pc.State == policy.PolicyStateDraft {
			return // found at least one
		}
	}
	t.Fatal("no DRAFT policy cards found in initial WorldState; " +
		"the Submit button in the policy tab will never be visible")
}

// TestContractBudgetTab_AllFiveDeptKeysPresent verifies that
// LastBudget.Departments contains all five department keys defined in the
// government package. The budget tab iterates these keys to render allocation
// bars; a missing key renders as zero without warning.
// Ref: contract section 7, "Department IDs".
func TestContractBudgetTab_AllFiveDeptKeysPresent(t *testing.T) {
	w := loadWorld(t)
	depts := []string{
		government.DeptPower,
		government.DeptTransport,
		government.DeptBuildings,
		government.DeptIndustry,
		government.DeptCross,
	}
	for _, d := range depts {
		_, ok := w.LastBudget.Departments[d]
		assert.Truef(t, ok,
			"dept key %q must be present in LastBudget.Departments (budget tab reads it)", d)
	}
}

// TestContractIndustryTab_CompaniesNonEmpty verifies that the company map is
// populated on a fresh world. The industry tab iterates this map; an empty map
// would render a blank panel with no feedback to the player.
// Ref: contract "Industry.Companies | map[string]industry.CompanyState".
func TestContractIndustryTab_CompaniesNonEmpty(t *testing.T) {
	w := loadWorld(t)
	assert.NotEmpty(t, w.Industry.Companies,
		"Industry.Companies must have at least one entry for the industry tab to render")
}

// TestContractEvidenceTab_OrgStateCountMatchesConfig verifies that every
// organisation defined in config has a corresponding OrgState entry. The
// evidence tab iterates Cfg.Organisations and calls findOrgState for each;
// a missing entry would silently skip that org.
// Ref: contract "OrgStates ([]evidence.OrgState) -- per-org: RelationshipScore, CoolingOffUntil, MuricanUnlocked".
func TestContractEvidenceTab_OrgStateCountMatchesConfig(t *testing.T) {
	w := loadWorld(t)
	require.NotNil(t, w.Cfg, "WorldState.Cfg must not be nil")
	assert.Equal(t, len(w.Cfg.Organisations), len(w.OrgStates),
		"OrgStates must have exactly one entry per organisation in config")
}

// TestContractMapTab_TilesNonEmpty verifies the tile slice is populated.
// The map tab renders a region/tile list; an empty slice yields a blank panel.
// Ref: contract "Tiles | []region.Tile".
func TestContractMapTab_TilesNonEmpty(t *testing.T) {
	w := loadWorld(t)
	assert.NotEmpty(t, w.Tiles,
		"Tiles must not be empty (map tab renders tile-level data)")
}

// TestContractMapTab_TileFieldsInRange verifies that the three per-tile floats
// read by the UI are within [0, 100] on a fresh world.
// Ref: contract "FuelPoverty | float64 | 0-100", "InsulationLevel | float64 | 0-100",
//
//	"LocalPoliticalOpinion | float64 | 0-100; 50 = neutral".
//
// NOTE: the contract names the field "LocalPoliticalOpinion" but the actual
// region.Tile struct field is "PoliticalOpinion". The contract should be
// corrected; this test uses the real field name.
func TestContractMapTab_TileFieldsInRange(t *testing.T) {
	w := loadWorld(t)
	for i, tile := range w.Tiles {
		assert.GreaterOrEqualf(t, tile.FuelPoverty, 0.0,
			"Tiles[%d].FuelPoverty must be >= 0", i)
		assert.LessOrEqualf(t, tile.FuelPoverty, 100.0,
			"Tiles[%d].FuelPoverty must be <= 100", i)
		assert.GreaterOrEqualf(t, tile.InsulationLevel, 0.0,
			"Tiles[%d].InsulationLevel must be >= 0", i)
		assert.LessOrEqualf(t, tile.InsulationLevel, 100.0,
			"Tiles[%d].InsulationLevel must be <= 100", i)
		// Contract names this LocalPoliticalOpinion; struct field is PoliticalOpinion.
		assert.GreaterOrEqualf(t, tile.PoliticalOpinion, 0.0,
			"Tiles[%d].PoliticalOpinion must be >= 0", i)
		assert.LessOrEqualf(t, tile.PoliticalOpinion, 100.0,
			"Tiles[%d].PoliticalOpinion must be <= 100", i)
	}
}

// TestContractFieldPath_LCRLastPollResultAccessible verifies the exact field
// path that hud.go and tab_overview.go use to render the LCR value. If the
// reputation package renames LastPollResult this test will fail to compile,
// making the breakage visible before a runtime panic.
// Ref: contract "LCR.Value | float64 | Low-Carbon Reputation; 0-100".
// Note: the UI reads world.LCR.LastPollResult (the noisy poll sample), not LCR.Value.
func TestContractFieldPath_LCRLastPollResultAccessible(t *testing.T) {
	w := loadWorld(t)
	// Access the exact field paths used in ui/hud.go and ui/tab_overview.go.
	v := w.LCR.LastPollResult
	assert.GreaterOrEqual(t, v, 0.0, "LCR.LastPollResult must be >= 0")
	assert.LessOrEqual(t, v, 100.0, "LCR.LastPollResult must be <= 100")
	// Also verify the underlying true value is accessible (used in simulation logic).
	assert.GreaterOrEqual(t, w.LCR.Value, 0.0, "LCR.Value must be >= 0")
	assert.LessOrEqual(t, w.LCR.Value, 100.0, "LCR.Value must be <= 100")
}

// ---------------------------------------------------------------------------
// B. Action round-trips (contract section 3)
// ---------------------------------------------------------------------------

// TestContractAction_ShockResponse_AllOptions_ClearShock verifies that all
// three ShockResponseOption values the UI can submit (ACCEPT, DECLINE, MITIGATE)
// cause the matching entry to be removed from PendingShockResponses.
//
// MITIGATE is the option added in the UI fix (modal_shock.go); this test
// prevents regressions where a new Detail string is silently ignored.
// Ref: contract "player.ActionTypeShockResponse | event def ID | ShockResponseOption string".
func TestContractAction_ShockResponse_AllOptions_ClearShock(t *testing.T) {
	options := []string{"ACCEPT", "DECLINE", "MITIGATE"}
	for _, opt := range options {
		opt := opt // capture for subtest
		t.Run(opt, func(t *testing.T) {
			w := loadWorld(t)
			w, _ = AdvanceWeek(w, nil)

			// Inject a pending shock directly (event firing is stochastic).
			const shockID = "test_shock_contract"
			w.PendingShockResponses = []event.PendingShockResponse{
				{EventDefID: shockID, Week: w.Week},
			}
			require.Len(t, w.PendingShockResponses, 1)

			w, _ = AdvanceWeek(w, []Action{
				{Type: player.ActionTypeShockResponse, Target: shockID, Detail: opt},
			})

			for _, psr := range w.PendingShockResponses {
				assert.NotEqual(t, shockID, psr.EventDefID,
					"shock %q must not remain in PendingShockResponses after %s response",
					shockID, opt)
			}
		})
	}
}

// TestContractAction_TickyPressure_DeclineAndNegotiate_ClearFlag verifies that
// DECLINE and NEGOTIATE both clear PendingTickyPressure.
// ACCEPT is already covered by TestTickyPressure_AcceptDeal_UnlocksOrgAndBoostsRelationship
// in integration_test.go; this test closes the gap for the two non-accept options.
// Ref: contract "player.ActionTypeRespondTickyPressure | ” | 'ACCEPT', 'DECLINE', or 'NEGOTIATE'".
func TestContractAction_TickyPressure_DeclineAndNegotiate_ClearFlag(t *testing.T) {
	options := []string{"DECLINE", "NEGOTIATE"}
	for _, detail := range options {
		detail := detail
		t.Run(detail, func(t *testing.T) {
			w := loadWorld(t)
			w, _ = AdvanceWeek(w, nil)

			// Force the pressure flag without needing Ticky in cabinet.
			w.PendingTickyPressure = true

			w, _ = AdvanceWeek(w, []Action{
				{
					Type:   player.ActionTypeRespondTickyPressure,
					Target: tickyStakeholderID,
					Detail: detail,
				},
			})

			assert.False(t, w.PendingTickyPressure,
				"PendingTickyPressure must be false after %s response", detail)
		})
	}
}

// TestContractAction_FireStaff_RemovesHiredStaff verifies the full hire-then-fire
// round-trip via AdvanceWeek. The UI does not yet expose a Fire button, but
// ActionTypeFireStaff is in the contract and must work correctly for future tabs.
// Ref: contract "player.ActionTypeFireStaff | staff member ID | ”".
func TestContractAction_FireStaff_RemovesHiredStaff(t *testing.T) {
	w := loadWorld(t)

	// Hire an analyst in week 1.
	w, _ = AdvanceWeek(w, []Action{
		{Type: player.ActionTypeHireStaff, Target: string(player.StaffRoleAnalyst)},
	})
	require.Len(t, w.Player.Staff, 1, "must have exactly one staff member after hire")
	staffID := w.Player.Staff[0].ID

	// Fire the hired analyst in week 2.
	w, _ = AdvanceWeek(w, []Action{
		{Type: player.ActionTypeFireStaff, Target: staffID},
	})

	for _, s := range w.Player.Staff {
		assert.NotEqual(t, staffID, s.ID,
			"fired staff member %q must not remain in Player.Staff", staffID)
	}
}

// TestContractAction_SubmitPolicy_StateVisibleToUI verifies that after
// ActionTypeSubmitPolicy the card's State transitions to UNDER_REVIEW in the
// same WorldState snapshot the UI reads. If the state change were deferred,
// the policy column would show the card in the wrong column for one frame.
// Ref: contract "ActionTypeSubmitPolicy | policy card ID | ”".
func TestContractAction_SubmitPolicy_StateVisibleToUI(t *testing.T) {
	w := loadWorld(t)
	w, _ = AdvanceWeek(w, nil) // ensure AP pool is set and tech maturities are seeded

	// Find a DRAFT card that is tech-unlocked and within the base AP pool (5).
	var targetID string
	for _, pc := range w.PolicyCards {
		if pc.State != policy.PolicyStateDraft || pc.Def == nil {
			continue
		}
		if pc.Def.APCost > 5 {
			continue
		}
		if !policy.IsUnlocked(pc, w.Tech.Maturities) {
			continue
		}
		targetID = pc.Def.ID
		break
	}
	require.NotEmpty(t, targetID,
		"need a DRAFT policy that is tech-unlocked and has APCost <= 5; check config/policies.go")

	w, _ = AdvanceWeek(w, []Action{
		{Type: player.ActionTypeSubmitPolicy, Target: targetID},
	})

	for _, pc := range w.PolicyCards {
		if pc.Def != nil && pc.Def.ID == targetID {
			assert.Equal(t, policy.PolicyStateUnderReview, pc.State,
				"submitted policy must appear as UNDER_REVIEW in the same WorldState the UI renders")
			return
		}
	}
	t.Fatalf("policy card %q not found in PolicyCards after submit", targetID)
}

// ---------------------------------------------------------------------------
// C. AP economy (contract section 5)
// ---------------------------------------------------------------------------

// TestContractAP_RestoredToFullPoolEachWeek verifies that Player.APRemaining
// equals WeeklyAPPool at the end of a week with no actions. This is the value
// the HUD's effectiveAP() method starts from; if it were stale the AP display
// would be wrong on week 2+.
// Ref: contract "AP is replenished at the start of each AdvanceWeek call."
func TestContractAP_RestoredToFullPoolEachWeek(t *testing.T) {
	w := loadWorld(t)

	// Advance two weeks with no actions to ensure replenishment is
	// repeatable and not a one-time initialisation artifact.
	for i := 0; i < 2; i++ {
		w, _ = AdvanceWeek(w, nil)
		expected := player.WeeklyAPPool(w.Player)
		assert.Equalf(t, expected, w.Player.APRemaining,
			"APRemaining must equal WeeklyAPPool(%d) after week %d with no actions",
			expected, i+1)
	}
}

// TestContractAP_LobbyMinister_DeductsThreeAP verifies that
// ActionTypeLobbyMinister deducts exactly lobbyAPCost (3) AP from
// APRemaining. The UI uses effectiveAP = APRemaining - pendingAPSpend to gate
// the Lobby button; if the simulation deducts a different amount the button
// stays enabled incorrectly.
// Ref: contract "other actions have fixed costs in the simulation layer".
func TestContractAP_LobbyMinister_DeductsThreeAP(t *testing.T) {
	w := loadWorld(t)

	// Find a minister in the initial cabinet.
	var targetID string
	for _, sid := range w.Government.CabinetByRole {
		targetID = sid
		break
	}
	require.NotEmpty(t, targetID, "need a cabinet minister to lobby")

	w, _ = AdvanceWeek(w, []Action{
		{Type: player.ActionTypeLobbyMinister, Target: targetID},
	})

	// APRemaining = WeeklyAPPool - lobbyAPCost (3).
	expectedRemaining := player.WeeklyAPPool(w.Player) - lobbyAPCost
	assert.Equal(t, expectedRemaining, w.Player.APRemaining,
		"LobbyMinister must deduct exactly %d AP (lobbyAPCost)", lobbyAPCost)
}

// TestContractAP_FullyRestoredAfterSpending verifies the week-over-week AP
// cycle: spend AP in week N, advance to week N+1 with no actions, and
// APRemaining is back to the full pool. This mirrors the UI's
// AdvanceWeekRequested() call that resets pendingAPSpend.
// Ref: contract "AP is replenished at the start of each AdvanceWeek call. It is not cumulative."
func TestContractAP_FullyRestoredAfterSpending(t *testing.T) {
	w := loadWorld(t)

	var targetID string
	for _, sid := range w.Government.CabinetByRole {
		targetID = sid
		break
	}
	require.NotEmpty(t, targetID)

	// Spend AP in week 1.
	w, _ = AdvanceWeek(w, []Action{
		{Type: player.ActionTypeLobbyMinister, Target: targetID},
	})
	spentWeek := player.WeeklyAPPool(w.Player) - w.Player.APRemaining
	require.Greater(t, spentWeek, 0, "AP must have been deducted in week 1")

	// Advance week 2 with no actions.
	w, _ = AdvanceWeek(w, nil)

	assert.Equal(t, player.WeeklyAPPool(w.Player), w.Player.APRemaining,
		"APRemaining must be restored to the full weekly pool in week 2 "+
			"regardless of how much was spent in week 1")
}
