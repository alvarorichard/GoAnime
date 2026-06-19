package scraper

import (
	"os"
	"testing"
	"time"
)

// TestSolverRebuildsAfterClose proves the fix for the "target closed" bug
// (added 2026-06-09): if the persistent browser context dies — e.g. the user
// closes the window — the next init() must launch a FRESH context instead of
// returning the dead handle forever (the old sync.Once behaviour).
//
// Live (launches the real bundled Chromium), env-gated.
//
//	GOANIME_RECON=1 go test ./internal/scraper/ -run TestSolverRebuildsAfterClose -v -count=1 -timeout 120s
func TestSolverRebuildsAfterClose(t *testing.T) {
	if os.Getenv("GOANIME_RECON") == "" {
		t.Skip("set GOANIME_RECON=1 (launches a real browser)")
	}

	s := &cfBrowserSolver{}
	t.Cleanup(s.Close)

	bctx1, err := s.init()
	if err != nil {
		t.Fatalf("first init: %v", err)
	}
	if bctx1 == nil {
		t.Fatal("first init returned nil context")
	}

	// Simulate the user closing the window: tear down just the context.
	if err := bctx1.Close(); err != nil {
		t.Fatalf("close context: %v", err)
	}

	// OnClose nil's s.pctx asynchronously; wait for it.
	deadline := time.Now().Add(10 * time.Second)
	for {
		s.lifeMu.Lock()
		cleared := s.pctx == nil
		s.lifeMu.Unlock()
		if cleared {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("OnClose did not clear s.pctx after context close")
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Next init() must rebuild a usable context rather than return the dead one.
	bctx2, err := s.init()
	if err != nil {
		t.Fatalf("second init (rebuild): %v", err)
	}
	if bctx2 == nil {
		t.Fatal("rebuild returned nil context")
	}

	// A fresh context must be able to open a page (the dead one could not).
	page, err := bctx2.NewPage()
	if err != nil {
		t.Fatalf("rebuilt context cannot open page: %v", err)
	}
	_ = page.Close()
}
