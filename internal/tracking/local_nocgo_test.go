//go:build !cgo

package tracking

import "testing"

func TestNewLocalTrackerWithoutCGOReturnsNil(t *testing.T) {
	tracker := NewLocalTracker(t.TempDir())
	if tracker != nil {
		t.Fatalf("expected nil tracker when CGO is disabled, got %#v", tracker)
	}
}

func TestNoCGOLocalTrackerContract(t *testing.T) {
	var tracker *LocalTracker

	if err := tracker.UpdateProgress(Anime{}); err != ErrTrackerNotInited {
		t.Fatalf("UpdateProgress error = %v, want %v", err, ErrTrackerNotInited)
	}

	gotAnime, err := tracker.GetAnime(1, "allanime-id")
	if err != ErrTrackerNotInited {
		t.Fatalf("GetAnime error = %v, want %v", err, ErrTrackerNotInited)
	}
	if gotAnime != nil {
		t.Fatalf("GetAnime returned %+v, want nil", gotAnime)
	}

	gotAll, err := tracker.GetAllAnime()
	if err != ErrTrackerNotInited {
		t.Fatalf("GetAllAnime error = %v, want %v", err, ErrTrackerNotInited)
	}
	if gotAll != nil {
		t.Fatalf("GetAllAnime returned %+v, want nil", gotAll)
	}

	if err := tracker.DeleteAnime(1, "allanime-id"); err != ErrTrackerNotInited {
		t.Fatalf("DeleteAnime error = %v, want %v", err, ErrTrackerNotInited)
	}

	if err := tracker.Close(); err != ErrTrackerNotInited {
		t.Fatalf("Close error = %v, want %v", err, ErrTrackerNotInited)
	}
}
