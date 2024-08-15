package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/errors"
)

const baseSiteURL = "https://animefire.plus/"

type Anime struct {
	Name     string
	URL      string
	Episodes []Episode
}

type Episode struct {
	Number string
	Num    int
	URL    string
}

// SearchAnime searches for an anime on a specific site based on the provided name.
// If the anime is found, it returns the URL of the anime's page. Otherwise, it returns an error.
//
// Parameters:
// - animeName: the name of the anime to search for.
//
// Returns:
// - string: the URL of the found anime's page.
// - error: an error if the anime is not found or if there is an issue during the search.
func SearchAnime(animeName string) (string, error) {
	// Construct the URL for the search page using the provided anime name.
	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseSiteURL, animeName)

	// Loop through the search results pages until the anime is found or no more pages are left.
	if util.IsDebug {
		log.Printf("Searching for anime with URL: %s", currentPageURL)
	}

	for {
		// Perform the search on the current page and get the anime URL (if found) or the URL for the next page.
		animeURL, nextPageURL, err := searchAnimeOnPage(currentPageURL)
		if err != nil {
			// Return an error if there is an issue with the search.
			return "", err
		}
		// If an anime URL is found, return this URL.
		if animeURL != "" {
			return animeURL, nil
		}

		// If there are no more pages to search, return an error indicating that the anime was not found.
		if nextPageURL == "" {
			return "", errors.New("no anime found with the given name")
		}
		// Update the current page URL to the next page of search results.
		currentPageURL = baseSiteURL + nextPageURL
	}
}

// searchAnimeOnPage performs the search for an anime on a given page URL.
// It parses the page to find the anime and returns the anime's URL if found, the URL of the next page if available,
// or an error if something goes wrong.
//
// Parameters:
// - url: the URL of the search results page to be queried.
//
// Returns:
// - string: the URL of the found anime.
// - string: the URL of the next page of search results, if available.
// - error: an error if the search fails or if there is an issue processing the page.
func searchAnimeOnPage(url string) (string, string, error) {
	// Send an HTTP GET request to the specified URL.
	response, err := http.Get(url)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to perform search request")
	}
	// Ensure the response body is closed after the function finishes, and log an error if closing fails.
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}(response.Body)

	// Check if the request was successful (status code 200 OK).
	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusForbidden {
			// If the server responds with 403 Forbidden, provide a specific error message.
			return "", "", errors.New("Connection refused: You need to be in Brazil or use a VPN to access the server.")
		}
		// Return an error if the status code is anything other than 200 OK.
		return "", "", errors.Errorf("Search failed, the server returned the error: %s", response.Status)
	}

	// Parse the HTML response using goquery.
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to parse response")
	}

	// Extract the list of animes from the parsed HTML document.
	animes := ParseAnimes(doc)
	if len(animes) > 0 {
		// If animes are found, prompt the user to select one using go-fuzzyfinder.
		selectedAnimeName, err := selectAnimeWithGoFuzzyFinder(animes)
		if err != nil {
			return "", "", err
		}
		// Return the URL of the selected anime if found.
		for _, anime := range animes {
			if anime.Name == selectedAnimeName {
				return anime.URL, "", nil
			}
		}
	}

	// Check if there is a next page of search results.
	nextPage, exists := doc.Find(".pagination .next a").Attr("href")
	if !exists {
		// If no next page is found, return nil for the nextPage URL.
		return "", "", nil
	}

	// Return an empty string for anime URL and the next page URL for further searching.
	return "", nextPage, nil
}

// sortAnimes sorts a list of Anime structs alphabetically by their Name field.
// It returns the sorted slice of Anime.
//
// Parameters:
// - animeList: a slice of Anime structs to be sorted.
//
// Returns:
// - []Anime: the sorted slice of Anime structs.
func sortAnimes(animeList []Anime) []Anime {
	// Sort the slice of Anime structs in place using the sort.Slice function.
	// The sorting is done based on the Name field of each Anime struct in alphabetical order.
	sort.Slice(animeList, func(i, j int) bool {
		return animeList[i].Name < animeList[j].Name
	})

	// Return the sorted slice.
	return animeList
}

// ParseAnimes extracts a list of Anime structs from the given goquery.Document.
// It looks for specific HTML elements that contain anime information and returns a slice of Anime structs.
//
// Parameters:
// - doc: a pointer to a goquery.Document which represents the parsed HTML content.
//
// Returns:
// - []Anime: a slice of Anime structs extracted from the HTML document.
func ParseAnimes(doc *goquery.Document) []Anime {
	// Initialize an empty slice to hold the Anime structs.
	var animes []Anime

	// Find all anchor elements within the specified CSS selector and iterate over them.
	doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
		// Extract the text (name of the anime) and the href attribute (URL) from each anchor element,
		// trim any surrounding whitespace, and create an Anime struct which is appended to the slice.
		animes = append(animes, Anime{
			Name: strings.TrimSpace(s.Text()),
			URL:  s.AttrOr("href", ""),
		})
	})

	// Return the slice of Anime structs.
	return animes
}

// selectAnimeWithGoFuzzyFinder allows the user to select an anime from a list using a fuzzy finder interface.
// It returns the name of the selected anime or an error if the selection fails.
//
// Parameters:
// - animes: a slice of Anime structs from which the user can choose.
//
// Returns:
// - string: the name of the selected anime.
// - error: an error if no animes are provided, the selection fails, or the selection index is invalid.
func selectAnimeWithGoFuzzyFinder(animes []Anime) (string, error) {
	// Check if the anime list is empty. If so, return an error.
	if len(animes) == 0 {
		return "", errors.New("no anime provided")
	}

	// Create a slice to hold the names of the animes for display in the fuzzy finder.
	animeNames := make([]string, len(animes))

	// Sort the animes alphabetically by name and populate the animeNames slice.
	for i, anime := range sortAnimes(animes) {
		animeNames[i] = anime.Name
	}

	// Use the fuzzyfinder library to allow the user to select an anime by name.
	idx, err := fuzzyfinder.Find(
		animeNames,
		func(i int) string {
			return animeNames[i]
		},
	)
	if err != nil {
		// Return an error if the fuzzy finder fails.
		return "", errors.Wrap(err, "failed to select anime with go-fuzzyfinder")
	}

	// Check if the selected index is out of bounds. If so, return an error.
	if idx < 0 || idx >= len(animes) {
		return "", errors.New("invalid index returned by fuzzyfinder")
	}

	// Return the name of the selected anime.
	return animes[idx].Name, nil
}
