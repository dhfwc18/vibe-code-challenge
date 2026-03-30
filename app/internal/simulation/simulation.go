package simulation

import (
	"math"
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
	"github.com/vibe-code-challenge/twenty-fifty/internal/mathutil"
	"github.com/vibe-code-challenge/twenty-fifty/internal/player"
	"github.com/vibe-code-challenge/twenty-fifty/internal/policy"
	"github.com/vibe-code-challenge/twenty-fifty/internal/polling"
	"github.com/vibe-code-challenge/twenty-fifty/internal/region"
	"github.com/vibe-code-challenge/twenty-fifty/internal/reputation"
	"github.com/vibe-code-challenge/twenty-fifty/internal/save"
	"github.com/vibe-code-challenge/twenty-fifty/internal/stakeholder"
	"github.com/vibe-code-challenge/twenty-fifty/internal/technology"
)

// lobbyAPCost is the action point cost for the LobbyMinister action.
const lobbyAPCost = 3

// ministerGraceWeeks is the number of weeks after appointment during which a
// minister is not held accountable for popularity (grace period).
const ministerGraceWeeks = 4

// signWeightMinor/Moderate/Major are multipliers applied to ideology conflict
// accumulation based on the policy card's Significance classification.
const signWeightMinor    = 1.0
const signWeightModerate = 2.0
const signWeightMajor    = 4.0

// consultancyAffinityBonusPerWeek is the weekly relationship bonus applied to
// affiliated orgs when a minister with ConsultancyAffinity is governing.
const consultancyAffinityBonusPerWeek = 0.3

// consultancyAversionPenaltyPerWeek is the weekly relationship hit the player
// takes with an averse minister for each active Consultancy-type commission.
const consultancyAversionPenaltyPerWeek = 0.8

// ministerSackingThreshold is the Popularity level below which a cabinet
// minister enters UNDER_PRESSURE. Reduced when GovernmentPopularity is high.
const ministerSackingThreshold = 25.0

// ministerSackingThresholdHighPop is used when GovernmentPopularity > 60.
const ministerSackingThresholdHighPop = 20.0

// ministerSackingWeeks is consecutive weeks UNDER_PRESSURE before SACKED.
const ministerSackingWeeks = 3

// techDeliveryThreshold is the AccumulatedQuality level at which a company
// delivers a technology maturity boost and resets its quality accumulator.
const techDeliveryThreshold = 200.0

// ideologyConflictWeight scales the raw ideology-conflict score (0-200 range)
// into the accumulated IdeologyConflictScore per approved step.
const ideologyConflictWeight = 0.015

// ideologyConflictResignThreshold is the accumulated score at which a cabinet
// minister transitions to RESIGNED due to sustained ideology conflict.
const ideologyConflictResignThreshold = 8.0

// ideologyConflictDecayRate is the weekly fraction by which IdeologyConflictScore
// decays toward zero. At 0.05, the score halves in approximately 14 weeks.
const ideologyConflictDecayRate = 0.05

// lobbyBudgetEffect is the per-action multiplier applied to the lobbied minister's
// departments in LobbyEffects. 1.10 = 10% budget share boost for that dept this quarter.
const lobbyBudgetEffect = 1.10

// ---------------------------------------------------------------------------
// Calibration constants
// ---------------------------------------------------------------------------

// tickyStakeholderID is the unique ID of TD Tennison, the only stakeholder
// with the Ticky pressure special mechanic.
const tickyStakeholderID = "ticky_tennison"

// tickyTier1OrgID is the Tier 1 Murican org that Ticky unlocks on Accept/Negotiate.
const tickyTier1OrgID = "murican_growth_alliance"

// tickyPressureMinWeeks is the minimum weeks between Ticky pressure events.
const tickyPressureMinWeeks = 6

// tickyPressureMaxVariance is added to tickyPressureMinWeeks via RNG to give
// a 6-10 week interval between pressure events.
const tickyPressureMaxVariance = 5

// tickyRelationshipAccept is the relationship delta when the player accepts.
const tickyRelationshipAccept = 8.0

// tickyRelationshipDecline is the relationship delta when the player declines.
const tickyRelationshipDecline = -5.0

// tickyRelationshipNegotiate is the relationship delta when the player negotiates.
const tickyRelationshipNegotiate = -2.0

const (
	// baselineYearlyMt is the 2010 Taitan baseline annual emissions.
	baselineYearlyMt = 590.0
	// baseWeeklyMt is the implied weekly emission rate before any policy reductions.
	baseWeeklyMt = baselineYearlyMt / 52.0

	// initialGovtPopularity seeds the hidden GovernmentPopularity at game start.
	initialGovtPopularity = 52.0

	// initialFossilDependency seeds FossilDependency at game start (~70% for 2010).
	initialFossilDependency = 70.0

	// initialElectionDueWeek places the first scheduled election at ~May 2015.
	// Game starts January 2010; 5 years = 260 weeks.
	initialElectionDueWeek = 260

	// pollWeeklyProb is the Bernoulli draw probability for a government/regional poll.
	pollWeeklyProb = 0.25

	// maxPollHistory is the maximum number of poll snapshots retained in PollHistory.
	// Older entries are dropped to prevent unbounded memory growth over a 40-year game.
	maxPollHistory = 200

	// maxTechDeliveryLog is the maximum number of tech delivery milestone strings retained.
	maxTechDeliveryLog = 100

	// maxInsightReports is the maximum number of delivered insight reports retained.
	maxInsightReports = 50
)

// ---------------------------------------------------------------------------
// Extended Ticky mechanics constants
// ---------------------------------------------------------------------------

// riskyTickyMinWeek is the earliest week Risky Ticky can fire (after game start).
const riskyTickyMinWeek = 52

// riskyTickyWeeklyProb is the per-week probability of Risky Ticky firing when
// a centrist or far-left government is in power.
const riskyTickyWeeklyProb = 0.03

// riskyTickyEndorseRelBoost is the relationship gain with Ticky on endorsement.
const riskyTickyEndorseRelBoost = 15.0

// riskyTickyEndorseCabinetPenalty is the relationship penalty applied to each
// current cabinet member when the player publicly endorses Ticky's party.
const riskyTickyEndorseCabinetPenalty = -8.0

// riskyTickyDeclinePenalty is the relationship penalty with Ticky on decline.
const riskyTickyDeclinePenalty = -3.0

// trickyTickyWeeklyProb is the per-week probability of Tricky Ticky firing when
// the far-right is in government.
const trickyTickyWeeklyProb = 0.05

// trickyTickyAcceptBudgetBoost is the GBP-millions budget added to DeptPower on
// accepting the Murican contract offer.
const trickyTickyAcceptBudgetBoost = 20.0

// trickyTickyAcceptCabinetPenalty is the relationship penalty applied to each
// far-right cabinet member on accepting (they resent being bypassed).
const trickyTickyAcceptCabinetPenalty = -5.0

// trickyTickyDeclinePenalty is the relationship penalty with Ticky on decline.
const trickyTickyDeclinePenalty = -3.0

// angryTickyRelThreshold is the Ticky relationship score below which Angry
// Ticky activates (a permanent policy-approval malus kicks in).
const angryTickyRelThreshold = 15.0

// angryTickyPolicyMalus is added to MinRelationshipScore for every approval
// step while Angry Ticky is active, making all policies harder to pass.
const angryTickyPolicyMalus = 15.0

// angryTickyDamageAPCost is the AP cost of the DamageTickyReputation action.
const angryTickyDamageAPCost = 3

// angryTickyDamageRelDelta is applied to Ticky each time the player uses
// the reputation-damage action.
const angryTickyDamageRelDelta = -20.0

// angryTickyDamageCabinetSplash is the relationship penalty applied to each
// unlocked non-Ticky FarRight stakeholder on each reputation-damage action.
const angryTickyDamageCabinetSplash = -3.0

// wimpyTickyRelThreshold is the Ticky relationship level below which the
// Wimpy Ticky outcome fires: Ticky backs down and the angry malus is lifted.
// RelationshipScore is clamped to [0,100] so the threshold must be >= 0.
const wimpyTickyRelThreshold = 5.0

// greatSneezeStartYear is the calendar year the Great Sneeze activates.
const greatSneezeStartYear = 2019

// greatSneezeEndYear is the calendar year the Great Sneeze deactivates.
const greatSneezeEndYear = 2021

// greatSneezeWeeklyPopPenalty is the weekly GovernmentPopularity delta
// applied every week the Great Sneeze is active.
const greatSneezeWeeklyPopPenalty = -0.3

// greatSneezeEmergencyBudgetBoostGBP is the one-time budget bonus per
// department when the Great Sneeze fires.
const greatSneezeEmergencyBudgetBoostGBP = 12.0

// ---------------------------------------------------------------------------
// WorldState
// ---------------------------------------------------------------------------

// ActiveDecayingShock is a runtime market-effect created when an event with
// a config.DecayingShockConfig fires. Each week the price deltas are applied
// to the energy market, multiplied by DecayRate, until WeeksRemaining hits zero.
type ActiveDecayingShock struct {
	EventID         string
	GasPctThisWeek  float64 // current weekly gas price % change
	OilPctThisWeek  float64 // current weekly oil price % change
	ElecPctThisWeek float64 // current weekly electricity price % change
	DecayRate       float64 // multiplied against each pct every week
	WeeksRemaining  int
}

// WorldState is the single source of truth for a game turn. All simulation
// logic reads and writes through AdvanceWeek, which returns a new copy each
// week. The RNG pointer and Cfg pointer are the only shared state across
// copies; both are treated as immutable (Cfg) or sequentially-advanced (RNG).
type WorldState struct {
	// Clock
	Week      int // absolute week; 0 = initial state, 1 = first processed week
	Year      int // calendar year derived from StartYear + Week/52
	Month     int // calendar month 1-12 derived from week-of-year
	Quarter   int // 1-4
	StartYear int // calendar year the scenario begins (e.g. 2010, 2019, 2026)

	// Scenario metadata
	ScenarioID            config.ScenarioID // which campaign start point is active
	ScandalRateMultiplier float64           // multiplier on base scandal probability (1.0 = normal)
	BaseWeeklyMt          float64           // scenario-calibrated baseline weekly emissions (MtCO2e)

	// Static config (shared pointer; never mutated at runtime)
	Cfg *config.Config

	// RNG (shared pointer; advances in place; do not copy for branching)
	RNG *rand.Rand

	// Political
	Government           government.GovernmentState
	Stakeholders         []stakeholder.Stakeholder
	GovernmentPopularity float64 // hidden true value; 0-100

	// Environment
	EnergyMarket     energy.EnergyMarket
	ClimateState     climate.ClimateState
	FossilDependency float64 // derived from energy mix; 0-100
	Carbon           carbon.CarbonBudgetState

	// Technology and industry
	Tech     technology.TechTracker
	Industry industry.IndustryState

	// Geography
	Regions []region.Region
	Tiles   []region.Tile

	// Economy
	Economy        economy.EconomyState
	LastTaxRevenue economy.TaxRevenue      // updated quarterly
	LastBudget     economy.BudgetAllocation // updated quarterly

	// Reputation
	LCR               reputation.LowCarbonReputation
	WeeksUntilLCRPoll int

	// Polling history
	PollHistory []polling.PollSnapshot

	// Policy cards (all cards, all states)
	PolicyCards []policy.PolicyCard

	// Evidence: advisory organisations, active commissions, delivered reports
	OrgStates   []evidence.OrgState
	Commissions []evidence.Commission
	Reports     []evidence.InsightReport

	// Events
	EventLog              event.EventLog
	PressureGroups        []event.PressureGroup
	PendingShockResponses []event.PendingShockResponse

	// Ticky pressure mechanic: tracks recurring pressure events from TD Tennison.
	TickyCountdown               int  // weeks until next pressure event; only decrements when Ticky is in cabinet
	PendingTickyPressure         bool // true when Ticky has applied pressure and player has not yet responded
	TickyPressureAcceptedThisQuarter bool // true if player accepted or negotiated this quarter; reset at quarter-end

	// Extended Ticky mechanics.
	PendingRiskyTicky   bool // Risky Ticky endorsement prompt awaiting player response
	PendingTrickyTicky  bool // Tricky Ticky Murican contract offer awaiting player response
	AngryTickyActive    bool // Angry Ticky is in effect; policy approval malus applied
	AngryTickyWimpy     bool // Wimpy Ticky triggered; malus has been lifted

	// Decaying energy market shocks.
	ActiveDecayingShocks []ActiveDecayingShock

	// Great Sneeze pandemic event.
	GreatSneezeActive  bool // true while Great Sneeze is ongoing
	GreatSneezeWeekEnd int  // week number when Great Sneeze ends
	GreatSneezeFired   bool // prevents repeat triggers across save/load

	// FiredOnceEvents tracks IDs of events with TriggerAtYear that have already
	// fired, preventing them from re-firing on save/load.
	FiredOnceEvents map[string]bool

	// Player
	Player player.CivilServant

	// Player-visible poll snapshots
	GovernmentLastPollResult float64            // most recent noisy govt approval sample (sigma=3)
	MinisterLastPollResults  map[string]float64 // keyed by stakeholder ID; sigma=5

	// Tech delivery milestones (persistent; never reset)
	TechDeliveryLog []string // one entry per company delivery event, e.g. "company_id delivered tech boost for OFFSHORE_WIND"

	// Weekly transient accumulators -- reset at the top of each AdvanceWeek call.
	WeeklyNetCarbonMt         float64 // net carbon emitted this week (base minus reductions)
	WeeklyPolicyReductionMt   float64 // total carbon removed by active policies
	WeeklyEventLCRDelta       float64 // LCR delta from fired events
	WeeklyPolicyLCRDelta      float64 // direct LCR delta from active policy LCRDeltaPerWeek fields
	WeeklyPolicyBudgetCostGBP float64 // total budget draw from active policies this week
}

// ---------------------------------------------------------------------------
// Action
// ---------------------------------------------------------------------------

// Action describes a player decision applied during Phase 13 (Player Action Phase).
type Action struct {
	Type   player.ActionType
	Target string // ID of the target entity (policy ID, stakeholder ID, org ID, etc.)
	Detail string // supplementary info (e.g. InsightType for commissions, staff ID for hire)
}

// ---------------------------------------------------------------------------
// SimulationReport
// ---------------------------------------------------------------------------

// SimulationReport summarises the outcome of a HeadlessRun. Used by tests
// and AI-driven playtesting to assert simulation invariants.
type SimulationReport struct {
	WeeksRun          int
	EventsFired       int
	PollsTaken        int
	FinalCarbon       carbon.CarbonBudgetState
	FinalGovtPop      float64
	FinalLCR          float64
	MinBudgetValueGBP float64  // minimum department budget seen (GBP millions)
	StakeholderIssues []string // populated if any stakeholder has an invalid state
}

// ---------------------------------------------------------------------------
// NewWorld
// ---------------------------------------------------------------------------

// NewWorld seeds a WorldState for the default (Humble Beginnings 2010) campaign.
// It is a convenience wrapper around NewWorldFromScenario.
func NewWorld(cfg *config.Config, masterSeed save.MasterSeed) WorldState {
	return NewWorldFromScenario(cfg, masterSeed, config.ScenarioByID(config.ScenarioHumbleBeginnings))
}

// NewWorldFromScenario seeds a complete initial WorldState from config, a master
// seed, and a ScenarioConfig. Week=0; Year=scenario.StartYear; Quarter=1.
// The first call to AdvanceWeek moves to Week=1.
func NewWorldFromScenario(cfg *config.Config, masterSeed save.MasterSeed, scenario config.ScenarioConfig) WorldState {
	rng := rand.New(rand.NewSource(int64(masterSeed.DeriveSubSeed("simulation"))))

	scandalMult := scenario.ScandalRateMultiplier
	if scandalMult <= 0 {
		scandalMult = 1.0
	}
	baseWeekly := scenario.InitialCarbonMt / 52.0
	if baseWeekly <= 0 {
		baseWeekly = baselineYearlyMt / 52.0
	}

	w := WorldState{
		Week:                     0,
		Year:                     scenario.StartYear,
		Month:                    1,
		Quarter:                  1,
		StartYear:                scenario.StartYear,
		ScenarioID:               scenario.ID,
		ScandalRateMultiplier:    scandalMult,
		BaseWeeklyMt:             baseWeekly,
		Cfg:                      cfg,
		RNG:                      rng,
		GovernmentPopularity:     scenario.InitialPopularity,
		GovernmentLastPollResult: scenario.InitialPopularity,
		FossilDependency:         scenario.InitialFossilDep,
		WeeksUntilLCRPoll:        10 + rng.Intn(7),
		TickyCountdown:           tickyPressureMinWeeks + rng.Intn(tickyPressureMaxVariance),
		MinisterLastPollResults:  make(map[string]float64),
		TechDeliveryLog:          []string{},
	}

	// Government: ruling party wins the scenario-start election.
	w.Government = government.NewGovernment(scenario.InitialParty, scenario.ElectionDueWeek)

	// Build a set of roles claimed by scenario overrides so the default pass
	// can skip them (they will be assigned in the override pass below).
	reservedRoles := make(map[config.Role]bool)
	for _, o := range scenario.StakeholderOverrides {
		if o.ForceRole != "" {
			reservedRoles[o.ForceRole] = true
		}
	}

	// Stakeholders: seed all; unlock TimingStart figures; assign ruling party cabinet
	// unless the role is reserved for an override.
	w.Stakeholders = stakeholder.SeedStakeholders(cfg.Stakeholders)
	for i, s := range w.Stakeholders {
		if s.EntryTiming != config.TimingStart {
			continue
		}
		w.Stakeholders[i] = stakeholder.UnlockStakeholder(s, 0)
		w.Stakeholders[i].State = stakeholder.MinisterStateAppointed
		if s.Party == scenario.InitialParty && !reservedRoles[s.Role] {
			w.Government = government.AssignMinister(w.Government, s.Role, s.ID)
		}
	}

	// Override pass: force-unlock stakeholders and assign them to overridden roles.
	for _, o := range scenario.StakeholderOverrides {
		if !o.ForceUnlock {
			continue
		}
		for i := range w.Stakeholders {
			if w.Stakeholders[i].ID != o.StakeholderID {
				continue
			}
			// Force unlock regardless of EntryWeekMin (scenario backstory justification).
			w.Stakeholders[i].IsUnlocked = true
			w.Stakeholders[i].State = stakeholder.MinisterStateAppointed
			role := w.Stakeholders[i].Role
			if o.ForceRole != "" {
				role = o.ForceRole
			}
			w.Government = government.AssignMinister(w.Government, role, w.Stakeholders[i].ID)
			break
		}
	}

	// Energy market: anchored to scenario start values.
	w.EnergyMarket = energy.NewMarket()

	// Climate: zero cumulative stock at scenario start.
	w.ClimateState = climate.DeriveClimateState(0.0)

	// Carbon: zero state; limits checked against cfg.CarbonBudgets each year.
	w.Carbon = carbon.CarbonBudgetState{}

	// Technology: seeded from logistic curve definitions; then apply scenario overrides.
	w.Tech = technology.NewTechTracker(cfg.Technologies)
	for tech, maturity := range scenario.PreUnlockTech {
		if _, ok := w.Tech.Maturities[tech]; ok {
			w.Tech.Maturities[tech] = maturity
		}
	}

	// Geography.
	w.Regions = region.SeedRegions(cfg.Regions)
	w.Tiles = region.SeedTiles(cfg.Tiles)

	// Economy.
	w.Economy = economy.NewEconomyState()

	// Reputation: start at scenario initial player rep.
	w.LCR = reputation.NewLCR()
	w.LCR.Value = scenario.InitialPlayerRep
	w.LCR.LastPollResult = scenario.InitialPlayerRep

	// Policies: all cards start in DRAFT.
	w.PolicyCards = policy.SeedPolicyCards(cfg.PolicyCards)

	// Industry: all companies seeded inactive.
	w.Industry = industry.SeedIndustry(cfg.Companies)

	// Evidence: org relationship states.
	w.OrgStates = evidence.SeedOrgStates(cfg.Organisations)

	// Events: empty log, default pressure groups.
	w.EventLog = event.NewEventLog()
	w.PressureGroups = event.DefaultPressureGroups()

	// Pre-fire scenario backstory events so they do not re-trigger in-game.
	if len(scenario.FiredOnceEvents) > 0 {
		w.FiredOnceEvents = make(map[string]bool, len(scenario.FiredOnceEvents))
		for _, id := range scenario.FiredOnceEvents {
			w.FiredOnceEvents[id] = true
		}
	}

	// Player.
	w.Player = player.NewCivilServant()

	// Initial budget allocation (recomputed each quarter).
	w.LastTaxRevenue = economy.ComputeTaxRevenue(w.Economy, 1, scenario.StartYear)
	w.LastBudget = economy.AllocateBudget(
		w.LastTaxRevenue,
		baseDeptFractions(),
		neutralMinisterWeights(),
		reputation.ChainToBudgetModifier(w.LCR.Value),
		w.Economy.LobbyEffects,
	)

	return w
}

// ---------------------------------------------------------------------------
// AdvanceWeek
// ---------------------------------------------------------------------------

// AdvanceWeek executes the full 18-phase weekly pipeline and returns the updated
// WorldState along with any events that fired this week.
// Pass nil (or an empty slice) for actions in headless/AI mode.
func AdvanceWeek(w WorldState, actions []Action) (WorldState, []event.EventEntry) {
	// Reset weekly transients.
	w.WeeklyNetCarbonMt = w.BaseWeeklyMt
	w.WeeklyPolicyReductionMt = 0
	w.WeeklyEventLCRDelta = 0
	w.WeeklyPolicyLCRDelta = 0
	w.WeeklyPolicyBudgetCostGBP = 0

	// Phase 1: Clock Advance.
	w = phaseClockAdvance(w)

	// Phase 2: Climate and Fossil Dependency Update.
	w = phaseClimateAndFossilUpdate(w)

	// Phase 3: Global Events Roll.
	var firedEvents []event.EventEntry
	w, firedEvents = phaseGlobalEventRoll(w)

	// Phase 3b: Time-gated events (Great Sneeze and other TriggerAtYear events).
	w = phaseTimeGatedEvents(w)

	// Phase 3c: Apply and decay active market shocks from war/crisis events.
	w = phaseDecayingShockTick(w)

	// Phase 4: Scandal and Pressure Roll.
	w = phaseScandalAndPressureRoll(w)
	w = phaseTickyPressureTick(w)
	w = phaseExtendedTickyMechanics(w)

	// Phase 5: Technology Progress Tick.
	w = phaseTechnologyProgressTick(w)

	// Phase 6: Regional World Tick.
	w = phaseRegionalWorldTick(w)

	// Phase 7: Tile Local Tick.
	w = phaseTileLocalTick(w)

	// Phase 8: Climate Event Impact on Tiles -- targeted tile deltas already
	// applied during Phase 3 via ResolvedEffect; no additional state change here.

	// Phase 9: Policy Resolution.
	w = phasePolicyResolution(w)

	// Phase 10: Carbon Budget Accounting.
	w = phaseCarbonBudgetAccounting(w)

	// Phase 11: Economy and Tax Revenue Tick.
	w = phaseEconomyTick(w)

	// Phase 11b: Great Sneeze ongoing effects (weekly popularity drain while active).
	w = phaseGreatSneezeTick(w)

	// Phase 12: Polling Check.
	w = phasePollingCheck(w)

	// Phase 13: Consequence Resolution (policy approval evaluation for cards
	// already UNDER_REVIEW). Runs before player actions so that policies
	// submitted this tick are not evaluated until the following week.
	w = phaseConsequenceResolution(w)

	// Phase 14: Player Action Phase.
	w = phasePlayerActions(w, actions)

	// Phase 15: Minister Health Check (includes Angry Ticky trigger detection).
	w = phaseMinisterHealthCheck(w)

	// Phase 16: Minister Transitions.
	w = phaseMinisterTransitions(w)

	// Phase 17: Consultancy Delivery Check.
	w = phaseConsultancyDelivery(w)

	// Phase 18: End-of-Week Render -- no-op in headless/simulation mode.

	return w, firedEvents
}

// ---------------------------------------------------------------------------
// HeadlessRun
// ---------------------------------------------------------------------------

// HeadlessRun advances the simulation for the given number of weeks with no
// player input. Returns the final WorldState and a SimulationReport for use
// in integration tests and AI playtesting.
func HeadlessRun(w WorldState, weeks int) (WorldState, SimulationReport) {
	report := SimulationReport{MinBudgetValueGBP: math.MaxFloat64}

	for i := 0; i < weeks; i++ {
		var fired []event.EventEntry
		w, fired = AdvanceWeek(w, nil)
		report.EventsFired += len(fired)
		report.PollsTaken = len(w.PollHistory)

		// Track minimum budget value across all departments this week.
		for _, v := range w.LastBudget.Departments {
			if v < report.MinBudgetValueGBP {
				report.MinBudgetValueGBP = v
			}
		}
	}

	report.WeeksRun = weeks
	report.FinalCarbon = w.Carbon
	report.FinalGovtPop = w.GovernmentPopularity
	report.FinalLCR = w.LCR.Value

	// Validate stakeholder states.
	for _, s := range w.Stakeholders {
		if s.IsUnlocked && !isValidMinisterState(s.State) {
			report.StakeholderIssues = append(report.StakeholderIssues,
				"invalid state on "+s.ID+": "+string(s.State))
		}
	}

	if report.MinBudgetValueGBP == math.MaxFloat64 {
		report.MinBudgetValueGBP = 0
	}
	return w, report
}

// ---------------------------------------------------------------------------
// Phase 1: Clock Advance
// ---------------------------------------------------------------------------

func phaseClockAdvance(w WorldState) WorldState {
	w.Week++
	// Year: week 1-52 = StartYear, week 53-104 = StartYear+1, etc.
	w.Year = w.StartYear + w.Week/52
	// Quarter: 1-4 within each 52-week year.
	w.Quarter = 1 + (w.Week%52)/13
	// Month 1-12 derived from week-of-year: weekOfYear in [0,51] -> month in [1,12].
	weekOfYear := (w.Week - 1) % 52
	w.Month = 1 + weekOfYear*12/52

	// Election check: trigger when the scheduled week arrives.
	if government.IsElectionDue(w.Government, w.Week) {
		// Winner is the leading party in the most recent poll; default to current ruler.
		winner := w.Government.RulingParty
		if len(w.PollHistory) > 0 {
			winner = polling.LeadingParty(w.PollHistory[len(w.PollHistory)-1])
		}
		// Clear cabinet; set new ruling party; schedule next election 5 years out.
		w.Government = government.TriggerElection(w.Government, winner, w.Week+260)
		// Lock in the vote share at the moment of the election for the hemicycle display.
		if len(w.PollHistory) > 0 {
			src := w.PollHistory[len(w.PollHistory)-1].NationalPolls
			vs := make(map[config.Party]float64, len(src))
			for k, v := range src {
				vs[k] = v
			}
			w.Government.LastElectionVoteShare = vs
		}

		// Rebuild cabinet from winning party's unlocked non-terminal ministers.
		// Losing party ministers in ACTIVE/APPOINTED/UNDER_PRESSURE/OPPOSITION_SHADOW
		// move to OPPOSITION_SHADOW. Those already on BACKBENCH stay BACKBENCH.
		for i := range w.Stakeholders {
			s := w.Stakeholders[i]
			if !s.IsUnlocked || isTerminalState(s.State) {
				continue
			}
			if s.Party == winner {
				w.Government = government.AssignMinister(w.Government, s.Role, s.ID)
				w.Stakeholders[i].GraceWeeksRemaining = ministerGraceWeeks
				w.Stakeholders[i].State = stakeholder.MinisterStateAppointed
				w.Stakeholders[i].WeeksUnderPressure = 0
			} else {
				// Backbench ministers of the losing party stay on backbench.
				if s.State != stakeholder.MinisterStateBackbench {
					w.Stakeholders[i].State = stakeholder.MinisterStateOppositionShadow
				}
			}
		}
	}
	return w
}

// ---------------------------------------------------------------------------
// Phase 2: Climate and Fossil Dependency Update
// ---------------------------------------------------------------------------

func phaseClimateAndFossilUpdate(w WorldState) WorldState {
	w.ClimateState = climate.DeriveClimateState(w.Carbon.CumulativeStock)

	// FossilDependency: derived from the fraction of generation that is non-renewable.
	gridShare := w.EnergyMarket.RenewableGridShare
	gasFrac := (1.0 - gridShare/100.0) * 0.65
	oilFrac := (1.0 - gridShare/100.0) * 0.15
	w.FossilDependency = economy.DeriveFossilDependency(gasFrac, oilFrac)

	// Carbon trajectory: simple linear projection for the current year.
	weeksInYear := w.Week % 52
	if weeksInYear == 0 {
		weeksInYear = 52
	}
	w.Carbon.Trajectory = carbon.ProjectTrajectory(w.Carbon, weeksInYear)

	// Advance energy market ring buffers with this week's post-shock prices.
	// Policy-driven grid share changes will be plumbed in a simulation tuning pass;
	// zero gridShareDelta keeps prices stable apart from event shocks.
	w.EnergyMarket = energy.TickPrices(w.EnergyMarket, 0, 0, 0, 0)
	return w
}

// ---------------------------------------------------------------------------
// Phase 3: Global Events Roll
// ---------------------------------------------------------------------------

func phaseGlobalEventRoll(w WorldState) (WorldState, []event.EventEntry) {
	def, fired := event.DrawEvent(w.Cfg.Events, w.ClimateState.Level, w.FossilDependency, w.RNG)
	if !fired {
		return w, nil
	}

	resolved := event.ResolveEffect(
		def.BaseEffects,
		w.Cfg.Regions,
		w.Stakeholders,
		companyStateSlice(w.Industry),
		companyDefMap(w.Cfg),
	)

	// Apply global deltas.
	w.EnergyMarket = energy.ApplyShock(w.EnergyMarket, def.BaseEffects)
	w.GovernmentPopularity = mathutil.Clamp(w.GovernmentPopularity+resolved.GovtPopularityDelta, 0, 100)
	w.Economy.Value = mathutil.Clamp(w.Economy.Value+resolved.EconomyDelta, 0, 100)
	w.WeeklyEventLCRDelta += resolved.LCRDelta
	w.LCR.Value = mathutil.Clamp(w.LCR.Value+resolved.LCRDelta, 0, 100)

	// Apply region deltas.
	for i, r := range w.Regions {
		if d, ok := resolved.RegionDeltas[r.ID]; ok {
			w.Regions[i].InstallerCapacity = mathutil.Clamp(r.InstallerCapacity+d.InstallerCapacityDelta, 0, 10000)
			w.Regions[i].SkillsNetwork = mathutil.Clamp(r.SkillsNetwork+d.SkillsNetworkDelta, 0, 100)
		}
	}

	// Apply tile deltas.
	for i, t := range w.Tiles {
		if d, ok := resolved.TileDeltas[t.ID]; ok {
			w.Tiles[i] = region.ApplyClimateEventDamage(t, d.FuelPovertyDelta)
			w.Tiles[i].InsulationLevel = mathutil.Clamp(w.Tiles[i].InsulationLevel-d.InsulationDamage, 0, 100)
		}
	}

	// Apply stakeholder relationship deltas.
	for i, s := range w.Stakeholders {
		if d, ok := resolved.StakeholderDeltas[s.ID]; ok {
			w.Stakeholders[i] = stakeholder.TickRelationship(s, 0, d.RelDelta)
		}
	}

	// Apply company work rate and quality deltas from events.
	w = applyCompanyDeltas(w, resolved.CompanyDeltas)

	// Queue shock response opportunity if the event offers one.
	if def.OffersShockResponse {
		w.PendingShockResponses = append(w.PendingShockResponses, event.PendingShockResponse{
			EventDefID: def.ID,
			Week:       w.Week,
		})
	}

	// If the event has a decaying shock config, spawn an ActiveDecayingShock.
	if def.DecayingShock.MaxWeeks > 0 {
		shock := ActiveDecayingShock{
			EventID:         def.ID,
			GasPctThisWeek:  def.DecayingShock.InitialGasPctPerWeek,
			OilPctThisWeek:  def.DecayingShock.InitialOilPctPerWeek,
			ElecPctThisWeek: def.DecayingShock.InitialElecPctPerWeek,
			DecayRate:       def.DecayingShock.DecayRate,
			WeeksRemaining:  def.DecayingShock.MaxWeeks,
		}
		w.ActiveDecayingShocks = append(w.ActiveDecayingShocks, shock)
	}

	entry := event.EventEntry{DefID: def.ID, Name: def.Name, Week: w.Week, Effects: resolved}
	w.EventLog = event.AppendEventLog(w.EventLog, entry)

	// Design: at most 2 events per week. Attempt a second draw at CRITICAL/EMERGENCY climate.
	if w.ClimateState.Level >= carbon.ClimateLevelCritical {
		def2, fired2 := event.DrawEvent(w.Cfg.Events, w.ClimateState.Level, w.FossilDependency, w.RNG)
		if fired2 {
			resolved2 := event.ResolveEffect(
				def2.BaseEffects, w.Cfg.Regions, w.Stakeholders,
				companyStateSlice(w.Industry), companyDefMap(w.Cfg),
			)
			w.EnergyMarket = energy.ApplyShock(w.EnergyMarket, def2.BaseEffects)
			w.GovernmentPopularity = mathutil.Clamp(w.GovernmentPopularity+resolved2.GovtPopularityDelta, 0, 100)
			w.Economy.Value = mathutil.Clamp(w.Economy.Value+resolved2.EconomyDelta, 0, 100)
			w.WeeklyEventLCRDelta += resolved2.LCRDelta
			w.LCR.Value = mathutil.Clamp(w.LCR.Value+resolved2.LCRDelta, 0, 100)
			for i, r := range w.Regions {
				if d, ok := resolved2.RegionDeltas[r.ID]; ok {
					w.Regions[i].InstallerCapacity = mathutil.Clamp(r.InstallerCapacity+d.InstallerCapacityDelta, 0, 10000)
					w.Regions[i].SkillsNetwork = mathutil.Clamp(r.SkillsNetwork+d.SkillsNetworkDelta, 0, 100)
				}
			}
			for i, t := range w.Tiles {
				if d, ok := resolved2.TileDeltas[t.ID]; ok {
					w.Tiles[i] = region.ApplyClimateEventDamage(t, d.FuelPovertyDelta)
					w.Tiles[i].InsulationLevel = mathutil.Clamp(w.Tiles[i].InsulationLevel-d.InsulationDamage, 0, 100)
				}
			}
			for i, s := range w.Stakeholders {
				if d, ok := resolved2.StakeholderDeltas[s.ID]; ok {
					w.Stakeholders[i] = stakeholder.TickRelationship(s, 0, d.RelDelta)
				}
			}
			w = applyCompanyDeltas(w, resolved2.CompanyDeltas)
			if def2.OffersShockResponse {
				w.PendingShockResponses = append(w.PendingShockResponses, event.PendingShockResponse{
					EventDefID: def2.ID,
					Week:       w.Week,
				})
			}
			entry2 := event.EventEntry{DefID: def2.ID, Name: def2.Name, Week: w.Week, Effects: resolved2}
			w.EventLog = event.AppendEventLog(w.EventLog, entry2)
			return w, []event.EventEntry{entry, entry2}
		}
	}

	return w, []event.EventEntry{entry}
}

// ---------------------------------------------------------------------------
// Phase 4: Scandal and Pressure Roll
// ---------------------------------------------------------------------------

func phaseScandalAndPressureRoll(w WorldState) WorldState {
	// Number of scandal rolls per minister per week, shaped by ScandalRateMultiplier.
	// Default (1.0) = one roll; Rising Storm (2.0) = two rolls.
	multiplier := w.ScandalRateMultiplier
	if multiplier < 1.0 {
		multiplier = 1.0
	}
	scandalRolls := int(multiplier + 0.5) // round to nearest int

	// Scandal roll for each active unlocked stakeholder.
	for _, s := range w.Stakeholders {
		if !s.IsUnlocked || isTerminalState(s.State) {
			continue
		}
		for i := 0; i < scandalRolls; i++ {
			if event.RollScandal(s, s.WeeksUnderPressure, w.RNG) {
				w.GovernmentPopularity = mathutil.Clamp(w.GovernmentPopularity-2.0, 0, 100)
			}
		}
	}

	// Pressure groups apply weekly modifiers to GovernmentPopularity and LCR.
	results := event.ApplyPressureGroups(w.PressureGroups, w.Carbon.Trajectory, w.LCR.Value)
	for _, pr := range results {
		w.GovernmentPopularity = mathutil.Clamp(w.GovernmentPopularity+pr.GovtPopularityDelta, 0, 100)
		w.LCR.Value = mathutil.Clamp(w.LCR.Value+pr.LCRDelta, 0, 100)
	}
	return w
}

// ---------------------------------------------------------------------------
// Phase 5: Technology Progress Tick
// ---------------------------------------------------------------------------

func phaseTechnologyProgressTick(w WorldState) WorldState {
	bonuses := activePolicyRDBonuses(w.PolicyCards)
	for _, curve := range w.Cfg.Technologies {
		bonus := bonuses[curve.ID]
		w.Tech = technology.AdvanceTick(w.Tech, curve, bonus)
	}
	return w
}

// ---------------------------------------------------------------------------
// Phase 6: Regional World Tick
// ---------------------------------------------------------------------------

func phaseRegionalWorldTick(w WorldState) WorldState {
	// Organic skills growth: slow baseline learning over time.
	for i := range w.Regions {
		w.Regions[i].SkillsNetwork = mathutil.Clamp(w.Regions[i].SkillsNetwork+0.01, 0, 100)
	}

	// Industry company tick: accumulate quality for each active company.
	// Average installer capacity across all regions is used as the national proxy.
	avgCap := avgInstallerCapacity(w.Regions)
	defs := companyDefMap(w.Cfg)
	for defID, cs := range w.Industry.Companies {
		if !cs.IsActive {
			continue
		}
		def, ok := defs[defID]
		if !ok {
			continue
		}
		w.Industry = industry.TickCompany(w.Industry, defID, def, avgCap)
		// Tech delivery: fire when accumulated quality crosses the threshold.
		if cs, ok := w.Industry.Companies[defID]; ok && cs.IsActive && cs.AccumulatedQuality >= techDeliveryThreshold {
			tech := cs.ContractedTech
			w.Industry, w.Tech = industry.DeliverTech(w.Industry, defID, w.Tech, def)
			w.TechDeliveryLog = append(w.TechDeliveryLog,
				defID+" delivered tech maturity boost for "+string(tech))
			if len(w.TechDeliveryLog) > maxTechDeliveryLog {
				w.TechDeliveryLog = w.TechDeliveryLog[len(w.TechDeliveryLog)-maxTechDeliveryLog:]
			}
		}
	}
	return w
}

// ---------------------------------------------------------------------------
// Phase 7: Tile Local Tick
// ---------------------------------------------------------------------------

func phaseTileLocalTick(w WorldState) WorldState {
	hpMaturity := w.Tech.Maturity(config.TechHeatPumps)
	seasonal := seasonalMultiplier(w.Week)

	for i, t := range w.Tiles {
		// TrueRetrofitRate = ObservedRetrofitRate * (InstallerQuality / 100).
		w.Tiles[i].TrueRetrofitRate = region.ComputeTrueRetrofitRate(
			t.ObservedRetrofitRate, t.InstallerQuality)

		// FuelPoverty: recomputed from scratch each week so it responds
		// immediately to energy price changes and insulation improvements.
		oldFP := t.FuelPoverty
		w.Tiles[i].FuelPoverty = region.ComputeFuelPoverty(region.FuelPovertyInput{
			GasPrice:              w.EnergyMarket.GasPrice,
			ElectricityPrice:      w.EnergyMarket.ElectricityPrice,
			OilPrice:              w.EnergyMarket.OilPrice,
			HeatingType:           t.HeatingType,
			InsulationLevel:       t.InsulationLevel,
			LocalIncome:           t.LocalIncome,
			HeatingCapacity:       t.HeatingCapacity,
			TechMaturityHeatPumps: hpMaturity,
			SeasonalMultiplier:    seasonal,
		})

		// LocalPoliticalOpinion shifts from fuel poverty changes.
		fpDelta := w.Tiles[i].FuelPoverty - oldFP
		w.Tiles[i] = region.UpdateLocalPoliticalOpinion(
			w.Tiles[i], fpDelta, 0.0, w.ClimateState.Level)
	}
	return w
}

// ---------------------------------------------------------------------------
// Phase 9: Policy Resolution
// ---------------------------------------------------------------------------

func phasePolicyResolution(w WorldState) WorldState {
	// Tick under-review and active cards.
	for i, card := range w.PolicyCards {
		switch card.State {
		case policy.PolicyStateUnderReview:
			w.PolicyCards[i] = policy.TickUnderReview(card)
		case policy.PolicyStateActive:
			w.PolicyCards[i] = policy.TickActive(card)
		}
	}

	// Compute aggregate carbon reduction from active policies.
	// Each policy uses the installer capacity of regions relevant to its sector.
	// PROXY: per-policy region scoping is not in config; sector heuristics are used
	// (INDUSTRY -> industrial-tagged regions; all others -> all-region average).
	avgRetrofit := avgTrueRetrofitRate(w.Tiles)

	for _, card := range w.PolicyCards {
		if card.State != policy.PolicyStateActive {
			continue
		}
		techFrac := techFractionForPolicy(*card.Def, w.Tech)
		capRegion := region.Region{
			InstallerCapacity: sectorInstallerCapacity(
				w.Regions, w.Cfg.Regions, card.Def.WeeklyEffect.Sector),
		}
		delta := policy.ResolveWeeklyEffect(card, capRegion, techFrac, avgRetrofit)
		// DeltaMt convention: negative = reduction. WeeklyPolicyReductionMt stores
		// the net reduction as a positive number (reduction = removed carbon).
		w.WeeklyPolicyReductionMt -= delta.DeltaMt
		w.WeeklyPolicyBudgetCostGBP += delta.BudgetCostPerWeek
		// Direct LCR bonus from the policy config (e.g. wind planning reform gives
		// +0.06/week as a credibility signal beyond carbon tonnage alone).
		w.WeeklyPolicyLCRDelta += card.Def.LCRDeltaPerWeek
	}

	// Net carbon: base emissions minus policy reductions (reductions stored as
	// positive). Clamped to prevent negative values.
	w.WeeklyNetCarbonMt = mathutil.Clamp(
		w.BaseWeeklyMt-w.WeeklyPolicyReductionMt, 0, w.BaseWeeklyMt*2)
	return w
}

// ---------------------------------------------------------------------------
// Phase 10: Carbon Budget Accounting
// ---------------------------------------------------------------------------

func phaseCarbonBudgetAccounting(w WorldState) WorldState {
	w.Carbon = carbon.AccumulateWeekly(w.Carbon, w.WeeklyNetCarbonMt)

	// Year-end: check annual carbon budget limit against CCC targets.
	// w.Year is already incremented by phaseClockAdvance for the new year,
	// so subtract 1 to check the year that just completed.
	if w.Week%52 == 0 {
		_, w.Carbon = carbon.CheckAnnualBudget(w.Carbon, w.Year-1, w.Cfg.CarbonBudgets)
	}
	return w
}

// ---------------------------------------------------------------------------
// Phase 11: Economy and Tax Revenue Tick
// ---------------------------------------------------------------------------

func phaseEconomyTick(w WorldState) WorldState {
	// Aggregate fuel poverty as a drag on consumer spending -> Economy.
	fpDrag := avgFuelPoverty(w.Tiles) / 200.0 // normalise to ~0-0.5

	w.Economy = economy.TickEconomy(
		w.Economy,
		w.ClimateState.Severity*0.2,  // climate damage: up to 0.2 at peak severity
		fpDrag,
		0.0,                           // shock severity: events already applied directly
		0.0,                           // policy bonus: wired up in simulation tuning pass
		w.FossilDependency/500.0,      // fossil drag: ~0-0.2 range
	)

	// LCR tick: capture delta for downstream popularity chains.
	prevLCR := w.LCR.Value
	w.LCR = reputation.TickReputation(w.LCR, w.WeeklyPolicyReductionMt, w.WeeklyEventLCRDelta+w.WeeklyPolicyLCRDelta)
	lcrDelta := w.LCR.Value - prevLCR

	// Government popularity chain from LCR movement.
	w.GovernmentPopularity = mathutil.Clamp(
		w.GovernmentPopularity+reputation.ChainToGovtPopularity(lcrDelta), 0, 100)

	// Per-minister popularity from LCR chain and minister stats (cabinet ministers only).
	for i := range w.Stakeholders {
		s := w.Stakeholders[i]
		if !s.IsUnlocked || isTerminalState(s.State) {
			continue
		}
		if !isInCabinet(w.Government, s.ID) {
			continue
		}
		stats := government.ComputeMinisterStats(s, lcrDelta)
		minPopDelta := reputation.ChainToMinisterPopularity(lcrDelta) + stats.PopularityModifier
		w.Stakeholders[i].Popularity = mathutil.Clamp(s.Popularity+minPopDelta, 0, 100)
	}

	// Quarter-end: allocate budget using accumulated lobby effects, then clear them.
	if w.Week%13 == 0 {
		w.TickyPressureAcceptedThisQuarter = false
		w.LastTaxRevenue = economy.ComputeTaxRevenue(w.Economy, w.Quarter, w.Year)
		w.LastBudget = economy.AllocateBudget(
			w.LastTaxRevenue,
			baseDeptFractions(),
			cabinetBudgetWeights(w.Government, w.Stakeholders, lcrDelta),
			reputation.ChainToBudgetModifier(w.LCR.Value),
			w.Economy.LobbyEffects,
		)
		// Clear lobby effects after they have been applied to this quarter's budget.
		w.Economy = economy.ClearLobbyEffectsAtQuarterEnd(w.Economy)
	}
	return w
}

// ---------------------------------------------------------------------------
// Phase 12: Polling Check
// ---------------------------------------------------------------------------

func phasePollingCheck(w WorldState) WorldState {
	// Party vote share, government approval, and minister polls: one Bernoulli draw.
	if w.RNG.Float64() < pollWeeklyProb {
		snap := polling.TakePollSnapshot(w.Week, w.Tiles, w.Regions, w.RNG)

		// Government approval rating (sigma=3 noise on true hidden value).
		snap.GovernmentApprovalRating = mathutil.Clamp(
			w.GovernmentPopularity+w.RNG.NormFloat64()*3.0, 0, 100)
		w.GovernmentLastPollResult = snap.GovernmentApprovalRating

		// Per-minister popularity polls (sigma=5 noise on true hidden value).
		for _, s := range w.Stakeholders {
			if !s.IsUnlocked || isTerminalState(s.State) {
				continue
			}
			w.MinisterLastPollResults[s.ID] = mathutil.Clamp(
				s.Popularity+w.RNG.NormFloat64()*5.0, 0, 100)
		}

		// Swing computation against previous snapshot.
		if len(w.PollHistory) > 0 {
			swings := polling.SwingFromLast(snap, w.PollHistory[len(w.PollHistory)-1])
			for rid, swing := range swings {
				if rp, ok := snap.RegionPolls[rid]; ok {
					rp.Swing = swing
					snap.RegionPolls[rid] = rp
				}
			}
		}

		w.PollHistory = append(w.PollHistory, snap)
		if len(w.PollHistory) > maxPollHistory {
			w.PollHistory = w.PollHistory[len(w.PollHistory)-maxPollHistory:]
		}
	}

	// LCR poll: fires on a randomised 10-16 week interval.
	w.WeeksUntilLCRPoll--
	if w.WeeksUntilLCRPoll <= 0 {
		w.LCR = reputation.PollLCR(w.LCR, w.RNG)
		w.WeeksUntilLCRPoll = 10 + w.RNG.Intn(7)
	}
	return w
}

// ---------------------------------------------------------------------------
// Phase 13: Player Action Phase
// ---------------------------------------------------------------------------

func phasePlayerActions(w WorldState, actions []Action) WorldState {
	w.Player = player.StartWeekAPPool(w.Player)
	for _, a := range actions {
		w = applyAction(w, a)
	}
	return w
}

func applyAction(w WorldState, a Action) WorldState {
	switch a.Type {

	case player.ActionTypeSubmitPolicy:
		for i, card := range w.PolicyCards {
			if card.Def.ID != a.Target || card.State != policy.PolicyStateDraft {
				continue
			}
			// Tech unlock gate: cannot submit if gate not yet met.
			if !policy.IsUnlocked(card, w.Tech.Maturities) {
				break
			}
			cs, ok := player.SpendAP(w.Player, card.Def.APCost)
			if !ok {
				break
			}
			w.Player = cs
			w.PolicyCards[i] = policy.SubmitPolicy(card)
			w.Player = player.RecordAction(w.Player, player.ActionRecord{
				ActionType: a.Type, Week: w.Week, APCost: card.Def.APCost,
			})
			break
		}

	case player.ActionTypeCommissionReport:
		// a.Target = orgID; a.Detail = insight type string (config.InsightType).
		w = applyCommissionReport(w, a)

	case player.ActionTypeLobbyMinister:
		// a.Target = stakeholder ID; costs lobbyAPCost AP.
		for i, s := range w.Stakeholders {
			if s.ID != a.Target || !s.IsUnlocked || isTerminalState(s.State) {
				continue
			}
			cs, ok := player.SpendAP(w.Player, lobbyAPCost)
			if !ok {
				break
			}
			w.Player = cs
			w.Stakeholders[i] = stakeholder.TickRelationship(s, 5.0, 0)
			for _, dept := range roleToDepts(s.Role) {
				w.Economy = economy.AccumulateLobbyEffect(w.Economy, dept, lobbyBudgetEffect)
			}
			w.Player = player.RecordAction(w.Player, player.ActionRecord{
				ActionType: a.Type, Week: w.Week, APCost: lobbyAPCost,
			})
			break
		}

	case player.ActionTypeRespondTickyPressure:
		// a.Detail = "ACCEPT" | "DECLINE" | "NEGOTIATE"
		// Only executes if a Ticky pressure event is pending.
		w = applyTickyPressureResponse(w, a)

	case player.ActionTypeRespondRiskyTicky:
		w = applyRiskyTickyResponse(w, a)

	case player.ActionTypeRespondTrickyTicky:
		w = applyTrickyTickyResponse(w, a)

	case player.ActionTypeDamageTickyReputation:
		w = applyDamageTickyReputation(w, a)

	case player.ActionTypeGreatSneezeLobby:
		// Free-AP emergency lobby; only available while Great Sneeze is active.
		if w.GreatSneezeActive {
			// Reuse the normal lobby logic but deduct 0 AP.
			lobbyAction := Action{
				Type:   player.ActionTypeLobbyMinister,
				Target: a.Target,
			}
			// Apply lobby effect without AP deduction.
			w = applyLobbyWithCost(w, lobbyAction, 0)
		}

	case player.ActionTypeShockResponse:
		// a.Target = EventDefID; a.Detail = ShockResponseOption string.
		w = applyShockResponse(w, a)

	case player.ActionTypeHireStaff:
		w.Player = player.HireStaff(w.Player, staffFromAction(a, w.Week))

	case player.ActionTypeFireStaff:
		w.Player = player.FireStaff(w.Player, a.Target)
	}
	return w
}

// applyCommissionReport handles ActionTypeCommissionReport.
func applyCommissionReport(w WorldState, a Action) WorldState {
	// Find org def.
	var orgDef config.OrgDefinition
	found := false
	for _, d := range w.Cfg.Organisations {
		if d.ID == a.Target {
			orgDef = d
			found = true
			break
		}
	}
	if !found {
		return w
	}

	// Find org state; check availability.
	for _, os := range w.OrgStates {
		if os.OrgID != a.Target {
			continue
		}
		if os.CoolingOffUntil > w.Week {
			return w // still cooling off after a failed commission
		}
		// Check Murican org availability.
		tickyPresent := false
		for _, sid := range w.Government.CabinetByRole {
			if sid == tickyStakeholderID {
				tickyPresent = true
				break
			}
		}
		if !evidence.MuracanOrgAvailable(orgDef, os, tickyPresent, w.TickyPressureAcceptedThisQuarter) {
			return w
		}
		comm := evidence.CreateCommission(
			orgDef,
			config.InsightType(a.Detail),
			a.Detail,     // scope
			a.Detail,     // topicKey (same as scope for now)
			w.Week,
			w.RNG,
		)
		w.Commissions = append(w.Commissions, comm)
		return w
	}
	return w
}

// applyShockResponse handles ActionTypeShockResponse.
func applyShockResponse(w WorldState, a Action) WorldState {
	for j, psr := range w.PendingShockResponses {
		if psr.EventDefID != a.Target {
			continue
		}
		option := climate.ShockResponseOption(a.Detail)
		backfireProb := climate.BackfireProbability(w.LCR.Value, w.Player.Reputation)
		card := climate.ShockResponseCard{
			ID:                  psr.EventDefID + "_response",
			EventDefID:          psr.EventDefID,
			BackfireProbability: backfireProb,
		}
		outcome := climate.ShockResponseOutcome(card, option, w.RNG)
		w.LCR.Value = mathutil.Clamp(w.LCR.Value+outcome.LCRDelta, 0, 100)
		w.GovernmentPopularity = mathutil.Clamp(w.GovernmentPopularity+outcome.PopularityDelta, 0, 100)
		// Remove handled shock response.
		w.PendingShockResponses = append(w.PendingShockResponses[:j], w.PendingShockResponses[j+1:]...)
		return w
	}
	return w
}

// ---------------------------------------------------------------------------
// Phase 4b: Ticky Pressure Tick
// ---------------------------------------------------------------------------

// phaseTickyPressureTick manages TD Tennison's recurring pressure mechanic.
// When Ticky is in cabinet, the countdown decrements each week. On reaching
// zero, PendingTickyPressure is set and the countdown resets to 6-10 weeks.
// The player must respond via ActionTypeRespondTickyPressure before the next
// pressure fires; an unanswered pressure is silently replaced.
func phaseTickyPressureTick(w WorldState) WorldState {
	tickyInCabinet := false
	for _, sid := range w.Government.CabinetByRole {
		if sid == tickyStakeholderID {
			tickyInCabinet = true
			break
		}
	}
	if !tickyInCabinet {
		return w
	}
	if w.TickyCountdown > 0 {
		w.TickyCountdown--
		return w
	}
	// Countdown reached zero: fire pressure event.
	w.PendingTickyPressure = true
	w.TickyCountdown = tickyPressureMinWeeks + w.RNG.Intn(tickyPressureMaxVariance)
	return w
}

// applyTickyPressureResponse resolves the player's response to a Ticky pressure event.
//
//   - ACCEPT:    +8 relationship with Ticky; unlocks murican_growth_alliance this quarter.
//   - DECLINE:   -5 relationship with Ticky; no Murican org access.
//   - NEGOTIATE: -2 relationship with Ticky; unlocks murican_growth_alliance this quarter.
//
// Has no effect if no pressure event is pending.
func applyTickyPressureResponse(w WorldState, a Action) WorldState {
	if !w.PendingTickyPressure {
		return w
	}
	var relDelta float64
	switch a.Detail {
	case "ACCEPT":
		relDelta = tickyRelationshipAccept
		w = unlockTickyOrg(w)
		w.TickyPressureAcceptedThisQuarter = true
	case "DECLINE":
		relDelta = tickyRelationshipDecline
	case "NEGOTIATE":
		relDelta = tickyRelationshipNegotiate
		w = unlockTickyOrg(w)
		w.TickyPressureAcceptedThisQuarter = true
	default:
		return w
	}
	for i, s := range w.Stakeholders {
		if s.ID == tickyStakeholderID {
			w.Stakeholders[i] = stakeholder.TickRelationship(s, relDelta, 0)
			break
		}
	}
	w.PendingTickyPressure = false
	w.Player = player.RecordAction(w.Player, player.ActionRecord{
		ActionType: a.Type, Week: w.Week, APCost: 0,
	})
	return w
}

// unlockTickyOrg marks the Tier 1 Murican org as unlocked for this quarter.
func unlockTickyOrg(w WorldState) WorldState {
	for i, os := range w.OrgStates {
		if os.OrgID == tickyTier1OrgID {
			w.OrgStates[i] = evidence.UnlockMuricanOrg(os)
			return w
		}
	}
	return w
}

// applyLobbyWithCost applies the lobby action effect, deducting apCost AP from
// the player. Pass apCost=0 for free actions (e.g. Great Sneeze emergency lobby).
func applyLobbyWithCost(w WorldState, a Action, apCost int) WorldState {
	for i, s := range w.Stakeholders {
		if s.ID != a.Target || !s.IsUnlocked || isTerminalState(s.State) {
			continue
		}
		if apCost > 0 {
			cs, ok := player.SpendAP(w.Player, apCost)
			if !ok {
				break
			}
			w.Player = cs
		}
		w.Stakeholders[i] = stakeholder.TickRelationship(s, 5.0, 0)
		for _, dept := range roleToDepts(s.Role) {
			w.Economy = economy.AccumulateLobbyEffect(w.Economy, dept, lobbyBudgetEffect)
		}
		w.Player = player.RecordAction(w.Player, player.ActionRecord{
			ActionType: a.Type, Week: w.Week, APCost: apCost,
		})
		break
	}
	return w
}

// applyRiskyTickyResponse resolves the Risky Ticky endorsement prompt.
//
//   - ENDORSE: +15 relationship with Ticky; -8 to every current cabinet member.
//   - DECLINE: -3 relationship with Ticky.
func applyRiskyTickyResponse(w WorldState, a Action) WorldState {
	if !w.PendingRiskyTicky {
		return w
	}
	switch a.Detail {
	case "ENDORSE":
		// Boost Ticky relationship.
		for i, s := range w.Stakeholders {
			if s.ID == tickyStakeholderID {
				w.Stakeholders[i] = stakeholder.TickRelationship(s, riskyTickyEndorseRelBoost, 0)
				break
			}
		}
		// Damage relationship with every current cabinet minister.
		cabinetIDs := make(map[string]bool, len(w.Government.CabinetByRole))
		for _, sid := range w.Government.CabinetByRole {
			cabinetIDs[sid] = true
		}
		for i, s := range w.Stakeholders {
			if cabinetIDs[s.ID] && s.ID != tickyStakeholderID {
				w.Stakeholders[i] = stakeholder.TickRelationship(s, riskyTickyEndorseCabinetPenalty, 0)
			}
		}
	case "DECLINE":
		for i, s := range w.Stakeholders {
			if s.ID == tickyStakeholderID {
				w.Stakeholders[i] = stakeholder.TickRelationship(s, riskyTickyDeclinePenalty, 0)
				break
			}
		}
	}
	w.PendingRiskyTicky = false
	w.Player = player.RecordAction(w.Player, player.ActionRecord{
		ActionType: a.Type, Week: w.Week, APCost: 0,
	})
	return w
}

// applyTrickyTickyResponse resolves the Tricky Ticky Murican contract offer.
//
//   - ACCEPT: +20 GBP budget boost to DeptPower; -5 relationship with each far-right
//     cabinet member (they resent being bypassed on the deal).
//   - DECLINE: -3 relationship with Ticky.
func applyTrickyTickyResponse(w WorldState, a Action) WorldState {
	if !w.PendingTrickyTicky {
		return w
	}
	switch a.Detail {
	case "ACCEPT":
		// Budget boost to Power department.
		if w.LastBudget.Departments != nil {
			w.LastBudget.Departments[government.DeptPower] += trickyTickyAcceptBudgetBoost
		}
		// Relationship penalty for every far-right cabinet member.
		cabinetIDs := make(map[string]bool, len(w.Government.CabinetByRole))
		for _, sid := range w.Government.CabinetByRole {
			cabinetIDs[sid] = true
		}
		for i, s := range w.Stakeholders {
			if cabinetIDs[s.ID] && s.Party == config.PartyFarRight {
				w.Stakeholders[i] = stakeholder.TickRelationship(s, trickyTickyAcceptCabinetPenalty, 0)
			}
		}
	case "DECLINE":
		for i, s := range w.Stakeholders {
			if s.ID == tickyStakeholderID {
				w.Stakeholders[i] = stakeholder.TickRelationship(s, trickyTickyDeclinePenalty, 0)
				break
			}
		}
	}
	w.PendingTrickyTicky = false
	w.Player = player.RecordAction(w.Player, player.ActionRecord{
		ActionType: a.Type, Week: w.Week, APCost: 0,
	})
	return w
}

// applyDamageTickyReputation is the player's counter-move against Angry Ticky.
// Costs angryTickyDamageAPCost AP. Requires that at least one unlocked non-Ticky
// far-right stakeholder has RelationshipScore >= 40 (tacit political cover).
// If Ticky's relationship drops below wimpyTickyRelThreshold after this action,
// Wimpy Ticky triggers: Angry Ticky is deactivated.
func applyDamageTickyReputation(w WorldState, a Action) WorldState {
	if !w.AngryTickyActive {
		return w
	}
	if w.Player.APRemaining < angryTickyDamageAPCost {
		return w
	}
	// Check political cover: need at least one non-Ticky FarRight with rel >= 40.
	hasCover := false
	for _, s := range w.Stakeholders {
		if s.IsUnlocked && s.ID != tickyStakeholderID &&
			s.Party == config.PartyFarRight && s.RelationshipScore >= 40 {
			hasCover = true
			break
		}
	}
	if !hasCover {
		return w
	}
	// Deduct AP.
	w.Player.APRemaining -= angryTickyDamageAPCost
	// Damage Ticky's relationship.
	var tickyRel float64
	for i, s := range w.Stakeholders {
		if s.ID == tickyStakeholderID {
			w.Stakeholders[i] = stakeholder.TickRelationship(s, angryTickyDamageRelDelta, 0)
			tickyRel = w.Stakeholders[i].RelationshipScore
			break
		}
	}
	// Splash damage: other FarRight stakeholders take minor relationship penalty.
	for i, s := range w.Stakeholders {
		if s.ID != tickyStakeholderID && s.IsUnlocked && s.Party == config.PartyFarRight {
			w.Stakeholders[i] = stakeholder.TickRelationship(s, angryTickyDamageCabinetSplash, 0)
		}
	}
	// Wimpy Ticky trigger.
	if tickyRel < wimpyTickyRelThreshold {
		w.AngryTickyActive = false
		w.AngryTickyWimpy = true
		entry := event.EventEntry{
			DefID: "wimpy_ticky",
			Name:  "Ticky Backs Down",
			Week:  w.Week,
		}
		w.EventLog = event.AppendEventLog(w.EventLog, entry)
	}
	w.Player = player.RecordAction(w.Player, player.ActionRecord{
		ActionType: a.Type, Week: w.Week, APCost: angryTickyDamageAPCost,
	})
	return w
}

// ---------------------------------------------------------------------------
// Phase 3b: Time-gated events
// ---------------------------------------------------------------------------

// phaseTimeGatedEvents fires EventDef entries whose TriggerAtYear matches the
// current game year. Each such event fires exactly once; FiredOnceEvents prevents
// repeat fires. The Great Sneeze is the primary user of this mechanism.
func phaseTimeGatedEvents(w WorldState) WorldState {
	if w.FiredOnceEvents == nil {
		w.FiredOnceEvents = make(map[string]bool)
	}
	// Only check on the first week of a new year (Week % 52 == 1).
	if w.Week%52 != 1 {
		return w
	}
	for _, def := range w.Cfg.Events {
		if def.TriggerAtYear == 0 || def.TriggerAtYear != w.Year {
			continue
		}
		if w.FiredOnceEvents[def.ID] {
			continue
		}
		// Fire the event.
		w.FiredOnceEvents[def.ID] = true
		resolved := event.ResolveEffect(def.BaseEffects, w.Cfg.Regions, w.Stakeholders, companyStateSlice(w.Industry), companyDefMap(w.Cfg))
		w.EnergyMarket = energy.ApplyShock(w.EnergyMarket, def.BaseEffects)
		w.GovernmentPopularity = mathutil.Clamp(w.GovernmentPopularity+resolved.GovtPopularityDelta, 0, 100)
		w.Economy.Value = mathutil.Clamp(w.Economy.Value+resolved.EconomyDelta, 0, 100)
		w.WeeklyEventLCRDelta += resolved.LCRDelta
		w.LCR.Value = mathutil.Clamp(w.LCR.Value+resolved.LCRDelta, 0, 100)
		for i, r := range w.Regions {
			if d, ok := resolved.RegionDeltas[r.ID]; ok {
				w.Regions[i].InstallerCapacity = mathutil.Clamp(r.InstallerCapacity+d.InstallerCapacityDelta, 0, 100)
				w.Regions[i].SkillsNetwork = mathutil.Clamp(r.SkillsNetwork+d.SkillsNetworkDelta, 0, 100)
			}
		}
		for i, t := range w.Tiles {
			if d, ok := resolved.TileDeltas[t.RegionID]; ok {
				w.Tiles[i].FuelPoverty = mathutil.Clamp(t.FuelPoverty+d.FuelPovertyDelta, 0, 100)
				w.Tiles[i].InsulationLevel = mathutil.Clamp(t.InsulationLevel-d.InsulationDamage, 0, 100)
			}
		}
		for i, s := range w.Stakeholders {
			if d, ok := resolved.StakeholderDeltas[s.ID]; ok {
				w.Stakeholders[i] = stakeholder.TickRelationship(s, d.RelDelta, float64(d.PressureDelta))
			}
		}
		w = applyCompanyDeltas(w, resolved.CompanyDeltas)
		// Great Sneeze special handling.
		if def.ID == "great_sneeze" && !w.GreatSneezeFired {
			w.GreatSneezeFired = true
			w.GreatSneezeActive = true
			// Great Sneeze lasts until end of greatSneezeEndYear.
			// Week offset: (greatSneezeEndYear - 2010 + 1) * 52 = roughly week 624 from start.
			w.GreatSneezeWeekEnd = (greatSneezeEndYear-2010+1)*52 + 1
			// Emergency spending: boost LastBudget for all departments.
			for dept := range w.LastBudget.Departments {
				w.LastBudget.Departments[dept] += greatSneezeEmergencyBudgetBoostGBP
			}
		}
		entry := event.EventEntry{DefID: def.ID, Name: def.Name, Week: w.Week}
		w.EventLog = event.AppendEventLog(w.EventLog, entry)
	}
	return w
}

// ---------------------------------------------------------------------------
// Phase 3c: Decaying shock tick
// ---------------------------------------------------------------------------

// phaseDecayingShockTick applies all active decaying market shocks this week,
// then multiplies their weekly deltas by DecayRate and decrements WeeksRemaining.
// Shocks that have expired (WeeksRemaining <= 0) are removed.
func phaseDecayingShockTick(w WorldState) WorldState {
	if len(w.ActiveDecayingShocks) == 0 {
		return w
	}
	active := w.ActiveDecayingShocks[:0:len(w.ActiveDecayingShocks)]
	for _, shock := range w.ActiveDecayingShocks {
		if shock.WeeksRemaining <= 0 {
			continue
		}
		// Apply this week's price deltas.
		effect := config.EventEffect{
			GasPriceDeltaPct:         shock.GasPctThisWeek,
			OilPriceDeltaPct:         shock.OilPctThisWeek,
			ElectricityPriceDeltaPct: shock.ElecPctThisWeek,
		}
		w.EnergyMarket = energy.ApplyShock(w.EnergyMarket, effect)
		// Decay and decrement.
		shock.GasPctThisWeek *= shock.DecayRate
		shock.OilPctThisWeek *= shock.DecayRate
		shock.ElecPctThisWeek *= shock.DecayRate
		shock.WeeksRemaining--
		if shock.WeeksRemaining > 0 {
			active = append(active, shock)
		}
	}
	w.ActiveDecayingShocks = active
	return w
}

// ---------------------------------------------------------------------------
// Phase 11b: Great Sneeze tick
// ---------------------------------------------------------------------------

// phaseGreatSneezeTick applies ongoing weekly effects while the Great Sneeze
// is active (weekly popularity drain reflecting emergency chaos). It also checks
// whether the sneeze has run its course and deactivates it.
func phaseGreatSneezeTick(w WorldState) WorldState {
	if !w.GreatSneezeActive {
		return w
	}
	if w.Week >= w.GreatSneezeWeekEnd {
		w.GreatSneezeActive = false
		return w
	}
	w.GovernmentPopularity = mathutil.Clamp(w.GovernmentPopularity+greatSneezeWeeklyPopPenalty, 0, 100)
	return w
}

// phaseExtendedTickyMechanics handles Risky Ticky (endorsement prompt when
// Left/FarLeft governs) and Tricky Ticky (Murican contract offer when FarRight
// governs). These are independent of the standard Ticky pressure countdown.
func phaseExtendedTickyMechanics(w WorldState) WorldState {
	// Risky Ticky: fires probabilistically when centrist-left or far-left governs,
	// after the first year, and only if a Risky prompt is not already pending.
	if !w.PendingRiskyTicky && w.Week >= riskyTickyMinWeek {
		party := w.Government.RulingParty
		if (party == config.PartyLeft || party == config.PartyFarLeft) &&
			w.RNG.Float64() < riskyTickyWeeklyProb {
			w.PendingRiskyTicky = true
		}
	}
	// Tricky Ticky: fires probabilistically when far-right governs,
	// only if a Tricky prompt is not already pending.
	if !w.PendingTrickyTicky {
		if w.Government.RulingParty == config.PartyFarRight &&
			w.RNG.Float64() < trickyTickyWeeklyProb {
			w.PendingTrickyTicky = true
		}
	}
	return w
}

// activePolicyRDBonuses sums the RDBonus values from all ACTIVE policy cards.
// Returns a map from Technology to total weekly acceleration bonus.
func activePolicyRDBonuses(cards []policy.PolicyCard) map[config.Technology]float64 {
	out := make(map[config.Technology]float64)
	for _, card := range cards {
		if card.State != policy.PolicyStateActive {
			continue
		}
		for tech, bonus := range card.Def.RDBonus {
			out[tech] += bonus
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Phase 14: Minister Health Check
// ---------------------------------------------------------------------------

func phaseMinisterHealthCheck(w WorldState) WorldState {
	partyShares := nationalPartyShares(w)

	// Sacking threshold shifts down when GovernmentPopularity is high (PM shielded).
	sackingThreshold := ministerSackingThreshold
	if w.GovernmentPopularity > 60.0 {
		sackingThreshold = ministerSackingThresholdHighPop
	}

	for i := range w.Stakeholders {
		s := w.Stakeholders[i]
		if !s.IsUnlocked {
			continue
		}
		// Advance special mechanics (Ticky pressure counter, Dizzy surge, etc.).
		w.Stakeholders[i] = stakeholder.TickSpecialMechanic(s, w.ClimateState.Level, w.RNG)
		// Update influence from current polling data.
		w.Stakeholders[i] = stakeholder.ComputeInfluence(
			w.Stakeholders[i], partyShares, stakeholder.DefaultRoleWeights)
		// Passive relationship decay every week (no player action, no event).
		w.Stakeholders[i] = stakeholder.TickRelationship(w.Stakeholders[i], 0, 0)

		updated := w.Stakeholders[i]

		// Ideology conflict score: decay each week for non-terminal ministers.
		if !isTerminalState(updated.State) && updated.IdeologyConflictScore > 0 {
			w.Stakeholders[i].IdeologyConflictScore = mathutil.Clamp(
				updated.IdeologyConflictScore*(1.0-ideologyConflictDecayRate), 0, 100)
			updated = w.Stakeholders[i]
		}

		// Consultancy affinity: governing ministers with org ties boost those orgs' relationships.
		if isInCabinet(w.Government, updated.ID) && len(updated.ConsultancyAffinity) > 0 {
			for _, orgID := range updated.ConsultancyAffinity {
				for j, os := range w.OrgStates {
					if os.OrgID == orgID {
						w.OrgStates[j].RelationshipScore = mathutil.Clamp(
							w.OrgStates[j].RelationshipScore+consultancyAffinityBonusPerWeek, 0, 100)
						break
					}
				}
			}
		}

		// Consultancy aversion: averse governing ministers lose relationship with player
		// for each active Consultancy-type commission.
		if updated.ConsultancyAversion && isInCabinet(w.Government, updated.ID) {
			activeConsOrgs := activeConsultancyOrgIDs(w.Commissions, w.Cfg.Organisations)
			for range activeConsOrgs {
				w.Stakeholders[i] = stakeholder.TickRelationship(w.Stakeholders[i], 0, -consultancyAversionPenaltyPerWeek)
			}
			updated = w.Stakeholders[i]
		}

		// Ideology conflict resignation: cabinet ministers only.
		if !isTerminalState(updated.State) &&
			updated.State != stakeholder.MinisterStateAppointed &&
			isInCabinet(w.Government, updated.ID) &&
			updated.IdeologyConflictScore >= ideologyConflictResignThreshold {
			w.Stakeholders[i].State = stakeholder.MinisterStateResigned
			for role, sid := range w.Government.CabinetByRole {
				if sid == updated.ID {
					w.Government = government.RemoveMinister(w.Government, role)
					break
				}
			}
			continue
		}

		// Grace period: newly appointed ministers are not held accountable for
		// popularity for ministerGraceWeeks weeks after appointment.
		if updated.GraceWeeksRemaining > 0 {
			w.Stakeholders[i].GraceWeeksRemaining--
			continue
		}

		// Popularity-based state transitions for cabinet ministers only.
		if isTerminalState(updated.State) || updated.State == stakeholder.MinisterStateAppointed {
			continue
		}
		if !isInCabinet(w.Government, updated.ID) {
			continue
		}
		switch updated.State {
		case stakeholder.MinisterStateActive:
			if updated.Popularity < sackingThreshold {
				w.Stakeholders[i].State = stakeholder.MinisterStateUnderPressure
			}
		case stakeholder.MinisterStateUnderPressure:
			if updated.Popularity >= sackingThreshold {
				// Recovery: popularity has recovered above threshold.
				w.Stakeholders[i].State = stakeholder.MinisterStateActive
				w.Stakeholders[i].WeeksUnderPressure = 0
			}
		}
	}

	// Angry Ticky detection: if Ticky's relationship drops below the threshold
	// and AngryTickyActive has not yet been triggered, activate it.
	if !w.AngryTickyActive && !w.AngryTickyWimpy {
		for _, s := range w.Stakeholders {
			if s.ID == tickyStakeholderID && s.IsUnlocked && s.RelationshipScore < angryTickyRelThreshold {
				w.AngryTickyActive = true
				// Log the event so the player sees it.
				entry := event.EventEntry{
					DefID: "angry_ticky",
					Name:  "TD Tennison Turns Hostile",
					Week:  w.Week,
				}
				w.EventLog = event.AppendEventLog(w.EventLog, entry)
				break
			}
		}
	}

	return w
}

// ---------------------------------------------------------------------------
// Phase 15: Minister Transitions
// ---------------------------------------------------------------------------

func phaseMinisterTransitions(w WorldState) WorldState {
	for i := range w.Stakeholders {
		s := w.Stakeholders[i]
		if !s.IsUnlocked {
			continue
		}
		updated := w.Stakeholders[i]
		switch s.State {
		case stakeholder.MinisterStateAppointed:
			// APPOINTED -> ACTIVE: set grace period so the minister is not held
			// accountable for popularity for ministerGraceWeeks weeks.
			w.Stakeholders[i].GraceWeeksRemaining = ministerGraceWeeks
			w.Stakeholders[i].State = stakeholder.MinisterStateActive

		case stakeholder.MinisterStateUnderPressure:
			// Increment consecutive pressure weeks; sack after threshold.
			w.Stakeholders[i].WeeksUnderPressure++
			if w.Stakeholders[i].WeeksUnderPressure >= ministerSackingWeeks {
				w.Stakeholders[i].State = stakeholder.MinisterStateSacked
				w.Stakeholders[i].WeeksUnderPressure = 0
				// Vacate cabinet role.
				for role, sid := range w.Government.CabinetByRole {
					if sid == s.ID {
						w.Government = government.RemoveMinister(w.Government, role)
						break
					}
				}
			}

		case stakeholder.MinisterStateSacked:
			// SACKED is a transient state: move to BACKBENCH.
			w.Stakeholders[i].State = stakeholder.MinisterStateBackbench

		case stakeholder.MinisterStateResigned:
			// RESIGNED is a transient state: move to BACKBENCH.
			w.Stakeholders[i].State = stakeholder.MinisterStateBackbench

		case stakeholder.MinisterStateBackbench:
			// Low-popularity backbenchers eventually retire.
			// Reuse WeeksUnderPressure as a backbench accountability counter.
			if updated.Popularity < 15.0 {
				w.Stakeholders[i].WeeksUnderPressure++
				if w.Stakeholders[i].WeeksUnderPressure >= 12 {
					w.Stakeholders[i].State = stakeholder.MinisterStateDeparted
					w.Stakeholders[i].WeeksUnderPressure = 0
				}
			} else {
				w.Stakeholders[i].WeeksUnderPressure = 0
			}
		}
	}
	return w
}

// evaluateApprovalWithAngryTicKyMalus wraps policy.EvaluateApproval to apply
// the Angry Ticky policy malus when AngryTickyActive is true. The malus adds
// angryTickyPolicyMalus points to MinRelationshipScore on every approval step,
// making all UNDER_REVIEW policies harder to push through.
func evaluateApprovalWithAngryTicKyMalus(w WorldState, card policy.PolicyCard, stakeholders []stakeholder.Stakeholder) policy.PolicyCard {
	if !w.AngryTickyActive || card.State != policy.PolicyStateUnderReview {
		return policy.EvaluateApproval(card, stakeholders)
	}
	// Create a temporary copy of the def with elevated MinRelationshipScore.
	defCopy := *card.Def
	steps := make([]config.ApprovalRequirement, len(defCopy.ApprovalSteps))
	for i, step := range defCopy.ApprovalSteps {
		step.MinRelationshipScore += angryTickyPolicyMalus
		steps[i] = step
	}
	defCopy.ApprovalSteps = steps
	cardCopy := card
	cardCopy.Def = &defCopy
	return policy.EvaluateApproval(cardCopy, stakeholders)
}

// ---------------------------------------------------------------------------
// Phase 16: Consequence Resolution
// ---------------------------------------------------------------------------

func phaseConsequenceResolution(w WorldState) WorldState {
	// Evaluate approval steps for UNDER_REVIEW policies.
	// Runs before player actions so submissions this tick are not evaluated until next tick.
	// Only cabinet ministers (governing party) participate in policy approval.
	active := cabinetStakeholders(w.Government, w.Stakeholders)

	// Build role -> stakeholder map for ideology conflict tracking.
	byRole := make(map[config.Role]stakeholder.Stakeholder)
	for _, s := range active {
		if _, exists := byRole[s.Role]; !exists {
			byRole[s.Role] = s
		}
	}

	for i, card := range w.PolicyCards {
		if card.State != policy.PolicyStateUnderReview {
			continue
		}
		prevSteps := card.StepsCleared
		w.PolicyCards[i] = evaluateApprovalWithAngryTicKyMalus(w, card, active)
		// If all approval steps cleared, activate immediately.
		if w.PolicyCards[i].State == policy.PolicyStateApproved {
			w.PolicyCards[i] = policy.ActivatePolicy(w.PolicyCards[i])
		}
		// For each newly cleared step, accumulate ideology conflict on the approving minister.
		if w.PolicyCards[i].StepsCleared > prevSteps {
			for step := prevSteps; step < w.PolicyCards[i].StepsCleared && step < len(card.Def.ApprovalSteps); step++ {
				req := card.Def.ApprovalSteps[step]
				s, ok := byRole[req.Role]
				if !ok {
					continue
				}
				conflict := policy.IdeologyConflict(*card.Def, s) * ideologyConflictWeight * significanceMultiplier(card.Def.Significance)
				for j := range w.Stakeholders {
					if w.Stakeholders[j].ID == s.ID {
						w.Stakeholders[j].IdeologyConflictScore = mathutil.Clamp(
							w.Stakeholders[j].IdeologyConflictScore+conflict, 0, 100)
						break
					}
				}
			}
		}
	}
	return w
}

// ---------------------------------------------------------------------------
// Phase 17: Consultancy Delivery Check
// ---------------------------------------------------------------------------

func phaseConsultancyDelivery(w WorldState) WorldState {
	defs := orgDefMap(w.Cfg)
	updated, delivered := evidence.TickDelivery(w.Commissions, defs, w.Week, w.RNG)
	w.Commissions = updated

	// Track which org IDs received a relationship event this week (to avoid double-decay).
	eventedOrgs := make(map[string]bool, len(delivered))

	for _, c := range delivered {
		orgDef, ok := defs[c.OrgID]
		if !ok {
			continue
		}

		// Update org relationship state.
		for j, os := range w.OrgStates {
			if os.OrgID != c.OrgID {
				continue
			}
			ev := evidence.RelationshipEventDelivered
			if c.Failed {
				ev = evidence.RelationshipEventFailed
			}
			w.OrgStates[j] = evidence.UpdateOrgRelationship(os, ev, w.Week)
			eventedOrgs[c.OrgID] = true
		}

		if c.Failed {
			continue
		}

		// Generate the insight report using the true world-state value for this topic.
		rawVal := rawValueForInsight(c.InsightType, w)
		report := evidence.GenerateReport(c, orgDef, string(c.InsightType), rawVal, 0.0, w.RNG)
		w.Reports = append(w.Reports, report)
		if len(w.Reports) > maxInsightReports {
			w.Reports = w.Reports[len(w.Reports)-maxInsightReports:]
		}
	}

	// Apply natural relationship decay for orgs that had no event this week.
	for j, os := range w.OrgStates {
		if !eventedOrgs[os.OrgID] {
			w.OrgStates[j] = evidence.UpdateOrgRelationship(os, "", w.Week)
		}
	}
	return w
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func companyStateSlice(ind industry.IndustryState) []industry.CompanyState {
	out := make([]industry.CompanyState, 0, len(ind.Companies))
	for _, cs := range ind.Companies {
		out = append(out, cs)
	}
	return out
}

func companyDefMap(cfg *config.Config) map[string]config.CompanyDef {
	m := make(map[string]config.CompanyDef, len(cfg.Companies))
	for _, d := range cfg.Companies {
		m[d.ID] = d
	}
	return m
}

func orgDefMap(cfg *config.Config) map[string]config.OrgDefinition {
	m := make(map[string]config.OrgDefinition, len(cfg.Organisations))
	for _, d := range cfg.Organisations {
		m[d.ID] = d
	}
	return m
}

func baseDeptFractions() map[string]float64 {
	return map[string]float64{
		government.DeptPower:     0.25,
		government.DeptTransport: 0.20,
		government.DeptBuildings: 0.25,
		government.DeptIndustry:  0.20,
		government.DeptCross:     0.10,
	}
}

func neutralMinisterWeights() map[string]float64 {
	return map[string]float64{
		government.DeptPower:     1.0,
		government.DeptTransport: 1.0,
		government.DeptBuildings: 1.0,
		government.DeptIndustry:  1.0,
		government.DeptCross:     1.0,
	}
}

// seasonalMultiplier returns a heating demand multiplier that peaks in January
// (week 0, ~1.30) and troughs in July (week 26, ~0.70).
func seasonalMultiplier(week int) float64 {
	weekInYear := week % 52
	angle := 2.0 * math.Pi * float64(weekInYear) / 52.0
	return 1.0 + 0.3*math.Cos(angle)
}

func avgFuelPoverty(tiles []region.Tile) float64 {
	if len(tiles) == 0 {
		return 0
	}
	sum := 0.0
	for _, t := range tiles {
		sum += t.FuelPoverty
	}
	return sum / float64(len(tiles))
}

func avgTrueRetrofitRate(tiles []region.Tile) float64 {
	if len(tiles) == 0 {
		return 0
	}
	sum := 0.0
	for _, t := range tiles {
		sum += t.TrueRetrofitRate
	}
	return sum / float64(len(tiles))
}

// techFractionForPolicy returns the 0-1 tech maturity fraction for a policy's
// unlock gate, or 1.0 if the policy has no tech dependency.
func techFractionForPolicy(def config.PolicyCardDef, tech technology.TechTracker) float64 {
	if def.TechUnlockGate == "" {
		return 1.0
	}
	return mathutil.Clamp(tech.Maturity(def.TechUnlockGate)/100.0, 0, 1)
}

func unlockedStakeholders(stakeholders []stakeholder.Stakeholder) []stakeholder.Stakeholder {
	out := make([]stakeholder.Stakeholder, 0)
	for _, s := range stakeholders {
		if s.IsUnlocked {
			out = append(out, s)
		}
	}
	return out
}

// cabinetStakeholders returns only the stakeholders currently assigned to the
// governing cabinet (via CabinetByRole). Used for policy approval evaluation so
// that opposition-party ministers do not participate in the approval pipeline.
func cabinetStakeholders(gov government.GovernmentState, stakeholders []stakeholder.Stakeholder) []stakeholder.Stakeholder {
	cabinetIDs := make(map[string]bool, len(gov.CabinetByRole))
	for _, sid := range gov.CabinetByRole {
		cabinetIDs[sid] = true
	}
	out := make([]stakeholder.Stakeholder, 0, len(cabinetIDs))
	for _, s := range stakeholders {
		if cabinetIDs[s.ID] && s.IsUnlocked && !isTerminalState(s.State) {
			out = append(out, s)
		}
	}
	return out
}

// nationalPartyShares returns the most recent national poll shares, or neutral
// equal shares if no polls have been taken yet.
func nationalPartyShares(w WorldState) map[config.Party]float64 {
	if len(w.PollHistory) == 0 {
		return map[config.Party]float64{
			config.PartyLeft:     25.0,
			config.PartyRight:    25.0,
			config.PartyFarLeft:  25.0,
			config.PartyFarRight: 25.0,
		}
	}
	return w.PollHistory[len(w.PollHistory)-1].NationalPolls
}

func staffFromAction(a Action, week int) player.StaffMember {
	role := player.StaffRole(a.Target)
	apBonus := 1
	if role == player.StaffRoleChiefOfStaff {
		apBonus = 2
	}
	return player.StaffMember{
		ID:        a.Detail,
		Role:      role,
		APBonus:   apBonus,
		WeekHired: week,
	}
}

// isInCabinet returns true if the given stakeholder ID occupies any cabinet role.
func isInCabinet(g government.GovernmentState, stakeholderID string) bool {
	for _, sid := range g.CabinetByRole {
		if sid == stakeholderID {
			return true
		}
	}
	return false
}

// avgInstallerCapacity returns the mean InstallerCapacity across all regions.
// Returns 0 if no regions exist.
func avgInstallerCapacity(regions []region.Region) float64 {
	if len(regions) == 0 {
		return 0
	}
	sum := 0.0
	for _, r := range regions {
		sum += r.InstallerCapacity
	}
	return sum / float64(len(regions))
}

// applyCompanyDeltas applies event-driven work rate and quality deltas to companies.
// Creates a new Companies map so the original IndustryState is not aliased.
func applyCompanyDeltas(w WorldState, deltas map[string]event.CompanyDelta) WorldState {
	if len(deltas) == 0 {
		return w
	}
	newCompanies := make(map[string]industry.CompanyState, len(w.Industry.Companies))
	for k, v := range w.Industry.Companies {
		newCompanies[k] = v
	}
	for defID, d := range deltas {
		if cs, ok := newCompanies[defID]; ok {
			cs.WorkRate = mathutil.Clamp(cs.WorkRate+d.WorkRateDelta, 0, 100)
			cs.AccumulatedQuality = mathutil.Clamp(cs.AccumulatedQuality+d.QualityDelta, 0, 1e6)
			newCompanies[defID] = cs
		}
	}
	w.Industry.Companies = newCompanies
	return w
}

// isValidMinisterState returns true for all recognised MinisterState values.
func isValidMinisterState(s stakeholder.MinisterState) bool {
	switch s {
	case stakeholder.MinisterStateActive,
		stakeholder.MinisterStateUnderPressure,
		stakeholder.MinisterStateLeadershipChallenge,
		stakeholder.MinisterStateDeparted,
		stakeholder.MinisterStateBackbench,
		stakeholder.MinisterStateOppositionShadow,
		stakeholder.MinisterStateAppointed,
		stakeholder.MinisterStateSacked,
		stakeholder.MinisterStateResigned,
		stakeholder.MinisterStateElectionOut:
		return true
	}
	return false
}

// isTerminalState returns true for states where no further simulation actions apply.
// SACKED and RESIGNED are transient states that transition to BACKBENCH; they are
// not terminal. BACKBENCH and OPPOSITION_SHADOW are live states.
func isTerminalState(s stakeholder.MinisterState) bool {
	return s == stakeholder.MinisterStateDeparted ||
		s == stakeholder.MinisterStateElectionOut
}

// rawValueForInsight returns the true world-state value (0-100) for the given
// insight type. This is the ground truth before quality and bias distortion are
// applied by evidence.GenerateReport.
func rawValueForInsight(t config.InsightType, w WorldState) float64 {
	switch t {
	case config.InsightPower:
		// Renewable grid share (0-100): how clean the national grid is.
		return mathutil.Clamp(w.EnergyMarket.RenewableGridShare, 0, 100)
	case config.InsightTransport:
		// Invert fossil dependency: high fossil = low decarbonisation progress.
		return mathutil.Clamp(100.0-w.FossilDependency, 0, 100)
	case config.InsightBuildings:
		// Average insulation quality across all housing tiles.
		return avgInsulationLevel(w.Tiles)
	case config.InsightIndustry:
		// Fraction of LCT companies currently active, scaled to 0-100.
		return activeCompanyFraction(w.Industry) * 100.0
	case config.InsightEconomy:
		return mathutil.Clamp(w.Economy.Value, 0, 100)
	case config.InsightPolicy:
		// Fraction of policy cards currently in force, scaled to 0-100.
		return activePolicyFraction(w.PolicyCards) * 100.0
	case config.InsightClimate:
		// Combine discrete level (0-3) and continuous intra-level severity (0-1)
		// into a monotone 0-100 signal.
		levelBase := float64(w.ClimateState.Level) / 3.0
		return mathutil.Clamp((levelBase+w.ClimateState.Severity/3.0)*100.0, 0, 100)
	case config.InsightFuelPoverty:
		return avgFuelPoverty(w.Tiles)
	case config.InsightRetrofit:
		return avgTrueRetrofitRate(w.Tiles)
	case config.InsightEnergyMarket:
		// Renewable grid share as proxy for market stability; gas price index
		// expansion is deferred pending energy market model growth.
		// PROXY: gas-price-based index not yet implemented.
		return mathutil.Clamp(w.EnergyMarket.RenewableGridShare, 0, 100)
	default:
		return 50.0 // neutral fallback for unrecognised insight types
	}
}

// sectorInstallerCapacity returns the average InstallerCapacity for regions
// relevant to the given policy sector.
//
// PROXY: per-policy region scoping is not stored in config; sector heuristics apply:
//   - INDUSTRY: average of regions tagged "industrial" (fallback: all regions).
//   - All other sectors: average of all regions.
func sectorInstallerCapacity(regions []region.Region, regionDefs []config.RegionDef, sector config.PolicySector) float64 {
	if len(regions) == 0 {
		return 0
	}
	// Build ID -> tags lookup from static config.
	tagsByID := make(map[string][]string, len(regionDefs))
	for _, d := range regionDefs {
		tagsByID[d.ID] = d.Tags
	}

	if sector == config.PolicySectorIndustry {
		sum, count := 0.0, 0
		for _, r := range regions {
			for _, tag := range tagsByID[r.ID] {
				if tag == "industrial" {
					sum += r.InstallerCapacity
					count++
					break
				}
			}
		}
		if count > 0 {
			return sum / float64(count)
		}
	}
	// Default: average of all regions.
	sum := 0.0
	for _, r := range regions {
		sum += r.InstallerCapacity
	}
	return sum / float64(len(regions))
}

// cabinetBudgetWeights derives per-department budget weights from the current
// cabinet's ideology and net-zero sympathy. Returns weights on the same scale
// as neutralMinisterWeights (1.0 = neutral; range 0.5-2.0).
//
// Role-to-department mapping:
//   - RoleLeader        -> DeptCross
//   - RoleChancellor    -> DeptBuildings, DeptIndustry
//   - RoleEnergy        -> DeptPower, DeptTransport
//   - RoleForeignSecretary -> no domestic dept (neutral retained)
// roleToDepts maps a ministerial Role to the domestic department IDs it
// influences for budget allocation. RoleForeignSecretary has no domestic
// departments and returns nil.
func roleToDepts(role config.Role) []string {
	switch role {
	case config.RoleLeader:
		return []string{government.DeptCross}
	case config.RoleChancellor:
		return []string{government.DeptBuildings, government.DeptIndustry}
	case config.RoleEnergy:
		return []string{government.DeptPower, government.DeptTransport}
	}
	return nil
}

func cabinetBudgetWeights(g government.GovernmentState, stakeholders []stakeholder.Stakeholder, lcrDelta float64) map[string]float64 {
	weights := neutralMinisterWeights()

	byID := make(map[string]stakeholder.Stakeholder, len(stakeholders))
	for _, s := range stakeholders {
		byID[s.ID] = s
	}

	for role, sid := range g.CabinetByRole {
		s, ok := byID[sid]
		if !ok {
			continue
		}
		depts := roleToDepts(role)
		if len(depts) == 0 {
			continue
		}
		stats := government.ComputeMinisterStats(s, lcrDelta)
		for _, dept := range depts {
			weights[dept] = stats.BudgetAllocationBias[dept]
		}
	}
	return weights
}

// avgInsulationLevel returns the mean InsulationLevel (0-100) across all tiles.
// Returns 50.0 if no tiles exist.
func avgInsulationLevel(tiles []region.Tile) float64 {
	if len(tiles) == 0 {
		return 50.0
	}
	sum := 0.0
	for _, t := range tiles {
		sum += t.InsulationLevel
	}
	return sum / float64(len(tiles))
}

// activeCompanyFraction returns the fraction of all LCT companies that are
// currently active, in [0, 1]. Returns 0 if there are no companies.
func activeCompanyFraction(ind industry.IndustryState) float64 {
	if len(ind.Companies) == 0 {
		return 0
	}
	active := 0
	for _, cs := range ind.Companies {
		if cs.IsActive {
			active++
		}
	}
	return float64(active) / float64(len(ind.Companies))
}

// significanceMultiplier returns the ideology conflict accumulation multiplier
// for a given policy significance level.
func significanceMultiplier(s config.PolicySignificance) float64 {
	switch s {
	case config.PolicySignificanceMajor:
		return signWeightMajor
	case config.PolicySignificanceModerate:
		return signWeightModerate
	default:
		return signWeightMinor
	}
}

// activeConsultancyOrgIDs returns the set of org IDs that have active (PENDING)
// commissions AND are of OrgType Consultancy.
func activeConsultancyOrgIDs(commissions []evidence.Commission, orgDefs []config.OrgDefinition) map[string]bool {
	defMap := make(map[string]config.OrgDefinition, len(orgDefs))
	for _, d := range orgDefs {
		defMap[d.ID] = d
	}
	out := make(map[string]bool)
	for _, c := range commissions {
		if !c.Delivered && !c.Failed {
			if def, ok := defMap[c.OrgID]; ok && def.OrgType == config.OrgConsultancy {
				out[c.OrgID] = true
			}
		}
	}
	return out
}

// activePolicyFraction returns the fraction of policy cards currently ACTIVE,
// in [0, 1]. Returns 0 if there are no cards.
func activePolicyFraction(cards []policy.PolicyCard) float64 {
	if len(cards) == 0 {
		return 0
	}
	active := 0
	for _, c := range cards {
		if c.State == policy.PolicyStateActive {
			active++
		}
	}
	return float64(active) / float64(len(cards))
}
