package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// jargonWords must never appear in a user-facing SuperFlix message: a lay user
// has to understand it at a glance. Guarded by tests below.
var jargonWords = []string{
	"$DISPLAY", "WAYLAND", "Cloudflare", "Turnstile", "Playwright",
	"Chromium", "headless", "cf_clearance", "--sf-",
}

func assertNoJargon(t *testing.T, msg string) {
	t.Helper()
	low := strings.ToLower(msg)
	for _, w := range jargonWords {
		assert.NotContains(t, low, strings.ToLower(w),
			"user-facing message must avoid jargon %q: %q", w, msg)
	}
}

func TestFriendlyError(t *testing.T) {
	t.Parallel()

	cause := errors.New("technical jargon: Playwright Chromium boom")
	fe := &friendlyError{msg: "⚠️  Something simple happened.", cause: cause}

	// Error() shows ONLY the friendly text — the raw cause must not leak.
	assert.Equal(t, "⚠️  Something simple happened.", fe.Error())
	assert.NotContains(t, fe.Error(), "Playwright")

	// Unwrap keeps the cause reachable for errors.Is / debug tooling.
	assert.Equal(t, cause, fe.Unwrap())
	assert.ErrorIs(t, fe, cause)

	// A nil cause is safe (Unwrap returns nil, no panic).
	assert.NoError(t, (&friendlyError{msg: "x"}).Unwrap())
}

func TestIsGateTimeout(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"unrelated", errors.New("connection refused"), false},
		{"gate timeout", errors.New("CF gate not cleared within 1m30s"), true},
		{"wrapped gate timeout", fmt.Errorf("solve: %w", errors.New("gate not cleared")), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isGateTimeout(tt.err))
		})
	}
}

func TestDescribeSuperFlixErr(t *testing.T) {
	t.Parallel()

	other := errors.New("boom")

	tests := []struct {
		name       string
		in         error
		wantNil    bool
		wantSame   bool   // returned error == input (unwrapped passthrough)
		wantWraps  error  // errors.Is(out, wantWraps) must hold
		wantSubstr string // translated message must contain this hint
		translated bool   // message is user-facing -> must be jargon-free
	}{
		{name: "nil passes through", in: nil, wantNil: true},
		{name: "unrelated error untouched", in: other, wantSame: true},
		{
			name:       "playwright unavailable -> setup hint, still matchable",
			in:         fmt.Errorf("launch failed: %w", scraper.ErrPlaywrightUnavailable),
			wantWraps:  scraper.ErrPlaywrightUnavailable,
			wantSubstr: "helper browser",
			translated: true,
		},
		{
			name:       "no servers -> try later hint",
			in:         fmt.Errorf("fetch: %w", scraper.ErrSuperFlixNoServers),
			wantWraps:  scraper.ErrSuperFlixNoServers,
			wantSubstr: "No video sources",
			translated: true,
		},
		{
			name:       "context deadline -> are you human hint",
			in:         fmt.Errorf("solve: %w", context.DeadlineExceeded),
			wantWraps:  context.DeadlineExceeded,
			wantSubstr: "didn't finish in time",
			translated: true,
		},
		{
			name:       "gate timeout string -> are you human hint",
			in:         errors.New("CF gate not cleared within 1m30s (auto-solve attempted)"),
			wantSubstr: "are you human",
			translated: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := describeSuperFlixErr(tt.in)

			if tt.wantNil {
				assert.NoError(t, out)
				return
			}
			require.Error(t, out)
			if tt.wantSame {
				assert.Equal(t, tt.in, out)
			}
			if tt.wantWraps != nil {
				assert.ErrorIs(t, out, tt.wantWraps,
					"translated error must still satisfy errors.Is for the root cause")
			}
			if tt.wantSubstr != "" {
				assert.Contains(t, out.Error(), tt.wantSubstr)
			}
			if tt.translated {
				assertNoJargon(t, out.Error())
			}
		})
	}
}

// withPreflightMocks swaps the preflight indirection points for recorders and
// restores them on cleanup. Returns pointers to the captured warn/info messages.
func withPreflightMocks(t *testing.T, headless, pending bool) (warns, infos *[]string) {
	t.Helper()
	origHeadless, origPending := sfHeadlessEnvFn, sfSetupPendingFn
	origWarn, origInfo := sfWarnFn, sfInfoFn
	t.Cleanup(func() {
		sfHeadlessEnvFn, sfSetupPendingFn = origHeadless, origPending
		sfWarnFn, sfInfoFn = origWarn, origInfo
	})

	var w, i []string
	sfHeadlessEnvFn = func() bool { return headless }
	sfSetupPendingFn = func() bool { return pending }
	sfWarnFn = func(msg any, _ ...any) { w = append(w, fmt.Sprint(msg)) }
	sfInfoFn = func(msg any, _ ...any) { i = append(i, fmt.Sprint(msg)) }
	return &w, &i
}

func TestPreflightSuperFlixBrowser(t *testing.T) {
	// Not parallel: mutates package-level indirection vars.
	tests := []struct {
		name      string
		headless  bool
		pending   bool
		wantWarns int
		wantInfos int
	}{
		{name: "display present, already set up -> silent", headless: false, pending: false, wantWarns: 0, wantInfos: 0},
		{name: "headless host -> only warn", headless: true, pending: false, wantWarns: 1, wantInfos: 0},
		{name: "first run -> only info", headless: false, pending: true, wantWarns: 0, wantInfos: 1},
		{name: "headless + first run -> both", headless: true, pending: true, wantWarns: 1, wantInfos: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warns, infos := withPreflightMocks(t, tt.headless, tt.pending)

			preflightSuperFlixBrowser()

			require.Len(t, *warns, tt.wantWarns)
			require.Len(t, *infos, tt.wantInfos)
			if tt.wantWarns > 0 {
				assert.Contains(t, (*warns)[0], "no screen was found")
				assertNoJargon(t, (*warns)[0])
			}
			if tt.wantInfos > 0 {
				assert.Contains(t, (*infos)[0], "First time on SuperFlix")
				assertNoJargon(t, (*infos)[0])
			}
		})
	}
}

// TestSfBrowserSpinnerHintIsLayFriendly guards the spinner hint string itself:
// it must stay jargon-free since it is shown verbatim to every user.
func TestSfBrowserSpinnerHintIsLayFriendly(t *testing.T) {
	t.Parallel()
	assertNoJargon(t, sfBrowserSpinnerHint)
	assert.Contains(t, sfBrowserSpinnerHint, "human")
}
