package superflix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

// SuperFlix rotates its domain every few weeks, and every rotation used to break
// playback until SuperFlixBase was hand-edited AND a release shipped: Go's
// http.Client follows the 301 but downgrades the POST to a GET and drops the
// body, so /player/bootstrap answers HTML and JSON decoding fails. Issue #199 is
// that failure on v1.8.6 (.pro → .monster), weeks after dev had already learned
// to follow a rotation on its own — the fix existed, it just never reached the
// people running the released binary.
//
// So discovery is built to keep working for a binary that is never updated. It
// resolves the live host through independent layers, each able to rescue the
// ones before it:
//
//  1. GOANIME_SF_HOST            manual pin, no network
//  2. last host that worked      remembered on disk from a previous run
//  3. every alias ever observed  walked concurrently, first live one wins
//  4. a pointer in the repo      one line a maintainer edits, read only when
//     everything above has failed
//  5. the compiled default       what a failed probe falls back to
//
// Two facts measured on 2026-09-14 shaped this. One seed is not enough: .sbs,
// the compiled seed only weeks earlier, now times out on a different IP, and a
// binary whose only seed dies has nothing left to walk from — while every other
// retired alias still redirected to .monster. And "answered" is not "is
// SuperFlix": a dead alias can come back as a parked page serving 200, which the
// old status-only check would have accepted as the live host. A host is only
// accepted once its page proves it is SuperFlix.
const (
	// hostEnvOverride pins the host manually, skipping discovery entirely —
	// the escape hatch for a rotation that redirects somewhere the family
	// pattern below rejects. Value is a bare host ("superflixapi.monster") or
	// a full origin ("https://superflixapi.monster").
	hostEnvOverride = "GOANIME_SF_HOST"
	// hostProbeBudget caps the whole resolution, pointer included. Discovery is
	// on the critical path of the first SuperFlix request, so it must fail fast.
	hostProbeBudget = 12 * time.Second
	// hostProbeMaxHops bounds each redirect chain. Observed rotations are one
	// hop; spare hops cover a stale alias pointing at another stale alias.
	hostProbeMaxHops = 5
	// hostVerifyReadLimit is how much of a page is read to verify it. The
	// <title> sits in the first kilobyte; this leaves room for a bloated head
	// without buffering the whole ~470KB homepage per candidate.
	hostVerifyReadLimit = 64 << 10
	// hostPointerURL is a one-line file in the repository naming the current
	// host. Editing it repairs every installed binary without a release — the
	// one thing that could have spared #199. Consulted only when every local
	// layer has failed, so a normal launch never touches GitHub.
	hostPointerURL = "https://raw.githubusercontent.com/alvarorichard/GoAnime/main/superflix-host.txt"
	// hostStateFileName holds the last host that verified, under the user
	// cache dir alongside the stream cache.
	hostStateFileName = "superflix-host.json"
)

// retiredSuperFlixHosts is every SuperFlix domain observed so far, newest first.
// Each is a discovery seed: SuperFlix keeps old aliases redirecting to the live
// host, so any one of them that is still up leads to it. Append the old host
// here whenever the compiled default is rotated — never delete entries, since
// an alias that looks dead today may be the last one still redirecting later.
var retiredSuperFlixHosts = []string{
	"superflixapi.baby",
	"superflixapi.beer",
	"superflixapi.sbs",
	"superflixapi.pro",
	"superflixapi.lifestyle",
	"superflixapi.cyou",
	"superflixapi.fit",
	"superflixapi.best",
	"superflixapi.online",
	"superflixapi.rest",
}

// superflixHostRe matches the SuperFlix domain family. Discovery only accepts a
// host inside it: retired aliases are attacker-attractive (expired domains get
// re-registered), and following an arbitrary Location into a fresh origin would
// hand that origin our requests and cookies.
var superflixHostRe = regexp.MustCompile(`^superflixapi\.[a-z0-9-]{2,24}$`)

// superflixPageMarker is the brand in the live site's <title>
// ("SuperFlixAPI - Início"). It is matched case-sensitively as a title, not as a
// loose substring: a parking page for an expired alias typically repeats the
// lowercase domain name ("superflixapi.sbs is for sale") but not this title.
var superflixPageMarker = []byte("<title>SuperFlixAPI")

// Discovery state. A plain mutex rather than sync.Once + atomic so that a
// concurrent reader BLOCKS on an in-flight probe instead of racing past it and
// building a URL on the stale seed. The probe runs once and every later call is
// an uncontended lock, so the cost is nil on the request path.
var (
	hostMu sync.Mutex
	// hostFound is the discovered host, empty until discovery succeeds.
	hostFound string
	// hostProbed records that discovery has already run (successfully or not),
	// so a failed probe is not retried on every request.
	hostProbed bool
)

// liveEmbedHost returns the host to target: the discovered one when discovery
// has run and succeeded, otherwise the compiled default. Call ensureLiveHost
// first to actually run discovery.
func liveEmbedHost() string {
	hostMu.Lock()
	defer hostMu.Unlock()
	if hostFound != "" {
		return hostFound
	}
	return SuperFlixEmbedHost
}

// liveBase is liveEmbedHost as an origin, the rotation-aware SuperFlixBase.
func liveBase() string { return "https://" + liveEmbedHost() }

// LiveBase returns the origin of the SuperFlix host that is live right now,
// running discovery (once per process) if it has not run yet. Callers outside
// this package that need to build a SuperFlix URL should use this instead of
// SuperFlixBase.
func LiveBase(ctx context.Context) string {
	ensureLiveHost(ctx)
	return liveBase()
}

// discoveryAllowed reports whether this process may spend a network round trip
// discovering the live host.
//
// Test binaries may not, unless they opted in with GOANIME_LIVE. Discovery used
// to be gated on a guess — "this client still has the default base URL and the
// real solver, so it must be production" — and a unit test that built a client
// with NewSuperFlixClient() and pointed only its http.Client at an httptest
// server slipped straight through it. That put a live network call inside
// `go test -short`, and worse, it left the discovered host in a package global
// where an unrelated parallel test compared it against the compiled constant
// and failed the moment the domain rotated.
//
// testing.Testing() is the supported way to ask, and unlike the guess it cannot
// be defeated by how a particular test happens to build its client.
func discoveryAllowed() bool {
	if !testing.Testing() {
		return true
	}
	return os.Getenv("GOANIME_LIVE") != ""
}

// ensureLiveHost resolves the currently live SuperFlix host, once per process.
// It is best-effort: any failure leaves liveEmbedHost on the compiled default.
func ensureLiveHost(ctx context.Context) {
	hostMu.Lock()
	defer hostMu.Unlock()
	if hostProbed {
		return
	}
	hostProbed = true

	// The manual pin comes first: it costs no network, so it applies in tests
	// too and stays the escape hatch when discovery is unavailable.
	if pinned := hostFromEnv(); pinned != "" {
		hostFound = pinned
		util.Debug("SuperFlix host pinned via " + hostEnvOverride + ": " + pinned)
		return
	}
	if !discoveryAllowed() {
		util.Debug("SuperFlix host discovery skipped under `go test`; using the compiled default",
			"default", SuperFlixEmbedHost)
		return
	}
	host, err := probeLiveHost(ctx)
	if err != nil {
		util.Debug("SuperFlix host discovery failed, using compiled default",
			"default", SuperFlixEmbedHost, "err", err)
		return
	}
	hostFound = host
	if host != SuperFlixEmbedHost {
		util.Debug("SuperFlix host rotated", "from", SuperFlixEmbedHost, "to", host)
	}
}

// hostFromEnv reads and normalizes the GOANIME_SF_HOST override, accepting
// either a bare host or a full URL.
func hostFromEnv() string {
	return normalizeHost(os.Getenv(hostEnvOverride))
}

// normalizeHost reduces "superflixapi.x", "https://superflixapi.x/" and similar
// to a bare lowercase host, or "" for blank input.
func normalizeHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if u, err := neturl.Parse(raw); err == nil && u.Host != "" {
		return strings.ToLower(u.Host)
	}
	return strings.ToLower(strings.TrimSuffix(raw, "/"))
}

// probeLiveHost resolves the live host through every network layer. Kept as the
// single entry point the live tests exercise.
func probeLiveHost(ctx context.Context) (string, error) {
	return newHostProber().resolve(ctx)
}

// hostProber runs discovery. Its network and filesystem touch points are fields
// so tests can drive every layer offline with a fake transport and a temp dir.
type hostProber struct {
	// client must not follow redirects: each hop is vetted against the family
	// pattern before it is taken.
	client *http.Client
	// seeds are tried concurrently, persisted host first.
	seeds []string
	// pointerURL is the last-resort pointer; empty disables it.
	pointerURL string
	// statePath remembers the last verified host; empty disables persistence.
	statePath string
}

func newHostProber() *hostProber {
	return &hostProber{
		client: &http.Client{
			Transport: netx.SafeScraperTransport(hostProbeBudget),
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		seeds:      append([]string{SuperFlixEmbedHost}, retiredSuperFlixHosts...),
		pointerURL: hostPointerURL,
		statePath:  defaultHostStatePath(),
	}
}

// defaultHostStatePath is where the last verified host is remembered, or "" if
// the platform has no cache dir (persistence is then simply skipped).
func defaultHostStatePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "goanime", hostStateFileName)
}

// resolve walks the layers in order and remembers whichever host verifies.
func (p *hostProber) resolve(parent context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(parent, hostProbeBudget)
	defer cancel()

	seeds := dedupeHosts(append([]string{p.loadPersisted()}, p.seeds...))
	host, seedErr := p.race(ctx, seeds)
	if seedErr == nil {
		p.persist(host)
		return host, nil
	}

	// Every built-in alias failed. That is the case no amount of client code
	// can fix — SuperFlix moved somewhere none of our aliases point to — so ask
	// the one source that can be corrected without shipping a binary.
	pointed, ptrErr := p.fetchPointer(ctx)
	if ptrErr != nil {
		return "", fmt.Errorf("superflix host discovery: %w; pointer fallback: %v", seedErr, ptrErr)
	}
	host, err := p.walkAndVerify(ctx, pointed)
	if err != nil {
		return "", fmt.Errorf("superflix host discovery: %w; pointer named %s but %v", seedErr, pointed, err)
	}
	util.Debug("SuperFlix host resolved from the repository pointer", "host", host)
	p.persist(host)
	return host, nil
}

// race walks every seed at once and returns the first host that verifies.
// Concurrency is what makes a large seed list affordable: a dead alias costs
// its own timeout, not the whole budget, and the live chain usually answers in
// a few hundred milliseconds.
func (p *hostProber) race(parent context.Context, seeds []string) (string, error) {
	if len(seeds) == 0 {
		return "", errors.New("no seeds to walk")
	}
	ctx, cancel := context.WithCancel(parent)
	defer cancel() // stop the losers as soon as there is a winner

	type outcome struct {
		host string
		err  error
	}
	// Buffered to len(seeds) so every walker can report and exit even after
	// the race is decided — nothing is left blocked on a send.
	results := make(chan outcome, len(seeds))
	for _, seed := range seeds {
		go func(seed string) {
			host, err := p.walkAndVerify(ctx, seed)
			results <- outcome{host, err}
		}(seed)
	}

	var errs []error
	for range seeds {
		r := <-results
		if r.err == nil {
			return r.host, nil
		}
		errs = append(errs, r.err)
	}
	return "", fmt.Errorf("none of %d known aliases led to a live SuperFlix host: %w", len(seeds), errors.Join(errs...))
}

// walkAndVerify follows the redirect chain from seed and accepts where it lands
// only if the page there is SuperFlix.
func (p *hostProber) walkAndVerify(ctx context.Context, seed string) (string, error) {
	host := normalizeHost(seed)
	if !superflixHostRe.MatchString(host) {
		return "", fmt.Errorf("%q is outside the SuperFlix domain family", seed)
	}
	for hop := 0; hop < hostProbeMaxHops; hop++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+host+"/", http.NoBody)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", SuperFlixUserAgent)

		resp, err := p.client.Do(req)
		if err != nil {
			return "", fmt.Errorf("%s: %w", host, err)
		}
		next, done, err := nextHopHost(host, resp.StatusCode, resp.Header.Get("Location"))
		if err != nil || !done {
			_ = resp.Body.Close()
			if err != nil {
				return "", err
			}
			host = next
			continue
		}

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, hostVerifyReadLimit))
		_ = resp.Body.Close()
		if readErr != nil {
			return "", fmt.Errorf("%s: reading page: %w", host, readErr)
		}
		if !looksLikeSuperFlix(resp, body) {
			return "", fmt.Errorf("%s answered %d but is not SuperFlix (parked or retired alias?)", host, resp.StatusCode)
		}
		return host, nil
	}
	return "", fmt.Errorf("%s: redirect chain exceeded %d hops", seed, hostProbeMaxHops)
}

// looksLikeSuperFlix reports whether a terminal page is the live site.
//
// Two shapes count. The site itself, recognised by its title. Or a Cloudflare
// challenge in front of it: SuperFlix sits behind Turnstile, and a live host
// that challenges the probe is still the live host — clearing it is the browser
// solver's job later. The challenge is recognised by Cloudflare's own markers
// rather than by status alone, so a parked page answering 403 is not mistaken
// for one.
func looksLikeSuperFlix(resp *http.Response, body []byte) bool {
	if bytes.Contains(body, superflixPageMarker) {
		return true
	}
	return resp.Header.Get("cf-mitigated") == "challenge" || bodyHasChallengeMarker(body)
}

// fetchPointer reads the repository pointer and returns the first host in it
// that belongs to the family. The file is plain text: one host per line, blank
// lines and #-comments ignored, so a maintainer can annotate it.
func (p *hostProber) fetchPointer(ctx context.Context) (string, error) {
	if p.pointerURL == "" {
		return "", errors.New("pointer disabled")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.pointerURL, http.NoBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", netx.APIUserAgent)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pointer answered %s", resp.Status)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	if err != nil {
		return "", err
	}
	return parseHostPointer(raw)
}

// parseHostPointer extracts the first family host from pointer file contents.
func parseHostPointer(raw []byte) (string, error) {
	for line := range strings.SplitSeq(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if host := normalizeHost(line); superflixHostRe.MatchString(host) {
			return host, nil
		}
	}
	return "", errors.New("pointer names no SuperFlix host")
}

// hostState is the on-disk record of the last verified host.
type hostState struct {
	Host       string `json:"host"`
	VerifiedAt int64  `json:"verified_at"`
}

// loadPersisted returns the remembered host, or "" when there is none. It is
// only ever used as the first seed — it still has to verify like any other, so
// a stale record costs one failed walk, never a wrong host.
func (p *hostProber) loadPersisted() string {
	if p.statePath == "" {
		return ""
	}
	raw, err := os.ReadFile(p.statePath)
	if err != nil {
		return ""
	}
	var st hostState
	if json.Unmarshal(raw, &st) != nil {
		return ""
	}
	if host := normalizeHost(st.Host); superflixHostRe.MatchString(host) {
		return host
	}
	return ""
}

// persist remembers a verified host so the next launch starts from it. This is
// what keeps a binary working after every alias it shipped with has died: the
// host it last reached is still on disk. Best-effort; a failure only costs that
// shortcut.
func (p *hostProber) persist(host string) {
	if p.statePath == "" {
		return
	}
	raw, err := json.Marshal(hostState{Host: host, VerifiedAt: time.Now().Unix()})
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(p.statePath), 0o700); err != nil {
		return
	}
	_ = os.WriteFile(p.statePath, raw, 0o600)
}

// dedupeHosts normalizes seeds and drops blanks and repeats, keeping order.
func dedupeHosts(hosts []string) []string {
	seen := make(map[string]bool, len(hosts))
	out := make([]string, 0, len(hosts))
	for _, h := range hosts {
		h = normalizeHost(h)
		if h == "" || seen[h] {
			continue
		}
		seen[h] = true
		out = append(out, h)
	}
	return out
}

// nextHopHost interprets one hop of the redirect walk: it returns the host to
// settle on (done) or the host to try next.
//
// A non-redirect status is terminal even when it is 403 or 503 — SuperFlix
// fronts every host with Cloudflare, and an interstitial means "this host is
// answering". Whether it is SuperFlix answering is decided afterwards, from the
// page, by looksLikeSuperFlix.
func nextHopHost(current string, status int, location string) (host string, done bool, err error) {
	if status < 300 || status > 399 || location == "" {
		return current, true, nil
	}
	next, parseErr := neturl.Parse(location)
	if parseErr != nil {
		return "", false, fmt.Errorf("superflix host discovery: unparseable Location %q: %w", location, parseErr)
	}
	// A relative or same-host Location (http→https, trailing-slash
	// canonicalization) is not a rotation — this host is already the live one.
	if next.Host == "" || next.Host == current {
		return current, true, nil
	}
	// Retired SuperFlix domains expire and get re-registered, so a Location
	// pointing outside the family is not a rotation to follow — following it
	// would hand a stranger our requests and Cloudflare cookies.
	if !superflixHostRe.MatchString(next.Host) {
		return "", false, fmt.Errorf("superflix host discovery: %s redirected outside the domain family (%s)",
			current, next.Host)
	}
	return next.Host, false, nil
}
