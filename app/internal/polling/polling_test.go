package polling

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"twenty-fifty/internal/config"
	"twenty-fifty/internal/region"
)

func seededRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

func makeTile(id, regionID string, opinion, capacity float64) region.Tile {
	return region.Tile{
		ID:              id,
		RegionID:        regionID,
		PoliticalOpinion: opinion,
		HeatingCapacity: capacity,
	}
}

func makeRegion(id string, capacity float64) region.Region {
	return region.Region{ID: id, InstallerCapacity: capacity}
}

// ---------------------------------------------------------------------------
// AggregateRegionPoll
// ---------------------------------------------------------------------------

func TestAggregateRegionPoll_SingleTile_SharesSumTo100(t *testing.T) {
	tiles := []region.Tile{makeTile("t1", "r1", 50, 10)}
	rp := AggregateRegionPoll(tiles, "r1", seededRNG(1))
	total := 0.0
	for _, v := range rp.PartyShares {
		total += v
	}
	assert.InDelta(t, 100.0, total, 0.5)
}

func TestAggregateRegionPoll_FiltersByRegionID(t *testing.T) {
	tiles := []region.Tile{
		makeTile("t1", "r1", 50, 10),
		makeTile("t2", "r2", 90, 10), // should be ignored
	}
	rp := AggregateRegionPoll(tiles, "r1", seededRNG(1))
	assert.Equal(t, "r1", rp.RegionID)
}

func TestAggregateRegionPoll_AllPartiesPresent(t *testing.T) {
	tiles := []region.Tile{makeTile("t1", "r1", 50, 10)}
	rp := AggregateRegionPoll(tiles, "r1", seededRNG(1))
	assert.Contains(t, rp.PartyShares, config.PartyFarLeft)
	assert.Contains(t, rp.PartyShares, config.PartyLeft)
	assert.Contains(t, rp.PartyShares, config.PartyRight)
	assert.Contains(t, rp.PartyShares, config.PartyFarRight)
}

func TestAggregateRegionPoll_HighRightOpinion_RightShareDominates(t *testing.T) {
	// Opinion=90 sits near the FarRight position (90); Right should dominate.
	tiles := []region.Tile{makeTile("t1", "r1", 90, 10)}
	// Use a deterministic RNG and take many samples to average out noise.
	totalRight := 0.0
	totalFarLeft := 0.0
	n := 100
	for i := 0; i < n; i++ {
		rp := AggregateRegionPoll(tiles, "r1", seededRNG(int64(i)))
		totalRight += rp.PartyShares[config.PartyFarRight]
		totalFarLeft += rp.PartyShares[config.PartyFarLeft]
	}
	assert.Greater(t, totalRight/float64(n), totalFarLeft/float64(n))
}

func TestAggregateRegionPoll_HighLeftOpinion_LeftShareDominates(t *testing.T) {
	tiles := []region.Tile{makeTile("t1", "r1", 10, 10)}
	totalFarLeft := 0.0
	totalFarRight := 0.0
	n := 100
	for i := 0; i < n; i++ {
		rp := AggregateRegionPoll(tiles, "r1", seededRNG(int64(i)))
		totalFarLeft += rp.PartyShares[config.PartyFarLeft]
		totalFarRight += rp.PartyShares[config.PartyFarRight]
	}
	assert.Greater(t, totalFarLeft/float64(n), totalFarRight/float64(n))
}

func TestAggregateRegionPoll_HeatingCapacityWeightsOpinion(t *testing.T) {
	// Two tiles: one far-left opinion with high capacity, one far-right with low.
	// Result should lean left.
	tiles := []region.Tile{
		makeTile("t1", "r1", 5, 100),  // far-left, high weight
		makeTile("t2", "r1", 95, 1),   // far-right, low weight
	}
	totalFarLeft := 0.0
	totalFarRight := 0.0
	n := 100
	for i := 0; i < n; i++ {
		rp := AggregateRegionPoll(tiles, "r1", seededRNG(int64(i)))
		totalFarLeft += rp.PartyShares[config.PartyFarLeft]
		totalFarRight += rp.PartyShares[config.PartyFarRight]
	}
	assert.Greater(t, totalFarLeft/float64(n), totalFarRight/float64(n))
}

func TestAggregateRegionPoll_ZeroCapacityTiles_FallbackToSimpleMean(t *testing.T) {
	// Both tiles have zero capacity; function should fall back to simple mean without panic.
	tiles := []region.Tile{
		makeTile("t1", "r1", 20, 0),
		makeTile("t2", "r1", 80, 0),
	}
	rp := AggregateRegionPoll(tiles, "r1", seededRNG(1))
	total := 0.0
	for _, v := range rp.PartyShares {
		total += v
	}
	assert.InDelta(t, 100.0, total, 0.5)
}

func TestAggregateRegionPoll_AllSharesAboveMinFloor(t *testing.T) {
	tiles := []region.Tile{makeTile("t1", "r1", 50, 10)}
	for i := 0; i < 50; i++ {
		rp := AggregateRegionPoll(tiles, "r1", seededRNG(int64(i)))
		for _, v := range rp.PartyShares {
			assert.GreaterOrEqual(t, v, 0.0)
		}
	}
}

func TestAggregateRegionPoll_EmptyRegion_ReturnsNeutralShares(t *testing.T) {
	// No tiles for regionID "r99"; should return equal-ish shares, no panic.
	tiles := []region.Tile{makeTile("t1", "r1", 50, 10)}
	rp := AggregateRegionPoll(tiles, "r99", seededRNG(1))
	total := 0.0
	for _, v := range rp.PartyShares {
		total += v
	}
	assert.InDelta(t, 100.0, total, 0.5)
}

// ---------------------------------------------------------------------------
// AggregateNationalPoll
// ---------------------------------------------------------------------------

func TestAggregateNationalPoll_SharesSumTo100(t *testing.T) {
	rps := []RegionPoll{
		{RegionID: "r1", PartyShares: map[config.Party]float64{
			config.PartyLeft: 40, config.PartyRight: 35,
			config.PartyFarLeft: 15, config.PartyFarRight: 10,
		}},
	}
	regions := []region.Region{makeRegion("r1", 50)}
	national := AggregateNationalPoll(rps, regions)
	total := 0.0
	for _, v := range national {
		total += v
	}
	assert.InDelta(t, 100.0, total, 0.5)
}

func TestAggregateNationalPoll_HighWeightRegion_DominatesNational(t *testing.T) {
	rps := []RegionPoll{
		{RegionID: "r1", PartyShares: map[config.Party]float64{
			config.PartyLeft: 70, config.PartyRight: 10,
			config.PartyFarLeft: 10, config.PartyFarRight: 10,
		}},
		{RegionID: "r2", PartyShares: map[config.Party]float64{
			config.PartyLeft: 10, config.PartyRight: 70,
			config.PartyFarLeft: 10, config.PartyFarRight: 10,
		}},
	}
	regions := []region.Region{
		makeRegion("r1", 100), // much higher weight
		makeRegion("r2", 1),
	}
	national := AggregateNationalPoll(rps, regions)
	assert.Greater(t, national[config.PartyLeft], national[config.PartyRight])
}

func TestAggregateNationalPoll_EmptyInput_ReturnsEqualShares(t *testing.T) {
	national := AggregateNationalPoll(nil, nil)
	for _, v := range national {
		assert.InDelta(t, 25.0, v, 0.001)
	}
}

// ---------------------------------------------------------------------------
// TakePollSnapshot
// ---------------------------------------------------------------------------

func TestTakePollSnapshot_SetsWeek(t *testing.T) {
	tiles := []region.Tile{makeTile("t1", "r1", 50, 10)}
	regions := []region.Region{makeRegion("r1", 50)}
	snap := TakePollSnapshot(52, tiles, regions, seededRNG(1))
	assert.Equal(t, 52, snap.Week)
}

func TestTakePollSnapshot_RegionCountMatchesTileRegions(t *testing.T) {
	tiles := []region.Tile{
		makeTile("t1", "r1", 50, 10),
		makeTile("t2", "r2", 60, 10),
	}
	regions := []region.Region{makeRegion("r1", 50), makeRegion("r2", 50)}
	snap := TakePollSnapshot(1, tiles, regions, seededRNG(1))
	assert.Equal(t, 2, len(snap.RegionPolls))
}

func TestTakePollSnapshot_NationalPollsPresent(t *testing.T) {
	tiles := []region.Tile{makeTile("t1", "r1", 50, 10)}
	regions := []region.Region{makeRegion("r1", 50)}
	snap := TakePollSnapshot(1, tiles, regions, seededRNG(1))
	assert.Equal(t, 4, len(snap.NationalPolls))
}

func TestTakePollSnapshot_DoesNotMutateInput(t *testing.T) {
	tiles := []region.Tile{makeTile("t1", "r1", 50, 10)}
	regions := []region.Region{makeRegion("r1", 50)}
	TakePollSnapshot(1, tiles, regions, seededRNG(1))
	assert.InDelta(t, 50.0, tiles[0].PoliticalOpinion, 0.001)
}

// ---------------------------------------------------------------------------
// LeadingParty
// ---------------------------------------------------------------------------

func TestLeadingParty_SingleDominantParty_ReturnsIt(t *testing.T) {
	snap := PollSnapshot{
		NationalPolls: map[config.Party]float64{
			config.PartyLeft:     55,
			config.PartyRight:    25,
			config.PartyFarLeft:  10,
			config.PartyFarRight: 10,
		},
	}
	assert.Equal(t, config.PartyLeft, LeadingParty(snap))
}

func TestLeadingParty_TiedParties_DeterministicResult(t *testing.T) {
	snap := PollSnapshot{
		NationalPolls: map[config.Party]float64{
			config.PartyLeft:     40,
			config.PartyRight:    40,
			config.PartyFarLeft:  10,
			config.PartyFarRight: 10,
		},
	}
	// Call twice; must return same party each time.
	p1 := LeadingParty(snap)
	p2 := LeadingParty(snap)
	assert.Equal(t, p1, p2)
}

func TestLeadingParty_EmptySnapshot_ReturnsDefault(t *testing.T) {
	snap := PollSnapshot{}
	assert.Equal(t, config.PartyLeft, LeadingParty(snap))
}

// ---------------------------------------------------------------------------
// SwingFromLast
// ---------------------------------------------------------------------------

func TestSwingFromLast_NoChange_ReturnsZeroSwing(t *testing.T) {
	shares := map[config.Party]float64{
		config.PartyLeft: 50, config.PartyRight: 30,
		config.PartyFarLeft: 10, config.PartyFarRight: 10,
	}
	rp := RegionPoll{RegionID: "r1", PartyShares: shares}
	current := PollSnapshot{RegionPolls: map[string]RegionPoll{"r1": rp}}
	previous := PollSnapshot{RegionPolls: map[string]RegionPoll{"r1": rp}}
	swings := SwingFromLast(current, previous)
	assert.InDelta(t, 0.0, swings["r1"], 0.001)
}

func TestSwingFromLast_LeadingShareIncreased_PositiveSwing(t *testing.T) {
	prev := RegionPoll{RegionID: "r1", PartyShares: map[config.Party]float64{
		config.PartyLeft: 40, config.PartyRight: 30,
		config.PartyFarLeft: 15, config.PartyFarRight: 15,
	}}
	curr := RegionPoll{RegionID: "r1", PartyShares: map[config.Party]float64{
		config.PartyLeft: 50, config.PartyRight: 25,
		config.PartyFarLeft: 15, config.PartyFarRight: 10,
	}}
	current := PollSnapshot{RegionPolls: map[string]RegionPoll{"r1": curr}}
	previous := PollSnapshot{RegionPolls: map[string]RegionPoll{"r1": prev}}
	swings := SwingFromLast(current, previous)
	assert.Greater(t, swings["r1"], 0.0)
}

func TestSwingFromLast_LeadingShareDecreased_NegativeSwing(t *testing.T) {
	prev := RegionPoll{RegionID: "r1", PartyShares: map[config.Party]float64{
		config.PartyLeft: 50, config.PartyRight: 25,
		config.PartyFarLeft: 15, config.PartyFarRight: 10,
	}}
	curr := RegionPoll{RegionID: "r1", PartyShares: map[config.Party]float64{
		config.PartyLeft: 40, config.PartyRight: 30,
		config.PartyFarLeft: 15, config.PartyFarRight: 15,
	}}
	current := PollSnapshot{RegionPolls: map[string]RegionPoll{"r1": curr}}
	previous := PollSnapshot{RegionPolls: map[string]RegionPoll{"r1": prev}}
	swings := SwingFromLast(current, previous)
	assert.Less(t, swings["r1"], 0.0)
}

func TestSwingFromLast_NewRegionInCurrent_SwingZero(t *testing.T) {
	curr := RegionPoll{RegionID: "r_new", PartyShares: map[config.Party]float64{
		config.PartyLeft: 40, config.PartyRight: 30,
		config.PartyFarLeft: 15, config.PartyFarRight: 15,
	}}
	current := PollSnapshot{RegionPolls: map[string]RegionPoll{"r_new": curr}}
	previous := PollSnapshot{RegionPolls: map[string]RegionPoll{}}
	swings := SwingFromLast(current, previous)
	assert.InDelta(t, 0.0, swings["r_new"], 0.001)
}

func TestSwingFromLast_DoesNotMutateInputs(t *testing.T) {
	shares := map[config.Party]float64{
		config.PartyLeft: 50, config.PartyRight: 30,
		config.PartyFarLeft: 10, config.PartyFarRight: 10,
	}
	rp := RegionPoll{RegionID: "r1", PartyShares: shares, Swing: 0}
	current := PollSnapshot{RegionPolls: map[string]RegionPoll{"r1": rp}}
	previous := PollSnapshot{RegionPolls: map[string]RegionPoll{"r1": rp}}
	SwingFromLast(current, previous)
	assert.InDelta(t, 0.0, current.RegionPolls["r1"].Swing, 0.001)
}
