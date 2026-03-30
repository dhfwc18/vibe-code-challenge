package config

// regionDefs lists the 12 Taitan regions. TileIDs must match entries in tileDefs.
// InitialSkillsNetwork, InitialSupplyChain are 0-100.
// InitialInstallerCapacity is installs per week at game start.
var regionDefs = []RegionDef{
	{
		ID:   "northern_highlands",
		Name: "Northern Highlands",
		TileIDs: []string{
			"nh_coastal",
			"nh_interior",
			"nh_islands",
		},
		Tags:                     []string{"coastal", "rural"},
		InitialSkillsNetwork:     30.0,
		InitialInstallerCapacity: 4.0,
		InitialSupplyChain:       28.0,
	},
	{
		ID:   "eastern_lowlands",
		Name: "Eastern Lowlands",
		TileIDs: []string{
			"el_capital_city",
			"el_harbour",
		},
		Tags:                     []string{"coastal", "urban"},
		InitialSkillsNetwork:     52.0,
		InitialInstallerCapacity: 8.0,
		InitialSupplyChain:       48.0,
	},
	{
		ID:   "northern_industrial",
		Name: "Northern Industrial Belt",
		TileIDs: []string{
			"ni_shipyards",
			"ni_coalfields",
			"ni_port_towns",
		},
		Tags:                     []string{"industrial", "coastal"},
		InitialSkillsNetwork:     44.0,
		InitialInstallerCapacity: 9.0,
		InitialSupplyChain:       42.0,
	},
	{
		ID:   "pennine_corridor",
		Name: "Pennine Corridor",
		TileIDs: []string{
			"pc_mill_towns",
			"pc_dales",
		},
		Tags:                     []string{"rural", "industrial"},
		InitialSkillsNetwork:     46.0,
		InitialInstallerCapacity: 10.0,
		InitialSupplyChain:       44.0,
	},
	{
		ID:   "north_west_cities",
		Name: "North West Cities",
		TileIDs: []string{
			"nw_metropolis",
			"nw_port_city",
			"nw_commuter_belt",
		},
		Tags:                     []string{"urban"},
		InitialSkillsNetwork:     55.0,
		InitialInstallerCapacity: 14.0,
		InitialSupplyChain:       52.0,
	},
	{
		ID:   "east_midlands",
		Name: "East Midlands",
		TileIDs: []string{
			"em_farmlands",
			"em_market_towns",
		},
		Tags:                     []string{"rural", "agricultural"},
		InitialSkillsNetwork:     42.0,
		InitialInstallerCapacity: 8.0,
		InitialSupplyChain:       40.0,
	},
	{
		ID:   "west_midlands",
		Name: "West Midlands",
		TileIDs: []string{
			"wm_conurbation",
			"wm_automotive",
		},
		Tags:                     []string{"industrial", "urban"},
		InitialSkillsNetwork:     50.0,
		InitialInstallerCapacity: 12.0,
		InitialSupplyChain:       50.0,
	},
	{
		ID:   "eastern_counties",
		Name: "Eastern Counties",
		TileIDs: []string{
			"ec_fens",
			"ec_coastal_plain",
		},
		Tags:                     []string{"coastal", "rural", "agricultural"},
		InitialSkillsNetwork:     38.0,
		InitialInstallerCapacity: 7.0,
		InitialSupplyChain:       36.0,
	},
	{
		ID:   "capital_region",
		Name: "Capital Region",
		TileIDs: []string{
			"cr_inner_city",
			"cr_suburbs",
			"cr_outer_commuter",
		},
		Tags:                     []string{"urban"},
		InitialSkillsNetwork:     72.0,
		InitialInstallerCapacity: 20.0,
		InitialSupplyChain:       68.0,
	},
	{
		ID:   "south_east",
		Name: "South East",
		TileIDs: []string{
			"se_commuter_towns",
			"se_coastal_resorts",
			"se_downs",
		},
		Tags:                     []string{"coastal", "urban"},
		InitialSkillsNetwork:     58.0,
		InitialInstallerCapacity: 13.0,
		InitialSupplyChain:       55.0,
	},
	{
		ID:   "south_west",
		Name: "South West",
		TileIDs: []string{
			"sw_peninsula",
			"sw_estuary",
		},
		Tags:                     []string{"coastal", "rural"},
		InitialSkillsNetwork:     40.0,
		InitialInstallerCapacity: 7.0,
		InitialSupplyChain:       37.0,
	},
	{
		ID:   "western_coast",
		Name: "Western Coast",
		TileIDs: []string{
			"wc_valleys",
			"wc_north_coast",
			"wc_border_towns",
		},
		Tags:                     []string{"coastal", "rural"},
		InitialSkillsNetwork:     36.0,
		InitialInstallerCapacity: 6.0,
		InitialSupplyChain:       33.0,
	},
}

// tileDefs lists the 30 Taitan map tiles.
// InitialLocalIncome: 50 = median Taitan household income.
// InitialPoliticalOpinion: 50 = neutral; >50 leans right; <50 leans left.
var tileDefs = []TileDef{
	// Northern Highlands
	{
		ID: "nh_coastal", RegionID: "northern_highlands", Name: "Highland Coast",
		InitialInsulationLevel: 28.0, InitialHeatingType: HeatingOil,
		InitialLocalIncome: 38.0, InitialPoliticalOpinion: 42.0,
		InitialHeatingCapacity: 45.0, InitialInstallerQuality: 32.0,
	},
	{
		ID: "nh_interior", RegionID: "northern_highlands", Name: "Highland Interior",
		InitialInsulationLevel: 22.0, InitialHeatingType: HeatingOil,
		InitialLocalIncome: 32.0, InitialPoliticalOpinion: 40.0,
		InitialHeatingCapacity: 38.0, InitialInstallerQuality: 26.0,
	},
	{
		ID: "nh_islands", RegionID: "northern_highlands", Name: "Outer Islands",
		InitialInsulationLevel: 18.0, InitialHeatingType: HeatingOil,
		InitialLocalIncome: 30.0, InitialPoliticalOpinion: 38.0,
		InitialHeatingCapacity: 30.0, InitialInstallerQuality: 22.0,
	},

	// Eastern Lowlands
	{
		ID: "el_capital_city", RegionID: "eastern_lowlands", Name: "Eastern Capital",
		InitialInsulationLevel: 45.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 58.0, InitialPoliticalOpinion: 45.0,
		InitialHeatingCapacity: 65.0, InitialInstallerQuality: 52.0,
	},
	{
		ID: "el_harbour", RegionID: "eastern_lowlands", Name: "Harbour Quarter",
		InitialInsulationLevel: 38.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 52.0, InitialPoliticalOpinion: 48.0,
		InitialHeatingCapacity: 58.0, InitialInstallerQuality: 46.0,
	},

	// Northern Industrial Belt
	{
		ID: "ni_shipyards", RegionID: "northern_industrial", Name: "Shipyard Towns",
		InitialInsulationLevel: 30.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 42.0, InitialPoliticalOpinion: 38.0,
		InitialHeatingCapacity: 50.0, InitialInstallerQuality: 40.0,
	},
	{
		ID: "ni_coalfields", RegionID: "northern_industrial", Name: "Former Coalfields",
		InitialInsulationLevel: 25.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 36.0, InitialPoliticalOpinion: 44.0,
		InitialHeatingCapacity: 44.0, InitialInstallerQuality: 35.0,
	},
	{
		ID: "ni_port_towns", RegionID: "northern_industrial", Name: "Port Towns",
		InitialInsulationLevel: 32.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 40.0, InitialPoliticalOpinion: 50.0,
		InitialHeatingCapacity: 48.0, InitialInstallerQuality: 38.0,
	},

	// Pennine Corridor
	{
		ID: "pc_mill_towns", RegionID: "pennine_corridor", Name: "Mill Towns",
		InitialInsulationLevel: 30.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 40.0, InitialPoliticalOpinion: 46.0,
		InitialHeatingCapacity: 50.0, InitialInstallerQuality: 42.0,
	},
	{
		ID: "pc_dales", RegionID: "pennine_corridor", Name: "Rural Dales",
		InitialInsulationLevel: 26.0, InitialHeatingType: HeatingMixed,
		InitialLocalIncome: 46.0, InitialPoliticalOpinion: 60.0,
		InitialHeatingCapacity: 40.0, InitialInstallerQuality: 34.0,
	},

	// North West Cities
	{
		ID: "nw_metropolis", RegionID: "north_west_cities", Name: "Northern Metropolis",
		InitialInsulationLevel: 40.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 52.0, InitialPoliticalOpinion: 36.0,
		InitialHeatingCapacity: 62.0, InitialInstallerQuality: 52.0,
	},
	{
		ID: "nw_port_city", RegionID: "north_west_cities", Name: "Northern Port City",
		InitialInsulationLevel: 35.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 44.0, InitialPoliticalOpinion: 34.0,
		InitialHeatingCapacity: 55.0, InitialInstallerQuality: 46.0,
	},
	{
		ID: "nw_commuter_belt", RegionID: "north_west_cities", Name: "North West Commuter Belt",
		InitialInsulationLevel: 44.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 58.0, InitialPoliticalOpinion: 56.0,
		InitialHeatingCapacity: 58.0, InitialInstallerQuality: 50.0,
	},

	// East Midlands
	{
		ID: "em_farmlands", RegionID: "east_midlands", Name: "East Midlands Farmlands",
		InitialInsulationLevel: 28.0, InitialHeatingType: HeatingMixed,
		InitialLocalIncome: 44.0, InitialPoliticalOpinion: 62.0,
		InitialHeatingCapacity: 40.0, InitialInstallerQuality: 36.0,
	},
	{
		ID: "em_market_towns", RegionID: "east_midlands", Name: "East Midlands Market Towns",
		InitialInsulationLevel: 35.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 48.0, InitialPoliticalOpinion: 58.0,
		InitialHeatingCapacity: 50.0, InitialInstallerQuality: 44.0,
	},

	// West Midlands
	{
		ID: "wm_conurbation", RegionID: "west_midlands", Name: "West Midlands Conurbation",
		InitialInsulationLevel: 38.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 46.0, InitialPoliticalOpinion: 44.0,
		InitialHeatingCapacity: 58.0, InitialInstallerQuality: 48.0,
	},
	{
		ID: "wm_automotive", RegionID: "west_midlands", Name: "Automotive Corridor",
		InitialInsulationLevel: 36.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 50.0, InitialPoliticalOpinion: 52.0,
		InitialHeatingCapacity: 55.0, InitialInstallerQuality: 46.0,
	},

	// Eastern Counties
	{
		ID: "ec_fens", RegionID: "eastern_counties", Name: "The Fens",
		InitialInsulationLevel: 24.0, InitialHeatingType: HeatingMixed,
		InitialLocalIncome: 40.0, InitialPoliticalOpinion: 64.0,
		InitialHeatingCapacity: 38.0, InitialInstallerQuality: 32.0,
	},
	{
		ID: "ec_coastal_plain", RegionID: "eastern_counties", Name: "Eastern Coastal Plain",
		InitialInsulationLevel: 30.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 44.0, InitialPoliticalOpinion: 60.0,
		InitialHeatingCapacity: 44.0, InitialInstallerQuality: 38.0,
	},

	// Capital Region
	{
		ID: "cr_inner_city", RegionID: "capital_region", Name: "Capital Inner City",
		InitialInsulationLevel: 42.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 68.0, InitialPoliticalOpinion: 28.0,
		InitialHeatingCapacity: 75.0, InitialInstallerQuality: 70.0,
	},
	{
		ID: "cr_suburbs", RegionID: "capital_region", Name: "Capital Suburbs",
		InitialInsulationLevel: 50.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 72.0, InitialPoliticalOpinion: 52.0,
		InitialHeatingCapacity: 70.0, InitialInstallerQuality: 66.0,
	},
	{
		ID: "cr_outer_commuter", RegionID: "capital_region", Name: "Outer Commuter Zone",
		InitialInsulationLevel: 48.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 65.0, InitialPoliticalOpinion: 58.0,
		InitialHeatingCapacity: 62.0, InitialInstallerQuality: 58.0,
	},

	// South East
	{
		ID: "se_commuter_towns", RegionID: "south_east", Name: "South East Commuter Towns",
		InitialInsulationLevel: 46.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 66.0, InitialPoliticalOpinion: 62.0,
		InitialHeatingCapacity: 62.0, InitialInstallerQuality: 58.0,
	},
	{
		ID: "se_coastal_resorts", RegionID: "south_east", Name: "Coastal Resorts",
		InitialInsulationLevel: 32.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 48.0, InitialPoliticalOpinion: 60.0,
		InitialHeatingCapacity: 48.0, InitialInstallerQuality: 42.0,
	},
	{
		ID: "se_downs", RegionID: "south_east", Name: "The Downs",
		InitialInsulationLevel: 40.0, InitialHeatingType: HeatingMixed,
		InitialLocalIncome: 62.0, InitialPoliticalOpinion: 65.0,
		InitialHeatingCapacity: 52.0, InitialInstallerQuality: 48.0,
	},

	// South West
	{
		ID: "sw_peninsula", RegionID: "south_west", Name: "South West Peninsula",
		InitialInsulationLevel: 30.0, InitialHeatingType: HeatingOil,
		InitialLocalIncome: 44.0, InitialPoliticalOpinion: 56.0,
		InitialHeatingCapacity: 40.0, InitialInstallerQuality: 36.0,
	},
	{
		ID: "sw_estuary", RegionID: "south_west", Name: "Estuary Towns",
		InitialInsulationLevel: 38.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 52.0, InitialPoliticalOpinion: 50.0,
		InitialHeatingCapacity: 52.0, InitialInstallerQuality: 46.0,
	},

	// Western Coast
	{
		ID: "wc_valleys", RegionID: "western_coast", Name: "The Valleys",
		InitialInsulationLevel: 28.0, InitialHeatingType: HeatingGas,
		InitialLocalIncome: 36.0, InitialPoliticalOpinion: 35.0,
		InitialHeatingCapacity: 44.0, InitialInstallerQuality: 36.0,
	},
	{
		ID: "wc_north_coast", RegionID: "western_coast", Name: "Western North Coast",
		InitialInsulationLevel: 24.0, InitialHeatingType: HeatingOil,
		InitialLocalIncome: 34.0, InitialPoliticalOpinion: 42.0,
		InitialHeatingCapacity: 38.0, InitialInstallerQuality: 30.0,
	},
	{
		ID: "wc_border_towns", RegionID: "western_coast", Name: "Border Towns",
		InitialInsulationLevel: 32.0, InitialHeatingType: HeatingMixed,
		InitialLocalIncome: 40.0, InitialPoliticalOpinion: 50.0,
		InitialHeatingCapacity: 46.0, InitialInstallerQuality: 38.0,
	},
}
