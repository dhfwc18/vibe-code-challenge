# 20-50 -- Implementation Plan

Last updated: 2026-03-28

This document is the authoritative build plan. It derives from the design in
game_design.md. Read that document first if anything here is unclear.

---

## Current State

Four Go files exist:
  app/cmd/app/main.go               -- opens a 1280x720 Ebitengine window
  app/internal/game/game.go         -- Game struct, Update/Draw/Layout stubs
  app/internal/game/colours.go      -- background colour constant
  app/internal/game/game_test.go    -- 3 passing tests

Nothing else exists. All game logic is to be built.

---

## Build Strategy

Build bottom-up, one layer at a time. Each package must be fully tested before
the next layer begins. No skipping ahead to UI before simulation is solid.

Each layer is independent of layers above it. Dependencies flow downward only.
The simulation package (Layer 5) is the only package that calls all the others;
the ui package (Layer 6) only reads WorldState and never mutates it.

Rule: every exported function gets at least one test before the next function
is written. The headless simulation test (Layer 5) is the integration gate --
if 100 weeks run clean with no panics, no negative budgets, and no invalid
state machine transitions, the simulation is ready for UI work.

---

## Layer 0 -- Foundation

Packages: config, save
No game dependencies. These are the bedrock everything else reads from.

### config

Purpose: load and expose all static data at startup. Zero game logic.

Structs to define:
  PolicyCardDef         -- id, name, sector, apCost, budgetCost, approvalRequirements,
                           weeklyEffects (carbon delta formula params), techUnlockGate
  OrgDefinition         -- id, name, orgType, origin, baseCost, deliveryDist,
                           qualityRange, biasType, biasDirection, clientBiasWeight,
                           popularityRisk, baseFailureProbability, specialisms
  StakeholderSeed       -- id, party, role, entryTiming, name, nickname, biography,
                           ideologyScore, netZeroSympathy, riskTolerance, populismScore,
                           diplomaticSkill, consultancyAffinity, signals, specialMechanic
  CompanyDef            -- id, name, techCategory, originSize, baseQuality,
                           baseWorkRate, taitonHq
  EventDef              -- id, eventType, severity, baseProbability, climateMultiplier,
                           fossilMultiplier, effects
  TechCurveDef          -- id, name, sector, logisticMidpoint, logisticSteepness,
                           baseAdoptionRate, costAnchors
  CarbonBudgetDef       -- year, annualLimitMtCO2e (from CCC targets)
  RegionDef             -- id, name, tileIDs, initialSkillsNetwork,
                           initialInstallerCapacity, initialSupplyChain
  TileDef               -- id, regionID, name, initialInsulationLevel,
                           initialHeatingType, initialLocalIncome, initialPoliticalOpinion

Data to seed:
  - 16+ stakeholder seeds per party pool (all 4 parties, with entry timing)
  - 18 organisation definitions (15 local + 3 Murican)
  - 15 base LCT company definitions
  - 40-60 policy card definitions across 4 sectors
  - 40-60 event definitions across 6 event types
  - 8 technology curve definitions
  - Carbon budget table 2010-2050 (from green_book_reference.md)
  - 12 region definitions with 60-80 tile definitions total
  - Historical energy price anchors (2010: gas GBP25/MWh, electricity GBP50/MWh,
    oil GBP80/barrel)

Tests: loader returns correct counts, required fields are non-zero,
all IDs are unique, delivery distributions are valid (min <= mode <= max).

### save

Purpose: serialise and deserialise WorldState. Own the master seed.

Key responsibilities:
  - WriteState(path, WorldState) error
  - ReadState(path) (WorldState, error)
  - Version tag validation: if save version != current version, return
    ErrIncompatibleSave -- never attempt migration
  - Master seed management: NewMasterSeed() generates a secure random seed;
    all stochastic sub-seeds derive deterministically from it via hash chain
  - Sub-seed derivation: SubSeed(master uint64, domain string) uint64

Tests: round-trip (write then read produces identical state), incompatible
version returns correct error, sub-seeds are deterministic (same master ->
same sub-seed for same domain string), different domains produce different seeds.

---

## Layer 1 -- Pure Domain Models

Packages: carbon, technology, region
These model the physical/spatial world. Pure functions where possible.

### carbon

Key types:
  CarbonBudgetState -- runningAnnualTotal, cumulativeStock, overshootAccumulator,
                       currentBudgetLimit, trajectory float64 (0 on-track, positive
                       means overshoot pace)

Key functions:
  AccumulateWeekly(state, weeklyNetMtCO2e) CarbonBudgetState
  CheckAnnualBudget(state, year) BudgetCheckResult
  UpdateCumulativeStock(state) CarbonBudgetState
  ProjectTrajectory(state, remainingWeeks) float64  -- simple linear projection
  ClimateStockToState(cumulativeStock) ClimateLevel  -- 4 severity levels

Anchors from green_book_reference.md must be embedded as constants:
  Baseline2010 = 590.0  // MtCO2e/yr
  Target2050   = 0.0
  Carbon budget limits per CCC table

Tests: accumulation arithmetic, budget check fires at correct year boundary,
trajectory projection is monotone, climate level boundaries match design spec.

### technology

Key types:
  TechTracker  -- maturity float64 (0-100) per technology
  TechSnapshot -- all 8 maturities at a point in time

Key functions:
  AdvanceTick(tracker, curve, accelerationBonus) TechTracker
  EvaluateLogistic(x, midpoint, steepness) float64  -- pure maths
  ApplyAccelerationBonus(tracker, bonusMap) TechTracker

Technologies: OffshoreWind, OnshoreWind, SolarPV, Nuclear, HeatPumps,
              ElectricVehicles, Hydrogen, IndustrialCCS

Tests: logistic curve properties (0->1 monotone, midpoint = 0.5, steepness
controls slope), bonus correctly accelerates maturity, maturity clamps to 100.

### region

Key types:
  Region  -- id, name, skillsNetwork, installerCapacity, supplyChain, tileIDs
  Tile    -- (see data structures in game_design.md for full field list)
  FuelPovertyInput  -- prices, heatingType, insulationLevel, localIncome,
                       heatingCapacity, techMaturityHeatPumps, seasonalMultiplier

Key functions:
  ComputeFuelPoverty(FuelPovertyInput) float64     -- full formula from design
  ComputeTrueRetrofitRate(observed, quality) float64
  UpdateLocalPoliticalOpinion(tile, fuelPovertyDelta, climateEventImpact) Tile
  CapacityMultiplier(region) float64               -- for policy resolution
  ApplyClimateEventDamage(tile, event) Tile
  RevealAttribute(tile, attribute) Tile            -- fog-of-war unlock

Fuel poverty formula (from design):
  heatingDemand    = baseHeatingDemand * insulationFactor * seasonalMultiplier
  totalFuelCost    = heatingCostPerUnit * heatingDemand
  FuelPoverty      = clamp((totalFuelCost / localIncome) * povertyScalingWeight, 0, 100)
  insulationFactor = 1 - (insulationLevel / 200)      // 0=no insulation, 1=perfect
  heatingCostPerUnit depends on HeatingType (Gas/Oil/Electric/HeatPump/Mixed)
  COP factor applies for HeatPump: effectiveCost = electricityPrice / COP(techMaturity)

Tests: fuel poverty formula at boundary conditions (zero income edge case,
fully insulated tile, heat pump COP scaling), true retrofit rate is always <=
observed, capacity multiplier is in [0,1], fog-of-war correctly gates visibility.

---

## Layer 2 -- Derived World Models

Packages: energy, climate, economy, reputation

### energy

Key types:
  EnergyMarket  -- gasPrice, electricityPrice, oilPrice, renewableGridShare,
                   priceHistory (ring buffer, 52 weeks per fuel)

Key functions:
  TickPrices(market, globalDelta, carbonLevy, renewableSubsidy, gridShare) EnergyMarket
  GridCouplingModel(gasPrice, gridShare) float64  -- electricity price floor when
                                                   -- renewables are low
  ApplyShock(market, shockEvent) EnergyMarket
  HistoricalAnchor(year int) EnergyMarket         -- seeds starting prices from real data

Grid coupling rule (from design):
  if renewableGridShare < 40%:
    electricityPriceFloor = gasPrice * conversionFactor * (1 - gridShare/100)
  else:
    floor decouples; electricity tracks renewable LCOE trajectory

Tests: historical anchor values match green_book_reference.md constants,
grid coupling floor applies when gridShare < 40 and not above 40, shock
raises price by expected magnitude, ring buffer wraps correctly.

### climate

Key types:
  ClimateState    -- level ClimateLevel, severity float64 (continuous within level)
  ClimateEvent    -- type, severity, affectedRegions, weekOfYear
  EnergyShockEvent -- shockType, priceDelta, fossilMultiplier
  ShockResponseCard -- id, shockType, options (Accept/Decline/Mitigate),
                        backfireProbability

Key functions:
  DeriveClimateState(cumulativeStock) ClimateState
  RollClimateEvent(state, fossilDependency, weekOfYear, rng) *ClimateEvent
  RollEnergyShock(fossilDependency, rng) *EnergyShockEvent
  ShockResponseOutcome(card, option, lcr, playerReputation, rng) ResponseResult
  BackfireProbability(lcr, playerReputation) float64

ClimateLevel thresholds (from design): STABLE < 200 MtCO2e cumulative above
baseline, ELEVATED 200-400, CRITICAL 400-600, EMERGENCY 600+. (Exact numbers
to be tuned during balance pass; these are starting anchors.)

Tests: deterministic outcomes from fixed rng seed, climate level transitions
at correct thresholds, shock probability scales correctly with fossilDependency,
backfire probability is monotone-decreasing in lcr.

### economy

Key types:
  EconomyState  -- hidden float64 (0-100)
  TaxRevenue    -- GBP bn, computed quarterly
  FossilDependency -- derived float64 (0-100%), semi-visible

Key functions:
  TickEconomy(state, climateDamage, fuelPovertyAggregate, shockSeverity,
              policyBonus, fossilDrag) EconomyState
  ComputeTaxRevenue(economy, quarter) float64
  AllocateBudget(taxRevenue, departments, ministerWeights, lcrModifier,
                 lobbyEffect) BudgetAllocation
  AccumulateLobbyEffect(current, newEffect) float64
  ClearLobbyEffectsAtQuarterEnd(state) EconomyState
  DeriveFossilDependency(energyMix) float64

Budget allocation formula (from design):
  share_d = baseFraction_d * ministerPopWeight_d * lcrModifier * lobbyEffect_d
  Normalised across departments so total = TaxRevenue * discretionaryFraction

Tests: economy is in [0,100] after any tick, tax revenue is non-negative,
budget shares sum to <= total, lobby effects clear at quarter boundary,
fossil dependency derivation matches expected range given energy mix inputs.

### reputation

Key types:
  LowCarbonReputation  -- value float64 (0-100), lastPollResult float64

Key functions:
  TickReputation(lcr, weeklyPolicyCarbonDelta, eventImpact) float64
  ChainToMinisterPopularity(lcrDelta) float64    -- returns delta to apply
  ChainToGovtPopularity(lcrDelta) float64
  ChainToBudgetModifier(lcr) float64
  PollLCR(trueValue, rng) float64                -- adds Gaussian noise sigma=4
  CapitalisationSuccess(lcr, playerReputation, rng) bool

Tests: chain effects have correct sign and magnitude, poll noise has correct
sigma, capitalisation probability is monotone in lcr and playerReputation.

---

## Layer 3 -- Agent Models

Packages: stakeholder, government, polling, industry

### stakeholder

This is the most complex package. Contains all 16 political figures + PM.

Key types:
  Stakeholder       -- full field list from game_design.md (all hidden + tracked fields)
  StakeholderState  -- ACTIVE | UNDER_PRESSURE | LEADERSHIP_CHALLENGE | DEPARTED |
                       BACKBENCH | OPPOSITION_SHADOW
  PartyPool         -- map[Party]map[Role][]StakeholderSeed with entry timing gates
  RelationshipScore -- float64 with 5-label helper

Key functions:
  SeedPartyPool(config, masterSeed) map[Party]PartyPool
  TickEntryGates(pool, currentWeek, govtHistory) PartyPool  -- unlock MID/LATE figures
  EvaluateMinisterState(s Stakeholder, worldSnapshot) []TransitionEvent
  EvaluatePMState(pm Stakeholder, govtState) []TransitionEvent
  TickConsultancyAffinity(s Stakeholder, orgs OrganisationState) OrgRelDelta
  GenerateTickyPressure(ticky Stakeholder, week int, rng) *TickyPressureEvent
  EvaluateDizzySurge(truscott Stakeholder, isPM bool) *DizzySurgeEvent
  EvaluateElectoralFatigue(jj Stakeholder, electionLossCount int) bool
  InteractionAPCost(s Stakeholder, playerRel float64) int

State transitions from design (minister 8-state machine + PM 4-state machine):
  Implement as pure EvaluateXState(snapshot) []TransitionEvent pattern.
  Mutating the state lives in simulation only.

Tests: each state transition fires under correct conditions and not others,
AP cost scales with seniority and relationship, consultancy affinity bonus
applies only when governing and does not exceed 100, Ticky pressure event
fires within 6-10 week window, electoral fatigue triggers after exactly 2 losses.

### government

Key types:
  GovernmentState      -- STABLE | UNDER_PRESSURE | CONFIDENCE_VOTE | SNAP_ELECTION |
                          CARETAKER | ELECTION_CAMPAIGN
  GovernmentPopularity -- aggregated from tile -> region -> national
  ElectionResult       -- winningParty, marginOfVictory, seatCounts

Key functions:
  AggregatePopularity(tiles []Tile, regions []Region) float64
  EvaluateGovtState(state GovernmentState, popularity, weeksUnstable) []TransitionEvent
  ResolveElection(govtPopularity, opponentPopularity, rng) ElectionResult
  ScheduleElection(currentWeek, maxTermWeeks) int  -- returns election week
  PopularityHistory(history [52]float64) PopularityStats  // trend, volatility

Tests: popularity aggregation is weighted correctly, election fires at max
term if not triggered early, election result is monotone in popularity gap.

### polling

Regional political polling using a Gaussian proximity kernel. Built richer than
the original plan; the full regional model is preferred because it makes tile
opinion changes visible in the polls and gives the player meaningful feedback on
local policy impacts.

Key types:
  RegionPoll    -- regionID, partyShares map[Party]float64 (sums to ~100), swing float64
  PollSnapshot  -- week, regionPolls map[string]RegionPoll, nationalPolls map[Party]float64

Algorithm (AggregateRegionPoll):
  1. Filter tiles to regionID; compute HeatingCapacity-weighted mean PoliticalOpinion.
  2. Map opinion (0-100) to party shares via Gaussian proximity kernel (bandwidth=25).
     Party axis positions: FarLeft=10, Left=35, Right=65, FarRight=90.
  3. Add per-party Gaussian noise (sigma=3) and renormalise to 100.
  4. Apply minimum share floor (0.5%) and renormalise again.

National poll: InstallerCapacity-weighted average of all region polls.

Key functions:
  AggregateRegionPoll(tiles, regionID, rng) RegionPoll
  AggregateNationalPoll(regionPolls, regions) map[Party]float64
  TakePollSnapshot(week, tiles, regions, rng) PollSnapshot
  LeadingParty(snap) Party              -- deterministic tie-break by party ID sort
  SwingFromLast(current, previous) map[string]float64

Tests: leading party is the one with highest share, swing sign is correct,
national poll is weighted average of regions, noise is bounded, minimum share
floor prevents any party vanishing, empty-region fallback returns neutral 50.

### industry

Key types:
  Company           -- static seed + dynamic state (workRate, quality, state)
  CompanyState      -- ACTIVE | STRUGGLING | BANKRUPT | ABSORBED | STARTUP
  CompanyWeeklyWork -- techMaturityDelta, installerCapacityDelta

Key functions:
  SeedCompanies(config, masterSeed) []Company
  WeeklyWork(company, techMaturity, regionCapacity) CompanyWeeklyWork
  ApplyPlayerIntervention(company, intervention) Company
  EvaluateCompanyState(company, economyState, competitorCount) []TransitionEvent
  GenerateLobbyingEvent(company, rng) *LobbyEvent
  SpawnStartup(techCategory, masterSeed, week int) Company

Tests: weekly work output is bounded, struggling transition fires below
threshold, bankrupt fires below lower threshold, lobbying event is within
expected frequency, startup spawn produces valid company.

---

## Layer 4 -- Player-Facing Mechanics

Packages: policy, evidence, event, player

### policy

Key types:
  PolicyCard     -- static def + dynamic approval state
  PolicyState    -- DRAFT | SUBMITTED | UNDER_REVIEW | APPROVED | REJECTED |
                    ACTIVE | ARCHIVED
  ApprovalStep   -- ministerRole, required relationship, ideology threshold
  WeeklyEffect   -- carbonDeltaFormula, budgetCost, capacityModifier

Key functions:
  SubmitPolicy(card, worldSnapshot) PolicyCard      -- moves to UNDER_REVIEW
  EvaluateApproval(card, stakeholders) PolicyState  -- pure; called weekly
  ResolveWeeklyEffect(card, region, techMaturity, trueRetrofitRate) CarbonDelta
  ArchivePolicy(card, reason) PolicyCard

Note: carbon effect uses TrueRetrofitRate (not Observed). The gap is silent
until a quality audit report surfaces it.

Tests: approval pipeline state machine, weekly effect uses correct retrofit
rate, archived policy stops contributing effects, unlock gate blocks submission
when tech maturity < threshold.

### evidence

Key types:
  Commission       -- full field list from game_design.md
  InsightReport    -- rawValue, reportedValue (quality + bias distorted), topicKey
  OrgState         -- relationship score, commissionCount, state, coolingOffUntil
  DeliveryTimer    -- draws from triangular distribution at commission time

Key functions:
  CreateCommission(org, insightType, scope, week, budget, rng) Commission
  TickDelivery(commissions, week) []Commission         -- returns delivered ones
  GenerateReport(commission, worldSnapshot, rng) InsightReport
  ApplyBias(rawValue, biasType, biasDirection, clientBias,
            recentPlayerDecisions) float64
  UpdateOrgRelationship(org, event RelationshipEvent) OrgState
  MuracanOrgAvailable(org, tickyPresent, tickyPressureAccepted) bool

Bias model:
  ClientConfirmation: reportedValue shifts toward validating player's last 3 decisions
  Ideological: reportedValue shifts toward biasDirection
  None: small Gaussian noise only

Tests: delivery timer draws from correct distribution (run 1000, check mean ~
mode, check min/max respected), client bias shifts in correct direction,
ideological bias has correct sign, Murican orgs are locked until Ticky event,
failure mechanic fires at correct probability.

### event

Design principle: all event definitions live in config seed data. The event
package is a pure dispatcher -- no per-event hardcoded logic. New events can
be added or edited in config/events.go without touching any event package code.

Key types:
  EventEntry       -- one fired event: def ID, week, resolved effects
  EventLog         -- circular buffer, player-visible, 52 weeks retention
  ShockResponseCard -- queued for player action when OffersShockResponse=true
  PressureGroup    -- persistent actor; accumulates pop/carbon/LCR-triggered events
  PressureResult   -- output from ApplyPressureGroups
  ResolvedEffect   -- flat map of concrete changes produced by ResolveEffect
  RegionDelta      -- InstallerCapacityDelta, SkillsNetworkDelta per region ID
  TileDelta        -- FuelPovertyDelta, InsulationDamage per tile ID
  StakeholderDelta -- RelDelta, PressureDelta per stakeholder ID
  CompanyDelta     -- WorkRateDelta, QualityDelta per company ID

Key functions:
  ComputeEventProbability(def, climateLevel, fossilDependency) float64
  DrawEvent(defs, climateLevel, fossilDependency, rng) (config.EventDef, bool)
  RollScandal(stakeholder, weeklyPressure, rng) bool
  NewEventLog() EventLog
  AppendEventLog(log, entry) EventLog
  DefaultPressureGroups() []PressureGroup
  ApplyPressureGroups(groups, carbonTrajectory, lcr) ([]PressureGroup, []PressureResult)
  MatchRegions(filter string, regions []config.RegionDef) []string   -- returns region IDs
  MatchStakeholders(filter string, stakeholders []stakeholder.Stakeholder) []string
  MatchCompanies(filter string, companies []industry.CompanyState) []string
  ResolveEffect(effect config.EventEffect, regions, stakeholders, companies) ResolvedEffect

Targeting filter semantics (resolved by MatchRegions / MatchStakeholders / MatchCompanies):
  RegionFilter:      "COASTAL","RURAL","URBAN","INDUSTRIAL","AGRICULTURAL" = tag match;
                     any other string = exact region ID; empty = all regions.
  StakeholderFilter: "CABINET" = all 4 role-holders; "ROLE:ENERGY" etc = role match;
                     "ALL" = all unlocked stakeholders.
  CompanyFilter:     "ALL" = all active companies; "TECH:EVS" etc = TechCategory match.

Tests:
  ComputeEventProbability scales correctly with climate and fossil multipliers.
  DrawEvent returns false when all probabilities are zero.
  MatchRegions_CoastalFilter_ReturnsOnlyCoastalRegions
  MatchRegions_RegionIDFilter_ReturnsThatRegion
  MatchStakeholders_CabinetFilter_ReturnsFourRoleHolders
  MatchCompanies_TechFilter_ReturnsOnlyMatchingCategory
  ResolveEffect_CoastalFloodingEffect_PopulatesRegionAndTileDeltas
  ResolveEffect_EmptyFilters_ProducesNoTargetedDeltas
  ApplyPressureGroups_HighCarbon_GeneratesEvent
  RollScandal_ZeroPressure_NeverFires (10 000 iterations)

### player

Key types:
  CivilServant   -- apPool, staffRoster, reputationGrade, actionHistory
  StaffMember    -- role, apBonus, weekHired
  ActionRecord   -- action type, week, AP cost, outcome

Key functions:
  StartWeekAPPool(player, staffRoster) int
  SpendAP(player, cost) (CivilServant, error)
  RecordAction(player, action) CivilServant
  ReputationGrade(playerRep float64) string   -- maps to civil service grade labels
  ProcessShockResponse(player, card, choice, worldSnapshot) ActionRecord

Tests: AP pool starts at correct value including staff bonus, overspend
returns error, reputation grade boundaries map correctly.

---

## Layer 5 -- Orchestration

Package: simulation

This is the most important package. It owns WorldState and runs the 17-phase
pipeline. It imports all Layer 0-4 packages.

Key types:
  WorldState   -- single source of truth (see full struct in game_design.md)
  TurnEngine   -- owns WorldState, exposes AdvanceWeek() and PlayerAction()
  EventBus     -- lightweight pub/sub for cross-package event delivery

17-phase pipeline (must execute in this exact order):
  Phase  1: ClockAdvance
  Phase  2: ClimateAndFossilUpdate
  Phase  3: GlobalEventRoll
  Phase  4: ScandalAndPressureRoll
  Phase  5: TechnologyProgressTick
  Phase  6: RegionalWorldTick
  Phase  7: TileLocalTick
  Phase  8: ClimateEventImpactOnTiles
  Phase  9: PolicyResolution
  Phase 10: CarbonBudgetAccounting
  Phase 11: EconomyAndTaxRevenueTick
  Phase 12: PollingCheck
  Phase 13: PlayerActionPhase  (blocks; waits for UI input or headless input)
  Phase 14: MinisterHealthCheck
  Phase 15: MinisterTransitions
  Phase 16: ConsequenceResolution
  Phase 17: ConsultancyDeliveryCheck
  Phase 18: EndOfWeekRender (UI only; no-op in headless mode)

Key functions:
  NewWorld(config, masterSeed) WorldState
  AdvanceWeek(world, playerActions []Action) (WorldState, []Event)
  HeadlessRun(world, weeks int) SimulationReport

Integration test (the gate to Layer 6):
  Run 100 weeks headless from a fixed seed.
  Assert: no panics, no negative budgets, no invalid state machine states,
  carbon accumulation is bounded above 0, all stakeholders have valid states,
  at least one climate event has fired, at least one poll has been generated.

---

## Layer 6 -- Presentation

Package: ui

Built on Ebitengine + ebitenui. Reads WorldState; never mutates it.
All player intent expressed as Action structs passed back to simulation.

8 tabs (one scene each):

  Overview
    - Game clock (year, week, quarter)
    - AP remaining this week + breakdown
    - Event log feed (last 10 entries, scrollable)
    - Key indicators: GovernmentPopularity (poll), LastLCRPoll, TaxRevenue
    - Next election countdown

  Map
    - Choropleth of 12 regions, coloured by LocalPoliticalOpinion aggregate
    - Click region to expand tile list
    - Click tile to see revealed attributes (fog-of-war gates unrevealed ones)
    - Climate event overlay (show affected tiles)
    - Legend for all colouring modes

  Politics
    - 4 party columns, each with 4 role rows + pool depth indicator
    - Each figure: name, role, state badge, relationship score label
    - Click figure: detail panel (signals, available actions, AP cost)
    - PM panel: state, popularity, weeks remaining
    - Election calendar

  Policy
    - Policy card browser: filter by sector, search by name
    - Card detail: effects, approval requirements, budget cost
    - Approval pipeline: DRAFT -> ... -> ACTIVE column view
    - Active policies: list with weekly carbon contribution

  Energy
    - Price dashboard: GasPrice, ElectricityPrice, OilPrice (live + 52w chart)
    - RenewableGridShare gauge
    - Grid coupling indicator (is electricity still tracking gas?)
    - Fuel poverty risk index (aggregate, no tile detail)

  Industry
    - Company roster: 15 base + emergent
    - State badge per company (ACTIVE / STRUGGLING / BANKRUPT etc.)
    - TechCategory grouping
    - Intervention buttons: Contract, Grant, Set Standards, Investigate
    - TechMaturity progress bars (visible after consultancy reveals)

  Evidence
    - Organisation list: 15 local (always visible) + 3 Murican (visible only
      after TICKY_PRESSURE event)
    - Active commissions with delivery countdown
    - Report inbox: delivered reports with reportedValue display
    - Cross-reference view: group all reports by topicKey side by side
    - Relationship score label per org

  Budget
    - Department budget bars (current allocation)
    - Tax revenue history (quarterly, 4-quarter view)
    - Lobby effect tracker (pending effects for next quarter)
    - Budget allocation breakdown (minister weight contributions)

UI rules:
  - Never display a true value that design marks as hidden
  - All hidden values shown only if a consultancy report has revealed them
  - Fog-of-war: tile attributes masked with "?" until reveal
  - All monetary values in GBP millions or billions as appropriate
  - All carbon values in MtCO2e
  - Ebitengine target: 60fps, layout 1280x720, resizable

---

## Layer 7 -- Entry Point

File: app/cmd/app/main.go (already exists)

Expand to:
  - Load config from embedded data files
  - Initialise or load WorldState (new game or save)
  - Wire simulation.TurnEngine and ui.Root
  - Pass player actions from UI to simulation each tick
  - Save on quit

This layer is trivial once Layers 0-6 exist. Estimated: 50 lines.

---

## Playable Milestones

Milestone 1 -- Headless Green (end of Layer 5)
  100 weeks run headless from fixed seed with all invariants passing.
  Nothing visible yet. The game exists as a testable simulation.

Milestone 2 -- Clock Ticking (early Layer 6)
  Overview tab live. Clock advances visibly. Event log shows real events.
  AP pool displayed. GovernmentPopularity poll result appears. Can press
  "end week" to advance. Earliest point the game is visually "alive".

Milestone 3 -- Map and Politics (mid Layer 6)
  Map tab live with regional colours and tile drill-down.
  Politics tab live with stakeholder states and relationship labels.
  Player can "meet" a stakeholder and see AP deducted.

Milestone 4 -- First Policy (mid Layer 6)
  Policy tab live. Player can draft, submit, and (after approval pipeline
  resolves) see a policy go ACTIVE and contribute weekly carbon reductions.
  Carbon budget running total visible in Overview.

Milestone 5 -- Full Tab Complement (end of Layer 6)
  All 8 tabs functional. Energy prices charted. Industry companies visible.
  Evidence commissions can be created and reports delivered.
  Budget allocation visible.

Milestone 6 -- First Complete Game (integration + balance)
  Player can start in 2010, make decisions, reach 2050 (or fail).
  Win condition: net zero achieved. Fail condition: 2050 clock hits with
  positive emissions.
  Initial balance pass: tune carbon reduction rates, event frequencies,
  AP costs so the game is challenging but winnable.

---

## Scale Estimate

  Packages:           20
  Exported structs:   ~60-80
  Exported functions: ~200-300
  Test functions:     ~400-600
  Lines of Go (excl. tests): ~10,000-15,000
  Lines of tests:     ~5,000-8,000

The config data (stakeholder seeds, policy cards, org definitions, event
definitions) will likely be the single largest block of non-logic code.
Consider embedding as JSON or TOML loaded at startup rather than hardcoded
Go structs, to allow future content editing without recompilation.

---

## Known Risks

1. Simulation correctness: 17-phase pipeline has many interdependencies.
   The headless integration test is the main guard. Run it frequently.

2. Config data volume: 40-60 policy cards, 40-60 events, 30+ stakeholder
   seeds, 18 orgs, 15 companies, 80 tiles. This is a lot of content to
   balance. Design first, tune after Milestone 6.

3. UI performance: 80 tiles recomputing FuelPoverty every week is fine
   for simulation; rendering all tile details in the map is a draw-call
   concern. Use dirty-flag pattern: only re-render tiles that changed.

4. Save format evolution: as design evolves during build, WorldState fields
   will change. The version-tag rejection means saves break on every
   schema change. Acceptable during development; document clearly.

5. Fog-of-war consistency: the revealedAttributes map on each tile must be
   respected everywhere the UI reads tile state. A single missed check is
   a bug that breaks game balance. Consider a typed accessor (TileView)
   that enforces visibility rules at compile time.
