package api

import (
	"io"
	"regexp"
	"sort"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/pkg/errors"
)

// GetAnimeEpisodes fetches and parses the list of episodes for a given anime.
// It returns a sorted slice of Episode structs, ordered by episode number.
//
// Parameters:
// - animeURL: the URL of the anime's page.
//
// Returns:
// - []models.Episode: a slice of Episode structs, sorted by episode number.
// - error: an error if the process fails at any step.
func GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	// Send an HTTP GET request to retrieve the anime details.
	resp, err := SafeGet(animeURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get anime details")
	}
	// Ensure the response body is closed after the function finishes, and log an error if closing fails.
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			util.Debugf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	// Parse the HTML response using goquery.
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse anime details")
	}

	// Extract the episodes from the parsed HTML document.
	episodes := parseEpisodes(doc)
	// Sort the episodes by their numerical order.
	sortEpisodesByNum(episodes)

	// Return the sorted list of episodes.
	return episodes, nil
}

// parseEpisodes extracts a list of Episode structs from the given goquery.Document.
// It looks for specific HTML elements that contain episode information and returns a slice of Episode structs.
//
// Parameters:
// - doc: a pointer to a goquery.Document which represents the parsed HTML content.
//
// Returns:
// - []models.Episode: a slice of Episode structs extracted from the HTML document.
func parseEpisodes(doc *goquery.Document) []models.Episode {
	var episodes []models.Episode

	// Find all anchor elements within the specified CSS selector that represent episodes and iterate over them.
	doc.Find("a.lEp.epT.divNumEp.smallbox.px-2.mx-1.text-left.d-flex").Each(func(i int, s *goquery.Selection) {
		// Extract the episode number (as text) and the href attribute (URL) from each element.
		episodeNum := s.Text()
		episodeURL, _ := s.Attr("href")

		// Parse the episode number as an integer.
		num, err := parseEpisodeNumber(episodeNum)
		if err != nil {
			util.Debugf("Error parsing episode number '%s': %v", episodeNum, err)
			return
		}

		// Append the parsed episode information to the episodes slice.
		episodes = append(episodes, models.Episode{
			Number: episodeNum,
			Num:    num,
			URL:    episodeURL,
		})
	})
	return episodes
}

// parseEpisodeNumber extracts the numeric portion of an episode number string.
// It uses a regular expression to find the first sequence of digits and returns it as an integer.
//
// Parameters:
// - episodeNum: the string containing the episode number.
//
// Returns:
// - int: the parsed episode number.
// - error: an error if the string cannot be converted to an integer.
func parseEpisodeNumber(episodeStr string) (int, error) {
	re := regexp.MustCompile(`(?i)epis[oó]dio\s+(\d+)`)
	matches := re.FindStringSubmatch(episodeStr)
	if len(matches) >= 2 {
		return strconv.Atoi(matches[1])
	}
	return 1, nil
}

// sortEpisodesByNum sorts a slice of Episode structs in ascending order by the episode number.
//
// Parameters:
// - episodes: a slice of Episode structs to be sorted.
func sortEpisodesByNum(episodes []models.Episode) {
	// Sort the episodes slice in place using the sort.Slice function.
	// The sorting is done based on the Num field of each Episode struct.
	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].Num < episodes[j].Num
	})
}
