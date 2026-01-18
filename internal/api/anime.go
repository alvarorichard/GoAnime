package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/api/movie"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/errors"
)

// Common HTTP client instance
var httpClient = &http.Client{}

// GetEpisodeData fetches episode data using multiple providers with fallback support.
// It tries Jikan (MyAnimeList) first, then falls back to AniList and Kitsu if needed.
// This provides robust episode data retrieval even when primary APIs are unavailable.
func GetEpisodeData(animeID int, episodeNo int, anime *models.Anime) error {
	return GetEpisodeDataWithFallback(animeID, episodeNo, anime)
}

// GetMovieData fetches movie/OVA data for a given anime ID from Jikan API
func GetMovieData(animeID int, anime *models.Anime) error {

	url := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d", animeID)

	response, err := makeGetRequest(url, nil)
	if err != nil {
		return fmt.Errorf("error fetching data from Jikan (MyAnimeList) API: %w", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid response structure: missing or invalid 'data' field")
	}

	// Helper functions to safely get values
	getStringValue := func(field string) string {
		if value, ok := data[field].(string); ok {
			return value
		}
		return ""
	}

	getIntValue := func(field string) int {
		if value, ok := data[field].(float64); ok {
			return int(value)
		}
		return 0
	}

	getBoolValue := func(field string) bool {
		if value, ok := data[field].(bool); ok {
			return value
		}
		return false
	}

	// Assign values to the Anime struct
	if len(anime.Episodes) == 0 {
		anime.Episodes = make([]models.Episode, 1)
	}
	anime.Episodes[0].Title.Romaji = getStringValue("title_romanji")
	anime.Episodes[0].Title.English = getStringValue("title")
	anime.Episodes[0].Title.Japanese = getStringValue("title_japanese")
	anime.Episodes[0].Aired = getStringValue("aired")
	anime.Episodes[0].Duration = getIntValue("duration")
	anime.Episodes[0].IsFiller = getBoolValue("filler")
	anime.Episodes[0].IsRecap = getBoolValue("recap")
	anime.Episodes[0].Synopsis = getStringValue("synopsis")

	return nil
}

// FetchAnimeDetails retrieves additional information for the selected anime
func FetchAnimeDetails(anime *models.Anime) error {
	response, err := http.Get(anime.URL)
	if err != nil {
		return errors.Wrap(err, "failed to get anime details page")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("error get details")

		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get anime details page: %s", response.Status)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return errors.Wrap(err, "failed to parse anime details page")
	}

	imageURL, exists := doc.Find(`meta[property="og:image"]`).Attr("content")
	if !exists || imageURL == "" {
		return errors.New("cover image URL not found")
	}

	return nil
}

func SearchAnime(animeName string) (*models.Anime, error) {
	start := time.Now()
	util.Debugf("[PERF] SearchAnime started for %s", animeName)

	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", models.AnimeFireURL, url.PathEscape(animeName))

	for {
		selectedAnime, nextPageURL, err := searchAnimeOnPage(currentPageURL)
		if err != nil {
			util.Debugf("[PERF] SearchAnime failed for %s after %v", animeName, time.Since(start))
			return nil, err
		}
		if selectedAnime != nil {
			if err := enrichAnimeData(selectedAnime); err != nil {
				util.Errorf("Error enriching anime data: %v", err)
			}
			util.Debugf("[PERF] SearchAnime completed for %s in %v", animeName, time.Since(start))
			return selectedAnime, nil
		}

		if nextPageURL == "" {
			util.Debugf("[PERF] No results found for %s after %v", animeName, time.Since(start))
			return nil, errors.New("no anime found with the given name")
		}
		currentPageURL = models.AnimeFireURL + nextPageURL
	}
}

// Unified function to fetch anime data from Jikan API
func FetchAnimeData(animeID int, episodeNo int, anime *models.Anime) error {
	endpoint := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d", animeID)
	if episodeNo > 0 {
		endpoint = fmt.Sprintf("%s/episodes/%d", endpoint, episodeNo)
	}

	data, err := makeGetRequest(endpoint, nil)
	if err != nil {
		return fmt.Errorf("jikan API request failed: %w", err)
	}

	responseData, ok := data["data"].(map[string]interface{})
	if !ok {
		return errors.New("invalid response structure from Jikan API")
	}

	// Ensure anime has at least one episode
	if len(anime.Episodes) == 0 {
		anime.Episodes = make([]models.Episode, 1)
	}

	// Populate episode data from response
	ep := &anime.Episodes[0]
	ep.Title.Romaji = getStringValue(responseData, "title_romanji")
	ep.Title.English = getStringValue(responseData, "title")
	ep.Title.Japanese = getStringValue(responseData, "title_japanese")
	ep.Aired = getStringValue(responseData, "aired")
	ep.Duration = getIntValue(responseData, "duration")
	ep.IsFiller = getBoolValue(responseData, "filler")
	ep.IsRecap = getBoolValue(responseData, "recap")
	ep.Synopsis = getStringValue(responseData, "synopsis")

	return nil
}

// Helper functions for map value extraction
func getStringValue(data map[string]interface{}, field string) string {
	val, _ := data[field].(string)
	return val
}

func getIntValue(data map[string]interface{}, field string) int {
	switch val := data[field].(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	}
	return 0
}

func getBoolValue(data map[string]interface{}, field string) bool {
	val, _ := data[field].(bool)
	return val
}

// Enrich anime data from AniList
func enrichAnimeData(anime *models.Anime) error {
	// Use TMDB enrichment for FlixHQ movies/TV shows
	if anime.Source == "FlixHQ" || anime.MediaType == models.MediaTypeMovie || anime.MediaType == models.MediaTypeTV {
		util.Debug("Using TMDB enrichment for movie/TV content", "name", anime.Name)
		return movie.EnrichMedia(anime)
	}

	aniListInfo, err := FetchAnimeFromAniList(anime.Name)
	if err != nil {
		util.Debugf("Warning: AniList enrichment failed for '%s': %v", anime.Name, err)
		return fmt.Errorf("AniList enrichment failed: %w", err)
	}

	anime.AnilistID = aniListInfo.Data.Media.ID
	anime.MalID = aniListInfo.Data.Media.IDMal
	anime.Details = aniListInfo.Data.Media

	if cover := aniListInfo.Data.Media.CoverImage.Large; cover != "" {
		anime.ImageURL = cover
	} else {
		util.Debugf("Cover image not found for: %s", anime.Name)
	}

	util.Debugf("AniList Data: ID:%d, MAL:%d, Title:%s",
		aniListInfo.Data.Media.ID,
		aniListInfo.Data.Media.IDMal,
		aniListInfo.Data.Media.Title.Romaji)
	return nil
}

func searchAnimeOnPage(pageURL string) (*models.Anime, string, error) {
	resp, err := httpGetWithUA(pageURL)
	if err != nil {
		return nil, "", errors.Wrap(err, "HTTP request failed")
	}
	defer safeClose(resp.Body, "search page response body")

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusForbidden {
			return nil, "", errors.New("access restricted: VPN required")
		}
		return nil, "", fmt.Errorf("server returned: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, "", errors.Wrap(err, "HTML parsing failed")
	}

	animes := ParseAnimes(doc)
	util.Debugf("Found %d anime(s)", len(animes))

	if len(animes) > 0 {
		selectedAnime, err := selectAnimeWithGoFuzzyFinder(animes)
		return selectedAnime, "", err
	}

	if nextPage, exists := doc.Find(".pagination .next a").Attr("href"); exists {
		return nil, nextPage, nil
	}
	return nil, "", nil
}

func ParseAnimes(doc *goquery.Document) []models.Anime {
	var animes []models.Anime

	doc.Find(".row.ml-1.mr-1 a").Each(func(_ int, s *goquery.Selection) {
		if urlPath, exists := s.Attr("href"); exists {
			name := strings.TrimSpace(s.Text())
			animes = append(animes, models.Anime{
				Name: name,
				URL:  resolveURL(models.AnimeFireURL, urlPath),
			})
			util.Debugf("Parsed: %s", name)
		}
	})
	return animes
}

func FetchAnimeFromAniList(animeName string) (*models.AniListResponse, error) {
	cleanedName := CleanTitle(animeName)
	util.Debugf("Querying AniList for: '%s' (original: '%s')", cleanedName, animeName)

	// Check cache first
	cache := util.GetAniListCache()
	cacheKey := "anilist:" + strings.ToLower(cleanedName)
	if cached, found := cache.Get(cacheKey); found {
		var result models.AniListResponse
		if err := json.Unmarshal(cached, &result); err == nil && result.Data.Media.ID != 0 {
			util.Debugf("AniList cache hit for: '%s'", cleanedName)
			return &result, nil
		}
	}

	// Try multiple search variations for better matching
	searchVariations := generateSearchVariations(cleanedName)

	query := `query ($search: String) {
        Media(search: $search, type: ANIME) {
            id
            title { romaji english native }
            idMal
            coverImage { large }
            synonyms
        }
    }`

	var lastErr error
	for _, searchTerm := range searchVariations {
		util.Debugf("Trying AniList search with: '%s'", searchTerm)

		jsonData, err := json.Marshal(map[string]interface{}{
			"query": query,
			"variables": map[string]interface{}{
				"search": searchTerm,
			},
		})
		if err != nil {
			lastErr = fmt.Errorf("JSON marshal failed: %w", err)
			continue
		}

		resp, err := httpPostFast("https://graphql.anilist.co", jsonData)
		if err != nil {
			lastErr = fmt.Errorf("AniList request failed: %w", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		safeClose(resp.Body, "AniList response body")

		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		util.Debugf("AniList response status: %d", resp.StatusCode)

		if resp.StatusCode != http.StatusOK {
			util.Debugf("AniList error response: %s", string(body))
			lastErr = fmt.Errorf("AniList returned: %s", resp.Status)
			continue
		}

		var result models.AniListResponse
		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = fmt.Errorf("JSON decode failed: %w", err)
			continue
		}

		if result.Data.Media.ID == 0 {
			lastErr = errors.New("no matching anime found on AniList")
			continue
		}

		// Cache the successful result
		cache.Set(cacheKey, body)

		util.Debugf("AniList found: ID=%d, MAL=%d, Title=%s",
			result.Data.Media.ID,
			result.Data.Media.IDMal,
			result.Data.Media.Title.Romaji)

		return &result, nil
	}

	return nil, lastErr
}

func selectAnimeWithGoFuzzyFinder(animes []models.Anime) (*models.Anime, error) {
	if len(animes) == 0 {
		return nil, errors.New("no anime available for selection")
	}

	sort.Slice(animes, func(i, j int) bool {
		return animes[i].Name < animes[j].Name
	})

	idx, err := fuzzyfinder.Find(animes, func(i int) string {
		name := animes[i].Name
		name = strings.ReplaceAll(name, "[AllAnime]", "[English]")
		name = strings.ReplaceAll(name, "[AnimeFire]", "[Portuguese]")
		return name
	})
	if err != nil {
		return nil, fmt.Errorf("fuzzy selection failed: %w", err)
	}

	if idx < 0 || idx >= len(animes) {
		return nil, errors.New("invalid selection index")
	}
	return &animes[idx], nil
}

// HTTP helper functions
func httpGetWithUA(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	return util.GetSharedClient().Do(req)
}

func httpPost(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "GoAnime/1.0")
	return util.GetSharedClient().Do(req)
}

// httpPostFast uses the fast client for quick API requests
func httpPostFast(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "GoAnime/1.0")
	return util.GetFastClient().Do(req)
}

func makeGetRequest(url string, headers map[string]string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	// Set default User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	// Set additional headers if provided
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET request failed: %w", err)
	}
	defer safeClose(resp.Body, "API response body")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %w", err)
	}
	return responseData, nil
}

// Utility functions
func resolveURL(base, ref string) string {
	baseURL, _ := url.Parse(base)
	refURL, _ := url.Parse(ref)
	return baseURL.ResolveReference(refURL).String()
}

// generateSearchVariations creates multiple search term variations for better AniList matching
// This is especially important for Brazilian sources that have localized titles
func generateSearchVariations(cleanedName string) []string {
	variations := []string{cleanedName}
	seen := make(map[string]bool)
	seen[cleanedName] = true

	addVariation := func(v string) {
		v = strings.TrimSpace(v)
		if v != "" && !seen[v] {
			seen[v] = true
			variations = append(variations, v)
		}
	}

	// Variation: Title case (for all lowercase names from URLs)
	if strings.ToLower(cleanedName) == cleanedName {
		words := strings.Fields(cleanedName)
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(string(w[0])) + w[1:]
			}
		}
		addVariation(strings.Join(words, " "))
	}

	// Variation: Remove common subtitle patterns after colon
	if idx := strings.Index(cleanedName, ":"); idx > 0 {
		addVariation(strings.TrimSpace(cleanedName[:idx]))
	}

	// Variation: Remove trailing roman numerals (seasons like II, III, IV)
	reRoman := regexp.MustCompile(`\s+(?:II|III|IV|V|VI|VII|VIII|IX|X)\s*$`)
	if match := reRoman.FindString(cleanedName); match != "" {
		addVariation(strings.TrimSpace(reRoman.ReplaceAllString(cleanedName, "")))
	}

	// Variation: Remove trailing numbers that might be season indicators (2, 3, 4, etc.)
	reTrailingNum := regexp.MustCompile(`\s+\d+\s*$`)
	if match := reTrailingNum.FindString(cleanedName); match != "" {
		addVariation(strings.TrimSpace(reTrailingNum.ReplaceAllString(cleanedName, "")))
	}

	// Variation: Common Japanese title adaptations
	// Try removing "no" particles which are sometimes omitted
	if strings.Contains(cleanedName, " no ") {
		addVariation(strings.ReplaceAll(cleanedName, " no ", " "))
	}

	// Variation: Try with common alternative title patterns
	// Some anime have "The" prefix in English but not in romaji
	if strings.HasPrefix(strings.ToLower(cleanedName), "the ") {
		addVariation(cleanedName[4:])
	}

	// Variation: For very long titles, try first few words
	words := strings.Fields(cleanedName)
	if len(words) > 4 {
		// Try first 3 words
		addVariation(strings.Join(words[:3], " "))
		// Try first 4 words
		addVariation(strings.Join(words[:4], " "))
	}

	util.Debugf("Generated %d search variations for '%s': %v", len(variations), cleanedName, variations)
	return variations
}

func CleanTitle(title string) string {
	cleaned := title

	// Remove media type tags like [Movies/TV], [Anime], [Series], [Movie] at the start
	reMediaTags := regexp.MustCompile(`^\s*\[(?:Movies?(?:/TV)?|TV|Anime|Series|Show)\]\s*`)
	cleaned = strings.TrimSpace(reMediaTags.ReplaceAllString(cleaned, ""))

	// Remove language tags like [English], [Portuguese], [Português] at the start
	reLangTags := regexp.MustCompile(`^\s*\[(?:English|Portuguese|Português|Japonês|Japanese)\]\s*`)
	cleaned = strings.TrimSpace(reLangTags.ReplaceAllString(cleaned, ""))

	// Remove source tags like 🔥[AnimeFire], 🌐[AllAnime], or [AnimeDrive]
	re1 := regexp.MustCompile(`(?i)[🔥🌐]?\[(?:animefire|allanime|animedrive)\]\s*`)
	cleaned = strings.TrimSpace(re1.ReplaceAllString(cleaned, ""))

	// Remove everything after em-dash or en-dash (typically subtitles like "– Todos os Episódios")
	// This handles both em-dash (—), en-dash (–), and regular dash with spaces ( - )
	reEmDash := regexp.MustCompile(`\s*[–—]\s+.*$`)
	cleaned = strings.TrimSpace(reEmDash.ReplaceAllString(cleaned, ""))
	reSpaceDash := regexp.MustCompile(`\s+-\s+.*$`)
	cleaned = strings.TrimSpace(reSpaceDash.ReplaceAllString(cleaned, ""))

	// Remove content in parentheses if it contains language info (do this BEFORE removing standalone language indicators)
	re6 := regexp.MustCompile(`(?i)\s*\([^)]*(?:dublado|legendado|dub|sub)[^)]*\)`)
	cleaned = strings.TrimSpace(re6.ReplaceAllString(cleaned, ""))

	// Remove standalone language indicators (not in parentheses) - more comprehensive for Brazilian sources
	re2 := regexp.MustCompile(`(?i)\s+(?:dublado|legendado|dub|sub|dual\s*[aá]udio)\s*$`)
	cleaned = strings.TrimSpace(re2.ReplaceAllString(cleaned, ""))

	// Remove "Todos os Episodios" and similar Brazilian phrases (in case em-dash removal didn't catch it)
	re3 := regexp.MustCompile(`(?i)[-–—]?\s*todos\s+os\s+epis[oó]dios`)
	cleaned = strings.TrimSpace(re3.ReplaceAllString(cleaned, ""))

	// Remove "Completo" or "Episodio X" suffixes common in Brazilian sources
	reCompleto := regexp.MustCompile(`(?i)\s+(?:completo|episodio\s*\d+|ep\s*\d+)\s*$`)
	cleaned = strings.TrimSpace(reCompleto.ReplaceAllString(cleaned, ""))

	// Remove season indicators like "X Temporada", "Season X", "Temporada X", "Xª Temporada"
	reSeasonPt := regexp.MustCompile(`(?i)\s*[-–—]?\s*(?:\d+[ªº]?\s*temporada|temporada\s*\d+|season\s*\d+|\d+(?:st|nd|rd|th)\s*season)\s*$`)
	cleaned = strings.TrimSpace(reSeasonPt.ReplaceAllString(cleaned, ""))

	// Remove "Parte X" (Part X) common in Brazilian titles
	rePart := regexp.MustCompile(`(?i)\s*[-–—]?\s*(?:parte\s*\d+|part\s*\d+)\s*$`)
	cleaned = strings.TrimSpace(rePart.ReplaceAllString(cleaned, ""))

	// Remove season/episode indicators like "2.0 A2" at the end (but NOT plain season numbers)
	re4 := regexp.MustCompile(`\s+\d+(\.\d+)?\s+A\d+\s*$`)
	cleaned = strings.TrimSpace(re4.ReplaceAllString(cleaned, ""))

	// Remove decimal version numbers at the end like "3.5" (but NOT "Season 2")
	re5 := regexp.MustCompile(`\s+\d+\.\d+\s*$`)
	cleaned = strings.TrimSpace(re5.ReplaceAllString(cleaned, ""))

	// Remove episode count like "(171 episodes)" or "(1 eps)" or Portuguese equivalents
	re7 := regexp.MustCompile(`(?i)\s*\(\d+\s+(?:episodes?|eps?|epis[oó]dios?)\)`)
	cleaned = strings.TrimSpace(re7.ReplaceAllString(cleaned, ""))

	// Remove special titles and additions after colon
	re8 := regexp.MustCompile(`(?i):\s*(?:Jump Festa \d+|The All Magic Knights|Sword of the Wizard King|Mahou Tei no Ken).*$`)
	cleaned = strings.TrimSpace(re8.ReplaceAllString(cleaned, ""))

	// Remove N/A ratings and similar suffixes
	re9 := regexp.MustCompile(`(?i)\s+N/A\s*$`)
	cleaned = strings.TrimSpace(re9.ReplaceAllString(cleaned, ""))

	// Remove rating scores like "7.12" or "8.5" at the end (only decimal numbers)
	re10 := regexp.MustCompile(`\s+\d+\.\d+\s*$`)
	cleaned = strings.TrimSpace(re10.ReplaceAllString(cleaned, ""))

	// Remove empty parentheses that may be left after other cleanups
	reEmptyParens := regexp.MustCompile(`\s*\(\s*\)`)
	cleaned = strings.TrimSpace(reEmptyParens.ReplaceAllString(cleaned, ""))

	// Remove trailing colons that may be left after removing season/part info
	cleaned = strings.TrimSuffix(strings.TrimSpace(cleaned), ":")
	cleaned = strings.TrimSpace(cleaned)

	// Replace hyphens with spaces (for URL-style names like "black-clover")
	// But only if surrounded by letters (not em-dashes already handled above)
	cleaned = regexp.MustCompile(`([a-zA-Z])-([a-zA-Z])`).ReplaceAllString(cleaned, "$1 $2")

	// Replace underscores with spaces
	cleaned = strings.ReplaceAll(cleaned, "_", " ")

	// Normalize whitespace
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	util.Debugf("CleanTitle: '%s' -> '%s'", title, cleaned)

	return cleaned
}

func safeClose(closer io.Closer, name string) {
	if err := closer.Close(); err != nil {
		util.Debugf("Error closing %s: %v", name, err)
	}
}
