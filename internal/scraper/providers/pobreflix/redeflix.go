package pobreflix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// Since 2026-11-02 the site lists, next to the Turnstile player, a RedeFlixApi
// player page (redeflixapi.bar/filme/<tmdb>). It has no Turnstile: the page
// hands every visitor a signed ticket, and the player's own script trades it in
// two POSTs, exactly as a browser does:
//
//  1. <origin>/playback-resolve.php  {ticket, server}  ->  {launchTicket, nextTicket}
//  2. <origin>/playerkys/source.php  {launchTicket}    ->  {url}
//
// "server" is "default" (the dubbed track the page lists) or "legendado". Both
// calls carry the player's own origin, as its script's requests do. Nothing
// here names another site or fakes an embedding page: a request without that
// origin is answered "Origem de reprodução não autorizada".
//
// The host is not hard-coded: the catalog's server list says where the player
// is, and ResolveRedeflix recognizes the format from the page itself.

// ErrNotRedeflixPlayer means the page is not in the RedeFlixApi format, so the
// caller should try the Turnstile path.
var ErrNotRedeflixPlayer = errors.New("pobreflix: the player page is not a RedeFlixApi player")

var (
	redeflixTicketRe     = regexp.MustCompile(`var\s+playbackResolveTicket\s*=\s*"([A-Za-z0-9._~-]+)"`)
	redeflixResolveURLRe = regexp.MustCompile(`var\s+playbackResolveUrl\s*=\s*"([^"\s]+)"`)
)

const (
	redeflixSourcePath = "/playerkys/source.php"
	redeflixShellPath  = "/playerkys/index.html"
	maxPlayerJSONBytes = 1 << 20
)

// envOpenPlayerHosts lists extra hosts (comma separated) known to serve the
// RedeFlixApi player, for when the site moves it without a release.
const envOpenPlayerHosts = "GOANIME_POBREFLIX_OPEN_PLAYER_HOSTS"

// IsSolverFreePlayer reports whether rawURL is on a host that serves the
// RedeFlixApi player, which needs no Turnstile solver. It is judged by the host
// name alone, with no request, so that a title whose player pages are all
// behind the Turnstile costs nothing when no solver is installed. The page's
// format is still checked by ResolveRedeflix itself.
func (c *Client) IsSolverFreePlayer(rawURL string) bool {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Hostname() == "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if strings.Contains(host, "redeflix") {
		return true
	}
	for _, h := range strings.Split(os.Getenv(envOpenPlayerHosts), ",") {
		if h = strings.ToLower(strings.TrimSpace(h)); h != "" && host == h {
			return true
		}
	}
	return false
}

// HasSolverFreePlayer reports whether any listed player page is on such a host.
func (c *Client) HasSolverFreePlayer(servers []Server) bool {
	for _, s := range GatedServers(servers) {
		if c.IsSolverFreePlayer(s.URL) {
			return true
		}
	}
	return false
}

// playerResponse is what both RedeFlixApi endpoints answer.
type playerResponse struct {
	URL          string `json:"url"`
	LaunchTicket string `json:"launchTicket"`
	Error        string `json:"error"`
}

// ResolveRedeflix returns the media URL behind a RedeFlixApi player page,
// choosing the track by want. It needs no solver. A page in another format
// yields ErrNotRedeflixPlayer without any further request.
func (c *Client) ResolveRedeflix(ctx context.Context, serverURL string, want Audio) (Stream, error) {
	pageURL := strings.TrimSpace(serverURL)
	if pageURL == "" {
		return Stream{}, ErrNotRedeflixPlayer
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return Stream{}, err
	}
	hc := &http.Client{Jar: jar, Timeout: playerHTTPTimeout, Transport: c.playerTransport}

	body, finalURL, status, err := c.fetchPlayerPage(ctx, hc, pageURL, c.base()+"/")
	if err != nil {
		return Stream{}, fmt.Errorf("pobreflix: player page: %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return Stream{}, fmt.Errorf("pobreflix: player page answered HTTP %d", status)
	}
	m := redeflixTicketRe.FindSubmatch(body)
	if m == nil {
		return Stream{}, ErrNotRedeflixPlayer
	}
	ticket := string(m[1])

	page, err := url.Parse(finalURL)
	if err != nil || page.Host == "" || page.Scheme != "https" && page.Scheme != "http" {
		return Stream{}, ErrNotRedeflixPlayer
	}
	origin := page.Scheme + "://" + page.Host

	// The page names its resolve endpoint; it must be on the page's own host, so
	// a ticket is never sent anywhere else.
	resolveURL := origin + "/playback-resolve.php"
	if r := redeflixResolveURLRe.FindSubmatch(body); r != nil {
		u, perr := url.Parse(string(r[1]))
		if perr != nil || !strings.EqualFold(u.Host, page.Host) {
			return Stream{}, fmt.Errorf("pobreflix: the player names a resolve endpoint on another host")
		}
		resolveURL = u.String()
	}

	server := "default"
	have := AudioDubbed
	switch want {
	case AudioSubtitled:
		server, have = "legendado", AudioSubtitled
	case AudioUnknown:
		have = AudioUnknown
	}

	var step1 playerResponse
	if err := c.postPlayerJSON(ctx, hc, resolveURL, origin, finalURL,
		map[string]string{"ticket": ticket, "server": server}, &step1); err != nil {
		return Stream{}, fmt.Errorf("pobreflix: resolving the %s track: %w", server, err)
	}
	mediaURL := step1.URL
	if mediaURL == "" {
		if step1.LaunchTicket == "" {
			return Stream{}, errors.New("pobreflix: the player answered with neither a media URL nor a launch ticket")
		}
		var step2 playerResponse
		if err := c.postPlayerJSON(ctx, hc, origin+redeflixSourcePath, origin, origin+redeflixShellPath,
			map[string]string{"ticket": step1.LaunchTicket}, &step2); err != nil {
			return Stream{}, fmt.Errorf("pobreflix: opening the %s source: %w", server, err)
		}
		mediaURL = step2.URL
	}
	mu, err := url.Parse(mediaURL)
	if err != nil || mu.Host == "" || mu.Scheme != "https" {
		return Stream{}, errors.New("pobreflix: the player returned no https media URL")
	}
	return Stream{URL: mu.String(), Audio: have}, nil
}

// postPlayerJSON sends one of the player script's own POSTs and decodes the
// JSON answer into out. The server's error text is kept; tickets are not
// echoed anywhere.
func (c *Client) postPlayerJSON(ctx context.Context, hc *http.Client, endpoint, origin, referer string, payload map[string]string, out *playerResponse) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "RedeFlixPlayer")
	req.Header.Set("Origin", origin)
	req.Header.Set("Referer", referer)
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxPlayerJSONBytes))
	if err != nil {
		return err
	}
	_ = json.Unmarshal(b, out) // an HTML error page leaves out empty; the status says it
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if out.Error != "" {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, out.Error)
		}
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}
