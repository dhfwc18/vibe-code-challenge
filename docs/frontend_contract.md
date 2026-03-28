# Frontend Contract

This document is the reference for the Layer 6 frontend developer. It describes how to initialise and drive the simulation backend, which WorldState fields to read for each UI concern, which actions the player can submit, and how save/load works. All package paths are relative to the `app/` Go module root.

---

## 1. Game Loop

```
state := simulation.NewWorld(cfg, masterSeed)

// Each game tick (once per rendered week):
state, firedEvents := simulation.AdvanceWeek(state, actions)

// Render by reading fields from state -- do NOT call AdvanceWeek again
// until the player is ready to advance to the next week.
```

- `NewWorld(cfg *config.Config, masterSeed save.MasterSeed) WorldState` -- seeds a complete initial state. Week=0, Year=2010, Quarter=1.
- `AdvanceWeek(w WorldState, actions []Action) (WorldState, []event.EventEntry)` -- advances the clock by one week, applies all player actions, returns the new state plus any events that fired this week.
- The frontend renders fields from the returned `WorldState` each frame. Do not mutate the returned state directly.

---

## 2. WorldState Fields by UI Concern

### Clock display
| Field | Type | Notes |
|---|---|---|
| `Week` | `int` | Absolute week number; 0 = initial state |
| `Year` | `int` | Calendar year 2010-2050 |
| `Quarter` | `int` | 1-4 |

### Carbon and climate display
| Field | Type | Notes |
|---|---|---|
| `Carbon.CumulativeStockMt` | `float64` | Cumulative MtCO2e emitted since 2010 |
| `Carbon.AnnualBudgetExceeded` | `bool` | True when current year exceeds legal limit |
| `ClimateState.Level` | `string` | Enum: LOW / MEDIUM / HIGH / CRITICAL / EMERGENCY |
| `WeeklyNetCarbonMt` | `float64` | Net carbon this week (base minus policy reductions) |
| `WeeklyPolicyReductionMt` | `float64` | Carbon removed by active policies this week |

### Reputation and LCR
| Field | Type | Notes |
|---|---|---|
| `LCR.Value` | `float64` | Low-Carbon Reputation; 0-100 |
| `GovernmentLastPollResult` | `float64` | Most recent noisy government approval sample |
| `MinisterLastPollResults` | `map[string]float64` | Keyed by stakeholder ID; most recent noisy poll per minister |

### Energy prices
| Field | Type | Notes |
|---|---|---|
| `EnergyMarket.GasPrice` | `float64` | GBP per MWh |
| `EnergyMarket.ElectricityPrice` | `float64` | GBP per MWh |
| `EnergyMarket.RenewableGridShare` | `float64` | 0-1 fraction of grid from renewables |

The `EnergyMarket` struct also carries ring-buffer price history for sparkline charts; see `energy.EnergyMarket` for field details.

### Economy and budget
| Field | Type | Notes |
|---|---|---|
| `LastBudget.Departments` | `map[string]float64` | Key = department ID (see Section 7); value = GBPm allocated |
| `LastBudget.TotalGBPm` | `float64` | Total weekly departmental spend in GBPm |
| `LastTaxRevenue.GBPBillions` | `float64` | Most recent quarterly tax revenue |

### Technology
`Tech` is a `technology.TechTracker`. Call `Tech.Maturity(config.Technology)` to get the 0-1 maturity fraction for any of the 8 technologies (see Section 8).

### Government
| Field | Type | Notes |
|---|---|---|
| `Government.RulingParty` | `config.Party` | String ID of the governing party |
| `Government.CabinetByRole` | `map[config.Role]string` | Maps role to stakeholder ID; absent key = vacant |
| `GovernmentPopularity` | `float64` | Hidden true approval; 0-100 |

### Ministers and stakeholders
`Stakeholders` is `[]stakeholder.Stakeholder`. Relevant fields per entry:

| Field | Type |
|---|---|
| `State` | `stakeholder.MinisterState` (string enum) |
| `Popularity` | `float64` (0-100) |
| `RelationshipScore` | `float64` (0-100) |
| `IdeologyConflictScore` | `float64` (0-100) |
| `GraceWeeksRemaining` | `int` |
| `ConsultancyAffinity` | `[]string` (org IDs) |
| `ConsultancyAversion` | `bool` |

### Policy cards
`PolicyCards` is `[]policy.PolicyCard`. Relevant fields per entry:

| Field | Type |
|---|---|
| `Def.ID` | `string` |
| `Def.Name` | `string` |
| `Def.Description` | `string` |
| `Def.APCost` | `int` |
| `State` | `policy.PolicyState` (string enum) |
| `StepsCleared` | `int` |
| `WeeksUnderReview` | `int` |

### Player
`Player` is a `player.CivilServant`. Relevant fields:

| Field | Type |
|---|---|
| `Player.APRemaining` | `int` |
| `Player.Reputation` | `float64` (0-100) |
| `Player.Staff` | `[]player.StaffMember` |

Use `player.WeeklyAPPool(state.Player)` to get the total AP available this week (base 5 + staff bonuses). `Player.APRemaining` is decremented as actions are applied in `AdvanceWeek`.

### Events
- `EventLog` (`event.EventLog`) -- rolling 52-entry ring buffer. Call `state.EventLog.Entries()` to get events in chronological order.
- `firedEvents` -- the `[]event.EventEntry` second return value of `AdvanceWeek`; display these as notifications for the current week. Each entry has `DefID string`, `Name string`, `Week int`.

### Ticky pressure
| Field | Type | Notes |
|---|---|---|
| `PendingTickyPressure` | `bool` | Show the Ticky pressure UI prompt when true |

### Polling history
`PollHistory` is `[]polling.PollSnapshot`, capped at 200 entries. Each snapshot has a `PartyShares map[config.Party]float64` suitable for chart rendering. Index 0 is the oldest retained entry.

### Fuel poverty and tiles
`Tiles` is `[]region.Tile`. Per-tile fields of interest:

| Field | Type |
|---|---|
| `FuelPoverty` | `float64` (0-100) |
| `InsulationLevel` | `float64` (0-100) |
| `LocalPoliticalOpinion` | `float64` (0-100; 50 = neutral) |

### Evidence
- `OrgStates` (`[]evidence.OrgState`) -- per-org: `RelationshipScore float64`, `CoolingOffUntil int`, `MuricanUnlocked bool`.
- `Commissions` (`[]evidence.Commission`) -- active and completed commission records.
- `Reports` (`[]evidence.InsightReport`) -- delivered insight reports, capped at 50 entries.

### Industry
`Industry.Companies` is `map[string]industry.CompanyState`. Per-company fields of interest:

| Field | Type |
|---|---|
| `IsActive` | `bool` |
| `AccumulatedQuality` | `float64` |
| `WorkRate` | `float64` |

---

## 3. Player Actions

Actions are passed as `[]simulation.Action` to `AdvanceWeek`. Each `Action` has `Type player.ActionType`, `Target string`, and `Detail string`.

| ActionType constant | Target | Detail |
|---|---|---|
| `player.ActionTypeSubmitPolicy` | policy card ID | `""` |
| `player.ActionTypeCommissionReport` | org ID | `config.InsightType` string (e.g. `"POWER"`, `"CLIMATE"`) |
| `player.ActionTypeLobbyMinister` | stakeholder ID | `""` |
| `player.ActionTypeRespondTickyPressure` | `""` | `"ACCEPT"`, `"DECLINE"`, or `"NEGOTIATE"` |
| `player.ActionTypeShockResponse` | event def ID | ShockResponseOption string |
| `player.ActionTypeHireStaff` | staff role string | `""` |
| `player.ActionTypeFireStaff` | staff member ID | `""` |

---

## 4. Events Returned from AdvanceWeek

`AdvanceWeek` returns `(WorldState, []event.EventEntry)`. Each `EventEntry`:

```go
type EventEntry struct {
    DefID   string
    Name    string
    Week    int
    Effects ResolvedEffect
}
```

Display these as in-week notification banners. The same entries are also written into `EventLog` for persistent history access.

---

## 5. AP Economy

- Base pool: 5 AP per week.
- Staff bonuses: each hired `StaffMember.APBonus` is added. Call `player.WeeklyAPPool(state.Player)` for the total.
- AP is replenished at the start of each `AdvanceWeek` call. It is not cumulative across weeks.
- Most actions cost AP (see `Def.APCost` on policy cards; other actions have fixed costs in the simulation layer).
- Before rendering an action as available, call `player.SpendAP(state.Player, cost)` to check affordability; it returns `(updated CivilServant, ok bool)`. Use the `ok` result only -- do not apply the returned state until the action is submitted.

---

## 6. Save / Load Flow

### Saving
```go
ss, err := save.NewSaveState(playerName)  // new game: generates OS-entropy seed
ss.GameWeek = state.Week
ss.GameYear = state.Year
err = save.Write(path, ss)
```

### Loading
```go
ss, err := save.Read(path)
// Restore seed, rebuild world, replay:
w := simulation.NewWorld(cfg, ss.MasterSeed)
for i := 0; i < ss.GameWeek; i++ {
    w, _ = simulation.AdvanceWeek(w, nil)
}
// w is now at the saved week
```

### Design note
The save system is intentionally seed-based: it stores only `MasterSeed` and `GameWeek`. Loading replays up to `GameWeek` `AdvanceWeek` calls. At roughly 1 ms per call, a full 2080-week (40-year) game loads in approximately 2 seconds. A full `WorldState` serialisation system should be considered once struct churn stops (after the content-complete milestone). When implementing, use `encoding/json` with the versioned envelope already in place in `save.go` (`currentVersion` constant, `SaveState.Version` field).

---

## 7. Department IDs

Used as keys in `LastBudget.Departments` and in `government.MinisterStats.BudgetAllocationBias`.

| Constant | String value |
|---|---|
| `government.DeptPower` | `"power"` |
| `government.DeptTransport` | `"transport"` |
| `government.DeptBuildings` | `"buildings"` |
| `government.DeptIndustry` | `"industry"` |
| `government.DeptCross` | `"cross"` |

---

## 8. config.Technology Values

Used as the argument to `Tech.Maturity(config.Technology)` and as keys in tech-related maps.

| Constant | String value |
|---|---|
| `config.TechOffshoreWind` | `"OFFSHORE_WIND"` |
| `config.TechOnshoreWind` | `"ONSHORE_WIND"` |
| `config.TechSolarPV` | `"SOLAR_PV"` |
| `config.TechNuclear` | `"NUCLEAR"` |
| `config.TechHeatPumps` | `"HEAT_PUMPS"` |
| `config.TechEVs` | `"EVS"` |
| `config.TechHydrogen` | `"HYDROGEN"` |
| `config.TechIndustrialCCS` | `"INDUSTRIAL_CCS"` |
