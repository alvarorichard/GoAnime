//package api
//
//import (
//	"fmt"
//	"io"
//	"log"
//	"net/http"
//	"net/url"
//	"os"
//	"sort"
//	"strings"
//
//	"github.com/PuerkitoBio/goquery"
//	"github.com/alvarorichard/Goanime/internal/util"
//	"github.com/ktr0731/go-fuzzyfinder"
//	"github.com/pkg/errors"
//)
//
//const baseSiteURL = "https://animefire.plus"
//
//type Anime struct {
//	Name     string
//	URL      string
//	ImageURL string
//	Episodes []Episode
//}
//
//type Episode struct {
//	Number string
//	Num    int
//	URL    string
//}
//
//func SearchAnime(animeName string) (*Anime, error) {
//	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseSiteURL, url.PathEscape(animeName))
//
//	if util.IsDebug {
//		log.Printf("Searching for anime with URL: %s", currentPageURL)
//	}
//
//	for {
//		selectedAnime, nextPageURL, err := searchAnimeOnPage(currentPageURL)
//		if err != nil {
//			return nil, err
//		}
//		if selectedAnime != nil {
//			// After selecting the anime, fetch its details to get the image URL
//			err := FetchAnimeDetails(selectedAnime) // Use FetchAnimeDetails
//			if err != nil {
//				if util.IsDebug {
//					log.Printf("Failed to fetch anime details: %v", err)
//				}
//			}
//			return selectedAnime, nil
//		}
//
//		if nextPageURL == "" {
//			return nil, errors.New("no anime found with the given name")
//		}
//		currentPageURL = baseSiteURL + nextPageURL
//	}
//}
//
//func searchAnimeOnPage(pageURL string) (*Anime, string, error) {
//	response, err := getHTTPResponse(pageURL)
//	if err != nil {
//		return nil, "", errors.Wrap(err, "failed to perform search request")
//	}
//	defer response.Body.Close()
//
//	if response.StatusCode != http.StatusOK {
//		if response.StatusCode == http.StatusForbidden {
//			return nil, "", errors.New("Connection refused: You need to be in Brazil or use a VPN to access the server.")
//		}
//		return nil, "", errors.Errorf("Search failed, the server returned the error: %s", response.Status)
//	}
//
//	doc, err := goquery.NewDocumentFromReader(response.Body)
//	if err != nil {
//		return nil, "", errors.Wrap(err, "failed to parse response")
//	}
//
//	animes := ParseAnimes(doc)
//	if util.IsDebug {
//		log.Printf("Number of animes found: %d", len(animes))
//	}
//
//	if len(animes) > 0 {
//		selectedAnime, err := selectAnimeWithGoFuzzyFinder(animes)
//		if err != nil {
//			return nil, "", err
//		}
//		return selectedAnime, "", nil
//	}
//
//	// Update the selector for the next page link if necessary
//	nextPage, exists := doc.Find(".pagination .next a").Attr("href")
//	if !exists {
//		return nil, "", nil
//	}
//
//	return nil, nextPage, nil
//}
//
//// ParseAnimes extracts a list of Anime structs from the search results page.
//func ParseAnimes(doc *goquery.Document) []Anime {
//	var animes []Anime
//
//	// Update the selectors based on the actual HTML structure of the website
//	doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
//		// Extract the URL
//		urlPath, exists := s.Attr("href")
//		if !exists {
//			return
//		}
//		url := resolveURL(baseSiteURL, urlPath)
//
//		// Extract the name
//		name := strings.TrimSpace(s.Text())
//
//		if util.IsDebug {
//			log.Printf("Parsed Anime - Name: %s, URL: %s", name, url)
//		}
//
//		animes = append(animes, Anime{
//			Name: name,
//			URL:  url,
//		})
//	})
//
//	return animes
//}
//
////func fetchAnimeDetails(anime *Anime) error {
////	if util.IsDebug {
////		log.Printf("Fetching details for anime: %s, URL: %s", anime.Name, anime.URL)
////	}
////
////	response, err := getHTTPResponse(anime.URL)
////	if err != nil {
////		return errors.Wrap(err, "failed to get anime details page")
////	}
////	defer response.Body.Close()
////
////	if response.StatusCode != http.StatusOK {
////		return errors.Errorf("Failed to get anime details page: %s", response.Status)
////	}
////
////	bodyBytes, err := io.ReadAll(response.Body)
////	if err != nil {
////		return errors.Wrap(err, "failed to read response body")
////	}
////
////	if util.IsDebug {
////		log.Printf("Anime detail page HTML:\n%s", string(bodyBytes))
////	}
////
////	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
////	if err != nil {
////		return errors.Wrap(err, "failed to parse anime details page")
////	}
////
////	// Try using meta tag
////	metaImage := doc.Find(`meta[property="og:image"]`)
////	if metaImage.Length() > 0 {
////		imageURL, exists := metaImage.Attr("content")
////		if exists && imageURL != "" {
////			imageURL = resolveURL(baseSiteURL, imageURL)
////			anime.ImageURL = imageURL
////			if util.IsDebug {
////				log.Printf("Cover image URL from meta tag: %s", anime.ImageURL)
////			}
////			// Proceed to download the image
////			err = downloadMedia(anime.ImageURL, "cover")
////			if err != nil {
////				return errors.Wrap(err, "failed to download cover image")
////			}
////			return nil
////		}
////	}
////
////	// Fallback to previous method with adjusted selector
////	imageSelection := doc.Find(".anime-poster img") // Adjust this selector
////
////	if imageSelection.Length() == 0 {
////		if util.IsDebug {
////			log.Printf("Cover image element not found in anime details page.")
////		}
////		return errors.New("cover image element not found")
////	}
////
////	imageURL, exists := imageSelection.Attr("src")
////	if !exists || imageURL == "" {
////		return errors.New("cover image URL not found")
////	}
////
////	imageURL = resolveURL(baseSiteURL, imageURL)
////	anime.ImageURL = imageURL
////
////	if util.IsDebug {
////		log.Printf("Cover image URL: %s", anime.ImageURL)
////	}
////
////	// Now download the image
////	err = downloadMedia(anime.ImageURL, "cover")
////	if err != nil {
////		return errors.Wrap(err, "failed to download cover image")
////	}
////
////	return nil
////}
//
//func FetchAnimeDetails(anime *Anime) error {
//	response, err := http.Get(anime.URL)
//	if err != nil {
//		return errors.Wrap(err, "failed to get anime details page")
//	}
//	defer response.Body.Close()
//
//	if response.StatusCode != http.StatusOK {
//		return fmt.Errorf("failed to get anime details page: %s", response.Status)
//	}
//
//	// Parse the HTML to find the cover image URL
//	doc, err := goquery.NewDocumentFromReader(response.Body)
//	if err != nil {
//		return errors.Wrap(err, "failed to parse anime details page")
//	}
//
//	// Try to find the cover image from a meta tag or another selector
//	imageURL, exists := doc.Find(`meta[property="og:image"]`).Attr("content")
//	if !exists || imageURL == "" {
//		return errors.New("cover image URL not found")
//	}
//
//	anime.ImageURL = imageURL // Store the cover image URL in the Anime struct
//
//	// Optionally, download the cover image
//	err = downloadMedia(imageURL, "cover")
//	if err != nil {
//		return errors.Wrap(err, "failed to download cover image")
//	}
//
//	log.Printf("Cover image URL set for anime: %s", anime.Name)
//	return nil
//}
//
//// downloadMedia downloads the media from a URL and saves it as cover.webp
//func downloadMedia(mediaURL, filename string) error {
//	resp, err := http.Get(mediaURL)
//	if err != nil {
//		return errors.Wrap(err, "failed to download media")
//	}
//	defer resp.Body.Close()
//
//	out, err := os.Create(fmt.Sprintf("%s.webp", filename))
//	if err != nil {
//		return errors.Wrap(err, "failed to create media file")
//	}
//	defer out.Close()
//
//	_, err = io.Copy(out, resp.Body)
//	if err != nil {
//		return errors.Wrap(err, "failed to save media to file")
//	}
//
//	log.Printf("Media saved as: %s.webp", filename)
//	return nil
//}
//func selectAnimeWithGoFuzzyFinder(animes []Anime) (*Anime, error) {
//	if len(animes) == 0 {
//		return nil, errors.New("no anime provided")
//	}
//
//	sortedAnimes := sortAnimes(animes)
//	idx, err := fuzzyfinder.Find(
//		sortedAnimes,
//		func(i int) string {
//			return sortedAnimes[i].Name
//		},
//	)
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to select anime with go-fuzzyfinder")
//	}
//
//	if idx < 0 || idx >= len(sortedAnimes) {
//		return nil, errors.New("invalid index returned by fuzzyfinder")
//	}
//
//	return &sortedAnimes[idx], nil
//}
//
//func sortAnimes(animeList []Anime) []Anime {
//	sort.Slice(animeList, func(i, j int) bool {
//		return animeList[i].Name < animeList[j].Name
//	})
//	return animeList
//}
//
////// downloadMedia downloads a media file from the given URL and saves it to the specified file path.
////func downloadMedia(mediaURL, filename string) error {
////	if mediaURL == "" {
////		return errors.New("media URL is empty")
////	}
////
////	resp, err := getHTTPResponse(mediaURL)
////	if err != nil {
////		return errors.Wrap(err, "failed to download media")
////	}
////	defer resp.Body.Close()
////
////	// Determine the file extension
////	u, err := url.Parse(mediaURL)
////	if err != nil {
////		return errors.Wrap(err, "invalid media URL")
////	}
////	ext := path.Ext(u.Path)
////	filepath := filename + ext
////
////	out, err := os.Create(filepath)
////	if err != nil {
////		return errors.Wrap(err, "failed to create media file")
////	}
////	defer out.Close()
////
////	_, err = io.Copy(out, resp.Body)
////	if err != nil {
////		return errors.Wrap(err, "failed to save media to file")
////	}
////
////	if util.IsDebug {
////		log.Printf("Media saved to: %s", filepath)
////	}
////
////	return nil
////}
//
//// resolveURL resolves relative URLs to absolute URLs based on the base URL.
//func resolveURL(base, ref string) string {
//	baseURL, err := url.Parse(base)
//	if err != nil {
//		return ""
//	}
//	refURL, err := url.Parse(ref)
//	if err != nil {
//		return ""
//	}
//	return baseURL.ResolveReference(refURL).String()
//}
//
//// getHTTPResponse sends an HTTP GET request with headers and returns the response.
//func getHTTPResponse(url string) (*http.Response, error) {
//	client := &http.Client{}
//	req, err := http.NewRequest("GET", url, nil)
//	if err != nil {
//		return nil, err
//	}
//	// Set the User-Agent header to mimic a regular browser
//	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
//	resp, err := client.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	return resp, nil
//}
//
//// GetAnimeEpisodes retrieves the list of episodes for a given anime.

//package api
//
//import (
//	"encoding/json"
//	"fmt"
//	"io"
//	"log"
//	"net/http"
//	"net/url"
//	"os"
//	"regexp"
//	"sort"
//	"strings"
//
//	"github.com/PuerkitoBio/goquery"
//	"github.com/alvarorichard/Goanime/internal/util"
//	"github.com/ktr0731/go-fuzzyfinder"
//	"github.com/pkg/errors"
//)
//
//const baseSiteURL = "https://animefire.plus"
//
//type Anime struct {
//	Name      string
//	URL       string
//	ImageURL  string
//	Episodes  []Episode
//	AnilistID int
//	Details   AniListDetails
//}
//
//type Episode struct {
//	Number string
//	Num    int
//	URL    string
//}
//
//type AniListResponse struct {
//	Data struct {
//		Media AniListDetails `json:"Media"`
//	} `json:"data"`
//}
//
//type AniListDetails struct {
//	ID           int      `json:"id"`
//	Title        Title    `json:"title"`
//	Description  string   `json:"description"`
//	Genres       []string `json:"genres"`
//	AverageScore int      `json:"averageScore"`
//	Episodes     int      `json:"episodes"`
//	Status       string   `json:"status"`
//}
//
//type Title struct {
//	Romaji  string `json:"romaji"`
//	English string `json:"english"`
//}
//
//// SearchAnime searches for an anime on AnimeFire and fetches additional information from AniList API
//func SearchAnime(animeName string) (*Anime, error) {
//	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseSiteURL, url.PathEscape(animeName))
//
//	if util.IsDebug {
//		log.Printf("Searching for anime with URL: %s", currentPageURL)
//	}
//
//	for {
//		selectedAnime, nextPageURL, err := searchAnimeOnPage(currentPageURL)
//		if err != nil {
//			return nil, err
//		}
//		if selectedAnime != nil {
//			// After selecting the anime, fetch its details to get the image URL
//			err := FetchAnimeDetails(selectedAnime)
//			if err != nil {
//				log.Printf("Failed to fetch anime details: %v", err)
//			}
//
//			// Fetch additional details from AniList API
//			aniListInfo, err := FetchAnimeFromAniList(selectedAnime.Name)
//			if err != nil {
//				log.Printf("Error fetching additional data from AniList: %v", err)
//			} else {
//				selectedAnime.AnilistID = aniListInfo.Data.Media.ID
//				selectedAnime.Details = aniListInfo.Data.Media
//				log.Printf("AniList ID: %d, Title: %s, Score: %d", aniListInfo.Data.Media.ID, aniListInfo.Data.Media.Title.Romaji, aniListInfo.Data.Media.AverageScore)
//			}
//
//			return selectedAnime, nil
//		}
//
//		if nextPageURL == "" {
//			return nil, errors.New("no anime found with the given name")
//		}
//		currentPageURL = baseSiteURL + nextPageURL
//	}
//}
//
//func searchAnimeOnPage(pageURL string) (*Anime, string, error) {
//	response, err := getHTTPResponse(pageURL)
//	if err != nil {
//		return nil, "", errors.Wrap(err, "failed to perform search request")
//	}
//	defer response.Body.Close()
//
//	if response.StatusCode != http.StatusOK {
//		if response.StatusCode == http.StatusForbidden {
//			return nil, "", errors.New("Connection refused: You need to be in Brazil or use a VPN to access the server.")
//		}
//		return nil, "", errors.Errorf("Search failed, the server returned the error: %s", response.Status)
//	}
//
//	doc, err := goquery.NewDocumentFromReader(response.Body)
//	if err != nil {
//		return nil, "", errors.Wrap(err, "failed to parse response")
//	}
//
//	animes := ParseAnimes(doc)
//	if util.IsDebug {
//		log.Printf("Number of animes found: %d", len(animes))
//	}
//
//	if len(animes) > 0 {
//		selectedAnime, err := selectAnimeWithGoFuzzyFinder(animes)
//		if err != nil {
//			return nil, "", err
//		}
//		return selectedAnime, "", nil
//	}
//
//	nextPage, exists := doc.Find(".pagination .next a").Attr("href")
//	if !exists {
//		return nil, "", nil
//	}
//
//	return nil, nextPage, nil
//}
//
//// ParseAnimes extracts a list of Anime structs from the search results page.
//func ParseAnimes(doc *goquery.Document) []Anime {
//	var animes []Anime
//
//	doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
//		urlPath, exists := s.Attr("href")
//		if !exists {
//			return
//		}
//		url := resolveURL(baseSiteURL, urlPath)
//
//		name := strings.TrimSpace(s.Text())
//
//		if util.IsDebug {
//			log.Printf("Parsed Anime - Name: %s, URL: %s", name, url)
//		}
//
//		animes = append(animes, Anime{
//			Name: name,
//			URL:  url,
//		})
//	})
//
//	return animes
//}
//
//// FetchAnimeDetails retrieves additional information for the selected anime
//func FetchAnimeDetails(anime *Anime) error {
//	response, err := http.Get(anime.URL)
//	if err != nil {
//		return errors.Wrap(err, "failed to get anime details page")
//	}
//	defer response.Body.Close()
//
//	if response.StatusCode != http.StatusOK {
//		return fmt.Errorf("failed to get anime details page: %s", response.Status)
//	}
//
//	doc, err := goquery.NewDocumentFromReader(response.Body)
//	if err != nil {
//		return errors.Wrap(err, "failed to parse anime details page")
//	}
//
//	imageURL, exists := doc.Find(`meta[property="og:image"]`).Attr("content")
//	if !exists || imageURL == "" {
//		return errors.New("cover image URL not found")
//	}
//
//	anime.ImageURL = imageURL
//	err = downloadMedia(imageURL, "cover")
//	if err != nil {
//		return errors.Wrap(err, "failed to download cover image")
//	}
//
//	log.Printf("Cover image URL set for anime: %s", anime.Name)
//	return nil
//}
//
//// FetchAnimeFromAniList retrieves additional information from the AniList API
////func FetchAnimeFromAniList(animeName string) (*AniListResponse, error) {
////	query := `
////	query ($search: String) {
////		Media(search: $search, type: ANIME) {
////			id
////			title { romaji english }
////			description
////			genres
////			averageScore
////			episodes
////			status
////		}
////	}`
////	variables := fmt.Sprintf(`{"search": "%s"}`, animeName)
////	jsonData := fmt.Sprintf(`{"query": %q, "variables": %s}`, query, variables)
////
////	req, err := http.NewRequest("POST", "https://graphql.anilist.co", strings.NewReader(jsonData))
////	if err != nil {
////		return nil, err
////	}
////	req.Header.Set("Content-Type", "application/json")
////
////	client := &http.Client{}
////	resp, err := client.Do(req)
////	if err != nil {
////		return nil, err
////	}
////	defer resp.Body.Close()
////
////	var result AniListResponse
////	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
////		return nil, err
////	}
////	return &result, nil
////}
//
//func CleanTitle(title string) string {
//	// Remove terms comuns que podem interferir na busca
//	re := regexp.MustCompile(`(?i)(dublado|legendado|todos os episodios)`)
//	title = re.ReplaceAllString(title, "")
//
//	// Remover classificações e códigos de idade, como "7.12  A14"
//	re = regexp.MustCompile(`\s+\d+(\.\d+)?\s+A\d+$`)
//	title = re.ReplaceAllString(title, "")
//
//	title = strings.TrimSpace(title)
//	return title
//}
//
//func FetchAnimeFromAniList(animeName string) (*AniListResponse, error) {
//	cleanedName := CleanTitle(animeName)
//	log.Printf("Attempting AniList search with title: %s", cleanedName)
//
//	query := `
//	query ($search: String) {
//		Media(search: $search, type: ANIME) {
//			id
//			title { romaji english }
//			description
//			genres
//			averageScore
//			episodes
//			status
//		}
//	}`
//	variables := fmt.Sprintf(`{"search": "%s"}`, cleanedName)
//	jsonData := fmt.Sprintf(`{"query": %q, "variables": %s}`, query, variables)
//
//	req, err := http.NewRequest("POST", "https://graphql.anilist.co", strings.NewReader(jsonData))
//	if err != nil {
//		return nil, fmt.Errorf("failed to create request: %v", err)
//	}
//	req.Header.Set("Content-Type", "application/json")
//
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return nil, fmt.Errorf("failed to fetch data from AniList API: %v", err)
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		body, _ := io.ReadAll(resp.Body)
//		return nil, fmt.Errorf("AniList API request failed with status %d: %s", resp.StatusCode, string(body))
//	}
//
//	var result AniListResponse
//	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
//		return nil, fmt.Errorf("failed to parse AniList API response: %v", err)
//	}
//
//	if result.Data.Media.ID == 0 {
//		log.Printf("No results found on AniList for anime: %s", cleanedName)
//		return nil, fmt.Errorf("no results found on AniList for anime: %s", cleanedName)
//	}
//
//	log.Printf("AniList ID: %d, Title: %s, Score: %d", result.Data.Media.ID, result.Data.Media.Title.Romaji, result.Data.Media.AverageScore)
//	return &result, nil
//}
//
//// downloadMedia downloads the media from a URL and saves it as cover.webp
//func downloadMedia(mediaURL, filename string) error {
//	resp, err := http.Get(mediaURL)
//	if err != nil {
//		return errors.Wrap(err, "failed to download media")
//	}
//	defer resp.Body.Close()
//
//	out, err := os.Create(fmt.Sprintf("%s.webp", filename))
//	if err != nil {
//		return errors.Wrap(err, "failed to create media file")
//	}
//	defer out.Close()
//
//	_, err = io.Copy(out, resp.Body)
//	if err != nil {
//		return errors.Wrap(err, "failed to save media to file")
//	}
//
//	log.Printf("Media saved as: %s.webp", filename)
//	return nil
//}
//
//func selectAnimeWithGoFuzzyFinder(animes []Anime) (*Anime, error) {
//	if len(animes) == 0 {
//		return nil, errors.New("no anime provided")
//	}
//
//	sortedAnimes := sortAnimes(animes)
//	idx, err := fuzzyfinder.Find(
//		sortedAnimes,
//		func(i int) string {
//			return sortedAnimes[i].Name
//		},
//	)
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to select anime with go-fuzzyfinder")
//	}
//
//	if idx < 0 || idx >= len(sortedAnimes) {
//		return nil, errors.New("invalid index returned by fuzzyfinder")
//	}
//
//	return &sortedAnimes[idx], nil
//}
//
//func sortAnimes(animeList []Anime) []Anime {
//	sort.Slice(animeList, func(i, j int) bool {
//		return animeList[i].Name < animeList[j].Name
//	})
//	return animeList
//}
//
//func resolveURL(base, ref string) string {
//	baseURL, err := url.Parse(base)
//	if err != nil {
//		return ""
//	}
//	refURL, err := url.Parse(ref)
//	if err != nil {
//		return ""
//	}
//	return baseURL.ResolveReference(refURL).String()
//}
//
//func getHTTPResponse(url string) (*http.Response, error) {
//	client := &http.Client{}
//	req, err := http.NewRequest("GET", url, nil)
//	if err != nil {
//		return nil, err
//	}
//	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
//	resp, err := client.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	return resp, nil
//}

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
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
	Number   string
	Num      int
	URL      string
	Title    TitleDetails
	Aired    string
	Duration int
	IsFiller bool
	IsRecap  bool
	Synopsis string
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

// // SearchAnime searches for an anime on AnimeFire and fetches additional information from AniList API
//
//	func SearchAnime(animeName string) (*Anime, error) {
//		currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseSiteURL, url.PathEscape(animeName))
//
//		if util.IsDebug {
//			log.Printf("Searching for anime with URL: %s", currentPageURL)
//		}
//
//		for {
//			selectedAnime, nextPageURL, err := searchAnimeOnPage(currentPageURL)
//			if err != nil {
//				return nil, err
//			}
//			if selectedAnime != nil {
//				// After selecting the anime, fetch its details to get the image URL
//				err := FetchAnimeDetails(selectedAnime)
//				if err != nil {
//					log.Printf("Failed to fetch anime details: %v", err)
//				}
//
//				// Fetch additional details from AniList API
//				aniListInfo, err := FetchAnimeFromAniList(selectedAnime.Name)
//				if err != nil {
//					log.Printf("Error fetching additional data from AniList: %v", err)
//				} else {
//					selectedAnime.AnilistID = aniListInfo.Data.Media.ID
//					selectedAnime.Details = aniListInfo.Data.Media
//					log.Printf("AniList ID: %d, Title: %s, Score: %d", aniListInfo.Data.Media.ID, aniListInfo.Data.Media.Title.Romaji, aniListInfo.Data.Media.AverageScore)
//				}
//
//				return selectedAnime, nil
//			}
//
//			if nextPageURL == "" {
//				return nil, errors.New("no anime found with the given name")
//			}
//			currentPageURL = baseSiteURL + nextPageURL
//		}
//	}
//
// // GetEpisodeData fetches episode data for a given anime ID and episode number from Jikan API
//
//	func GetEpisodeData(animeID int, episodeNo int, anime *Anime) error {
//		url := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d/episodes/%d", animeID, episodeNo)
//
//		response, err := makeGetRequest(url, nil)
//		if err != nil {
//			return fmt.Errorf("error fetching data from Jikan (MyAnimeList) API: %w", err)
//		}
//
//		data, ok := response["data"].(map[string]interface{})
//		if !ok {
//			return fmt.Errorf("invalid response structure: missing or invalid 'data' field")
//		}
//
//		// Helper functions to safely get values
//		getStringValue := func(field string) string {
//			if value, ok := data[field].(string); ok {
//				return value
//			}
//			return ""
//		}
//
//		getIntValue := func(field string) int {
//			if value, ok := data[field].(float64); ok {
//				return int(value)
//			}
//			return 0
//		}
//
//		getBoolValue := func(field string) bool {
//			if value, ok := data[field].(bool); ok {
//				return value
//			}
//			return false
//		}
//
//		// Assign values to the Anime struct
//		anime.Episodes[episodeNo-1].Title.Romaji = getStringValue("title_romanji")
//		anime.Episodes[episodeNo-1].Title.English = getStringValue("title")
//		anime.Episodes[episodeNo-1].Title.Japanese = getStringValue("title_japanese")
//		anime.Episodes[episodeNo-1].Aired = getStringValue("aired")
//		anime.Episodes[episodeNo-1].Duration = getIntValue("duration")
//		anime.Episodes[episodeNo-1].IsFiller = getBoolValue("filler")
//		anime.Episodes[episodeNo-1].IsRecap = getBoolValue("recap")
//		anime.Episodes[episodeNo-1].Synopsis = getStringValue("synopsis")
//
//		return nil
//	}
//

// FetchAnimeCoverFromJikan fetches cover image URL using Jikan API
func FetchAnimeCoverFromJikan(malID int) (string, error) {
	url := fmt.Sprintf("https://api.jikan.moe/v4/anime/%d", malID)

	response, err := makeGetRequest(url, nil)
	if err != nil {
		return "", fmt.Errorf("error fetching data from Jikan API: %w", err)
	}

	// Parse the JSON response to retrieve the image URL
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response structure: missing 'data' field")
	}

	imageURL, ok := data["images"].(map[string]interface{})["jpg"].(map[string]interface{})["image_url"].(string)
	if !ok || imageURL == "" {
		return "", fmt.Errorf("cover image URL not found in Jikan response")
	}

	return imageURL, nil
}

//// SearchAnime searches for an anime on AnimeFire and fetches additional information from AniList API
//func SearchAnime(animeName string) (*Anime, error) {
//	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseSiteURL, url.PathEscape(animeName))
//
//	if util.IsDebug {
//		log.Printf("Searching for anime with URL: %s", currentPageURL)
//	}
//
//	for {
//		selectedAnime, nextPageURL, err := searchAnimeOnPage(currentPageURL)
//		if err != nil {
//			return nil, err
//		}
//		if selectedAnime != nil {
//			// After selecting the anime, fetch its details to get the image URL
//			err := FetchAnimeDetails(selectedAnime)
//			if err != nil {
//				log.Printf("Failed to fetch anime details: %v", err)
//			}
//
//			// Fetch additional details from AniList API
//			aniListInfo, err := FetchAnimeFromAniList(selectedAnime.Name)
//			if err != nil {
//				log.Printf("Error fetching additional data from AniList: %v", err)
//			} else {
//				selectedAnime.AnilistID = aniListInfo.Data.Media.ID
//				selectedAnime.MalID = aniListInfo.Data.Media.IDMal // Armazene o ID do MAL para usar com a Jikan API
//				selectedAnime.Details = aniListInfo.Data.Media
//				log.Printf("AniList ID: %d, MAL ID: %d, Title: %s, Score: %d",
//					aniListInfo.Data.Media.ID, aniListInfo.Data.Media.IDMal,
//					aniListInfo.Data.Media.Title.Romaji, aniListInfo.Data.Media.AverageScore)
//			}
//
//			// Fetch episode data from Jikan API for the first episode as an example
//			if selectedAnime.MalID > 0 {
//				err = GetEpisodeData(selectedAnime.MalID, 1, selectedAnime)
//				if err != nil {
//					log.Printf("Error fetching episode data from Jikan API: %v", err)
//				} else if util.IsDebug && len(selectedAnime.Episodes) > 0 {
//					// Print detailed episode information in debug mode
//					firstEpisode := selectedAnime.Episodes[0]
//					log.Printf("Jikan Episode Data - Title: %s, Aired: %s, Duration: %d, Filler: %t, Synopsis: %s",
//						firstEpisode.Title.English,
//						firstEpisode.Aired,
//						firstEpisode.Duration,
//						firstEpisode.IsFiller,
//						firstEpisode.Synopsis)
//				}
//			} else {
//				log.Printf("No MAL ID found for anime: %s. Unable to fetch episode data from Jikan API.", selectedAnime.Name)
//			}
//
//			return selectedAnime, nil
//		}
//
//		if nextPageURL == "" {
//			return nil, errors.New("no anime found with the given name")
//		}
//		currentPageURL = baseSiteURL + nextPageURL
//	}
//
//}

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
					log.Printf("Cover image URL retrieved from AniList: %s", selectedAnime.ImageURL)
				} else {
					log.Printf("Cover image URL not found in AniList response for anime: %s", selectedAnime.Name)
				}

				log.Printf("AniList ID: %d, MAL ID: %d, Title: %s, Score: %d, Cover Image URL: %s",
					aniListInfo.Data.Media.ID, aniListInfo.Data.Media.IDMal,
					aniListInfo.Data.Media.Title.Romaji, aniListInfo.Data.Media.AverageScore,
					selectedAnime.ImageURL)
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
	defer response.Body.Close()

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
	defer response.Body.Close()

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

	anime.ImageURL = imageURL
	err = downloadMedia(imageURL, "cover")
	if err != nil {
		return errors.Wrap(err, "failed to download cover image")
	}

	log.Printf("Cover image URL set for anime: %s", anime.Name)
	return nil
}

//// FetchAnimeFromAniList retrieves additional information from the AniList API
//func FetchAnimeFromAniList(animeName string) (*AniListResponse, error) {
//	cleanedName := CleanTitle(animeName)
//	log.Printf("Attempting AniList search with title: %s", cleanedName)
//
//	query := `
//	query ($search: String) {
//		Media(search: $search, type: ANIME) {
//			id
//			title { romaji english }
//			description
//			genres
//			averageScore
//			episodes
//			status
//		}
//	}`
//	variables := fmt.Sprintf(`{"search": "%s"}`, cleanedName)
//	jsonData := fmt.Sprintf(`{"query": %q, "variables": %s}`, query, variables)
//
//	req, err := http.NewRequest("POST", "https://graphql.anilist.co", strings.NewReader(jsonData))
//	if err != nil {
//		return nil, fmt.Errorf("failed to create request: %v", err)
//	}
//	req.Header.Set("Content-Type", "application/json")
//
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return nil, fmt.Errorf("failed to fetch data from AniList API: %v", err)
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		body, _ := io.ReadAll(resp.Body)
//		return nil, fmt.Errorf("AniList API request failed with status %d: %s", resp.StatusCode, string(body))
//	}
//
//	var result AniListResponse
//	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
//		return nil, fmt.Errorf("failed to parse AniList API response: %v", err)
//	}
//
//	if result.Data.Media.ID == 0 {
//		log.Printf("No results found on AniList for anime: %s", cleanedName)
//		return nil, fmt.Errorf("no results found on AniList for anime: %s", cleanedName)
//	}
//
//	log.Printf("AniList ID: %d, Title: %s, Score: %d", result.Data.Media.ID, result.Data.Media.Title.Romaji, result.Data.Media.AverageScore)
//	return &result, nil
//}

//// FetchAnimeFromAniList retrieves additional information from the AniList API
//func FetchAnimeFromAniList(animeName string) (*AniListResponse, error) {
//	cleanedName := CleanTitle(animeName)
//	log.Printf("Attempting AniList search with title: %s", cleanedName)
//
//	query := `
//	query ($search: String) {
//		Media(search: $search, type: ANIME) {
//			id
//			idMal    # Obtenha o MAL ID para uso com a API Jikan
//			title { romaji english }
//			description
//			genres
//			averageScore
//			episodes
//			status
//		}
//	}`
//	variables := fmt.Sprintf(`{"search": "%s"}`, cleanedName)
//	jsonData := fmt.Sprintf(`{"query": %q, "variables": %s}`, query, variables)
//
//	req, err := http.NewRequest("POST", "https://graphql.anilist.co", strings.NewReader(jsonData))
//	if err != nil {
//		return nil, fmt.Errorf("failed to create request: %v", err)
//	}
//	req.Header.Set("Content-Type", "application/json")
//
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return nil, fmt.Errorf("failed to fetch data from AniList API: %v", err)
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		body, _ := io.ReadAll(resp.Body)
//		return nil, fmt.Errorf("AniList API request failed with status %d: %s", resp.StatusCode, string(body))
//	}
//
//	var result AniListResponse
//	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
//		return nil, fmt.Errorf("failed to parse AniList API response: %v", err)
//	}
//
//	if result.Data.Media.ID == 0 {
//		log.Printf("No results found on AniList for anime: %s", cleanedName)
//		return nil, fmt.Errorf("no results found on AniList for anime: %s", cleanedName)
//	}
//
//	if util.IsDebug {
//		log.Printf("AniList ID: %d, MAL ID: %d, Title: %s, Score: %d",
//			result.Data.Media.ID, result.Data.Media.IDMal, result.Data.Media.Title.Romaji, result.Data.Media.AverageScore)
//
//	}
//	return &result, nil
//}

func FetchAnimeFromAniList(animeName string) (*AniListResponse, error) {
	cleanedName := CleanTitle(animeName)
	log.Printf("Attempting AniList search with title: %s", cleanedName)

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
	defer resp.Body.Close()

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

	log.Printf("AniList ID: %d, MAL ID: %d, Title: %s, Score: %d, Cover Image URL: %s",
		result.Data.Media.ID, result.Data.Media.IDMal,
		result.Data.Media.Title.Romaji, result.Data.Media.AverageScore,
		result.Data.Media.CoverImage.Large)

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

// downloadMedia downloads the media from a URL and saves it as cover.webp
func downloadMedia(mediaURL, filename string) error {
	resp, err := http.Get(mediaURL)
	if err != nil {
		return errors.Wrap(err, "failed to download media")
	}
	defer resp.Body.Close()

	out, err := os.Create(fmt.Sprintf("%s.webp", filename))
	if err != nil {
		return errors.Wrap(err, "failed to create media file")
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to save media to file")
	}

	if util.IsDebug {
		log.Printf("Media saved as: %s.webp", filename)
	}

	return nil
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
	defer resp.Body.Close()

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
