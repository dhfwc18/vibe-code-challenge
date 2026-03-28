package stakeholder

import (
	"math/rand"

	"twenty-fifty/internal/carbon"
	"twenty-fifty/internal/config"
	"twenty-fifty/internal/mathutil"
)

// MinisterState represents the lifecycle of a political figure in the game.
// Transition logic is deferred to the simulation layer (Layer 5).
type MinisterState string

const (
	MinisterStateActive              MinisterState = "ACTIVE"
	MinisterStateUnderPressure       MinisterState = "UNDER_PRESSURE"
	MinisterStateLeadershipChallenge MinisterState = "LEADERSHIP_CHALLENGE"
	MinisterStateDeparted            MinisterState = "DEPARTED"
	MinisterStateBackbench           MinisterState = "BACKBENCH"
	MinisterStateOppositionShadow    MinisterState = "OPPOSITION_SHADOW"
	MinisterStateAppointed           MinisterState = "APPOINTED"    // just assigned a role this week
	MinisterStateSacked              MinisterState = "SACKED"       // removed by the PM
	MinisterStateResigned            MinisterState = "RESIGNED"     // resigned voluntarily
	MinisterStateElectionOut         MinisterState = "ELECTION_OUT" // lost their seat at a general election
)

// Stakeholder holds all runtime state for one political NPC.
// Identity fields are copied from the seed at game start and never change.
// Live fields are updated each tick.
type Stakeholder struct {
	// Identity (immutable after seeding)
	ID                  string
	Party               config.Party
	Role                config.Role
	Name                string
	Nickname            string
	Biography           string
	IdeologyScore       float64 // -100 (far-left) to +100 (far-right)
	NetZeroSympathy     float64 // 0-100
	RiskTolerance       float64 // 0-100
	PopulismScore       float64 // 0-100
	DiplomaticSkill     float64 // 0-100; relevant for ForeignSecretary role
	ConsultancyAffinity []string // org IDs; drives passive relationship bonus
	ConsultancyAversion bool     // true = hostile to private consultancy spend
	Signals             []string // observable personality signals shown on appointment
	EntryWeekMin        int
	EntryWeekMax        int
	EntryTiming         config.EntryTiming
	SpecialMechanic     config.SpecialMechanic

	// Live relationship and influence
	RelationshipScore float64 // 0-100; starts at 50; decays toward 50
	InfluenceScore    float64 // 0-100; computed from party polling and role weight
	IsUnlocked        bool

	// Minister popularity (distinct from GovernmentPopularity)
	Popularity             float64 // 0-100 hidden true value; polled with sigma=5
	WeeksUnderPressure     int     // increments while State==UNDER_PRESSURE; resets on recovery
	GraceWeeksRemaining    int     // countdown after appointment; minister not held accountable during grace
	IdeologyConflictScore  float64 // accumulates when minister approves ideologically-opposed policies

	// Signal queue (appended externally by simulation layer)
	PendingSignals []string

	// Lifecycle state (transition logic in simulation layer)
	State MinisterState

	// Special mechanic counters
	TickyPressureCounter  int  // increments under ClimateLevelElevated+
	DizzySurgeActive      bool // toggled by the DizzySurge mechanic
	ElectoralFatigueCount int  // increments unconditionally each tick for ELECTORAL_FATIGUE
}

// Calibration constants for the stakeholder model.
const (
	startingRelationship  = 50.0
	decayTarget           = 50.0
	decayRate             = 0.02  // fraction of (score - 50) that reverts per week
	maxRelationshipDelta  = 5.0   // max relationship change per tick from a player action
	maxEventImpact        = 3.0   // max relationship change per tick from a single event

	tickyPressureThreshold    = 12  // weeks at ELEVATED+ before Murican-org event is signalled
	dizzySurgeProbability     = 0.08 // 8% per week chance of DizzySurge toggle
	electoralFatigueThreshold = 156  // 3 years (156 weeks) before fatigue signals departure

	// ApprovalChance weights
	ideologyAlignmentWeight = 0.40
	netZeroSympathyWeight   = 0.35
	relationshipWeight      = 0.25
)

// DefaultRoleWeights maps each role to an influence multiplier.
// Used by ComputeInfluence; callers may supply an override map.
var DefaultRoleWeights = map[config.Role]float64{
	config.RoleLeader:           1.40,
	config.RoleChancellor:       1.20,
	config.RoleForeignSecretary: 0.90,
	config.RoleEnergy:           1.00,
}

// policyIdeologyPosition maps each policy sector to its position on the
// ideology axis (-100 = far-left, +100 = far-right).
var policyIdeologyPosition = map[config.PolicySector]float64{
	config.PolicySectorPower:     0.0,
	config.PolicySectorTransport: -15.0,
	config.PolicySectorBuildings: -20.0,
	config.PolicySectorIndustry:  20.0,
	config.PolicySectorCross:     0.0,
}

// PolicyIdeologyPosition returns the ideology axis position for a given policy
// sector. Returns 0.0 for unrecognised sectors. Exported so the policy package
// can call it in EvaluateApprovalStep without duplicating the map.
func PolicyIdeologyPosition(sector config.PolicySector) float64 {
	pos, ok := policyIdeologyPosition[sector]
	if !ok {
		return 0.0
	}
	return pos
}

// SeedStakeholders creates live Stakeholder values from config seed definitions.
// All stakeholders start with RelationshipScore=50 and are unlocked only if their
// EntryTiming is TimingStart.
func SeedStakeholders(defs []config.StakeholderSeed) []Stakeholder {
	out := make([]Stakeholder, 0, len(defs))
	for _, d := range defs {
		s := Stakeholder{
			ID:                  d.ID,
			Party:               d.Party,
			Role:                d.Role,
			Name:                d.Name,
			Nickname:            d.Nickname,
			Biography:           d.Biography,
			IdeologyScore:       d.IdeologyScore,
			NetZeroSympathy:     d.NetZeroSympathy,
			RiskTolerance:       d.RiskTolerance,
			PopulismScore:       d.PopulismScore,
			DiplomaticSkill:     d.DiplomaticSkill,
			ConsultancyAffinity: d.ConsultancyAffinity,
			ConsultancyAversion: d.ConsultancyAversion,
			Signals:             d.Signals,
			EntryWeekMin:        d.EntryWeekMin,
			EntryWeekMax:        d.EntryWeekMax,
			EntryTiming:         d.EntryTiming,
			SpecialMechanic:     d.SpecialMechanic,
			RelationshipScore:   startingRelationship,
			Popularity:            50.0,
			IdeologyConflictScore: 0.0,
			IsUnlocked:            d.EntryTiming == config.TimingStart,
			State:               MinisterStateActive,
			PendingSignals:      []string{},
		}
		out = append(out, s)
	}
	return out
}

// UnlockStakeholder returns a copy of s with IsUnlocked set to true if
// currentWeek >= s.EntryWeekMin. It is a no-op if the stakeholder is already
// unlocked. It never sets IsUnlocked to false.
func UnlockStakeholder(s Stakeholder, currentWeek int) Stakeholder {
	if s.IsUnlocked {
		return s
	}
	if currentWeek >= s.EntryWeekMin {
		s.IsUnlocked = true
	}
	return s
}

// TickRelationship updates the relationship score for one week.
// Natural decay pulls the score toward 50 at rate decayRate.
// playerAction and eventImpact are clamped to their respective maxima.
func TickRelationship(s Stakeholder, playerAction, eventImpact float64) Stakeholder {
	decay := (s.RelationshipScore - decayTarget) * decayRate
	action := mathutil.Clamp(playerAction, -maxRelationshipDelta, maxRelationshipDelta)
	impact := mathutil.Clamp(eventImpact, -maxEventImpact, maxEventImpact)
	s.RelationshipScore = mathutil.Clamp(s.RelationshipScore-decay+action+impact, 0, 100)
	return s
}

// ComputeInfluence updates s.InfluenceScore from the current party polling shares
// and a role-weight map. If s.Party is absent from partyPolling, 25.0 is assumed.
// If s.Role is absent from roleWeights, 1.0 is assumed.
func ComputeInfluence(
	s Stakeholder,
	partyPolling map[config.Party]float64,
	roleWeights map[config.Role]float64,
) Stakeholder {
	polling, ok := partyPolling[s.Party]
	if !ok {
		polling = 25.0
	}
	roleFactor, ok := roleWeights[s.Role]
	if !ok {
		roleFactor = 1.0
	}
	raw := (polling*0.60 + s.RelationshipScore*0.40) * roleFactor
	s.InfluenceScore = mathutil.Clamp(raw, 0, 100)
	return s
}

// TickSpecialMechanic advances the special mechanic state for one week.
// rng is required for the DizzySurge probability roll; pass nil to skip the roll
// (useful in deterministic tests where DizzySurge behaviour is not under test).
func TickSpecialMechanic(s Stakeholder, carbonLevel carbon.ClimateLevel, rng *rand.Rand) Stakeholder {
	switch s.SpecialMechanic {
	case config.MechanicTickyPressure:
		if carbonLevel >= carbon.ClimateLevelElevated {
			s.TickyPressureCounter++
		} else {
			s.TickyPressureCounter = 0
		}
	case config.MechanicDizzySurge:
		if rng != nil && rng.Float64() < dizzySurgeProbability {
			s.DizzySurgeActive = !s.DizzySurgeActive
		}
	case config.MechanicElectoralFatigue:
		s.ElectoralFatigueCount++
	}
	return s
}

// TickyPressureThresholdReached reports whether the TickyPressure counter has
// reached the trigger threshold. The caller is responsible for taking action
// (e.g. queuing a Murican-org event) and may reset the counter externally.
func TickyPressureThresholdReached(s Stakeholder) bool {
	return s.SpecialMechanic == config.MechanicTickyPressure &&
		s.TickyPressureCounter >= tickyPressureThreshold
}

// ElectoralFatigueThresholdReached reports whether JJ Cameron's fatigue has
// accumulated past the departure threshold.
func ElectoralFatigueThresholdReached(s Stakeholder) bool {
	return s.SpecialMechanic == config.MechanicElectoralFatigue &&
		s.ElectoralFatigueCount >= electoralFatigueThreshold
}

// ApprovalChance returns the probability (in [0.05, 0.95]) that this stakeholder
// approves a given policy card during an approval step.
//
// Three factors are combined with fixed weights:
//   - Ideology alignment (40%): how closely the stakeholder's ideology matches
//     the policy sector's lean.
//   - NetZeroSympathy (35%): the stakeholder's baseline support for net-zero.
//   - RelationshipScore (25%): current player-stakeholder relationship.
func ApprovalChance(s Stakeholder, policy config.PolicyCardDef) float64 {
	// Ideology alignment: 1.0 = perfect match, 0.0 = maximum conflict
	policyPos, ok := policyIdeologyPosition[policy.Sector]
	if !ok {
		policyPos = 0.0
	}
	stakeholderNorm := (s.IdeologyScore + 100.0) / 200.0
	policyNorm := (policyPos + 100.0) / 200.0
	alignment := 1.0 - abs(stakeholderNorm-policyNorm)

	sympathy := s.NetZeroSympathy / 100.0
	relContrib := s.RelationshipScore / 100.0

	raw := ideologyAlignmentWeight*alignment +
		netZeroSympathyWeight*sympathy +
		relationshipWeight*relContrib

	return mathutil.Clamp(raw, 0.05, 0.95)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
