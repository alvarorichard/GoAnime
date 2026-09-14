package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// Jikan is the unofficial MyAnimeList API, used here as the stand-in for
// AniList when AniList has switched its own API off (see ErrAniListAPIDisabled).
//
// It covers everything enrichment actually consumes — MAL id, the three title
// forms, a cover image and synonyms — with one exception: there is no AniList
// id, because Jikan indexes MyAnimeList. That is already safe. Both readers of
// Anime.AnilistID guard on `> 0` (the Discord Rich Presence link and the
// metadata enricher), so leaving it zero drops the "view on AniList" button and
// changes nothing else.

// jikanSearchLimit keeps the response small: enrichment only ever reads the
// best match.
const jikanSearchLimit = 1

// Jikan fails in bursts. Measured 2026-09-09 on one title: a run of 504s across
// several minutes, then 12 consecutive 200s with no change on this side. A
// couple of quick retries turn that kind of blip into a normal answer, and cost
// nothing when the first attempt already worked.
//
// Only transient conditions are retried — a 5xx or a connection error. A 404 or
// an empty result set is a real answer and repeating it just adds latency to a
// path the user is waiting on.
// The backoff sits above Jikan's published rate limit (3 requests/second).
// A shorter one was self-defeating: retrying a 504 after 400ms produced 429 Too
// Many Requests, so the retry manufactured the failure it was meant to absorb.
//
// Two attempts, not more. Enrichment runs before playback starts, so the user
// waits for it; one extra second is a fair price for rescuing a blip, three is
// not — and metadata is best-effort, so giving up is a supported outcome.
const (
	jikanAttempts = 2
	jikanBackoff  = 1100 * time.Millisecond
)

// transientJikanStatus reports whether a status is worth another attempt.
func transientJikanStatus(code int) bool {
	return code >= 500 || code == http.StatusTooManyRequests
}

// jikanSearchResponse is the slice of Jikan's /anime payload this package needs.
type jikanSearchResponse struct {
	Data []struct {
		MalID         int      `json:"mal_id"`
		Title         string   `json:"title"`
		TitleEnglish  string   `json:"title_english"`
		TitleJapanese string   `json:"title_japanese"`
		Synonyms      []string `json:"title_synonyms"`
		Synopsis      string   `json:"synopsis"`
		Status        string   `json:"status"`
		Episodes      int      `json:"episodes"`
		Score         float64  `json:"score"`
		Genres        []struct {
			Name string `json:"name"`
		} `json:"genres"`
		Images struct {
			JPG struct {
				LargeImageURL string `json:"large_image_url"`
				ImageURL      string `json:"image_url"`
			} `json:"jpg"`
		} `json:"images"`
	} `json:"data"`
}

// jikanSearch performs the request, retrying the transient failures Jikan is
// prone to. The last error is returned once the attempts run out.
func jikanSearch(endpoint string) (*jikanSearchResponse, error) {
	var lastErr error
	for attempt := 1; attempt <= jikanAttempts; attempt++ {
		if attempt > 1 {
			time.Sleep(jikanBackoff)
		}

		out, retry, err := jikanAttempt(endpoint)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
		util.Debugf("Jikan attempt %d/%d failed (%v); retrying", attempt, jikanAttempts, err)
	}
	return nil, lastErr
}

// jikanAttempt is one request. Split out so the response body is closed by a
// plain defer in a scope of its own: closing it across the branches of a retry
// loop is easy to get wrong on a later edit, and hides the guarantee from
// linters too.
//
// retry reports whether the failure is worth another attempt.
func jikanAttempt(endpoint string) (out *jikanSearchResponse, retry bool, err error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, false, fmt.Errorf("jikan: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", netx.APIUserAgent)

	resp, err := aniListClient.Do(req) // #nosec G704 -- endpoint is built from the package's own base URL
	if err != nil {
		// A connection-level failure is exactly the kind of blip worth retrying.
		return nil, true, fmt.Errorf("jikan request failed: %w", err)
	}
	defer safeClose(resp.Body, "Jikan response body")

	if resp.StatusCode != http.StatusOK {
		return nil, transientJikanStatus(resp.StatusCode), fmt.Errorf("jikan returned: %s", resp.Status)
	}

	var decoded jikanSearchResponse
	if err := jsonx.Decode(resp.Body, maxJSONResponseBytes, &decoded); err != nil {
		return nil, false, fmt.Errorf("jikan decode failed: %w", err)
	}
	return &decoded, false, nil
}

// fetchAnimeFromJikan looks a title up on MyAnimeList and shapes the result
// like an AniList response, so callers keep one code path.
func fetchAnimeFromJikan(animeName string) (*models.AniListResponse, error) {
	cleaned := CleanTitle(animeName)
	if strings.TrimSpace(cleaned) == "" {
		return nil, fmt.Errorf("jikan: empty search title")
	}

	endpoint := fmt.Sprintf("%s/anime?q=%s&limit=%d",
		strings.TrimSuffix(jikanBaseURL, "/"), url.QueryEscape(cleaned), jikanSearchLimit)

	out, err := jikanSearch(endpoint)
	if err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, fmt.Errorf("no matching anime found on MyAnimeList for %q", cleaned)
	}

	m := out.Data[0]
	genres := make([]string, 0, len(m.Genres))
	for _, g := range m.Genres {
		genres = append(genres, g.Name)
	}
	cover := m.Images.JPG.LargeImageURL
	if cover == "" {
		cover = m.Images.JPG.ImageURL
	}

	var result models.AniListResponse
	result.Data.Media = models.AniListDetails{
		// ID stays 0: this is a MyAnimeList record and has no AniList id.
		IDMal: m.MalID,
		Title: models.Title{
			Romaji:  m.Title,
			English: m.TitleEnglish,
			Native:  m.TitleJapanese,
		},
		Description: m.Synopsis,
		Genres:      genres,
		// Jikan scores out of 10, AniList out of 100.
		AverageScore: int(m.Score * 10),
		Episodes:     m.Episodes,
		Status:       strings.ToUpper(m.Status),
		CoverImage:   models.CoverImages{Large: cover},
		Synonyms:     m.Synonyms,
	}

	util.Debugf("Jikan found: MAL=%d, Title=%s (search: %q)", m.MalID, m.Title, cleaned)
	return &result, nil
}
