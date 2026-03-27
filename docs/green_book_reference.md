# Green Book Reference Data

Source: HM Treasury Green Book (2022), DESNZ Valuation of Energy Use and GHG Emissions
(November 2023 update), CCC Seventh Carbon Budget (March 2025).

All values used as the authoritative numerical basis for game mechanics.

---

## UK Emissions Baseline

- 1990 territorial GHG emissions: 800 MtCO2e/year (statutory baseline)
- 2010 emissions (game start): approx 590 MtCO2e/year (approx 26% below 1990)
- 2023 emissions: approx 390 MtCO2e/year (51% below 1990)
- Net zero target: 0 MtCO2e net by 2050
- All figures measured in MtCO2e using IPCC AR5 GWP-100 values

---

## Carbon Budgets (legally binding 5-year caps)

| Budget | Period    | Total (MtCO2e) | Implied annual avg (MtCO2e) |
|--------|-----------|----------------|-----------------------------|
| CB4    | 2023-2027 | 583            | 117                         |
| CB5    | 2028-2032 | 535            | 107                         |
| CB6    | 2033-2037 | 78% below 1990 | approx 176                  |
| CB7    | 2038-2042 | 535            | 107                         |

Interim milestone: 78% reduction vs 1990 by 2035 = approx 176 MtCO2e/year.

---

## Shadow Price of Carbon (Non-Traded Sectors)

Source: DESNZ November 2023. Price base: GBP 2020 real.
Applies to: transport, buildings, agriculture, waste, smaller industry.
Growth rate post-2050: 1.5% real per annum.

| Year | Low (GBP/tCO2e) | Central (GBP/tCO2e) | High (GBP/tCO2e) |
|------|-----------------|---------------------|------------------|
| 2020 | 120             | 241                 | 361              |
| 2025 | 130             | 260                 | 390              |
| 2030 | 140             | 280                 | 420              |
| 2035 | 151             | 302                 | 453              |
| 2040 | 163             | 326                 | 489              |
| 2045 | 176             | 351                 | 527              |
| 2050 | 189             | 378                 | 568              |

Key calibration point: 2040 central = GBP 326/tCO2e (GloCaF model anchor).
Interpolation between years: approximately GBP 2/tCO2e per year increase (central).

Pre-2021 note: before the 2021 revision, central values were approx GBP 55/tCO2e (2011
prices) for 2010 -- roughly 4x lower. The revision reflected adoption of the 1.5C pathway.
Game uses post-2021 methodology throughout for consistency.

---

## Social Discount Rate (Green Book STPR)

Formula: STPR = delta + L + (mu * g) = 0.5 + 1.0 + (1.0 * 2.0) = 3.5%

| Time horizon | Discount rate |
|--------------|---------------|
| Years 0-30   | 3.5%          |
| Years 31-75  | 3.0%          |
| Years 76-125 | 2.5%          |
| Years 126+   | 2.0%          |

Used to compute NPV of policy interventions in game mechanics.

---

## Key Sectors and Decarbonisation Role

| Sector            | 2010 share (approx) | Route to zero            | Residual in 2050 |
|-------------------|---------------------|--------------------------|------------------|
| Power             | 27%                 | Wind, solar, nuclear     | Near zero        |
| Surface transport | 22%                 | EV transition, rail      | Near zero        |
| Buildings         | 17%                 | Heat pumps, insulation   | Near zero        |
| Industry (heavy)  | 17%                 | Electrification, CCS     | Small residual   |
| Agriculture       | 10%                 | Efficiency, diet shift   | 20-30 MtCO2e/yr  |
| Waste             | 4%                  | Landfill reduction       | Small residual   |
| Land use          | 3%                  | Carbon sinks (offsetting)| Net negative     |

Agriculture and land use are the hardest sectors to fully abate. Net zero requires
land-based sinks and engineered removals (BECCS, DACCS) to offset residual agriculture.

---

## Technology Cost Anchors

Source: DESNZ Electricity Generation Costs 2023 (GBP 2023 prices).

| Technology              | Cost range        | Unit    |
|-------------------------|-------------------|---------|
| Offshore wind           | 44-57             | GBP/MWh |
| Onshore wind            | 33-44             | GBP/MWh |
| Solar PV (utility)      | 25-40             | GBP/MWh |
| Nuclear                 | 125-160           | GBP/MWh |
| Gas CCGT (unabated)     | 60-80             | GBP/MWh |
| Gas with CCS            | 85-120            | GBP/MWh |
| Heat pump premium       | 5,000-10,000      | GBP/home|
| Industrial CCS          | 150-250           | GBP/tCO2|
| DACCS (direct air)      | 300-500 (target 200 by 2030s) | GBP/tCO2 |

---

## Whole-Economy Cost Reference

- Total net cost of UK net zero 2020-2050: GBP 321 billion
- Gross investment: GBP 1,312 billion
- Net operating savings: GBP 991 billion
- As share of GDP: approx 0.5-1% of cumulative GDP
- Annual government budget envelope implied: approx GBP 10-15 billion/year average

---

## GWP Values (IPCC AR5, used for CO2e conversion)

| Gas             | GWP-100 |
|-----------------|---------|
| CO2             | 1       |
| Methane (CH4)   | 28-30   |
| Nitrous oxide   | 265     |
| SF6             | 23,500  |
| HFCs            | varies  |
