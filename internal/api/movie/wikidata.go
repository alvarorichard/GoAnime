// Package movie — Wikidata as a keyless bridge between TMDB and IMDb ids.
package movie

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// Wikidata records a work's TMDB id (P4947 for movies, P4983 for TV series)
// next to its IMDb id (P345), English label and dates, and answers SPARQL
// without an API key. That is what lets a title known by only one of the two
// ids — StartFlix names shows by TMDB id and movies by IMDb id — reach the
// other: IMDb's title lookup is keyed by IMDb id, and Plex/Jellyfin folders
// carry both.
const WikidataSPARQLURL = "https://query.wikidata.org/sparql"

var (
	imdbIDRe = regexp.MustCompile(`^tt\d{7,10}$`)

	// ErrNoWikidataLink means Wikidata has no item linking the two ids.
	ErrNoWikidataLink = errors.New("wikidata has no TMDB/IMDb link for this title")
)

// WikidataClient queries Wikidata's SPARQL endpoint.
type WikidataClient struct {
	client   *http.Client
	endpoint string
}

// NewWikidataClient creates a client on the SSRF-guarded movie transport.
func NewWikidataClient() *WikidataClient {
	return &WikidataClient{
		client: &http.Client{
			Timeout:   15 * time.Second,
			Transport: safeMovieTransport(15 * time.Second),
		},
		endpoint: WikidataSPARQLURL,
	}
}

func wikidataTMDBProperty(movie bool) string {
	if movie {
		return "P4947"
	}
	return "P4983"
}

// WikidataRecord is what Wikidata files about a movie or show.
type WikidataRecord struct {
	IMDBID string
	TMDBID int
	Label  string // English label
	Year   string // earliest publication (or, for a show, start) year
}

// Lookup finds the work with this IMDb id or, failing that, this TMDB id, and
// returns both of its ids with its English label and year in one query.
func (c *WikidataClient) Lookup(tmdbID int, imdbID string, movie bool) (*WikidataRecord, error) {
	tmdbProp := wikidataTMDBProperty(movie)
	var anchor string
	switch {
	case imdbID != "":
		if !imdbIDRe.MatchString(imdbID) {
			return nil, fmt.Errorf("invalid IMDb id %q", imdbID)
		}
		anchor = `wdt:P345 "` + imdbID + `"`
	case tmdbID > 0:
		anchor = `wdt:` + tmdbProp + ` "` + strconv.Itoa(tmdbID) + `"`
	default:
		return nil, errors.New("no id to look up")
	}
	query := `SELECT ?imdb ?tmdb ?label ?date WHERE { ?item ` + anchor + ` .` +
		` OPTIONAL { ?item wdt:P345 ?imdb }` +
		` OPTIONAL { ?item wdt:` + tmdbProp + ` ?tmdb }` +
		` OPTIONAL { ?item rdfs:label ?label FILTER(LANG(?label) = "en") }` +
		` OPTIONAL { ?item wdt:P577|wdt:P580 ?date } } LIMIT 50`
	rows, err := c.selectRows(query)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrNoWikidataLink
	}

	rec := &WikidataRecord{}
	for _, row := range rows {
		if v := row["imdb"]; rec.IMDBID == "" && imdbIDRe.MatchString(v) {
			rec.IMDBID = v
		}
		if id, err := strconv.Atoi(row["tmdb"]); rec.TMDBID == 0 && err == nil && id > 0 {
			rec.TMDBID = id
		}
		if rec.Label == "" {
			rec.Label = row["label"]
		}
		if y := leadingYear(row["date"]); y != "" && (rec.Year == "" || y < rec.Year) {
			rec.Year = y
		}
	}
	return rec, nil
}

// selectRows runs a query and returns each result row as variable -> value.
func (c *WikidataClient) selectRows(query string) ([]map[string]string, error) {
	endpoint := c.endpoint + "?" + url.Values{"query": {query}, "format": {"json"}}.Encode()
	req, err := http.NewRequest(http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/sparql-results+json")
	// Wikimedia's policy asks API clients to name themselves; requests with
	// a generic or browser User-Agent may be throttled or refused.
	req.Header.Set("User-Agent", netx.APIUserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wikidata returned status: %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var result struct {
		Results struct {
			Bindings []map[string]struct {
				Value string `json:"value"`
			} `json:"bindings"`
		} `json:"results"`
	}
	if err := jsonx.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse wikidata response: %w", err)
	}
	rows := make([]map[string]string, 0, len(result.Results.Bindings))
	for _, b := range result.Results.Bindings {
		row := make(map[string]string, len(b))
		for k, v := range b {
			row[k] = v.Value
		}
		rows = append(rows, row)
	}
	return rows, nil
}
