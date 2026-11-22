package tui

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A terminal that never answers OSC 11 (tmux, the Linux console, many SSH
// sessions) once left the probe's read running for the life of the process.
// It then took the next keys typed, up to 64 bytes, from whichever menu came
// next, and the episode picker of the first title looked frozen to the user
// (found 2026-11-21 by driving goanime in tmux). These tests hold the probe
// to one rule: when readOSCReply returns, nothing it started is reading.

// fakeTerminal is stdin as the probe sees it: bytes the terminal sends, and a
// read that blocks, as a terminal's does, until there is one or the read is
// cancelled.
type fakeTerminal struct {
	mu        sync.Mutex
	pending   []byte
	ready     chan struct{} // closed and replaced whenever bytes arrive or Cancel is called
	cancelled bool
	reading   atomic.Int32 // reads under way
	readsLate atomic.Int32 // reads that started after Cancel
	closed    atomic.Bool
}

func newFakeTerminal(sent string) *fakeTerminal {
	return &fakeTerminal{pending: []byte(sent), ready: make(chan struct{})}
}

var errCancelled = errors.New("read cancelled")

func (f *fakeTerminal) Read(p []byte) (int, error) {
	f.reading.Add(1)
	defer f.reading.Add(-1)
	for {
		f.mu.Lock()
		if f.cancelled {
			f.readsLate.Add(1)
			f.mu.Unlock()
			// A cancelled read returns soon, not at once: a probe that does
			// not wait for it returns with the read still under way.
			time.Sleep(30 * time.Millisecond)
			return 0, errCancelled
		}
		if len(f.pending) > 0 {
			n := copy(p, f.pending)
			f.pending = f.pending[n:]
			f.mu.Unlock()
			return n, nil
		}
		ready := f.ready
		f.mu.Unlock()
		<-ready
	}
}

// send is the terminal (or the user) writing bytes later.
func (f *fakeTerminal) send(b string) {
	f.mu.Lock()
	f.pending = append(f.pending, b...)
	close(f.ready)
	f.ready = make(chan struct{})
	f.mu.Unlock()
}

func (f *fakeTerminal) Cancel() bool {
	f.mu.Lock()
	f.cancelled = true
	close(f.ready)
	f.ready = make(chan struct{})
	f.mu.Unlock()
	return true
}

func (f *fakeTerminal) Close() error { f.closed.Store(true); return nil }

func (f *fakeTerminal) left() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return string(f.pending)
}

func TestOSC11ReplyIsRead(t *testing.T) {
	for name, sent := range map[string]string{
		"ended by BEL": "\x1b]11;rgb:ffff/ffff/ffff\a",
		"ended by ST":  "\x1b]11;rgb:0000/0000/0000\x1b\\",
	} {
		t.Run(name, func(t *testing.T) {
			term := newFakeTerminal(sent)
			reply, ok := readOSCReply(term, time.Second)
			require.True(t, ok)
			assert.Contains(t, reply, "11;rgb:")
			assert.NotContains(t, reply, "\x1b\\")
			assert.Zero(t, term.reading.Load(), "nothing is left reading")
		})
	}
}

func TestOSC11SilentTerminalLeavesNothingReading(t *testing.T) {
	term := newFakeTerminal("")
	start := time.Now()
	_, ok := readOSCReply(term, 50*time.Millisecond)
	assert.False(t, ok, "a terminal that says nothing has no background")
	assert.Less(t, time.Since(start), osc11Drain, "the probe does not hang on a silent terminal")
	assert.Zero(t, term.reading.Load(), "the read it started was called off before it returned")

	// The user now types in the next menu: every key reaches that menu.
	term.mu.Lock()
	term.cancelled = false // the next screen's reader, not the probe's
	term.mu.Unlock()
	term.send("jq\r")
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, "jq\r", term.left(), "no key was taken by the probe")
}

func TestOSC11LateReplyAfterTheTimeoutIsNotReadByTheProbe(t *testing.T) {
	term := newFakeTerminal("")
	_, ok := readOSCReply(term, 30*time.Millisecond)
	require.False(t, ok)
	term.mu.Lock()
	term.cancelled = false
	term.mu.Unlock()
	term.send("\x1b]11;rgb:ffff/ffff/ffff\a")
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, "\x1b]11;rgb:ffff/ffff/ffff\a", term.left(), "left for the drain, not eaten by a reader nobody waits for")
}

func TestOSC11ProbeDoesNotReadPastTheReply(t *testing.T) {
	term := newFakeTerminal("\x1b]11;rgb:ffff/ffff/ffff\ajk")
	_, ok := readOSCReply(term, time.Second)
	require.True(t, ok)
	assert.Equal(t, "jk", term.left(), "keys typed after the reply are the next screen's")
}
