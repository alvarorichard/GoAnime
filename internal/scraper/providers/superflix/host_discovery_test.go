package superflix

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeHosts answers requests by host name, so every discovery layer can be
// driven offline with the real domain names and the real family pattern. A host
// missing from the map behaves like a dead domain.
type fakeHosts map[string]func(*http.Request) *http.Response

func (f fakeHosts) RoundTrip(r *http.Request) (*http.Response, error) {
	if h, ok := f[r.URL.Host]; ok {
		return h(r), nil
	}
	return nil, fmt.Errorf("dial tcp: lookup %s: i/o timeout", r.URL.Host)
}

func respond(status int, header http.Header, body string) func(*http.Request) *http.Response {
	return func(r *http.Request) *http.Response {
		if header == nil {
			header = http.Header{}
		}
		return &http.Response{
			StatusCode: status,
			Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
			Header:     header,
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}
	}
}

func redirectTo(location string) func(*http.Request) *http.Response {
	return respond(http.StatusMovedPermanently, http.Header{"Location": {location}}, "")
}

// liveSuperFlixHome is the head of the real homepage, captured 2026-09-14.
const liveSuperFlixHome = `<!DOCTYPE html><html lang="pt-BR"><head><meta charset="UTF-8">` +
	`<title>SuperFlixAPI - Início</title></head><body>...</body></html>`

func newTestProber(t *testing.T, hosts fakeHosts, seeds ...string) *hostProber {
	t.Helper()
	return &hostProber{
		client: &http.Client{
			Transport: hosts,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		seeds:     seeds,
		statePath: filepath.Join(t.TempDir(), hostStateFileName),
	}
}

// TestHostProber_Issue199_StaleAliasFollowsToMonster reproduces issue #199 as
// reported: v1.8.6 still targets .pro, which now 301s to .monster.
func TestHostProber_Issue199_StaleAliasFollowsToMonster(t *testing.T) {
	t.Parallel()
	p := newTestProber(t, fakeHosts{
		"superflixapi.pro":     redirectTo("https://superflixapi.monster/"),
		"superflixapi.monster": respond(200, nil, liveSuperFlixHome),
	}, "superflixapi.pro")

	host, err := p.resolve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "superflixapi.monster", host)
}

// TestHostProber_DeadSeedDoesNotStrandDiscovery is the case one seed could not
// survive. .sbs was the compiled seed only weeks before it stopped answering; a
// binary shipping it as the ONLY seed would have had nothing left to walk from,
// while every other retired alias still pointed at the live host.
func TestHostProber_DeadSeedDoesNotStrandDiscovery(t *testing.T) {
	t.Parallel()
	p := newTestProber(t, fakeHosts{
		// superflixapi.sbs deliberately absent: it times out.
		"superflixapi.baby":    redirectTo("https://superflixapi.monster/"),
		"superflixapi.monster": respond(200, nil, liveSuperFlixHome),
	}, "superflixapi.sbs", "superflixapi.baby")

	host, err := p.resolve(context.Background())
	require.NoError(t, err, "a dead seed must not sink discovery while another alias still redirects")
	assert.Equal(t, "superflixapi.monster", host)
}

// TestHostProber_RejectsParkedPage closes the hole the old status-only check
// left open: any non-redirect answer counted as "live", so an expired alias
// re-registered as a parking page serving 200 would have been adopted — and
// every SuperFlix request sent to a stranger.
func TestHostProber_RejectsParkedPage(t *testing.T) {
	t.Parallel()
	p := newTestProber(t, fakeHosts{
		"superflixapi.sbs": respond(200, nil,
			`<html><head><title>superflixapi.sbs is for sale</title></head><body>Buy this domain</body></html>`),
	}, "superflixapi.sbs")

	_, err := p.resolve(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not SuperFlix")
}

// A parked alias must not win the race just by answering first.
func TestHostProber_ParkedAliasDoesNotWinTheRace(t *testing.T) {
	t.Parallel()
	p := newTestProber(t, fakeHosts{
		"superflixapi.sbs":     respond(200, nil, `<title>Parked Domain</title>`),
		"superflixapi.baby":    redirectTo("https://superflixapi.monster/"),
		"superflixapi.monster": respond(200, nil, liveSuperFlixHome),
	}, "superflixapi.sbs", "superflixapi.baby")

	host, err := p.resolve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "superflixapi.monster", host)
}

// SuperFlix sits behind Turnstile: a live host that challenges the probe is
// still the live host. Recognised by Cloudflare's own signal, not by status.
func TestHostProber_CloudflareChallengeCountsAsLive(t *testing.T) {
	t.Parallel()
	p := newTestProber(t, fakeHosts{
		"superflixapi.monster": respond(403, http.Header{"Cf-Mitigated": {"challenge"}},
			`<html><head><title>Verificação</title></head></html>`),
	}, "superflixapi.monster")

	host, err := p.resolve(context.Background())
	require.NoError(t, err, "a Turnstile interstitial on the live host is not a dead host")
	assert.Equal(t, "superflixapi.monster", host)
}

// A bare 403 with no Cloudflare signal and no SuperFlix title proves nothing.
func TestHostProber_Plain403IsNotProofOfLife(t *testing.T) {
	t.Parallel()
	p := newTestProber(t, fakeHosts{
		"superflixapi.sbs": respond(403, nil, `<html><body>Forbidden</body></html>`),
	}, "superflixapi.sbs")

	_, err := p.resolve(context.Background())
	require.Error(t, err)
}

// TestHostProber_RemembersTheLastLiveHost covers the layer that keeps an old
// binary alive after every alias it shipped with is gone: the host it last
// reached is on disk and seeds the next launch.
func TestHostProber_RemembersTheLastLiveHost(t *testing.T) {
	t.Parallel()
	live := fakeHosts{
		"superflixapi.baby":    redirectTo("https://superflixapi.monster/"),
		"superflixapi.monster": respond(200, nil, liveSuperFlixHome),
	}
	first := newTestProber(t, live, "superflixapi.baby")
	_, err := first.resolve(context.Background())
	require.NoError(t, err)

	// Next launch: every shipped alias is dead, only the remembered host answers.
	second := newTestProber(t, fakeHosts{
		"superflixapi.monster": respond(200, nil, liveSuperFlixHome),
	}, "superflixapi.baby", "superflixapi.sbs")
	second.statePath = first.statePath

	host, err := second.resolve(context.Background())
	require.NoError(t, err, "the remembered host must rescue a launch where every shipped alias is dead")
	assert.Equal(t, "superflixapi.monster", host)
}

// A remembered host is a seed, not an authority: it must verify like any other.
func TestHostProber_StaleRememberedHostIsNotTrustedBlindly(t *testing.T) {
	t.Parallel()
	p := newTestProber(t, fakeHosts{
		"superflixapi.old":     respond(200, nil, `<title>Parked</title>`),
		"superflixapi.monster": respond(200, nil, liveSuperFlixHome),
	}, "superflixapi.monster")
	require.NoError(t, os.WriteFile(p.statePath, []byte(`{"host":"superflixapi.old","verified_at":1}`), 0o600))

	host, err := p.resolve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "superflixapi.monster", host)
}

func TestHostProber_PersistedHostOutsideFamilyIsIgnored(t *testing.T) {
	t.Parallel()
	p := newTestProber(t, fakeHosts{})
	require.NoError(t, os.WriteFile(p.statePath, []byte(`{"host":"evil.example"}`), 0o600))
	assert.Empty(t, p.loadPersisted(), "a tampered state file must not inject a foreign host")
}

// TestHostProber_PointerRescuesWhenEveryAliasIsGone is the layer that makes a
// release unnecessary: SuperFlix moved somewhere none of the shipped aliases
// point to, and a one-line edit in the repository repairs installed binaries.
func TestHostProber_PointerRescuesWhenEveryAliasIsGone(t *testing.T) {
	t.Parallel()
	p := newTestProber(t, fakeHosts{
		"raw.githubusercontent.com": respond(200, nil, "# current host\n\nsuperflixapi.newhost\n"),
		"superflixapi.newhost":      respond(200, nil, liveSuperFlixHome),
	}, "superflixapi.baby", "superflixapi.sbs")
	p.pointerURL = hostPointerURL

	host, err := p.resolve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "superflixapi.newhost", host)

	remembered := &hostProber{statePath: p.statePath}
	assert.Equal(t, "superflixapi.newhost", remembered.loadPersisted(),
		"a host found through the pointer is remembered too, so the pointer is read once")
}

// The pointer is last resort: a normal launch must never touch GitHub.
func TestHostProber_PointerNotReadWhenAnAliasWorks(t *testing.T) {
	t.Parallel()
	var pointerHits atomic.Int32
	p := newTestProber(t, fakeHosts{
		"raw.githubusercontent.com": func(r *http.Request) *http.Response {
			pointerHits.Add(1)
			return respond(200, nil, "superflixapi.monster")(r)
		},
		"superflixapi.monster": respond(200, nil, liveSuperFlixHome),
	}, "superflixapi.monster")
	p.pointerURL = hostPointerURL

	_, err := p.resolve(context.Background())
	require.NoError(t, err)
	assert.Zero(t, pointerHits.Load())
}

// Whatever the pointer says still has to prove it is SuperFlix.
func TestHostProber_PointerCannotAdoptAnUnverifiedHost(t *testing.T) {
	t.Parallel()
	p := newTestProber(t, fakeHosts{
		"raw.githubusercontent.com": respond(200, nil, "superflixapi.parked\n"),
		"superflixapi.parked":       respond(200, nil, `<title>Parked</title>`),
	}, "superflixapi.sbs")
	p.pointerURL = hostPointerURL

	_, err := p.resolve(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pointer named superflixapi.parked")
}

func TestParseHostPointer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, in, want string
		wantErr        bool
	}{
		{name: "bare host", in: "superflixapi.monster", want: "superflixapi.monster"},
		{name: "origin form", in: "https://superflixapi.monster/", want: "superflixapi.monster"},
		{name: "comments and blanks skipped", in: "# note\n\n  superflixapi.monster  \n", want: "superflixapi.monster"},
		{name: "first family host wins", in: "evil.example\nsuperflixapi.aa\nsuperflixapi.bb", want: "superflixapi.aa"},
		// The family pattern requires a TLD of at least two characters.
		{name: "single-letter TLD is not a real domain", in: "superflixapi.a", wantErr: true},
		{name: "uppercase normalised", in: "SuperFlixAPI.Monster", want: "superflixapi.monster"},
		{name: "outside the family", in: "evil.example", wantErr: true},
		{name: "lookalike subdomain", in: "superflixapi.monster.evil.example", wantErr: true},
		{name: "only comments", in: "# nothing here", wantErr: true},
		{name: "empty", in: "", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseHostPointer([]byte(tc.in))
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestRepositoryPointerFileIsValid guards the pointer file itself: if a
// maintainer typoes it, the rescue layer silently becomes a no-op on the day
// it is needed.
func TestRepositoryPointerFileIsValid(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "superflix-host.txt"))
	require.NoError(t, err, "superflix-host.txt must exist at the repository root")
	host, err := parseHostPointer(raw)
	require.NoError(t, err)
	assert.Equal(t, SuperFlixEmbedHost, host,
		"the pointer should name the same host as the compiled default; update both on rotation")
}

// A dead alias that hangs must cost only its own slot, not delay the winner.
func TestHostProber_SlowDeadAliasDoesNotDelayTheWinner(t *testing.T) {
	t.Parallel()
	p := newTestProber(t, fakeHosts{
		"superflixapi.sbs": func(r *http.Request) *http.Response {
			select { // hang until the race cancels us
			case <-r.Context().Done():
			case <-time.After(10 * time.Second):
			}
			return respond(504, nil, "")(r)
		},
		"superflixapi.monster": respond(200, nil, liveSuperFlixHome),
	}, "superflixapi.sbs", "superflixapi.monster")

	start := time.Now()
	host, err := p.resolve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "superflixapi.monster", host)
	assert.Less(t, time.Since(start), 2*time.Second, "the race must not wait for a hanging alias")
}

func TestHostProber_AllFailReportsEveryAlias(t *testing.T) {
	t.Parallel()
	p := newTestProber(t, fakeHosts{}, "superflixapi.baby", "superflixapi.sbs")

	_, err := p.resolve(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of 2 known aliases")
	assert.Contains(t, err.Error(), "superflixapi.baby")
	assert.Contains(t, err.Error(), "superflixapi.sbs")
	var target interface{ Unwrap() []error }
	assert.True(t, errors.As(err, &target) || strings.Contains(err.Error(), "superflixapi.sbs"))
}

// TestSeeds_CoverEveryObservedAlias pins the seed list. Removing an entry
// throws away a path back to the live host; the compiled default must lead.
func TestSeeds_CoverEveryObservedAlias(t *testing.T) {
	t.Parallel()
	p := newHostProber()
	require.NotEmpty(t, p.seeds)
	assert.Equal(t, SuperFlixEmbedHost, p.seeds[0], "the compiled default is tried first")

	for _, alias := range []string{
		"superflixapi.pro", // the host v1.8.6 shipped — issue #199
		"superflixapi.baby", "superflixapi.beer", "superflixapi.sbs",
		"superflixapi.lifestyle", "superflixapi.cyou", "superflixapi.fit",
		"superflixapi.best", "superflixapi.online", "superflixapi.rest",
	} {
		assert.Contains(t, p.seeds, alias)
	}
	for _, s := range p.seeds {
		assert.Regexpf(t, superflixHostRe, s, "seed %q must be inside the family", s)
	}
	assert.NotContains(t, retiredSuperFlixHosts, SuperFlixEmbedHost,
		"the live default must not also be listed as retired")
}

func TestDedupeHosts(t *testing.T) {
	t.Parallel()
	assert.Equal(t,
		[]string{"superflixapi.monster", "superflixapi.baby"},
		dedupeHosts([]string{"", "superflixapi.monster", "SUPERFLIXAPI.MONSTER", "https://superflixapi.baby/", "superflixapi.baby"}))
}
