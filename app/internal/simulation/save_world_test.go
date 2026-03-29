package simulation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/save"
	"github.com/vibe-code-challenge/twenty-fifty/internal/simulation"
)

func newWorldForSaveTest(t *testing.T) (simulation.WorldState, *config.Config) {
	t.Helper()
	cfg, err := config.Load()
	assert.NoError(t, err)
	seed, err := save.NewMasterSeed()
	assert.NoError(t, err)
	return simulation.NewWorld(cfg, seed), cfg
}

// TestSaveWorld_RoundTrip verifies that SaveWorld -> RestoreWorld produces a
// WorldState whose key fields match the original.
func TestSaveWorld_RoundTrip(t *testing.T) {
	world, cfg := newWorldForSaveTest(t)

	data := simulation.SaveWorld(world)
	restored, err := simulation.RestoreWorld(data, cfg)

	assert.NoError(t, err)
	assert.Equal(t, world.Week, restored.Week)
	assert.Equal(t, world.Year, restored.Year)
	assert.Equal(t, world.Month, restored.Month)
	assert.Equal(t, world.StartYear, restored.StartYear)
	assert.Equal(t, world.ScenarioID, restored.ScenarioID)
	assert.InDelta(t, world.GovernmentPopularity, restored.GovernmentPopularity, 0.001)
	assert.InDelta(t, world.FossilDependency, restored.FossilDependency, 0.001)
	assert.InDelta(t, world.BaseWeeklyMt, restored.BaseWeeklyMt, 0.001)
}

// TestSaveWorld_PolicyCards_PreserveState verifies that policy card state is
// preserved through a save/restore cycle.
func TestSaveWorld_PolicyCards_PreserveState(t *testing.T) {
	world, cfg := newWorldForSaveTest(t)

	// Count cards in DRAFT (should be all of them at start).
	draftCount := 0
	for _, pc := range world.PolicyCards {
		if pc.State == "DRAFT" {
			draftCount++
		}
	}
	assert.Greater(t, draftCount, 0)

	data := simulation.SaveWorld(world)
	restored, err := simulation.RestoreWorld(data, cfg)
	assert.NoError(t, err)

	restoredDraft := 0
	for _, pc := range restored.PolicyCards {
		if pc.State == "DRAFT" {
			restoredDraft++
		}
	}
	assert.Equal(t, draftCount, restoredDraft)
}

// TestSaveWorld_PolicyCards_DefPointerRestored verifies Def pointers are re-linked.
func TestSaveWorld_PolicyCards_DefPointerRestored(t *testing.T) {
	world, cfg := newWorldForSaveTest(t)

	data := simulation.SaveWorld(world)
	restored, err := simulation.RestoreWorld(data, cfg)
	assert.NoError(t, err)

	for _, pc := range restored.PolicyCards {
		assert.NotNil(t, pc.Def, "policy card %q has nil Def after restore", pc.State)
	}
}

// TestSaveWorld_RNGNotNil verifies the restored world has a non-nil RNG.
func TestSaveWorld_RNGNotNil(t *testing.T) {
	world, cfg := newWorldForSaveTest(t)
	data := simulation.SaveWorld(world)
	restored, err := simulation.RestoreWorld(data, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, restored.RNG)
}

// TestSaveWorld_FiredOnceEvents_Preserved verifies the FiredOnceEvents map survives round-trip.
// HumbleBeginnings has no pre-fired events so the map may be nil; we just check no error occurs.
func TestSaveWorld_FiredOnceEvents_Preserved(t *testing.T) {
	world, cfg := newWorldForSaveTest(t)

	data := simulation.SaveWorld(world)
	restored, err := simulation.RestoreWorld(data, cfg)
	assert.NoError(t, err)

	// Nil and empty map are both valid for a fresh HumbleBeginnings world.
	assert.Equal(t, len(world.FiredOnceEvents), len(restored.FiredOnceEvents))
}

// TestSaveWorld_Stakeholders_Preserved verifies stakeholder count is unchanged.
func TestSaveWorld_Stakeholders_Preserved(t *testing.T) {
	world, cfg := newWorldForSaveTest(t)

	data := simulation.SaveWorld(world)
	restored, err := simulation.RestoreWorld(data, cfg)
	assert.NoError(t, err)

	assert.Equal(t, len(world.Stakeholders), len(restored.Stakeholders))
}

// TestSaveWorld_AfterAdvance_WeekIncremented verifies that advancing a week
// and then saving/restoring preserves the new week number.
func TestSaveWorld_AfterAdvance_WeekIncremented(t *testing.T) {
	world, cfg := newWorldForSaveTest(t)

	world, _ = simulation.AdvanceWeek(world, nil)
	assert.Equal(t, 1, world.Week)

	data := simulation.SaveWorld(world)
	restored, err := simulation.RestoreWorld(data, cfg)
	assert.NoError(t, err)
	assert.Equal(t, 1, restored.Week)
}
