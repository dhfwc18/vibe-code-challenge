package simulation

import (
	"math"
	"math/rand"

	"twenty-fifty/internal/carbon"
	"twenty-fifty/internal/climate"
	"twenty-fifty/internal/config"
	"twenty-fifty/internal/economy"
	"twenty-fifty/internal/energy"
	"twenty-fifty/internal/event"
	"twenty-fifty/internal/evidence"
	"twenty-fifty/internal/government"
	"twenty-fifty/internal/industry"
	"twenty-fifty/internal/mathutil"
	"twenty-fifty/internal/player"
	"twenty-fifty/internal/policy"
	"twenty-fifty/internal/polling"
	"twenty-fifty/internal/region"
	"twenty-fifty/internal/reputation"
	"twenty-fifty/internal/save"
	"twenty-fifty/internal/stakeholder"
	"twenty-fifty/internal/technology"
)

// lobbyAPCost is the action point cost for the LobbyMinister action.
const lobbyAPCost = 3

// ministerSackingThreshold is the Popularity level below which a cabinet
// minister enters UNDER_PRESSURE. Reduced when GovernmentPopularity is high.
const ministerSackingThreshold = 25.0

// ministerSackingThresholdHighPop is used when GovernmentPopularity > 60.
const ministerSackingThresholdHighPop = 20.0

// ministerSackingWeeks is consecutive weeks UNDER_PRESSURE before SACKED.
const ministerSackingWeeks = 3

// ---------------------------------------------------------------------------
// Calibration constants
// ---------------------------------------------------------------------------

const (
	// baselineYearlyMt is the 2010 UK-equivalent baseline annual emissions.
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
)

// ---------------------------------------------------------------------------
// WorldState
// ---------------------------------------------------------------------------

// WorldState is the single source of truth for a game turn. All simulation
// logic reads and writes through AdvanceWeek, which returns a new copy each
// week. The RNG pointer and Cfg pointer are the only shared state across
// copies; both are treated as immutable (Cfg) or sequentially-advanced (RNG).
type WorldState struct {
	// Clock
	Week    int // absolute week; 0 = initial state, 1 = first processed week
	Year    int // calendar year: 2010-2050
	Quarter int // 1-4

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

	// Player
	Player player.CivilServant

	// Player-visible poll snapshots
	GovernmentLastPollResult float64            // most recent noisy govt approval sample (sigma=3)
	MinisterLastPollResults  map[string]float64 // keyed by stakeholder ID; sigma=5

	// Weekly transient accumulators -- reset at the top of each AdvanceWeek call.
	WeeklyNetCarbonMt        float64 // net carbon emitted this week (base minus reductions)
	WeeklyPolicyReductionMt  float64 // total carbon removed by active policies
	WeeklyEventLCRDelta      float64 // LCR delta from fired events
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

// NewWorld seeds a complete initial WorldState from config and a master seed.
// Week=0; Year=2010; Quarter=1. The first call to AdvanceWeek moves to Week=1.
func NewWorld(cfg *config.Config, masterSeed save.MasterSeed) WorldState {
	rng := rand.New(rand.NewSource(int64(masterSeed.DeriveSubSeed("simulation"))))

	w := WorldState{
		Week:                    0,
		Year:                    2010,
		Quarter:                 1,
		Cfg:                     cfg,
		RNG:                     rng,
		GovernmentPopularity:    initialGovtPopularity,
		GovernmentLastPollResult: initialGovtPopularity,
		FossilDependency:        initialFossilDependency,
		WeeksUntilLCRPoll:       10 + rng.Intn(7),
		MinisterLastPollResults: make(map[string]float64),
	}

	// Government: Common Wealth rules from 2010; first election ~2015.
	w.Government = government.NewGovernment(config.PartyLeft, initialElectionDueWeek)

	// Stakeholders: seed all; unlock START-timing figures; assign ruling party cabinet.
	w.Stakeholders = stakeholder.SeedStakeholders(cfg.Stakeholders)
	for i, s := range w.Stakeholders {
		if s.EntryTiming != config.TimingStart {
			continue
		}
		w.Stakeholders[i] = stakeholder.UnlockStakeholder(s, 0)
		w.Stakeholders[i].State = stakeholder.MinisterStateAppointed
		if s.Party == config.PartyLeft {
			w.Government = government.AssignMinister(w.Government, s.Role, s.ID)
		}
	}

	// Energy market: anchored to 2010 values.
	w.EnergyMarket = energy.NewMarket()

	// Climate: zero cumulative stock at game start.
	w.ClimateState = climate.DeriveClimateState(0.0)

	// Carbon: zero state; limits checked against cfg.CarbonBudgets each year.
	w.Carbon = carbon.CarbonBudgetState{}

	// Technology: seeded from logistic curve definitions.
	w.Tech = technology.NewTechTracker(cfg.Technologies)

	// Geography.
	w.Regions = region.SeedRegions(cfg.Regions)
	w.Tiles = region.SeedTiles(cfg.Tiles)

	// Economy.
	w.Economy = economy.NewEconomyState()

	// Reputation.
	w.LCR = reputation.NewLCR()

	// Policies: all cards start in DRAFT.
	w.PolicyCards = policy.SeedPolicyCards(cfg.PolicyCards)

	// Industry: all companies seeded inactive.
	w.Industry = industry.SeedIndustry(cfg.Companies)

	// Evidence: org relationship states.
	w.OrgStates = evidence.SeedOrgStates(cfg.Organisations)

	// Events: empty log, default pressure groups.
	w.EventLog = event.NewEventLog()
	w.PressureGroups = event.DefaultPressureGroups()

	// Player.
	w.Player = player.NewCivilServant()

	// Initial budget allocation (recomputed each quarter).
	w.LastTaxRevenue = economy.ComputeTaxRevenue(w.Economy, 1, 2010)
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
	w.WeeklyNetCarbonMt = baseWeeklyMt
	w.WeeklyPolicyReductionMt = 0
	w.WeeklyEventLCRDelta = 0
	w.WeeklyPolicyBudgetCostGBP = 0

	// Phase 1: Clock Advance.
	w = phaseClockAdvance(w)

	// Phase 2: Climate and Fossil Dependency Update.
	w = phaseClimateAndFossilUpdate(w)

	// Phase 3: Global Events Roll.
	var firedEvents []event.EventEntry
	w, firedEvents = phaseGlobalEventRoll(w)

	// Phase 4: Scandal and Pressure Roll.
	w = phaseScandalAndPressureRoll(w)

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

	// Phase 12: Polling Check.
	w = phasePollingCheck(w)

	// Phase 13: Consequence Resolution (policy approval evaluation for cards
	// already UNDER_REVIEW). Runs before player actions so that policies
	// submitted this tick are not evaluated until the following week.
	w = phaseConsequenceResolution(w)

	// Phase 14: Player Action Phase.
	w = phasePlayerActions(w, actions)

	// Phase 15: Minister Health Check.
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
	// Year: week 1-52 = 2010, week 53-104 = 2011, etc.
	w.Year = 2010 + w.Week/52
	// Quarter: 1-4 within each 52-week year.
	w.Quarter = 1 + (w.Week%52)/13

	// Election check: trigger when the scheduled week arrives.
	if government.IsElectionDue(w.Government, w.Week) {
		// Winner is the leading party in the most recent poll; default to current ruler.
		winner := w.Government.RulingParty
		if len(w.PollHistory) > 0 {
			winner = polling.LeadingParty(w.PollHistory[len(w.PollHistory)-1])
		}
		// Clear cabinet; set new ruling party; schedule next election 5 years out.
		w.Government = government.TriggerElection(w.Government, winner, w.Week+260)

		// Rebuild cabinet from winning party's unlocked non-terminal ministers.
		// All other unlocked ministers move to OPPOSITION_SHADOW.
		for i := range w.Stakeholders {
			s := w.Stakeholders[i]
			if !s.IsUnlocked || isTerminalState(s.State) {
				continue
			}
			if s.Party == winner {
				w.Government = government.AssignMinister(w.Government, s.Role, s.ID)
				w.Stakeholders[i].State = stakeholder.MinisterStateAppointed
				w.Stakeholders[i].WeeksUnderPressure = 0
			} else {
				w.Stakeholders[i].State = stakeholder.MinisterStateOppositionShadow
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

	entry := event.EventEntry{DefID: def.ID, Name: def.Name, Week: w.Week, Effects: resolved}
	w.EventLog = event.AppendEventLog(w.EventLog, entry)
	return w, []event.EventEntry{entry}
}

// ---------------------------------------------------------------------------
// Phase 4: Scandal and Pressure Roll
// ---------------------------------------------------------------------------

func phaseScandalAndPressureRoll(w WorldState) WorldState {
	// Scandal roll for each active unlocked stakeholder.
	for _, s := range w.Stakeholders {
		if !s.IsUnlocked || isTerminalState(s.State) {
			continue
		}
		if event.RollScandal(s, s.WeeksUnderPressure, w.RNG) {
			w.GovernmentPopularity = mathutil.Clamp(w.GovernmentPopularity-2.0, 0, 100)
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
	// Advance each tech along its logistic curve. Acceleration bonuses from
	// R&D policies will be added once active policy effects are plumbed through.
	for _, curve := range w.Cfg.Technologies {
		w.Tech = technology.AdvanceTick(w.Tech, curve, 0.0)
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
	// Uses the first region as representative for capacity multiplier scaling.
	// National-scope BaseCarbonDeltaMt values are applied once per policy.
	repRegion := representativeRegion(w.Regions)
	avgRetrofit := avgTrueRetrofitRate(w.Tiles)

	for _, card := range w.PolicyCards {
		if card.State != policy.PolicyStateActive {
			continue
		}
		techFrac := techFractionForPolicy(card.Def, w.Tech)
		delta := policy.ResolveWeeklyEffect(card, repRegion, techFrac, avgRetrofit)
		w.WeeklyPolicyReductionMt += delta.DeltaMt
		w.WeeklyPolicyBudgetCostGBP += delta.BudgetCostPerWeek
	}

	// Net carbon: base emissions minus policy reductions. Clamped to prevent
	// negative values (policies cannot emit more carbon than the baseline).
	w.WeeklyNetCarbonMt = mathutil.Clamp(
		baseWeeklyMt-w.WeeklyPolicyReductionMt, 0, baseWeeklyMt*2)
	return w
}

// ---------------------------------------------------------------------------
// Phase 10: Carbon Budget Accounting
// ---------------------------------------------------------------------------

func phaseCarbonBudgetAccounting(w WorldState) WorldState {
	w.Carbon = carbon.AccumulateWeekly(w.Carbon, w.WeeklyNetCarbonMt)

	// Year-end: check annual carbon budget limit against CCC targets.
	if w.Week%52 == 0 {
		_, w.Carbon = carbon.CheckAnnualBudget(w.Carbon, w.Year, w.Cfg.CarbonBudgets)
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
	w.LCR = reputation.TickReputation(w.LCR, w.WeeklyPolicyReductionMt, w.WeeklyEventLCRDelta)
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

	// Quarter-end: clear lobby effects, recompute tax revenue and budget allocation.
	if w.Week%13 == 0 {
		w.Economy = economy.ClearLobbyEffectsAtQuarterEnd(w.Economy)
		w.LastTaxRevenue = economy.ComputeTaxRevenue(w.Economy, w.Quarter, w.Year)
		w.LastBudget = economy.AllocateBudget(
			w.LastTaxRevenue,
			baseDeptFractions(),
			neutralMinisterWeights(),
			reputation.ChainToBudgetModifier(w.LCR.Value),
			w.Economy.LobbyEffects,
		)
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
			w.Player = player.RecordAction(w.Player, player.ActionRecord{
				ActionType: a.Type, Week: w.Week, APCost: lobbyAPCost,
			})
			break
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
		// Check Murican org availability (tickyPresent/tickyPressureAccepted wired in tuning pass).
		if !evidence.MuracanOrgAvailable(orgDef, os, false, false) {
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

		// Popularity-based state transitions for cabinet ministers only.
		updated := w.Stakeholders[i]
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
		switch s.State {
		case stakeholder.MinisterStateAppointed:
			// APPOINTED -> ACTIVE: 4-week grace period deferred to simulation tuning;
			// immediate transition ensures valid states from week 1 onward.
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
		}
	}
	return w
}

// ---------------------------------------------------------------------------
// Phase 16: Consequence Resolution
// ---------------------------------------------------------------------------

func phaseConsequenceResolution(w WorldState) WorldState {
	// Evaluate approval steps for UNDER_REVIEW policies.
	// Runs before player actions so submissions this tick are not evaluated until next tick.
	active := unlockedStakeholders(w.Stakeholders)
	for i, card := range w.PolicyCards {
		if card.State != policy.PolicyStateUnderReview {
			continue
		}
		w.PolicyCards[i] = policy.EvaluateApproval(card, active)
		// If all approval steps cleared, activate immediately.
		if w.PolicyCards[i].State == policy.PolicyStateApproved {
			w.PolicyCards[i] = policy.ActivatePolicy(w.PolicyCards[i])
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
		}

		if c.Failed {
			continue
		}

		// Generate the insight report. rawValue is a neutral placeholder;
		// the simulation tuning pass will provide per-topic truth values.
		report := evidence.GenerateReport(c, orgDef, string(c.InsightType), 50.0, 0.0, w.RNG)
		w.Reports = append(w.Reports, report)
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

// representativeRegion returns the first region for use as a capacity proxy
// in national policy resolution. Returns a zero-value Region if none exist.
func representativeRegion(regions []region.Region) region.Region {
	if len(regions) == 0 {
		return region.Region{}
	}
	return regions[0]
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
func isTerminalState(s stakeholder.MinisterState) bool {
	return s == stakeholder.MinisterStateDeparted ||
		s == stakeholder.MinisterStateSacked ||
		s == stakeholder.MinisterStateElectionOut
}
