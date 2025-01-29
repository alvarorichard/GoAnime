package api

import (
	"fmt"
	"github.com/alvarorichard/Goanime/internal/util"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const baseAnimeSearchURL = "https://animefire.plus/pesquisar/"

type GUIAnime struct {
	Name string
	URL  string
}

// ✅ **New Exclusive Function for GUI Anime Search**
func SearchAnimeGUI(animeName string) ([]GUIAnime, error) {
	searchURL := fmt.Sprintf("%s%s", baseAnimeSearchURL, url.PathEscape(animeName))

	if util.IsDebug {
		log.Println("🔍 Searching Anime for GUI:", searchURL)

	}

	// Perform the HTTP request
	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch anime search page")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error: Server returned status %d", resp.StatusCode)
	}

	// Parse HTML response
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse anime search page")
	}

	var animes []GUIAnime
	animeSet := make(map[string]bool) // To prevent duplicates

	// Extract anime results
	doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Text())
		animeURL, exists := s.Attr("href")
		if !exists || name == "" {
			return
		}

		// Convert to absolute URL
		animeURL = resolveURL(baseAnimeSearchURL, animeURL)

		// Avoid duplicates
		if !animeSet[name] {
			animeSet[name] = true
			animes = append(animes, GUIAnime{Name: name, URL: animeURL})
		}
	})

	// Fetch additional details (concurrent requests)
	fetchAnimeDetails(animes)

	if len(animes) == 0 {
		return nil, errors.New("No anime found")
	}

	return animes, nil
}

// ✅ **Parallel fetching of anime details**
func fetchAnimeDetails(animes []GUIAnime) {
	var wg sync.WaitGroup
	wg.Add(len(animes))

	for i := range animes {
		go func(index int) {
			defer wg.Done()
			desc, err := getAnimeDescription(animes[index].URL)
			if err != nil {
				log.Println("❌ Failed to fetch details for:", animes[index].Name)
				return
			}
			animes[index].Name = fmt.Sprintf("%s - %s", animes[index].Name, desc)
		}(i)
	}

	wg.Wait()
}

// ✅ **Fetches additional anime description**
func getAnimeDescription(animeURL string) (string, error) {
	resp, err := http.Get(animeURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to load page: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	description := doc.Find("meta[name='description']").AttrOr("content", "Sem descrição disponível")
	return strings.TrimSpace(description), nil
}

// ✅ **Helper function to resolve relative URLs**
func ResolveURL(base, ref string) string {
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
