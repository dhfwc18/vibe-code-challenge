package energy

import (
	"twenty-fifty/internal/config"
	"twenty-fifty/internal/mathutil"
)

// Historical price anchors from Green Book / Ofgem data at game start (2010).
const (
	Anchor2010GasPriceMWh         = 25.0 // GBP/MWh
	Anchor2010ElectricityPriceMWh = 50.0 // GBP/MWh
	Anchor2010OilPriceBarrel      = 80.0 // GBP/barrel
	Anchor2010RenewableGridShare  = 15.0 // percent

	// GridCouplingThreshold is the renewable grid share below which electricity
	// price is coupled to gas. Above this share the coupling floor lifts.
	GridCouplingThreshold = 40.0

	// gasToElecConversionFactor is the approximate combined heat-rate and
	// distribution markup that converts gas price to an electricity floor price.
	gasToElecConversionFactor = 2.5
)

// PriceRing is a 52-slot circular buffer that tracks one year of weekly prices.
// All operations return new values; the input is never mutated.
type PriceRing struct {
	Values [52]float64
	Head   int // index of the most recently written slot
}

// PushPrice adds a new weekly price to the ring and advances the head.
func PushPrice(ring PriceRing, v float64) PriceRing {
	ring.Head = (ring.Head + 1) % 52
	ring.Values[ring.Head] = v
	return ring
}

// PriceAt returns the price n weeks ago (0 = most recent, 51 = oldest).
// n is clamped to [0, 51].
func PriceAt(ring PriceRing, n int) float64 {
	if n < 0 {
		n = 0
	}
	if n > 51 {
		n = 51
	}
	idx := ((ring.Head - n) % 52 + 52) % 52
	return ring.Values[idx]
}

// EnergyMarket holds the live state of the Taitan energy market.
type EnergyMarket struct {
	GasPrice           float64 // GBP/MWh
	ElectricityPrice   float64 // GBP/MWh
	OilPrice           float64 // GBP/barrel
	RenewableGridShare float64 // 0-100 percent
	GasHistory         PriceRing
	ElecHistory        PriceRing
	OilHistory         PriceRing
}

// NewMarket seeds the energy market from 2010 Green Book anchors.
// All 52 history slots are pre-filled with anchor values.
func NewMarket() EnergyMarket {
	var m EnergyMarket
	m.GasPrice = Anchor2010GasPriceMWh
	m.ElectricityPrice = Anchor2010ElectricityPriceMWh
	m.OilPrice = Anchor2010OilPriceBarrel
	m.RenewableGridShare = Anchor2010RenewableGridShare
	for i := range m.GasHistory.Values {
		m.GasHistory.Values[i] = Anchor2010GasPriceMWh
		m.ElecHistory.Values[i] = Anchor2010ElectricityPriceMWh
		m.OilHistory.Values[i] = Anchor2010OilPriceBarrel
	}
	return m
}

// TickPrices advances the energy market by one week.
//
// gasDeltaPct, elecDeltaPct, oilDeltaPct are percentage changes from all
// sources (events, carbon levy, renewable subsidy) summed before calling.
// gridShareDelta is the change in renewable grid share this week.
//
// The grid coupling floor is enforced after all deltas are applied.
// Prices are clamped to [0, 9999]. Returns a new EnergyMarket.
func TickPrices(m EnergyMarket, gasDeltaPct, elecDeltaPct, oilDeltaPct, gridShareDelta float64) EnergyMarket {
	m.GasPrice = mathutil.Clamp(m.GasPrice*(1+gasDeltaPct/100), 0, 9999)
	m.ElectricityPrice = mathutil.Clamp(m.ElectricityPrice*(1+elecDeltaPct/100), 0, 9999)
	m.OilPrice = mathutil.Clamp(m.OilPrice*(1+oilDeltaPct/100), 0, 9999)
	m.RenewableGridShare = mathutil.Clamp(m.RenewableGridShare+gridShareDelta, 0, 100)

	// Enforce grid coupling floor after all deltas.
	floor := GridCouplingFloor(m.GasPrice, m.RenewableGridShare)
	if m.ElectricityPrice < floor {
		m.ElectricityPrice = floor
	}

	m.GasHistory = PushPrice(m.GasHistory, m.GasPrice)
	m.ElecHistory = PushPrice(m.ElecHistory, m.ElectricityPrice)
	m.OilHistory = PushPrice(m.OilHistory, m.OilPrice)
	return m
}

// GridCouplingFloor returns the minimum electricity price enforced by gas coupling.
//
// When renewableGridShare < GridCouplingThreshold, electricity cannot fall below:
//
//	gasPrice * conversionFactor * (1 - gridShare/100)
//
// When renewableGridShare >= GridCouplingThreshold, the floor lifts and this
// function returns 0 (no floor).
func GridCouplingFloor(gasPrice, renewableGridShare float64) float64 {
	if renewableGridShare >= GridCouplingThreshold {
		return 0
	}
	return gasPrice * gasToElecConversionFactor * (1 - renewableGridShare/100)
}

// ApplyShock applies a config.EventEffect's percentage price deltas directly
// to the market without advancing the history ring buffers.
// Call TickPrices separately to record the post-shock price in history.
func ApplyShock(m EnergyMarket, effect config.EventEffect) EnergyMarket {
	m.GasPrice = mathutil.Clamp(m.GasPrice*(1+effect.GasPriceDeltaPct/100), 0, 9999)
	m.ElectricityPrice = mathutil.Clamp(m.ElectricityPrice*(1+effect.ElectricityPriceDeltaPct/100), 0, 9999)
	m.OilPrice = mathutil.Clamp(m.OilPrice*(1+effect.OilPriceDeltaPct/100), 0, 9999)
	return m
}

