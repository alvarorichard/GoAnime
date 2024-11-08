package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/errors"
)

const baseSiteURL = "https://animefire.plus"

type Anime struct {
	Name      string
	URL       string
	ImageURL  string
	Episodes  []Episode
	AnilistID int
	MalID     int // Adicionado para armazenar o ID do MAL
	Details   AniListDetails
}

type Episode struct {
	Number    string
	Num       int
	URL       string
	Title     TitleDetails
	Aired     string
	Duration  int
	IsFiller  bool
	IsRecap   bool
	Synopsis  string
	SkipTimes SkipTimes // Skip times for OP and ED
}

type TitleDetails struct {
	Romaji   string
	English  string
	Japanese string
}

type AniListResponse struct {
	Data struct {
		Media AniListDetails `json:"Media"`
	} `json:"data"`
}

type AniListDetails struct {
	ID           int         `json:"id"`
	IDMal        int         `json:"idMal"` // ID do MAL para integração com Jikan
	Title        Title       `json:"title"`
	Description  string      `json:"description"`
	Genres       []string    `json:"genres"`
	AverageScore int         `json:"averageScore"`
	Episodes     int         `json:"episodes"`
	Status       string      `json:"status"`
	CoverImage   CoverImages `json:"coverImage"`
}

type CoverImages struct {
	Large  string `json:"large"`
	Medium string `json:"medium"`
}

type Title struct {
	Romaji  string `json:"romaji"`
	English string `json:"english"`
}

func SearchAnime(animeName string) (*Anime, error) {
	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseSiteURL, url.PathEscape(animeName))

	if util.IsDebug {
		log.Printf("Searching for anime with URL: %s", currentPageURL)
	}

	for {
		selectedAnime, nextPageURL, err := searchAnimeOnPage(currentPageURL)
		if err != nil {
			return nil, err
		}
		if selectedAnime != nil {
			// Busca de detalhes adicionais pela AniList API, incluindo a imagem de capa
			aniListInfo, err := FetchAnimeFromAniList(selectedAnime.Name)
			if err != nil {
				log.Printf("Error fetching additional data from AniList: %v", err)
			} else {
				selectedAnime.AnilistID = aniListInfo.Data.Media.ID
				selectedAnime.MalID = aniListInfo.Data.Media.IDMal
				selectedAnime.Details = aniListInfo.Data.Media

				// Definindo a imagem de capa do AniList
				if aniListInfo.Data.Media.CoverImage.Large != "" {
					selectedAnime.ImageURL = aniListInfo.Data.Media.CoverImage.Large
					if util.IsDebug {
						log.Printf("Cover image URL retrieved from AniList: %s", selectedAnime.ImageURL)
					}
				} else {
					log.Printf("Cover image URL not found in AniList response for anime: %s", selectedAnime.Name)
				}
				if util.IsDebug {
					log.Printf("AniList ID: %d, MAL ID: %d, Title: %s, Score: %d, Cover Image URL: %s",
						aniListInfo.Data.Media.ID, aniListInfo.Data.Media.IDMal,
						aniListInfo.Data.Media.Title.Romaji, aniListInfo.Data.Media.AverageScore,
						selectedAnime.ImageURL)
				}

			}

			// Busca de dados do primeiro episódio na Jikan API, se o MAL ID estiver disponível
			if selectedAnime.MalID > 0 {
				err = GetEpisodeData(selectedAnime.MalID, 1, selectedAnime)
				if err != nil {
					log.Printf("Error fetching episode data from Jikan API: %v", err)
				} else if util.IsDebug && len(selectedAnime.Episodes) > 0 {
					firstEpisode := selectedAnime.Episodes[0]
					log.Printf("Jikan Episode Data - Title: %s, Aired: %s, Duration: %d, Filler: %t, Synopsis: %s",
						firstEpisode.Title.English,
						firstEpisode.Aired,
						firstEpisode.Duration,
						firstEpisode.IsFiller,
						firstEpisode.Synopsis)
				}
			} else {
				log.Printf("No MAL ID found for anime: %s. Unable to fetch episode data from Jikan API.", selectedAnime.Name)
			}

			return selectedAnime, nil
		}

		if nextPageURL == "" {
			return nil, errors.New("no anime found with the given name")
		}
		currentPageURL = baseSiteURL + nextPageURL
	}
}

// GetEpisodeData fetches episode data for a given anime ID and episode number from Jikan API
func GetEpisodeData(animeID int, episodeNo int, anime *Anime) error {

	url := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d/episodes/%d", animeID, episodeNo)

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
		anime.Episodes = make([]Episode, 1) // Ensure there is at least one episode slot
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

// searchAnimeOnPage searches for anime on a given page and returns the selected anime
func searchAnimeOnPage(pageURL string) (*Anime, string, error) {
	response, err := getHTTPResponse(pageURL)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to perform search request")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusForbidden {
			return nil, "", errors.New("connection refused: you need to be in Brazil or use a VPN to access the server")
		}
		return nil, "", errors.Errorf("search failed, server returned: %s", response.Status)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to parse response")
	}

	animes := ParseAnimes(doc)
	if util.IsDebug {
		log.Printf("Number of animes found: %d", len(animes))
	}

	if len(animes) > 0 {
		selectedAnime, err := selectAnimeWithGoFuzzyFinder(animes)
		if err != nil {
			return nil, "", err
		}
		return selectedAnime, "", nil
	}

	nextPage, exists := doc.Find(".pagination .next a").Attr("href")
	if !exists {
		return nil, "", nil
	}

	return nil, nextPage, nil
}

// ParseAnimes extracts a list of Anime structs from the search results page.
func ParseAnimes(doc *goquery.Document) []Anime {
	var animes []Anime

	doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
		urlPath, exists := s.Attr("href")
		if !exists {
			return
		}
		url := resolveURL(baseSiteURL, urlPath)

		name := strings.TrimSpace(s.Text())

		if util.IsDebug {
			log.Printf("Parsed Anime - Name: %s, URL: %s", name, url)
		}

		animes = append(animes, Anime{
			Name: name,
			URL:  url,
		})
	})

	return animes
}

// FetchAnimeDetails retrieves additional information for the selected anime
func FetchAnimeDetails(anime *Anime) error {
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

func FetchAnimeFromAniList(animeName string) (*AniListResponse, error) {
	cleanedName := CleanTitle(animeName)
	if util.IsDebug {
		log.Printf("Attempting AniList search with title: %s", cleanedName)

	}

	query := `
    query ($search: String) {
        Media(search: $search, type: ANIME) {
            id
            title { romaji english }
            description
            genres
            averageScore
            episodes
            status
            idMal
            coverImage { large medium }
        }
    }`

	variables := map[string]interface{}{
		"search": cleanedName,
	}

	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", "https://graphql.anilist.co", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from AniList API: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AniList API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result AniListResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read AniList API response: %v", err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AniList API response: %v", err)
	}

	if result.Data.Media.ID == 0 {
		log.Printf("No results found on AniList for anime: %s", cleanedName)
		return nil, fmt.Errorf("no results found on AniList for anime: %s", cleanedName)
	}

	if util.IsDebug {
		log.Printf("AniList ID: %d, MAL ID: %d, Title: %s, Score: %d, Cover Image URL: %s",
			result.Data.Media.ID, result.Data.Media.IDMal,
			result.Data.Media.Title.Romaji, result.Data.Media.AverageScore,
			result.Data.Media.CoverImage.Large)
	}

	return &result, nil
}

// selectAnimeWithGoFuzzyFinder allows the user to select an anime from a list using fuzzy search
func selectAnimeWithGoFuzzyFinder(animes []Anime) (*Anime, error) {
	if len(animes) == 0 {
		return nil, errors.New("no anime provided")
	}

	sortedAnimes := sortAnimes(animes)
	idx, err := fuzzyfinder.Find(
		sortedAnimes,
		func(i int) string {
			return sortedAnimes[i].Name
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select anime with go-fuzzyfinder")
	}

	if idx < 0 || idx >= len(sortedAnimes) {
		return nil, errors.New("invalid index returned by fuzzyfinder")
	}

	return &sortedAnimes[idx], nil
}

// sortAnimes sorts a list of Anime structs alphabetically by name
func sortAnimes(animeList []Anime) []Anime {
	sort.Slice(animeList, func(i, j int) bool {
		return animeList[i].Name < animeList[j].Name
	})
	return animeList
}

// resolveURL resolves relative URLs to absolute URLs based on the base URL.
func resolveURL(base, ref string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(refURL).String()
}

// CleanTitle removes unwanted words, numbers, and ratings for better API search results
func CleanTitle(title string) string {
	re := regexp.MustCompile(`(?i)(dublado|legendado|todos os episodios)`)
	title = re.ReplaceAllString(title, "")

	re = regexp.MustCompile(`\s+\d+(\.\d+)?\s+A\d+$`)
	title = re.ReplaceAllString(title, "")

	title = strings.TrimSpace(title)
	return title
}

func makeGetRequest(url string, headers map[string]string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed with status %d: %s", resp.StatusCode, body)
	}

	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return responseData, nil
}

func getHTTPResponse(url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
