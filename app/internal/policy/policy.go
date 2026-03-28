package policy

import (
	"twenty-fifty/internal/config"
	"twenty-fifty/internal/mathutil"
	"twenty-fifty/internal/region"
	"twenty-fifty/internal/stakeholder"
)

// significanceRefuseConflict is the ideology conflict threshold above which a
// minister issues a formal hard refusal for policies of the given significance,
// provided the card has also been stalled for the corresponding number of weeks.
const majorSignificanceRefuseConflict    = 75.0
const majorSignificanceRefuseWeeks      = 8
const moderateSignificanceRefuseConflict = 110.0
const moderateSignificanceRefuseWeeks   = 16

// PolicyState tracks where a policy card sits in the approval and lifecycle pipeline.
// Transition logic runs in the simulation layer; this package provides pure functions
// that return updated PolicyCard values.
type PolicyState string

const (
	// PolicyStateDraft is the initial state for all seeded cards.
	PolicyStateDraft PolicyState = "DRAFT"
	// PolicyStateUnderReview is entered when the player submits the card.
	// The card stays here until all approval steps resolve.
	PolicyStateUnderReview PolicyState = "UNDER_REVIEW"
	// PolicyStateApproved means all steps passed; card awaits activation.
	PolicyStateApproved PolicyState = "APPROVED"
	// PolicyStateRejected means at least one step failed the ideology hard gate.
	// Cards in this state cannot be resubmitted without a reshuffle.
	PolicyStateRejected PolicyState = "REJECTED"
	// PolicyStateActive means the card is in force and producing weekly effects.
	PolicyStateActive PolicyState = "ACTIVE"
	// PolicyStateArchived is the terminal state for deactivated or cancelled cards.
	PolicyStateArchived PolicyState = "ARCHIVED"
)

// ArchiveReason records why a policy left the ACTIVE or APPROVED state.
type ArchiveReason string

const (
	ArchiveReasonCompleted  ArchiveReason = "COMPLETED"  // policy ran to natural end
	ArchiveReasonSuperseded ArchiveReason = "SUPERSEDED" // replaced by a successor policy
	ArchiveReasonCancelled  ArchiveReason = "CANCELLED"  // player chose to cancel
	ArchiveReasonBudgetCut  ArchiveReason = "BUDGET_CUT" // insufficient budget to continue
)

// PolicyCard combines the immutable definition from config with the runtime
// approval and lifecycle state.
type PolicyCard struct {
	Def              config.PolicyCardDef
	State            PolicyState
	WeeksActive      int  // incremented each week while ACTIVE
	WeeksUnderReview int  // incremented each week while UNDER_REVIEW
	StepsCleared     int  // number of ApprovalSteps that have been approved so far
	ArchiveReason    ArchiveReason
}

// CarbonDelta is the output of ResolveWeeklyEffect: how much carbon reduction
// (negative = less emissions) the policy produced this week, and its budget cost.
type CarbonDelta struct {
	Sector            config.PolicySector
	DeltaMt           float64 // MtCO2e; negative = reduction, positive = increase
	BudgetCostPerWeek float64 // GBP millions this week
}

// SeedPolicyCards creates one PolicyCard per definition, all starting in DRAFT.
func SeedPolicyCards(defs []config.PolicyCardDef) []PolicyCard {
	cards := make([]PolicyCard, 0, len(defs))
	for _, d := range defs {
		cards = append(cards, PolicyCard{
			Def:   d,
			State: PolicyStateDraft,
		})
	}
	return cards
}

// IsUnlocked returns true if the tech unlock gate for this card has been met.
// Cards with no gate (TechUnlockGate == "") are always unlocked.
func IsUnlocked(card PolicyCard, techMaturity map[config.Technology]float64) bool {
	if card.Def.TechUnlockGate == "" {
		return true
	}
	maturity, ok := techMaturity[card.Def.TechUnlockGate]
	if !ok {
		return false
	}
	return maturity >= card.Def.TechUnlockThreshold
}

// SubmitPolicy moves a DRAFT card to UNDER_REVIEW. It is a no-op for any other state.
func SubmitPolicy(card PolicyCard) PolicyCard {
	if card.State != PolicyStateDraft {
		return card
	}
	card.State = PolicyStateUnderReview
	return card
}

// IdeologyConflict returns the absolute difference between a stakeholder's ideology
// and the policy sector's ideology position. Range: [0, 200].
func IdeologyConflict(def config.PolicyCardDef, s stakeholder.Stakeholder) float64 {
	policyPos := stakeholder.PolicyIdeologyPosition(def.Sector)
	diff := s.IdeologyScore - policyPos
	if diff < 0 {
		diff = -diff
	}
	return diff
}

// EvaluateApprovalStep tests one approval requirement against the matching stakeholder.
// Returns (approved, hardReject):
//   - hardReject=true if ideology conflict exceeds MaxIdeologyConflict (permanent block).
//   - hardReject=true if significance-based refusal threshold is crossed after sustained stalling.
//   - approved=true if relationship meets MinRelationshipScore and no hard reject.
//   - approved=false, hardReject=false means the step is not yet cleared (relationship
//     shortfall); the card stays UNDER_REVIEW.
func EvaluateApprovalStep(
	card PolicyCard,
	def config.PolicyCardDef,
	s stakeholder.Stakeholder,
	req config.ApprovalRequirement,
) (approved bool, hardReject bool) {
	if IdeologyConflict(def, s) > req.MaxIdeologyConflict {
		return false, true
	}
	// Significance-based formal refusal: a minister who strongly disagrees with a
	// high-significance policy will issue a hard refusal after sustained review,
	// even if the per-step ideology threshold was not crossed.
	conflict := IdeologyConflict(def, s)
	switch def.Significance {
	case config.PolicySignificanceMajor:
		if conflict > majorSignificanceRefuseConflict && card.WeeksUnderReview >= majorSignificanceRefuseWeeks {
			return false, true
		}
	case config.PolicySignificanceModerate:
		if conflict > moderateSignificanceRefuseConflict && card.WeeksUnderReview >= moderateSignificanceRefuseWeeks {
			return false, true
		}
	}
	if s.RelationshipScore < req.MinRelationshipScore {
		return false, false
	}
	return true, false
}

// EvaluateApproval runs through the outstanding approval steps for a card that is
// UNDER_REVIEW. It matches each ApprovalRequirement to the unlocked stakeholder with
// the correct role.
//
//   - If any step produces a hard reject the card moves to REJECTED (permanent).
//   - If all remaining steps are approved the card moves to APPROVED.
//   - Otherwise the card stays UNDER_REVIEW (relationship not yet sufficient).
//
// Only unlocked stakeholders are considered. If no unlocked stakeholder matches a
// required role the step is skipped (not cleared and not rejected).
func EvaluateApproval(card PolicyCard, stakeholders []stakeholder.Stakeholder) PolicyCard {
	if card.State != PolicyStateUnderReview {
		return card
	}

	// Build a role -> stakeholder lookup (first unlocked match wins).
	byRole := make(map[config.Role]stakeholder.Stakeholder)
	for _, s := range stakeholders {
		if !s.IsUnlocked {
			continue
		}
		if _, exists := byRole[s.Role]; !exists {
			byRole[s.Role] = s
		}
	}

	steps := card.Def.ApprovalSteps
	for i := card.StepsCleared; i < len(steps); i++ {
		req := steps[i]
		s, found := byRole[req.Role]
		if !found {
			// No unlocked minister in this role; cannot clear or reject the step yet.
			break
		}
		approved, hardReject := EvaluateApprovalStep(card, card.Def, s, req)
		if hardReject {
			card.State = PolicyStateRejected
			return card
		}
		if !approved {
			// Relationship shortfall; stop here and wait.
			break
		}
		card.StepsCleared++
	}

	if card.StepsCleared >= len(steps) {
		card.State = PolicyStateApproved
	}
	return card
}

// ActivatePolicy moves an APPROVED card to ACTIVE. It is a no-op for any other state.
func ActivatePolicy(card PolicyCard) PolicyCard {
	if card.State != PolicyStateApproved {
		return card
	}
	card.State = PolicyStateActive
	return card
}

// ArchivePolicy moves a card to ARCHIVED for the given reason. It works from any state
// except ARCHIVED (which is a no-op to prevent double-archiving).
func ArchivePolicy(card PolicyCard, reason ArchiveReason) PolicyCard {
	if card.State == PolicyStateArchived {
		return card
	}
	card.State = PolicyStateArchived
	card.ArchiveReason = reason
	return card
}

// TickActive increments WeeksActive on an ACTIVE card. No-op for other states.
func TickActive(card PolicyCard) PolicyCard {
	if card.State != PolicyStateActive {
		return card
	}
	card.WeeksActive++
	return card
}

// TickUnderReview increments WeeksUnderReview on an UNDER_REVIEW card. No-op otherwise.
func TickUnderReview(card PolicyCard) PolicyCard {
	if card.State != PolicyStateUnderReview {
		return card
	}
	card.WeeksUnderReview++
	return card
}

// ResolveWeeklyEffect computes the carbon reduction and budget cost for one active
// policy card this week. Returns a zero CarbonDelta if the card is not ACTIVE.
//
// The delta is modulated by up to three optional multipliers drawn from the
// WeeklyEffectDef:
//   - CapacityDependent: multiplied by region.CapacityMultiplier (0-1).
//   - TechDependent: multiplied by techMaturityFraction (0-1).
//   - RetrofitDependent: multiplied by trueRetrofitRate (0-1).
//
// Multipliers stack multiplicatively. A card with all three flags and a region at
// half capacity, 80% tech maturity, and 0.6 retrofit rate would apply a combined
// factor of 0.5 * 0.8 * 0.6 = 0.24.
func ResolveWeeklyEffect(
	card PolicyCard,
	r region.Region,
	techMaturityFraction float64,
	trueRetrofitRate float64,
) CarbonDelta {
	if card.State != PolicyStateActive {
		return CarbonDelta{}
	}
	ef := card.Def.WeeklyEffect
	base := ef.BaseCarbonDeltaMt

	multiplier := 1.0
	if ef.CapacityDependent {
		multiplier *= region.CapacityMultiplier(r)
	}
	if ef.TechDependent {
		multiplier *= mathutil.Clamp(techMaturityFraction, 0, 1)
	}
	if ef.RetrofitDependent {
		multiplier *= mathutil.Clamp(trueRetrofitRate, 0, 1)
	}

	return CarbonDelta{
		Sector:            ef.Sector,
		DeltaMt:           base * multiplier,
		BudgetCostPerWeek: ef.BudgetCostPerWeek,
	}
}
