package tui

import (
	"os"
	"testing"
	"time"

	"golang.org/x/term"
)

// Headless CI (no TTY) must fail fast instead of hanging.
//
// Bubble Tea blocks on Windows console input (getConsoleInput syscall) while
// it errors on Linux. Without an explicit guard, any picker opened headless
// hangs windows-latest for the full 10m go-test timeout (see
// internal/player TestChooseBatchSubtitlesHeadlessKeepsTracks). These tests
// pin the fast-fail contract: each public entry point that opens a real
// terminal screen returns within seconds when stdin is not a TTY.

func requireHeadless(t *testing.T) {
	t.Helper()
	if term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("requires non-interactive stdin")
	}
}

func TestHasTerminal_HeadlessIsFalse(t *testing.T) {
	requireHeadless(t)
	if HasTerminal() {
		t.Fatal("HasTerminal should be false without a TTY")
	}
}

func TestPickActioned_HeadlessFailsFast(t *testing.T) {
	requireHeadless(t)
	done := make(chan error, 1)
	go func() {
		_, _, err := PickActioned([]PickItem{{Label: "A"}}, PickOptions{})
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("PickActioned should error without a TTY")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("PickActioned hung without a TTY (deadlocks Windows CI)")
	}
}

func TestSelectSubtitleTrack_HeadlessFailsFast(t *testing.T) {
	requireHeadless(t)
	done := make(chan error, 1)
	go func() {
		_, err := SelectSubtitleTrack(SubtitleOptions{
			Tracks: []SubtitleTrack{{Label: "PT", Language: "por"}},
		})
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("SelectSubtitleTrack should error without a TTY")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("SelectSubtitleTrack hung without a TTY (deadlocks Windows CI)")
	}
}

func TestSearchHome_HeadlessFailsFast(t *testing.T) {
	requireHeadless(t)
	done := make(chan error, 1)
	go func() {
		_, err := SearchHome(SearchHomeOptions{})
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("SearchHome should error without a TTY")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("SearchHome hung without a TTY (deadlocks Windows CI)")
	}
}
