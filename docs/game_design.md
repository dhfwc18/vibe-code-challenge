# Net Zero: Game Design Document

Player role: UK civil servant, 2010-2050, pushing for net zero.
Engine: Ebitengine v2. All values calibrated to HM Treasury Green Book / DESNZ data.

---

## A. Weekly Turn Sequence (14 Phases)

Each game week executes in a fixed pipeline.

1.  Clock Advance -- increment date, check quarter (budget disbursement + tax revenue
    recalculation), check year (carbon budget accounting + climate state update), check
    scheduled election.
2.  Climate and Fossil Dependency Update -- recompute ClimateState from cumulative
    emissions stock. Recompute FossilDependency from active energy policy mix. Both
    feed into event severity multipliers used in Phase 3.
3.  Global Events Roll -- draw from weighted event deck. Event probability and severity
    are multiplied by ClimateState (for weather events) and FossilDependency (for oil/gas
    shocks). Economy is updated by shock severity. At most 2 events per week; most weeks
    zero. Major events (severity > threshold) queue a ShockResponse opportunity for Phase 11.
4.  Scandal and Pressure Roll -- each minister independently rolls scandal probability;
    pressure groups apply weekly modifiers to GovernmentPopularity and LowCarbonReputation
    based on carbon trajectory and recent climate events.
5.  Technology Progress Tick -- each of 8 technologies advances its adoption curve;
    acceleration bonuses from R&D policies apply.
6.  Regional World Tick -- 12 regions update SkillsNetwork, InstallerCapacity, SupplyChain
    based on active policies and organic growth. Climate damage events may degrade regional
    capacity (e.g. flooding reduces retrofit installer capacity in affected region).
7.  Policy Resolution -- active policies resolve weekly carbon impact (multiplied by
    regional capacity and tech maturity), budget draws, popularity modifiers, and
    LowCarbonReputation modifiers. Policies with empty budgets stall.
8.  Carbon Budget Accounting -- accumulate net weekly emissions into current-year total
    and cumulative stock. At year-end: check annual budget limit (overshoot hits
    GovernmentPopularity), update ClimateState from new stock level.
9.  Economy and Tax Revenue Tick -- update hidden Economy from: climate damage this week,
    oil/gas shock severity, active industrial policy bonuses, FossilDependency drag.
    At quarter-end: compute Tax Revenue from Economy, then run Budget Allocation formula
    (Economy * BaseRate * DeptPopularityWeights * LowCarbonReputationModifier).
10. Polling Check -- Bernoulli draw each week (avg interval 8-12 weeks). Generates noisy
    results for GovernmentPopularity (sigma=3) and per-minister (sigma=5). Also generates
    a LowCarbonReputation poll (sigma=4, interval 10-16 weeks).
11. Player Action Phase (interactive) -- player spends AP pool (default 5/week).
    If a ShockResponse was queued in Phase 3, player is offered 2-3 response cards
    before their normal AP spend. Each response card has: popularity gain potential,
    LowCarbonReputation gain potential, backfire risk, and AP cost.
12. Minister Health Check -- WeeksUnderPressure thresholds and IdeologyConflictScore.
13. Minister Transitions -- process queued appointments, sackings, resignations,
    election-outs.
14. Consequence Resolution -- resolve player action outcomes. Budget allocation lobbying
    results take effect at next quarter-end, not immediately.
15. Consultancy Delivery Check -- decrement timers; generate and deliver reports.
16. End of Week Render -- update UI, flush event log.

---

## B. State Machines

### Minister State Machine

States: POOL | APPOINTED | ACTIVE | UNDER_PRESSURE | SACKED | RESIGNED | ELECTION_OUT | DEPARTED

Transitions:
  POOL          -> APPOINTED      : PM decision (election win, reshuffle, vacancy)
  APPOINTED     -> ACTIVE         : 4-week grace period elapses
  ACTIVE        -> UNDER_PRESSURE : MinisterPopularity < SackingThreshold for 1 week
  UNDER_PRESSURE -> ACTIVE        : MinisterPopularity rises above SackingThreshold
  UNDER_PRESSURE -> SACKED        : MinisterPopularity < SackingThreshold for 3 consecutive weeks
  ACTIVE        -> RESIGNED       : IdeologyConflictScore > ResignationThreshold after major policy
  ACTIVE        -> SACKED         : PM discretionary sack (low probability, higher if GovPop < 35)
  ACTIVE        -> ELECTION_OUT   : Election result: party change
  APPOINTED     -> ELECTION_OUT   : Election result: party change
  SACKED        -> DEPARTED       : Immediate
  RESIGNED      -> DEPARTED       : Immediate
  ELECTION_OUT  -> DEPARTED       : Immediate

Sacking threshold: default 25, reduced by up to 10 if GovernmentPopularity > 60.
Cabinet ministers: threshold 20. Junior ministers: threshold 30.
Resignation trigger: (IdeologyDistance * PolicySignificance) > 80.
PolicySignificance: MINOR=10, MODERATE=30, MAJOR=70.
Resignation GovernmentPopularity penalty: -4 to -12 (larger than sacking: -2 to -8).

### Government State Machine

States: GOVERNING | PRE_ELECTION | ELECTION | TRANSITION | SNAP_ELECTION_RISK | DISSOLVED

Transitions:
  GOVERNING          -> PRE_ELECTION       : Scheduled election minus 6 weeks
  GOVERNING          -> SNAP_ELECTION_RISK : GovernmentPopularity < 30 for 4+ weeks
  SNAP_ELECTION_RISK -> GOVERNING          : GovernmentPopularity rises above 30
  SNAP_ELECTION_RISK -> DISSOLVED          : PM snap election decision (weekly probabilistic roll)
  GOVERNING          -> DISSOLVED          : Tactical election call (rare, if GovPop > 55)
  PRE_ELECTION       -> ELECTION           : Campaign period elapses
  DISSOLVED          -> PRE_ELECTION       : Immediate
  ELECTION           -> GOVERNING          : Incumbent wins
  ELECTION           -> TRANSITION         : Incumbent loses
  TRANSITION         -> GOVERNING          : New ministers appointed, 2-week timer elapses

PRE_ELECTION: player AP reduced to 3 (purdah). Policy approvals suspended.
Election outcome: logistic model on GovernmentPopularity + economic indicators + seeded stochastic term.

Scheduled UK elections (approximate, from 2010 game start):
  May 2010, May 2015, June 2017, December 2019, then player-influenced.

---

## C. Resource Table

| Resource                         | Range          | Visible to Player                              |
|----------------------------------|----------------|------------------------------------------------|
| Action Points (AP)               | 0-10/week      | Yes (exact)                                    |
| GovernmentPopularity             | 0-100          | No (poll only, noise sigma=3)                  |
| LastPollResult (Government)      | 0-100          | Yes (snapshot)                                 |
| MinisterPopularity (per)         | 0-100          | No (poll only, noise sigma=5)                  |
| LastPollResult (Minister)        | 0-100          | Yes (snapshot)                                 |
| LowCarbonReputation              | 0-100          | No (poll only, noise sigma=4, interval 10-16w) |
| LastPollResult (LowCarbon)       | 0-100          | Yes (snapshot)                                 |
| Economy                          | 0-100          | No (Tax Revenue is the only lagging signal)    |
| TaxRevenue                       | GBP bn/quarter | Yes (exact, quarterly)                         |
| DepartmentBudget (per dept)      | 0-500M GBP/qtr | Yes (exact)                                    |
| BudgetAllocationLobbyEffect      | modifier       | No (outcome revealed at next quarter-end)      |
| RelationshipScore (per minister) | -100 to +100   | Approximate (5 labels)                         |
| ClimateState                     | 0-100          | No (visible only via climate events + reports) |
| CumulativeEmissionsStock         | MtCO2e         | No (annual reporting only, approximate)        |
| FossilDependency                 | 0-100%         | No (consultancy reveals; inferred from events) |
| CarbonBudgetRunning              | MtCO2e         | Approximate (annual event only)                |
| AnnualCarbonBudgetLimit          | per CCC target | Yes (always)                                   |
| TechMaturity (per tech)          | 0-100          | No (consultancy reveals only)                  |
| RegionalSkillsNetwork            | 0-100          | No (consultancy reveals only)                  |
| RegionalInstallerCapacity        | 0-1000/week    | No (consultancy reveals only)                  |
| RegionalSupplyChain              | 0-100          | No (consultancy reveals only)                  |
| MinisterIdeologyScore            | -100 to +100   | No (inferred from behaviour)                   |
| MinisterNetZeroSympathy          | 0-100          | No                                             |
| MinisterRiskTolerance            | 0-100          | No                                             |
| MinisterPopulismScore            | 0-100          | No                                             |
| WeeksUnderPressure (minister)    | 0+             | No                                             |
| PlayerReputation                 | 0-100          | Yes (5 grade labels)                           |
| CarbonOvershootAccumulator       | 0+             | No (specific consultancy only)                 |

RelationshipScore labels: Hostile / Cool / Neutral / Warm / Ally.
PlayerReputation labels: Generalist / Executive Officer / Higher Executive / Grade 7 / Grade 6 / Deputy Director / Director / Director General / Permanent Secretary.

Player actions and AP costs:
  Meet minister:         1-3 AP (seniority-dependent, relationship threshold may apply)
  Submit policy:         2 AP
  Commission consultancy: 1 AP + budget
  Attend select committee: 2 AP (affects minister relationship)
  Hire civil service staff: 3 AP + budget (increases future AP pool)
  Read delivered report: 0 AP

---

## D. Package Map

internal/
  config/       Static data definitions (minister pools, policy cards, consultancy cards, event
                definitions, tech curves, carbon budget limits, region geometry). Loaded at startup.
                Every other package reads from here. No game logic.

  save/         WorldState serialisation/deserialisation, versioned save format, master seed
                management. All stochastic elements derive sub-seeds deterministically from master.

  carbon/       CarbonBudgetState, annual limit table (CCC targets), weekly accumulation,
                overshoot accounting, cumulative stock tracking, trajectory projection.
                Outputs cumulativeStock used by climate package. Pure functions.

  climate/      ClimateState derivation from cumulative stock. ClimateEvent and
                EnergyShockEvent structs. Event probability/severity computation using
                ClimateState and FossilDependency multipliers. ShockResponseCard definitions
                and backfire probability formula. Does not own the event queue -- hands
                events to the event package. Pure functions where possible.

  technology/   TechTracker, LogisticCurve evaluation, AccelerationBonus accumulation.
                Outputs TechMaturity per technology each tick. Pure functions.

  region/       12-region model. SkillsNetwork/InstallerCapacity/SupplyChain per region.
                CapacityMultiplier output consumed by policy resolution. Climate event
                regional capacity damage applied here. Region map geometry data for rendering.

  economy/      EconomyState (hidden), TaxRevenue (visible), FossilDependency (derived).
                Budget allocation formula: baseFraction * ministerPopWeight * LCRModifier
                * lobbyEffect. Tracks pending lobby effects and clears them at quarter-end.
                Depends on climate (damage inputs) and policy (independence bonus inputs).

  reputation/   LowCarbonReputation value, poll generation, weekly chain-effect computation
                (minister popularity delta, government popularity delta, budget modifier).
                ShockCapitalisation outcome resolution (success/backfire probability from
                LCR and PlayerReputation). Separate from government/minister packages to
                keep chain-effect logic in one place.

  minister/     Minister struct, MinisterStateMachineEvaluator, MinisterPool, MinisterFactory.
                Given minister + world snapshot, returns zero or more transition events.
                Does not mutate state.

  government/   GovernmentStateMachine, ElectionOutcomeResolver, PopularityModifier,
                PopularityHistory (52-week ring buffer). Computes election outcomes from
                logistic model. Tracks GovernmentPopularity and weekly modifiers.

  polling/      Poller, PollResult, PollScheduler, NoiseModel. Never holds true popularity;
                receives it as parameter. Gaussian noise model, configurable sigma.

  policy/       PolicyCard catalogue, ApprovalPipeline (DRAFT->SUBMITTED->UNDER_REVIEW->
                APPROVED|REJECTED->ACTIVE|ARCHIVED), weekly effect resolution.
                Depends on technology (unlock), region (capacity multiplier), carbon (accounting).

  consultancy/  ConsultancyCard deck, Commission management, delivery timer, InsightReport
                generation. Quality rolled at commission time (hidden until delivery).
                BiasModel for ideologically-motivated think tanks.

  event/        GlobalEvent deck (weighted draw), ScandalEvaluator (per minister weekly roll),
                PressureGroup persistent actors, EventLog (player-visible weekly feed).

  player/       CivilServant state, AP pool, StaffRoster, ActionRecord log, minister
                relationships. Passive state container; simulation reads and updates it.

  simulation/   WorldState (single source of truth), TurnEngine (14-phase weekly pipeline),
                EventBus. Imports all domain packages. Only package that mutates WorldState.

  ui/           All Ebitengine scenes and input handling. Reads WorldState, never writes.
                Expresses player intent as Action structs passed to simulation.
                Scenes: WeeklyTurn, Minister, RegionMap, Policy, Consultancy, EventLog.

  game/         Ebitengine bootstrap. Wires simulation and ui. Trivial once above exist.

---

## E. Key Data Structures

WorldState {
  date: { year, week }
  government: GovernmentState
  ministers: map[Department -> Minister]
  ministerPool: map[Party -> map[Department -> []MinisterDef]]
  departmentBudgets: map[Department -> BudgetAccount]
  policies: []PolicyInstance { card, state, weeksActive, budgetDrawn }
  pendingCommissions: []Commission
  deliveredReports: []InsightReport
  regions: [12]Region
  technologies: [8]TechTracker
  carbon: CarbonBudgetState
  player: CivilServant
  pollHistory: []PollResult
  eventLog: []WeeklyEventLog
  governmentPopularity: float64
  lastPollResult: PollResult
  rng: SeededRNG
}

Minister {
  id, name, party, department
  state: MinisterState
  graceWeeksRemaining: int
  // Hidden:
  ideologyScore: float64        // -100 frugal/conservative to +100 spendthrift/progressive
  netZeroSympathy: float64      // 0 hostile to 100 champion
  riskTolerance: float64
  populismScore: float64
  // Tracked:
  popularity: float64
  weeksUnderPressure: int
  ideologyConflictAccumulator: float64
  relationshipWithPlayer: float64
  personalityTraits: []Trait
  portfolioPriorities: []Priority
}

PolicyCard (static, from config) {
  id, name, description
  unlockYear: int
  requiredTechMaturity: map[Technology -> float64]
  apCost: int
  weeklyBudgetDraw: map[Department -> float64]
  weeklyBaselineCarbonImpact: float64   // tonnes CO2e/week at full capacity
  regionalCapacityRequirement: map[SectorSkill -> float64]
  popularityModifier: float64
  ministerPopularityModifier: map[Department -> float64]
  approvalRequirements: []ApprovalRequirement
  ideologyScore: float64
  significance: MINOR | MODERATE | MAJOR
}

ConsultancyCard (static, from config) {
  id, name
  organisationType: PRIVATE_CONSULTANCY | THINK_TANK | ACADEMIC | REGULATOR
  cost: float64
  deliveryWeeksRange: { min, max }
  insightType: InsightType
  insightScope: InsightScope
  biasDirection: float64   // -1 to +1; think tanks have non-zero values
  popularityRisk: float64
  // quality rolled at commission time, not stored here
}

Region {
  id: RegionID
  skillsNetwork: map[SectorSkill -> float64]
  installerCapacity: map[SectorSkill -> float64]
  supplyChain: float64
  revealedByPlayer: map[string -> bool]
}

SectorSkill enum: Retrofit | EVCharging | WindInstallation | HeatPump |
                  HydrogenInfrastructure | SolarInstallation | CCSInstallation

Technology {
  id: TechID
  historicalUnlockYear: int
  adoptionCurve: LogisticCurve { midpointYear, steepness }
  currentMaturity: float64
  accelerationBonusWeeks: float64
  isUnlocked: bool
}

TechID enum: OffshoreWind | OnshoreWind | Solar | EVs | HeatPumps | Hydrogen | CCUS | DACCS

CarbonBudgetState {
  annualLimits: map[int -> float64]       // year -> MtCO2e limit
  weeklyEmissionsHistory: []float64
  currentYearAccumulation: float64
  totalOvershootAccumulator: float64
  cumulativeStock: float64               // total CO2e emitted since 2010; drives ClimateState
  baselineWeeklyEmissions: float64
  trajectoryProjection: []float64         // next 5 years
}

ClimateState {
  // Derived from cumulativeStock each year-end. Not stored separately -- recomputed.
  // Drives: climate event probability weights, event severity multipliers.
  // Thresholds (illustrative, subject to tuning):
  //   stock < 5,000 MtCO2e  -> LOW    (baseline weather, no amplification)
  //   stock < 10,000         -> MEDIUM (cold snaps more frequent, flooding +20% prob)
  //   stock < 15,000         -> HIGH   (heatwaves, coastal flooding, agriculture stress)
  //   stock >= 15,000        -> SEVERE (extreme events quarterly, economy drag ongoing)
  level: ClimateLevel  // LOW | MEDIUM | HIGH | SEVERE
  eventProbabilityMultiplier: float64
  eventSeverityMultiplier: float64
  ongoingDamagePerWeek: float64          // permanent economy drag at HIGH/SEVERE
}

ClimateEvent {
  type: ClimateEventType  // ColdSnap | Flooding | Heatwave | Drought | StormDamage
  affectedRegions: []RegionID
  economySeverity: float64               // raw hit to Economy before FossilDependency scaling
  regionalCapacityDamage: map[RegionID -> map[SectorSkill -> float64]]
  govPopularityDelta: float64
  lowCarbonRepDelta: float64             // sign depends on ClimateState vs. player narrative
  shockResponseQueued: bool
}

EnergyShockEvent {
  type: EnergyShockType  // OilPriceSpike | GasSupplyDisruption | BlackoutRisk
  baseSeverity: float64                  // before FossilDependency multiplier
  actualSeverity: float64               // baseSeverity * (FossilDependency / 50)
  economyDelta: float64                  // negative; scales with actualSeverity
  deptBudgetDelta: float64              // Treasury forced to bail out energy costs
  shockResponseQueued: bool
}

EconomyState {
  value: float64                         // 0-100, hidden from player
  baseRate: float64                      // GBP bn/quarter tax revenue at value=100
  // Modifiers applied weekly:
  //   climate damage: -ongoingDamagePerWeek
  //   oil/gas shock: -actualSeverity * shockEconomyWeight
  //   clean energy independence bonus: +(100 - FossilDependency) * independenceRate
  //   active industrial policy: +policyBonus per active policy
  weeklyDeltaHistory: []float64
}

TaxRevenue {
  lastQuarterValue: float64             // GBP bn, visible to player
  allocationWeights: map[Department -> float64]  // recomputed each quarter
  // Allocation formula:
  //   baseFraction * ministerPopularityWeight * lowCarbonRepModifier * lobbyEffect
  // lobbyEffect: accumulated from player lobbying actions since last quarter-end
  pendingLobbyEffect: map[Department -> float64]  // cleared each quarter-end
}

FossilDependency {
  // Derived from active energy policy mix each week. Not separately stored.
  // = (baseline fossil MWh - clean policy MWh saved) / baseline fossil MWh * 100
  // Range: 0-100%. High value amplifies EnergyShockEvent severity.
  currentPercent: float64
}

LowCarbonReputation {
  value: float64                        // 0-100, true value, hidden from player
  lastPollResult: PollResult            // noisy, sigma=4, visible to player
  // Increases from:
  //   policy success attributed to net zero agenda
  //   climate event + player capitalises successfully
  //   international agreements signed
  //   favourable think tank report (if player commissions and publishes)
  // Decreases from:
  //   policy failure attributed to net zero agenda
  //   energy price spike while player is seen as pro-renewables
  //   climate event + player capitalises and it backfires
  //   hostile think tank report (if opponent commissions)
  // Chain effects applied each week:
  //   minister popularity delta += netZeroSympathy * (LCR - 50) * 0.01
  //   government popularity delta += (LCR - 50) * 0.005
  //   budget allocation modifier = 1.0 + (LCR - 50) * 0.004  (for green depts)
  weeklyDeltaHistory: []float64
}

ShockResponseCard {
  id: string
  name: string
  description: string
  apCost: int
  // Outcomes are probabilistic, weighted by LowCarbonReputation and PlayerReputation:
  successGovPopDelta: float64
  successLCRDelta: float64
  backfireGovPopDelta: float64
  backfireLCRDelta: float64
  // Backfire probability formula:
  //   max(0, 0.5 - (LCR - 50)*0.01 - (PlayerReputation - 50)*0.005)
  // i.e. high LCR and high PlayerReputation make backfire unlikely but not impossible
  budgetCost: float64
  availableForEventTypes: []EventType
}

---

## F. Implementation Order

Layer 0 (no game dependencies):    config, save
Layer 1 (pure domain models):      carbon, technology, region
Layer 2 (derived world models):    climate, economy, reputation
Layer 3 (agent models):            minister, government, polling
Layer 4 (player-facing mechanics): policy, consultancy, event, player
Layer 5 (orchestration):           simulation
Layer 6 (presentation):            ui
Layer 7 (entry point):             game

Key dependencies added by new systems:
  climate   <- carbon (cumulativeStock), event (shock queue)
  economy   <- climate (damage), policy (fossil dependency, industrial bonus)
  reputation <- government (popularity chain), minister (chain), economy (budget modifier)
  simulation <- all of the above, in 16-phase pipeline (was 14)

Build each layer with unit tests before moving to the next.
Headless simulation test: advance 100 weeks, verify no negative budgets,
no invalid minister states, carbon accumulation bounded.

---

## G. Open Design Questions

1. Minister legibility: how much of a minister's ideology is inferable from party, name,
   prior public record (free to read) vs. requiring player to probe via meetings?

2. Election stochasticity: should the seeded stochastic term be derived from the last 10
   player actions (making different strategies produce different outcomes), or purely random?

3. Carbon sector granularity: track 6 separate emission sectors (power, transport, buildings,
   industry, agriculture, land) or a single aggregated carbon number?

4. Consultancy cross-reference: can players directly compare two reports on the same topic
   to detect low-quality or biased output?

5. Regional partial transparency: should some regional information be freely visible from
   public data (qualitative level), with precise numbers requiring consultancy?

6. Minister AP costs: should AP cost to meet a minister be dynamic (higher for hostile ones)?

7. Average ministerial tenure target: real UK data ~2 years (104 weeks). Right for gameplay?

8. UI library: adopt ebitenui or build a bespoke minimal widget system?

9. Save file migration: refuse to load older saves, or write version migration functions?

10. Minister distinctiveness: procedurally generated observable personality signals
    (catchphrases, known public positions) to help players infer hidden attributes?

11. Climate event tone: when a climate event occurs (flooding, cold snap), should the
    LowCarbonReputation delta be automatically positive (public blames fossil fuels) or
    should it depend on whether the player successfully capitalises? I.e. is the event
    itself neutral and only the player response determines the reputation outcome?

12. Fossil dependency visibility: should FossilDependency be directly readable from a
    public statistics page (realistic -- DESNZ publishes this), or kept behind consultancy
    fog of war to increase strategic uncertainty?

13. Tax revenue granularity: should the player see a breakdown of tax revenue by source
    (income tax, fuel duty, VAT, business rates) or just a single GBP figure? A breakdown
    would make the oil shock -> fuel duty loss link tangible and educational.

14. Shock response timing: should ShockResponseCards expire after 1 week (use them or
    lose them) or persist until the player acts (simpler but less tense)?
