package player

import "twenty-fifty/internal/mathutil"

// ---------------------------------------------------------------------------
// Enumerations
// ---------------------------------------------------------------------------

// StaffRole identifies the type of support staff the player can hire.
type StaffRole string

const (
	StaffRoleAnalyst     StaffRole = "ANALYST"      // +1 AP/week, +evidence quality
	StaffRoleAdvisor     StaffRole = "ADVISOR"       // +1 AP/week, improved approval odds
	StaffRoleCommsOfficer StaffRole = "COMMS_OFFICER" // +1 AP/week, faster LCR polling
	StaffRoleChiefOfStaff StaffRole = "CHIEF_OF_STAFF" // +2 AP/week, all bonuses reduced
)

// ActionType identifies categories of player action recorded in the action log.
type ActionType string

const (
	ActionTypeSubmitPolicy     ActionType = "SUBMIT_POLICY"
	ActionTypeCommissionReport ActionType = "COMMISSION_REPORT"
	ActionTypeLobbyMinister    ActionType = "LOBBY_MINISTER"
	ActionTypeHireStaff        ActionType = "HIRE_STAFF"
	ActionTypeFireStaff        ActionType = "FIRE_STAFF"
	ActionTypeShockResponse    ActionType = "SHOCK_RESPONSE"
	ActionTypeOther            ActionType = "OTHER"
)

// ---------------------------------------------------------------------------
// Structs
// ---------------------------------------------------------------------------

// StaffMember represents one member of the player's support team.
type StaffMember struct {
	ID         string
	Role       StaffRole
	APBonus    int  // AP added to the weekly pool while this member is hired
	WeekHired  int
}

// ActionRecord logs one player action for history and reputation tracking.
type ActionRecord struct {
	ActionType ActionType
	Week       int
	APCost     int
	Detail     string // brief free-text summary (e.g. policy ID or minister ID)
}

// CivilServant holds all player state: AP, staff, reputation, and action history.
type CivilServant struct {
	APRemaining   int
	ActionHistory []ActionRecord
	Staff         []StaffMember
	Reputation    float64 // 0-100; displayed as a civil service grade label
}

// ---------------------------------------------------------------------------
// Calibration constants
// ---------------------------------------------------------------------------

const (
	baseAPPool = 5 // AP available per week before staff bonuses

	// Reputation grade boundaries (lower bound of each grade band).
	reputationGeneralist        = 0.0
	reputationHigherExecutive   = 15.0
	reputationSeniorExecutive   = 28.0
	reputationGrade7            = 40.0
	reputationGrade6            = 52.0
	reputationGrade5            = 63.0
	reputationGrade4            = 74.0
	reputationDeputySecretary   = 85.0
	reputationPermanentSecretary = 95.0
)

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

// NewCivilServant creates a fresh player state at game start.
func NewCivilServant() CivilServant {
	return CivilServant{
		APRemaining:   baseAPPool,
		ActionHistory: []ActionRecord{},
		Staff:         []StaffMember{},
		Reputation:    20.0, // starts as a Higher Executive Officer
	}
}

// ---------------------------------------------------------------------------
// AP management
// ---------------------------------------------------------------------------

// WeeklyAPPool returns the total AP available this week (base + staff bonuses).
func WeeklyAPPool(cs CivilServant) int {
	pool := baseAPPool
	for _, s := range cs.Staff {
		pool += s.APBonus
	}
	return pool
}

// StartWeekAPPool resets APRemaining to the full weekly pool.
func StartWeekAPPool(cs CivilServant) CivilServant {
	cs.APRemaining = WeeklyAPPool(cs)
	return cs
}

// SpendAP deducts cost from APRemaining. Returns the updated CivilServant and
// true on success, or the unchanged CivilServant and false if insufficient AP.
func SpendAP(cs CivilServant, cost int) (CivilServant, bool) {
	if cs.APRemaining < cost {
		return cs, false
	}
	cs.APRemaining -= cost
	return cs, true
}

// ---------------------------------------------------------------------------
// Action recording
// ---------------------------------------------------------------------------

// RecordAction appends an action to the history log.
func RecordAction(cs CivilServant, record ActionRecord) CivilServant {
	cs.ActionHistory = append(cs.ActionHistory, record)
	return cs
}

// ---------------------------------------------------------------------------
// Staff management
// ---------------------------------------------------------------------------

// HireStaff adds a staff member to the roster.
func HireStaff(cs CivilServant, member StaffMember) CivilServant {
	cs.Staff = append(cs.Staff, member)
	return cs
}

// FireStaff removes the staff member with the given ID. No-op if not found.
func FireStaff(cs CivilServant, staffID string) CivilServant {
	updated := make([]StaffMember, 0, len(cs.Staff))
	for _, s := range cs.Staff {
		if s.ID != staffID {
			updated = append(updated, s)
		}
	}
	cs.Staff = updated
	return cs
}

// TotalWeeklyStaffCost returns the total salary cost per week for all staff.
// Each staff role has a fixed weekly cost in GBP thousands.
func TotalWeeklyStaffCost(cs CivilServant) float64 {
	total := 0.0
	for _, s := range cs.Staff {
		total += staffRoleCost(s.Role)
	}
	return total
}

// staffRoleCost returns the weekly cost in GBP thousands for a given role.
func staffRoleCost(role StaffRole) float64 {
	switch role {
	case StaffRoleAnalyst:
		return 2.0
	case StaffRoleAdvisor:
		return 3.0
	case StaffRoleCommsOfficer:
		return 2.5
	case StaffRoleChiefOfStaff:
		return 5.0
	default:
		return 0.0
	}
}

// ---------------------------------------------------------------------------
// Reputation
// ---------------------------------------------------------------------------

// ReputationGrade maps a 0-100 reputation score to a civil service grade label.
// Labels follow Taitan's equivalent of the UK civil service grade structure.
func ReputationGrade(rep float64) string {
	rep = mathutil.Clamp(rep, 0, 100)
	switch {
	case rep >= reputationPermanentSecretary:
		return "Permanent Secretary"
	case rep >= reputationDeputySecretary:
		return "Deputy Secretary"
	case rep >= reputationGrade4:
		return "Director General"
	case rep >= reputationGrade5:
		return "Director"
	case rep >= reputationGrade6:
		return "Deputy Director"
	case rep >= reputationGrade7:
		return "Grade 7"
	case rep >= reputationSeniorExecutive:
		return "Senior Executive Officer"
	case rep >= reputationHigherExecutive:
		return "Higher Executive Officer"
	default:
		return "Generalist"
	}
}
