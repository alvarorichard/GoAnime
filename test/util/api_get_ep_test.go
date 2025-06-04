package test_util

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strconv"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"

	"github.com/alvarorichard/Goanime/internal/util"
)

// Mock SafeGet function for testing
func SafeGet(url string) (*http.Response, error) {
	return http.Get(url)
}

// Define the Episode struct
type Episode struct {
	Number string
	Num    int
	URL    string
}

// Test function for GetAnimeEpisodes
func TestGetAnimeEpisodes(t *testing.T) {
	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a HTML response
		html := `
		<html>
			<body>
				<a class="lEp epT divNumEp smallbox px-2 mx-1 text-left d-flex" href="/episode1">Episode 1</a>
				<a class="lEp epT divNumEp smallbox px-2 mx-1 text-left d-flex" href="/episode2">Episode 2</a>
			</body>
		</html>`
		_, err := w.Write([]byte(html))
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
			return
		}
	}))
	defer mockServer.Close()

	// Call the GetAnimeEpisodes function with the mock server URL
	episodes, err := GetAnimeEpisodes(mockServer.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Validate the results
	expectedEpisodes := []Episode{
		{Number: "Episode 1", Num: 1, URL: "/episode1"},
		{Number: "Episode 2", Num: 2, URL: "/episode2"},
	}

	if len(episodes) != len(expectedEpisodes) {
		t.Fatalf("Expected %d episodes, got %d", len(expectedEpisodes), len(episodes))
	}

	for i, episode := range episodes {
		if episode.Number != expectedEpisodes[i].Number || episode.Num != expectedEpisodes[i].Num || episode.URL != expectedEpisodes[i].URL {
			t.Errorf("Expected episode %v, got %v", expectedEpisodes[i], episode)
		}
	}
}

// Helper functions for GetAnimeEpisodes
func GetAnimeEpisodes(animeURL string) ([]Episode, error) {
	resp, err := SafeGet(animeURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get anime details")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			util.Debugf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse anime details")
	}

	episodes := parseEpisodes(doc)
	sortEpisodesByNum(episodes)

	return episodes, nil
}

func parseEpisodes(doc *goquery.Document) []Episode {
	var episodes []Episode
	doc.Find("a.lEp.epT.divNumEp.smallbox.px-2.mx-1.text-left.d-flex").Each(func(i int, s *goquery.Selection) {
		episodeNum := s.Text()
		episodeURL, _ := s.Attr("href")

		num, err := parseEpisodeNumber(episodeNum)
		if err != nil {
			util.Debugf("Error parsing episode number '%s': %v", episodeNum, err)
			return
		}

		episodes = append(episodes, Episode{
			Number: episodeNum,
			Num:    num,
			URL:    episodeURL,
		})
	})
	return episodes
}

func parseEpisodeNumber(episodeNum string) (int, error) {
	numRe := regexp.MustCompile(`\d+`)
	numStr := numRe.FindString(episodeNum)
	if numStr == "" {
		return 0, errors.New("no episode number found")
	}
	return strconv.Atoi(numStr)
}

func sortEpisodesByNum(episodes []Episode) {
	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].Num < episodes[j].Num
	})
}
