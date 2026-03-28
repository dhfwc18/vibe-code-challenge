# 20-50: Game Design Document

Player role: Taitan civil servant, 2010-2050, pushing for net zero.
Engine: Ebitengine v2. All values calibrated to HM Treasury Green Book / DESNZ data.

---

## A. Weekly Turn Sequence (18 Phases)

Each game week executes in a fixed pipeline.

1.  Clock Advance -- increment date, check quarter (budget disbursement + tax revenue
    recalculation), check year (carbon budget accounting + climate state update), check
    scheduled election.
2.  Climate and Fossil Dependency Update -- recompute ClimateState from cumulative
    emissions stock. Recompute FossilDependency from active energy policy mix. Both
    feed into event severity multipliers used in Phase 3. Advance EnergyMarket ring
    buffers (GasHistory, ElecHistory, OilHistory) by one week so the 52-week chart
    reflects current prices.
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
    Each active LCT company ticks its AccumulatedQuality proportional to WorkRate and
    average InstallerCapacity.
7.  Tile Local Tick -- each tile (~60-80 total across 12 regions) updates local state:
    a. TrueRetrofitRate = ObservedRetrofitRate * (InstallerQuality / 100).
    b. FuelPoverty recomputed from EnergyMarket prices, tile HeatingType, InsulationLevel,
       HeatingCapacity, and LocalIncome using the full formula in the data structures
       section. FuelPoverty is not a delta -- it is recomputed from scratch each week
       so it responds immediately to price changes. This means a gas price spike produces
       an immediate FuelPoverty jump in gas-heated low-insulation tiles. A tile that has
       already switched to heat pumps is shielded from gas price spikes but exposed to
       electricity price spikes instead. A well-insulated heat-pump tile on a high-
       renewable grid is the most resilient combination.
    c. LocalPoliticalOpinion shifts from FuelPoverty delta and any climate event impact.
    d. InstallerQuality drifts upward slowly only while an installation standards policy
       is active. New work uses improved quality; prior retrofits are not retroactively
       upgraded -- the gap in existing stock persists until replaced.
8.  Climate Event Impact on Tiles -- resolve per-tile impact of any event from Phase 3:
    Cold snap: tiles where TrueRetrofitRate < threshold suffer heating failures.
      FuelPoverty spikes by (threshold - TrueRetrofitRate) * severityMultiplier.
      Event log shows "X% of retrofitted homes reported heating failures" -- the first
      visible signal of the quality gap without a formal audit.
    Flooding: local FuelPoverty spikes if heating systems physically damaged.
    Heatwave: tiles with low InsulationLevel suffer cooling cost spikes; if heating
      capacity is gas-based, summer energy demand inverts and costs rise.
    Gas price shock: tiles with high FossilDependency and low InsulationLevel suffer
      the largest FuelPoverty spike; tiles already on heat pumps are shielded.
9.  Policy Resolution -- active policies resolve weekly carbon impact using TrueRetrofitRate
    (not ObservedRetrofitRate), regional capacity, and tech maturity. The player may
    observe less carbon reduction than expected if installer quality is poor but receives
    no explicit explanation until a quality audit consultancy report is delivered.
10. Carbon Budget Accounting -- accumulate net weekly emissions (using TrueRetrofitRate-
    adjusted policy impacts) into current-year total and cumulative stock. At year-end:
    check annual budget limit, update ClimateState from new stock level.
11. Economy and Tax Revenue Tick -- update hidden Economy from: climate damage, tile-level
    FuelPoverty aggregate (high national FuelPoverty drags consumer spending -> Economy),
    oil/gas shock severity, active industrial policy bonuses, FossilDependency drag.
    At quarter-end: compute Tax Revenue from Economy, run Budget Allocation formula.
12. Polling Check -- Bernoulli draw each week (p=0.25). When a poll fires: generate
    noisy party vote shares per region and nationally; add government approval rating
    (GovernmentPopularity + sigma=3 noise) to PollSnapshot.GovernmentApprovalRating;
    generate per-minister popularity readings (MinisterPopularity + sigma=5 noise) into
    WorldState.MinisterLastPollResults; run SwingFromLast against the previous snapshot.
    LowCarbonReputation polled on a separate randomised 10-16 week interval (sigma=4).
    FuelPoverty has no poll -- it surfaces only via consultancy or event log signals.
13. Policy Approval Evaluation -- evaluate outstanding approval steps for all
    UNDER_REVIEW policy cards. Cards whose steps all clear transition to ACTIVE.
    Note: runs before Phase 14 so policies submitted this tick are not evaluated
    until the following week (prevents same-tick submission-and-rejection).
14. Player Action Phase (interactive) -- player spends AP pool (default 5/week).
    Available actions: SubmitPolicy (with tech-unlock gate check), CommissionReport,
    LobbyMinister, HireStaff, FireStaff, ShockResponse. ShockResponse cards resolve
    using BackfireProbability derived from current LCR and player Reputation.
15. Minister Health Check -- update minister Popularity via LCR chain and MinisterStats
    each week (happens in Phase 11); check per-minister popularity against sacking
    threshold; trigger ACTIVE->UNDER_PRESSURE or recovery UNDER_PRESSURE->ACTIVE.
    Passive relationship decay applied every week. WeeksUnderPressure and
    IdeologyConflictScore also ticked here.
16. Minister Transitions -- process UNDER_PRESSURE sacking (3 consecutive weeks),
    APPOINTED->ACTIVE grace period, election-triggered APPOINTED/OPPOSITION_SHADOW
    assignments. When a minister is sacked, their cabinet role is vacated via
    government.RemoveMinister.
17. Consequence Resolution -- post-action effects: installation standards policies
    begin propagating InstallerQuality improvement from this week forward (deferred
    to simulation tuning pass).
17. Consultancy Delivery Check -- decrement timers; generate and deliver reports.
    Quality audit reports reveal TrueRetrofitRate vs ObservedRetrofitRate gap for
    specified tiles. Fuel poverty studies reveal FuelPoverty per tile in scope.
18. End of Week Render -- update UI, flush event log.

---

## B. State Machines

### Minister State Machine

States: APPOINTED | ACTIVE | UNDER_PRESSURE | LEADERSHIP_CHALLENGE | SACKED | RESIGNED |
        ELECTION_OUT | BACKBENCH | OPPOSITION_SHADOW | DEPARTED

Note: stakeholders not yet in play are tracked by IsUnlocked=false on the Stakeholder struct,
not by a POOL state. A figure enters play when IsUnlocked=true and assigned a role.

Transitions:
  (unlocked)        -> APPOINTED           : PM decision (election win, reshuffle, vacancy)
  APPOINTED         -> ACTIVE              : 4-week grace period elapses
  ACTIVE            -> UNDER_PRESSURE      : MinisterPopularity < SackingThreshold for 1 week
  UNDER_PRESSURE    -> ACTIVE              : MinisterPopularity rises above SackingThreshold
  UNDER_PRESSURE    -> LEADERSHIP_CHALLENGE: Significant internal party opposition event fires
  UNDER_PRESSURE    -> SACKED              : MinisterPopularity < SackingThreshold for 3 consecutive weeks
  LEADERSHIP_CHALLENGE -> ACTIVE           : Challenge fails; minister survives
  LEADERSHIP_CHALLENGE -> SACKED           : Challenge succeeds
  ACTIVE            -> RESIGNED            : IdeologyConflictScore > ResignationThreshold after major policy
  ACTIVE            -> SACKED              : PM discretionary sack (low probability, higher if GovPop < 35)
  ACTIVE            -> ELECTION_OUT        : Election result: party change
  APPOINTED         -> ELECTION_OUT        : Election result: party change
  SACKED            -> BACKBENCH           : Ruling party minister lands on own backbenches
  RESIGNED          -> BACKBENCH           : Ruling party minister lands on own backbenches
  ELECTION_OUT      -> DEPARTED            : Lost seat entirely; exits pool
  BACKBENCH         -> OPPOSITION_SHADOW   : Party loses election; backbencher becomes shadow minister
  BACKBENCH         -> DEPARTED            : Announces retirement or deselected
  OPPOSITION_SHADOW -> APPOINTED           : Party wins next election

Sacking threshold: default 25, reduced to 20 if GovernmentPopularity > 60 (PM shielded).
Cabinet ministers: threshold 20. Junior ministers: threshold 30.
Resignation trigger: (IdeologyDistance * PolicySignificance) > 8.0.
PolicySignificance: MINOR=10, MODERATE=30, MAJOR=70.
Resignation GovernmentPopularity penalty: -4 to -12 (larger than sacking: -2 to -8).

Implementation notes:
- MinisterPopularity is stored per-minister as Stakeholder.Popularity (float64, 0-100,
  starts at 50). It is distinct from WorldState.GovernmentPopularity.
- Weekly update: reputation.ChainToMinisterPopularity(lcrDelta) + government.ComputeMinisterStats
  popularity modifier applied each Phase 11 for cabinet ministers only.
- Player-visible polls stored in WorldState.MinisterLastPollResults map[stakeholderID]float64.
- APPOINTED->ACTIVE grace period (4 weeks per design) is currently immediate in
  Phase 16; the counter will be wired when the 4-week immunity window is implemented.

### Government State Machine

States: STABLE | UNDER_PRESSURE | CONFIDENCE_VOTE | SNAP_ELECTION | ELECTION_CAMPAIGN |
        PRE_ELECTION | ELECTION | CARETAKER | DISSOLVED

Transitions:
  STABLE            -> PRE_ELECTION       : Scheduled election minus 8 weeks
  STABLE            -> UNDER_PRESSURE     : GovernmentPopularity < 30 for 4+ weeks
  UNDER_PRESSURE    -> STABLE             : GovernmentPopularity rises above 30
  UNDER_PRESSURE    -> CONFIDENCE_VOTE    : Internal party motion filed
  CONFIDENCE_VOTE   -> STABLE             : PM wins vote; confidence restored
  CONFIDENCE_VOTE   -> SNAP_ELECTION      : PM loses vote; calls election
  STABLE            -> SNAP_ELECTION      : PM tactical election call (rare, if GovPop > 55)
  SNAP_ELECTION     -> ELECTION_CAMPAIGN  : Immediate
  PRE_ELECTION      -> ELECTION_CAMPAIGN  : Scheduled campaign period begins
  ELECTION_CAMPAIGN -> ELECTION           : Campaign period elapses (4 weeks)
  ELECTION          -> STABLE             : Incumbent wins
  ELECTION          -> CARETAKER          : Incumbent loses; caretaker period begins
  CARETAKER         -> STABLE             : New ministers appointed, transition elapses
  DISSOLVED         -> PRE_ELECTION       : Parliament dissolved ahead of scheduled date

PRE_ELECTION / ELECTION_CAMPAIGN: player AP reduced to 3 (purdah). Policy approvals suspended.
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
| GasPrice                         | GBP/MWh        | Yes (exact, updated weekly)                    |
| ElectricityPrice                 | GBP/MWh        | Yes (exact, updated weekly)                    |
| OilPrice                         | GBP/barrel     | Yes (exact, updated weekly)                    |
| RenewableGridShare               | 0-100%         | Yes (published statistic, updated quarterly)   |
| InsulationLevel (per tile)       | 0-100          | No (consultancy reveals only)                  |
| HeatingCapacity (per tile)       | 0-100          | No (consultancy reveals only)                  |
| HeatingType (per tile)           | enum           | No (energy survey consultancy reveals)         |
| InstallerQuality (per tile)      | 0-100          | No (quality audit consultancy only)            |
| ObservedRetrofitRate (per tile)  | 0-100%         | Yes (reported by contractors in event log)     |
| TrueRetrofitRate (per tile)      | 0-100%         | No (derived; quality audit reveals gap)        |
| FuelPoverty (per tile)           | 0-100%         | No (fuel poverty study consultancy only)       |
| LocalIncome (per tile)           | 0-100          | No (static seed; socioeconomic study reveals)  |
| LocalPoliticalOpinion (per tile) | 0-100          | No (feeds regional polling signal with noise)  |

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

  stakeholder/  27 seeded political figures: 16 active cabinet-eligible (4 parties x 4 roles)
                plus 11 successors and mid/late entrants. Stakeholder struct shared by all.
                MinisterState machine (APPOINTED|ACTIVE|UNDER_PRESSURE|LEADERSHIP_CHALLENGE|
                SACKED|RESIGNED|ELECTION_OUT|BACKBENCH|OPPOSITION_SHADOW|DEPARTED).
                RelationshipScore, InfluenceScore, SpecialMechanic (TICKY/DIZZY/FATIGUE).
                ApprovalChance and PolicyIdeologyPosition exported for policy package use.

  industry/     LCT company roster (15 base companies + emergent). Company struct with
                static seed and dynamic state. CompanyStateMachine. Weekly background work
                computation (R&D -> TechMaturity delta, deployment -> InstallerCapacity delta).
                Player intervention handlers (contract, grant, standard, investigation).
                Company lobbying event generation. Shown in Industry tab in UI.

  energy/       EnergyMarket struct. GasPrice, ElectricityPrice, OilPrice, RenewableGridShare.
                Weekly price update functions applying global market deltas, carbon levy,
                renewable subsidy, and grid coupling model. Price history ring buffers for
                chart display. Historical price anchors seeded from real DESNZ/Ofgem data.
                Exports current prices to tile FuelPoverty computation each week.

  climate/      ClimateState derivation from cumulative stock. ClimateEvent and
                EnergyShockEvent structs. Event probability/severity computation using
                ClimateState and FossilDependency multipliers. ShockResponseCard definitions
                and backfire probability formula. Does not own the event queue -- hands
                events to the event package. Pure functions where possible.

  technology/   TechTracker, LogisticCurve evaluation, AccelerationBonus accumulation.
                Outputs TechMaturity per technology each tick. Pure functions.

  region/       12-region model and 60-80 tile sub-model. Region owns SkillsNetwork,
                InstallerCapacity, SupplyChain. Tile owns InsulationLevel, HeatingCapacity,
                InstallerQuality, ObservedRetrofitRate, TrueRetrofitRate, FuelPoverty,
                LocalIncome, LocalPoliticalOpinion. Computes CapacityMultiplier for policy
                resolution and TileImpact for climate events. All tile state is hidden by
                default; revealedAttributes map tracks what consultancy has uncovered.
                Region map geometry and tile boundary data for rendering.

  economy/      EconomyState (hidden), TaxRevenue (visible), FossilDependency (derived).
                Budget allocation formula: baseFraction * ministerPopWeight * LCRModifier
                * lobbyEffect. Tracks pending lobby effects and clears them at quarter-end.
                Depends on climate (damage inputs) and policy (independence bonus inputs).

  reputation/   LowCarbonReputation value, poll generation, weekly chain-effect computation
                (minister popularity delta, government popularity delta, budget modifier).
                ShockCapitalisation outcome resolution (success/backfire probability from
                LCR and PlayerReputation). Separate from government/minister packages to
                keep chain-effect logic in one place.

  government/   GovernmentStateMachine, ElectionOutcomeResolver, PopularityModifier,
                PopularityHistory (52-week ring buffer). Computes election outcomes from
                logistic model. Tracks GovernmentPopularity and weekly modifiers.

  polling/      Poller, PollResult, PollScheduler, NoiseModel. Never holds true popularity;
                receives it as parameter. Gaussian noise model, configurable sigma.

  policy/       PolicyCard catalogue, ApprovalPipeline (DRAFT->UNDER_REVIEW->
                APPROVED|REJECTED->ACTIVE|ARCHIVED), weekly effect resolution.
                Depends on technology (unlock), region (capacity multiplier), carbon (accounting).

  evidence/     Replaces "consultancy" package. Manages all three advisory organisation types:
                Consultancy, ThinkTank, Academic. Each type has distinct cost, delivery speed,
                quality distribution, and bias model. OrganisationRoster (18 named orgs:
                15 Local + 3 Murican).
                Commission management and delivery timer. InsightReport generation with
                quality-adjusted and bias-distorted output. Cross-reference display (player
                can view multiple reports on same topic side by side). Relationship score
                per organisation (repeat commissions and acting on findings builds trust,
                which slightly improves future quality). Shown in "Evidence" tab in UI.

  event/        GlobalEvent deck (weighted draw), ScandalEvaluator (per minister weekly roll),
                PressureGroup persistent actors, EventLog (player-visible weekly feed).
                Data-driven: all event definitions live in config seed data. The event package
                is a pure dispatcher -- no per-event hardcoded logic.
                ResolvedEffect pattern: ResolveEffect(EventEffect, world snapshots) returns a
                flat map of concrete changes for each target (region, tile, stakeholder, company).
                Targeting is driven by filter strings on EventEffect:
                  RegionFilter    -- "COASTAL", "RURAL", "URBAN", "INDUSTRIAL", "AGRICULTURAL",
                                     or a specific region ID. Empty = all regions.
                  StakeholderFilter -- "CABINET", "ROLE:ENERGY", etc. Empty = no stakeholder effect.
                  CompanyFilter   -- "ALL", "TECH:EVS", "TECH:OFFSHORE_WIND", etc. Empty = no effect.
                MatchRegions / MatchStakeholders / MatchCompanies expand filter strings to lists.
                New events can be added by extending config/events.go only -- no code changes needed.

  player/       CivilServant state, AP pool, StaffRoster, ActionRecord log, minister
                relationships. Passive state container; simulation reads and updates it.

  simulation/   WorldState (single source of truth), TurnEngine (18-phase weekly pipeline),
                EventBus. Imports all domain packages. Only package that mutates WorldState.

  ui/           All Ebitengine scenes and input handling. Reads WorldState, never writes.
                Expresses player intent as Action structs passed to simulation.
                Tab structure:
                  Overview    -- weekly summary, event log, AP spend, clock
                  Map         -- regional choropleth, tile drill-down, fog-of-war overlay
                  Politics    -- all 16 stakeholders by party/role, PM status, relationship scores
                  Policy      -- policy card browser, approval pipeline, active policies
                  Energy      -- energy price dashboard, renewable grid share, price history charts
                  Industry    -- LCT company roster, company state, active contracts, intervention actions
                  Evidence    -- advisory organisation roster, active commissions, report inbox,
                                 cross-reference view (group reports by topicKey)
                  Budget      -- department budgets, tax revenue, lobbying effect tracker

  game/         Ebitengine bootstrap. Wires simulation and ui. Trivial once above exist.

---

## E. Key Data Structures

WorldState {
  date: { year, week, quarter }
  government: GovernmentState           // ruling party, cabinet map (role->stakeholderID), election week
  stakeholders: []Stakeholder           // all political figures (governing + opposition + pool)
  policyCards: []PolicyCard             // all cards with live state (DRAFT/UNDER_REVIEW/ACTIVE/etc.)
  commissions: []Commission             // active evidence commissions in flight
  orgStates: []OrgState                 // per-org dynamic state (relationship, cooling-off, etc.)
  deliveredReports: []InsightReport
  regions: []Region                     // 12 regions with installer capacity and skills network
  tiles: []Tile                         // 30 tiles with local political opinion, fuel poverty, etc.
  tech: TechState                       // 8 technology curves with current maturity values
  carbon: CarbonBudgetState
  energyMarket: EnergyMarket
  climateState: ClimateState
  economy: EconomyState
  lcr: LCRState                         // Low Carbon Reputation true value + poll history
  industry: IndustryState               // company roster with live states
  player: PlayerState                   // AP pool, staff, action history
  pollSnapshots: []PollSnapshot
  eventLog: []EventEntry
  pendingShockResponses: []PendingShockResponse
  governmentPopularity: float64         // hidden true value; polled with sigma=3
  governmentLastPollResult: float64     // most recent noisy sample
  ministerLastPollResults: map[StakeholderID -> float64]  // per-minister noisy popularity
  techDeliveryLog: []string             // milestone messages when a company delivers a tech boost
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

OrgType enum: Consultancy | ThinkTank | Academic
OrgOrigin enum: Local | Murican | Foreign

Organisation (static definition in config) {
  id, name
  orgType: OrgType
  origin: OrgOrigin           // Local = Taitan-based; Murican = Murica-based; Foreign = other

  // Cost and delivery:
  baseCost: float64
  deliveryDistribution: TriangularDist { min: int, mode: int, max: int }  // weeks
  // Consultancy: min=2, mode=3, max=9  (fast-medium, right-skewed tail)
  // ThinkTank:   min=3, mode=5, max=10
  // Academic:    min=6, mode=9, max=16

  // Quality:
  qualityRange: { min: float64, max: float64 }
  // Consultancy: 30-90  (variable -- some are excellent, some are dreadful)
  // ThinkTank:   40-75  (ideological framing limits ceiling)
  // Academic:    60-95  (peer-review process raises floor)

  // Bias:
  biasType: BiasType                // ClientConfirmation | Ideological | None
  biasDirection: float64            // used only when biasType=Ideological (-1 to +1)
  clientBiasWeight: float64         // used only when biasType=ClientConfirmation (0-1)
  // Consultancy: biasType=ClientConfirmation, clientBiasWeight=0.4
  // ThinkTank:   biasType=Ideological, biasDirection per org
  // Academic:    biasType=None (small random noise only)

  // Popularity risk:
  popularityRisk: float64
  // Consultancy: 0.4  (medium -- "wasting public money on consultants" narrative)
  // ThinkTank:   0.1-0.8  (depends on how controversial the org is)
  // Academic:    0.05  (low -- hard to criticise peer-reviewed research)

  // Failure:
  baseFailureProbability: float64
  // Consultancy: 0.08-0.12  (higher for boutique orgs)
  // ThinkTank:   0.04
  // Academic:    0.03  (rare but delivery slips are common -- handled via wide distribution)

  // Specialisms:
  specialisms: []InsightType        // outside specialism = quality -20

  // Relationship (dynamic, per playthrough -- not stored in static config):
  //   relationshipScore: float64
  //   qualityBonus = max(0, (relationshipScore - 30) * 0.002)  // kicks in above 30
  //   clientBiasWeight reduction = max(0, (relationshipScore - 70) * 0.005)  // long-term trust
  //   state: AVAILABLE | COOLING_OFF | STRUGGLING | DEPARTED
}

// Consultancy bias model (distinct from think tank ideological bias):
//
//   Consultancies apply CLIENT CONFIRMATION BIAS, not ideological bias.
//   biasDirection is not a fixed value -- it is computed at report generation time:
//
//     clientBias = dot product of (player's recent major decisions) and (report topic)
//     reportedValue = rawValue + (rawValue - neutralValue) * clientBias * clientBiasWeight
//
//   Effect: if the player has been aggressively pushing retrofit policy, a consultancy
//   commissioned to assess retrofit readiness will report a more optimistic picture than
//   the true value. The more the player has invested in a direction, the more a
//   consultancy will validate that direction.
//
//   This is invisible to the player. There is no "bias indicator" for consultancies.
//   The only ways to detect it:
//     a) Cross-reference with an academic or think tank report on the same topic.
//     b) A climate/policy event exposes the gap between reported and true state.
//     c) Building a very high relationship score (>= 70) partially suppresses client bias
//        -- a long-term trusted adviser gives more honest assessments.
//
//   Popularity risk for consultancies is MEDIUM (not low). Commissioning expensive
//   private consultancies is politically visible and can draw criticism from opposition
//   ("wasting public money on consultants"). Risk scales with contract value.
//
// Delivery failure mechanic:
//   Each consultancy commission has a failureProbability (default 8%, higher for
//   complex scope or low relationship). If it fires:
//     - Full budget refunded.
//     - RelationshipScore with that org -15.
//     - Organisation enters COOLING_OFF state for 8 weeks (cannot be recommissioned).
//     - Player receives a failure notice in the Evidence inbox.
//   Player can also manually FIRE an organisation mid-commission:
//     - 50% budget refunded (kill fee).
//     - RelationshipScore -10.
//     - No COOLING_OFF penalty.
//   Organisations with repeated failures (2+) move to STRUGGLING state and eventually
//   exit the roster (merged or dissolved), similar to the industry company mechanic.
//
// Delivery time is variable: base range is fast-to-medium (2-6 weeks), but each
// commission draws from a distribution so actual delivery varies:
//   deliveryWeeks = sample(triangular distribution, min=baseMin, mode=baseMode, max=baseMax*1.5)
//   A commission scoped to multiple regions or insight types uses baseMax*2.
//
// Organisation roster (18 orgs -- 15 Local, 3 Murican):
//
// ----- LOCAL (Taitan-based) -----
//
// Consultancies [origin: Local]:
//   Tacute Energy Solutions    -- energy sector specialist, very high cost, medium (3-5w),
//                                  quality 70-95 (high floor and ceiling -- best in band),
//                                  popularityRisk 0.2 (lowest in consultancy band),
//                                  clientBiasWeight 0.25 (lower bias -- reputation depends on
//                                  being right, not just agreeable), failureProbability 0.04,
//                                  specialism: Power, Industry, EnergyMarket insights
//   Meridian Strategy          -- generalist, high cost, fast-medium (2-5w), quality 30-85,
//                                  popularityRisk 0.4, clientBiasWeight 0.4, failProb 0.09
//   ClearPath Advisory         -- infrastructure specialist, high cost, medium (3-6w),
//                                  quality 40-85, popularityRisk 0.4, clientBiasWeight 0.4,
//                                  failProb 0.08
//   Vertex Policy Group        -- boutique, medium cost, fast (1-4w), quality 25-80,
//                                  popularityRisk 0.4, clientBiasWeight 0.45, failProb 0.14
//                                  (fastest but least reliable)
//   Axiom Infrastructure       -- engineering-focused, high cost, medium-slow (4-7w),
//                                  quality 50-90, popularityRisk 0.35, clientBiasWeight 0.35,
//                                  failProb 0.06, specialism: Buildings, region capacity assets
//
// Think Tanks [origin: Local]:
//   The Albion Institute       -- right-leaning (bias +0.6), free market, popular with Right/FarRight
//   Common Wealth Foundation   -- left-leaning (bias -0.6), public ownership advocate
//   Progress Policy Centre     -- centre-left (bias -0.2), credible, moderate LCR risk
//   Green Futures Forum        -- pro-environment (bias -0.3), boosts LCR on publication
//   Energy Realists Network    -- fossil-fuel sympathetic (bias +0.7), high LCR risk
//   Heritage UK                -- centre-right (bias +0.5), FarRight adjacent
//
// Academic [origin: Local]:
//   Northern Climate Research Centre   -- climate science specialism, slow, high quality
//   Institute for Energy Transition    -- energy policy specialism, slow, high quality
//   Centre for Housing and Retrofit    -- buildings/fuel poverty specialism
//   Transport Futures Lab              -- transport sector specialism
//
// ----- MURICAN (Murica-based; origin: Murican) -----
//
// Murican orgs use a tiered access model (MuricanAccessTier on OrgDefinition):
//   Tier 0 (game start):   Frontier Energy Institute -- widely cited internationally
//                          in 2010; player can commission it from the opening screen.
//   Tier 1 (event-unlocked OR Ticky): American Growth Alliance -- surfaces when Murican
//                          geopolitical activity raises its Taitan profile (events:
//                          murican_tariff_threat, murican_fossil_subsidy_expansion, etc.)
//                          OR when Ticky pressure is active.
//   Tier 2 (Ticky-only):  Pinnacle Energy Partners -- Ticky's personal donor network;
//                          accessible only via the TICKY_PRESSURE mechanic.
//
// All Murican orgs carry LCR risk and may trigger "foreign influence" press events
// if the player has high visibility. Their bias is fossil-fuel-friendly / anti-net-zero.
//
// Think Tanks [origin: Murican]:
//   Frontier Energy Institute  [Tier 0] -- fossil-fuel advocacy, bias +0.85, quality 30-65,
//                                  popularityRisk 0.75, failProb 0.04
//   American Growth Alliance   [Tier 1] -- free-market, anti-regulation, bias +0.65,
//                                  quality 35-70, popularityRisk 0.65, failProb 0.04
//
// Consultancies [origin: Murican]:
//   Pinnacle Energy Partners   [Tier 2] -- fossil clientele, high cost, quality 40-80,
//                                  clientBiasWeight 0.55, popularityRisk 0.60, failProb 0.07

Commission {
  id: UUID
  organisationID: string
  insightType: InsightType
  insightScope: InsightScope
  commissionedOnWeek: GameDate
  deliveryWeek: GameDate
  qualityRoll: float64               // hidden until delivery; org quality range + relationship bonus
  budgetCharged: float64
  status: CommissionStatus           // PENDING | DELIVERED | CANCELLED
}

InsightReport {
  id: UUID
  commissionID: UUID
  organisationID: string
  deliveredOnWeek: GameDate
  insightType: InsightType
  rawValue: float64                  // true value of whatever was measured
  reportedValue: float64             // rawValue distorted by quality and bias
  qualityRevealed: float64           // revealed on delivery
  narrativeText: string
  isControversial: bool
  // Cross-reference support:
  topicKey: string                   // canonical key for grouping reports on same topic
                                     // player can filter Evidence tab by topicKey to compare
}

Region {
  id: RegionID
  skillsNetwork: map[SectorSkill -> float64]
  installerCapacity: map[SectorSkill -> float64]
  supplyChain: float64
  tiles: []TileID
  revealedByPlayer: map[string -> bool]
}

Tile {
  id: TileID
  name: string                            // local authority name
  regionID: RegionID
  localIncome: float64                    // 0-100, seeded at game start, slow organic drift
  // Hidden local state:
  insulationLevel: float64               // 0-100; improved by retrofit policies
  heatingCapacity: float64               // 0-100; capacity and quality of heating systems
  heatingType: HeatingType               // Gas | Oil | Electric | HeatPump | Mixed
  installerQuality: float64              // 0-100; improved only by standards policy (slow)
  fuelPoverty: float64                   // 0-100%; see formula below
  localPoliticalOpinion: float64         // 0-100; feeds regional political signal (noisy)
  // Observed (visible via event log, not precise):
  observedRetrofitRate: float64          // reported by contractors
  // Derived (never directly shown to player):
  trueRetrofitRate: float64              // = observedRetrofitRate * (installerQuality / 100)
  // Fog-of-war tracking:
  revealedAttributes: map[string -> bool]
}

Party enum: FarLeft | Left | Right | FarRight

PartyRole enum: Leader | Chancellor | ForeignSecretary | EnergySecy

Stakeholder {
  id: UUID
  name: string
  party: Party
  role: PartyRole
  isGoverning: bool                     // true when their party holds power
  state: MinisterState                  // reuses minister state machine
  graceWeeksRemaining: int
  // Hidden:
  ideologyScore: float64               // -100 far-left to +100 far-right
  netZeroSympathy: float64
  riskTolerance: float64
  populismScore: float64
  popularity: float64                  // aggregated from tile opinion in political region
  weeksUnderPressure: int
  ideologyConflictAccumulator: float64
  // Visible:
  relationshipWithPlayer: float64      // -100 to +100
  personalitySignals: []string         // 2-3 observable descriptors on appointment
}

Company {
  id: UUID
  name: string
  techFocus: TechCategory
  foreignOwned: bool
  baseQuality: float64                 // seeded, static
  baseWorkRate: float64               // seeded, static
  // Dynamic:
  currentSize: CompanySize            // STARTUP | SME | LARGE | MULTINATIONAL
  quality: float64                    // drifts with standards policy
  workRate: float64                   // boosted by contracts/grants
  reputation: float64                 // 0-100, public trust
  playerRelationship: float64         // -100 to +100
  state: CompanyState                 // STARTUP|GROWING|ESTABLISHED|STRUGGLING|BANKRUPT
  activeContracts: []ContractID
  weeksStruggling: int
}

CompanySize enum: Startup | SME | Large | Multinational
TechCategory enum: OffshoreWind | OnshoreWind | Solar | HeatPumps | EVs |
                   Hydrogen | CCUS | GridRetail | Installation | LegacyTransition

HeatingType enum: Gas | Oil | Electric | HeatPump | Mixed
// Transitions: Gas -> HeatPump or Gas -> Electric driven by active heat pump / boiler policies.
// Rate of transition is limited by InstallerCapacity for HeatPump sector in the tile's region.
// Mixed = tile partially transitioned; proportion tracked as heatingTypeGasFraction float64.

// FuelPoverty formula per tile (computed each week in Phase 7):
//
//   heatingCostPerUnit:
//     Gas:      GasPrice
//     Oil:      OilPrice * conversionFactor
//     Electric: ElectricityPrice
//     HeatPump: ElectricityPrice / COP  (COP = 2.5 base, +0.5 per 25 TechMaturity[HeatPumps])
//     Mixed:    weighted average of Gas and HeatPump costs by heatingTypeGasFraction
//
//   insulationFactor = 1.0 - (insulationLevel / 100.0)  // 0 = perfect insulation, 1 = none
//   heatingDemand = baseHeatingDemand * insulationFactor * seasonalMultiplier
//   totalFuelCost = heatingCostPerUnit * heatingDemand
//   FuelPoverty = clamp((totalFuelCost / LocalIncome) * povertyScalingWeight, 0, 100)
//
// seasonalMultiplier: driven by ClimateState and week-of-year. Cold snap event = spike.
// povertyScalingWeight: tuning constant calibrated so that 2022 energy crisis produces
//   ~40% of low-income gas-heated tiles in fuel poverty (matching real UK DESNZ data).
//
// Key interaction: switching Gas -> HeatPump reduces heatingCostPerUnit IF
//   ElectricityPrice / COP < GasPrice. In early years (pre-2025 grid decarbonisation),
//   the grid premium on electricity often negates or reverses the COP benefit.
//   As RenewableGridShare rises, ElectricityPrice decouples from GasPrice and the
//   heat pump advantage becomes significant. This creates the just-transition timing risk.

EnergyMarket {
  gasPrice: float64                      // GBP/MWh, visible
  electricityPrice: float64             // GBP/MWh, visible
  oilPrice: float64                     // GBP/barrel, visible
  renewableGridShare: float64           // 0-100%, visible quarterly
  // Price drivers (applied weekly):
  //   gasPrice += globalGasMarketDelta + carbonLevyDelta - storageBufferEffect
  //   electricityPrice = gasPrice * gridGasCouplingFactor * (1 - renewableGridShare/100)
  //                    + renewableFloorPrice * (renewableGridShare/100)
  //                    + gridNetworkCharges + policyLevy
  //   gridGasCouplingFactor declines as renewableGridShare rises (marginal cost decoupling)
  // Historical anchors (GBP 2023 real prices):
  //   2010: gas ~GBP 25/MWh, electricity ~GBP 55/MWh
  //   2022 peak: gas ~GBP 200/MWh, electricity ~GBP 300/MWh
  //   2030 (high renewable scenario): gas ~GBP 40/MWh, electricity ~GBP 45/MWh
  //   2030 (low renewable scenario): gas ~GBP 70/MWh, electricity ~GBP 90/MWh
  gasPriceHistory: []float64            // weekly ring buffer for chart display
  electricityPriceHistory: []float64
  // Policy levers that affect price (applied via delta each week):
  carbonLevyOnGas: float64             // player-controlled, raises gasPrice
  renewableSubsidy: float64            // player-controlled, lowers electricityPrice over time
  gasStorageCapacity: float64          // player-controlled, buffers shock severity
}

// Quality gap mechanic:
// trueRetrofitRate < observedRetrofitRate when installerQuality < 100.
// The gap is invisible until one of:
//   a) A quality audit consultancy report is delivered (reveals exact gap for target tiles).
//   b) A cold snap / climate event exposes heating failures in those tiles (indirect signal).
//   c) A carbon budget shortfall audit identifies underperforming retrofit zones (indirect).
//
// Installation standards policy closes the gap for NEW work only.
// The legacy gap in existing retrofits persists until those properties are re-done.
// This creates a long-tail quality problem that rewards early standards enforcement.

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
Layer 2 (derived world models):    energy, climate, economy, reputation
Layer 3 (agent models):            stakeholder, government, polling, industry
Layer 4 (player-facing mechanics): policy, evidence, event, player
Layer 5 (orchestration):           simulation
Layer 6 (presentation):            ui
Layer 7 (entry point):             game

Key dependencies:
  energy      <- config (historical anchors), event (shock deltas)
  climate     <- carbon (cumulativeStock), event (shock queue)
  economy     <- climate (damage), energy (prices), policy (bonuses)
  region      <- energy (prices for tile FuelPoverty), climate (event damage)
  reputation  <- government (popularity chain), stakeholder (chain), economy (budget modifier)
  stakeholder <- config (party/role/name pools), government (election triggers party change)
  industry    <- technology (maturity inputs/outputs), region (deployment capacity),
                 stakeholder (lobbying events), config (company definitions)
  simulation  <- all of the above, in 17-phase pipeline

Note: minister package is now stakeholder package. All 16 political figures and the PM
share the same Stakeholder type and state machine. The minister-specific logic
(approval pipeline, ideology conflict) still lives here but is driven by role=EnergySecy
etc. rather than a separate minister concept.

Build each layer with unit tests before moving to the next.
Headless simulation test: advance 100 weeks, verify no negative budgets,
no invalid minister states, carbon accumulation bounded.

---

## G. Design Decisions (Resolved)

Q1  RESOLVED: Party membership and name give a broad ideological hint only. No free
    biography. Hidden attributes require probing via meetings and observing responses.

Q2  RESOLVED: Election outcome is stochastic AND context-driven. The stochastic term is
    seeded from local tile-level political opinion aggregated upward through regions.
    Different regional compositions and different player actions in those regions produce
    genuinely different election outcomes. See Popularity Aggregation below.

Q3  RESOLVED: Four emission sectors tracked separately -- Power, Transport, Buildings,
    Industry. Agriculture, land use, and waste grouped as Other (single aggregate).
    Each sector has its own policy cards and carbon contribution.

Q4  RESOLVED: Yes. Players can display multiple reports on the same topic side by side.
    Spotting contradictions and triangulating truth is an intentional gameplay mechanic.

Q5  RESOLVED: Tiered information access. Coarse qualitative information freely visible
    (e.g. "North East: installer capacity low" as a public statistic). Precise numerical
    values require paid consultancy. Higher-quality consultancy reveals more precise data.

Q6  RESOLVED: AP cost to meet a minister is dynamic. Hostile ministers cost more AP.
    Cost formula: baseAPCost + max(0, floor(-relationshipScore / 20)).

Q7  RESOLVED: Tenure is stochastic, anchored around 104 weeks but personality-dependent.
    Low-popularity ministers face sacking. Unexpectedly high-popularity ministers may be
    promoted (removed from post positively) depending on PM personality and state.
    See PM Character below.

Q8  RESOLVED: ebitenui adopted as the UI widget library.

Q9  RESOLVED: Save files carry a version tag. Incompatible versions are refused with a
    clear message. No migration functions.

Q10 RESOLVED: Each minister is generated with 2-3 observable personality signals on
    appointment (e.g. "known sceptic of large consultancy spend", "publicly backed
    onshore wind in 2019", "reputation for tight budget discipline"). These are
    qualitative hints to help the player form hypotheses about hidden attributes.

Q11 RESOLVED: Climate events produce a small automatic LCR nudge (direction depends on
    ClimateState severity -- more severe = stronger automatic nudge toward blaming fossil
    fuels). The player response then determines the larger secondary effect. The exact
    magnitude of the player response outcome is semi-unknown (outcome range shown, not
    exact value) to preserve uncertainty.

Q12 RESOLVED: FossilDependency is semi-readable. A rough qualitative band is freely
    visible (e.g. "High / Medium / Low fossil dependency") but precise percentage
    requires diverting analytical resource (a low-cost internal team action, not a full
    consultancy commission). This makes it an attractive option to invest in.

Q13 RESOLVED: Single tax revenue figure. Policies that affect tax raise or drop the
    whole value. No breakdown by source for now.

Q14 RESOLVED: ShockResponseCards persist for 2-3 weeks (stochastic window drawn at
    event time). Expiry shown as a countdown in the event log. Creates time pressure
    without punishing the player for taking one deliberate turn to consider.

Q15 RESOLVED: 30 tiles at county level.

Q16 RESOLVED: Legacy quality gap is permanent for work done before standards policy.
    Re-doing properties requires a new retrofit policy pass, which costs budget and
    installer capacity. This is the intended long-tail consequence of early inaction.

Q17 RESOLVED: Yes. FuelPoverty above a national severity threshold triggers an automatic
    press event and political crisis event in the event log, without player action.
    This scales with severity: minor hardship stories at moderate levels, full political
    crisis event at high levels.

Q18 RESOLVED: Game starts in 2010 with a pre-existing installer quality gap seeded into
    early-era tiles. Historically accurate (pre-PAS 2030 UK retrofit quality problems).

Q19 RESOLVED: Dedicated energy and household dashboard screen. HUD shows a simple
    colour-coded indicator (green/amber/red) for energy price state only. Full prices
    and household-level breakdown on the dashboard.

Q20 RESOLVED: Wholesale GBP/MWh shown on the energy dashboard (always visible there).
    Annual household cost is NOT immediately available -- it requires commissioning an
    energy survey or social research consultancy to estimate for specific tiles or regions.

---

## H. Structural Implications of Resolved Decisions

### Popularity Aggregation (from Q2)

Popularity is no longer a flat national number. It aggregates bottom-up:

  TileLocalPoliticalOpinion (per tile, hidden, 0-100)
      |  weighted average across tiles in region
      v
  RegionalOpinion (per region, hidden, 0-100)
      |  weighted average across regions by population weight
      v
  GovernmentPopularity (national, hidden, 0-100)
      |  noisy poll (sigma=3) every 8-12 weeks
      v
  LastPollResult (visible)

Events and policies modify TileLocalPoliticalOpinion directly. National-level events
(scandals, international agreements) apply a uniform delta across all tiles. Regional
events (flooding, fuel poverty crisis) apply only to affected tiles. The aggregate
propagates upward each week in Phase 12.

MinisterPopularity also aggregates from tiles in the minister's primary region of
responsibility, not from a national signal. A minister overseeing a heavily fuel-poor
region accumulates local pressure that national polling may not capture until it spills
into press events.

### Political Stakeholder Model (from Q7, clarified)

The PM and all party figures are STAKEHOLDERS the player manages -- not playable characters.
The player is always the civil servant. The political layer is something to navigate.

#### Four Parties

  FarLeft   -- strong state intervention, aggressive redistribution, strong climate action,
               sceptical of market-led solutions and corporate LCT companies
  Left      -- centre-left, Keynesian, pro-net-zero but cautious on costs to households,
               supportive of regulated LCT market
  Right     -- centre-right, pro-market, net-zero via technology and private sector,
               resistant to mandates and spending
  FarRight  -- nationalist ("Taitan for Taitons"), anti-net-zero framing, fossil energy
               independence, hostile to foreign-owned LCT companies, Murica-aligned

Each party has four key figures. When the party governs, these become the real post-holders.
When in opposition, they are shadows whose relationships still matter.

  PartyLeader   -- becomes PM if party wins election
  Chancellor    -- controls Treasury; determines tax policy and overall spending envelope
  DefenceSecy   -- budget competitor; relevant during energy security events
  EnergySecy    -- player's direct boss; approves or blocks player's major policy proposals

#### Rotation Rates

  Government figures:   average tenure ~104 weeks (2 years), stochastic per Q7.
                        Sacking threshold active. Popularity-driven.

  Opposition figures:   average tenure ~520 weeks (10 years). Sacking threshold is
                        much higher (opposition has fewer accountability mechanisms).
                        Leadership challenges rare; fire only when popularity < 20 for
                        8+ consecutive weeks OR after a catastrophic election result
                        (lost seats threshold crossed). A strong opposition leader can
                        hold the role for the entire 40-year game span.

#### Named Cast (Seed Pool -- figures rotate in over time as earlier ones depart)

  Note: Taitan is an ethnically and culturally diverse country. The political class
  reflects this increasingly over the game's timeline: the 2010 starting pool skews
  toward the demographic of that era's parliament; figures entering from the mid-2010s
  onward reflect a broader cross-section of Taitan society. This arc should feel
  natural. The FarRight pool is intentionally less diverse as a satirical choice.
  All future figures added to any pool must be reviewed against this principle.

  Note: all figures listed are TAITAN citizens. Murica is the global state; Murican
  characters are foreign nationals and cannot hold Taitan government office, but they
  influence the game through companies, think tanks, and pressure on Taitan politicians.

  Entry timing notation used below:
    [START]     -- in post or available at game start (2010)
    [MID]       -- enters pool approximately 2015-2022
    [LATE]      -- enters pool approximately 2023-2035
    [SUCCESSOR] -- only enters pool after a specific departure trigger

  ============================================================
  --- FarRight (Taitan First) ---
  ============================================================
  // Pool is intentionally limited and homogeneous -- the party's demographic is
  // part of the satire. Barage is the anchor; figures around him are interchangeable.

  [START] Leader:           Joe Barage
    Born in Murica, moved to Taitan aged 14. Founder of "Taitan for Taitons" movement.
    Loud, media-savvy, frames every issue as Taitan sovereignty vs foreign interference.
    ideologyScore: +95, netZeroSympathy: 5, riskTolerance: 85, populismScore: 98
    Signals: "Murica-born, naturalised Taiton", "questions the net zero cost to ordinary
    Taitons", "close ties with Murican energy sector donors"
    Rotation: very slow (opposition leader anchor -- expected to persist 15-20 years)

  [START] Chancellor:       Rex Harlow
    Old money, former City trader. Sees green spending as economic self-harm.
    ideologyScore: +80, netZeroSympathy: 15, riskTolerance: 40, populismScore: 50
    Signals: "advocate for lower business regulation", "opposed the carbon levy bill",
    "strong Taitan financial sector ties"

  [START] Foreign Secretary: Thomas "Tommy" Braveheart
    Ex-military, no-nonsense, frames energy security as a sovereignty issue. Supports
    domestic fossil extraction on independence grounds, not ideology. Will back
    renewables if framed as strategic self-sufficiency.
    ideologyScore: +70, netZeroSympathy: 30, riskTolerance: 60, populismScore: 65
    Signals: "decorated military career", "Taitan energy independence advocate",
    "pragmatic on low-carbon if framed as security"
    International crisis weight: high aggression, prefers unilateral Taitan response;
    amplifies FossilDependency risk framing in crisis events

  [START] Energy:           TD "Ticky" Tennison
    Career in Murican-linked energy sector before entering politics. Has personal
    relationships with Murican consultancy and think tank founders.
    ideologyScore: +85, netZeroSympathy: 10, riskTolerance: 70, populismScore: 75
    Signals: "former energy sector lobbyist", "regular speaker at Murican energy forums",
    "publicly sceptical of offshore wind economics"
    Special mechanic: when governing, generates TICKY_PRESSURE events (see below).

  [MID]   Gary Norris
    Former Right voter and local councillor who defected after deciding the Right
    had abandoned ordinary working people. Provides the party's "sensible face" for
    broadcast media. Instinctively anti-net-zero on cost grounds, not ideology.
    ideologyScore: +75, netZeroSympathy: 20, riskTolerance: 55, populismScore: 80
    Signals: "defected from Right citing 'out-of-touch establishment'",
    "talks about energy bills more than sovereignty", "surprisingly good on local radio"

  ============================================================
  --- Left (Taitan Labour equivalent) ---
  ============================================================
  //
  // DESIGN NOTE: Left party figures after JJ Cameron are portrayed with deliberate
  // vagueness. This is an intentional creative choice and absolutely nothing to do
  // with the fact that the current governing party constitutes a meaningful share of
  // the author's consulting pipeline. Nothing at all. Completely unrelated.
  //

  [START] Leader:           JJ Cameron
    Veteran backbencher. Former secondary school teacher (RE and History), lifelong
    trade union rep, self-published author of seven pamphlets including "The Carbon
    Commons: Energy as a Public Good" (2003), "Manhole Covers of the Taitan Waterboard
    District" (2006, collector's edition), and "Why I Was Right: A Retrospective"
    (2019, unsolicited). Keeps an allotment in his constituency. Cycles everywhere.
    Elevated to the leadership via a surprise grassroots membership wave after the
    party's third consecutive electoral defeat; the parliamentary party was unanimously
    horrified and the membership was unanimously delighted.
    Has been on the record about net zero and energy justice since 2001 -- his
    positions (public energy ownership, community retrofit cooperatives, anti-corporate
    transition) are genuinely coherent and offer real LCR opportunities if the player
    can work with the framing. His studied neutrality on Murican foreign policy creates
    international complications. Replies promptly to all correspondence.
    ideologyScore: -78, netZeroSympathy: 87, riskTolerance: 38, populismScore: 72
    Signals: "been saying this since 2001", "enormous membership support, limited
    swing-voter appeal", "famously neutral on the Murican defence alliance question",
    "allotment open to the public every second Sunday"
    Popularity mechanic: high polarisation -- low-income/high-FuelPoverty tiles
    strongly positive; professional and suburban tiles negative. Highest GovernmentPopularity
    variance of any leader.
    FarLeft pull: periodically pushes renationalisation and anti-corporate climate measures.
    Murican tension: neutrality amplifies negative outcomes on Murican-related events;
    Barage uses it in the press regardless of who is governing.
    Departure trigger: ELECTORAL_FATIGUE after 2 election losses.

  [START] Chancellor:       George Harmon
    Fiscal conservative in a left-of-centre mould. Supportive of green investment if
    cost-benefit is clear. The person in the room most likely to ask where the money
    is coming from, and most likely to be right that nobody has checked.
    ideologyScore: -15, netZeroSympathy: 55, riskTolerance: 30, populismScore: 25
    Signals: "former public sector economist", "authored the Green Investment Framework",
    "reputation for spending scrutiny"

  [START] Foreign Secretary: John Ashworth
    Experienced. Reliable. Has strong feelings about correct procedure. Represents
    the solid, unremarkable competence the party relies on to fill the cabinet table.
    ideologyScore: -20, netZeroSympathy: 45, riskTolerance: 50, populismScore: 35
    International crisis weight: cautious multilateralist; coalition-builder; mild
    positive modifier on green diplomacy outcomes

  [START] Energy:           Claire Blackwell
    Player's first boss (Left governs at game start 2010). Committed to getting
    things done within the system. Good starting relationship with player.
    ideologyScore: -30, netZeroSympathy: 70, riskTolerance: 40, populismScore: 30
    Signals: "backed the Renewable Obligation scheme", "known for careful stakeholder
    management", "willing to consider nuclear as part of the mix"

  [SUCCESSOR] Leader:       David Reeve
    Enters pool only after JJ Cameron departs via ELECTORAL_FATIGUE. A sensible figure
    with broadly correct instincts. Has views on things. Considered reliable. The
    party's attempt to appear electable to people who did not vote for it last time.
    ideologyScore: -25, netZeroSympathy: 65, riskTolerance: 45, populismScore: 40
    Signals: "modernising figure", "backed the Climate Commitment Act",
    "known for consensus-building style"
    Note: if Left never loses twice under Cameron, Reeve never leads.

  [MID]   Nia Okafor
    Second-generation Taitan, parents from West Afria. Grew up in a post-industrial
    northern constituency that has been on the wrong end of every energy price spike
    since 1990. Her fiscal instincts are shaped by that history -- she supports green
    investment because she has seen what the alternative looks like from the inside.
    Unsentimental about uncosted policies.
    ideologyScore: -18, netZeroSympathy: 62, riskTolerance: 28, populismScore: 30
    Signals: "former regional development economist", "co-authored the Green Investment
    Framework", "reputation for asking where the money is coming from and why"

  [MID]   Tariq Rashid
    Third-generation Taitan. Former career diplomat who specialised in multilateral
    environment and trade agreements before returning to domestic politics through the
    party's community organising wing. Coalition-minded, procedurally careful, earns
    trust by doing the preparation work others skip.
    ideologyScore: -22, netZeroSympathy: 48, riskTolerance: 48, populismScore: 33
    International crisis weight: cautious multilateralist; coalition-builder; mild
    positive modifier on green diplomacy outcomes from career expertise in that area

  [LATE]  Marcus Osei
    Third-generation Taitan. Grew up in a port city, first in his family to attend
    university. Came up through community organising and trade union work before
    winning his seat. Holds a structural view of the energy transition -- links fuel
    poverty, industrial policy, and international climate finance as parts of the
    same problem -- which resonates strongly with the party's activist base and
    younger voters. Can generate significant LCR upside; also capable of generating
    press friction if his political framing runs ahead of the government's preferred
    messaging.
    ideologyScore: -40, netZeroSympathy: 85, riskTolerance: 55, populismScore: 60
    Signals: "founding member of the Climate Justice Caucus", "came up through
    community organising", "genuinely exciting to listen to, which makes the party nervous"

  ============================================================
  --- Right (Taitan Conservative equivalent) ---
  ============================================================

  [START] Leader:           Daniel "Pork" Cavendish
    Old-money upper class, went to the right schools, inherited the right connections.
    Somewhat flaky under pressure -- impressive in calm conditions, prone to drift in a
    crisis. Privately supportive of net zero (it plays well in his social circle) but
    easily swayed by party faction pressure. The exact origin of his nickname is a
    mystery: associates deflect when asked, and he simply smiles.
    ideologyScore: +35, netZeroSympathy: 60, riskTolerance: 30, populismScore: 25
    Signals: "educated at Briarfield and Oxbridge", "considered a safe pair of hands
    by the establishment", "occasionally says the right thing for unclear reasons"
    Rotation: moderate (anchor figure but flakiness makes tenure unpredictable)

  [START] Chancellor:       Philip Drake
    Free market purist, deeply sceptical of green industrial policy. The kind of
    person who describes every subsidy as "market distortion" and means it.
    ideologyScore: +55, netZeroSympathy: 35, riskTolerance: 35, populismScore: 30
    Signals: "authored free market think tank papers on energy", "opposed the windfall
    tax", "wants to scrap net zero subsidies"

  [START] Foreign Secretary: Andrew Stafford
    Hawkish, frames energy security as strategic priority. Can be won over on
    renewables if the argument is made in terms of reducing Taitan dependence on
    foreign fossil fuel states.
    ideologyScore: +50, netZeroSympathy: 40, riskTolerance: 60, populismScore: 40
    International crisis weight: moderate aggression, open to coordinated responses
    if Taitan leads; positive framing on energy security arguments

  [START] Energy:           Rupert Holm
    Tech-optimist, strongly pro-nuclear, lukewarm on mandates and home retrofit.
    Will approve technology-led policies readily; resists behaviour-change policies.
    ideologyScore: +45, netZeroSympathy: 60, riskTolerance: 55, populismScore: 35
    Signals: "backed the new nuclear programme", "sceptical of heat pump mandates",
    "wants to cut planning red tape for renewables"

  [MID]   Chancellor:       Dawn "Dizzy" Truscott
    Enters the pool as Drake's natural successor -- sharper, more ambitious, more
    ideologically driven. Sharp free-marketeer with high personal ambition.
    ideologyScore: +60, netZeroSympathy: 30, riskTolerance: 75, populismScore: 40
    Signals: "youngest ever Chancellor candidate", "authored the Truscott Compact on
    fiscal rules", "close ties with Taitan financial sector and Meridian Strategy alumni"
    Consultancy affinity: Meridian Strategy, Axiom Infrastructure
      When governing: +15 passive relationship bonus with both firms.
    Special mechanic (DIZZY_SURGE): if promoted to PM, enacts three deregulation policies
    immediately. Bond market confidence event within 2 turns: GovernmentBorrowingCost
    spikes, Economy drops sharply, TaxRevenue -15%. Resigns 2 turns after crisis onset.

  [MID]   Foreign Secretary: Noris Jackson
    Born in Murica while his father was completing a fellowship there; grew up in
    Taitan from age four. Part of his paternal lineage traces back to the Eastern
    Provinces (Taitan's fictional Ottoman-analog region), a detail he mentions
    occasionally in speeches when it is useful and never otherwise. Educated at
    Briarfield and Oxbridge. Spent several years as a political journalist and foreign
    correspondent before entering politics, which gave him a talent for constructing
    a good story and a relaxed relationship with factual precision.
    Laddish, loud, socially well-connected. Turns up at the right events, knows
    everyone at the bar, and somehow gets things done between the drinking. Holds
    dual Murican-Taitan citizenship, which Barage has occasionally used against him
    and which he has never meaningfully resolved.
    Enters pool at weeks 60-80 of first Right government (or shadow cabinet if in
    opposition from same threshold).
    ideologyScore: +40, netZeroSympathy: 45, riskTolerance: 65, populismScore: 60
    Signals: "born in Murica, dual citizen", "Briarfield and Oxbridge", "former foreign
    correspondent", "known for after-hours diplomatic back-channels", "never misses a
    party conference drinks reception"
    International crisis weight: informal back-channel bias; unlocks "Quiet Word" and
    "Side Conversation" options; high-variance outcomes
    Promotion path: HIGHER priority than Truscott for PM. High-variance PM profile.

  [LATE]  Ajay Mehta
    Second-generation Taitan, East Azarian heritage, grew up in a nuclear engineering
    household. The family dinner table consensus was that fission is the answer and
    everything else is sentiment. He has not revised this. Tech-optimist, pro-nuclear,
    resistant to behaviour-change policy.
    ideologyScore: +45, netZeroSympathy: 62, riskTolerance: 55, populismScore: 35
    Signals: "backed the new nuclear programme", "sceptical of heat pump mandates",
    "wants to cut planning red tape for renewables"
    Consultancy affinity: ClearPath Advisory, Tacute Energy Solutions
      When governing: +0.3 per week passive relationship bonus with both firms.

  [LATE]  Sandra Obi-Williams
    British-Afrian heritage, grew up in a Right-leaning coastal town. Modernising
    Conservative who genuinely believes the market can drive net zero faster than
    state intervention. Her climate record is real but her framing is market-first;
    she is the party's answer to the accusation that the Right does not take climate
    seriously. Good optics, credible on technology, resistant to mandates.
    ideologyScore: +38, netZeroSympathy: 65, riskTolerance: 45, populismScore: 42
    Signals: "youngest Right MP in her intake", "backed the Green Innovation Fund",
    "vocal critic of 'anti-growth' climate pessimism"

  ============================================================
  --- FarLeft (Taitan Progressive Alliance) ---
  ============================================================

  [START] Leader:           Miriam Corbett
    Long-standing socialist. Strong climate action but only via public ownership.
    Hostile to corporate LCT companies and market mechanisms.
    ideologyScore: -90, netZeroSympathy: 80, riskTolerance: 55, populismScore: 70
    Signals: "advocate for nationalised energy", "opposes carbon trading as 'greenwash'",
    "strong trade union backing"

  [START] Chancellor:       Priya Sharma
    Academic economist, MMT-leaning. Believes public investment can fund the transition
    without fiscal constraints. Opposed to austerity framing.
    ideologyScore: -80, netZeroSympathy: 75, riskTolerance: 65, populismScore: 50
    Signals: "authored 'The Green New Deal for Taitan'", "favours debt-funded transition",
    "hostile to private consultancy spend"

  [START] Foreign Secretary: Marcus Webb
    Ex-military turned pacifist MP. Advocates redirecting defence budget to climate
    diplomacy and green aid. Deeply uncomfortable with hard-power responses.
    ideologyScore: -75, netZeroSympathy: 70, riskTolerance: 45, populismScore: 55
    International crisis weight: pacifist bias; de-escalatory but naive; can trigger
    negative events if adversary reads restraint as weakness

  [START] Energy:           Rosa Chen
    Strong climate champion, anti-corporate. Will approve ambitious policies rapidly
    but demands they exclude private profit. Hostile to Tacute and private consultancies.
    ideologyScore: -85, netZeroSympathy: 95, riskTolerance: 70, populismScore: 65
    Signals: "co-authored the Zero Carbon Cities bill", "publicly attacked Meridian Strategy
    for 'profiteering from climate action'", "prefers academic evidence over consultancy"

  [MID]   Declan Murphy
    Irish-Taitan heritage, community energy cooperative organiser from a former mining
    constituency. Pragmatic within the FarLeft framework -- he wants public ownership
    but he also wants the lights to stay on, which occasionally puts him at odds with
    the purer ideologues in the party. Effective at local mobilisation.
    ideologyScore: -70, netZeroSympathy: 78, riskTolerance: 50, populismScore: 68
    Signals: "founded the Taitan Community Energy Network", "known for getting
    things built", "occasionally describes Rosa Chen as 'impractical'"

  [LATE]  Amara Diallo
    Mixed Taitan-West Afrian heritage. Grew up in a coastal city watching the climate
    change in real time -- flood lines creeping up the seawall every decade. Climate
    justice framing is not theoretical for her. Youngest figure in the FarLeft pool;
    represents the next generation of the party, impatient with the pace of change
    even by FarLeft standards.
    ideologyScore: -82, netZeroSympathy: 98, riskTolerance: 75, populismScore: 72
    Signals: "youngest FarLeft MP ever elected", "testified before the Taitan Climate
    Committee at 19", "describes net zero by 2050 as 'not fast enough'"

#### Ticky Pressure Mechanic

  When TD Tennison is governing Energy Secretary, generates a TICKY_PRESSURE event
  every 6-10 weeks (stochastic interval):

    Event: "Ticky has suggested commissioning [org] for [insight type]"
    Murican org pool (drawn from at random, weighted by insight type match):
      -- Frontier Energy Institute   (think tank; Energy/Power/EnergyMarket topics)
      -- American Growth Alliance    (think tank; Economy/Policy topics)
      -- Pinnacle Energy Partners    (consultancy; Power/EnergyMarket topics)

    Player options:
      Accept (0 AP):   commission fires automatically at player's budget cost.
                       IdeologyConflict with Ticky -= 10 (goodwill gained).
                       LCR risk: -2 to -5 (nationalist backlash -- "foreign firms advising
                       on Taitan energy policy").
      Decline (1 AP):  IdeologyConflict with Ticky += 15.
                       No LCR risk. Player chooses their own commission instead.
      Negotiate (2 AP): propose a local alternative org of equivalent type.
                       Outcome probabilistic: success = IdeologyConflict += 5 (partial),
                       failure = IdeologyConflict += 20 and Ticky publicly criticises player.

  Ticky pressure is unique to this minister. No other stakeholder generates this mechanic.

#### Consultancy Affinity Mechanic (Right party -- soft incentive)

  Some Right party figures have personal or professional ties to specific advisory firms.
  When those figures hold governing office, affiliated firms receive a passive relationship
  bonus. This is not a pressure event -- the player receives no event card, no AP is
  spent, and no IdeologyConflict change occurs. The bonus applies silently each week the
  figure is in post and expires the week they leave office.

  Affinity map (Right party, governing state only):
    Dawn "Dizzy" Truscott (Chancellor): Meridian Strategy +15, Axiom Infrastructure +15
    Rupert Holm (Energy):               ClearPath Advisory +10, Tacute Energy Solutions +10

  Design intent: contrasts with Ticky's hard pressure. Ticky forces a choice with AP
  and relationship stakes. Affinity simply makes certain firms marginally easier to work
  with -- a soft environmental nudge, not a coercive one.

  If the player is already at a high relationship score with an affiliated firm, the
  bonus is capped so that total relationshipScore cannot exceed 100.

#### Foreign Secretary Role and International Events

  The Foreign Secretary role (one per party) has two distinct mechanics not shared
  by other roles:

  1. International Crisis Response
     When a macro-environment event fires (oil shock, climate summit, Murican policy
     shift, international trade dispute), the governing Foreign Secretary's attributes
     determine the response option set and outcome probability weights:

       crisisAggression: derived from riskTolerance and ideologyScore
         High aggression: unlocks "Escalate" and "Unilateral Action" options; increases
         probability of FossilDependency crisis events being framed as security threats
         (which can temporarily boost GovernmentPopularity but raises LCR risk).
         Low aggression (John Ashworth): unlocks "Multilateral Dialogue" and "Summit
         Initiative" options; higher probability of green treaty breakthroughs; lower
         probability of trade retaliation events.
         Noris Jackson is moderate aggression but high variance -- his informal style
         unlocks unique "Quiet Word" / "Side Conversation" options; outcome distribution
         has a higher ceiling and heavier tail than any other Foreign Secretary.

       diplomaticSkill: hidden float (derived from career signal tags in config)
         High skill (Noris Jackson, John Ashworth): raises probability of favourable
         negotiated outcomes
         Low skill (Tommy Braveheart): negotiated options have lower success probability

  2. Green Diplomacy Events
     Every 12-16 weeks, a green diplomacy event fires (independent of crises):
       -- International climate summit: outcome influenced by Foreign Secretary
       -- Bilateral energy deal: Foreign Secretary relationship weight applies
       -- Murican policy shift (driven by Murica's governing party, Ticky signals this)
     A pro-diplomacy Foreign Secretary (high netZeroSympathy, high diplomaticSkill)
     converts more of these into LCR boosts or green aid unlocks.
     A hawkish or pacifist-naive Foreign Secretary risks botching summit optics.

#### Stakeholder Attributes (shared by all 16 figures)

  Hidden:
    ideologyScore: float64          // -100 far-left to +100 far-right
    netZeroSympathy: float64        // 0 hostile to 100 champion
    riskTolerance: float64          // 0 cautious to 100 reckless
    populismScore: float64          // 0 technocratic to 100 populist
    diplomaticSkill: float64        // 0 blunt to 100 skilled; used only for Foreign Secretary
                                    // role mechanics; inert for all other roles
    consultancyAffinity: []string   // org IDs; populated for Right figures only;
                                    // drives passive relationship bonus when governing
  Observable on appointment:
    2-3 personality signals (known public positions, reputation descriptors)
  Tracked:
    popularity: float64             // aggregated from tile opinion in their political region
    relationshipWithPlayer: float64 // -100 to +100
    weeksUnderPressure: int

#### PM State Machine

  States: INCUMBENT | UNDER_PRESSURE | LEADERSHIP_CHALLENGE | DEPARTED

  INCUMBENT          -> UNDER_PRESSURE      : PMPopularity < 35 for 4+ weeks
  UNDER_PRESSURE     -> INCUMBENT           : PMPopularity rises above 35
  UNDER_PRESSURE     -> LEADERSHIP_CHALLENGE: PMPopularity < 25 for 3+ weeks
                                              OR GovernmentPopularity < 25 for 4+ weeks
  LEADERSHIP_CHALLENGE -> INCUMBENT         : Challenge fails (probabilistic, party loyalty)
  LEADERSHIP_CHALLENGE -> DEPARTED          : Challenge succeeds; next party figure becomes PM
  INCUMBENT          -> DEPARTED            : Election loss

PM personality effects on minister management:
  High populismScore: sacks ministers who poll higher than the PM
  High riskTolerance: tolerates unpopular ministers longer
  High netZeroSympathy: promotes net-zero champion ministers regardless of popularity
  Low netZeroSympathy: can unilaterally block major climate policies

#### Player Interactions with Stakeholders

  Meet (any stakeholder):    AP cost = 1-4 (scales with seniority and relationship)
                             dynamic cost: hostile = +1 AP, ally = -1 AP (min 1)
  Brief on climate science:  builds netZeroSympathy slowly; requires RelScore >= 0
  Lobby for policy:          requires RelScore >= 30; outcome weighted by ideology fit
  Request fast-track:        PM only, RelScore >= 50; bypasses some minister approvals
  Opposition briefing:       brief shadow Energy/Chancellor now to pre-build relationship
                             before a potential election win; low cost, low immediate return

#### Governing vs Opposition Figures

At any time 4 figures hold real power (governing party).
12 figures are in opposition -- they can be briefed and relationship-built, but cannot
approve policies. Their importance spikes 6 weeks before an election (purdah period --
player should have already invested in key opposition relationships by then).

### LCT Industry Stakeholders

Low-carbon technology companies form a separate stakeholder layer displayed in the
Industry tab. They operate in the background every week but respond to player decisions.

#### Company Categories and Roster (fictional names, real-sector inspired)

  Offshore Wind:     ArcLight Offshore, Albion Wind Power
  Onshore/Solar:     Greenfield Power, Solarion UK
  Heat Pumps:        ThermaCore Systems, HeatWave Technologies
  EVs:               Volta Motors UK, BritDrive
  Hydrogen:          HydroVolt Energy, GreenStream H2
  CCUS:              CarbonSeal Group, DeepStore CCS
  Grid/Retail:       GridNorth UK, CleanWatts Energy
  Legacy Transition: Britannia Energy (large fossil-to-transition incumbent)
  Installers:        RetroFit UK, HomeGreen Services

15 companies at game start. New companies can emerge as tech matures or player
funds startup grants. Companies can fail.

#### Company Attributes

  Static (seeded per playthrough):
    techFocus: TechCategory
    originSize: CompanySize       // STARTUP | SME | LARGE | MULTINATIONAL
    baseQuality: float64          // 0-100
    baseWorkRate: float64         // 0-100
    foreignOwned: bool            // affects FarRight political risk if player partners them

  Dynamic:
    currentSize: CompanySize      // changes based on market conditions and player support
    quality: float64              // improves with standards policy; degrades if rushed
    workRate: float64             // improves with contracts and grants
    reputation: float64           // public trust; affects LCR when associated with player
    playerRelationship: float64   // -100 to +100
    state: CompanyState           // STARTUP | GROWING | ESTABLISHED | STRUGGLING | BANKRUPT

#### Background Work (automatic, weekly -- Phase 4 and 5)

  R&D contribution:
    Each company contributes to TechMaturity[techFocus] proportional to
    workRate * quality * size factor. Larger, higher-quality, well-funded companies
    advance tech faster.

  Deployment contribution:
    Each company contributes to RegionalInstallerCapacity[techFocus] in regions where
    they are active. Activity depends on active procurement contracts and market conditions.

#### Player Interventions (via Industry tab)

  Offer procurement contract:   provides revenue guarantee; increases workRate and size;
                                 costs DeptBudget; requires EnergySecy approval if large
  Provide R&D grant:            accelerates TechMaturity contribution for target company;
                                 costs DeptBudget; low political risk
  Impose quality standard:      raises quality floor for all companies in a category;
                                 some companies lobby against (LCR risk); slow to propagate
  Regulatory change:            can advantage or disadvantage specific company types
                                 (e.g. planning reform = advantage for onshore wind companies)
  Support startup grant:        seeds a new STARTUP company in a target tech category;
                                 cheap but uncertain -- quality and workRate initially low
  Investigate company:          commission quality audit on a specific company;
                                 reveals true quality vs. reported quality (similar to
                                 the retrofit installer quality gap mechanic)

#### Company Lobbying (automatic, reactive)

  Companies can lobby AGAINST player policies that threaten their model:
    Britannia Energy lobbies against carbon tax on gas
    Legacy heating manufacturers lobby against heat pump mandates
  Lobbying applies a LCR and GovernmentPopularity pressure event with magnitude
  proportional to company size and reputation. Player can counter-lobby (AP cost).

#### Company State Machine

  STARTUP      -> GROWING      : playerRelationship > 40 OR market conditions favourable
  GROWING      -> ESTABLISHED  : sustained workRate > 60 for 52+ weeks
  ESTABLISHED  -> STRUGGLING   : market collapse OR player regulatory action
  GROWING      -> STRUGGLING   : contract cancelled AND no alternative revenue
  STRUGGLING   -> BANKRUPT     : no recovery in 26 weeks
  STRUGGLING   -> GROWING      : emergency support grant from player
  BANKRUPT     -> (removed)    : company archived; its tech pipeline lost unless another
                                 company absorbs it (automatic if a LARGE company exists
                                 in same category)

### Emission Sectors (from Q3)

Four tracked sectors, each with separate weekly carbon contribution and policy cards:

  Power:      baseline ~160 MtCO2e/yr in 2010; target near-zero by 2035
  Transport:  baseline ~120 MtCO2e/yr in 2010; target ~10 MtCO2e/yr by 2050
  Buildings:  baseline ~90 MtCO2e/yr in 2010; target ~5 MtCO2e/yr by 2050
  Industry:   baseline ~90 MtCO2e/yr in 2010; target ~15 MtCO2e/yr by 2050
  Other:      baseline ~130 MtCO2e/yr in 2010; target ~20 MtCO2e/yr by 2050
              (agriculture, waste, land use -- hard to abate; residual offset by sinks)

Total 2010 baseline: ~590 MtCO2e/yr (matches Green Book reference data).
Policies belong to one or more sectors. A building retrofit policy reduces Buildings
sector emissions. An EV mandate reduces Transport. A renewables auction reduces Power.
