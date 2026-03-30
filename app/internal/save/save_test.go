package save

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// MasterSeed / DeriveSubSeed
// ---------------------------------------------------------------------------

func TestNewMasterSeed_ReturnsDifferentValuesEachCall(t *testing.T) {
	a, err := NewMasterSeed()
	require.NoError(t, err)
	b, err := NewMasterSeed()
	require.NoError(t, err)
	assert.NotEqual(t, a, b, "two calls to NewMasterSeed should produce different seeds")
}

func TestDeriveSubSeed_SameSeedSameSubsystem_Deterministic(t *testing.T) {
	m := MasterSeed(123456789)
	first := m.DeriveSubSeed("tech")
	second := m.DeriveSubSeed("tech")
	assert.Equal(t, first, second, "DeriveSubSeed must be deterministic")
}

func TestDeriveSubSeed_DifferentSubsystems_ProduceDifferentValues(t *testing.T) {
	m := MasterSeed(987654321)
	tech := m.DeriveSubSeed("tech")
	events := m.DeriveSubSeed("events")
	assert.NotEqual(t, tech, events, "different subsystem names should yield different sub-seeds")
}

func TestDeriveSubSeed_DifferentMasterSeeds_ProduceDifferentValues(t *testing.T) {
	a := MasterSeed(1)
	b := MasterSeed(2)
	assert.NotEqual(t, a.DeriveSubSeed("tech"), b.DeriveSubSeed("tech"),
		"different master seeds should yield different sub-seeds for the same subsystem")
}

// ---------------------------------------------------------------------------
// NewSaveState
// ---------------------------------------------------------------------------

func TestNewSaveState_ValidPlayerName_SetsDefaults(t *testing.T) {
	s, err := NewSaveState("Alice")
	require.NoError(t, err)
	assert.Equal(t, currentVersion, s.Version)
	assert.Equal(t, 0, s.GameWeek)
	assert.Equal(t, 2010, s.GameYear)
	assert.Equal(t, "Alice", s.PlayerName)
	assert.NotZero(t, s.MasterSeed, "MasterSeed must be non-zero")
}

// ---------------------------------------------------------------------------
// Write + Read round-trip
// ---------------------------------------------------------------------------

func TestWriteRead_RoundTrip_PreservesAllFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sav")

	original, err := NewSaveState("TestPlayer")
	require.NoError(t, err)
	original.GameWeek = 52
	original.GameYear = 2011

	require.NoError(t, Write(path, original))

	loaded, err := Read(path)
	require.NoError(t, err)
	assert.Equal(t, original.Version, loaded.Version)
	assert.Equal(t, original.MasterSeed, loaded.MasterSeed)
	assert.Equal(t, original.GameWeek, loaded.GameWeek)
	assert.Equal(t, original.GameYear, loaded.GameYear)
	assert.Equal(t, original.PlayerName, loaded.PlayerName)
}

func TestWrite_NilSaveState_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nil.sav")
	err := Write(path, nil)
	assert.Error(t, err)
}

func TestRead_NonExistentFile_ReturnsError(t *testing.T) {
	_, err := Read("/nonexistent/path/game.sav")
	assert.Error(t, err)
}

func TestRead_CorruptJSON_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "corrupt.sav")
	require.NoError(t, os.WriteFile(path, []byte("{not valid json"), 0o600))
	_, err := Read(path)
	assert.Error(t, err)
}

func TestRead_WrongVersion_ReturnsErrIncompatibleVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "old.sav")

	// Write a save with a different schema version.
	data := []byte(`{"version":999,"master_seed":42,"game_week":0,"game_year":2010,"player_name":"Old"}`)
	require.NoError(t, os.WriteFile(path, data, 0o600))

	_, err := Read(path)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrIncompatibleVersion),
		"expected ErrIncompatibleVersion, got: %v", err)
}

func TestWrite_Atomic_DoesNotLeakTempFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "atomic.sav")

	s, err := NewSaveState("AtomicTest")
	require.NoError(t, err)
	require.NoError(t, Write(path, s))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	// Only the final save file should exist; no .tmp leftovers.
	assert.Equal(t, 1, len(entries), "temp file must be cleaned up after atomic write")
	assert.Equal(t, "atomic.sav", entries[0].Name())
}

func TestWrite_OverwriteExistingFile_Succeeds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "overwrite.sav")

	first, err := NewSaveState("First")
	require.NoError(t, err)
	require.NoError(t, Write(path, first))

	second, err := NewSaveState("Second")
	require.NoError(t, err)
	second.GameWeek = 100
	require.NoError(t, Write(path, second))

	loaded, err := Read(path)
	require.NoError(t, err)
	assert.Equal(t, "Second", loaded.PlayerName)
	assert.Equal(t, 100, loaded.GameWeek)
}
