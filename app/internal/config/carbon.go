package config

// carbonBudgets lists the annual carbon limits derived from CCC carbon budget targets.
// Source: docs/green_book_reference.md / CCC Seventh Carbon Budget (March 2025).
// Years not listed use linear interpolation between the nearest entries.
// All values in MtCO2e per year.
var carbonBudgets = []CarbonBudgetEntry{
	// 2010-2022: approximate actuals used as calibration anchors
	{Year: 2010, AnnualLimitMtCO2e: 590.0}, // game start baseline
	{Year: 2015, AnnualLimitMtCO2e: 500.0},
	{Year: 2019, AnnualLimitMtCO2e: 450.0},
	{Year: 2022, AnnualLimitMtCO2e: 390.0}, // approx actual
	// CB4: 2023-2027 total 583 MtCO2e -> avg 117/yr
	{Year: 2023, AnnualLimitMtCO2e: 130.0},
	{Year: 2024, AnnualLimitMtCO2e: 122.0},
	{Year: 2025, AnnualLimitMtCO2e: 118.0},
	{Year: 2026, AnnualLimitMtCO2e: 115.0},
	{Year: 2027, AnnualLimitMtCO2e: 112.0},
	// CB5: 2028-2032 total 535 MtCO2e -> avg 107/yr
	{Year: 2028, AnnualLimitMtCO2e: 110.0},
	{Year: 2029, AnnualLimitMtCO2e: 108.0},
	{Year: 2030, AnnualLimitMtCO2e: 107.0},
	{Year: 2031, AnnualLimitMtCO2e: 106.0},
	{Year: 2032, AnnualLimitMtCO2e: 104.0},
	// CB6: 2033-2037 -- 78% below 1990 (800 MtCO2e) by 2035 -> 176 MtCO2e/yr
	{Year: 2033, AnnualLimitMtCO2e: 195.0},
	{Year: 2034, AnnualLimitMtCO2e: 185.0},
	{Year: 2035, AnnualLimitMtCO2e: 176.0}, // statutory 78% milestone
	{Year: 2036, AnnualLimitMtCO2e: 165.0},
	{Year: 2037, AnnualLimitMtCO2e: 155.0},
	// CB7: 2038-2042 total 535 MtCO2e -> avg 107/yr
	{Year: 2038, AnnualLimitMtCO2e: 130.0},
	{Year: 2039, AnnualLimitMtCO2e: 115.0},
	{Year: 2040, AnnualLimitMtCO2e: 107.0},
	{Year: 2041, AnnualLimitMtCO2e: 95.0},
	{Year: 2042, AnnualLimitMtCO2e: 88.0},
	// 2043-2050: steep ramp to net zero
	{Year: 2043, AnnualLimitMtCO2e: 70.0},
	{Year: 2044, AnnualLimitMtCO2e: 55.0},
	{Year: 2045, AnnualLimitMtCO2e: 40.0},
	{Year: 2046, AnnualLimitMtCO2e: 28.0},
	{Year: 2047, AnnualLimitMtCO2e: 18.0},
	{Year: 2048, AnnualLimitMtCO2e: 10.0},
	{Year: 2049, AnnualLimitMtCO2e: 4.0},
	{Year: 2050, AnnualLimitMtCO2e: 0.0}, // net zero target
}
