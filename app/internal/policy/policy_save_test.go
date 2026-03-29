package policy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vibe-code-challenge/twenty-fifty/internal/config"
	"github.com/vibe-code-challenge/twenty-fifty/internal/policy"
)

func loadDefs(t *testing.T) []config.PolicyCardDef {
	t.Helper()
	cfg, err := config.Load()
	assert.NoError(t, err)
	return cfg.PolicyCards
}

// TestSavePolicyCard_PreservesState verifies round-trip of a DRAFT card.
func TestSavePolicyCard_PreservesState(t *testing.T) {
	defs := loadDefs(t)
	cards := policy.SeedPolicyCards(defs)
	assert.NotEmpty(t, cards)

	original := cards[0]
	saved := policy.SavePolicyCard(original)
	assert.Equal(t, original.Def.ID, saved.DefID)
	assert.Equal(t, original.State, saved.State)
	assert.Equal(t, original.WeeksActive, saved.WeeksActive)
}

// TestRestorePolicyCard_LinksDefPointer verifies that restore relinks the Def pointer.
func TestRestorePolicyCard_LinksDefPointer(t *testing.T) {
	defs := loadDefs(t)
	cards := policy.SeedPolicyCards(defs)
	original := cards[0]

	saved := policy.SavePolicyCard(original)
	restored := policy.RestorePolicyCard(saved, defs)

	assert.NotNil(t, restored.Def)
	assert.Equal(t, original.Def.ID, restored.Def.ID)
}

// TestRestorePolicyCard_UnknownID_NilDef verifies that an unrecognised ID results in nil Def.
func TestRestorePolicyCard_UnknownID_NilDef(t *testing.T) {
	defs := loadDefs(t)
	saved := policy.PolicyCardSave{DefID: "does_not_exist", State: policy.PolicyStateDraft}
	restored := policy.RestorePolicyCard(saved, defs)
	assert.Nil(t, restored.Def)
}
