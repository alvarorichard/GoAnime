package superflix

import (
	"context"
	"fmt"
	"net/http"
	neturl "net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

// SuperFlix rotates its domain every few weeks (`.rest` → `.online` → `.best`
// → `.fit` → `.cyou` → `.lifestyle` → `.pro` → `.sbs` → `.beer` → …), and every
// rotation used to break playback until SuperFlixBase was hand-edited: Go's
// http.Client follows the 301 but downgrades the POST to a GET and drops the
// body, so /player/bootstrap answers HTML 404 and JSON decoding fails.
//
// Retired aliases keep 301-redirecting to whichever host is currently live, so
// the live host is discoverable at runtime: follow the chain from the compiled
// default and use wherever it lands. SuperFlixBase stays the seed and the
// fallback, so a failed probe degrades to today's behaviour rather than
// breaking.
const (
	// hostEnvOverride pins the host manually, skipping discovery entirely —
	// the escape hatch for a rotation that redirects somewhere the family
	// pattern below rejects. Value is a bare host ("superflixapi.beer") or a
	// full origin ("https://superflixapi.beer").
	hostEnvOverride = "GOANIME_SF_HOST"
	// hostProbeBudget caps the whole redirect walk. Discovery is on the
	// critical path of the first request, so it must fail fast.
	hostProbeBudget = 12 * time.Second
	// hostProbeMaxHops bounds the 301 chain. Observed rotations are one hop;
	// a few spare hops cover a stale alias pointing at another stale alias.
	hostProbeMaxHops = 5
)

// superflixHostRe matches the SuperFlix domain family. Discovery only accepts a
// redirect that lands inside it: the retired aliases are attacker-attractive
// (expired domains get re-registered), and following an arbitrary Location into
// a fresh origin would hand that origin our requests and cookies.
var superflixHostRe = regexp.MustCompile(`^superflixapi\.[a-z0-9-]{2,24}$`)

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
// running the redirect-chain discovery (once per process) if it has not run
// yet. Callers outside this package that need to build a SuperFlix URL and can
// afford one round trip should use this instead of SuperFlixBase.
func LiveBase(ctx context.Context) string {
	ensureLiveHost(ctx)
	return liveBase()
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

	if pinned := hostFromEnv(); pinned != "" {
		hostFound = pinned
		util.Debug("SuperFlix host pinned via " + hostEnvOverride + ": " + pinned)
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
	raw := strings.TrimSpace(os.Getenv(hostEnvOverride))
	if raw == "" {
		return ""
	}
	if u, err := neturl.Parse(raw); err == nil && u.Host != "" {
		return u.Host
	}
	return strings.TrimSuffix(raw, "/")
}

// probeLiveHost walks the 301 chain from the compiled default and returns the
// host that finally answers without redirecting.
//
// Redirects are handled manually (ErrUseLastResponse) rather than by the
// client: we need the host at each hop to vet it against superflixHostRe before
// following it.
func probeLiveHost(parent context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(parent, hostProbeBudget)
	defer cancel()

	client := &http.Client{
		Transport: netx.SafeScraperTransport(hostProbeBudget),
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	host := SuperFlixEmbedHost
	for hop := 0; hop < hostProbeMaxHops; hop++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+host+"/", http.NoBody)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", SuperFlixUserAgent)

		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		next, done, err := nextHopHost(host, resp.StatusCode, resp.Header.Get("Location"))
		_ = resp.Body.Close()
		if err != nil {
			return "", err
		}
		if done {
			return next, nil
		}
		host = next
	}
	return "", fmt.Errorf("superflix host discovery: redirect chain exceeded %d hops from %s",
		hostProbeMaxHops, SuperFlixEmbedHost)
}

// nextHopHost interprets one hop of the redirect walk: it returns the host to
// settle on (done) or the host to try next.
//
// A non-redirect status is terminal even when it is 403 or 503 — SuperFlix
// fronts every host with Cloudflare, and an interstitial means "this host is
// answering", which is exactly what discovery is looking for; clearing it is
// the browser solver's job, later.
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
