// Package metadata provides unified metadata enrichment for all media sources.
// It queries AniList (for anime) and TMDB (for movies/TV) to get canonical titles,
// years, and season/episode mappings needed for Jellyfin/Plex-compatible naming.
package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// HTTPClient is the interface for HTTP requests, allowing test mocks.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Enricher populates anime metadata from external APIs.
type Enricher struct {
	client  HTTPClient
	timeout time.Duration
}

// NewEnricher creates a metadata Enricher with the default HTTP client.
func NewEnricher() *Enricher {
	return &Enricher{
		client:  util.GetSharedClient(),
		timeout: 10 * time.Second,
	}
}

// NewEnricherWithClient creates an Enricher with a custom HTTP client (for testing).
func NewEnricherWithClient(client HTTPClient) *Enricher {
	return &Enricher{
		client:  client,
		timeout: 10 * time.Second,
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AniList request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AniList returned status %d", resp.StatusCode)
	}

	var result aniListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AniList request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AniList returned status %d", resp.StatusCode)
	}

	var result aniListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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

	req, err := http.NewRequestWithContext(ctx, "GET", findURL, nil)
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
	if err := json.NewDecoder(resp.Body).Decode(&findResult); err != nil {
		return "", nil
	}

	if len(findResult.TVResults) == 0 {
		return "", nil
	}

	// Get the IMDB ID from the TMDB TV details
	tvID := findResult.TVResults[0].ID
	detailsURL := fmt.Sprintf("https://api.themoviedb.org/3/tv/%d/external_ids?api_key=%s",
		tvID, url.QueryEscape(tmdbAPIKey))

	req2, err := http.NewRequestWithContext(ctx, "GET", detailsURL, nil)
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
	if err := json.NewDecoder(resp2.Body).Decode(&extIDs); err != nil {
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

		// Try 2: SuperFlix (scrapes TMDB data, no API key needed)
		if len(seasonMap) == 0 {
			if sfMap := e.buildSeasonMapFromSuperFlix(ctx, anime.Name); len(sfMap) > 1 {
				util.Debug("using SuperFlix season map",
					"seasons", len(sfMap), "anime", anime.Name)
				seasonMap = sfMap
			}
		}
	}

	util.Debug("EnrichAnime final result", "seasonMapLen", len(seasonMap), "anime", anime.Name)
	return seasonMap, nil
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

	req, err := http.NewRequestWithContext(ctx, "GET", findURL, nil)
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
	if err := json.NewDecoder(resp.Body).Decode(&findResult); err != nil || len(findResult.TVResults) == 0 {
		return nil
	}

	tvID := findResult.TVResults[0].ID

	// Step 2: Get TV details with seasons
	detailsURL := fmt.Sprintf("https://api.themoviedb.org/3/tv/%d?api_key=%s&language=en-US",
		tvID, url.QueryEscape(tmdbAPIKey))

	req2, err := http.NewRequestWithContext(ctx, "GET", detailsURL, nil)
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
	if err := json.NewDecoder(resp2.Body).Decode(&tvDetails); err != nil {
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
		return nil // No sequels → no useful season map; let TMDB/SuperFlix provide one
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

// --- SuperFlix-based TMDB season lookup (no API key needed) ---

var (
	// reSFTMDBID extracts TMDB IDs from SuperFlix search result HTML.
	// Matches: data-msg="Copiar TMDB" data-copy="73223"
	reSFTMDBID = regexp.MustCompile(`data-msg="Copiar TMDB"\s+data-copy="(\d+)"`)
	// reSFSerieLink detects serie links in SuperFlix HTML.
	// Matches: data-copy="http...//superflixapi.rest/serie/73223" or similar
	reSFSerieLink = regexp.MustCompile(`data-copy="[^"]*?/serie/(\d+)"`)
	// reSFAllEpisodes extracts ALL_EPISODES JS variable from player page.
	reSFAllEpisodes = regexp.MustCompile(`var ALL_EPISODES\s*=\s*(\{.+?\});`)
)

// buildSeasonMapFromSuperFlix searches SuperFlix for an anime by name, finds
// its TMDB ID, fetches the player page, and extracts per-season episode counts
// from the ALL_EPISODES JavaScript variable. No API key needed.
func (e *Enricher) buildSeasonMapFromSuperFlix(ctx context.Context, animeName string) []SeasonMapping {
	cleanName := cleanSearchName(animeName)
	if cleanName == "" {
		return nil
	}

	// Step 1: Search SuperFlix
	searchURL := "https://superflixapi.rest/pesquisar?s=" + url.QueryEscape(cleanName)
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := e.client.Do(req)
	if err != nil {
		util.Debug("SuperFlix search failed", "error", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024)) // 512KB max
	if err != nil {
		return nil
	}
	html := string(body)

	// Find all TMDB IDs that are for series (not movies)
	tmdbID := findSerieTMDBID(html, cleanName)
	if tmdbID == "" {
		util.Debug("SuperFlix: no serie TMDB ID found", "query", cleanName)
		return nil
	}

	util.Debug("SuperFlix: found serie TMDB ID", "tmdbID", tmdbID, "query", cleanName)

	// Step 2: Fetch episode data from player page
	// Must include Referer and Sec-Fetch-* headers or SuperFlix returns
	// "ACESSO RESTRITO" instead of the actual player page with ALL_EPISODES.
	epURL := "https://superflixapi.rest/serie/" + tmdbID
	req2, err := http.NewRequestWithContext(ctx, "GET", epURL, nil)
	if err != nil {
		return nil
	}
	req2.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req2.Header.Set("Referer", "https://superflixapi.rest/")
	req2.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req2.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
	req2.Header.Set("Sec-Fetch-Dest", "iframe")
	req2.Header.Set("Sec-Fetch-Mode", "navigate")
	req2.Header.Set("Sec-Fetch-Site", "cross-site")

	resp2, err := e.client.Do(req2)
	if err != nil {
		util.Debug("SuperFlix episode page failed", "error", err)
		return nil
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		return nil
	}

	body2, err := io.ReadAll(io.LimitReader(resp2.Body, 1024*1024)) // 1MB max
	if err != nil {
		return nil
	}

	// Parse ALL_EPISODES
	m := reSFAllEpisodes.FindSubmatch(body2)
	if len(m) < 2 {
		util.Debug("SuperFlix: ALL_EPISODES not found", "tmdbID", tmdbID)
		return nil
	}

	var allEpisodes map[string][]json.RawMessage
	if err := json.Unmarshal(m[1], &allEpisodes); err != nil {
		return nil
	}

	// Build season map from episode counts per season key
	type seasonInfo struct {
		num     int
		epCount int
	}
	var seasons []seasonInfo
	for key, eps := range allEpisodes {
		sNum, err := strconv.Atoi(key)
		if err != nil || sNum < 1 {
			continue
		}
		seasons = append(seasons, seasonInfo{num: sNum, epCount: len(eps)})
	}

	if len(seasons) <= 1 {
		return nil
	}

	// Sort by season number
	sort.Slice(seasons, func(i, j int) bool { return seasons[i].num < seasons[j].num })

	// Build cumulative episode ranges
	var result []SeasonMapping
	currentEp := 1
	for _, s := range seasons {
		if s.epCount <= 0 {
			continue
		}
		result = append(result, SeasonMapping{
			Season:       s.num,
			StartEp:      currentEp,
			EndEp:        currentEp + s.epCount - 1,
			EpisodeCount: s.epCount,
		})
		currentEp += s.epCount
	}

	util.Debug("SuperFlix season map built", "seasons", len(result), "tmdbID", tmdbID)
	return result
}

// reSFCardTitle matches the alt attribute from card images or h3 text content.
var reSFCardTitle = regexp.MustCompile(`(?i)(?:alt="([^"]+)"|<h3[^>]*>([^<]+)<)`)

// findSerieTMDBID extracts the TMDB ID from SuperFlix HTML that best matches
// the given search name. It considers only results that link to /serie/ (TV shows)
// and picks the one whose title is the closest match.
func findSerieTMDBID(html string, searchName string) string {
	type candidate struct {
		tmdbID string
		title  string
	}

	// Find all /serie/ links with their positions
	serieMatches := reSFSerieLink.FindAllStringSubmatchIndex(html, -1)
	if len(serieMatches) == 0 {
		// Fallback: use any TMDB ID
		tmdbMatches := reSFTMDBID.FindAllStringSubmatch(html, -1)
		if len(tmdbMatches) > 0 {
			return tmdbMatches[0][1]
		}
		return ""
	}

	normalizedSearch := strings.ToLower(strings.TrimSpace(searchName))

	var candidates []candidate
	for _, m := range serieMatches {
		tmdbID := html[m[2]:m[3]]
		// Look at the surrounding context (±3000 chars) for a title
		start := m[0] - 3000
		if start < 0 {
			start = 0
		}
		end := m[1] + 1000
		if end > len(html) {
			end = len(html)
		}
		block := html[start:end]

		// Extract all titles from the block (alt attrs and h3 tags)
		titleMatches := reSFCardTitle.FindAllStringSubmatch(block, -1)
		bestTitle := ""
		for _, tm := range titleMatches {
			t := tm[1]
			if t == "" {
				t = tm[2]
			}
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			// Prefer the title closest to the serie link (last match in block)
			bestTitle = t
		}
		util.Debug("SuperFlix candidate", "tmdbID", tmdbID, "title", bestTitle)
		candidates = append(candidates, candidate{tmdbID: tmdbID, title: bestTitle})
	}

	if len(candidates) == 0 {
		return ""
	}

	// If only one candidate, use it
	if len(candidates) == 1 {
		return candidates[0].tmdbID
	}

	// Pick the candidate whose title best matches the search name.
	// 1. Exact match (case-insensitive)
	for _, c := range candidates {
		if strings.EqualFold(c.title, normalizedSearch) {
			return c.tmdbID
		}
	}
	// 2. Search name contains candidate title
	for _, c := range candidates {
		if c.title != "" && strings.Contains(normalizedSearch, strings.ToLower(c.title)) {
			return c.tmdbID
		}
	}
	// 3. Candidate title contains search name
	for _, c := range candidates {
		if strings.Contains(strings.ToLower(c.title), normalizedSearch) {
			return c.tmdbID
		}
	}

	// No title match; return first
	return candidates[0].tmdbID
}
