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
const currentVersion = 2

// SaveState is the top-level structure written to disk. All game state that
// must survive a quit-and-reload lives here. Packages populate their own
// sub-structs; this file only owns the envelope and I/O logic.
//
// WorldData is typed as json.RawMessage so the simulation package can own
// its own WorldSaveData schema without creating an import cycle.
type SaveState struct {
	Version    int             `json:"version"`
	MasterSeed MasterSeed      `json:"master_seed"`
	GameWeek   int             `json:"game_week"` // weeks elapsed since scenario start
	GameYear   int             `json:"game_year"` // derived cache of the current calendar year
	PlayerName string          `json:"player_name"`
	WorldData  json.RawMessage `json:"world_data,omitempty"` // simulation.WorldSaveData
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

// AutoSaveDir returns the directory used for save files.
// If saveDir is empty, it uses a "saves" subdirectory next to the binary.
func AutoSaveDir(saveDir string) string {
	if saveDir != "" {
		return saveDir
	}
	exe, err := os.Executable()
	if err == nil {
		return filepath.Join(filepath.Dir(exe), "saves")
	}
	return "saves"
}

// AutoSavePath returns the path to the autosave file in the given save directory.
// Pass "" to use the default save directory (next to the binary).
func AutoSavePath(saveDir string) string {
	return filepath.Join(AutoSaveDir(saveDir), "autosave.json")
}

// EnsureSaveDir creates the save directory if it does not exist.
func EnsureSaveDir(saveDir string) error {
	return os.MkdirAll(saveDir, 0o755)
}

// Exists returns true if a file exists at path (save file presence check).
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("save write temp: %w", err)
	}
	if err = tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("save close temp: %w", err)
	}
	if err = os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
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
	defer func() { _ = f.Close() }()

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
