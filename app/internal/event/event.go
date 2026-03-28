package event

import (
	"math/rand"
	"strings"

	"twenty-fifty/internal/carbon"
	"twenty-fifty/internal/config"
	"twenty-fifty/internal/industry"
	"twenty-fifty/internal/mathutil"
	"twenty-fifty/internal/stakeholder"
)

// ---------------------------------------------------------------------------
// EventLog
// ---------------------------------------------------------------------------

// EventEntry records one fired global event with its week and resolved effects.
type EventEntry struct {
	DefID   string
	Name    string
	Week    int
	Effects ResolvedEffect
}

// EventLog is a circular buffer of player-visible weekly events.
const eventLogCapacity = 52

// EventLog holds the rolling 52-week history of fired events.
type EventLog struct {
	entries [eventLogCapacity]EventEntry
	head    int // next write position
	count   int // total entries written (up to capacity)
}

// NewEventLog returns an empty EventLog.
func NewEventLog() EventLog {
	return EventLog{}
}

// AppendEventLog adds an entry to the log, overwriting the oldest if full.
func AppendEventLog(log EventLog, entry EventEntry) EventLog {
	log.entries[log.head] = entry
	log.head = (log.head + 1) % eventLogCapacity
	if log.count < eventLogCapacity {
		log.count++
	}
	return log
}

// Entries returns the log's events in chronological order (oldest first).
func (l EventLog) Entries() []EventEntry {
	if l.count == 0 {
		return nil
	}
	out := make([]EventEntry, l.count)
	start := (l.head - l.count + eventLogCapacity) % eventLogCapacity
	for i := 0; i < l.count; i++ {
		out[i] = l.entries[(start+i)%eventLogCapacity]
	}
	return out
}

// ---------------------------------------------------------------------------
// PendingShockResponse
// ---------------------------------------------------------------------------

// PendingShockResponse is queued for the player when an event has OffersShockResponse=true.
// It records which event triggered the response opportunity and on which week it was offered.
// The full resolution logic (backfire probability, outcome) lives in climate.ShockResponseOutcome.
type PendingShockResponse struct {
	EventDefID string
	Week       int
}

// ---------------------------------------------------------------------------
// Pressure groups
// ---------------------------------------------------------------------------

// PressureGroup is a persistent actor that generates pressure events based on
// the state of carbon trajectory and low-carbon reputation.
type PressureGroup struct {
	ID           string
	Name         string
	ConstantPop  float64 // weekly GovernmentPopularity delta (always applied)
	CarbonPop    float64 // additional pop delta when CarbonTrajectory > carbonTriggerThreshold
	ConstantLCR  float64 // weekly LCR delta (always applied)
	LowLCRBoost  float64 // additional LCR delta when LCR < lcrLowThreshold
}

// PressureResult is the aggregate output of one tick's pressure group application.
type PressureResult struct {
	GroupID             string
	GovtPopularityDelta float64
	LCRDelta            float64
}

// calibration thresholds for pressure group triggers.
const (
	carbonTriggerThreshold = 500.0 // MtCO2e/year trajectory that activates carbon pressure
	lcrLowThreshold        = 30.0  // LCR below this level activates low-LCR boost
)

// DefaultPressureGroups returns the five fixed pressure group actors.
func DefaultPressureGroups() []PressureGroup {
	return []PressureGroup{
		{
			ID:          "greens_alliance",
			Name:        "Greens Alliance",
			ConstantPop: 0.0,
			CarbonPop:   -1.5, // penalises govt when trajectory is off-track
			ConstantLCR: 1.0,
			LowLCRBoost: 2.5,
		},
		{
			ID:          "fossil_lobby",
			Name:        "Fossil Fuel Lobby",
			ConstantPop: 0.3,
			CarbonPop:   0.0,
			ConstantLCR: -0.5,
			LowLCRBoost: -1.5, // amplifies LCR damage when LCR is already low
		},
		{
			ID:          "trade_unions",
			Name:        "Trade Unions",
			ConstantPop: -0.2,
			CarbonPop:   -0.8,
			ConstantLCR: 0.3,
			LowLCRBoost: 0.0,
		},
		{
			ID:          "business_council",
			Name:        "Business Council",
			ConstantPop: 0.5,
			CarbonPop:   -1.0,
			ConstantLCR: -0.2,
			LowLCRBoost: 0.0,
		},
		{
			ID:          "local_govt_network",
			Name:        "Local Government Network",
			ConstantPop: 0.1,
			CarbonPop:   -0.5,
			ConstantLCR: 0.5,
			LowLCRBoost: 1.0,
		},
	}
}

// ApplyPressureGroups computes the weekly delta from all pressure groups.
// carbonTrajectory is the projected annual MtCO2e total; lcr is 0-100.
func ApplyPressureGroups(groups []PressureGroup, carbonTrajectory, lcr float64) []PressureResult {
	results := make([]PressureResult, 0, len(groups))
	for _, g := range groups {
		popDelta := g.ConstantPop
		lcrDelta := g.ConstantLCR
		if carbonTrajectory > carbonTriggerThreshold {
			popDelta += g.CarbonPop
		}
		if lcr < lcrLowThreshold {
			lcrDelta += g.LowLCRBoost
		}
		results = append(results, PressureResult{
			GroupID:             g.ID,
			GovtPopularityDelta: popDelta,
			LCRDelta:            lcrDelta,
		})
	}
	return results
}

// ---------------------------------------------------------------------------
// Event probability and drawing
// ---------------------------------------------------------------------------

// ComputeEventProbability returns the adjusted weekly draw probability for one
// event definition given the current climate level and fossil dependency score.
func ComputeEventProbability(
	def config.EventDef,
	climateLevel carbon.ClimateLevel,
	fossilDependency float64,
) float64 {
	p := def.BaseProbability
	if climateLevel >= carbon.ClimateLevelElevated {
		p *= def.ClimateMultiplier
	}
	if fossilDependency > 60.0 {
		p *= def.FossilMultiplier
	}
	return mathutil.Clamp(p, 0, 1)
}

// DrawEvent selects at most one event from the deck this week via independent
// Bernoulli trials. Returns the drawn event and true, or zero value and false
// if no event fires. The first event whose trial succeeds is returned.
func DrawEvent(
	defs []config.EventDef,
	climateLevel carbon.ClimateLevel,
	fossilDependency float64,
	rng *rand.Rand,
) (config.EventDef, bool) {
	for _, def := range defs {
		p := ComputeEventProbability(def, climateLevel, fossilDependency)
		if p > 0 && rng.Float64() < p {
			return def, true
		}
	}
	return config.EventDef{}, false
}

// ---------------------------------------------------------------------------
// Scandal roll
// ---------------------------------------------------------------------------

// Scandal calibration constants.
const (
	scandalBaseProb       = 0.005
	scandalPressureFactor = 0.002
	scandalPopulismFactor = 0.0002
	scandalProbCap        = 0.10
)

// RollScandal performs the weekly scandal check for one stakeholder.
// weeksUnderPressure is the number of consecutive weeks this minister has
// been in MinisterStateUnderPressure. Returns true if a scandal fires.
func RollScandal(s stakeholder.Stakeholder, weeksUnderPressure int, rng *rand.Rand) bool {
	p := scandalBaseProb +
		float64(weeksUnderPressure)*scandalPressureFactor +
		s.PopulismScore*scandalPopulismFactor
	p = mathutil.Clamp(p, 0, scandalProbCap)
	return rng.Float64() < p
}

// ---------------------------------------------------------------------------
// Targeting resolvers
// ---------------------------------------------------------------------------

// MatchRegions returns the IDs of regions that match the given filter.
// Filter semantics (case-insensitive tag match or exact region ID):
//   - Empty string -> returns all region IDs.
//   - "COASTAL","RURAL","URBAN","INDUSTRIAL","AGRICULTURAL" -> tag match.
//   - Any other value -> exact region ID match.
func MatchRegions(filter string, regions []config.RegionDef) []string {
	if filter == "" {
		ids := make([]string, 0, len(regions))
		for _, r := range regions {
			ids = append(ids, r.ID)
		}
		return ids
	}
	tag := strings.ToLower(filter)
	knownTags := map[string]bool{
		"coastal": true, "rural": true, "urban": true,
		"industrial": true, "agricultural": true,
	}
	var ids []string
	if knownTags[tag] {
		for _, r := range regions {
			for _, t := range r.Tags {
				if strings.ToLower(t) == tag {
					ids = append(ids, r.ID)
					break
				}
			}
		}
	} else {
		// Treat filter as an exact region ID.
		for _, r := range regions {
			if r.ID == filter {
				ids = append(ids, r.ID)
				break
			}
		}
	}
	return ids
}

// MatchStakeholders returns the IDs of unlocked stakeholders that match the filter.
// Filter semantics:
//   - Empty string -> returns empty slice (no stakeholder effect intended).
//   - "ALL" -> all unlocked stakeholders.
//   - "CABINET" -> all unlocked stakeholders (all hold a cabinet role).
//   - "ROLE:LEADER","ROLE:CHANCELLOR","ROLE:FOREIGN_SECRETARY","ROLE:ENERGY" -> role match.
func MatchStakeholders(filter string, stakeholders []stakeholder.Stakeholder) []string {
	if filter == "" {
		return nil
	}
	var ids []string
	upper := strings.ToUpper(filter)
	for _, s := range stakeholders {
		if !s.IsUnlocked {
			continue
		}
		switch upper {
		case "ALL", "CABINET":
			ids = append(ids, s.ID)
		case "ROLE:LEADER":
			if s.Role == config.RoleLeader {
				ids = append(ids, s.ID)
			}
		case "ROLE:CHANCELLOR":
			if s.Role == config.RoleChancellor {
				ids = append(ids, s.ID)
			}
		case "ROLE:FOREIGN_SECRETARY":
			if s.Role == config.RoleForeignSecretary {
				ids = append(ids, s.ID)
			}
		case "ROLE:ENERGY":
			if s.Role == config.RoleEnergy {
				ids = append(ids, s.ID)
			}
		}
	}
	return ids
}

// MatchCompanies returns the DefIDs of active companies that match the filter.
// Filter semantics:
//   - Empty string -> returns empty slice (no company effect intended).
//   - "ALL" -> all active companies.
//   - "TECH:<category>" -> matches by TechCategory string (e.g. "TECH:EVS").
func MatchCompanies(filter string, companies []industry.CompanyState, defs map[string]config.CompanyDef) []string {
	if filter == "" {
		return nil
	}
	upper := strings.ToUpper(filter)
	var ids []string
	for _, c := range companies {
		if c.Status != industry.CompanyStatusActive &&
			c.Status != industry.CompanyStatusStartup &&
			c.Status != industry.CompanyStatusStruggling {
			continue
		}
		def, ok := defs[c.DefID]
		if !ok {
			continue
		}
		if upper == "ALL" {
			ids = append(ids, c.DefID)
			continue
		}
		if strings.HasPrefix(upper, "TECH:") {
			wantCat := strings.TrimPrefix(upper, "TECH:")
			if strings.ToUpper(string(def.TechCategory)) == wantCat {
				ids = append(ids, c.DefID)
			}
		}
	}
	return ids
}

// ---------------------------------------------------------------------------
// ResolvedEffect and delta types
// ---------------------------------------------------------------------------

// RegionDelta holds the targeted changes applied to one region.
type RegionDelta struct {
	InstallerCapacityDelta float64
	SkillsNetworkDelta     float64
}

// TileDelta holds the targeted changes applied to each tile in a matched region.
type TileDelta struct {
	FuelPovertyDelta float64
	InsulationDamage float64
}

// StakeholderDelta holds the targeted changes applied to one stakeholder.
type StakeholderDelta struct {
	RelDelta      float64
	PressureDelta int
}

// CompanyDelta holds the targeted changes applied to one company.
type CompanyDelta struct {
	WorkRateDelta float64
	QualityDelta  float64
}

// ResolvedEffect is the fully expanded output of ResolveEffect. The simulation
// iterates these maps and applies each delta to the matching entity.
type ResolvedEffect struct {
	// Global effects
	GasPriceDeltaPct         float64
	ElectricityPriceDeltaPct float64
	OilPriceDeltaPct         float64
	EconomyDelta             float64
	LCRDelta                 float64
	GovtPopularityDelta      float64
	CarbonEmissionsDeltaMt   float64

	// Targeted effects
	RegionDeltas      map[string]RegionDelta      // region ID -> delta
	TileDeltas        map[string]TileDelta        // region ID -> per-tile delta (applied to all tiles in region)
	StakeholderDeltas map[string]StakeholderDelta // stakeholder ID -> delta
	CompanyDeltas     map[string]CompanyDelta     // company DefID -> delta
}

// ResolveEffect expands the filter strings in an EventEffect into a concrete
// ResolvedEffect by matching them against the current entity snapshots.
func ResolveEffect(
	effect config.EventEffect,
	regionDefs []config.RegionDef,
	stakeholders []stakeholder.Stakeholder,
	companies []industry.CompanyState,
	companyDefs map[string]config.CompanyDef,
) ResolvedEffect {
	out := ResolvedEffect{
		GasPriceDeltaPct:         effect.GasPriceDeltaPct,
		ElectricityPriceDeltaPct: effect.ElectricityPriceDeltaPct,
		OilPriceDeltaPct:         effect.OilPriceDeltaPct,
		EconomyDelta:             effect.EconomyDelta,
		LCRDelta:                 effect.LCRDelta,
		GovtPopularityDelta:      effect.GovtPopularityDelta,
		CarbonEmissionsDeltaMt:   effect.CarbonEmissionsDeltaMt,
		RegionDeltas:             make(map[string]RegionDelta),
		TileDeltas:               make(map[string]TileDelta),
		StakeholderDeltas:        make(map[string]StakeholderDelta),
		CompanyDeltas:            make(map[string]CompanyDelta),
	}

	// Region-targeted effects
	if effect.RegionFilter != "" ||
		effect.InstallerCapacityDelta != 0 ||
		effect.SkillsNetworkDelta != 0 ||
		effect.TileFuelPovertyDelta != 0 ||
		effect.TileInsulationDamage != 0 {

		matchedRegions := MatchRegions(effect.RegionFilter, regionDefs)
		for _, rid := range matchedRegions {
			if effect.InstallerCapacityDelta != 0 || effect.SkillsNetworkDelta != 0 {
				out.RegionDeltas[rid] = RegionDelta{
					InstallerCapacityDelta: effect.InstallerCapacityDelta,
					SkillsNetworkDelta:     effect.SkillsNetworkDelta,
				}
			}
			if effect.TileFuelPovertyDelta != 0 || effect.TileInsulationDamage != 0 {
				out.TileDeltas[rid] = TileDelta{
					FuelPovertyDelta: effect.TileFuelPovertyDelta,
					InsulationDamage: effect.TileInsulationDamage,
				}
			}
		}
	}

	// Stakeholder-targeted effects
	if effect.StakeholderFilter != "" {
		matchedSH := MatchStakeholders(effect.StakeholderFilter, stakeholders)
		for _, sid := range matchedSH {
			out.StakeholderDeltas[sid] = StakeholderDelta{
				RelDelta:      effect.StakeholderRelDelta,
				PressureDelta: effect.StakeholderPressureDelta,
			}
		}
	}

	// Company-targeted effects
	if effect.CompanyFilter != "" {
		matchedCo := MatchCompanies(effect.CompanyFilter, companies, companyDefs)
		for _, cid := range matchedCo {
			out.CompanyDeltas[cid] = CompanyDelta{
				WorkRateDelta: effect.CompanyWorkRateDelta,
				QualityDelta:  effect.CompanyQualityDelta,
			}
		}
	}

	return out
}
