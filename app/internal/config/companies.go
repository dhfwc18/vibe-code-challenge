package config

// companyDefs lists the 15 LCT companies present at game start.
// BaseQuality and BaseWorkRate are seeded here; actual values per playthrough are
// derived by the industry package using the master seed for per-run variation.
var companyDefs = []CompanyDef{
	// Offshore Wind
	{
		ID: "arclight_offshore", Name: "ArcLight Offshore", TechCategory: TechCatOffshoreWind,
		OriginSize: CompanyLarge, BaseQuality: 72, BaseWorkRate: 68, TaitonHQ: true,
	},
	{
		ID: "albion_wind_power", Name: "Albion Wind Power", TechCategory: TechCatOffshoreWind,
		OriginSize: CompanySME, BaseQuality: 58, BaseWorkRate: 55, TaitonHQ: true,
	},

	// Onshore Wind / Solar
	{
		ID: "greenfield_power", Name: "Greenfield Power", TechCategory: TechCatOnshore,
		OriginSize: CompanyLarge, BaseQuality: 65, BaseWorkRate: 62, TaitonHQ: true,
	},
	{
		ID: "solarion_taitan", Name: "Solarion Taitan", TechCategory: TechCatOnshore,
		OriginSize: CompanySME, BaseQuality: 55, BaseWorkRate: 60, TaitonHQ: false,
	},

	// Heat Pumps
	{
		ID: "thermacore_systems", Name: "ThermaCore Systems", TechCategory: TechCatHeatPumps,
		OriginSize: CompanySME, BaseQuality: 62, BaseWorkRate: 58, TaitonHQ: true,
	},
	{
		ID: "heatwave_technologies", Name: "HeatWave Technologies", TechCategory: TechCatHeatPumps,
		OriginSize: CompanyStartup, BaseQuality: 48, BaseWorkRate: 52, TaitonHQ: true,
	},

	// EVs
	{
		ID: "volta_motors_taitan", Name: "Volta Motors Taitan", TechCategory: TechCatEVs,
		OriginSize: CompanyLarge, BaseQuality: 70, BaseWorkRate: 65, TaitonHQ: true,
	},
	{
		ID: "taitdrive", Name: "TaitDrive", TechCategory: TechCatEVs,
		OriginSize: CompanySME, BaseQuality: 52, BaseWorkRate: 58, TaitonHQ: true,
	},

	// Hydrogen
	{
		ID: "hydrovolt_energy", Name: "HydroVolt Energy", TechCategory: TechCatHydrogen,
		OriginSize: CompanyStartup, BaseQuality: 50, BaseWorkRate: 45, TaitonHQ: true,
	},
	{
		ID: "greenstream_h2", Name: "GreenStream H2", TechCategory: TechCatHydrogen,
		OriginSize: CompanySME, BaseQuality: 55, BaseWorkRate: 48, TaitonHQ: false,
	},

	// CCUS
	{
		ID: "carbonseal_group", Name: "CarbonSeal Group", TechCategory: TechCatCCUS,
		OriginSize: CompanyLarge, BaseQuality: 68, BaseWorkRate: 55, TaitonHQ: true,
	},
	{
		ID: "deepstore_ccs", Name: "DeepStore CCS", TechCategory: TechCatCCUS,
		OriginSize: CompanySME, BaseQuality: 58, BaseWorkRate: 50, TaitonHQ: true,
	},

	// Grid / Retail
	{
		ID: "gridnorth_taitan", Name: "GridNorth Taitan", TechCategory: TechCatGrid,
		OriginSize: CompanyMultinational, BaseQuality: 75, BaseWorkRate: 70, TaitonHQ: false,
	},
	{
		ID: "cleanwatts_energy", Name: "CleanWatts Energy", TechCategory: TechCatGrid,
		OriginSize: CompanySME, BaseQuality: 60, BaseWorkRate: 62, TaitonHQ: true,
	},

	// Legacy Transition
	{
		ID: "northgate_energy", Name: "Northgate Energy", TechCategory: TechCatLegacy,
		OriginSize: CompanyMultinational, BaseQuality: 65, BaseWorkRate: 75, TaitonHQ: true,
	},
}
