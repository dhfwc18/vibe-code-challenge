package polling

import (
	"math"
	"math/rand"
	"sort"

	"twenty-fifty/internal/config"
	"twenty-fifty/internal/mathutil"
	"twenty-fifty/internal/region"
)

// RegionPoll is the political snapshot for one region at one polling period.
type RegionPoll struct {
	RegionID    string
	PartyShares map[config.Party]float64 // always contains all four parties; sums to ~100
	Swing       float64                  // leading-party share delta from the previous poll
}

// PollSnapshot is the full national and regional polling picture at one game week.
type PollSnapshot struct {
	Week                  int
	RegionPolls           map[string]RegionPoll    // keyed by RegionID
	NationalPolls         map[config.Party]float64 // always contains all four parties; sums to ~100
	GovernmentApprovalRating float64               // 0-100 noisy sample of GovernmentPopularity (sigma=3)
}

// allParties lists the four parties in a fixed order used for deterministic iteration.
var allParties = []config.Party{
	config.PartyFarLeft,
	config.PartyLeft,
	config.PartyRight,
	config.PartyFarRight,
}

// partyPosition maps each party to its position on the 0-100 opinion axis.
// FarLeft sits near 10, FarRight near 90.
var partyPosition = map[config.Party]float64{
	config.PartyFarLeft:  10.0,
	config.PartyLeft:     35.0,
	config.PartyRight:    65.0,
	config.PartyFarRight: 90.0,
}

// Calibration constants.
const (
	pollNoiseSigma    = 3.0  // Gaussian noise standard deviation per party share
	kernelBandwidth   = 25.0 // bandwidth of the Gaussian proximity kernel
	minPartyShare     = 0.5  // minimum share floor to prevent parties vanishing
)

// AggregateRegionPoll computes the party-share distribution for one region from its tiles.
//
// Algorithm:
//  1. Filter tiles to regionID; weight each by HeatingCapacity.
//  2. Compute weighted mean PoliticalOpinion (falls back to simple mean at zero capacity).
//  3. Map opinion to party shares via a Gaussian proximity kernel.
//  4. Add per-party Gaussian noise and renormalise.
//  5. Apply minPartyShare floor and renormalise again.
//
// Swing is always 0.0; callers call SwingFromLast to fill it.
func AggregateRegionPoll(tiles []region.Tile, regionID string, rng *rand.Rand) RegionPoll {
	// Filter tiles belonging to this region.
	var regional []region.Tile
	for _, t := range tiles {
		if t.RegionID == regionID {
			regional = append(regional, t)
		}
	}

	// Weighted mean opinion.
	opinion := weightedMeanOpinion(regional)

	// Gaussian kernel proximity scores.
	raw := make(map[config.Party]float64, 4)
	total := 0.0
	for _, p := range allParties {
		pos := partyPosition[p]
		diff := (opinion - pos) / kernelBandwidth
		score := math.Exp(-0.5 * diff * diff)
		raw[p] = score
		total += score
	}
	// Normalise to sum = 100.
	shares := make(map[config.Party]float64, 4)
	if total > 0 {
		for _, p := range allParties {
			shares[p] = raw[p] / total * 100.0
		}
	} else {
		for _, p := range allParties {
			shares[p] = 25.0
		}
	}

	// Add Gaussian noise and renormalise.
	noiseTotal := 0.0
	for _, p := range allParties {
		noisy := shares[p] + rng.NormFloat64()*pollNoiseSigma
		if noisy < 0 {
			noisy = 0
		}
		shares[p] = noisy
		noiseTotal += noisy
	}
	if noiseTotal > 0 {
		for _, p := range allParties {
			shares[p] = shares[p] / noiseTotal * 100.0
		}
	}

	// Apply minimum share floor and renormalise.
	shares = applyFloorAndNormalise(shares)

	return RegionPoll{
		RegionID:    regionID,
		PartyShares: shares,
		Swing:       0.0,
	}
}

// AggregateNationalPoll computes the national party-share distribution as a
// weighted average of region polls. Regions are weighted by InstallerCapacity.
// If the total weight is zero, equal weighting is used.
func AggregateNationalPoll(regionPolls []RegionPoll, regions []region.Region) map[config.Party]float64 {
	// Build capacity weight map.
	weights := make(map[string]float64, len(regions))
	totalWeight := 0.0
	for _, r := range regions {
		weights[r.ID] = r.InstallerCapacity
		totalWeight += r.InstallerCapacity
	}

	national := make(map[config.Party]float64, 4)
	for _, p := range allParties {
		national[p] = 0.0
	}

	if totalWeight == 0 {
		// Equal weight fallback.
		if len(regionPolls) == 0 {
			for _, p := range allParties {
				national[p] = 25.0
			}
			return national
		}
		for _, rp := range regionPolls {
			for _, p := range allParties {
				national[p] += rp.PartyShares[p]
			}
		}
		n := float64(len(regionPolls))
		for _, p := range allParties {
			national[p] /= n
		}
		return national
	}

	// Weighted average.
	for _, rp := range regionPolls {
		w := weights[rp.RegionID]
		if w == 0 {
			// Region not in the regions slice; give it a weight equal to mean.
			w = totalWeight / float64(len(regions))
		}
		for _, p := range allParties {
			national[p] += w * rp.PartyShares[p]
		}
	}

	// Normalise to sum = 100.
	sum := 0.0
	for _, p := range allParties {
		sum += national[p]
	}
	if sum > 0 {
		for _, p := range allParties {
			national[p] = national[p] / sum * 100.0
		}
	}
	return national
}

// TakePollSnapshot produces a full PollSnapshot for the given week.
// It collects all unique regionIDs from tiles, aggregates each region's poll,
// then aggregates nationally. Swing fields are left at 0.0.
func TakePollSnapshot(week int, tiles []region.Tile, regions []region.Region, rng *rand.Rand) PollSnapshot {
	// Collect unique region IDs from tiles in a deterministic order.
	seen := make(map[string]bool)
	var regionIDs []string
	for _, t := range tiles {
		if !seen[t.RegionID] {
			seen[t.RegionID] = true
			regionIDs = append(regionIDs, t.RegionID)
		}
	}
	sort.Strings(regionIDs)

	regionPolls := make(map[string]RegionPoll, len(regionIDs))
	var pollSlice []RegionPoll
	for _, rid := range regionIDs {
		rp := AggregateRegionPoll(tiles, rid, rng)
		regionPolls[rid] = rp
		pollSlice = append(pollSlice, rp)
	}

	national := AggregateNationalPoll(pollSlice, regions)

	return PollSnapshot{
		Week:          week,
		RegionPolls:   regionPolls,
		NationalPolls: national,
	}
}

// LeadingParty returns the party with the highest share in snap.NationalPolls.
// On a tie, the result is deterministic (alphabetical party ID order).
// Returns config.PartyLeft as a safe default if the map is empty.
func LeadingParty(snap PollSnapshot) config.Party {
	if len(snap.NationalPolls) == 0 {
		return config.PartyLeft
	}
	// Sort party IDs for deterministic tie-breaking.
	keys := make([]string, 0, len(snap.NationalPolls))
	for p := range snap.NationalPolls {
		keys = append(keys, string(p))
	}
	sort.Strings(keys)

	var best config.Party
	bestShare := -1.0
	for _, k := range keys {
		p := config.Party(k)
		if snap.NationalPolls[p] > bestShare {
			bestShare = snap.NationalPolls[p]
			best = p
		}
	}
	return best
}

// SwingFromLast returns a map of regionID -> swing, where swing is the signed change
// in the leading party's share between current and previous snapshots.
// Regions present in current but absent in previous receive swing=0.
// Neither input is mutated.
func SwingFromLast(current, previous PollSnapshot) map[string]float64 {
	result := make(map[string]float64, len(current.RegionPolls))
	for rid, cur := range current.RegionPolls {
		prev, ok := previous.RegionPolls[rid]
		if !ok {
			result[rid] = 0.0
			continue
		}
		result[rid] = leadingShare(cur.PartyShares) - leadingShare(prev.PartyShares)
	}
	return result
}

// leadingShare returns the highest party share in a shares map.
func leadingShare(shares map[config.Party]float64) float64 {
	best := 0.0
	for _, v := range shares {
		if v > best {
			best = v
		}
	}
	return best
}

// weightedMeanOpinion computes the HeatingCapacity-weighted mean PoliticalOpinion
// for a set of tiles. Falls back to a simple mean when total capacity is zero.
func weightedMeanOpinion(tiles []region.Tile) float64 {
	if len(tiles) == 0 {
		return 50.0 // neutral default for empty regions
	}
	sumW := 0.0
	sumWO := 0.0
	for _, t := range tiles {
		sumW += t.HeatingCapacity
		sumWO += t.HeatingCapacity * t.PoliticalOpinion
	}
	if sumW == 0 {
		// Simple mean fallback.
		total := 0.0
		for _, t := range tiles {
			total += t.PoliticalOpinion
		}
		return total / float64(len(tiles))
	}
	return sumWO / sumW
}

// applyFloorAndNormalise applies minPartyShare to all parties and renormalises
// so shares still sum to 100.
func applyFloorAndNormalise(shares map[config.Party]float64) map[config.Party]float64 {
	out := make(map[config.Party]float64, len(shares))
	total := 0.0
	for _, p := range allParties {
		v := mathutil.Clamp(shares[p], minPartyShare, 100)
		out[p] = v
		total += v
	}
	if total > 0 {
		for _, p := range allParties {
			out[p] = out[p] / total * 100.0
		}
	}
	return out
}
