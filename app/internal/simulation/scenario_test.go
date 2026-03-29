package simulation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/save"
	"github.com/vibe-code-challenge/twenty-fifty/internal/stakeholder"
)

// loadScenario returns a freshly seeded WorldState for the given scenario ID.
func loadScenario(t *testing.T, id config.ScenarioID) WorldState {
	t.Helper()
	cfg, err := config.Load()
	require.NoError(t, err)
	return NewWorldFromScenario(cfg, fixedSeed, config.ScenarioByID(id))
}

// cabinetStakeholderID returns the stakeholder ID occupying the given role,
// or "" if the role is vacant.
func cabinetStakeholderID(w WorldState, role config.Role) string {
	return w.Government.CabinetByRole[role]
}

// findStakeholder returns the Stakeholder with the given ID, and whether it
// was found.
func findStakeholder(w WorldState, id string) (stakeholder.Stakeholder, bool) {
	for _, s := range w.Stakeholders {
		if s.ID == id {
			return s, true
		}
	}
	return stakeholder.Stakeholder{}, false
}

// ---------------------------------------------------------------------------
// NewWorldFromScenario -- HumbleBeginnings (2010)
// ---------------------------------------------------------------------------

func TestNewWorldFromScenario_HumbleBeginnings_StartYear(t *testing.T) {
	w := loadScenario(t, config.ScenarioHumbleBeginnings)
	assert.Equal(t, 2010, w.Year)
	assert.Equal(t, 2010, w.StartYear)
	assert.Equal(t, 0, w.Week)
}

func TestNewWorldFromScenario_HumbleBeginnings_RulingParty(t *testing.T) {
	w := loadScenario(t, config.ScenarioHumbleBeginnings)
	assert.Equal(t, config.PartyRight, w.Government.RulingParty)
}

func TestNewWorldFromScenario_HumbleBeginnings_DefaultCabinet(t *testing.T) {
	w := loadScenario(t, config.ScenarioHumbleBeginnings)
	// Default 2010 cabinet: Cavendish (Leader), Drake (Chancellor),
	// Stafford (ForeignSecretary), Holm (Energy).
	assert.Equal(t, "daniel_cavendish", cabinetStakeholderID(w, config.RoleLeader))
	assert.Equal(t, "philip_drake", cabinetStakeholderID(w, config.RoleChancellor))
	assert.Equal(t, "andrew_stafford", cabinetStakeholderID(w, config.RoleForeignSecretary))
	assert.Equal(t, "rupert_holm", cabinetStakeholderID(w, config.RoleEnergy))
}

func TestNewWorldFromScenario_HumbleBeginnings_ScandalMultiplierIsOne(t *testing.T) {
	w := loadScenario(t, config.ScenarioHumbleBeginnings)
	assert.Equal(t, 1.0, w.ScandalRateMultiplier)
}

func TestNewWorldFromScenario_HumbleBeginnings_MatchesNewWorld(t *testing.T) {
	// NewWorld must produce identical initial state to NewWorldFromScenario(HumbleBeginnings).
	cfg, err := config.Load()
	require.NoError(t, err)
	seed := save.MasterSeed(99)
	wOld := NewWorld(cfg, seed)
	wNew := NewWorldFromScenario(cfg, seed, config.ScenarioByID(config.ScenarioHumbleBeginnings))
	assert.Equal(t, wOld.Year, wNew.Year)
	assert.Equal(t, wOld.Government.RulingParty, wNew.Government.RulingParty)
	assert.Equal(t, wOld.GovernmentPopularity, wNew.GovernmentPopularity)
	assert.Equal(t, wOld.FossilDependency, wNew.FossilDependency)
	assert.Equal(t, wOld.Government.CabinetByRole, wNew.Government.CabinetByRole)
}

// ---------------------------------------------------------------------------
// NewWorldFromScenario -- RisingStorm (2019)
// ---------------------------------------------------------------------------

func TestNewWorldFromScenario_RisingStorm_StartYear(t *testing.T) {
	w := loadScenario(t, config.ScenarioRisingStorm)
	assert.Equal(t, 2019, w.Year)
	assert.Equal(t, 2019, w.StartYear)
}

func TestNewWorldFromScenario_RisingStorm_RulingParty(t *testing.T) {
	w := loadScenario(t, config.ScenarioRisingStorm)
	assert.Equal(t, config.PartyRight, w.Government.RulingParty)
}

func TestNewWorldFromScenario_RisingStorm_JacksonIsLeader(t *testing.T) {
	w := loadScenario(t, config.ScenarioRisingStorm)
	leaderID := cabinetStakeholderID(w, config.RoleLeader)
	assert.Equal(t, "noris_jackson", leaderID,
		"Rising Storm: noris_jackson must be forced to Leader role")
}

func TestNewWorldFromScenario_RisingStorm_TruscottIsChancellor(t *testing.T) {
	w := loadScenario(t, config.ScenarioRisingStorm)
	chancellorID := cabinetStakeholderID(w, config.RoleChancellor)
	assert.Equal(t, "dawn_truscott", chancellorID,
		"Rising Storm: dawn_truscott must be forced to Chancellor role")
}

func TestNewWorldFromScenario_RisingStorm_JacksonIsUnlocked(t *testing.T) {
	w := loadScenario(t, config.ScenarioRisingStorm)
	s, ok := findStakeholder(w, "noris_jackson")
	require.True(t, ok)
	assert.True(t, s.IsUnlocked, "noris_jackson must be unlocked in Rising Storm")
	assert.Equal(t, stakeholder.MinisterStateAppointed, s.State)
}

func TestNewWorldFromScenario_RisingStorm_TruscottIsUnlocked(t *testing.T) {
	w := loadScenario(t, config.ScenarioRisingStorm)
	s, ok := findStakeholder(w, "dawn_truscott")
	require.True(t, ok)
	assert.True(t, s.IsUnlocked, "dawn_truscott must be unlocked in Rising Storm")
}

func TestNewWorldFromScenario_RisingStorm_ScandalMultiplierIsTwo(t *testing.T) {
	w := loadScenario(t, config.ScenarioRisingStorm)
	assert.Equal(t, 2.0, w.ScandalRateMultiplier)
}

func TestNewWorldFromScenario_RisingStorm_ElectionDueEarlier(t *testing.T) {
	w := loadScenario(t, config.ScenarioRisingStorm)
	// Rising Storm has a shorter first mandate: 104 weeks vs default 260.
	assert.Equal(t, 104, w.Government.ElectionDueWeek)
}

func TestNewWorldFromScenario_RisingStorm_TexitEventsPrefired(t *testing.T) {
	w := loadScenario(t, config.ScenarioRisingStorm)
	require.NotNil(t, w.FiredOnceEvents)
	assert.True(t, w.FiredOnceEvents["texit_campaign_begins"])
	assert.True(t, w.FiredOnceEvents["texit_sovereignty_pivot"])
	assert.True(t, w.FiredOnceEvents["texit_settled"])
}

func TestNewWorldFromScenario_RisingStorm_TechMoreMatureThan2010(t *testing.T) {
	wBase := loadScenario(t, config.ScenarioHumbleBeginnings)
	wStorm := loadScenario(t, config.ScenarioRisingStorm)
	// Offshore wind should be more mature in 2019 than in 2010.
	base := wBase.Tech.Maturities[config.TechOffshoreWind]
	storm := wStorm.Tech.Maturities[config.TechOffshoreWind]
	assert.Greater(t, storm, base,
		"offshore wind maturity must be higher in Rising Storm (2019) than HumbleBeginnings (2010)")
}

// ---------------------------------------------------------------------------
// NewWorldFromScenario -- Crossroads (2026)
// ---------------------------------------------------------------------------

func TestNewWorldFromScenario_Crossroads_StartYear(t *testing.T) {
	w := loadScenario(t, config.ScenarioCrossroads)
	assert.Equal(t, 2026, w.Year)
	assert.Equal(t, 2026, w.StartYear)
}

func TestNewWorldFromScenario_Crossroads_RulingParty(t *testing.T) {
	w := loadScenario(t, config.ScenarioCrossroads)
	assert.Equal(t, config.PartyLeft, w.Government.RulingParty)
}

func TestNewWorldFromScenario_Crossroads_ReeveIsLeader(t *testing.T) {
	w := loadScenario(t, config.ScenarioCrossroads)
	leaderID := cabinetStakeholderID(w, config.RoleLeader)
	assert.Equal(t, "david_reeve", leaderID,
		"Crossroads: david_reeve must be forced to Leader role")
}

func TestNewWorldFromScenario_Crossroads_ReeveIsUnlocked(t *testing.T) {
	w := loadScenario(t, config.ScenarioCrossroads)
	s, ok := findStakeholder(w, "david_reeve")
	require.True(t, ok)
	assert.True(t, s.IsUnlocked, "david_reeve must be unlocked in Crossroads despite TimingSuccessor")
}

func TestNewWorldFromScenario_Crossroads_ScandalMultiplierIsOne(t *testing.T) {
	w := loadScenario(t, config.ScenarioCrossroads)
	assert.Equal(t, 1.0, w.ScandalRateMultiplier)
}

func TestNewWorldFromScenario_Crossroads_TechMoreMatureThan2019(t *testing.T) {
	wStorm := loadScenario(t, config.ScenarioRisingStorm)
	wCross := loadScenario(t, config.ScenarioCrossroads)
	base := wStorm.Tech.Maturities[config.TechOffshoreWind]
	cross := wCross.Tech.Maturities[config.TechOffshoreWind]
	assert.Greater(t, cross, base,
		"offshore wind maturity must be higher in Crossroads (2026) than Rising Storm (2019)")
}

func TestNewWorldFromScenario_Crossroads_LowerFossilDependency(t *testing.T) {
	wBase := loadScenario(t, config.ScenarioHumbleBeginnings)
	wCross := loadScenario(t, config.ScenarioCrossroads)
	assert.Less(t, wCross.FossilDependency, wBase.FossilDependency,
		"Crossroads 2026 must start with lower fossil dependency than 2010")
}

// ---------------------------------------------------------------------------
// Month calculation
// ---------------------------------------------------------------------------

func TestPhaseClockAdvance_Month_FirstWeek(t *testing.T) {
	w := loadWorld(t)
	w, _ = AdvanceWeek(w, nil)
	assert.Equal(t, 1, w.Month, "week 1 should be month 1 (January)")
}

func TestPhaseClockAdvance_Month_MidYear(t *testing.T) {
	w := loadWorld(t)
	w = advanceN(w, 26)
	// week 26: weekOfYear = (26-1)%52 = 25; month = 1 + 25*12/52 = 1+5 = 6
	assert.Equal(t, 6, w.Month, "week 26 should be around month 6 (June)")
}

func TestPhaseClockAdvance_Month_LastWeek(t *testing.T) {
	w := loadWorld(t)
	w = advanceN(w, 52)
	// week 52: weekOfYear = (52-1)%52 = 51; month = 1 + 51*12/52 = 1+11 = 12
	assert.Equal(t, 12, w.Month, "week 52 should be month 12 (December)")
}

func TestPhaseClockAdvance_Year_RisingStorm(t *testing.T) {
	w := loadScenario(t, config.ScenarioRisingStorm)
	// Advance one full year.
	w = advanceN(w, 52)
	assert.Equal(t, 2020, w.Year, "after 52 weeks from 2019, year should be 2020")
}

func TestPhaseClockAdvance_Year_Crossroads(t *testing.T) {
	w := loadScenario(t, config.ScenarioCrossroads)
	w = advanceN(w, 52)
	assert.Equal(t, 2027, w.Year, "after 52 weeks from 2026, year should be 2027")
}

// ---------------------------------------------------------------------------
// Headless playtests -- all three scenarios
// ---------------------------------------------------------------------------

func TestHeadlessRun_HumbleBeginnings_520Weeks(t *testing.T) {
	w := loadScenario(t, config.ScenarioHumbleBeginnings)
	_, report := HeadlessRun(w, 520)
	assert.Empty(t, report.StakeholderIssues)
	assert.GreaterOrEqual(t, report.EventsFired, 1)
	assert.GreaterOrEqual(t, report.FinalGovtPop, 0.0)
	assert.LessOrEqual(t, report.FinalGovtPop, 100.0)
}

func TestHeadlessRun_HumbleBeginnings_1040Weeks(t *testing.T) {
	w := loadScenario(t, config.ScenarioHumbleBeginnings)
	_, report := HeadlessRun(w, 1040)
	assert.Empty(t, report.StakeholderIssues)
	assert.GreaterOrEqual(t, report.FinalGovtPop, 0.0)
	assert.LessOrEqual(t, report.FinalGovtPop, 100.0)
}

func TestHeadlessRun_RisingStorm_520Weeks(t *testing.T) {
	w := loadScenario(t, config.ScenarioRisingStorm)
	_, report := HeadlessRun(w, 520)
	assert.Empty(t, report.StakeholderIssues)
	assert.GreaterOrEqual(t, report.FinalGovtPop, 0.0)
	assert.LessOrEqual(t, report.FinalGovtPop, 100.0)
}

func TestHeadlessRun_RisingStorm_1040Weeks(t *testing.T) {
	w := loadScenario(t, config.ScenarioRisingStorm)
	_, report := HeadlessRun(w, 1040)
	assert.Empty(t, report.StakeholderIssues)
	assert.GreaterOrEqual(t, report.FinalGovtPop, 0.0)
	assert.LessOrEqual(t, report.FinalGovtPop, 100.0)
}

func TestHeadlessRun_RisingStorm_1560Weeks(t *testing.T) {
	// 1560 weeks = 30 years; Rising Storm starts 2019, ends 2049.
	w := loadScenario(t, config.ScenarioRisingStorm)
	_, report := HeadlessRun(w, 1560)
	assert.Empty(t, report.StakeholderIssues)
	assert.GreaterOrEqual(t, report.FinalGovtPop, 0.0)
}

func TestHeadlessRun_Crossroads_520Weeks(t *testing.T) {
	w := loadScenario(t, config.ScenarioCrossroads)
	_, report := HeadlessRun(w, 520)
	assert.Empty(t, report.StakeholderIssues)
	assert.GreaterOrEqual(t, report.FinalGovtPop, 0.0)
	assert.LessOrEqual(t, report.FinalGovtPop, 100.0)
}

func TestHeadlessRun_Crossroads_1040Weeks(t *testing.T) {
	w := loadScenario(t, config.ScenarioCrossroads)
	_, report := HeadlessRun(w, 1040)
	assert.Empty(t, report.StakeholderIssues)
	assert.GreaterOrEqual(t, report.FinalGovtPop, 0.0)
	assert.LessOrEqual(t, report.FinalGovtPop, 100.0)
}

func TestHeadlessRun_Crossroads_1248Weeks(t *testing.T) {
	// 1248 weeks = 24 years; Crossroads starts 2026, target end 2050.
	w := loadScenario(t, config.ScenarioCrossroads)
	_, report := HeadlessRun(w, 1248)
	assert.Empty(t, report.StakeholderIssues)
	assert.GreaterOrEqual(t, report.FinalGovtPop, 0.0)
}

// ---------------------------------------------------------------------------
// BaseWeeklyMt is scenario-calibrated
// ---------------------------------------------------------------------------

func TestNewWorldFromScenario_BaseWeeklyMt_ScenarioCalibrated(t *testing.T) {
	wHB := loadScenario(t, config.ScenarioHumbleBeginnings)
	wRS := loadScenario(t, config.ScenarioRisingStorm)
	wCR := loadScenario(t, config.ScenarioCrossroads)
	// Each scenario should have a lower base weekly Mt than the previous.
	assert.Greater(t, wHB.BaseWeeklyMt, wRS.BaseWeeklyMt,
		"HumbleBeginnings 2010 must have higher base emissions than RisingStorm 2019")
	assert.Greater(t, wRS.BaseWeeklyMt, wCR.BaseWeeklyMt,
		"RisingStorm 2019 must have higher base emissions than Crossroads 2026")
}
