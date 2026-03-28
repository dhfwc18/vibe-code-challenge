package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// currentVersion is incremented whenever the SaveState schema changes in a way
// that is incompatible with previous save files. Saves with a different version
// are rejected at load time.
const currentVersion = 1

// SaveState is the top-level structure written to disk. All game state that
// must survive a quit-and-reload lives here. Packages populate their own
// sub-structs; this file only owns the envelope and I/O logic.
type SaveState struct {
	Version    int        `json:"version"`
	MasterSeed MasterSeed `json:"master_seed"`
	GameWeek   int        `json:"game_week"`  // weeks elapsed since 2010-01-01
	GameYear   int        `json:"game_year"`  // derived cache of the current calendar year
	PlayerName string     `json:"player_name"`
}

// NewSaveState creates a fresh save state for a new game.
// MasterSeed is generated from OS entropy; it must not be called on load.
func NewSaveState(playerName string) (*SaveState, error) {
	seed, err := NewMasterSeed()
	if err != nil {
		return nil, fmt.Errorf("new save state: %w", err)
	}
	return &SaveState{
		Version:    currentVersion,
		MasterSeed: seed,
		GameWeek:   0,
		GameYear:   2010,
		PlayerName: playerName,
	}, nil
}

// Write serialises s to the file at path, creating or truncating it.
// The directory must already exist.
func Write(path string, s *SaveState) error {
	if s == nil {
		return errors.New("save: cannot write nil SaveState")
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("save marshal: %w", err)
	}
	// Write to a temp file then rename for atomicity.
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "save_*.tmp")
	if err != nil {
		return fmt.Errorf("save create temp: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("save write temp: %w", err)
	}
	if err = tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("save close temp: %w", err)
	}
	if err = os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("save rename: %w", err)
	}
	return nil
}

// Read deserialises a SaveState from the file at path.
// Returns ErrIncompatibleVersion if the file was written by a different schema.
func Read(path string) (*SaveState, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("save open: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("save read: %w", err)
	}

	var s SaveState
	if err = json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("save unmarshal: %w", err)
	}
	if s.Version != currentVersion {
		return nil, fmt.Errorf("%w: file version %d, current version %d",
			ErrIncompatibleVersion, s.Version, currentVersion)
	}
	return &s, nil
}

// ErrIncompatibleVersion is returned when the save file schema version does not
// match the running binary. The caller should offer to start a new game.
var ErrIncompatibleVersion = errors.New("incompatible save version")
