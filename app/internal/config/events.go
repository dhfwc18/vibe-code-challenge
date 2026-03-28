package config

// eventDefs lists the global event deck available to the event engine.
// BaseProbability is the per-week draw chance. ClimateMultiplier and
// FossilMultiplier are applied when climate severity is elevated or
// the national fossil dependency score exceeds 60 respectively.
var eventDefs = []EventDef{

	// ------------------------------------------------------------------
	// WEATHER
	// ------------------------------------------------------------------

	{
		ID:                "polar_cold_snap",
		Name:              "Polar Cold Snap",
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
}
