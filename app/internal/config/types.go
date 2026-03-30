package config

// ---------------------------------------------------------------------------
// Enumerations
// ---------------------------------------------------------------------------

// Party identifies the four political parties in Taitan.
// The string value is the stable ID used in save files and config data.
type Party string

const (
	PartyLeft     Party = "common_wealth"      // The Common Wealth
	PartyRight    Party = "union_party"         // The Union Party
	PartyFarLeft  Party = "renewal"             // Renewal
	PartyFarRight Party = "taitan_restoration"  // Taitan Restoration
)

// PartyNames maps each Party ID to its display name shown in the UI.
var PartyNames = map[Party]string{
	PartyLeft:     "The Common Wealth",
	PartyRight:    "The Union Party",
	PartyFarLeft:  "Renewal",
	PartyFarRight: "Taitan Restoration",
}

// Role identifies the four ministerial roles within each party.
type Role string

const (
	RoleLeader           Role = "LEADER"
	RoleChancellor       Role = "CHANCELLOR"
	RoleForeignSecretary Role = "FOREIGN_SECRETARY"
	RoleEnergy           Role = "ENERGY"
)

// EntryTiming controls when a stakeholder seed becomes available for selection.
type EntryTiming string

const (
	TimingStart     EntryTiming = "START"     // available at game start (2010)
	TimingMid       EntryTiming = "MID"       // enters pool approx 2015-2022
	TimingLate      EntryTiming = "LATE"      // enters pool approx 2023-2035
	TimingSuccessor EntryTiming = "SUCCESSOR" // only after a specific departure trigger
)

// SpecialMechanic flags unique event-generating behaviours on individual stakeholders.
type SpecialMechanic string

const (
	MechanicNone             SpecialMechanic = ""
	MechanicTickyPressure    SpecialMechanic = "TICKY_PRESSURE"
	MechanicDizzySurge       SpecialMechanic = "DIZZY_SURGE"
	MechanicElectoralFatigue SpecialMechanic = "ELECTORAL_FATIGUE"
)

// OrgType classifies advisory organisations by delivery model.
type OrgType string

const (
	OrgConsultancy OrgType = "CONSULTANCY"
	OrgThinkTank   OrgType = "THINK_TANK"
	OrgAcademic    OrgType = "ACADEMIC"
)

// OrgOrigin classifies advisory organisations by country of operation.
type OrgOrigin string

const (
	OrgLocal   OrgOrigin = "LOCAL"   // Taitan-based
	OrgMurican OrgOrigin = "MURICAN" // Murica-based; surfaced only via TICKY_PRESSURE
	OrgForeign OrgOrigin = "FOREIGN" // other foreign
)

// BiasType controls how an organisation's reports are distorted.
type BiasType string

const (
	BiasClientConfirmation BiasType = "CLIENT_CONFIRMATION" // shifts toward validating player decisions
	BiasIdeological        BiasType = "IDEOLOGICAL"         // shifts toward fixed political position
	BiasNone               BiasType = "NONE"                // small Gaussian noise only
)

// InsightType classifies what a commission is investigating.
type InsightType string

const (
	InsightPower        InsightType = "POWER"
	InsightTransport    InsightType = "TRANSPORT"
	InsightBuildings    InsightType = "BUILDINGS"
	InsightIndustry     InsightType = "INDUSTRY"
	InsightEconomy      InsightType = "ECONOMY"
	InsightPolicy       InsightType = "POLICY"
	InsightClimate      InsightType = "CLIMATE"
	InsightFuelPoverty  InsightType = "FUEL_POVERTY"
	InsightRetrofit     InsightType = "RETROFIT"
	InsightEnergyMarket InsightType = "ENERGY_MARKET"
)

// TechCategory maps LCT companies to the technology sector they operate in.
type TechCategory string

const (
	TechCatOffshoreWind  TechCategory = "OFFSHORE_WIND"
	TechCatOnshore       TechCategory = "ONSHORE_SOLAR"
	TechCatHeatPumps     TechCategory = "HEAT_PUMPS"
	TechCatEVs           TechCategory = "EVS"
	TechCatHydrogen      TechCategory = "HYDROGEN"
	TechCatCCUS          TechCategory = "CCUS"
	TechCatGrid          TechCategory = "GRID_RETAIL"
	TechCatLegacy        TechCategory = "LEGACY_TRANSITION"
	TechCatInstallers    TechCategory = "INSTALLERS"
	TechCatNuclear       TechCategory = "NUCLEAR"
)

// CompanySize describes the scale of an LCT company at game start.
type CompanySize string

const (
	CompanyStartup       CompanySize = "STARTUP"
	CompanySME           CompanySize = "SME"
	CompanyLarge         CompanySize = "LARGE"
	CompanyMultinational CompanySize = "MULTINATIONAL"
)

// Technology names the eight decarbonisation technologies tracked by the tech model.
type Technology string

const (
	TechOffshoreWind  Technology = "OFFSHORE_WIND"
	TechOnshoreWind   Technology = "ONSHORE_WIND"
	TechSolarPV       Technology = "SOLAR_PV"
	TechNuclear       Technology = "NUCLEAR"
	TechHeatPumps     Technology = "HEAT_PUMPS"
	TechEVs           Technology = "EVS"
	TechHydrogen      Technology = "HYDROGEN"
	TechIndustrialCCS Technology = "INDUSTRIAL_CCS"
)

// Sector names the five emission sectors tracked in carbon accounting.
type Sector string

const (
	SectorPower     Sector = "POWER"
	SectorTransport Sector = "TRANSPORT"
	SectorBuildings Sector = "BUILDINGS"
	SectorIndustry  Sector = "INDUSTRY"
	SectorOther     Sector = "OTHER"
)

// HeatingType describes the primary heating system in a tile.
type HeatingType string

const (
	HeatingGas      HeatingType = "GAS"
	HeatingOil      HeatingType = "OIL"
	HeatingElectric HeatingType = "ELECTRIC"
	HeatingHeatPump HeatingType = "HEAT_PUMP"
	HeatingMixed    HeatingType = "MIXED"
)

// EventType classifies global events by their origin.
type EventType string

const (
	EventWeather       EventType = "WEATHER"
	EventEnergyShock   EventType = "ENERGY_SHOCK"
	EventInternational EventType = "INTERNATIONAL"
	EventEconomic      EventType = "ECONOMIC"
	EventSocial        EventType = "SOCIAL"
	EventTech          EventType = "TECHNOLOGICAL"
)

// EventSeverity classifies global event magnitude.
type EventSeverity string

const (
	SeverityMinor    EventSeverity = "MINOR"
	SeverityModerate EventSeverity = "MODERATE"
	SeverityMajor    EventSeverity = "MAJOR"
)

// PolicySignificance classifies the political weight and transformative impact
// of a policy card. Used to scale ideology conflict accumulation and to gate
// minister hard-refusals on sustained opposition.
type PolicySignificance string

const (
	PolicySignificanceMinor    PolicySignificance = "MINOR"
	PolicySignificanceModerate PolicySignificance = "MODERATE"
	PolicySignificanceMajor    PolicySignificance = "MAJOR"
)

// PolicySector classifies which emission sector a policy card targets.
type PolicySector string

const (
	PolicySectorPower     PolicySector = "POWER"
	PolicySectorTransport PolicySector = "TRANSPORT"
	PolicySectorBuildings PolicySector = "BUILDINGS"
	PolicySectorIndustry  PolicySector = "INDUSTRY"
	PolicySectorCross     PolicySector = "CROSS_CUTTING"
)

// ---------------------------------------------------------------------------
// Composite value types
// ---------------------------------------------------------------------------

// TriangularDist defines a triangular probability distribution for delivery times.
type TriangularDist struct {
	Min  int // weeks, minimum delivery
	Mode int // weeks, most likely delivery
	Max  int // weeks, maximum delivery
}

// QualityRange defines the minimum and maximum quality an organisation can achieve.
type QualityRange struct {
	Min float64 // 0-100
	Max float64 // 0-100
}

// ---------------------------------------------------------------------------
// Static definition structs (config data; never mutated at runtime)
// ---------------------------------------------------------------------------

// StakeholderSeed is the immutable definition of a political figure loaded from config.
// Runtime state (popularity, relationship, state machine) lives in the stakeholder package.
type StakeholderSeed struct {
	ID                  string
	Party               Party
	Role                Role
	EntryTiming         EntryTiming
	EntryWeekMin        int // earliest week this figure can enter the pool (MID/LATE only)
	EntryWeekMax        int // latest week
	Name                string
	Nickname            string
	Biography           string
	IdeologyScore       float64 // -100 (far-left) to +100 (far-right)
	NetZeroSympathy     float64 // 0 (hostile) to 100 (champion)
	RiskTolerance       float64 // 0 (cautious) to 100 (reckless)
	PopulismScore       float64 // 0 (technocratic) to 100 (populist)
	DiplomaticSkill     float64 // 0-100; used only for Foreign Secretary role
	ConsultancyAffinity []string // org IDs; drives passive relationship bonus
	ConsultancyAversion bool     // true = hostile to private consultancy spend; commissioning while governing costs minister relationship
	Signals             []string // 2-3 observable personality signals shown on appointment
	SpecialMechanic     SpecialMechanic
}

// OrgDefinition is the immutable definition of an advisory organisation.
type OrgDefinition struct {
	ID                     string
	Name                   string
	OrgType                OrgType
	Origin                 OrgOrigin
	// MuricanAccessTier controls when a Murican-origin org becomes accessible.
	// Ignored for non-Murican orgs (always accessible).
	//   0 = available from game start
	//   1 = unlocked by any Murican-related international event firing (or Ticky)
	//   2 = unlocked only via the Ticky pressure mechanic
	MuricanAccessTier      int
	BaseCost               float64        // GBP thousands per commission
	DeliveryDist           TriangularDist // weeks
	Quality                QualityRange
	BiasType               BiasType
	BiasDirection          float64      // -1 to +1; used only when BiasType=IDEOLOGICAL
	ClientBiasWeight       float64      // 0-1; used only when BiasType=CLIENT_CONFIRMATION
	PopularityRisk         float64      // 0-1; weekly hit to GovernmentPopularity when active
	BaseFailureProbability float64      // 0-1; probability commission fails entirely
	Specialisms            []InsightType // outside specialism = quality -20
}

// CompanyDef is the immutable definition of an LCT company seed.
type CompanyDef struct {
	ID           string
	Name         string
	TechCategory TechCategory
	OriginSize   CompanySize
	BaseQuality  float64 // 0-100
	BaseWorkRate float64 // 0-100
	TaitonHQ     bool    // false = multinational with foreign parent
}

// TechCurveDef defines the adoption curve for one decarbonisation technology.
type TechCurveDef struct {
	ID                Technology
	Name              string
	Sector            Sector
	LogisticMidpoint  float64 // game week at which maturity reaches 50%
	LogisticSteepness float64 // controls slope of the S-curve
	BaseAdoptionRate  float64 // weekly maturity increase without any policy boost
	InitialMaturity   float64 // maturity at game start (week 0 / year 2010)
}

// CarbonBudgetEntry records the legally binding annual carbon limit for a given year.
// Values sourced from CCC carbon budget targets (see docs/green_book_reference.md).
type CarbonBudgetEntry struct {
	Year              int
	AnnualLimitMtCO2e float64
}

// RegionDef is the immutable seed for one of the 12 Taitan regions.
type RegionDef struct {
	ID                       string
	Name                     string
	TileIDs                  []string
	Tags                     []string // geographic tags e.g. "coastal", "rural", "urban", "industrial", "agricultural"
	InitialSkillsNetwork     float64  // 0-100
	InitialInstallerCapacity float64  // installs per week
	InitialSupplyChain       float64  // 0-100
}

// TileDef is the immutable seed for one map tile.
type TileDef struct {
	ID                      string
	RegionID                string
	Name                    string
	InitialInsulationLevel  float64     // 0-100
	InitialHeatingType      HeatingType
	InitialLocalIncome      float64 // 0-100 (50 = median Taitan household income)
	InitialPoliticalOpinion float64 // 0-100 (50 = neutral)
	InitialHeatingCapacity  float64 // 0-100
	InitialInstallerQuality float64 // 0-100
}

// ApprovalRequirement defines what a minister must accept before a policy advances.
type ApprovalRequirement struct {
	Role                 Role
	MinRelationshipScore float64 // player's relationship with this role must be >= this
	MaxIdeologyConflict  float64 // ideology conflict score must be <= this
}

// WeeklyEffectDef describes the ongoing impact of an active policy each week.
type WeeklyEffectDef struct {
	Sector            PolicySector
	BaseCarbonDeltaMt float64 // MtCO2e reduction per week at full capacity
	CapacityDependent bool    // multiply by regional InstallerCapacity fraction
	TechDependent     bool    // multiply by TechMaturity fraction
	RetrofitDependent bool    // use TrueRetrofitRate (not Observed)
	BudgetCostPerWeek float64 // GBP millions per week
}

// PolicyCardDef is the immutable definition of a policy card.
type PolicyCardDef struct {
	ID                  string
	Name                string
	Sector              PolicySector
	Description         string
	APCost              int
	BudgetCostToSubmit  float64 // GBP millions one-off cost to submit
	TechUnlockGate      Technology // zero value = no gate required
	TechUnlockThreshold float64    // minimum TechMaturity to unlock card
	ApprovalSteps       []ApprovalRequirement
	WeeklyEffect        WeeklyEffectDef
	LCRDeltaPerWeek       float64            // weekly change to LowCarbonReputation when active
	PopularityRiskPerWeek float64            // weekly GovernmentPopularity delta when active
	Significance          PolicySignificance // political weight: MINOR, MODERATE, or MAJOR
	RDBonus map[Technology]float64 // weekly tech maturity acceleration per tech when ACTIVE; nil = no R&D effect
}

// EventEffect describes the quantitative impact of a global event when it fires.
// Global fields apply to the whole country. Targeted fields require matching filters.
//
// Region targeting: RegionFilter is a comma-separated list of tags or region IDs.
// Recognised values: "COASTAL", "RURAL", "URBAN", "INDUSTRIAL", "AGRICULTURAL", or a region ID.
// An empty filter means the effect applies to ALL regions.
//
// Stakeholder targeting: StakeholderFilter matches by role or special value.
// Recognised values: "CABINET" (all 4 role-holders), "ROLE:LEADER", "ROLE:CHANCELLOR",
// "ROLE:FOREIGN_SECRETARY", "ROLE:ENERGY", "ALL". Empty = no stakeholder effect.
//
// Company targeting: CompanyFilter matches by tech category or special value.
// Recognised values: "ALL", "TECH:OFFSHORE_WIND", "TECH:ONSHORE_SOLAR", "TECH:HEAT_PUMPS",
// "TECH:EVS", "TECH:HYDROGEN", "TECH:CCUS", "TECH:GRID_RETAIL", "TECH:LEGACY_TRANSITION",
// "TECH:INSTALLERS", "TECH:NUCLEAR". Empty = no company effect.
type EventEffect struct {
	// Global effects
	GasPriceDeltaPct         float64 // percentage change to GasPrice
	ElectricityPriceDeltaPct float64
	OilPriceDeltaPct         float64
	EconomyDelta             float64 // direct delta to hidden Economy (0-100 scale)
	LCRDelta                 float64 // direct delta to LowCarbonReputation
	GovtPopularityDelta      float64
	CarbonEmissionsDeltaMt   float64 // additional MtCO2e this week (positive = more emissions)

	// Region-targeted effects (filter selects which regions are affected)
	RegionFilter             string  // see filter rules above
	InstallerCapacityDelta   float64 // installs-per-week delta applied to matched regions
	SkillsNetworkDelta       float64 // 0-100 scale delta applied to matched regions
	TileFuelPovertyDelta     float64 // applied to every tile in matched regions
	TileInsulationDamage     float64 // positive = insulation degraded (subtracted from InsulationLevel)

	// Stakeholder-targeted effects (filter selects which ministers are affected)
	StakeholderFilter        string  // see filter rules above
	StakeholderRelDelta      float64 // delta to player relationship with matched stakeholders
	StakeholderPressureDelta int     // +1/-1 to pressure counter on matched stakeholders

	// Company-targeted effects (filter selects which active LCT companies are affected)
	CompanyFilter            string  // see filter rules above
	CompanyWorkRateDelta     float64 // 0-100 scale delta applied to matched companies
	CompanyQualityDelta      float64 // 0-100 scale delta applied to matched companies
}

// DecayingShockConfig defines an ongoing market effect that diminishes over time.
// When an EventDef with a non-zero DecayingShock fires, the simulation creates an
// ActiveDecayingShock in WorldState that applies the price deltas each week and
// multiplies them by DecayRate until WeeksRemaining reaches zero.
type DecayingShockConfig struct {
	InitialGasPctPerWeek  float64 // first-week gas price % change (e.g. 2.5 = +2.5%)
	InitialOilPctPerWeek  float64 // first-week oil price % change
	InitialElecPctPerWeek float64 // first-week electricity price % change
	DecayRate             float64 // multiply deltas by this each week (e.g. 0.93 = 7% decay)
	MaxWeeks              int     // remove shock after this many weeks
}

// EventDef is the immutable definition of a global event card.
type EventDef struct {
	ID                  string
	Name                string
	Headline            string        // short newspaper-style headline shown in event notifications
	EventType           EventType
	Severity            EventSeverity
	BaseProbability     float64       // probability per week of this event firing; 0 = not probabilistic
	ClimateMultiplier   float64       // multiplied against BaseProbability when climate is ELEVATED+
	FossilMultiplier    float64       // multiplied against BaseProbability when FossilDependency > 60
	TriggerAtYear       int           // if > 0, fires once automatically at the start of this game year
	BaseEffects         EventEffect
	DecayingShock       DecayingShockConfig // zero value = no ongoing decaying effect
	Narrative           string        // full text shown in the player-visible event log
	OffersShockResponse bool          // if true, queues a ShockResponseCard for the player
}
