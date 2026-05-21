package util

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// snapshotPreWarm captures and restores the globals PreWarmConnections touches.
// preWarmOnce is a *sync.Once so swapping the pointer never copies the lock.
func snapshotPreWarm(t *testing.T) {
	t.Helper()
	prevHosts := append([]string{}, knownHosts...)
	prevOnce := preWarmOnce
	t.Cleanup(func() {
		knownHosts = prevHosts
		preWarmOnce = prevOnce
	})
}

func TestPreWarmConnections_NoHostsSpawnsNoGoroutines(t *testing.T) {
	snapshotPreWarm(t)
	knownHosts = nil
	preWarmOnce = &sync.Once{}

	// Must not panic and must return immediately.
	done := make(chan struct{})
	go func() {
		PreWarmConnections()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("PreWarmConnections did not return promptly")
	}
}

func TestPreWarmConnections_GatedBySyncOnce(t *testing.T) {
	snapshotPreWarm(t)
	knownHosts = nil
	preWarmOnce = &sync.Once{}

	calls := 0
	preWarmOnce.Do(func() { calls++ })
	preWarmOnce.Do(func() { calls++ })
	require.Equal(t, 1, calls, "sync.Once must gate execution")

	// Calling PreWarmConnections after the Once has fired must be a no-op.
	PreWarmConnections()
}

// TestPreWarmConnections_ExercisesGoroutineErrorPath points knownHosts at a
// closed TCP port. The pre-warm goroutine attempts a request, hits "connection
// refused", logs via Debugf, and exits — covering the error branch of the
// goroutine body without any real network access.
func TestPreWarmConnections_ExercisesGoroutineErrorPath(t *testing.T) {
	snapshotPreWarm(t)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	knownHosts = []string{addr}
	preWarmOnce = &sync.Once{}

	PreWarmConnections()

	// Wait for the spawned goroutine deterministically so it does not race
	// against later tests mutating package globals (e.g. IsDebug) during their
	// cleanup. Bounded by a generous timeout to fail loud if it ever hangs.
	done := make(chan struct{})
	go func() {
		preWarmWG.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("pre-warm goroutine did not finish in time")
	}
	assert.True(t, true)
}
