package downloader

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
)

// ---- monotonicPercent ----

type pctFn func(received, total int64, peak float64) (float64, float64)

var pctCases = []struct {
	name            string
	received, total int64
	peak            float64
	wantPct, wantPk float64
}{
	{"nothing received", 0, 100, 0, 0, 0},
	{"unknown total", 50, 0, 0, 0, 0},
	{"negative total", 50, -5, 0, 0, 0},
	{"half way", 50, 100, 0, 0.5, 0.5},
	{"progress raises the peak", 75, 100, 0.5, 0.75, 0.75},
	{"a shrinking ratio keeps the bar where it was", 40, 100, 0.75, 0.75, 0.75},
	{"a larger estimate does not pull the bar back", 50, 400, 0.5, 0.5, 0.5},
	{"capped just below full", 100, 100, 0.5, 0.99, 0.99},
	{"over-delivery is capped too", 150, 100, 0, 0.99, 0.99},
	{"already at the cap", 100, 100, 0.99, 0.99, 0.99},
	{"nothing received keeps the peak", 0, 100, 0.6, 0.6, 0.6},
	{"unknown total keeps the peak", 10, 0, 0.6, 0.6, 0.6},
	{"exactly the peak", 60, 100, 0.6, 0.6, 0.6},
	{"a one byte total counts", 1, 1, 0, 0.99, 0.99},
	{"one byte received counts", 1, 100, 0, 0.01, 0.01},
}

func pctFailures(fn pctFn) []string {
	var bad []string
	for _, c := range pctCases {
		pct, pk := fn(c.received, c.total, c.peak)
		if math.Abs(pct-c.wantPct) > 1e-9 || math.Abs(pk-c.wantPk) > 1e-9 {
			bad = append(bad, fmt.Sprintf("%s: got (%v,%v), want (%v,%v)", c.name, pct, pk, c.wantPct, c.wantPk))
		}
	}
	return bad
}

func TestMonotonicPercent(t *testing.T) {
	if bad := pctFailures(monotonicPercent); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

func TestMonotonicPercent_NeverGoesBackwardNeverReachesOne(t *testing.T) {
	peak := 0.0
	for _, r := range []int64{0, 10, 50, 30, 80, 20, 100, 120, 5} {
		pct, np := monotonicPercent(r, 100, peak)
		if pct < peak || np < peak || pct >= 1 || np >= 1 {
			t.Fatalf("received %d: pct %v peak %v -> %v/%v", r, pct, peak, pct, np)
		}
		peak = np
	}
}

func TestMeta_MonotonicPercent_CatchesBrokenImplementations(t *testing.T) {
	mutants := map[string]pctFn{
		"no peak (can go backward)": func(r, tot int64, pk float64) (float64, float64) {
			p, _ := monotonicPercent(r, tot, 0)
			return p, p
		},
		"no cap": func(r, tot int64, pk float64) (float64, float64) {
			if tot > 0 && r > 0 {
				p := float64(r) / float64(tot)
				if p > pk {
					return p, p
				}
			}
			return monotonicPercent(r, tot, pk)
		},
		"cap at one": func(r, tot int64, pk float64) (float64, float64) {
			p, np := monotonicPercent(r, tot, pk)
			if r >= tot && tot > 0 {
				return 1, 1
			}
			return p, np
		},
		"peak not updated": func(r, tot int64, pk float64) (float64, float64) {
			p, _ := monotonicPercent(r, tot, pk)
			return p, pk
		},
		"drops the peak when idle": func(r, tot int64, pk float64) (float64, float64) {
			p, np := monotonicPercent(r, tot, pk)
			if p == pk {
				return p, 0
			}
			return p, np
		},
		"divides by received": func(r, tot int64, pk float64) (float64, float64) {
			return monotonicPercent(tot, r, pk)
		},
		"ignores unknown total": func(r, tot int64, pk float64) (float64, float64) {
			return monotonicPercent(r, max(tot, 1), pk)
		},
	}
	for name, fn := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(pctFailures(fn)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}

// ---- Update ----

func newModel(total, received int64) *progressModel {
	m := newProgressModel("Show", "3 episodes")
	m.totalBytes, m.received = total, received
	return m
}

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

// cmdKind names what a command produces, running it (a tick waits 100ms).
func cmdKind(cmd tea.Cmd) string {
	if cmd == nil {
		return "nil"
	}
	switch cmd().(type) {
	case tea.BatchMsg:
		return "bar+tick"
	case tickMsg:
		return "tick"
	case tea.QuitMsg:
		return "quit"
	default:
		return "other"
	}
}

type updateFn func(m *progressModel, msg tea.Msg) (tea.Model, tea.Cmd)

func realUpdate(m *progressModel, msg tea.Msg) (tea.Model, tea.Cmd) { return m.Update(msg) }

// updateFailures drives one model through a scripted conversation and
// reports each deviation from the screen's contract.
func updateFailures(fn updateFn) []string {
	var bad []string
	fail := func(f string, a ...any) { bad = append(bad, fmt.Sprintf(f, a...)) }
	now := time.Now()

	// ctrl+c quits and marks the screen done.
	m := newModel(100, 0)
	if _, cmd := fn(m, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}); !isQuit(cmd) || !m.done {
		fail("ctrl+c: quit=%v done=%v", isQuit(cmd), m.done)
	}
	// Any other key does nothing.
	m = newModel(100, 0)
	if _, cmd := fn(m, tea.KeyPressMsg{Code: 'x'}); cmd != nil || m.done {
		fail("other key: cmd=%v done=%v", cmd != nil, m.done)
	}

	// A resize sets the width and resizes the bar.
	m = newModel(100, 0)
	if _, cmd := fn(m, tea.WindowSizeMsg{Width: 30, Height: 24}); cmd != nil || m.width != 30 {
		fail("resize: cmd=%v width=%d", cmd != nil, m.width)
	}
	if got, want := m.progress.Width(), tui.TransferBarWidth(30); got != want {
		fail("resize to 30: the bar is %d wide, want %d", got, want)
	}
	fn(m, tea.WindowSizeMsg{Width: 200, Height: 24})
	if got, want := m.progress.Width(), tui.TransferBarWidth(200); m.width != 200 || got != want {
		fail("resize to 200: width=%d bar=%d, want bar %d", m.width, got, want)
	}

	// A tick on a finished screen quits instead of rescheduling.
	m = newModel(100, 50)
	m.done = true
	if _, cmd := fn(m, tickMsg(now)); !isQuit(cmd) {
		fail("tick when done must quit")
	}
	// A tick with progress moves the bar and keeps ticking.
	m = newModel(100, 50)
	if _, cmd := fn(m, tickMsg(now)); cmdKind(cmd) != "bar+tick" || m.peakPct != 0.5 {
		fail("tick with progress: cmd=%s peak=%v, want bar+tick 0.5", cmdKind(cmd), m.peakPct)
	}
	// A tick before anything arrived is a plain tick and the bar stays empty.
	m = newModel(100, 0)
	if _, cmd := fn(m, tickMsg(now)); cmdKind(cmd) != "tick" || m.peakPct != 0 {
		fail("tick before any progress: cmd=%s peak=%v, want a plain tick and 0", cmdKind(cmd), m.peakPct)
	}
	// A tick with nothing new keeps the bar where it got to.
	m = newModel(100, 0)
	m.peakPct = 0.3
	if _, cmd := fn(m, tickMsg(now)); cmdKind(cmd) != "bar+tick" || m.peakPct != 0.3 {
		fail("tick with nothing new: cmd=%s peak=%v, want bar+tick and 0.3", cmdKind(cmd), m.peakPct)
	}
	// A tick never reaches 100%.
	m = newModel(100, 100)
	fn(m, tickMsg(now))
	if m.peakPct != 0.99 {
		fail("tick at full: peak=%v, want 0.99", m.peakPct)
	}

	// A status message is stored.
	m = newModel(100, 0)
	if _, cmd := fn(m, statusMsg("Downloading episode 2...")); cmd != nil || m.status != "Downloading episode 2..." {
		fail("status: cmd=%v status=%q", cmd != nil, m.status)
	}

	// A progress message stores the byte counts and moves the bar forward only.
	m = newModel(0, 0)
	if _, cmd := fn(m, progressMsg{received: 50, totalBytes: 100}); cmd == nil || m.received != 50 || m.totalBytes != 100 || m.peakPct != 0.5 {
		fail("progress 50/100: cmd=%v recv=%d total=%d peak=%v", cmd != nil, m.received, m.totalBytes, m.peakPct)
	}
	fn(m, progressMsg{received: 20, totalBytes: 100})
	if m.received != 20 || m.peakPct != 0.5 {
		fail("progress going back: recv=%d peak=%v (bytes follow the message, the bar does not)", m.received, m.peakPct)
	}
	fn(m, progressMsg{received: 100, totalBytes: 100})
	if m.peakPct != 0.99 {
		fail("progress at full: peak=%v, want 0.99", m.peakPct)
	}
	m = newModel(0, 0)
	if _, cmd := fn(m, progressMsg{received: 10, totalBytes: 0}); cmd != nil || m.peakPct != 0 {
		fail("progress with unknown total: cmd=%v peak=%v", cmd != nil, m.peakPct)
	}

	// A bar animation frame is handled without touching the download state.
	m = newModel(100, 40)
	if mod, _ := fn(m, progress.FrameMsg{}); mod != tea.Model(m) || m.received != 40 || m.done {
		fail("frame: model changed identity or state (recv=%d done=%v)", m.received, m.done)
	}

	// Unknown messages change nothing.
	m = newModel(100, 40)
	if _, cmd := fn(m, struct{}{}); cmd != nil || m.received != 40 || m.done {
		fail("unknown message: cmd=%v recv=%d done=%v", cmd != nil, m.received, m.done)
	}
	return bad
}

func TestProgressModelUpdate(t *testing.T) {
	if bad := updateFailures(realUpdate); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

func TestMeta_ProgressModelUpdate_CatchesBrokenImplementations(t *testing.T) {
	mutants := map[string]updateFn{
		"ctrl+c does not quit": func(m *progressModel, msg tea.Msg) (tea.Model, tea.Cmd) {
			if k, ok := msg.(tea.KeyPressMsg); ok && k.String() == "ctrl+c" {
				return m, nil
			}
			return m.Update(msg)
		},
		"every key quits": func(m *progressModel, msg tea.Msg) (tea.Model, tea.Cmd) {
			if _, ok := msg.(tea.KeyPressMsg); ok {
				m.done = true
				return m, tea.Quit
			}
			return m.Update(msg)
		},
		"resize ignored": func(m *progressModel, msg tea.Msg) (tea.Model, tea.Cmd) {
			if _, ok := msg.(tea.WindowSizeMsg); ok {
				return m, nil
			}
			return m.Update(msg)
		},
		"tick when done keeps going": func(m *progressModel, msg tea.Msg) (tea.Model, tea.Cmd) {
			if _, ok := msg.(tickMsg); ok && m.done {
				return m, tickCmd()
			}
			return m.Update(msg)
		},
		"tick stops ticking": func(m *progressModel, msg tea.Msg) (tea.Model, tea.Cmd) {
			if _, ok := msg.(tickMsg); ok && !m.done {
				m.Update(msg)
				return m, nil
			}
			return m.Update(msg)
		},
		"status ignored": func(m *progressModel, msg tea.Msg) (tea.Model, tea.Cmd) {
			if _, ok := msg.(statusMsg); ok {
				return m, nil
			}
			return m.Update(msg)
		},
		"progress bytes ignored": func(m *progressModel, msg tea.Msg) (tea.Model, tea.Cmd) {
			if _, ok := msg.(progressMsg); ok {
				return m, nil
			}
			return m.Update(msg)
		},
		"progress can go backward": func(m *progressModel, msg tea.Msg) (tea.Model, tea.Cmd) {
			if pm, ok := msg.(progressMsg); ok {
				m.Update(msg)
				m.peakPct = float64(pm.received) / float64(max(pm.totalBytes, 1))
				return m, nil
			}
			return m.Update(msg)
		},
		"progress forgets the total": func(m *progressModel, msg tea.Msg) (tea.Model, tea.Cmd) {
			if pm, ok := msg.(progressMsg); ok {
				pm.totalBytes = 0
				return m.Update(pm)
			}
			return m.Update(msg)
		},
	}
	for name, fn := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(updateFailures(fn)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}

func TestProgressModel_ViewByState(t *testing.T) {
	render := func(setup func(m *progressModel)) string {
		m := newModel(1000, 400)
		m.width = 100
		setup(m)
		return m.View().Content
	}
	running := render(func(m *progressModel) { m.peakPct = 0.4; m.status = "Downloading episode 2..." })
	done := render(func(m *progressModel) { m.done = true })
	failed := render(func(m *progressModel) { m.done, m.failed = true, true })
	if !strings.Contains(running, "Downloading episode 2...") {
		t.Error("the running view must show the status")
	}
	if done == running || failed == running || done == failed {
		t.Error("running, done and failed must render differently")
	}
}
