package evidence

import (
	"math"
	"math/rand"

	"twenty-fifty/internal/config"
	"twenty-fifty/internal/mathutil"
)

// ---------------------------------------------------------------------------
// Runtime state types
// ---------------------------------------------------------------------------

// Commission represents one active advisory commission in progress.
type Commission struct {
	ID              string
	OrgID           string
	InsightType     config.InsightType
	Scope           string  // free-text scope string visible in the UI
	CommissionedWeek int
	DeliveryWeek    int    // week number when delivery occurs
	BudgetCost      float64 // GBP thousands, fixed at commission time
	Failed          bool   // set to true if the failure roll fires on delivery
	Delivered       bool
}

// InsightReport is produced when a Commission is delivered.
type InsightReport struct {
	CommissionID string
	OrgID        string
	InsightType  config.InsightType
	Scope        string
	RawValue     float64 // 0-100: the true underlying value before quality/bias distortion
	ReportedValue float64 // 0-100: the value the organisation presents to the player
	TopicKey     string  // stable key for cross-referencing multiple reports (e.g. "energy:lcr")
	DeliveryWeek int
	QualityScore float64 // 0-100: the quality roll that determined distortion magnitude
	SpecialismBonus bool // true if the org had this InsightType in its Specialisms
}

// OrgState tracks the runtime relationship and availability of one advisory org.
type OrgState struct {
	OrgID             string
	RelationshipScore float64 // 0-100; starts at 50; improves with repeat use
	CommissionCount   int
	CoolingOffUntil   int  // week number after a failed commission; org unavailable until then
}

// RelationshipEvent is passed to UpdateOrgRelationship to describe what happened.
type RelationshipEvent string

const (
	RelationshipEventDelivered RelationshipEvent = "DELIVERED"    // commission completed successfully
	RelationshipEventActedOn   RelationshipEvent = "ACTED_ON"     // player acted on a report finding
	RelationshipEventIgnored   RelationshipEvent = "IGNORED"      // report finding ignored for 4+ weeks
	RelationshipEventFailed    RelationshipEvent = "FAILED"       // commission failed entirely
)

// ---------------------------------------------------------------------------
// Calibration constants
// ---------------------------------------------------------------------------

const (
	startingOrgRelationship = 50.0
	orgRelDecayTarget       = 50.0
	orgRelDecayRate         = 0.01 // fraction of (score-50) that reverts per week

	relationshipDeliveredDelta = 2.0
	relationshipActedOnDelta   = 4.0
	relationshipIgnoredDelta   = -1.5
	relationshipFailedDelta    = -6.0

	coolingOffWeeks = 8 // org unavailable for 8 weeks after a failure

	ideologicalBiasStrength    = 12.0 // max shift on 0-100 scale from ideological bias
	clientBiasStrength         = 8.0  // max shift from client confirmation bias
	noneNoiseSigma             = 2.0  // Gaussian noise sigma for BiasNone organisations

	outsideSpecialismQualityPenalty = 20.0 // quality deducted when InsightType not in Specialisms
)

// ---------------------------------------------------------------------------
// Triangular distribution
// ---------------------------------------------------------------------------

// DrawDeliveryWeek samples a delivery week offset from the triangular distribution
// defined by dist, then adds it to commissionedWeek. Uses the inverse-CDF method.
func DrawDeliveryWeek(dist config.TriangularDist, commissionedWeek int, rng *rand.Rand) int {
	a := float64(dist.Min)
	c := float64(dist.Mode)
	b := float64(dist.Max)

	u := rng.Float64()
	fc := (c - a) / (b - a)

	var x float64
	if u < fc {
		x = a + math.Sqrt(u*(b-a)*(c-a))
	} else {
		x = b - math.Sqrt((1.0-u)*(b-a)*(b-c))
	}

	offset := int(math.Round(x))
	if offset < dist.Min {
		offset = dist.Min
	}
	if offset > dist.Max {
		offset = dist.Max
	}
	return commissionedWeek + offset
}

// ---------------------------------------------------------------------------
// Commission management
// ---------------------------------------------------------------------------

// SeedOrgStates creates an OrgState for each OrgDefinition, starting at default values.
func SeedOrgStates(defs []config.OrgDefinition) []OrgState {
	out := make([]OrgState, 0, len(defs))
	for _, d := range defs {
		out = append(out, OrgState{
			OrgID:             d.ID,
			RelationshipScore: startingOrgRelationship,
		})
	}
	return out
}

// CreateCommission creates a Commission for the given org and insight type.
// The delivery week is drawn from the org's delivery distribution.
// The failure roll is NOT made here; it fires on delivery (see TickDelivery).
func CreateCommission(
	org config.OrgDefinition,
	insightType config.InsightType,
	scope string,
	topicKey string,
	week int,
	rng *rand.Rand,
) Commission {
	deliveryWeek := DrawDeliveryWeek(org.DeliveryDist, week, rng)
	return Commission{
		ID:               org.ID + "_" + scope + "_" + itoa(week),
		OrgID:            org.ID,
		InsightType:      insightType,
		Scope:            scope,
		CommissionedWeek: week,
		DeliveryWeek:     deliveryWeek,
		BudgetCost:       org.BaseCost,
	}
}

// itoa converts an int to a string without importing strconv (avoids unused import).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// TickDelivery advances all commissions by one week. Commissions whose DeliveryWeek
// equals currentWeek are marked Delivered (or Failed via the failure roll).
// Returns the commissions that were newly delivered this week.
func TickDelivery(
	commissions []Commission,
	orgDefs map[string]config.OrgDefinition,
	currentWeek int,
	rng *rand.Rand,
) (updated []Commission, delivered []Commission) {
	updated = make([]Commission, 0, len(commissions))
	delivered = make([]Commission, 0)
	for _, c := range commissions {
		if c.Delivered || c.Failed {
			updated = append(updated, c)
			continue
		}
		if currentWeek >= c.DeliveryWeek {
			def, ok := orgDefs[c.OrgID]
			if ok && rng.Float64() < def.BaseFailureProbability {
				c.Failed = true
			} else {
				c.Delivered = true
			}
			delivered = append(delivered, c)
		}
		updated = append(updated, c)
	}
	return updated, delivered
}

// ---------------------------------------------------------------------------
// Report generation
// ---------------------------------------------------------------------------

// GenerateReport produces an InsightReport from a delivered Commission.
// rawValue is the true 0-100 value that the simulation computed for the topic.
// Quality is sampled from the org's quality range, then adjusted for specialism.
// The reported value is the raw value distorted by quality and bias.
func GenerateReport(
	commission Commission,
	orgDef config.OrgDefinition,
	topicKey string,
	rawValue float64,
	recentDecisionBias float64, // -1 to +1; used by ClientConfirmation only
	rng *rand.Rand,
) InsightReport {
	// 1. Draw quality from the org's quality range (uniform).
	quality := orgDef.Quality.Min + rng.Float64()*(orgDef.Quality.Max-orgDef.Quality.Min)

	// 2. Specialism check.
	hasSpecialism := false
	for _, s := range orgDef.Specialisms {
		if s == commission.InsightType {
			hasSpecialism = true
			break
		}
	}
	if !hasSpecialism {
		quality = mathutil.Clamp(quality-outsideSpecialismQualityPenalty, 0, 100)
	}

	// 3. Quality distortion: lower quality -> more noise.
	noiseFactor := 1.0 - quality/100.0 // 0 at perfect quality, 1 at zero quality
	qualityNoise := rng.NormFloat64() * noiseFactor * 15.0
	distorted := rawValue + qualityNoise

	// 4. Bias distortion.
	reported := ApplyBias(distorted, orgDef, recentDecisionBias, rng)
	reported = mathutil.Clamp(reported, 0, 100)

	return InsightReport{
		CommissionID:    commission.ID,
		OrgID:           commission.OrgID,
		InsightType:     commission.InsightType,
		Scope:           commission.Scope,
		RawValue:        rawValue,
		ReportedValue:   reported,
		TopicKey:        topicKey,
		DeliveryWeek:    commission.DeliveryWeek,
		QualityScore:    quality,
		SpecialismBonus: hasSpecialism,
	}
}

// ApplyBias distorts a value according to the org's bias type.
// recentDecisionBias is a caller-supplied summary in [-1, +1] of the player's
// recent decisions (positive = green-leaning, negative = fossil-leaning).
func ApplyBias(
	value float64,
	orgDef config.OrgDefinition,
	recentDecisionBias float64,
	rng *rand.Rand,
) float64 {
	switch orgDef.BiasType {
	case config.BiasNone:
		return value + rng.NormFloat64()*noneNoiseSigma

	case config.BiasIdeological:
		// Shift value toward the org's fixed position.
		shift := orgDef.BiasDirection * ideologicalBiasStrength
		return value + shift

	case config.BiasClientConfirmation:
		// Shift value toward validating the player's recent direction.
		shift := recentDecisionBias * orgDef.ClientBiasWeight * clientBiasStrength
		return value + shift

	default:
		return value
	}
}

// ---------------------------------------------------------------------------
// Relationship management
// ---------------------------------------------------------------------------

// UpdateOrgRelationship applies a relationship event delta and natural decay.
// Call once per week per org with the appropriate event (or call with no-event
// variant to apply decay only -- pass empty string or a custom value of zero).
func UpdateOrgRelationship(org OrgState, event RelationshipEvent, currentWeek int) OrgState {
	delta := 0.0
	switch event {
	case RelationshipEventDelivered:
		delta = relationshipDeliveredDelta
	case RelationshipEventActedOn:
		delta = relationshipActedOnDelta
	case RelationshipEventIgnored:
		delta = relationshipIgnoredDelta
	case RelationshipEventFailed:
		delta = relationshipFailedDelta
		org.CoolingOffUntil = currentWeek + coolingOffWeeks
	}
	// Natural decay toward 50
	decay := (org.RelationshipScore - orgRelDecayTarget) * orgRelDecayRate
	org.RelationshipScore = mathutil.Clamp(org.RelationshipScore-decay+delta, 0, 100)
	return org
}

// MuracanOrgAvailable returns whether a Murican-origin org is currently accessible.
// These organisations are unlocked only when a Ticky-mechanic minister is in cabinet
// AND the player has previously accepted a Ticky offer.
func MuracanOrgAvailable(
	org config.OrgDefinition,
	tickyPresent bool,
	tickyPressureAccepted bool,
) bool {
	if org.Origin != config.OrgMurican {
		// Non-Murican orgs have their own availability rules; always visible here.
		return true
	}
	return tickyPresent && tickyPressureAccepted
}
