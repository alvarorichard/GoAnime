// Package movie — IMDb's own keyless title lookup.
package movie

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
)

// IMDbSuggestionURL is the endpoint behind the search box on imdb.com. Asked
// for an IMDb id, it answers with that title's name and year as IMDb lists
// it, with no API key and no account. The name is IMDb's English title even
// from a Brazilian address: "tt7526136" is "Heart of the Beast", not the
// "Coração Selvagem" StartFlix lists it as (checked 2026-10-08).
const IMDbSuggestionURL = "https://v3.sg.media-imdb.com/suggestion"

// ErrIMDbTitleNotFound means IMDb's answer did not include the asked-for id.
var ErrIMDbTitleNotFound = errors.New("IMDb has no title for this id")

// IMDbTitle is one title as IMDb's suggestion endpoint describes it.
type IMDbTitle struct {
	ID    string
	Title string
	Year  int
	Kind  string // IMDb's type id: "movie", "tvSeries", "tvMiniSeries", ...
}

// IMDbClient looks titles up on IMDb's suggestion endpoint.
type IMDbClient struct {
	client  *http.Client
	baseURL string
}

// NewIMDbClient creates a client on the SSRF-guarded movie transport.
func NewIMDbClient() *IMDbClient {
	return &IMDbClient{
		client: &http.Client{
			Timeout:   15 * time.Second,
			Transport: safeMovieTransport(15 * time.Second),
		},
		baseURL: IMDbSuggestionURL,
	}
}

// suggestion is one entry of IMDb's suggestion list.
type suggestion struct {
	ID    string `json:"id"`
	Label string `json:"l"`
	Year  int    `json:"y"`
	Kind  string `json:"qid"`
}

// suggest asks the suggestion endpoint for one shard-and-query path
// ("/t/tt0972534.json", "/i/icarly.json").
func (c *IMDbClient) suggest(path string) ([]suggestion, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", netx.EnglishAcceptLanguage)
	req.Header.Set("User-Agent", netx.APIUserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IMDb returned status: %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	var result struct {
		D []suggestion `json:"d"`
	}
	if err := jsonx.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse IMDb response: %w", err)
	}
	return result.D, nil
}

// TitleByID returns IMDb's name and year for an IMDb id.
func (c *IMDbClient) TitleByID(imdbID string) (*IMDbTitle, error) {
	if !imdbIDRe.MatchString(imdbID) {
		return nil, fmt.Errorf("invalid IMDb id %q", imdbID)
	}
	// The endpoint shards queries by their first letter: ids live under /t/.
	list, err := c.suggest("/t/" + imdbID + ".json")
	if err != nil {
		return nil, err
	}
	// The answer is a suggestion list; only the entry for this exact id counts.
	for _, d := range list {
		if d.ID == imdbID && d.Label != "" {
			return &IMDbTitle{ID: d.ID, Title: d.Label, Year: d.Year, Kind: d.Kind}, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrIMDbTitleNotFound, imdbID)
}

// SearchTitles returns the titles IMDb suggests for a name, best match first.
// Entries that are not titles (people, with no "tt" id) are left out. The name
// is IMDb's English one, so a Portuguese title often finds nothing.
func (c *IMDbClient) SearchTitles(name string) ([]IMDbTitle, error) {
	q := strings.ToLower(strings.TrimSpace(name))
	if q == "" {
		return nil, nil
	}
	// The endpoint shards queries by their first character, a letter or a
	// digit; anything else lands under "x".
	shard := "x"
	if r := q[0]; r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
		shard = string(r)
	}
	list, err := c.suggest("/" + shard + "/" + url.PathEscape(q) + ".json")
	if err != nil {
		return nil, err
	}
	var out []IMDbTitle
	for _, d := range list {
		if imdbIDRe.MatchString(d.ID) && d.Label != "" {
			out = append(out, IMDbTitle{ID: d.ID, Title: d.Label, Year: d.Year, Kind: d.Kind})
		}
	}
	return out, nil
}

// YearString is the title's year as text, or "" when IMDb gave none.
func (t *IMDbTitle) YearString() string {
	if t.Year <= 0 {
		return ""
	}
	return strconv.Itoa(t.Year)
}
