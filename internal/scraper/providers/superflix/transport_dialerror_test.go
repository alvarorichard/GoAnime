package superflix

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errTransport fails every request the way a refused connection does.
type errTransport struct{ calls int }

func (e *errTransport) RoundTrip(*http.Request) (*http.Response, error) {
	e.calls++
	return nil, errors.New("dial tcp: connection reset by peer")
}

func newErrTransport(gate *rateLimitGate) (*cfFallbackTransport, *errTransport) {
	base := &errTransport{}
	return &cfFallbackTransport{base: base, timeout: time.Second, gate: gate}, base
}

func get(t *testing.T, tr *cfFallbackTransport, url string) error {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	require.NoError(t, err)
	_, rtErr := tr.RoundTrip(req)
	return rtErr
}

// A refused connection on an endpoint that has already rate limited us is the
// same block escalating, and it must arm the gate exactly like a 429 does.
//
// This is the failure the user saw as "still no SuperFlix results" after the
// header fixes had landed. The endpoint had stopped answering 429 and started
// dropping the connection; the transport returned every dial error straight
// out, so the back-off never armed, so every search went back at an endpoint
// only time could release — renewing the block each time.
func TestTransport_ARefusedConnectionArmsTheGateOnAKnownEndpoint(t *testing.T) {
	t.Parallel()
	// A past 429 whose block has since lapsed: the endpoint is KNOWN to refuse
	// us, but is not currently gated. The clock is driven rather than slept
	// through, because the floor is a minute.
	now := time.Now()
	gate := newRateLimitGate("")
	gate.now = func() time.Time { return now }
	gate.arm("superflixapi.quest", "/pesquisar", time.Second)
	now = now.Add(time.Hour)
	require.Zero(t, gate.retryIn("superflixapi.quest", "/pesquisar"))
	require.True(t, gate.knows("superflixapi.quest", "/pesquisar"))

	tr, base := newErrTransport(gate)
	require.Error(t, get(t, tr, "https://superflixapi.quest/pesquisar?s=matrix"))
	assert.Equal(t, 1, base.calls)

	assert.Positive(t, gate.retryIn("superflixapi.quest", "/pesquisar"),
		"the dial error left the endpoint open for the next search to hammer")

	// And the next request is suppressed rather than sent.
	require.Error(t, get(t, tr, "https://superflixapi.quest/pesquisar?s=matrix"))
	assert.Equal(t, 1, base.calls, "a suppressed request must not reach the network")
}

// An endpoint that has never refused us is not suspect: a dial error there is
// an ordinary network failure, and gating it would turn a flaky connection into
// a self-imposed outage.
func TestTransport_ARefusedConnectionOnAHealthyEndpointIsJustAFailure(t *testing.T) {
	t.Parallel()
	gate := newRateLimitGate("")

	tr, base := newErrTransport(gate)
	require.Error(t, get(t, tr, "https://superflixapi.quest/pesquisar?s=matrix"))
	assert.Zero(t, gate.retryIn("superflixapi.quest", "/pesquisar"),
		"a first-ever network failure must not gate the endpoint")

	require.Error(t, get(t, tr, "https://superflixapi.quest/pesquisar?s=matrix"))
	assert.Equal(t, 2, base.calls, "both requests should have been attempted")
}

// The gating is per endpoint, like the host's own guard: a blocked /pesquisar
// must not stop us reading the homepage, which measurably kept answering 200
// throughout.
func TestTransport_ARefusedConnectionGatesOnlyThatEndpoint(t *testing.T) {
	t.Parallel()
	now := time.Now()
	gate := newRateLimitGate("")
	gate.now = func() time.Time { return now }
	gate.arm("superflixapi.quest", "/pesquisar", time.Second)
	now = now.Add(time.Hour)

	tr, base := newErrTransport(gate)
	require.Error(t, get(t, tr, "https://superflixapi.quest/pesquisar?s=matrix"))
	require.Positive(t, gate.retryIn("superflixapi.quest", "/pesquisar"))

	before := base.calls
	require.Error(t, get(t, tr, "https://superflixapi.quest/"))
	assert.Equal(t, before+1, base.calls,
		"the homepage was suppressed by a block that belongs to /pesquisar")
	assert.Zero(t, gate.retryIn("superflixapi.quest", "/"))
}
