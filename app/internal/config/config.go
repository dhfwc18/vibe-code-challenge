package config

import "fmt"

// Config holds all static game data loaded at startup.
// It is read-only after Load() returns; no field should ever be mutated at runtime.
type Config struct {
	Stakeholders  []StakeholderSeed
	Organisations []OrgDefinition
	Companies     []CompanyDef
	Technologies  []TechCurveDef
	CarbonBudgets []CarbonBudgetEntry
	Regions       []RegionDef
	Tiles         []TileDef
	PolicyCards   []PolicyCardDef
	Events        []EventDef
}

// Load assembles the Config from the package-level definition slices and validates
// basic structural invariants. It returns an error if any constraint is violated,
// so callers can fail fast at startup rather than at runtime.
func Load() (*Config, error) {
	cfg := &Config{
		Stakeholders:  stakeholderSeeds,
		Organisations: orgDefinitions,
		Companies:     companyDefs,
		Technologies:  techCurveDefs,
		CarbonBudgets: carbonBudgets,
		Regions:       regionDefs,
		Tiles:         tileDefs,
		PolicyCards:   policyCardDefs,
		Events:        eventDefs,
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	return cfg, nil
}

// validate checks structural invariants across all definition slices.
func (c *Config) validate() error {
	if err := checkUniqueIDs("stakeholder", stakeholderIDs(c.Stakeholders)); err != nil {
		return err
	}
	if err := checkUniqueIDs("organisation", orgIDs(c.Organisations)); err != nil {
		return err
	}
	if err := checkUniqueIDs("company", companyIDList(c.Companies)); err != nil {
		return err
	}
	if err := checkUniqueIDs("technology", techIDs(c.Technologies)); err != nil {
		return err
	}
	if err := checkUniqueIDs("region", regionIDList(c.Regions)); err != nil {
		return err
	}
	if err := checkUniqueIDs("tile", tileIDList(c.Tiles)); err != nil {
		return err
	}
	if err := checkUniqueIDs("policy_card", policyIDs(c.PolicyCards)); err != nil {
		return err
	}
	if err := checkUniqueIDs("event", eventIDList(c.Events)); err != nil {
		return err
	}

	if err := c.validateTileRegionRefs(); err != nil {
		return err
	}
	if err := c.validateRegionTileRefs(); err != nil {
		return err
	}
	if err := c.validateCarbonBudgetOrder(); err != nil {
		return err
	}

	return nil
}

// validateTileRegionRefs checks that every TileDef.RegionID names a known region.
func (c *Config) validateTileRegionRefs() error {
	known := make(map[string]bool, len(c.Regions))
	for _, r := range c.Regions {
		known[r.ID] = true
	}
	for _, t := range c.Tiles {
		if !known[t.RegionID] {
			return fmt.Errorf("tile %q references unknown region %q", t.ID, t.RegionID)
		}
	}
	return nil
}

// validateRegionTileRefs checks that every RegionDef.TileID names a known tile.
func (c *Config) validateRegionTileRefs() error {
	known := make(map[string]bool, len(c.Tiles))
	for _, t := range c.Tiles {
		known[t.ID] = true
	}
	for _, r := range c.Regions {
		for _, tid := range r.TileIDs {
			if !known[tid] {
				return fmt.Errorf("region %q references unknown tile %q", r.ID, tid)
			}
		}
	}
	return nil
}

// validateCarbonBudgetOrder checks that carbon budget years are strictly ascending.
func (c *Config) validateCarbonBudgetOrder() error {
	for i := 1; i < len(c.CarbonBudgets); i++ {
		if c.CarbonBudgets[i].Year <= c.CarbonBudgets[i-1].Year {
			return fmt.Errorf(
				"carbon budget years not strictly ascending: entry %d year %d <= entry %d year %d",
				i, c.CarbonBudgets[i].Year, i-1, c.CarbonBudgets[i-1].Year,
			)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// ID extraction helpers
// ---------------------------------------------------------------------------

func stakeholderIDs(ss []StakeholderSeed) []string {
	ids := make([]string, len(ss))
	for i, s := range ss {
		ids[i] = s.ID
	}
	return ids
}

func orgIDs(os []OrgDefinition) []string {
	ids := make([]string, len(os))
	for i, o := range os {
		ids[i] = o.ID
	}
	return ids
}

func companyIDList(cs []CompanyDef) []string {
	ids := make([]string, len(cs))
	for i, c := range cs {
		ids[i] = c.ID
	}
	return ids
}

func techIDs(ts []TechCurveDef) []string {
	ids := make([]string, len(ts))
	for i, t := range ts {
		ids[i] = string(t.ID)
	}
	return ids
}

func regionIDList(rs []RegionDef) []string {
	ids := make([]string, len(rs))
	for i, r := range rs {
		ids[i] = r.ID
	}
	return ids
}

func tileIDList(ts []TileDef) []string {
	ids := make([]string, len(ts))
	for i, t := range ts {
		ids[i] = t.ID
	}
	return ids
}

func policyIDs(ps []PolicyCardDef) []string {
	ids := make([]string, len(ps))
	for i, p := range ps {
		ids[i] = p.ID
	}
	return ids
}

func eventIDList(es []EventDef) []string {
	ids := make([]string, len(es))
	for i, e := range es {
		ids[i] = e.ID
	}
	return ids
}

// checkUniqueIDs returns an error if any ID appears more than once in the slice.
func checkUniqueIDs(kind string, ids []string) error {
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id == "" {
			return fmt.Errorf("%s definition has empty ID", kind)
		}
		if seen[id] {
			return fmt.Errorf("%s ID %q is duplicated", kind, id)
		}
		seen[id] = true
	}
	return nil
}
