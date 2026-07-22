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
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/pkg/errors"
)

// Common HTTP client instance - reuse the shared singleton for connection pooling
var httpClient = util.GetSharedClient()

// jikanBaseURL is the Jikan (MyAnimeList unofficial) API root. It is a var so
// tests can point it at a local httptest server.
var jikanBaseURL = "https://api.jikan.moe/v4"

// GetEpisodeData fetches episode data using multiple providers with fallback support.
// It tries Jikan (MyAnimeList) first, then falls back to AniList and Kitsu if needed.
// This provides robust episode data retrieval even when primary APIs are unavailable.
func GetEpisodeData(animeID int, episodeNo int, anime *models.Anime) error {
	return GetEpisodeDataWithFallback(animeID, episodeNo, anime)
}

// GetMovieData fetches movie/OVA data for a given anime ID from Jikan API
func GetMovieData(animeID int, anime *models.Anime) error {

	url := fmt.Sprintf("%s/anime/%d", jikanBaseURL, animeID)

	response, err := makeGetRequest(url, nil)
	if err != nil {
		return fmt.Errorf("error fetching data from Jikan (MyAnimeList) API: %w", err)
	}

	data, ok := response["data"].(map[string]any)
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
	response, err := SafeGet(anime.URL)
	if err != nil {
		return errors.Wrap(err, "failed to get anime details page")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			util.Debugf("Failed to close response body: %v", err)
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
				// Best-effort only: enrichment adds cover art / MAL id. Playback,
				// episode listing and download all work without it, so this must
				// not read as a failure (issue #184).
				util.Warn("Metadata enrichment unavailable; continuing without it", "anime", selectedAnime.Name, "error", err)
			}
			util.Debugf("[PERF] SearchAnime completed for %s in %v", animeName, time.Since(start))
			return selectedAnime, nil
		}

		if nextPageURL == "" {
			util.Debugf("[PERF] No results found for %s after %v", animeName, time.Since(start))
			return nil, errors.New("no anime found with the given name")
		}
		// Validate scraped next-page URL to prevent open redirects
		parsedNext, err := url.Parse(nextPageURL)
		if err != nil || (parsedNext.Host != "" && !strings.Contains(parsedNext.Host, "animefire")) {
			return nil, fmt.Errorf("suspicious next page URL rejected: %s", nextPageURL)
		}
		currentPageURL = models.AnimeFireURL + nextPageURL
	}
}

// Unified function to fetch anime data from Jikan API
func FetchAnimeData(animeID int, episodeNo int, anime *models.Anime) error {
	endpoint := fmt.Sprintf("%s/anime/%d", jikanBaseURL, animeID)
	if episodeNo > 0 {
		endpoint = fmt.Sprintf("%s/episodes/%d", endpoint, episodeNo)
	}

	data, err := makeGetRequest(endpoint, nil)
	if err != nil {
		return fmt.Errorf("jikan API request failed: %w", err)
	}

	responseData, ok := data["data"].(map[string]any)
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
func getStringValue(data map[string]any, field string) string {
	val, _ := data[field].(string)
	return val
}

func getIntValue(data map[string]any, field string) int {
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

func getBoolValue(data map[string]any, field string) bool {
	val, _ := data[field].(bool)
	return val
}

// Enrich anime data from AniList
func enrichAnimeData(anime *models.Anime) error {
	// Use TMDB/OMDb enrichment for movie/TV catalogs. SuperFlix is included by
	// SOURCE, not just media type: its catalog tags western animation (e.g.
	// "Os Simpsons") as anime, which would otherwise fall through to AniList —
	// a query that can't match (TMDB-indexed content) and pays a Cloudflare
	// challenge for nothing. Mirrors appflow.fetchAnimeDetailsCore.
	if anime.HasInteractiveEpisodeFlow() {
		util.Debug("Using TMDB enrichment for movie/TV content", "name", anime.Name)
		return movie.EnrichMedia(anime)
	}

	aniListInfo, err := FetchAnimeFromAniListWithURL(anime.Name, anime.URL)
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
	return FetchAnimeFromAniListWithURL(animeName, "")
}

func selectAnimeWithGoFuzzyFinder(animes []models.Anime) (*models.Anime, error) {
	if len(animes) == 0 {
		return nil, errors.New("no anime available for selection")
	}

	sort.Slice(animes, func(i, j int) bool {
		return animes[i].Name < animes[j].Name
	})

	idx, err := tui.Find(animes, func(i int) string {
		name := animes[i].Name
		name = strings.ReplaceAll(name, "[AllAnime]", "[English]")
		name = strings.ReplaceAll(name, "[AnimeFire]", "[PT-BR]")
		// Append release year if available and not already in the name
		if animes[i].Year != "" && !strings.Contains(name, "("+animes[i].Year+")") {
			name += " (" + animes[i].Year + ")"
		}
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
	return util.GetSharedClient().Do(req) // #nosec G704
}

// NOTE: the old httpPost/httpPostFast helpers (backed by the shared surf clients)
// were deleted. They only ever served AniList, and routing AniList through an
// impersonating client is precisely the bug in issue #184 — surf rewrites the
// User-Agent to Chrome's, and AniList 403s browser UAs. Use aniListPost instead.

// aniListClient is a plain net/http client reserved for AniList.
//
// It deliberately bypasses util.GetSharedClient()/GetFastClient(): those are
// backed by surf with Chrome impersonation, which REWRITES the User-Agent to
// Chrome's on every request. That forced browser UA is exactly what AniList
// rejects (see AniListUserAgent), so no header set by the caller can rescue the
// shared clients — the transport itself has to be a plain one.
var aniListClient = &http.Client{Timeout: 20 * time.Second}

// aniListPost sends a GraphQL request to AniList through the plain client with a
// non-browser User-Agent. The returned response body is already read and closed;
// callers use the returned bytes and may still read StatusCode/Status.
func aniListPost(endpoint string, jsonData []byte) (*http.Response, []byte, error) {
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", netx.APIUserAgent)

	resp, err := aniListClient.Do(req) // #nosec G704
	if err != nil {
		return nil, nil, err
	}
	body, rerr := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	safeClose(resp.Body, "AniList response body")
	if rerr != nil {
		return resp, nil, rerr
	}
	return resp, body, nil
}

func makeGetRequest(url string, headers map[string]string) (map[string]any, error) {
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

	resp, err := httpClient.Do(req) // #nosec G704
	if err != nil {
		return nil, fmt.Errorf("GET request failed: %w", err)
	}
	defer safeClose(resp.Body, "API response body")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var responseData map[string]any
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

// normalizeAccents replaces common accented characters with their ASCII equivalents.
func normalizeAccents(s string) string {
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "ã", "a", "â", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "õ", "o", "ô", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c", "ñ", "n",
		"Á", "A", "À", "A", "Ã", "A", "Â", "A", "Ä", "A",
		"É", "E", "È", "E", "Ê", "E", "Ë", "E",
		"Í", "I", "Ì", "I", "Î", "I", "Ï", "I",
		"Ó", "O", "Ò", "O", "Õ", "O", "Ô", "O", "Ö", "O",
		"Ú", "U", "Ù", "U", "Û", "U", "Ü", "U",
		"Ç", "C", "Ñ", "N",
	)
	return replacer.Replace(s)
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
	if match := reRoman.FindString(cleanedName); match != "" {
		addVariation(strings.TrimSpace(reRoman.ReplaceAllString(cleanedName, "")))
	}

	// Variation: Remove trailing numbers that might be season indicators (2, 3, 4, etc.)
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

	// Variation: Remove common PT-BR descriptive suffixes (Clássico, Classic, etc.)
	// These are used by Brazilian sites to distinguish series (e.g. "Naruto Clássico" vs "Naruto Shippuden")
	if rePtBRSuffix.MatchString(cleanedName) {
		addVariation(strings.TrimSpace(rePtBRSuffix.ReplaceAllString(cleanedName, "")))
	}

	// Variation: Normalize accented characters to ASCII (e.g. Clássico → Classico)
	normalized := normalizeAccents(cleanedName)
	if normalized != cleanedName {
		addVariation(normalized)
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

// Compiled once at init: CleanTitle and generateSearchVariations run per search
// result, so recompiling these on every call dominated the search hot path.
var (
	// generateSearchVariations patterns
	reRoman       = regexp.MustCompile(`\s+(?:II|III|IV|V|VI|VII|VIII|IX|X)\s*$`)
	reTrailingNum = regexp.MustCompile(`\s+\d+\s*$`)
	rePtBRSuffix  = regexp.MustCompile(`(?i)\s+(?:cl[aá]ssico|classic|shippuuden|next\s+generations?)\s*$`)

	// CleanTitle patterns, applied in order
	reMediaTags = regexp.MustCompile(`^\s*\[(?:Movies?(?:/TV)?|TV|Anime|Series|Show)\]\s*`)
	reLangTags  = regexp.MustCompile(`^\s*\[(?:English|PT-BR|Portuguese|Português|Japonês|Japanese|Multilanguage)\]\s*`)
	reSourceTag = regexp.MustCompile(`(?i)[🔥🌐]?\[(?:animefire|allanime|animedrive|9anime)\]\s*`)
	reEmDash    = regexp.MustCompile(`\s*[–—]\s+.*$`)
	// For the regular hyphen ( - ) we cannot strip blindly — it appears inside legitimate
	// titles such as "Jujutsu Kaisen: Shimetsu Kaiyuu - Zenpen" / "- Kouhen" (前編/後編,
	// Part 1 / Part 2 on AniList). Only strip when the suffix matches known PT-BR noise
	// or season/episode markers; otherwise keep the dash and what follows so AniList
	// resolves the correct entry.
	reSpaceDashNoise = regexp.MustCompile(`(?i)\s+-\s+(?:` +
		`dublado|legendado|dual\s*[aá]udio|dub|sub|completo|` +
		`todos\s+os\s+epis[oó]dios?|` +
		`epis[oó]dios?\s*\d+|ep\s*\d+|` +
		`\d+[ªº]?\s*temporada|temporada\s*\d*|` +
		`season\s*\d+|\d+(?:st|nd|rd|th)\s*season|` +
		`parte\s*\d+|part\s*\d+|` +
		`allanime|animefire|animedrive|9anime|goyabu|superflix|flixhq|sflix` +
		`).*$`)
	reLangParens    = regexp.MustCompile(`(?i)\s*\([^)]*(?:dublado|legendado|dub|sub)[^)]*\)`)
	reLangSuffix    = regexp.MustCompile(`(?i)\s+(?:dublado|legendado|dub|sub|dual\s*[aá]udio)\s*$`)
	reTodosEps      = regexp.MustCompile(`(?i)[-–—]?\s*todos\s+os\s+epis[oó]dios`)
	reCompleto      = regexp.MustCompile(`(?i)\s+(?:completo|episodio\s*\d+|ep\s*\d+)\s*$`)
	reSeasonPt      = regexp.MustCompile(`(?i)\s*[-–—]?\s*(?:\d+[ªº]?\s*temporada|temporada\s*\d+|season\s*\d+|\d+(?:st|nd|rd|th)\s*season)\s*$`)
	rePart          = regexp.MustCompile(`(?i)\s*[-–—]?\s*(?:parte\s*\d+|part\s*\d+)\s*$`)
	reSeasonEpTag   = regexp.MustCompile(`\s+\d+(\.\d+)?\s+A\d+\s*$`)
	reDecimalSuffix = regexp.MustCompile(`\s+\d+\.\d+\s*$`)
	reEpCount       = regexp.MustCompile(`(?i)\s*\(\d+\s+(?:episodes?|eps?|epis[oó]dios?)\)`)
	re9AnimeEpInfo  = regexp.MustCompile(`(?i)\s*\((?:HD\s+)?(?:(?:SUB|DUB)\s+)*Ep\s+\d+/\d+\)`)
	reSpecialTitles = regexp.MustCompile(`(?i):\s*(?:Jump Festa \d+|The All Magic Knights|Sword of the Wizard King|Mahou Tei no Ken).*$`)
	reNASuffix      = regexp.MustCompile(`(?i)\s+N/A\s*$`)
	reEmptyParens   = regexp.MustCompile(`\s*\(\s*\)`)
	reHyphenLetters = regexp.MustCompile(`([a-zA-Z])-([a-zA-Z])`)
	reWhitespaceRun = regexp.MustCompile(`\s+`)
)

func CleanTitle(title string) string {
	cleaned := title

	// Remove media type tags like [Movies/TV], [Anime], [Series], [Movie] at the start
	cleaned = strings.TrimSpace(reMediaTags.ReplaceAllString(cleaned, ""))

	// Remove language tags like [English], [PT-BR], [Portuguese], [Português], [Multilanguage] at the start
	cleaned = strings.TrimSpace(reLangTags.ReplaceAllString(cleaned, ""))

	// Remove source tags like 🔥[AnimeFire], 🌐[AllAnime], [AnimeDrive], or [9Anime]
	cleaned = strings.TrimSpace(reSourceTag.ReplaceAllString(cleaned, ""))

	// Remove everything after em-dash or en-dash (typically subtitles like "– Todos os Episódios").
	// Em/en dashes are almost always used as noise separators in PT-BR scrapers.
	cleaned = strings.TrimSpace(reEmDash.ReplaceAllString(cleaned, ""))

	// Strip " - <noise>" suffixes (see reSpaceDashNoise for why plain hyphens
	// can't be stripped blindly).
	cleaned = strings.TrimSpace(reSpaceDashNoise.ReplaceAllString(cleaned, ""))

	// Remove content in parentheses if it contains language info (do this BEFORE removing standalone language indicators)
	cleaned = strings.TrimSpace(reLangParens.ReplaceAllString(cleaned, ""))

	// Remove standalone language indicators (not in parentheses) - more comprehensive for Brazilian sources
	cleaned = strings.TrimSpace(reLangSuffix.ReplaceAllString(cleaned, ""))

	// Remove "Todos os Episodios" and similar Brazilian phrases (in case em-dash removal didn't catch it)
	cleaned = strings.TrimSpace(reTodosEps.ReplaceAllString(cleaned, ""))

	// Remove "Completo" or "Episodio X" suffixes common in Brazilian sources
	cleaned = strings.TrimSpace(reCompleto.ReplaceAllString(cleaned, ""))

	// Remove season indicators like "X Temporada", "Season X", "Temporada X", "Xª Temporada"
	cleaned = strings.TrimSpace(reSeasonPt.ReplaceAllString(cleaned, ""))

	// Remove "Parte X" (Part X) common in Brazilian titles
	cleaned = strings.TrimSpace(rePart.ReplaceAllString(cleaned, ""))

	// Remove season/episode indicators like "2.0 A2" at the end (but NOT plain season numbers)
	cleaned = strings.TrimSpace(reSeasonEpTag.ReplaceAllString(cleaned, ""))

	// Remove decimal version numbers at the end like "3.5" (but NOT "Season 2")
	cleaned = strings.TrimSpace(reDecimalSuffix.ReplaceAllString(cleaned, ""))

	// Remove episode count like "(171 episodes)" or "(1 eps)" or Portuguese equivalents
	cleaned = strings.TrimSpace(reEpCount.ReplaceAllString(cleaned, ""))

	// Remove 9Anime-style episode info like "(HD SUB DUB Ep 170/170)" or "(SUB Ep 12/12)"
	cleaned = strings.TrimSpace(re9AnimeEpInfo.ReplaceAllString(cleaned, ""))

	// Remove special titles and additions after colon
	cleaned = strings.TrimSpace(reSpecialTitles.ReplaceAllString(cleaned, ""))

	// Remove N/A ratings and similar suffixes
	cleaned = strings.TrimSpace(reNASuffix.ReplaceAllString(cleaned, ""))

	// Remove rating scores like "7.12" or "8.5" at the end (only decimal numbers)
	cleaned = strings.TrimSpace(reDecimalSuffix.ReplaceAllString(cleaned, ""))

	// Remove empty parentheses that may be left after other cleanups
	cleaned = strings.TrimSpace(reEmptyParens.ReplaceAllString(cleaned, ""))

	// Remove trailing colons that may be left after removing season/part info
	cleaned = strings.TrimSuffix(strings.TrimSpace(cleaned), ":")
	cleaned = strings.TrimSpace(cleaned)

	// Remove trailing hyphens left over when a later rule consumed noise that was
	// preceded by " - " (e.g. "Anime - Dublado" → "Anime -" → "Anime").
	cleaned = strings.TrimSuffix(strings.TrimSpace(cleaned), "-")
	cleaned = strings.TrimSpace(cleaned)

	// Replace hyphens with spaces (for URL-style names like "black-clover")
	// But only if surrounded by letters (not em-dashes already handled above)
	cleaned = reHyphenLetters.ReplaceAllString(cleaned, "$1 $2")

	// Replace underscores with spaces
	cleaned = strings.ReplaceAll(cleaned, "_", " ")

	// Normalize whitespace
	cleaned = reWhitespaceRun.ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	util.Debugf("CleanTitle: '%s' -> '%s'", title, cleaned)

	return cleaned
}

func safeClose(closer io.Closer, name string) {
	if err := closer.Close(); err != nil {
		util.Debugf("Error closing %s: %v", name, err)
	}
}
