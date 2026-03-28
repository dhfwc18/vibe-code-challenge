package evidence

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"twenty-fifty/internal/config"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func fixedRNG() *rand.Rand {
	return rand.New(rand.NewSource(42))
}

func consultancyOrg() config.OrgDefinition {
	return config.OrgDefinition{
		ID:                     "c1",
		OrgType:                config.OrgConsultancy,
		Origin:                 config.OrgLocal,
		BaseCost:               50.0,
		DeliveryDist:           config.TriangularDist{Min: 2, Mode: 3, Max: 9},
		Quality:                config.QualityRange{Min: 30.0, Max: 90.0},
		BiasType:               config.BiasNone,
		BaseFailureProbability: 0.0,
		Specialisms:            []config.InsightType{config.InsightPower},
	}
}

func muricanOrg() config.OrgDefinition {
	org := consultancyOrg()
	org.Origin = config.OrgMurican
	return org
}

// ---------------------------------------------------------------------------
// DrawDeliveryWeek
// ---------------------------------------------------------------------------

func TestDrawDeliveryWeek_ResultWithinBounds(t *testing.T) {
	dist := config.TriangularDist{Min: 2, Mode: 5, Max: 12}
	rng := fixedRNG()
	for i := 0; i < 500; i++ {
		week := DrawDeliveryWeek(dist, 0, rng)
		assert.GreaterOrEqual(t, week, dist.Min, "delivery week below Min")
		assert.LessOrEqual(t, week, dist.Max, "delivery week above Max")
	}
}

func TestDrawDeliveryWeek_MeanNearMode(t *testing.T) {
	// For triangular(a=2, c=5, b=12): mean = (a+b+c)/3 = 19/3 ~ 6.3
	// Mode is not the mean; mean should be approximately 6.3.
	dist := config.TriangularDist{Min: 2, Mode: 5, Max: 12}
	rng := fixedRNG()
	sum := 0
	n := 5000
	for i := 0; i < n; i++ {
		sum += DrawDeliveryWeek(dist, 0, rng)
	}
	mean := float64(sum) / float64(n)
	// Allow +-0.5 tolerance on the mean
	assert.InDelta(t, 6.3, mean, 0.5)
}

func TestDrawDeliveryWeek_AddsCommissionedWeek(t *testing.T) {
	dist := config.TriangularDist{Min: 3, Mode: 3, Max: 3} // degenerate: always 3
	week := DrawDeliveryWeek(dist, 10, fixedRNG())
	assert.Equal(t, 13, week)
}

// ---------------------------------------------------------------------------
// SeedOrgStates
// ---------------------------------------------------------------------------

func TestSeedOrgStates_StartAtDefaultRelationship(t *testing.T) {
	defs := []config.OrgDefinition{consultancyOrg()}
	states := SeedOrgStates(defs)
	assert.Len(t, states, 1)
	assert.Equal(t, startingOrgRelationship, states[0].RelationshipScore)
	assert.Equal(t, "c1", states[0].OrgID)
}

// ---------------------------------------------------------------------------
// CreateCommission
// ---------------------------------------------------------------------------

func TestCreateCommission_DeliveryWeekAfterCommissioned(t *testing.T) {
	org := consultancyOrg()
	c := CreateCommission(org, config.InsightPower, "national", "energy:lcr", 10, fixedRNG())
	assert.GreaterOrEqual(t, c.DeliveryWeek, 10+org.DeliveryDist.Min)
	assert.LessOrEqual(t, c.DeliveryWeek, 10+org.DeliveryDist.Max)
}

func TestCreateCommission_FieldsSet(t *testing.T) {
	org := consultancyOrg()
	c := CreateCommission(org, config.InsightPower, "region:nr", "energy:power", 5, fixedRNG())
	assert.Equal(t, "c1", c.OrgID)
	assert.Equal(t, config.InsightPower, c.InsightType)
	assert.Equal(t, 5, c.CommissionedWeek)
	assert.Equal(t, org.BaseCost, c.BudgetCost)
	assert.False(t, c.Delivered)
	assert.False(t, c.Failed)
}

// ---------------------------------------------------------------------------
// TickDelivery
// ---------------------------------------------------------------------------

func TestTickDelivery_NotYetDue_NotDelivered(t *testing.T) {
	org := consultancyOrg()
	c := Commission{
		ID: "x", OrgID: "c1", DeliveryWeek: 20,
	}
	defs := map[string]config.OrgDefinition{"c1": org}
	_, delivered := TickDelivery([]Commission{c}, defs, 15, fixedRNG())
	assert.Len(t, delivered, 0)
}

func TestTickDelivery_DueWeek_MarksDelivered(t *testing.T) {
	org := consultancyOrg() // failure prob = 0
	c := Commission{ID: "x", OrgID: "c1", DeliveryWeek: 10}
	defs := map[string]config.OrgDefinition{"c1": org}
	updated, delivered := TickDelivery([]Commission{c}, defs, 10, fixedRNG())
	assert.Len(t, delivered, 1)
	assert.True(t, updated[0].Delivered)
}

func TestTickDelivery_FailureProbability_FiresAtRate(t *testing.T) {
	org := consultancyOrg()
	org.BaseFailureProbability = 1.0 // always fails
	c := Commission{ID: "x", OrgID: "c1", DeliveryWeek: 5}
	defs := map[string]config.OrgDefinition{"c1": org}
	updated, _ := TickDelivery([]Commission{c}, defs, 5, fixedRNG())
	assert.True(t, updated[0].Failed)
	assert.False(t, updated[0].Delivered)
}

// ---------------------------------------------------------------------------
// ApplyBias
// ---------------------------------------------------------------------------

func TestApplyBias_None_CloseToCentral(t *testing.T) {
	org := consultancyOrg()
	org.BiasType = config.BiasNone
	rng := rand.New(rand.NewSource(99))
	sum := 0.0
	n := 1000
	for i := 0; i < n; i++ {
		sum += ApplyBias(50.0, org, 0.0, rng)
	}
	mean := sum / float64(n)
	assert.InDelta(t, 50.0, mean, 0.5)
}

func TestApplyBias_Ideological_ShiftsCorrectDirection(t *testing.T) {
	org := consultancyOrg()
	org.BiasType = config.BiasIdeological
	org.BiasDirection = 1.0 // positive direction -> value should increase
	result := ApplyBias(50.0, org, 0.0, fixedRNG())
	assert.Greater(t, result, 50.0)
}

func TestApplyBias_IdeologicalNegative_ShiftsDown(t *testing.T) {
	org := consultancyOrg()
	org.BiasType = config.BiasIdeological
	org.BiasDirection = -1.0
	result := ApplyBias(50.0, org, 0.0, fixedRNG())
	assert.Less(t, result, 50.0)
}

func TestApplyBias_ClientConfirmation_ShiftsWithDecisionBias(t *testing.T) {
	org := consultancyOrg()
	org.BiasType = config.BiasClientConfirmation
	org.ClientBiasWeight = 1.0

	// Positive decision bias -> reported value > raw value
	pos := ApplyBias(50.0, org, 1.0, fixedRNG())
	neg := ApplyBias(50.0, org, -1.0, fixedRNG())
	assert.Greater(t, pos, 50.0)
	assert.Less(t, neg, 50.0)
}

// ---------------------------------------------------------------------------
// GenerateReport
// ---------------------------------------------------------------------------

func TestGenerateReport_ReportedValueClamped(t *testing.T) {
	org := consultancyOrg()
	c := Commission{
		ID: "r1", OrgID: "c1", InsightType: config.InsightPower,
		Scope: "national", DeliveryWeek: 5,
	}
	report := GenerateReport(c, org, "energy:power", 50.0, 0.0, fixedRNG())
	assert.GreaterOrEqual(t, report.ReportedValue, 0.0)
	assert.LessOrEqual(t, report.ReportedValue, 100.0)
}

func TestGenerateReport_OutsideSpecialism_LowerQuality(t *testing.T) {
	org := consultancyOrg()
	// InsightType not in Specialisms -> quality penalty
	c := Commission{
		ID: "r1", OrgID: "c1", InsightType: config.InsightClimate,
		DeliveryWeek: 5,
	}
	report := GenerateReport(c, org, "climate:state", 50.0, 0.0, rand.New(rand.NewSource(1)))
	assert.False(t, report.SpecialismBonus)
	// With the penalty, quality should be below the normal minimum minus penalty = 30-20=10
	assert.LessOrEqual(t, report.QualityScore, org.Quality.Max-outsideSpecialismQualityPenalty+1)
}

func TestGenerateReport_InSpecialism_HasBonus(t *testing.T) {
	org := consultancyOrg()
	c := Commission{
		ID: "r1", OrgID: "c1", InsightType: config.InsightPower,
		DeliveryWeek: 5,
	}
	report := GenerateReport(c, org, "energy:power", 50.0, 0.0, fixedRNG())
	assert.True(t, report.SpecialismBonus)
}

// ---------------------------------------------------------------------------
// UpdateOrgRelationship
// ---------------------------------------------------------------------------

func TestUpdateOrgRelationship_Delivered_IncreasesScore(t *testing.T) {
	org := OrgState{OrgID: "c1", RelationshipScore: 50.0}
	updated := UpdateOrgRelationship(org, RelationshipEventDelivered, 10)
	assert.Greater(t, updated.RelationshipScore, 50.0)
}

func TestUpdateOrgRelationship_Failed_DecreaseScoreAndSetCooling(t *testing.T) {
	org := OrgState{OrgID: "c1", RelationshipScore: 50.0}
	updated := UpdateOrgRelationship(org, RelationshipEventFailed, 10)
	assert.Less(t, updated.RelationshipScore, 50.0)
	assert.Equal(t, 10+coolingOffWeeks, updated.CoolingOffUntil)
}

func TestUpdateOrgRelationship_NaturalDecay_MovesToward50(t *testing.T) {
	org := OrgState{OrgID: "c1", RelationshipScore: 80.0}
	updated := UpdateOrgRelationship(org, "", 1)
	assert.Less(t, updated.RelationshipScore, 80.0)
	assert.Greater(t, updated.RelationshipScore, 50.0)
}

// ---------------------------------------------------------------------------
// MuracanOrgAvailable
// ---------------------------------------------------------------------------

func TestMuracanOrgAvailable_LocalOrg_AlwaysTrue(t *testing.T) {
	org := consultancyOrg()
	assert.True(t, MuracanOrgAvailable(org, false, false))
}

func TestMuracanOrgAvailable_MuricanOrg_RequiresBothFlags(t *testing.T) {
	org := muricanOrg()
	assert.False(t, MuracanOrgAvailable(org, false, false))
	assert.False(t, MuracanOrgAvailable(org, true, false))
	assert.False(t, MuracanOrgAvailable(org, false, true))
	assert.True(t, MuracanOrgAvailable(org, true, true))
}
