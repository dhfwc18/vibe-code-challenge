package config

// ScenarioID identifies the three playable campaign starting points.
type ScenarioID string

const (
	ScenarioHumbleBeginnings ScenarioID = "humble_beginnings" // 2010 center-right start
	ScenarioRisingStorm      ScenarioID = "rising_storm"      // 2019 populist right start
	ScenarioCrossroads       ScenarioID = "crossroads"        // 2026 center-left start
)

// StakeholderOverride forces a specific stakeholder to be unlocked and/or
// assigned a different role at scenario start, regardless of EntryTiming.
type StakeholderOverride struct {
	StakeholderID string // must match a StakeholderSeed.ID
	ForceUnlock   bool   // unlock immediately at scenario start
	ForceRole     Role   // if non-empty, assign this role instead of the seed default
}

// ScenarioConfig holds the starting conditions for one campaign start point.
// All numeric fields represent initial world state at game start (week 0 of
// the scenario, regardless of StartYear).
type ScenarioConfig struct {
	ID                    ScenarioID
	Name                  string
	ShortName             string
	Description           string
	StartYear             int
	InitialParty          Party
	InitialPopularity     float64 // GovernmentPopularity [0,100]
	InitialFossilDep      float64 // FossilDependency [0,100]
	InitialPlayerRep      float64 // LowCarbonReputation [0,100]
	InitialCarbonMt       float64 // annual MtCO2e at scenario start
	InitialBudget         float64 // GBP millions available at start
	ElectionDueWeek       int     // first election due at this game week (relative to week 0)
	ScandalRateMultiplier float64 // multiplied against base scandal probability each week
	StakeholderOverrides  []StakeholderOverride
	PreUnlockTech         map[Technology]float64 // technology -> initial maturity [0,100]
	FiredOnceEvents       []string               // event IDs treated as already fired at start
}

// Scenarios is the ordered list of playable campaign start points.
// The first entry is the default for new games.
var Scenarios = []ScenarioConfig{
	scenarioHumbleBeginnings,
	scenarioRisingStorm,
	scenarioCrossroads,
}

// ScenarioByID returns the ScenarioConfig for the given ID.
// Returns HumbleBeginnings if the ID is not found.
func ScenarioByID(id ScenarioID) ScenarioConfig {
	for _, s := range Scenarios {
		if s.ID == id {
			return s
		}
	}
	return scenarioHumbleBeginnings
}

// ---------------------------------------------------------------------------
// Scenario definitions
// ---------------------------------------------------------------------------

// scenarioHumbleBeginnings is the default 2010 center-right campaign start.
// The Union Party has just won the Taitan general election; the civil service
// net zero team is small and underfunded. Everything is to play for.
var scenarioHumbleBeginnings = ScenarioConfig{
	ID:        ScenarioHumbleBeginnings,
	Name:      "Humble Beginnings",
	ShortName: "2010",
	Description: "Taitan 2010. The Union Party has won a narrow majority. " +
		"Cavendish leads a center-right cabinet cautiously open to the energy " +
		"transition. You are a junior civil servant with 40 years to reach net zero. " +
		"Budget is tight; political capital is scarcer still.",
	StartYear:             2010,
	InitialParty:          PartyRight,
	InitialPopularity:     52.0,
	InitialFossilDep:      70.0,
	InitialPlayerRep:      50.0,
	InitialCarbonMt:       590.0,
	InitialBudget:         500.0,
	ElectionDueWeek:       260,
	ScandalRateMultiplier: 1.0,
	// Default Union Party cabinet (all TimingStart): Cavendish, Drake, Stafford, Holm.
	StakeholderOverrides: nil,
	PreUnlockTech:        nil,
	FiredOnceEvents:      nil,
}

// scenarioRisingStorm is the 2019 populist center-right campaign start.
// Noris Jackson leads an unstable government after an unexpected leadership
// contest. Dizzy Truscott is Chancellor -- ambitious and ideologically combative.
// The Texit crisis is behind Taitan but the economic disruption lingers.
// Scandal probability is elevated; the government's grip on power is fragile.
var scenarioRisingStorm = ScenarioConfig{
	ID:        ScenarioRisingStorm,
	Name:      "Rising Storm",
	ShortName: "2019",
	Description: "Taitan 2019. Noris Jackson leads a fractious Union Party government " +
		"after a chaotic leadership contest. Dawn Truscott at the Treasury is already " +
		"briefing against him. The Texit vote is settled -- narrowly Remain -- but the " +
		"economy is still absorbing the shock. You have 31 years left, a thin majority, " +
		"and a Chancellor who reads the net zero budget as wasted money.",
	StartYear:             2019,
	InitialParty:          PartyRight,
	InitialPopularity:     45.0,
	InitialFossilDep:      62.0,
	InitialPlayerRep:      50.0,
	InitialCarbonMt:       470.0,
	InitialBudget:         420.0,
	ElectionDueWeek:       104, // unstable government; election due approx 2 years in
	ScandalRateMultiplier: 2.0,
	StakeholderOverrides: []StakeholderOverride{
		// Noris Jackson is normally TimingMid ForeignSecretary (weeks 60-80).
		// For Rising Storm he has won the leadership contest.
		{StakeholderID: "noris_jackson", ForceUnlock: true, ForceRole: RoleLeader},
		// Dawn Truscott is normally TimingMid Chancellor (weeks 260-460).
		// For Rising Storm she is appointed Chancellor from day one.
		{StakeholderID: "dawn_truscott", ForceUnlock: true, ForceRole: RoleChancellor},
		// Andrew Stafford (TimingStart Right ForeignSecretary) and Rupert Holm
		// (TimingStart Right Energy) fill the remaining cabinet posts unchanged.
	},
	PreUnlockTech: map[Technology]float64{
		TechOffshoreWind:  35.0, // nine years of deployment since 2010
		TechOnshoreWind:   42.0,
		TechSolarPV:       20.0,
		TechNuclear:       45.0, // existing fleet; slow natural growth
		TechHeatPumps:     10.0,
		TechEVs:           12.0,
		TechHydrogen:      5.0,
		TechIndustrialCCS: 7.0,
	},
	// Texit chain events are backstory; treat as already fired.
	FiredOnceEvents: []string{
		"texit_campaign_begins",
		"texit_sovereignty_pivot",
		"texit_settled",
	},
}

// scenarioCrossroads is the 2026 center-left campaign start.
// The Common Wealth has returned to government after the Union Party's
// fractious decade. David Reeve leads a modernising administration with
// genuine green ambitions but a stretched budget and a polarised parliament.
// More technology is mature; less time remains.
var scenarioCrossroads = ScenarioConfig{
	ID:        ScenarioCrossroads,
	Name:      "Crossroads",
	ShortName: "2026",
	Description: "Taitan 2026. The Common Wealth is back in government after sixteen " +
		"years in opposition. David Reeve inherits a country halfway through its energy " +
		"transition and not moving fast enough. You have 24 years to close the gap. " +
		"The technologies exist; the politics are harder.",
	StartYear:             2026,
	InitialParty:          PartyLeft,
	InitialPopularity:     55.0,
	InitialFossilDep:      52.0,
	InitialPlayerRep:      50.0,
	InitialCarbonMt:       380.0,
	InitialBudget:         480.0,
	ElectionDueWeek:       260,
	ScandalRateMultiplier: 1.0,
	StakeholderOverrides: []StakeholderOverride{
		// David Reeve is TimingSuccessor (only available after JJ Cameron's departure
		// via ElectoralFatigue). For Crossroads, JJ Cameron's era is backstory.
		{StakeholderID: "david_reeve", ForceUnlock: true, ForceRole: RoleLeader},
		// George Harmon, John Ashworth, Claire Blackwell are TimingStart Left and
		// fill the remaining cabinet posts without overrides.
	},
	PreUnlockTech: map[Technology]float64{
		TechOffshoreWind:  55.0, // sixteen years of deployment; well established
		TechOnshoreWind:   58.0,
		TechSolarPV:       42.0,
		TechNuclear:       52.0, // new builds underway alongside ageing fleet
		TechHeatPumps:     22.0,
		TechEVs:           30.0,
		TechHydrogen:      12.0,
		TechIndustrialCCS: 10.0,
	},
	// Texit chain and earlier scripted events are backstory.
	FiredOnceEvents: []string{
		"texit_campaign_begins",
		"texit_sovereignty_pivot",
		"texit_settled",
		"the_coming_winter",
		"the_amber_coast_war",
	},
}
