package superflix

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The browser window is the part of this program the user actually sees, and it
// has been left on screen twice by the same mechanism: "whoever opens it is
// supposed to close it". fetchSuperFlixSeasons opened one for a season list and
// released it only when a stream was eventually resolved, so it sat through
// every picker in between and never closed at all if the user backed out.
// SniffStream opened a tab per call and closed none.
//
// These tests pin the replacement: the context counts its users and closes
// itself when the last one leaves. They run without a browser, because the
// mechanism is the thing being tested, not Chromium.

// waitClosed reports whether the watchdog fired within d, and how many
// operations held the browser AT THE MOMENT it closed.
//
// That second value is the whole point. Reading the count after the fact cannot
// tell "closed, then a new user arrived" (correct) from "closed while a user
// held it" (the bug) — the first version of this file got that wrong and passed
// a solver that really could close a context out from under a live solve.
func waitClosed(t *testing.T, closed <-chan int, d time.Duration) (bool, int) {
	t.Helper()
	select {
	case n := <-closed:
		return true, n
	case <-time.After(d):
		return false, 0
	}
}

// newProbe returns a watchdog whose close records the in-flight count observed
// under the same lock the close is made under.
func newProbe() (*idleState, <-chan int) {
	closed := make(chan int, 64)
	st := &idleState{after: 30 * time.Millisecond}
	// Called with st.mu already held, so this reads the field directly rather
	// than through busy(), which would deadlock.
	st.closeFn = func() { closed <- st.inFlight }
	return st, closed
}

func TestIdleWatchdog_ClosesAfterTheLastUserLeaves(t *testing.T) {
	t.Parallel()
	st, closed := newProbe()

	st.beginUse()
	st.endUse(func() { t.Error("closeFn must take precedence in the probe") })

	fired, held := waitClosed(t, closed, time.Second)
	assert.True(t, fired, "a window nobody is using any more must close itself, with no explicit release")
	assert.Zero(t, held)
}

// The reason the explicit close had to go: a search runs its sources
// concurrently. Closing the shared context at the end of ONE solve tears down
// the browser another source is in the middle of using, which then relaunches
// onto the same persistent profile while the old instance still holds its lock.
func TestIdleWatchdog_KeepsTheWindowWhileAnyoneStillNeedsIt(t *testing.T) {
	t.Parallel()
	st, closed := newProbe()

	st.beginUse()
	st.beginUse()
	st.endUse(func() {})

	fired, _ := waitClosed(t, closed, 150*time.Millisecond)
	assert.False(t, fired, "one source finishing must not close the window another source is still using")

	st.endUse(func() {})
	fired, held := waitClosed(t, closed, time.Second)
	assert.True(t, fired, "once the last user leaves, it must close")
	assert.Zero(t, held)
}

// A resolve is several browser steps back to back — clear the gate, read the
// server list, sniff the stream. Each one releases, and the next one arrives
// while the timer is counting down. Re-arming on every acquire is what keeps
// that sequence on ONE window instead of relaunching between each step.
func TestIdleWatchdog_ANewUserDisarmsAPendingClose(t *testing.T) {
	t.Parallel()
	st, closed := newProbe()

	st.beginUse()
	st.endUse(func() {})
	time.Sleep(10 * time.Millisecond) // timer armed, not yet fired
	st.beginUse()

	fired, _ := waitClosed(t, closed, 150*time.Millisecond)
	assert.False(t, fired, "an operation that arrives before the timer fires must keep the context it is about to use")

	st.endUse(func() {})
	fired, held := waitClosed(t, closed, time.Second)
	assert.True(t, fired)
	assert.Zero(t, held)
}

// The timer callback and a new acquire can land at the same moment. Closing the
// context out from under a solve that just started would kill it.
func TestIdleWatchdog_DoesNotCloseUnderARaceWithANewUser(t *testing.T) {
	t.Parallel()
	for range 200 {
		st, closed := newProbe()
		st.beginUse()
		st.endUse(func() {})

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(30 * time.Millisecond) // aim at the firing instant
			st.beginUse()
		}()
		wg.Wait()

		// Either ordering is legitimate. The invariant is only about what was
		// true AT the close: it must never have happened with a user inside.
		if fired, held := waitClosed(t, closed, 120*time.Millisecond); fired {
			require.Zerof(t, held,
				"the context was closed while %d operation(s) held it — a live solve would die with it", held)
		}
		st.endUse(func() {})
	}
}

// An unbalanced release must not drive the count negative, which would make the
// next endUse see a positive count and never arm the timer — a leak that only
// shows up on the run AFTER the buggy one.
func TestIdleWatchdog_SurvivesAnUnbalancedRelease(t *testing.T) {
	t.Parallel()
	st, closed := newProbe()

	st.endUse(func() {}) // release with nothing acquired
	<-closed             // it arms; that is fine

	st.beginUse()
	st.endUse(func() {})
	fired, held := waitClosed(t, closed, time.Second)
	assert.True(t, fired, "the count must not go negative; a later real release still has to close the window")
	assert.Zero(t, held)
}

// ── The structural guarantee ────────────────────────────────────────────────

// Every path that starts the browser must go through acquire(), because
// acquire() is what registers the user that the watchdog counts. A new entry
// point that calls s.init() directly gets a window nothing will ever close —
// which is precisely how both leaks were written, each time by someone
// following the shape of the function next to it.
//
// A scan is the only check that catches that: no unit test fails when a NEW
// function forgets, because the test for it does not exist yet.
func TestEveryBrowserEntryPointGoesThroughAcquire(t *testing.T) {
	t.Parallel()

	// acquire() is the one place allowed to call init(), plus init's own
	// declaration.
	allowed := regexp.MustCompile(`func \(s \*cfBrowserSolver\) (acquire|init)\(`)
	initCall := regexp.MustCompile(`\bs\.init\(\)`)

	dir := "."
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	var offenders []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		body, readErr := os.ReadFile(filepath.Join(dir, name)) // #nosec G304 -- the package's own sources, in a test
		require.NoError(t, readErr)

		var fn string
		for i, line := range strings.Split(string(body), "\n") {
			if strings.HasPrefix(line, "func ") {
				fn = line
			}
			if !initCall.MatchString(line) || strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			if allowed.MatchString(fn) {
				continue
			}
			offenders = append(offenders, name+":"+itoa(i+1)+" in "+firstWords(fn))
		}
	}

	assert.Emptyf(t, offenders,
		"these start the browser without registering a user, so the idle watchdog can never "+
			"close the window they open — call s.acquire() and defer its release instead: %v", offenders)
}

// Every browser operation must serialize through browserWorkGate as well as
// acquire the shared context. This structural scan catches a future entrypoint
// that opens the browser safely but forgets the cancelable serialization gate.
func TestEveryBrowserEntryPointUsesCancelableGate(t *testing.T) {
	t.Parallel()

	want := map[string]bool{
		"Solve":            false,
		"solveGate":        false,
		"SniffStream":      false,
		"SniffEmbedStream": false,
	}
	entries, err := os.ReadDir(".")
	require.NoError(t, err)

	var offenders []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		file, parseErr := parser.ParseFile(fset, name, nil, 0)
		require.NoError(t, parseErr)

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Body == nil || receiverName(fn) != "cfBrowserSolver" {
				continue
			}

			method := fn.Name.Name
			_, expected := want[method]
			usesAcquire, locksGate, defersUnlock := false, false, false
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.CallExpr:
					switch selectorPath(n.Fun) {
					case "s.acquire":
						usesAcquire = true
					case "s.browserWorkGate.lock":
						locksGate = locksGate || (len(n.Args) == 1 && selectorPath(n.Args[0]) == "ctx")
					}
				case *ast.DeferStmt:
					if selectorPath(n.Call.Fun) == "s.browserWorkGate.unlock" {
						defersUnlock = true
					}
				}
				return true
			})

			if expected {
				want[method] = true
				if !usesAcquire {
					offenders = append(offenders, method+" no longer calls acquire()")
				}
			}
			if usesAcquire && method != "acquire" {
				if !locksGate {
					offenders = append(offenders, method+" does not take browserWorkGate with ctx")
				}
				if !defersUnlock {
					offenders = append(offenders, method+" does not defer browserWorkGate.unlock()")
				}
			}
		}
	}

	for method, found := range want {
		if !found {
			offenders = append(offenders, "missing browser entrypoint "+method)
		}
	}
	assert.Emptyf(t, offenders, "browser entrypoints must acquire and release the shared cancelable gate: %v", offenders)
}

func receiverName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	typ := fn.Recv.List[0].Type
	if star, ok := typ.(*ast.StarExpr); ok {
		typ = star.X
	}
	if ident, ok := typ.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func selectorPath(expr ast.Expr) string {
	switch node := expr.(type) {
	case *ast.Ident:
		return node.Name
	case *ast.SelectorExpr:
		return selectorPath(node.X) + "." + node.Sel.Name
	default:
		return ""
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func firstWords(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i > 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
