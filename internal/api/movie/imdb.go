// Package movie — IMDb's own keyless title lookup.
package movie

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
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

// TitleByID returns IMDb's name and year for an IMDb id.
func (c *IMDbClient) TitleByID(imdbID string) (*IMDbTitle, error) {
	if !imdbIDRe.MatchString(imdbID) {
		return nil, fmt.Errorf("invalid IMDb id %q", imdbID)
	}
	// The endpoint shards queries by their first letter: ids live under /t/.
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/t/"+imdbID+".json", http.NoBody)
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
		D []struct {
			ID    string `json:"id"`
			Label string `json:"l"`
			Year  int    `json:"y"`
			Kind  string `json:"qid"`
		} `json:"d"`
	}
	if err := jsonx.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse IMDb response: %w", err)
	}
	// The answer is a suggestion list; only the entry for this exact id counts.
	for _, d := range result.D {
		if d.ID == imdbID && d.Label != "" {
			return &IMDbTitle{ID: d.ID, Title: d.Label, Year: d.Year, Kind: d.Kind}, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrIMDbTitleNotFound, imdbID)
}

// YearString is the title's year as text, or "" when IMDb gave none.
func (t *IMDbTitle) YearString() string {
	if t.Year <= 0 {
		return ""
	}
	return strconv.Itoa(t.Year)
}
