package api

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const baseAnimeSearchURL = "https://animefire.plus/pesquisar/"

type GUIAnime struct {
	Name     string
	URL      string
	Episodes []GUIEpisode
}

type GUIEpisode struct {
	Number string
	URL    string
}

// ✅ **New GUI Search Function**
func SearchAnimeGUI(animeName string) ([]GUIAnime, error) {
	searchURL := fmt.Sprintf("%s%s", baseAnimeSearchURL, url.PathEscape(animeName))

	log.Println("🔍 Searching Anime for GUI:", searchURL)

	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch anime search page")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error: Server returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse anime search page")
	}

	var animes []GUIAnime
	animeSet := make(map[string]bool)

	// Extract anime results
	doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Text())
		animeURL, exists := s.Attr("href")
		if !exists || name == "" {
			return
		}

		animeURL = resolveURL(baseAnimeSearchURL, animeURL)

		if !animeSet[name] {
			animeSet[name] = true
			animes = append(animes, GUIAnime{Name: name, URL: animeURL})
		}
	})

	if len(animes) == 0 {
		return nil, errors.New("No anime found")
	}

	return animes, nil
}

// ✅ **Fetch episodes for the selected anime (REAL API)**
func FetchEpisodes(anime *GUIAnime) error {
	log.Println("🔍 Fetching episodes for:", anime.Name, "->", anime.URL)

	resp, err := http.Get(anime.URL)
	if err != nil {
		return errors.Wrap(err, "Failed to fetch anime episode list")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to load anime page: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Failed to parse anime page")
	}

	var episodes []GUIEpisode

	// ✅ **Ensure we're scraping the correct episode list section**
	doc.Find(".episodes-list a").Each(func(i int, s *goquery.Selection) {
		episodeNumber := strings.TrimSpace(s.Text())
		episodeURL, exists := s.Attr("href")
		if !exists || episodeNumber == "" {
			return
		}

		episodeURL = resolveURL(baseAnimeSearchURL, episodeURL)
		log.Println("✅ Episode Found:", episodeNumber, "->", episodeURL)

		episodes = append(episodes, GUIEpisode{Number: episodeNumber, URL: episodeURL})
	})

	if len(episodes) == 0 {
		return errors.New("No episodes found")
	}

	anime.Episodes = episodes
	return nil
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
