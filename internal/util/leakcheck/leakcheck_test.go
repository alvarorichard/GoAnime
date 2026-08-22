package leakcheck_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util/leakcheck"
)

// TestCheckerDetectsARealLeak is the self-test that keeps the CI gate honest: a
// leak detector that can never fire is worse than no detector at all.
func TestCheckerDetectsARealLeak(t *testing.T) {
	baseline := leakcheck.Count(t)

	// A channel that becomes unreachable while a goroutine waits on it can never
	// be sent to, so the collector can prove this goroutine is stuck. (The
	// runtime does not currently prove the equivalent case for sync.Mutex, which
	// is why this uses a channel.)
	func() {
		ch := make(chan int)
		go func() { <-ch }() // deliberate leak: nothing can ever send on ch
	}()

	fake := &recordingTB{TB: t}
	leakcheck.AssertNoNewLeaks(fake, baseline)

	if !fake.failed {
		t.Fatal("AssertNoNewLeaks did not report a deliberately leaked goroutine")
	}
	if !strings.Contains(fake.msg, "goroutine leak detected") {
		t.Errorf("unexpected failure message: %q", fake.msg)
	}
	if !strings.Contains(fake.msg, "goroutineleak profile") {
		t.Error("the failure should include the profile so the leak can be located")
	}
}

// TestCleanShutdownIsNotFlaggedAsALeak is the false-positive guard: goroutines
// that finish, and goroutines still working on live channels, must not trip it.
func TestCleanShutdownIsNotFlaggedAsALeak(t *testing.T) {
	baseline := leakcheck.Count(t)

	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			ch := make(chan int, 1)
			ch <- 1
			<-ch
		})
	}
	wg.Wait()

	// A goroutine that is still running against a channel the test holds is not
	// a leak: it can still make progress.
	live := make(chan struct{})
	done := make(chan struct{})
	go func() {
		<-live
		close(done)
	}()

	leakcheck.AssertNoNewLeaks(t, baseline)

	close(live)
	<-done
}

// recordingTB captures a Fatalf instead of aborting the test.
type recordingTB struct {
	testing.TB
	failed bool
	msg    string
}

func (r *recordingTB) Fatalf(format string, args ...any) {
	r.failed = true
	r.msg = sprintf(format, args...)
}

func (r *recordingTB) Helper() {}

func sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
