package providers

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"testing"

	"charm.land/log/v2"
	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// What a user saw when every source declined, before this type existed:
//
//	Search failed for "dexter": failed to search: no results for "dexter" (all
//	sources failed) — SuperFlix blocked the request: HTTP 429/challenge; HiAnime
//	temporarily unavailable: HTTP 503: SuperFlix: server returned: 429 Too Many
//	Requests
//	HiAnime: HiAnime search: upstream unavailable with HTTP 503
//
// One fact, stated three times, in a paragraph. The failure is data now, so the
// terminal gets one short line per source and the log keeps the causes.

func TestSearchFailure_ErrorIsOneShortLine(t *testing.T) {
	t.Parallel()
	f := twoSourceFailure()

	msg := f.Error()

	assert.Contains(t, msg, `"dexter"`)
	assert.Contains(t, msg, "SuperFlix")
	assert.Contains(t, msg, "HiAnime")
	assert.NotContains(t, msg, "\n", "the one-line form must stay one line")
	assert.NotContains(t, msg, "server returned",
		"raw causes belong in Detail() and the chain, not in what is printed")
	assert.Less(t, len(msg), 200, "a terminal line nobody reads is the same as no message")
}

// Reading the log must still reach the real errors.
func TestSearchFailure_DetailKeepsTheCauses(t *testing.T) {
	t.Parallel()
	f := twoSourceFailure()

	detail := f.Detail()

	assert.Contains(t, detail, "429", "the status the host actually returned")
	assert.Contains(t, detail, "503")
	assert.Contains(t, detail, "cause:")
	assert.Greater(t, strings.Count(detail, "\n"), 0, "one line per source")
}

// errors.Is/As must reach through the failure, or every caller that classifies
// an error (the retry loop, the circuit breaker, the tests) stops working.
func TestSearchFailure_UnwrapsToEveryCause(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("upstream is on fire")
	f := &SearchFailure{Query: "dexter", Sources: []SourceFailure{
		{Kind: source.SuperFlix, Reason: "failed", Err: errors.New("something")},
		{Kind: source.HiAnime, Reason: "failed", Err: sentinel},
	}}

	assert.ErrorIs(t, f, sentinel, "a cause from any source must be reachable")
}

func TestSearchFailure_RateLimitedFlagsTheActionableCase(t *testing.T) {
	t.Parallel()

	assert.True(t, twoSourceFailure().RateLimited(),
		"one throttled source is enough: retrying immediately cannot help")

	none := &SearchFailure{Query: "x", Sources: []SourceFailure{
		{Kind: source.HiAnime, Reason: "is temporarily unavailable (HTTP 503)"},
	}}
	assert.False(t, none.RateLimited(), "a source being down is not us being throttled")
}

// A nil failure must not panic on the error path.
func TestSearchFailure_NilIsInert(t *testing.T) {
	t.Parallel()
	var f *SearchFailure

	assert.NotPanics(t, func() {
		_ = f.Error()
		_ = f.Detail()
		_ = f.RateLimited()
		_ = f.Unwrap()
	})
}

// describeFailure is what turns a diagnostic into the line a person reads. It
// must never leak the raw error, and must flag the throttled case.
func TestDescribeFailure_PhrasesEachClassForAPerson(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name        string
		diag        *netx.SourceDiagnostic
		err         error
		wantSaid    string
		wantLimited bool
	}{
		{
			name:        "throttled by status",
			diag:        &netx.SourceDiagnostic{Kind: netx.DiagnosticBlockedChallenge, StatusCode: http.StatusTooManyRequests},
			wantSaid:    "rate limiting",
			wantLimited: true,
		},
		{
			name:        "suppressed by our own back-off",
			diag:        &netx.SourceDiagnostic{Kind: netx.DiagnosticUnknown},
			err:         errors.New("failed to make request: superflix: rate limited, backing off: host asked us to wait"),
			wantSaid:    "rate limiting",
			wantLimited: true,
		},
		{
			name:     "origin down",
			diag:     &netx.SourceDiagnostic{Kind: netx.DiagnosticSourceUnavailable, StatusCode: http.StatusServiceUnavailable},
			wantSaid: "temporarily unavailable (HTTP 503)",
		},
		{
			name:     "captcha",
			diag:     &netx.SourceDiagnostic{Kind: netx.DiagnosticBlockedChallenge},
			wantSaid: "captcha",
		},
		{
			name:     "parser drifted",
			diag:     &netx.SourceDiagnostic{Kind: netx.DiagnosticParserBroken},
			wantSaid: "could not read it",
		},
		{
			name:     "no diagnostic at all",
			wantSaid: "failed",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reason, limited := describeFailure(tt.diag, tt.err)

			assert.Contains(t, reason, tt.wantSaid)
			assert.Equal(t, tt.wantLimited, limited)
			assert.NotContains(t, reason, "HTTP 429",
				"the throttled line tells the user to wait; the code belongs in the log")
		})
	}
}

// The reason is a clause, not a sentence about a source: the caller prints the
// source name itself, and repeating it reads like a stutter.
func TestDescribeFailure_ReasonDoesNotRepeatTheSourceName(t *testing.T) {
	t.Parallel()
	reason, _ := describeFailure(&netx.SourceDiagnostic{
		Source: "SuperFlix", Kind: netx.DiagnosticSourceUnavailable, StatusCode: 503,
	}, nil)

	assert.NotContains(t, reason, "SuperFlix")
}

func twoSourceFailure() *SearchFailure {
	return &SearchFailure{Query: "dexter", Sources: []SourceFailure{
		{
			Kind:        source.SuperFlix,
			Reason:      "is rate limiting this network — wait a few minutes",
			RateLimited: true,
			Err:         errors.New("SuperFlix: server returned: 429 Too Many Requests"),
		},
		{
			Kind:   source.HiAnime,
			Reason: "is temporarily unavailable (HTTP 503)",
			Err:    errors.New("HiAnime: HiAnime search: upstream unavailable with HTTP 503"),
		},
	}}
}

// finishSearch is the one place that decides between "nothing matched" and
// "nothing answered". Conflating them is what produced "No anime found" for a
// host that was refusing us.
func TestFinishSearch_SeparatesEmptyFromFailed(t *testing.T) {
	t.Parallel()

	_, err := finishSearch("dexter", 2, nil, nil)
	require.Error(t, err)
	var asFailure *SearchFailure
	assert.False(t, errors.As(err, &asFailure),
		"every source answered with nothing; that is not a failure to report per source")
	assert.Contains(t, err.Error(), "no results found for")

	_, err = finishSearch("dexter", 2, nil, twoSourceFailure().Sources)
	require.ErrorAs(t, err, &asFailure)
	assert.Equal(t, "dexter", asFailure.Query)

	results, err := finishSearch("dexter", 3, []*models.Anime{{Name: "Dexter"}}, twoSourceFailure().Sources)
	require.NoError(t, err, "one source answering is a successful search, whatever the others did")
	assert.Len(t, results, 1)
}

// A search that succeeds while most of its sources are broken must say so.
//
// It used to hand back the results and drop the reasons, so a run where three
// of four sources failed — a rewritten site, a rate-limited network, a host
// returning 503 — was indistinguishable from a run that only ever had one
// source. That is what "it is only searching Goyabu" described: it was
// searching all four, and silently losing three.
func TestFinishSearch_PartialFailureIsReportedNotSwallowed(t *testing.T) {
	// Captures package-level logging — not parallel.
	restore := captureWarnings()
	defer restore()

	results, err := finishSearch("naruto", 3,
		[]*models.Anime{{Name: "Naruto", Source: "Goyabu"}},
		twoSourceFailure().Sources)

	require.NoError(t, err, "one source answering is still a successful search")
	assert.Len(t, results, 1)

	logged := warnings()
	assert.Contains(t, logged, "SuperFlix", "the user must learn which sources were missing")
	assert.Contains(t, logged, "HiAnime")
	assert.Contains(t, logged, "incomplete", "and that the result set is smaller than usual")
}

// The quiet path has to stay quiet: a clean search must not warn about nothing.
func TestFinishSearch_NoWarningWhenEverySourceAnswered(t *testing.T) {
	restore := captureWarnings()
	defer restore()

	_, err := finishSearch("naruto", 1, []*models.Anime{{Name: "Naruto"}}, nil)

	require.NoError(t, err)
	assert.Empty(t, warnings(), "nothing failed; there is nothing to report")
}

// captureWarnings redirects the shared logger into a buffer for one test.
//
// util.Logger is package-level, so a test that swaps it CANNOT run in parallel —
// a sibling would read the other's output. Every test using this is serial.
var warnBuf *bytes.Buffer

func captureWarnings() func() {
	prev := util.Logger
	warnBuf = &bytes.Buffer{}
	util.Logger = log.NewWithOptions(warnBuf, log.Options{Level: log.WarnLevel})
	return func() { util.Logger = prev }
}

func warnings() string { return warnBuf.String() }

// Nothing found is not everything broken.
//
// Searching "o-todo-poderoso" on 2026-09-24 had HiAnime, AnimeFire and Goyabu
// answer normally with no match while SuperFlix refused the connection, and it
// was reported as "every source failed". The two call for opposite reactions —
// try another title, versus wait for a host — so the failure has to carry
// enough to tell them apart.
func TestSearchFailure_DistinguishesNoMatchFromEverythingBroken(t *testing.T) {
	t.Parallel()

	partial := &SearchFailure{
		Query:    "o-todo-poderoso",
		Searched: 4,
		Sources:  twoSourceFailure().Sources[:1],
	}
	assert.False(t, partial.AllFailed(),
		"three sources answered; claiming every source failed tells the user their install is broken")
	assert.Equal(t, 3, partial.Answered())
	assert.NotContains(t, partial.Error(), "all sources failed")
	assert.Contains(t, partial.Error(), "no match")

	total := &SearchFailure{
		Query:    "dexter",
		Searched: 2,
		Sources:  twoSourceFailure().Sources,
	}
	assert.True(t, total.AllFailed(), "nothing answered at all; that IS every source failing")
	assert.Zero(t, total.Answered())
	assert.Contains(t, total.Error(), "all sources failed")
}

// A caller that has not been updated to fill Searched must not start making a
// new claim. Zero means unknown, and unknown keeps the old wording.
func TestSearchFailure_UnsetCountKeepsTheConservativeWording(t *testing.T) {
	t.Parallel()

	f := &SearchFailure{Query: "x", Sources: twoSourceFailure().Sources}
	assert.True(t, f.AllFailed())
	assert.Zero(t, f.Answered())
	assert.Contains(t, f.Error(), "all sources failed")
}

// finishSearch is what fills the count, and it has to be the number of sources
// actually asked — not the number that failed.
func TestFinishSearch_CarriesHowManySourcesWereAsked(t *testing.T) {
	t.Parallel()

	_, err := finishSearch("o-todo-poderoso", 4, nil, twoSourceFailure().Sources[:1])
	var f *SearchFailure
	require.ErrorAs(t, err, &f)
	assert.Equal(t, 4, f.Searched)
	assert.False(t, f.AllFailed())
	assert.Equal(t, 3, f.Answered())
}
