//go:build !cgo

package tracking

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNoCGOLocalTrackerContract(t *testing.T) {
	t.Parallel()

	if IsCgoEnabled {
		t.Fatal("IsCgoEnabled must be false in a no-CGO build")
	}

	dbPath := filepath.Join(t.TempDir(), "progress.db")
	tracker := NewLocalTracker(dbPath)
	if tracker != nil {
		t.Fatalf("NewLocalTracker(%q) = %#v, want nil when CGO is disabled", dbPath, tracker)
	}

	if _, err := os.Stat(dbPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("no-CGO tracker must not create a SQLite database, stat err = %v", err)
	}

	if err := tracker.UpdateProgress(Anime{AllanimeID: "anime:ep1", Duration: 1}); !errors.Is(err, ErrTrackerNotInited) {
		t.Fatalf("UpdateProgress on no-CGO tracker error = %v, want ErrTrackerNotInited", err)
	}

	if got, err := tracker.GetAnime(1, "anime:ep1"); got != nil || !errors.Is(err, ErrTrackerNotInited) {
		t.Fatalf("GetAnime on no-CGO tracker = (%#v, %v), want (nil, ErrTrackerNotInited)", got, err)
	}

	if got, err := tracker.GetAllAnime(); got != nil || !errors.Is(err, ErrTrackerNotInited) {
		t.Fatalf("GetAllAnime on no-CGO tracker = (%#v, %v), want (nil, ErrTrackerNotInited)", got, err)
	}
}
