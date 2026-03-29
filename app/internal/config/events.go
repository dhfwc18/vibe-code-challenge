package config

// eventDefs lists the global event deck available to the event engine.
// BaseProbability is the per-week draw chance (0 for time-gated events).
// ClimateMultiplier and FossilMultiplier are applied when climate severity
// is elevated or the national fossil dependency score exceeds 60.
// TriggerAtYear fires the event automatically once at the start of that year.
// DecayingShock creates a persistent market effect that diminishes each week.
var eventDefs = []EventDef{

	// ------------------------------------------------------------------
	// WEATHER
	// ------------------------------------------------------------------

	{
		ID:                "polar_cold_snap",
		Name:              "Polar Cold Snap",
		Headline:          "Deep freeze grips nation as gas bills surge",
		EventType:         EventWeather,
		Severity:          SeverityModerate,
		BaseProbability:   0.010,
		ClimateMultiplier: 1.4,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			GasPriceDeltaPct:         8.0,
			ElectricityPriceDeltaPct: 5.0,
			TileFuelPovertyDelta:     3.0,
			CarbonEmissionsDeltaMt:   1.2,
		},
		Narrative:           "A prolonged polar vortex pushes gas demand above seasonal norms, straining the distribution network and driving up bills for the poorest households.",
		OffersShockResponse: false,
	},
	{
		ID:                "summer_heatwave",
		Name:              "Summer Heatwave",
		Headline:          "Record heat wave kills dozens; ministers face scrutiny",
		EventType:         EventWeather,
		Severity:          SeverityModerate,
		BaseProbability:   0.008,
		ClimateMultiplier: 1.8,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			ElectricityPriceDeltaPct: 6.0,
			GovtPopularityDelta:      -1.5,
			TileFuelPovertyDelta:     1.5,
			CarbonEmissionsDeltaMt:   0.4,
		},
		Narrative:           "Record temperatures overwhelm cooling demand. Hospital admissions climb; the opposition demands a national heat resilience plan.",
		OffersShockResponse: true,
	},
	{
		ID:                "coastal_flooding",
		Name:              "Coastal Flooding Event",
		Headline:          "Storm surge swamps coast towns; hundreds evacuated",
		EventType:         EventWeather,
		Severity:          SeverityMajor,
		BaseProbability:   0.005,
		ClimateMultiplier: 2.2,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			EconomyDelta:           -2.0,
			GovtPopularityDelta:    -3.0,
			LCRDelta:               2.0, // climate salience boosts net-zero sympathy
			CarbonEmissionsDeltaMt: 0.0,
			// Coastal regions lose installer capacity; insulation is physically damaged
			RegionFilter:           "COASTAL",
			InstallerCapacityDelta: -6.0,
			TileInsulationDamage:   3.0,
		},
		Narrative:           "Storm surge overwhelms sea defences along the eastern coast. Evacuation orders affect four coastal towns; infrastructure repair costs mount.",
		OffersShockResponse: true,
	},
	{
		ID:                "wind_drought",
		Name:              "Extended Wind Drought",
		Headline:          "Calm skies push electricity prices to quarterly high",
		EventType:         EventWeather,
		Severity:          SeverityMinor,
		BaseProbability:   0.012,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			GasPriceDeltaPct:         4.0,
			ElectricityPriceDeltaPct: 3.0,
			CarbonEmissionsDeltaMt:   0.8,
		},
		Narrative:           "Three weeks of unusually low wind across the country forces greater reliance on gas peakers, pushing up wholesale electricity prices.",
		OffersShockResponse: false,
	},

	// ------------------------------------------------------------------
	// ENERGY SHOCK
	// ------------------------------------------------------------------

	{
		ID:                "gas_supply_disruption",
		Name:              "Continental Gas Supply Disruption",
		Headline:          "Import pipeline fault cuts gas supply by a quarter",
		EventType:         EventEnergyShock,
		Severity:          SeverityMajor,
		BaseProbability:   0.004,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.6,
		BaseEffects: EventEffect{
			GasPriceDeltaPct:         22.0,
			ElectricityPriceDeltaPct: 14.0,
			OilPriceDeltaPct:         6.0,
			EconomyDelta:             -3.0,
			GovtPopularityDelta:      -4.0,
			TileFuelPovertyDelta:     6.0,
			CarbonEmissionsDeltaMt:   0.5,
			// Cabinet ministers come under parliamentary pressure
			StakeholderFilter:        "CABINET",
			StakeholderPressureDelta: 2,
		},
		Narrative:           "A pipeline fault and maintenance outage at a major Taitan import hub cuts continental gas supply by a quarter for six weeks.",
		OffersShockResponse: true,
	},
	{
		ID:                "oil_price_spike",
		Name:              "Global Oil Price Spike",
		Headline:          "Crude surges 18 percent on geopolitical tension",
		EventType:         EventEnergyShock,
		Severity:          SeverityModerate,
		BaseProbability:   0.007,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.4,
		BaseEffects: EventEffect{
			OilPriceDeltaPct:     18.0,
			GasPriceDeltaPct:     5.0,
			EconomyDelta:         -1.5,
			GovtPopularityDelta:  -2.0,
			TileFuelPovertyDelta: 2.5,
		},
		Narrative:           "Geopolitical tensions in a major oil-producing region send crude prices sharply higher, feeding into transport costs and home heating oil.",
		OffersShockResponse: false,
	},
	{
		ID:                "energy_price_cap_pressure",
		Name:              "Regulator Lifts Price Cap",
		Headline:          "Household energy bills to rise after regulator review",
		EventType:         EventEnergyShock,
		Severity:          SeverityModerate,
		BaseProbability:   0.006,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.2,
		BaseEffects: EventEffect{
			GasPriceDeltaPct:         10.0,
			ElectricityPriceDeltaPct: 8.0,
			GovtPopularityDelta:      -3.0,
			TileFuelPovertyDelta:     4.0,
		},
		Narrative:           "The energy regulator raises the household price cap in response to sustained wholesale market pressure, drawing immediate parliamentary criticism.",
		OffersShockResponse: true,
	},
	{
		ID:                "lng_terminal_fire",
		Name:              "LNG Terminal Fire",
		Headline:          "Fire at Taitan's largest gas terminal sends prices to four-year high",
		EventType:         EventEnergyShock,
		Severity:          SeverityMajor,
		BaseProbability:   0.003,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			GasPriceDeltaPct:         15.0,
			ElectricityPriceDeltaPct: 9.0,
			EconomyDelta:             -1.0,
			GovtPopularityDelta:      -2.5,
			TileFuelPovertyDelta:     3.0,
			CarbonEmissionsDeltaMt:   0.3,
		},
		Narrative:           "A fire at Taitan's largest LNG import terminal removes 8 bcm/year of import capacity during the repair period, driving gas spot prices to a four-year high.",
		OffersShockResponse: true,
	},

	// ------------------------------------------------------------------
	// INTERNATIONAL
	// ------------------------------------------------------------------

	{
		ID:                "murican_tariff_threat",
		Name:              "Murica Threatens Green Tech Tariffs",
		Headline:          "Murica signals punitive tariffs on clean energy exports",
		EventType:         EventInternational,
		Severity:          SeverityModerate,
		BaseProbability:   0.005,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			LCRDelta:             -3.0,
			EconomyDelta:         -1.5,
			GovtPopularityDelta:  -1.0,
			// All active LCT companies face slower order books under tariff uncertainty
			CompanyFilter:        "ALL",
			CompanyWorkRateDelta: -5.0,
		},
		Narrative:           "The Murican administration signals punitive tariffs on Taitan-made clean energy equipment, citing 'market distortion'. Domestic manufacturers lobby for retaliation.",
		OffersShockResponse: true,
	},
	{
		ID:                "international_climate_summit",
		Name:              "International Climate Summit",
		Headline:          "Taitan hailed as climate leader at international summit",
		EventType:         EventInternational,
		Severity:          SeverityMinor,
		BaseProbability:   0.008,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			LCRDelta:            3.0,
			GovtPopularityDelta: 1.5,
		},
		Narrative:           "Taitan hosts a bilateral climate summit. A joint communique on carbon pricing and technology cooperation draws favourable international coverage.",
		OffersShockResponse: false,
	},
	{
		ID:                "murican_fossil_subsidy_expansion",
		Name:              "Murica Expands Fossil Subsidies",
		Headline:          "Murican fossil package floods market with cheap energy",
		EventType:         EventInternational,
		Severity:          SeverityModerate,
		BaseProbability:   0.006,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.3,
		BaseEffects: EventEffect{
			OilPriceDeltaPct:    -8.0,
			GasPriceDeltaPct:    -4.0,
			LCRDelta:            -2.0,
			GovtPopularityDelta: -1.0,
		},
		Narrative:           "Murica announces a major new fossil fuel subsidy package. Cheap imports undercut domestic LCT investment and embolden the FarRight to demand policy reversal.",
		OffersShockResponse: false,
	},
	{
		ID:                "foreign_green_tech_investment",
		Name:              "Foreign Green Tech Investment",
		Headline:          "Overseas gigafactory to create 8,000 clean energy jobs",
		EventType:         EventInternational,
		Severity:          SeverityMinor,
		BaseProbability:   0.007,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			LCRDelta:               2.5,
			EconomyDelta:           1.5,
			GovtPopularityDelta:    1.0,
			// Gigafactory landing in the Northern Industrial Belt boosts local capacity
			RegionFilter:           "northern_industrial",
			InstallerCapacityDelta: 5.0,
			SkillsNetworkDelta:     3.0,
		},
		Narrative:           "A major overseas manufacturer announces a gigafactory site in the Northern Reaches, citing Taitan's policy stability as the decisive factor.",
		OffersShockResponse: false,
	},

	// ------------------------------------------------------------------
	// ECONOMIC
	// ------------------------------------------------------------------

	{
		ID:                "recession_shock",
		Name:              "Economic Recession",
		Headline:          "Taitan enters recession as GDP falls for second quarter",
		EventType:         EventEconomic,
		Severity:          SeverityMajor,
		BaseProbability:   0.003,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			EconomyDelta:           -6.0,
			GovtPopularityDelta:    -5.0,
			TileFuelPovertyDelta:   3.0,
			CarbonEmissionsDeltaMt: -1.5, // lower activity reduces emissions short-term
		},
		Narrative:           "GDP contracts for two consecutive quarters. Treasury revises down tax receipts; the Chancellor faces pressure to delay green spending commitments.",
		OffersShockResponse: true,
	},
	{
		ID:                "supply_chain_crunch",
		Name:              "Clean Energy Supply Chain Crunch",
		Headline:          "Material shortages stall dozens of clean energy projects",
		EventType:         EventEconomic,
		Severity:          SeverityModerate,
		BaseProbability:   0.006,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			EconomyDelta: -1.0,
			LCRDelta:     -1.5,
			// All LCT companies face project delays
			CompanyFilter:        "ALL",
			CompanyWorkRateDelta: -8.0,
		},
		Narrative:           "Global shortages of key materials (transformers, heat pump refrigerants) cause project delays and cost overruns across the Taitan LCT sector.",
		OffersShockResponse: false,
	},
	{
		ID:                "bond_market_wobble",
		Name:              "Bond Market Credibility Concerns",
		Headline:          "Bond yields spike as markets question fiscal discipline",
		EventType:         EventEconomic,
		Severity:          SeverityMajor,
		BaseProbability:   0.002,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			EconomyDelta:        -4.0,
			GovtPopularityDelta: -3.0,
		},
		Narrative:           "Yields on Taitan government bonds spike after markets react to fiscal policy announcements, forcing the Chancellor to signal urgent spending restraint.",
		OffersShockResponse: true,
	},

	// ------------------------------------------------------------------
	// SOCIAL
	// ------------------------------------------------------------------

	{
		ID:                "fuel_poverty_protest",
		Name:              "National Fuel Poverty Protests",
		Headline:          "Thousands march demanding action on soaring energy bills",
		EventType:         EventSocial,
		Severity:          SeverityModerate,
		BaseProbability:   0.006,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.2,
		BaseEffects: EventEffect{
			GovtPopularityDelta: -4.0,
			LCRDelta:            1.0, // public demand for green solutions rises
		},
		Narrative:           "A coalition of charities and trade unions organises marches in twelve cities demanding government action on household energy bills.",
		OffersShockResponse: true,
	},
	{
		ID:                "planning_revolt",
		Name:              "Local Planning Revolts",
		Headline:          "Residents block wind cable route and two solar farms",
		EventType:         EventSocial,
		Severity:          SeverityMinor,
		BaseProbability:   0.010,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			GovtPopularityDelta: -2.0,
			LCRDelta:            -1.0,
		},
		Narrative:           "NIMBY campaigns block planning permission for a new offshore wind cable route and two solar farms, citing landscape impact and property values.",
		OffersShockResponse: false,
	},
	{
		ID:                "net_zero_public_backlash",
		Name:              "Net Zero Public Backlash",
		Headline:          "Viral campaign frames net zero as attack on ordinary families",
		EventType:         EventSocial,
		Severity:          SeverityModerate,
		BaseProbability:   0.005,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.3,
		BaseEffects: EventEffect{
			GovtPopularityDelta: -3.5,
			LCRDelta:            -2.0,
		},
		Narrative:           "A viral campaign frames net zero policies as a cost-of-living attack on working families. FarRight politicians amplify the messaging across social media.",
		OffersShockResponse: true,
	},

	// ------------------------------------------------------------------
	// TECHNOLOGICAL
	// ------------------------------------------------------------------

	{
		ID:                "battery_breakthrough",
		Name:              "Battery Cost Breakthrough",
		Headline:          "Taitan lab cuts battery cost by 30 percent",
		EventType:         EventTech,
		Severity:          SeverityMinor,
		BaseProbability:   0.004,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			LCRDelta:            3.0,
			EconomyDelta:        1.0,
			GovtPopularityDelta: 1.5,
			// EV companies directly benefit from faster cost reduction
			CompanyFilter:        "TECH:EVS",
			CompanyWorkRateDelta: 8.0,
		},
		Narrative:           "A Taitan research consortium announces a 30% reduction in battery cell manufacturing cost, accelerating the business case for grid storage and EVs.",
		OffersShockResponse: false,
	},
	{
		ID:                "cyber_attack_grid",
		Name:              "Cyber Attack on Grid Infrastructure",
		Headline:          "Hackers breach grid systems; brief blackouts reported",
		EventType:         EventTech,
		Severity:          SeverityMajor,
		BaseProbability:   0.003,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.0,
		BaseEffects: EventEffect{
			ElectricityPriceDeltaPct: 7.0,
			EconomyDelta:             -2.0,
			GovtPopularityDelta:      -3.0,
		},
		Narrative:           "A sophisticated cyber attack disrupts smart meter communications and SCADA systems at two regional distribution operators, causing brief supply interruptions.",
		OffersShockResponse: true,
	},

	// ------------------------------------------------------------------
	// INTERNATIONAL -- decaying energy war shocks
	// The Coming Winter and The Amber Coast War are high-impact, low-probability
	// events. When they fire they produce an immediate price spike (BaseEffects)
	// AND create a persistent decaying market shock (DecayingShock) that
	// continues to depress/inflate prices for up to two years afterwards.
	// Both events are named to be clearly fictional and disconnected from
	// any real-world conflict; the intent is to model energy supply disruption
	// mechanics only.
	// ------------------------------------------------------------------

	{
		ID:                "the_coming_winter",
		Name:              "The Coming Winter",
		Headline:          "War in the East chokes off continental gas supply",
		EventType:         EventInternational,
		Severity:          SeverityMajor,
		BaseProbability:   0.002,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.5,
		BaseEffects: EventEffect{
			GasPriceDeltaPct:         18.0,
			ElectricityPriceDeltaPct: 10.0,
			OilPriceDeltaPct:         4.0,
			EconomyDelta:             -3.5,
			GovtPopularityDelta:      -4.0,
			TileFuelPovertyDelta:     7.0,
			CarbonEmissionsDeltaMt:   0.6,
			StakeholderFilter:        "CABINET",
			StakeholderPressureDelta: 2,
		},
		DecayingShock: DecayingShockConfig{
			InitialGasPctPerWeek:  2.5,
			InitialOilPctPerWeek:  0.3,
			InitialElecPctPerWeek: 1.2,
			DecayRate:             0.93,
			MaxWeeks:              104,
		},
		Narrative:           "Armed conflict between Northland and the Gold Trident severs a critical gas transit corridor. Spot gas prices soar as Taitan scrambles for alternative LNG supplies; energy poverty spreads rapidly through the northern regions.",
		OffersShockResponse: true,
	},
	{
		ID:                "the_amber_coast_war",
		Name:              "The Amber Coast War",
		Headline:          "Far-shore conflict sends oil markets into turmoil",
		EventType:         EventInternational,
		Severity:          SeverityMajor,
		BaseProbability:   0.002,
		ClimateMultiplier: 1.0,
		FossilMultiplier:  1.5,
		BaseEffects: EventEffect{
			OilPriceDeltaPct:         25.0,
			GasPriceDeltaPct:         5.0,
			ElectricityPriceDeltaPct: 3.0,
			EconomyDelta:             -4.0,
			GovtPopularityDelta:      -3.5,
			TileFuelPovertyDelta:     5.0,
			LCRDelta:                 -1.5,
		},
		DecayingShock: DecayingShockConfig{
			InitialGasPctPerWeek:  0.5,
			InitialOilPctPerWeek:  3.5,
			InitialElecPctPerWeek: 0.6,
			DecayRate:             0.93,
			MaxWeeks:              78,
		},
		Narrative:           "Military intervention by Murica in the Amber Coast region disrupts shipping lanes and triggers panic buying in crude markets. Taitan's transport costs spike; the Chancellor suspends climate spending reviews pending a full fiscal reassessment.",
		OffersShockResponse: true,
	},

	// ------------------------------------------------------------------
	// SOCIAL -- time-gated pandemic event
	// The Great Sneeze fires automatically at the start of 2019. It models
	// the effect of a sudden national health crisis on government capacity and
	// public finances, without referencing any real event. BaseProbability=0
	// prevents it from firing probabilistically; TriggerAtYear=2019 fires it once.
	// ------------------------------------------------------------------

	{
		ID:              "great_sneeze",
		Name:            "The Great Sneeze",
		Headline:        "National health emergency declared as illness sweeps Taitan",
		EventType:       EventSocial,
		Severity:        SeverityMajor,
		BaseProbability: 0.0,
		TriggerAtYear:   2019,
		BaseEffects: EventEffect{
			GovtPopularityDelta:  -3.0,
			EconomyDelta:         -4.0,
			TileFuelPovertyDelta: 2.5,
			LCRDelta:             -2.0,
			// All companies face operational disruption
			CompanyFilter:        "ALL",
			CompanyWorkRateDelta: -15.0,
		},
		Narrative:           "A rapidly spreading illness forces closure of workplaces and public spaces across Taitan. Parliament passes emergency spending legislation; the government's popularity plummets as the scale of the crisis becomes clear. Emergency measures will remain in place until the situation stabilises.",
		OffersShockResponse: false,
	},
}
