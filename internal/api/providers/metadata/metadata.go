// Package metadata provides unified metadata enrichment for all media sources.
// It queries AniList (for anime) and TMDB (for movies/TV) to get canonical titles,
// years, and season/episode mappings needed for Jellyfin/Plex-compatible naming.
package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// maxJSONResponseBytes caps how much of an HTTP response jsonx.Decode will read
// before failing with jsonx.ErrTooLarge. The decoders this replaced
// (json.NewDecoder(resp.Body)) had no bound at all, so a hostile or broken
// upstream could stream until the process ran out of memory.
const maxJSONResponseBytes = 10 << 20 // 10 MiB

// setAniListHeaders applies the headers AniList expects from an API client.
//
// The User-Agent must NOT look like a browser: AniList answers browser UAs with
// an HTTP 403 ("The AniList API has been temporarily disabled due to severe
// stability issues") while serving plain API clients normally. This only holds if
// the request also goes through Enricher.aniListClient — the shared surf client
// would overwrite the UA with Chrome's. See issue #184.
func setAniListHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", netx.APIUserAgent)
}

// allowedAPIHosts lists the hosts that metadata HTTP requests may target.
// Any dynamically-built URL is validated against this set before a request
// is created, preventing SSRF via tainted input (gosec G704 / CWE-918).
var allowedAPIHosts = map[string]bool{
	"api.themoviedb.org": true,
}

// safeNewRequest builds an *http.Request after validating that the URL
// targets one of the allowedAPIHosts. Returns an error if the host is
// not in the allowlist, mitigating SSRF (gosec G704).
func safeNewRequest(ctx context.Context, method, rawURL string, body io.Reader) (*http.Request, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if !allowedAPIHosts[parsed.Hostname()] {
		return nil, errors.New("disallowed host: " + parsed.Hostname())
	}
	return http.NewRequestWithContext(ctx, method, parsed.String(), body) // #nosec G704 -- host validated above
}

// HTTPClient is the interface for HTTP requests, allowing test mocks.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Enricher populates anime metadata from external APIs.
type Enricher struct {
	client HTTPClient
	// aniListClient is used ONLY for AniList. It must stay a plain client: the
	// shared client impersonates Chrome and rewrites the User-Agent to a browser
	// one, which AniList answers with a 403 (see setAniListHeaders, issue #184).
	aniListClient HTTPClient
	timeout       time.Duration
}

// NewEnricher creates a metadata Enricher with the default HTTP client.
func NewEnricher() *Enricher {
	return &Enricher{
		client:        util.GetSharedClient(),
		aniListClient: &http.Client{Timeout: 20 * time.Second},
		timeout:       10 * time.Second,
	}
}

// NewEnricherWithClient creates an Enricher with a custom HTTP client (for testing).
// The injected client serves AniList too, so tests keep full control of every call.
func NewEnricherWithClient(client HTTPClient) *Enricher {
	return &Enricher{
		client:        client,
		aniListClient: client,
		timeout:       10 * time.Second,
	}
}

// AnimeMetadata contains the enriched metadata for naming purposes.
type AnimeMetadata struct {
	// TitleEnglish is the English title (preferred for Jellyfin/Plex).
	TitleEnglish string

	// TitleRomaji is the romanized Japanese title.
	TitleRomaji string

	// Year is the first air year.
	Year string

	// TotalEpisodes is the total episode count (0 if ongoing/unknown).
	TotalEpisodes int

	// AniListID is the AniList database ID.
	AniListID int

	// MalID is the MyAnimeList database ID.
	MalID int

	// IMDBID is the IMDB ID (from TMDB cross-reference).
	IMDBID string

	// Season mappings: for long-running anime, maps absolute episode numbers
	// to season/episode pairs. Nil if not applicable (single-season anime).
	SeasonMap []SeasonMapping
}

// SeasonMapping maps a range of absolute episode numbers to a season.
type SeasonMapping struct {
	Season       int // Season number (1-based)
	StartEp      int // First absolute episode number in this season
	EndEp        int // Last absolute episode number in this season
	EpisodeCount int // Number of episodes in this season
}

// AbsoluteToSeason converts an absolute episode number to (season, episode) pair.
// Returns (1, absoluteEp) if no season mapping is available.
func (m *AnimeMetadata) AbsoluteToSeason(absoluteEp int) (season, episode int) {
	if len(m.SeasonMap) == 0 {
		return 1, absoluteEp
	}
	for _, sm := range m.SeasonMap {
		if absoluteEp >= sm.StartEp && absoluteEp <= sm.EndEp {
			return sm.Season, absoluteEp - sm.StartEp + 1
		}
	}
	// Episode beyond known range: put in last season
	last := m.SeasonMap[len(m.SeasonMap)-1]
	return last.Season, absoluteEp - last.StartEp + 1
}

// EnrichFromAniList fetches metadata from AniList by anime name.
func (e *Enricher) EnrichFromAniList(ctx context.Context, animeName string) (*AnimeMetadata, error) {
	cleanName := cleanSearchName(animeName)
	if cleanName == "" {
		return nil, fmt.Errorf("empty anime name after cleaning")
	}

	query := `query ($search: String) {
		Media(search: $search, type: ANIME) {
			id
			idMal
			title { romaji english native }
			startDate { year }
			episodes
			status
			relations {
				edges {
					relationType
					node {
						id
						title { romaji english }
						episodes
						format
						startDate { year month }
					}
				}
			}
		}
	}`

	body, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": map[string]any{"search": cleanName},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal AniList query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://graphql.anilist.co", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	setAniListHeaders(req)

	resp, err := e.aniListClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AniList request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AniList returned status %d", resp.StatusCode)
	}

	var result aniListResponse
	if err := jsonx.Decode(resp.Body, maxJSONResponseBytes, &result); err != nil {
		return nil, fmt.Errorf("decode AniList response: %w", err)
	}

	media := result.Data.Media
	if media.ID == 0 {
		return nil, fmt.Errorf("no AniList result for %q", cleanName)
	}

	meta := &AnimeMetadata{
		TitleEnglish:  media.Title.English,
		TitleRomaji:   media.Title.Romaji,
		TotalEpisodes: media.Episodes,
		AniListID:     media.ID,
		MalID:         media.IDMal,
	}

	if media.StartDate.Year > 0 {
		meta.Year = fmt.Sprintf("%d", media.StartDate.Year)
	}

	// Build season map from sequel relations
	meta.SeasonMap = buildSeasonMap(media)

	return meta, nil
}

// EnrichFromAniListByID fetches metadata from AniList by AniList ID.
func (e *Enricher) EnrichFromAniListByID(ctx context.Context, anilistID int) (*AnimeMetadata, error) {
	query := `query ($id: Int) {
		Media(id: $id, type: ANIME) {
			id
			idMal
			title { romaji english native }
			startDate { year }
			episodes
			status
			relations {
				edges {
					relationType
					node {
						id
						title { romaji english }
						episodes
						format
						startDate { year month }
					}
				}
			}
		}
	}`

	body, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": map[string]any{"id": anilistID},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://graphql.anilist.co", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	setAniListHeaders(req)

	resp, err := e.aniListClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AniList request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AniList returned status %d", resp.StatusCode)
	}

	var result aniListResponse
	if err := jsonx.Decode(resp.Body, maxJSONResponseBytes, &result); err != nil {
		return nil, err
	}

	media := result.Data.Media
	if media.ID == 0 {
		return nil, fmt.Errorf("no AniList result for ID %d", anilistID)
	}

	meta := &AnimeMetadata{
		TitleEnglish:  media.Title.English,
		TitleRomaji:   media.Title.Romaji,
		TotalEpisodes: media.Episodes,
		AniListID:     media.ID,
		MalID:         media.IDMal,
	}

	if media.StartDate.Year > 0 {
		meta.Year = fmt.Sprintf("%d", media.StartDate.Year)
	}

	meta.SeasonMap = buildSeasonMap(media)

	return meta, nil
}

// LookupIMDBID tries to find the IMDB ID for an anime using TMDB's find endpoint.
// Requires a TMDB API key. Returns "" if not found.
func (e *Enricher) LookupIMDBID(ctx context.Context, malID int, tmdbAPIKey string) (string, error) {
	if malID <= 0 || tmdbAPIKey == "" {
		return "", nil
	}

	// TMDB find by MAL ID (external_source = myanimelist)
	findURL := fmt.Sprintf("https://api.themoviedb.org/3/find/mal-%d?api_key=%s&external_source=myanimelist",
		malID, url.QueryEscape(tmdbAPIKey))

	req, err := http.NewRequestWithContext(ctx, "GET", findURL, http.NoBody)
	if err != nil {
		return "", err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("TMDB find request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil // not found is not an error
	}

	var findResult struct {
		TVResults []struct {
			ID int `json:"id"`
		} `json:"tv_results"`
	}
	if err := jsonx.Decode(resp.Body, maxJSONResponseBytes, &findResult); err != nil {
		return "", nil
	}

	if len(findResult.TVResults) == 0 {
		return "", nil
	}

	// Get the IMDB ID from the TMDB TV details
	tvID := findResult.TVResults[0].ID
	detailsURL := fmt.Sprintf("https://api.themoviedb.org/3/tv/%d/external_ids?api_key=%s",
		tvID, url.QueryEscape(tmdbAPIKey))

	req2, err := http.NewRequestWithContext(ctx, "GET", detailsURL, http.NoBody)
	if err != nil {
		return "", err
	}

	resp2, err := e.client.Do(req2)
	if err != nil {
		return "", nil
	}
	defer resp2.Body.Close()

	var extIDs struct {
		IMDBID string `json:"imdb_id"`
	}
	if err := jsonx.Decode(resp2.Body, maxJSONResponseBytes, &extIDs); err != nil {
		return "", nil
	}

	return extIDs.IMDBID, nil
}

// ApplyToAnime enriches a models.Anime with metadata from AniList.
// This is the main integration point between metadata enrichment and the existing model.
func (e *Enricher) ApplyToAnime(ctx context.Context, anime *models.Anime) error {
	_, err := e.EnrichAnime(ctx, anime)
	return err
}

// EnrichAnime enriches the anime model with AniList metadata and returns the
// season map (if any). Callers that need per-episode season resolution should
// use the returned SeasonMap with player.SetSeasonMap.
func (e *Enricher) EnrichAnime(ctx context.Context, anime *models.Anime) ([]SeasonMapping, error) {
	if anime == nil {
		return nil, nil
	}

	util.Debug("EnrichAnime called", "name", anime.Name, "anilistID", anime.AnilistID, "malID", anime.MalID)

	var meta *AnimeMetadata
	var err error

	if anime.AnilistID > 0 {
		meta, err = e.EnrichFromAniListByID(ctx, anime.AnilistID)
	} else {
		meta, err = e.EnrichFromAniList(ctx, anime.Name)
	}

	if err != nil {
		util.Debug("metadata enrichment failed, continuing with scraped data", "error", err)
		return nil, nil // non-fatal: scraped data is still usable
	}

	util.Debug("AniList enrichment result",
		"titleEN", meta.TitleEnglish, "anilistID", meta.AniListID,
		"malID", meta.MalID, "totalEps", meta.TotalEpisodes,
		"seasonMapLen", len(meta.SeasonMap))
	if meta.TitleEnglish != "" && anime.Details.Title.English == "" {
		anime.Details.Title.English = meta.TitleEnglish
	}
	if meta.TitleRomaji != "" && anime.Details.Title.Romaji == "" {
		anime.Details.Title.Romaji = meta.TitleRomaji
	}
	if meta.Year != "" && anime.Year == "" {
		anime.Year = meta.Year
	}
	if meta.AniListID > 0 && anime.AnilistID == 0 {
		anime.AnilistID = meta.AniListID
	}
	if meta.MalID > 0 && anime.MalID == 0 {
		anime.MalID = meta.MalID
	}
	if meta.TotalEpisodes > 0 && anime.Details.Episodes == 0 {
		anime.Details.Episodes = meta.TotalEpisodes
	}
	if anime.CurrentSeason <= 0 {
		if season := inferSeasonNumber(anime.Name, meta.TitleEnglish, meta.TitleRomaji); season > 1 {
			anime.CurrentSeason = season
			util.Debug("inferred anime season from title", "season", season, "anime", anime.Name)
		}
	}

	seasonMap := meta.SeasonMap

	// If AniList returned no useful season map (no sequels), try external
	// sources for a proper season breakdown. Many long-running anime
	// (Black Clover, Naruto, etc.) are single entries on AniList but TMDB
	// splits them into proper seasons.
	if len(seasonMap) == 0 && meta.TotalEpisodes > 13 {
		// Try 1: TMDB API (if key is configured)
		if tmdbAPIKey := os.Getenv("TMDB_API_KEY"); tmdbAPIKey != "" && meta.MalID > 0 {
			if tmdbMap := e.buildSeasonMapFromTMDB(ctx, meta.MalID, tmdbAPIKey); len(tmdbMap) > 1 {
				util.Debug("using TMDB API season map",
					"seasons", len(tmdbMap), "malID", meta.MalID)
				seasonMap = tmdbMap
			}
		}

		// Try 2: StartFlix (its panel is keyed by TMDB id; no API key needed)
		if len(seasonMap) == 0 {
			if sfxMap := e.buildSeasonMapFromStartFlix(ctx, anime.Name); len(sfxMap) > 1 {
				util.Debug("using StartFlix season map",
					"seasons", len(sfxMap), "anime", anime.Name)
				seasonMap = sfxMap
			}
		}
	}

	util.Debug("EnrichAnime final result", "seasonMapLen", len(seasonMap), "anime", anime.Name)
	return seasonMap, nil
}

func inferSeasonNumber(titles ...string) int {
	for _, title := range titles {
		if season := inferSeasonNumberFromTitle(title); season > 1 {
			return season
		}
	}
	return 0
}

// seasonNumberPatterns match "Season N" / "Nth Season" markers in titles.
var seasonNumberPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bseason\s+(\d+)\b`),
	regexp.MustCompile(`(?i)\b(\d+)\s*(?:st|nd|rd|th)?\s+season\b`),
}

func inferSeasonNumberFromTitle(title string) int {
	title = strings.TrimSpace(title)
	if title == "" {
		return 0
	}

	for _, pattern := range seasonNumberPatterns {
		matches := pattern.FindStringSubmatch(title)
		if len(matches) < 2 {
			continue
		}
		season, err := strconv.Atoi(matches[1])
		if err == nil && season > 1 {
			return season
		}
	}
	return 0
}

// --- Internal types ---

type aniListResponse struct {
	Data struct {
		Media aniListMedia `json:"Media"`
	} `json:"data"`
}

type aniListMedia struct {
	ID    int `json:"id"`
	IDMal int `json:"idMal"`
	Title struct {
		Romaji  string `json:"romaji"`
		English string `json:"english"`
		Native  string `json:"native"`
	} `json:"title"`
	StartDate struct {
		Year  int `json:"year"`
		Month int `json:"month"`
	} `json:"startDate"`
	Episodes  int    `json:"episodes"`
	Status    string `json:"status"`
	Relations struct {
		Edges []struct {
			RelationType string `json:"relationType"`
			Node         struct {
				ID    int `json:"id"`
				Title struct {
					Romaji  string `json:"romaji"`
					English string `json:"english"`
				} `json:"title"`
				Episodes  int    `json:"episodes"`
				Format    string `json:"format"`
				StartDate struct {
					Year  int `json:"year"`
					Month int `json:"month"`
				} `json:"startDate"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"relations"`
}

// buildSeasonMap constructs season mappings from AniList relations.
// Season 1 is the queried anime; sequels become season 2, 3, etc.
// buildSeasonMapFromTMDB fetches seasons from TMDB and constructs a season map.
// It uses the MAL ID to find the corresponding TMDB TV entry, then reads the
// episode_count of each season to build cumulative episode ranges.
func (e *Enricher) buildSeasonMapFromTMDB(ctx context.Context, malID int, tmdbAPIKey string) []SeasonMapping {
	// Step 1: Find TMDB TV ID via MAL external ID
	findURL := fmt.Sprintf("https://api.themoviedb.org/3/find/mal-%d?api_key=%s&external_source=myanimelist",
		malID, url.QueryEscape(tmdbAPIKey))

	req, err := safeNewRequest(ctx, "GET", findURL, nil)
	if err != nil {
		return nil
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var findResult struct {
		TVResults []struct {
			ID int `json:"id"`
		} `json:"tv_results"`
	}
	if err := jsonx.Decode(resp.Body, maxJSONResponseBytes, &findResult); err != nil || len(findResult.TVResults) == 0 {
		return nil
	}

	tvID := findResult.TVResults[0].ID

	// Step 2: Get TV details with seasons
	detailsURL := fmt.Sprintf("https://api.themoviedb.org/3/tv/%d?api_key=%s&language=en-US",
		tvID, url.QueryEscape(tmdbAPIKey))

	req2, err := safeNewRequest(ctx, "GET", detailsURL, nil)
	if err != nil {
		return nil
	}

	resp2, err := e.client.Do(req2)
	if err != nil {
		return nil
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		return nil
	}

	var tvDetails struct {
		Seasons []struct {
			SeasonNumber int `json:"season_number"`
			EpisodeCount int `json:"episode_count"`
		} `json:"seasons"`
	}
	if err := jsonx.Decode(resp2.Body, maxJSONResponseBytes, &tvDetails); err != nil {
		return nil
	}

	// Build season map from TMDB seasons, skipping season 0 (specials)
	var seasons []SeasonMapping
	currentEp := 1
	for _, s := range tvDetails.Seasons {
		if s.SeasonNumber < 1 || s.EpisodeCount <= 0 {
			continue
		}
		seasons = append(seasons, SeasonMapping{
			Season:       s.SeasonNumber,
			StartEp:      currentEp,
			EndEp:        currentEp + s.EpisodeCount - 1,
			EpisodeCount: s.EpisodeCount,
		})
		currentEp += s.EpisodeCount
	}

	if len(seasons) <= 1 {
		return nil // Not useful if TMDB also has only 1 season
	}

	return seasons
}

func buildSeasonMap(media aniListMedia) []SeasonMapping {
	if media.Episodes <= 0 {
		return nil
	}

	seasons := []SeasonMapping{
		{Season: 1, StartEp: 1, EndEp: media.Episodes, EpisodeCount: media.Episodes},
	}

	// Collect SEQUEL relations with known episode counts
	type sequel struct {
		episodes int
		year     int
		month    int
	}
	var sequels []sequel

	for _, edge := range media.Relations.Edges {
		if edge.RelationType != "SEQUEL" {
			continue
		}
		node := edge.Node
		if node.Episodes <= 0 || node.Format == "SPECIAL" || node.Format == "OVA" {
			continue
		}
		sequels = append(sequels, sequel{
			episodes: node.Episodes,
			year:     node.StartDate.Year,
			month:    node.StartDate.Month,
		})
	}

	if len(sequels) == 0 {
		return nil // No sequels → no useful season map; let TMDB/StartFlix provide one
	}

	// Sort sequels chronologically (simple: by year then month)
	for i := 0; i < len(sequels); i++ {
		for j := i + 1; j < len(sequels); j++ {
			if sequels[j].year < sequels[i].year ||
				(sequels[j].year == sequels[i].year && sequels[j].month < sequels[i].month) {
				sequels[i], sequels[j] = sequels[j], sequels[i]
			}
		}
	}

	// Build cumulative season map
	currentEp := media.Episodes + 1
	for i, seq := range sequels {
		seasons = append(seasons, SeasonMapping{
			Season:       i + 2,
			StartEp:      currentEp,
			EndEp:        currentEp + seq.episodes - 1,
			EpisodeCount: seq.episodes,
		})
		currentEp += seq.episodes
	}

	return seasons
}

// cleanSearchName removes source tags and normalizes an anime name for API search.
func cleanSearchName(name string) string {
	// Remove square-bracket tags like [English], [PT-BR], [AnimeFire], etc.
	tagStart := strings.Index(name, "[")
	for tagStart >= 0 {
		tagEnd := strings.Index(name[tagStart:], "]")
		if tagEnd < 0 {
			break
		}
		name = name[:tagStart] + name[tagStart+tagEnd+1:]
		tagStart = strings.Index(name, "[")
	}
	// Remove parenthetical tags common in PT-BR sources:
	// (Dublado), (Legendado), (Sub), (Dub), etc.
	name = reParenTag.ReplaceAllString(name, "")
	return strings.TrimSpace(name)
}

// reParenTag matches parenthetical tags that should be stripped from anime names.
var reParenTag = regexp.MustCompile(`\s*\((?i:dublado|legendado|sub|dub|dual[- ]?audio|completo|todos os epis[oó]dios)\)`)

// --- StartFlix-based season lookup (no API key needed) ---
//
// StartFlix's video panel is keyed by TMDB id and lists each season's
// episodes, which is exactly the season structure TMDB gives, without a key.

// enricherTransport lets a StartFlix client send its requests through the
// enricher's own HTTPClient, so production uses the shared client and tests
// serve the pages from the same mock as every other lookup here.
type enricherTransport struct{ c HTTPClient }

func (t enricherTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.c.Do(req)
}

// buildSeasonMapFromStartFlix searches StartFlix for an anime by name, opens
// the matching series' panel, and turns its per-season episode lists into a
// season map.
func (e *Enricher) buildSeasonMapFromStartFlix(ctx context.Context, animeName string) []SeasonMapping {
	cleanName := cleanSearchName(animeName)
	if cleanName == "" {
		return nil
	}
	c := startflix.NewClientWithHTTP(&http.Client{Transport: enricherTransport{e.client}})

	results, err := c.Search(ctx, cleanName)
	if err != nil {
		util.Debug("StartFlix search failed", "error", err)
		return nil
	}
	target, ok := matchStartFlixSeries(results, cleanName)
	if !ok {
		util.Debug("StartFlix: no series matches", "query", cleanName)
		return nil
	}
	panel, err := c.Panel(ctx, target.URL)
	if err != nil {
		util.Debug("StartFlix: no panel for the series", "url", target.URL, "error", err)
		return nil
	}
	seasons, err := c.Seasons(ctx, panel)
	if err != nil {
		util.Debug("StartFlix: no seasons on the panel", "panel", panel.URL, "error", err)
		return nil
	}

	result := seasonMapFromStartFlix(seasons)
	if len(result) <= 1 {
		return nil
	}
	util.Debug("StartFlix season map built", "seasons", len(result), "tmdbID", panel.TMDBID)
	return result
}

// seasonMapFromStartFlix builds cumulative episode ranges. A season's size is
// its highest episode number across both audio lists: the dub and the
// subtitled list are often at different points in the same season.
func seasonMapFromStartFlix(seasons []startflix.Season) []SeasonMapping {
	var result []SeasonMapping
	currentEp := 1
	for _, s := range seasons {
		if s.Number < 1 {
			continue
		}
		count := 0
		for _, list := range [][]startflix.Episode{s.Dubbed, s.Subtitled} {
			for _, ep := range list {
				count = max(count, ep.Number)
			}
		}
		if count <= 0 {
			continue
		}
		result = append(result, SeasonMapping{
			Season:       s.Number,
			StartEp:      currentEp,
			EndEp:        currentEp + count - 1,
			EpisodeCount: count,
		})
		currentEp += count
	}
	return result
}

// matchStartFlixSeries picks the series whose title best matches the search:
// exact, then the search containing the title, then the title containing the
// search, then the first series. Movies never match.
func matchStartFlixSeries(results []startflix.Media, searchName string) (startflix.Media, bool) {
	var series []startflix.Media
	for _, r := range results {
		if r.Kind == startflix.KindSeries {
			series = append(series, r)
		}
	}
	if len(series) == 0 {
		return startflix.Media{}, false
	}
	search := strings.ToLower(strings.TrimSpace(searchName))
	rules := []func(title string) bool{
		func(t string) bool { return t == search },
		func(t string) bool { return t != "" && strings.Contains(search, t) },
		func(t string) bool { return strings.Contains(t, search) },
	}
	for _, rule := range rules {
		for _, s := range series {
			if rule(strings.ToLower(strings.TrimSpace(s.Title))) {
				return s, true
			}
		}
	}
	return series[0], true
}
