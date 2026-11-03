package suggest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
)

const (
	kitsuBase = "https://kitsu.io/api/edge"
	imdbBase  = "https://v3.sg.media-imdb.com"
	// maxBody caps a suggestion response; real ones are a few KiB.
	maxBody = 1 << 20
)

// Doer is the HTTP client the providers use; *http.Client satisfies it.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewHTTPClient is the SSRF-guarded client the default providers share.
func NewHTTPClient() *http.Client {
	return &http.Client{Timeout: 5 * time.Second, Transport: netx.SafeScraperTransport(5 * time.Second)}
}

// getJSON GETs rawURL and decodes the capped JSON body into out.
func getJSON(ctx context.Context, hc Doer, rawURL, accept string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("User-Agent", netx.APIUserAgent)
	resp, err := hc.Do(req) // #nosec G704 -- fixed provider hosts, query URL-escaped
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	if err := jsonx.Decode(resp.Body, maxBody, out); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

// yearOf reads the leading four-digit year of a date such as "2002-10-03".
func yearOf(date string) int {
	if len(date) < 4 {
		return 0
	}
	year := 0
	for _, r := range date[:4] {
		if r < '0' || r > '9' {
			return 0
		}
		year = year*10 + int(r-'0')
	}
	return year
}

// Kitsu suggests anime. Its text filter matches partial words, which makes it
// fit for as-you-type use (AniList only matches whole words).
type Kitsu struct {
	hc   Doer
	base string
}

// NewKitsu builds the Kitsu provider on hc.
func NewKitsu(hc Doer) *Kitsu { return &Kitsu{hc: hc, base: kitsuBase} }

// Name implements Provider.
func (*Kitsu) Name() string { return "kitsu" }

// Kinds implements Provider.
func (*Kitsu) Kinds() []Kind { return []Kind{KindAnime} }

type kitsuReply struct {
	Data []struct {
		Attributes struct {
			CanonicalTitle string            `json:"canonicalTitle"`
			Titles         map[string]string `json:"titles"`
			StartDate      string            `json:"startDate"`
		} `json:"attributes"`
	} `json:"data"`
}

// Suggest implements Provider.
func (k *Kitsu) Suggest(ctx context.Context, query string, limit int) ([]Suggestion, error) {
	params := url.Values{}
	params.Set("filter[text]", query)
	params.Set("page[limit]", fmt.Sprint(limit))
	params.Set("fields[anime]", "canonicalTitle,titles,startDate")
	var reply kitsuReply
	if err := getJSON(ctx, k.hc, k.base+"/anime?"+params.Encode(), "application/vnd.api+json", &reply); err != nil {
		return nil, err
	}
	out := make([]Suggestion, 0, len(reply.Data))
	for _, d := range reply.Data {
		a := d.Attributes
		title := strings.TrimSpace(a.CanonicalTitle)
		if title == "" {
			title = strings.TrimSpace(a.Titles["en"])
		}
		if title == "" {
			title = strings.TrimSpace(a.Titles["en_jp"])
		}
		if title == "" {
			continue
		}
		out = append(out, Suggestion{Title: title, Year: yearOf(a.StartDate), Kind: KindAnime, Source: k.Name()})
	}
	return out, nil
}

// IMDb suggests movies and series from IMDb's public suggestion feed.
type IMDb struct {
	hc   Doer
	base string
}

// NewIMDb builds the IMDb provider on hc.
func NewIMDb(hc Doer) *IMDb { return &IMDb{hc: hc, base: imdbBase} }

// Name implements Provider.
func (*IMDb) Name() string { return "imdb" }

// Kinds implements Provider.
func (*IMDb) Kinds() []Kind { return []Kind{KindMovie, KindSeries} }

type imdbReply struct {
	D []struct {
		ID    string `json:"id"`
		Title string `json:"l"`
		Type  string `json:"qid"`
		Year  int    `json:"y"`
	} `json:"d"`
}

// imdbKind maps IMDb's type id to a Kind; ok is false for people, games,
// episodes and the other rows a title search does not want.
func imdbKind(qid string) (Kind, bool) {
	switch qid {
	case "movie", "tvMovie":
		return KindMovie, true
	case "tvSeries", "tvMiniSeries":
		return KindSeries, true
	default:
		return "", false
	}
}

// Suggest implements Provider. The feed shards by the query's first character
// and wants letters, digits and spaces, so the query is folded first.
func (i *IMDb) Suggest(ctx context.Context, query string, limit int) ([]Suggestion, error) {
	folded := fold(query)
	if folded == "" {
		return nil, nil
	}
	first, _ := utf8.DecodeRuneInString(folded)
	shard := string(first)
	var reply imdbReply
	rawURL := i.base + "/suggestion/" + url.PathEscape(shard) + "/" + url.PathEscape(folded) + ".json"
	if err := getJSON(ctx, i.hc, rawURL, "application/json", &reply); err != nil {
		return nil, err
	}
	out := make([]Suggestion, 0, len(reply.D))
	for _, d := range reply.D {
		kind, ok := imdbKind(d.Type)
		title := strings.TrimSpace(d.Title)
		if !ok || title == "" || !strings.HasPrefix(d.ID, "tt") {
			continue
		}
		out = append(out, Suggestion{Title: title, Year: d.Year, Kind: kind, Source: i.Name()})
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

// Default returns the process-wide engine: Kitsu for anime, IMDb for movies
// and series, with no key or setup.
func Default() *Engine {
	defaultOnce.Do(func() {
		hc := NewHTTPClient()
		defaultEngine = New(Options{Providers: []Provider{NewKitsu(hc), NewIMDb(hc)}})
	})
	return defaultEngine
}
