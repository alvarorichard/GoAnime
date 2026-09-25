package superflix

import (
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
)

// Nothing may be left holding a browser window open.
//
// The window's lifetime used to be a convention: whoever opened it was supposed
// to call ReleaseSharedBrowser at the end of its own API-level operation. That
// convention failed twice in one week, in both directions. fetchSuperFlixSeasons
// opened a window for a season list and released it only when a stream was
// eventually resolved, so it sat on screen through every picker in between — and
// never closed at all if the user backed out. Both sniff paths open a window and
// release it nowhere.
//
// A convention that has to be remembered at four call sites, in a package whose
// whole job is a browser the user can see, is the wrong mechanism. This is the
// backstop: the shared context counts who is using it, and when the last user
// leaves it closes itself. Explicit releases stay — they make the window go away
// in milliseconds instead of seconds, which is worth having — but nothing
// depends on them being there.

// browserIdleTimeout is how long the shared context may sit unused before it
// closes itself.
//
// Long enough that a resolve made of several back-to-back browser steps — solve
// the gate, read the server list, sniff the stream — keeps one window instead of
// relaunching between each, and short enough that a window nobody released is
// gone before the user can wonder why it is there.
const browserIdleTimeout = 15 * time.Second

// idleState is the use-count and its timer. A zero value is ready to use, which
// is what lets defaultCFSolver stay a bare &cfBrowserSolver{}.
type idleState struct {
	mu       sync.Mutex
	inFlight int
	timer    *time.Timer
	// after is swappable so a test can drive the timeout in milliseconds.
	after time.Duration
	// closeFn is swappable so a test can observe the close without a browser.
	closeFn func()
}

// busy reports whether any operation currently holds the browser.
func (s *idleState) busy() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.inFlight > 0
}

func (s *idleState) timeout() time.Duration {
	if s.after > 0 {
		return s.after
	}
	return browserIdleTimeout
}

// beginUse marks the browser busy and disarms any pending auto-close.
//
// Called by every path that reaches s.init(), so a new operation that arrives
// while the timer is counting down keeps the context it is about to use.
func (s *idleState) beginUse() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inFlight++
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
}

// endUse marks one operation finished and arms the auto-close when it was the
// last one.
func (s *idleState) endUse(closeContext func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inFlight > 0 {
		s.inFlight--
	}
	if s.inFlight > 0 {
		return
	}
	if s.timer != nil {
		s.timer.Stop()
	}
	d := s.timeout()
	s.timer = time.AfterFunc(d, func() {
		// The whole callback runs under the lock, close included.
		//
		// Checking the count, releasing the lock and THEN closing looks
		// equivalent and is not: an acquire landing in that gap raises the count
		// and gets handed the context this goroutine is already committed to
		// closing. Caught by TestIdleWatchdog_DoesNotCloseUnderARaceWithANewUser
		// under -race, which is exactly the interleaving a concurrent search
		// produces. beginUse takes the same lock, so an acquire arriving during
		// a close waits it out and then gets a fresh context — acquire claims
		// BEFORE it calls init() for that reason.
		//
		// Safe to hold: closeContext only detaches the handle and hands the
		// actual teardown to a goroutine, so this does not block on Chromium.
		s.mu.Lock()
		defer s.mu.Unlock()
		s.timer = nil
		if s.inFlight > 0 {
			return
		}
		util.Debug("SuperFlix: closing the idle browser window", "after", d)
		if s.closeFn != nil {
			s.closeFn()
			return
		}
		closeContext()
	})
}
