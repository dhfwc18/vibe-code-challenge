package config

// techCurveDefs defines adoption S-curves for the eight tracked technologies.
// LogisticMidpoint is expressed as game weeks from start (week 0 = Jan 2010).
// BaseAdoptionRate is the weekly maturity gain with no active policy boosting it.
// InitialMaturity reflects deployment level at game start (2010).
// All values are calibrated to DESNZ electricity generation cost data.
var techCurveDefs = []TechCurveDef{
	{
		ID:                TechOffshoreWind,
		Name:              "Offshore Wind",
		Sector:            SectorPower,
		LogisticMidpoint:  520,  // ~week 520 = approx 2020
		LogisticSteepness: 0.01,
		BaseAdoptionRate:  0.08,
		InitialMaturity:   18.0, // meaningful but early-stage in 2010
	},
	{
		ID:                TechOnshoreWind,
		Name:              "Onshore Wind",
		Sector:            SectorPower,
		LogisticMidpoint:  400,  // ~2018; already established by 2010
		LogisticSteepness: 0.012,
		BaseAdoptionRate:  0.07,
		InitialMaturity:   30.0,
	},
	{
		ID:                TechSolarPV,
		Name:              "Solar PV",
		Sector:            SectorPower,
		LogisticMidpoint:  600,  // ~2021-22; cost collapse drives later surge
		LogisticSteepness: 0.013,
		BaseAdoptionRate:  0.09,
		InitialMaturity:   8.0, // minimal in 2010
	},
	{
		ID:                TechNuclear,
		Name:              "Nuclear",
		Sector:            SectorPower,
		LogisticMidpoint:  1040, // ~2030; long lead times
		LogisticSteepness: 0.006,
		BaseAdoptionRate:  0.02, // slow without explicit policy
		InitialMaturity:   40.0, // existing fleet; mature but ageing
	},
	{
		ID:                TechHeatPumps,
		Name:              "Heat Pumps",
		Sector:            SectorBuildings,
		LogisticMidpoint:  780,  // ~2025
		LogisticSteepness: 0.009,
		BaseAdoptionRate:  0.04,
		InitialMaturity:   5.0, // niche in 2010
	},
	{
		ID:                TechEVs,
		Name:              "Electric Vehicles",
		Sector:            SectorTransport,
		LogisticMidpoint:  700,  // ~2023
		LogisticSteepness: 0.011,
		BaseAdoptionRate:  0.06,
		InitialMaturity:   3.0, // negligible in 2010
	},
	{
		ID:                TechHydrogen,
		Name:              "Green Hydrogen",
		Sector:            SectorIndustry,
		LogisticMidpoint:  1200, // ~2033; emerging technology
		LogisticSteepness: 0.007,
		BaseAdoptionRate:  0.02,
		InitialMaturity:   2.0,
	},
	{
		ID:                TechIndustrialCCS,
		Name:              "Industrial CCS",
		Sector:            SectorIndustry,
		LogisticMidpoint:  1100, // ~2031
		LogisticSteepness: 0.007,
		BaseAdoptionRate:  0.02,
		InitialMaturity:   4.0,
	},
}
