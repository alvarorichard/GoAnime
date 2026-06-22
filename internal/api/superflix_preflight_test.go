package api

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDescribeSuperFlixErr(t *testing.T) {
	t.Parallel()

	other := errors.New("boom")

	tests := []struct {
		name        string
		in          error
		wantNil     bool
		wantSame    bool   // returned error == input (unwrapped passthrough)
		wantWrapsPW bool   // errors.Is(out, ErrPlaywrightUnavailable)
		wantSubstr  string // message must contain this hint
	}{
		{name: "nil passes through", in: nil, wantNil: true},
		{name: "unrelated error untouched", in: other, wantSame: true},
		{
			name:        "playwright error enriched but still matchable",
			in:          fmt.Errorf("launch failed: %w", scraper.ErrPlaywrightUnavailable),
			wantWrapsPW: true,
			wantSubstr:  "installing Google Chrome",
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
			if tt.wantWrapsPW {
				assert.ErrorIs(t, out, scraper.ErrPlaywrightUnavailable,
					"enriched error must still satisfy errors.Is for the root cause")
			}
			if tt.wantSubstr != "" {
				assert.Contains(t, out.Error(), tt.wantSubstr)
			}
		})
	}
}

// withPreflightMocks swaps the preflight indirection points for recorders and
// restores them on cleanup. Returns pointers to the captured warn/info messages
// (nil pointer slice element means "not called").
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

			assert.Len(t, *warns, tt.wantWarns)
			assert.Len(t, *infos, tt.wantInfos)
			if tt.wantWarns > 0 {
				assert.Contains(t, (*warns)[0], "--sf-headless")
			}
			if tt.wantInfos > 0 {
				assert.Contains(t, (*infos)[0], "First SuperFlix run")
			}
		})
	}
}
