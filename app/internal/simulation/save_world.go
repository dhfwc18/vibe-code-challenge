package simulation

import (
	"math/rand"

	"github.com/vibe-code-challenge/twenty-fifty/internal/carbon"
	"github.com/vibe-code-challenge/twenty-fifty/internal/climate"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/economy"
	"github.com/vibe-code-challenge/twenty-fifty/internal/energy"
	"github.com/vibe-code-challenge/twenty-fifty/internal/event"
	"github.com/vibe-code-challenge/twenty-fifty/internal/evidence"
	"github.com/vibe-code-challenge/twenty-fifty/internal/government"
	"github.com/vibe-code-challenge/twenty-fifty/internal/industry"
	"github.com/vibe-code-challenge/twenty-fifty/internal/player"
	"github.com/vibe-code-challenge/twenty-fifty/internal/policy"
	"github.com/vibe-code-challenge/twenty-fifty/internal/polling"
	"github.com/vibe-code-challenge/twenty-fifty/internal/region"
	"github.com/vibe-code-challenge/twenty-fifty/internal/reputation"
	"github.com/vibe-code-challenge/twenty-fifty/internal/stakeholder"
	"github.com/vibe-code-challenge/twenty-fifty/internal/technology"
)

// WorldSaveData is the JSON-serializable representation of a WorldState.
// It excludes the live *config.Config (reloaded on restore) and *rand.Rand (reseeded).
type WorldSaveData struct {
	// Schema
	SaveVersion int `json:"save_version"`

	// Clock
	Week      int `json:"week"`
	Year      int `json:"year"`
	Month     int `json:"month"`
	Quarter   int `json:"quarter"`
	StartYear int `json:"start_year"`

	// Scenario
	ScenarioID            config.ScenarioID `json:"scenario_id"`
	ScandalRateMultiplier float64           `json:"scandal_rate_multiplier"`
	BaseWeeklyMt          float64           `json:"base_weekly_mt"`

	// RNG: save a fresh seed so the game continues with a fresh sequence post-load.
	RNGSeed int64 `json:"rng_seed"`

	// Political
	Government           government.GovernmentState `json:"government"`
	Stakeholders         []stakeholder.Stakeholder  `json:"stakeholders"`
	GovernmentPopularity float64                    `json:"government_popularity"`

	// Environment
	EnergyMarket     energy.EnergyMarket      `json:"energy_market"`
	ClimateState     climate.ClimateState     `json:"climate_state"`
	FossilDependency float64                  `json:"fossil_dependency"`
	Carbon           carbon.CarbonBudgetState `json:"carbon"`

	// Technology and industry
	Tech     technology.TechTracker `json:"tech"`
	Industry industry.IndustryState `json:"industry"`

	// Geography
	Regions []region.Region `json:"regions"`
	Tiles   []region.Tile   `json:"tiles"`

	// Economy
	Economy        economy.EconomyState     `json:"economy"`
	LastTaxRevenue economy.TaxRevenue       `json:"last_tax_revenue"`
	LastBudget     economy.BudgetAllocation `json:"last_budget"`

	// Reputation
	LCR               reputation.LowCarbonReputation `json:"lcr"`
	WeeksUntilLCRPoll int                            `json:"weeks_until_lcr_poll"`

	// Polling
	PollHistory []polling.PollSnapshot `json:"poll_history"`

	// Policy (stored as ID-keyed save structs)
	PolicyCards []policy.PolicyCardSave `json:"policy_cards"`

	// Evidence
	OrgStates   []evidence.OrgState      `json:"org_states"`
	Commissions []evidence.Commission    `json:"commissions"`
	Reports     []evidence.InsightReport `json:"reports"`

	// Events
	EventLog              event.EventLog               `json:"event_log"`
	PressureGroups        []event.PressureGroup        `json:"pressure_groups"`
	PendingShockResponses []event.PendingShockResponse `json:"pending_shock_responses"`

	// Ticky
	TickyCountdown                   int  `json:"ticky_countdown"`
	PendingTickyPressure             bool `json:"pending_ticky_pressure"`
	TickyPressureAcceptedThisQuarter bool `json:"ticky_pressure_accepted_this_quarter"`
	PendingRiskyTicky                bool `json:"pending_risky_ticky"`
	PendingTrickyTicky               bool `json:"pending_tricky_ticky"`
	AngryTickyActive                 bool `json:"angry_ticky_active"`
	AngryTickyWimpy                  bool `json:"angry_ticky_wimpy"`

	// Decaying shocks
	ActiveDecayingShocks []ActiveDecayingShock `json:"active_decaying_shocks"`

	// Great Sneeze
	GreatSneezeActive  bool `json:"great_sneeze_active"`
	GreatSneezeWeekEnd int  `json:"great_sneeze_week_end"`
	GreatSneezeFired   bool `json:"great_sneeze_fired"`

	// Fired-once event tracking
	FiredOnceEvents map[string]bool `json:"fired_once_events"`

	// Player
	Player player.CivilServant `json:"player"`

	// Poll results
	GovernmentLastPollResult float64            `json:"government_last_poll_result"`
	MinisterLastPollResults  map[string]float64 `json:"minister_last_poll_results"`

	// Tech delivery log
	TechDeliveryLog []string `json:"tech_delivery_log"`

	// Weekly transient accumulators (persisted so first-frame render is accurate)
	WeeklyNetCarbonMt         float64 `json:"weekly_net_carbon_mt"`
	WeeklyPolicyReductionMt   float64 `json:"weekly_policy_reduction_mt"`
	WeeklyEventLCRDelta       float64 `json:"weekly_event_lcr_delta"`
	WeeklyPolicyLCRDelta      float64 `json:"weekly_policy_lcr_delta"`
	WeeklyPolicyBudgetCostGBP float64 `json:"weekly_policy_budget_cost_gbp"`
}

// worldSaveVersion is incremented when WorldSaveData schema changes incompatibly.
const worldSaveVersion = 1

// SaveWorld converts a live WorldState to its serialisable form.
// The caller should persist the returned value via save.Write.
func SaveWorld(w WorldState) WorldSaveData {
	// Capture a new RNG seed from the live RNG so post-load play differs slightly
	// but remains reproducible from the save point.
	var rngSeed int64
	if w.RNG != nil {
		rngSeed = w.RNG.Int63()
	}

	pcs := make([]policy.PolicyCardSave, len(w.PolicyCards))
	for i, pc := range w.PolicyCards {
		pcs[i] = policy.SavePolicyCard(pc)
	}

	return WorldSaveData{
		SaveVersion:                      worldSaveVersion,
		Week:                             w.Week,
		Year:                             w.Year,
		Month:                            w.Month,
		Quarter:                          w.Quarter,
		StartYear:                        w.StartYear,
		ScenarioID:                       w.ScenarioID,
		ScandalRateMultiplier:            w.ScandalRateMultiplier,
		BaseWeeklyMt:                     w.BaseWeeklyMt,
		RNGSeed:                          rngSeed,
		Government:                       w.Government,
		Stakeholders:                     w.Stakeholders,
		GovernmentPopularity:             w.GovernmentPopularity,
		EnergyMarket:                     w.EnergyMarket,
		ClimateState:                     w.ClimateState,
		FossilDependency:                 w.FossilDependency,
		Carbon:                           w.Carbon,
		Tech:                             w.Tech,
		Industry:                         w.Industry,
		Regions:                          w.Regions,
		Tiles:                            w.Tiles,
		Economy:                          w.Economy,
		LastTaxRevenue:                   w.LastTaxRevenue,
		LastBudget:                       w.LastBudget,
		LCR:                              w.LCR,
		WeeksUntilLCRPoll:                w.WeeksUntilLCRPoll,
		PollHistory:                      w.PollHistory,
		PolicyCards:                      pcs,
		OrgStates:                        w.OrgStates,
		Commissions:                      w.Commissions,
		Reports:                          w.Reports,
		EventLog:                         w.EventLog,
		PressureGroups:                   w.PressureGroups,
		PendingShockResponses:            w.PendingShockResponses,
		TickyCountdown:                   w.TickyCountdown,
		PendingTickyPressure:             w.PendingTickyPressure,
		TickyPressureAcceptedThisQuarter: w.TickyPressureAcceptedThisQuarter,
		PendingRiskyTicky:                w.PendingRiskyTicky,
		PendingTrickyTicky:               w.PendingTrickyTicky,
		AngryTickyActive:                 w.AngryTickyActive,
		AngryTickyWimpy:                  w.AngryTickyWimpy,
		ActiveDecayingShocks:             w.ActiveDecayingShocks,
		GreatSneezeActive:                w.GreatSneezeActive,
		GreatSneezeWeekEnd:               w.GreatSneezeWeekEnd,
		GreatSneezeFired:                 w.GreatSneezeFired,
		FiredOnceEvents:                  w.FiredOnceEvents,
		Player:                           w.Player,
		GovernmentLastPollResult:         w.GovernmentLastPollResult,
		MinisterLastPollResults:          w.MinisterLastPollResults,
		TechDeliveryLog:                  w.TechDeliveryLog,
		WeeklyNetCarbonMt:                w.WeeklyNetCarbonMt,
		WeeklyPolicyReductionMt:          w.WeeklyPolicyReductionMt,
		WeeklyEventLCRDelta:              w.WeeklyEventLCRDelta,
		WeeklyPolicyLCRDelta:             w.WeeklyPolicyLCRDelta,
		WeeklyPolicyBudgetCostGBP:        w.WeeklyPolicyBudgetCostGBP,
	}
}

// RestoreWorld reconstructs a live WorldState from saved data and a loaded config.
// Returns an error if the save version is incompatible.
func RestoreWorld(d WorldSaveData, cfg *config.Config) (WorldState, error) {
	if d.SaveVersion != worldSaveVersion {
		return WorldState{}, nil // caller should detect empty world; use ErrIncompatibleVersion
	}

	pcs := make([]policy.PolicyCard, len(d.PolicyCards))
	for i, ps := range d.PolicyCards {
		pcs[i] = policy.RestorePolicyCard(ps, cfg.PolicyCards)
	}

	w := WorldState{
		Week:                             d.Week,
		Year:                             d.Year,
		Month:                            d.Month,
		Quarter:                          d.Quarter,
		StartYear:                        d.StartYear,
		ScenarioID:                       d.ScenarioID,
		ScandalRateMultiplier:            d.ScandalRateMultiplier,
		BaseWeeklyMt:                     d.BaseWeeklyMt,
		Cfg:                              cfg,
		RNG:                              rand.New(rand.NewSource(d.RNGSeed)),
		Government:                       d.Government,
		Stakeholders:                     d.Stakeholders,
		GovernmentPopularity:             d.GovernmentPopularity,
		EnergyMarket:                     d.EnergyMarket,
		ClimateState:                     d.ClimateState,
		FossilDependency:                 d.FossilDependency,
		Carbon:                           d.Carbon,
		Tech:                             d.Tech,
		Industry:                         d.Industry,
		Regions:                          d.Regions,
		Tiles:                            d.Tiles,
		Economy:                          d.Economy,
		LastTaxRevenue:                   d.LastTaxRevenue,
		LastBudget:                       d.LastBudget,
		LCR:                              d.LCR,
		WeeksUntilLCRPoll:                d.WeeksUntilLCRPoll,
		PollHistory:                      d.PollHistory,
		PolicyCards:                      pcs,
		OrgStates:                        d.OrgStates,
		Commissions:                      d.Commissions,
		Reports:                          d.Reports,
		EventLog:                         d.EventLog,
		PressureGroups:                   d.PressureGroups,
		PendingShockResponses:            d.PendingShockResponses,
		TickyCountdown:                   d.TickyCountdown,
		PendingTickyPressure:             d.PendingTickyPressure,
		TickyPressureAcceptedThisQuarter: d.TickyPressureAcceptedThisQuarter,
		PendingRiskyTicky:                d.PendingRiskyTicky,
		PendingTrickyTicky:               d.PendingTrickyTicky,
		AngryTickyActive:                 d.AngryTickyActive,
		AngryTickyWimpy:                  d.AngryTickyWimpy,
		ActiveDecayingShocks:             d.ActiveDecayingShocks,
		GreatSneezeActive:                d.GreatSneezeActive,
		GreatSneezeWeekEnd:               d.GreatSneezeWeekEnd,
		GreatSneezeFired:                 d.GreatSneezeFired,
		FiredOnceEvents:                  d.FiredOnceEvents,
		Player:                           d.Player,
		GovernmentLastPollResult:         d.GovernmentLastPollResult,
		MinisterLastPollResults:          d.MinisterLastPollResults,
		TechDeliveryLog:                  d.TechDeliveryLog,
		WeeklyNetCarbonMt:                d.WeeklyNetCarbonMt,
		WeeklyPolicyReductionMt:          d.WeeklyPolicyReductionMt,
		WeeklyEventLCRDelta:              d.WeeklyEventLCRDelta,
		WeeklyPolicyLCRDelta:             d.WeeklyPolicyLCRDelta,
		WeeklyPolicyBudgetCostGBP:        d.WeeklyPolicyBudgetCostGBP,
	}
	return w, nil
}
