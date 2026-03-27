# Net Zero: Game Design Document

Player role: UK civil servant, 2010-2050, pushing for net zero.
Engine: Ebitengine v2. All values calibrated to HM Treasury Green Book / DESNZ data.

---

## A. Weekly Turn Sequence (14 Phases)

Each game week executes in a fixed pipeline.

1. Clock Advance -- increment date, check quarter (budget disbursement), check year (carbon accounting), check scheduled election.
2. Global Events Roll -- draw from weighted event deck (oil shock, international agreement, recession, extreme weather). At most 2 per week; most weeks zero.
3. Scandal and Pressure Roll -- each minister independently rolls scandal probability; pressure groups apply weekly GovernmentPopularity modifiers based on carbon trajectory.
4. Technology Progress Tick -- each of 8 technologies advances its adoption curve; acceleration bonuses from R&D policies apply.
5. Regional World Tick -- 12 regions update SkillsNetwork, InstallerCapacity, SupplyChain based on active policies and organic growth.
6. Policy Resolution -- all active policies resolve weekly carbon impact (multiplied by regional capacity and tech maturity), budget draws, popularity modifiers. Policies with empty budgets stall.
7. Carbon Budget Accounting -- accumulate net weekly emissions. At year-end, check against annual carbon budget limit. Overshoot hits GovernmentPopularity.
8. Polling Check -- Bernoulli draw each week (avg interval 8-12 weeks). On fire: generate noisy poll result (Gaussian sigma=3 for government, sigma=5 for ministers). Store as LastPollResult.
9. Minister Health Check -- check WeeksUnderPressure thresholds and IdeologyConflictScore for resignation triggers.
10. Minister Transitions -- process queued appointments, sackings, resignations, election-outs.
11. Player Action Phase (interactive) -- player spends AP pool (default 5/week) on actions listed below.
12. Consequence Resolution -- resolve outcomes of player actions: meeting results, policy pipeline advancement, budget changes.
13. Consultancy Delivery Check -- decrement delivery timers; generate and deliver reports at zero.
14. End of Week Render -- update UI, flush event log.

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

| Resource                         | Range          | Visible to Player                   |
|----------------------------------|----------------|-------------------------------------|
| Action Points (AP)               | 0-10/week      | Yes (exact)                         |
| GovernmentPopularity             | 0-100          | No (poll only, noise sigma=3)       |
| LastPollResult (Government)      | 0-100          | Yes (snapshot)                      |
| MinisterPopularity (per)         | 0-100          | No (poll only, noise sigma=5)       |
| LastPollResult (Minister)        | 0-100          | Yes (snapshot)                      |
| DepartmentBudget (per dept)      | 0-500M GBP/qtr | Yes (exact)                         |
| RelationshipScore (per minister) | -100 to +100   | Approximate (5 labels)              |
| CarbonBudgetRunning              | unbounded      | Approximate (annual event only)     |
| AnnualCarbonBudgetLimit          | per CCC target | Yes (always)                        |
| TechMaturity (per tech)          | 0-100          | No (consultancy reveals only)       |
| RegionalSkillsNetwork            | 0-100          | No (consultancy reveals only)       |
| RegionalInstallerCapacity        | 0-1000/week    | No (consultancy reveals only)       |
| RegionalSupplyChain              | 0-100          | No (consultancy reveals only)       |
| MinisterIdeologyScore            | -100 to +100   | No (inferred from behaviour)        |
| MinisterNetZeroSympathy          | 0-100          | No                                  |
| MinisterRiskTolerance            | 0-100          | No                                  |
| MinisterPopulismScore            | 0-100          | No                                  |
| WeeksUnderPressure (minister)    | 0+             | No                                  |
| PlayerReputation                 | 0-100          | Yes (5 grade labels)                |
| CarbonOvershootAccumulator       | 0+             | No (specific consultancy only)      |

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
                overshoot accounting, trajectory projection. Pure functions, no side effects.

  technology/   TechTracker, LogisticCurve evaluation, AccelerationBonus accumulation.
                Outputs TechMaturity per technology each tick. Pure functions.

  region/       12-region model. SkillsNetwork/InstallerCapacity/SupplyChain per region.
                CapacityMultiplier output consumed by policy resolution. Region map geometry
                data for rendering.

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
  baselineWeeklyEmissions: float64
  trajectoryProjection: []float64         // next 5 years
}

---

## F. Implementation Order

Layer 0 (no game dependencies):   config, save
Layer 1 (pure domain models):     carbon, technology, region
Layer 2 (agent models):           minister, government, polling
Layer 3 (player-facing mechanics): policy, consultancy, event, player
Layer 4 (orchestration):          simulation
Layer 5 (presentation):           ui
Layer 6 (entry point):            game

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
