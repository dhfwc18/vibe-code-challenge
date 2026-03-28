package region

import (
	"github.com/vibe-code-challenge/app/internal/config"
	"github.com/vibe-code-challenge/app/internal/technology"
)

// FuelPovertyInput bundles all inputs to the fuel poverty formula for one tile
// in one week.
type FuelPovertyInput struct {
	GasPrice              float64            // GBP/MWh
	ElectricityPrice      float64            // GBP/MWh
	OilPrice              float64            // GBP/MWh; used only for OIL heating
	HeatingType           config.HeatingType
	InsulationLevel       float64 // 0-100
	LocalIncome           float64 // 0-100 (50 = median Taitan household)
	HeatingCapacity       float64 // 0-100
	TechMaturityHeatPumps float64 // 0-100; used to compute COP for HeatPump type
	SeasonalMultiplier    float64 // 1.0 baseline; >1 winter, <1 summer
}

// Calibration constants. Tuned so a median-income, half-insulated tile on gas
// at 2010 reference prices (GBP 25/MWh) starts at approximately 35 fuel poverty.
const (
	baseHeatingDemand    = 100.0 // arbitrary demand units per week at reference
	povertyScalingWeight = 0.70  // scales cost/income ratio into 0-100 index
	minLocalIncome       = 0.01  // prevents division by zero
)

// ComputeFuelPoverty returns the fuel poverty index (0-100) for a tile.
//
// Formula:
//
//	insulationFactor = 1 - (insulationLevel / 200)   // 1.0 at 0%, 0.5 at 100%
//	heatingDemand    = baseHeatingDemand * insulationFactor * seasonalMultiplier
//	costPerUnit      = price for HeatingType (COP-adjusted for HeatPump)
//	FuelPoverty      = clamp((costPerUnit * heatingDemand / localIncome) * weight, 0, 100)
func ComputeFuelPoverty(in FuelPovertyInput) float64 {
	insulationFactor := 1.0 - (clamp(in.InsulationLevel, 0, 100) / 200.0)
	heatingDemand := baseHeatingDemand * insulationFactor * in.SeasonalMultiplier
	costPerUnit := heatingCostPerUnit(in)
	income := in.LocalIncome
	if income < minLocalIncome {
		income = minLocalIncome
	}
	return clamp((costPerUnit*heatingDemand/income)*povertyScalingWeight, 0, 100)
}

// heatingCostPerUnit returns the effective energy price per demand unit for
// a given heating type. Heat pumps are divided by COP to reflect efficiency gains.
func heatingCostPerUnit(in FuelPovertyInput) float64 {
	switch in.HeatingType {
	case config.HeatingGas:
		return in.GasPrice
	case config.HeatingOil:
		return in.OilPrice
	case config.HeatingElectric:
		return in.ElectricityPrice
	case config.HeatingHeatPump:
		cop := technology.HeatPumpCOP(in.TechMaturityHeatPumps)
		return in.ElectricityPrice / cop
	case config.HeatingMixed:
		return (in.GasPrice + in.ElectricityPrice) * 0.5
	default:
		return in.GasPrice
	}
}
