// Package leakcheck turns Go 1.27's goroutineleak profile into a test
// assertion.
//
// The profile reports goroutines the garbage collector has proven can never
// make progress: blocked on a channel, sync.Mutex, sync.WaitGroup or sync.Cond
// that nothing reachable can ever signal. That is a much narrower — and far less
// flaky — definition than "a goroutine is still running", which is what
// goroutine-count based leak detectors use. A worker parked on a live ticker or
// waiting on a context that someone still holds is not reported.
//
// Detection only happens when the profile is written: Profile.Count() on its own
// returns the count from the previous run. Snapshot always writes first, so the
// numbers it returns are current.
//
// Known limitation, inherited from the runtime: a goroutine blocked on a
// primitive that is still reachable through a package-level variable cannot be
// proven dead and will not be reported. Absence of a finding is therefore not
// proof of absence of a leak — but any finding is a real one.
package leakcheck

import (
	"bytes"
	"fmt"
	"runtime/pprof"
	"testing"
	"time"
)

// Snapshot runs a leak-detecting GC cycle and returns the number of leaked
// goroutines together with the human-readable profile.
func Snapshot(tb testing.TB) (leaked int, profile string) {
	tb.Helper()
	p := pprof.Lookup("goroutineleak")
	if p == nil {
		tb.Fatal("goroutineleak profile is unavailable; Go 1.27+ is required")
	}
	var buf bytes.Buffer
	if err := p.WriteTo(&buf, 1); err != nil {
		tb.Fatalf("writing goroutineleak profile: %v", err)
	}
	return p.Count(), buf.String()
}

// Count is Snapshot without the profile text.
func Count(tb testing.TB) int {
	tb.Helper()
	n, _ := Snapshot(tb)
	return n
}

// AssertNoNewLeaks fails the test if more goroutines are leaked now than at the
// baseline.
func AssertNoNewLeaks(tb testing.TB, baseline int) {
	tb.Helper()
	n, profile := settledCount(tb)
	if n > baseline {
		tb.Fatalf("goroutine leak detected: %d leaked goroutines, baseline was %d\n%s",
			n, baseline, indent(profile))
	}
}

// settledCount samples the profile several times and returns the highest
// reading.
//
// A single sample is not enough: a goroutine spawned moments ago may not have
// parked yet, and the channel or WaitGroup it will block on may still be
// reachable from a stack that has not been collected, so an immediate snapshot
// happily reports zero. Sampling repeatedly closes that window. Taking the
// maximum rather than the last value is safe because leaked goroutines never
// recover — the count only ever grows.
func settledCount(tb testing.TB) (int, string) {
	tb.Helper()
	const (
		samples = 4
		spacing = 25 * time.Millisecond
	)
	var maxN int
	var worst string
	for i := range samples {
		if i > 0 {
			time.Sleep(spacing)
		}
		n, profile := Snapshot(tb)
		if n >= maxN {
			maxN, worst = n, profile
		}
	}
	return maxN, worst
}

// Guard takes a baseline and registers AssertNoNewLeaks as test cleanup.
func Guard(tb testing.TB) {
	tb.Helper()
	baseline := Count(tb)
	tb.Cleanup(func() { AssertNoNewLeaks(tb, baseline) })
}

func indent(s string) string {
	const limit = 4000
	if len(s) > limit {
		s = s[:limit] + "\n... (profile truncated)"
	}
	return fmt.Sprintf("--- goroutineleak profile ---\n%s", s)
}
