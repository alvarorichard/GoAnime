package util

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
)

// Saved source choices: what the user switched on or off in the TUI's SOURCES
// screen. A choice is "on", "off", or absent (the source's own default). It is
// a SEARCH choice only: a source switched off here stays resolvable, so history
// saved from it still opens (see internal/api/source/enablement.go for why the
// two must not be conflated). The env lists (GOANIME_ENABLED_SOURCES /
// GOANIME_DISABLED_SOURCES) outrank it, which keeps scripts and CI
// predictable whatever a desktop session saved.

// sourcePrefsFile is the file name inside the goanime config folder.
const sourcePrefsFile = "sources.json"

// sourcePrefsVersion is the file format; a newer file than this is not read.
const sourcePrefsVersion = 1

var (
	sourcePrefsMu sync.RWMutex
	sourcePrefs   = map[string]bool{}

	// writeSourcePrefs is the seam tests replace to fail the write.
	writeSourcePrefs = writeFileAtomic
	// userConfigDir is the seam tests replace to fail the lookup.
	userConfigDir = os.UserConfigDir
)

type sourcePrefsDoc struct {
	Version int             `json:"version"`
	Sources map[string]bool `json:"sources"`
}

// SourcePref returns the saved choice for name. set is false when the user
// never chose, so the source's own default applies.
func SourcePref(name string) (on, set bool) {
	token := canonSourceToken(name)
	if token == "" {
		return false, false
	}
	sourcePrefsMu.RLock()
	defer sourcePrefsMu.RUnlock()
	on, set = sourcePrefs[token]
	return on, set
}

// SetSourcePref records a choice in memory (not on disk). An empty name is
// ignored.
func SetSourcePref(name string, on bool) {
	token := canonSourceToken(name)
	if token == "" {
		return
	}
	sourcePrefsMu.Lock()
	sourcePrefs[token] = on
	sourcePrefsMu.Unlock()
}

// ClearSourcePref forgets the choice for name, back to the source's default.
func ClearSourcePref(name string) {
	token := canonSourceToken(name)
	sourcePrefsMu.Lock()
	delete(sourcePrefs, token)
	sourcePrefsMu.Unlock()
}

// ResetSourcePrefs forgets every choice (memory only).
func ResetSourcePrefs() {
	sourcePrefsMu.Lock()
	sourcePrefs = map[string]bool{}
	sourcePrefsMu.Unlock()
}

// SourcePrefsPath is where the choices live: <user config dir>/goanime/sources.json.
func SourcePrefsPath() (string, error) {
	dir, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("source preferences: no config folder: %w", err)
	}
	return filepath.Join(dir, "goanime", sourcePrefsFile), nil
}

// LoadSourcePrefs replaces the in-memory choices with the file's. A missing
// file is not an error (nothing chosen yet). A file that cannot be read or
// parsed leaves the current choices untouched and returns the error, so a
// corrupt file never silently flips sources.
func LoadSourcePrefs(path string) error {
	data, err := os.ReadFile(path) // #nosec G304 -- the user's own config file
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("source preferences: read %s: %w", path, err)
	}
	var doc sourcePrefsDoc
	if err := jsonx.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("source preferences: %s is not valid: %w", path, err)
	}
	if doc.Version > sourcePrefsVersion {
		return fmt.Errorf("source preferences: %s is version %d, this GoAnime reads up to %d", path, doc.Version, sourcePrefsVersion)
	}
	next := make(map[string]bool, len(doc.Sources))
	for name, on := range doc.Sources {
		if token := canonSourceToken(name); token != "" {
			next[token] = on
		}
	}
	sourcePrefsMu.Lock()
	sourcePrefs = next
	sourcePrefsMu.Unlock()
	return nil
}

// LoadSavedSourcePrefs loads the choices from their usual place at startup.
func LoadSavedSourcePrefs() error {
	path, err := SourcePrefsPath()
	if err != nil {
		return err
	}
	return LoadSourcePrefs(path)
}

// SaveSourcePrefs writes the in-memory choices to path atomically.
func SaveSourcePrefs(path string) error {
	sourcePrefsMu.RLock()
	snapshot := make(map[string]bool, len(sourcePrefs))
	for k, v := range sourcePrefs {
		snapshot[k] = v
	}
	sourcePrefsMu.RUnlock()

	data, err := jsonx.Marshal(sourcePrefsDoc{Version: sourcePrefsVersion, Sources: snapshot})
	if err != nil {
		return fmt.Errorf("source preferences: encode: %w", err)
	}
	if err := writeSourcePrefs(path, data); err != nil {
		return fmt.Errorf("source preferences: save %s: %w", path, err)
	}
	return nil
}

// PersistSourcePref records a choice and saves it. If the save fails the
// previous choice is restored, so memory never claims more than the disk will
// remember and the caller can tell the user.
func PersistSourcePref(path, name string, on bool) error {
	prevOn, hadPrev := SourcePref(name)
	SetSourcePref(name, on)
	if err := SaveSourcePrefs(path); err != nil {
		if hadPrev {
			SetSourcePref(name, prevOn)
		} else {
			ClearSourcePref(name)
		}
		return err
	}
	return nil
}

// writeFileAtomic writes data next to path and renames it over, so a crash
// never leaves half a file. The folder is created when missing.
func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".sources-*.tmp")
	if err != nil {
		return err
	}
	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if err := errors.Join(werr, cerr); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return nil
}
