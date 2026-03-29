package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Load
// ---------------------------------------------------------------------------

func TestLoad_ValidData_ReturnsConfig(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

// ---------------------------------------------------------------------------
// Counts
// ---------------------------------------------------------------------------

func TestLoad_Stakeholders_ExpectedCount(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 27, len(cfg.Stakeholders), "expected 27 stakeholder seeds across 4 parties")
}

func TestLoad_Organisations_ExpectedCount(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 18, len(cfg.Organisations), "expected 18 organisations (15 local + 3 Murican)")
}

func TestLoad_Companies_ExpectedCount(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 15, len(cfg.Companies), "expected 15 LCT companies")
}

func TestLoad_Technologies_ExpectedCount(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 8, len(cfg.Technologies), "expected 8 technology curves")
}

func TestLoad_Regions_ExpectedCount(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 12, len(cfg.Regions), "expected 12 Taitan regions")
}

func TestLoad_Tiles_ExpectedCount(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 30, len(cfg.Tiles), "expected 30 map tiles")
}

func TestLoad_PolicyCards_AtLeastMinimum(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(cfg.PolicyCards), 20, "expected at least 20 policy cards")
}

func TestLoad_Events_AtLeastMinimum(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(cfg.Events), 18, "expected at least 18 event definitions")
}

// ---------------------------------------------------------------------------
// Carbon budgets
// ---------------------------------------------------------------------------

func TestLoad_CarbonBudgets_GameStartYear2010(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	require.NotEmpty(t, cfg.CarbonBudgets)
	assert.Equal(t, 2010, cfg.CarbonBudgets[0].Year)
}

func TestLoad_CarbonBudgets_NetZeroBy2050(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	last := cfg.CarbonBudgets[len(cfg.CarbonBudgets)-1]
	assert.Equal(t, 2050, last.Year)
	assert.Equal(t, 0.0, last.AnnualLimitMtCO2e)
}

func TestLoad_CarbonBudgets_StrictlyAscendingYears(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	for i := 1; i < len(cfg.CarbonBudgets); i++ {
		assert.Greater(t,
			cfg.CarbonBudgets[i].Year,
			cfg.CarbonBudgets[i-1].Year,
			"carbon budget years must be strictly ascending at index %d", i,
		)
	}
}

// ---------------------------------------------------------------------------
// Unique IDs
// ---------------------------------------------------------------------------

func TestLoad_StakeholderIDs_AllUnique(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assertUniqueIDs(t, "stakeholder", stakeholderIDs(cfg.Stakeholders))
}

func TestLoad_OrgIDs_AllUnique(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assertUniqueIDs(t, "organisation", orgIDs(cfg.Organisations))
}

func TestLoad_CompanyIDs_AllUnique(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assertUniqueIDs(t, "company", companyIDList(cfg.Companies))
}

func TestLoad_TechIDs_AllUnique(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assertUniqueIDs(t, "technology", techIDs(cfg.Technologies))
}

func TestLoad_RegionIDs_AllUnique(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assertUniqueIDs(t, "region", regionIDList(cfg.Regions))
}

func TestLoad_TileIDs_AllUnique(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assertUniqueIDs(t, "tile", tileIDList(cfg.Tiles))
}

func TestLoad_PolicyCardIDs_AllUnique(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assertUniqueIDs(t, "policy_card", policyIDs(cfg.PolicyCards))
}

func TestLoad_EventIDs_AllUnique(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assertUniqueIDs(t, "event", eventIDList(cfg.Events))
}

// ---------------------------------------------------------------------------
// Cross-reference integrity
// ---------------------------------------------------------------------------

func TestLoad_TileRegionRefs_AllValid(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	knownRegions := make(map[string]bool, len(cfg.Regions))
	for _, r := range cfg.Regions {
		knownRegions[r.ID] = true
	}
	for _, tile := range cfg.Tiles {
		assert.True(t, knownRegions[tile.RegionID],
			"tile %q references unknown region %q", tile.ID, tile.RegionID)
	}
}

func TestLoad_RegionTileRefs_AllValid(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	knownTiles := make(map[string]bool, len(cfg.Tiles))
	for _, t2 := range cfg.Tiles {
		knownTiles[t2.ID] = true
	}
	for _, r := range cfg.Regions {
		for _, tid := range r.TileIDs {
			assert.True(t, knownTiles[tid],
				"region %q references unknown tile %q", r.ID, tid)
		}
	}
}

// ---------------------------------------------------------------------------
// Domain-specific validity
// ---------------------------------------------------------------------------

func TestLoad_Orgs_MuricanOrgsHaveHighPopularityRisk(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	for _, o := range cfg.Organisations {
		if o.Origin == OrgMurican {
			assert.GreaterOrEqual(t, o.PopularityRisk, 0.50,
				"Murican org %q should have PopularityRisk >= 0.5", o.ID)
		}
	}
}

func TestLoad_TechCurves_InitialMaturityWithinBounds(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	for _, tech := range cfg.Technologies {
		assert.GreaterOrEqual(t, tech.InitialMaturity, 0.0,
			"tech %q InitialMaturity must be >= 0", tech.ID)
		assert.LessOrEqual(t, tech.InitialMaturity, 100.0,
			"tech %q InitialMaturity must be <= 100", tech.ID)
	}
}

func TestLoad_Stakeholders_SuccessorTimingHasNonZeroEntryWeek(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	for _, s := range cfg.Stakeholders {
		if s.EntryTiming == TimingMid || s.EntryTiming == TimingLate {
			assert.Greater(t, s.EntryWeekMin, 0,
				"stakeholder %q with timing %s must have EntryWeekMin > 0",
				s.ID, s.EntryTiming)
		}
	}
}

func TestLoad_PolicyCards_APCostPositive(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	for _, p := range cfg.PolicyCards {
		assert.Greater(t, p.APCost, 0,
			"policy %q must have APCost > 0", p.ID)
	}
}

func TestLoad_Events_BaseProbabilityPositive(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	for _, e := range cfg.Events {
		if e.TriggerAtYear > 0 {
			continue // time-gated events fire deterministically; BaseProbability intentionally 0
		}
		assert.Greater(t, e.BaseProbability, 0.0,
			"event %q must have BaseProbability > 0", e.ID)
	}
}

func TestLoad_Tiles_PoliticalOpinionWithinBounds(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	for _, tile := range cfg.Tiles {
		assert.GreaterOrEqual(t, tile.InitialPoliticalOpinion, 0.0,
			"tile %q InitialPoliticalOpinion must be >= 0", tile.ID)
		assert.LessOrEqual(t, tile.InitialPoliticalOpinion, 100.0,
			"tile %q InitialPoliticalOpinion must be <= 100", tile.ID)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func assertUniqueIDs(t *testing.T, kind string, ids []string) {
	t.Helper()
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		assert.NotEmpty(t, id, "%s has an empty ID", kind)
		assert.False(t, seen[id], "%s ID %q is duplicated", kind, id)
		seen[id] = true
	}
}
