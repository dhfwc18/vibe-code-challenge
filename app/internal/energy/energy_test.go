package energy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"twenty-fifty/internal/config"
)

// ---------------------------------------------------------------------------
// NewMarket
// ---------------------------------------------------------------------------

func TestNewMarket_SetsGreenBookAnchors(t *testing.T) {
	m := NewMarket()
	assert.Equal(t, Anchor2010GasPriceMWh, m.GasPrice)
	assert.Equal(t, Anchor2010ElectricityPriceMWh, m.ElectricityPrice)
	assert.Equal(t, Anchor2010OilPriceBarrel, m.OilPrice)
	assert.Equal(t, Anchor2010RenewableGridShare, m.RenewableGridShare)
}

func TestNewMarket_HistoryPrefilledWithAnchors(t *testing.T) {
	m := NewMarket()
	for i := 0; i < 52; i++ {
		assert.Equal(t, Anchor2010GasPriceMWh, m.GasHistory.Values[i],
			"gas history slot %d should be anchor value", i)
	}
}

// ---------------------------------------------------------------------------
// PriceRing / PushPrice / PriceAt
// ---------------------------------------------------------------------------

func TestPriceAt_ZeroIndex_ReturnsMostRecent(t *testing.T) {
	ring := PriceRing{}
	ring = PushPrice(ring, 100.0)
	assert.Equal(t, 100.0, PriceAt(ring, 0))
}

func TestPriceAt_OneIndex_ReturnsPreviousWeek(t *testing.T) {
	ring := PriceRing{}
	ring = PushPrice(ring, 10.0)
	ring = PushPrice(ring, 20.0)
	assert.Equal(t, 20.0, PriceAt(ring, 0))
	assert.Equal(t, 10.0, PriceAt(ring, 1))
}

func TestPushPrice_WrapsAt52(t *testing.T) {
	ring := PriceRing{}
	for i := 0; i < 53; i++ {
		ring = PushPrice(ring, float64(i))
	}
	// After 53 pushes, Head has wrapped; most recent is 52, one ago is 51
	assert.Equal(t, 52.0, PriceAt(ring, 0))
	assert.Equal(t, 51.0, PriceAt(ring, 1))
}

func TestPriceAt_NegativeIndex_ClampsToZero(t *testing.T) {
	ring := PriceRing{}
	ring = PushPrice(ring, 42.0)
	assert.Equal(t, PriceAt(ring, 0), PriceAt(ring, -5))
}

func TestPriceAt_OverMax_ClampsTo51(t *testing.T) {
	ring := PriceRing{}
	ring = PushPrice(ring, 42.0)
	assert.Equal(t, PriceAt(ring, 51), PriceAt(ring, 999))
}

func TestPushPrice_DoesNotMutateOriginal(t *testing.T) {
	ring := PriceRing{}
	PushPrice(ring, 99.0)
	assert.Equal(t, 0.0, PriceAt(ring, 0), "PushPrice must not mutate the original ring")
}

// ---------------------------------------------------------------------------
// GridCouplingFloor
// ---------------------------------------------------------------------------

func TestGridCouplingFloor_BelowThreshold_ReturnsPositiveFloor(t *testing.T) {
	floor := GridCouplingFloor(25.0, 20.0)
	assert.Greater(t, floor, 0.0)
}

func TestGridCouplingFloor_AtThreshold_ReturnsZero(t *testing.T) {
	assert.Equal(t, 0.0, GridCouplingFloor(25.0, GridCouplingThreshold))
}

func TestGridCouplingFloor_AboveThreshold_ReturnsZero(t *testing.T) {
	assert.Equal(t, 0.0, GridCouplingFloor(25.0, 60.0))
}

func TestGridCouplingFloor_HigherGas_RaisesFloor(t *testing.T) {
	low := GridCouplingFloor(20.0, 20.0)
	high := GridCouplingFloor(40.0, 20.0)
	assert.Greater(t, high, low)
}

func TestGridCouplingFloor_HigherGridShare_LowersFloor(t *testing.T) {
	low := GridCouplingFloor(25.0, 10.0)
	high := GridCouplingFloor(25.0, 35.0)
	assert.Less(t, high, low)
}

// ---------------------------------------------------------------------------
// TickPrices
// ---------------------------------------------------------------------------

func TestTickPrices_PositiveDelta_RaisesPrices(t *testing.T) {
	m := NewMarket()
	updated := TickPrices(m, 10.0, 10.0, 10.0, 0)
	assert.Greater(t, updated.GasPrice, m.GasPrice)
	assert.Greater(t, updated.ElectricityPrice, m.ElectricityPrice)
	assert.Greater(t, updated.OilPrice, m.OilPrice)
}

func TestTickPrices_NegativeDelta_LowersPrices(t *testing.T) {
	m := NewMarket()
	updated := TickPrices(m, -10.0, -10.0, -10.0, 0)
	assert.Less(t, updated.GasPrice, m.GasPrice)
}

func TestTickPrices_ZeroDelta_PricesUnchangedExceptCouplingCheck(t *testing.T) {
	m := NewMarket()
	updated := TickPrices(m, 0, 0, 0, 0)
	// Grid coupling floor may push electricity up slightly; gas and oil unchanged
	assert.InDelta(t, m.GasPrice, updated.GasPrice, 0.001)
	assert.InDelta(t, m.OilPrice, updated.OilPrice, 0.001)
}

func TestTickPrices_EnforcesGridCouplingFloor(t *testing.T) {
	m := NewMarket()
	m.RenewableGridShare = 10.0   // low renewables; coupling active
	m.GasPrice = 50.0
	m.ElectricityPrice = 1.0      // artificially low
	updated := TickPrices(m, 0, 0, 0, 0)
	floor := GridCouplingFloor(50.0, 10.0)
	assert.GreaterOrEqual(t, updated.ElectricityPrice, floor)
}

func TestTickPrices_AboveCouplingThreshold_FloorDoesNotApply(t *testing.T) {
	m := NewMarket()
	m.RenewableGridShare = 60.0  // above threshold
	m.GasPrice = 100.0
	m.ElectricityPrice = 5.0     // very low but above threshold uncoupled
	updated := TickPrices(m, 0, 0, 0, 0)
	assert.InDelta(t, 5.0, updated.ElectricityPrice, 0.01,
		"above coupling threshold, low electricity price should not be raised")
}

func TestTickPrices_PushesToHistory(t *testing.T) {
	m := NewMarket()
	updated := TickPrices(m, 10.0, 0, 0, 0)
	assert.Equal(t, updated.GasPrice, PriceAt(updated.GasHistory, 0))
}

func TestTickPrices_DoesNotMutateOriginal(t *testing.T) {
	m := NewMarket()
	original := m.GasPrice
	TickPrices(m, 10.0, 0, 0, 0)
	assert.Equal(t, original, m.GasPrice)
}

func TestTickPrices_ClampsToZero(t *testing.T) {
	m := NewMarket()
	updated := TickPrices(m, -100.0, -100.0, -100.0, 0)
	assert.Equal(t, 0.0, updated.GasPrice)
}

// ---------------------------------------------------------------------------
// ApplyShock
// ---------------------------------------------------------------------------

func TestApplyShock_RaisesPricesByExpectedPercentage(t *testing.T) {
	m := NewMarket()
	effect := config.EventEffect{GasPriceDeltaPct: 20.0}
	updated := ApplyShock(m, effect)
	assert.InDelta(t, Anchor2010GasPriceMWh*1.2, updated.GasPrice, 0.001)
}

func TestApplyShock_DoesNotAdvanceHistory(t *testing.T) {
	m := NewMarket()
	effect := config.EventEffect{GasPriceDeltaPct: 20.0}
	updated := ApplyShock(m, effect)
	// History head should not have advanced
	assert.Equal(t, m.GasHistory.Head, updated.GasHistory.Head)
}

func TestApplyShock_ZeroEffect_NoPriceChange(t *testing.T) {
	m := NewMarket()
	updated := ApplyShock(m, config.EventEffect{})
	assert.Equal(t, m.GasPrice, updated.GasPrice)
	assert.Equal(t, m.ElectricityPrice, updated.ElectricityPrice)
	assert.Equal(t, m.OilPrice, updated.OilPrice)
}
