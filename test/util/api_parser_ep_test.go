package test_util

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/pkg/errors"

	"github.com/PuerkitoBio/goquery"

	"github.com/alvarorichard/Goanime/internal/util"
)

func TestParseEpisodes(t *testing.T) {
	// HTML simulado para teste
	html := `
	<html>
		<body>
			<a class="lEp epT divNumEp smallbox px-2 mx-1 text-left d-flex" href="/episode1">Episode 1</a>
			<a class="lEp epT divNumEp smallbox px-2 mx-1 text-left d-flex" href="/episode2">Episode 2</a>
		</body>
	</html>`

	// Cria um documento goquery a partir do HTML simulado
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document from reader: %v", err)
	}

	// Chama a função parseEpisodes
	episodes := parseEpisodes(doc)

	// Resultados esperados
	expectedEpisodes := []Episode{
		{Number: "Episode 1", Num: 1, URL: "/episode1"},
		{Number: "Episode 2", Num: 2, URL: "/episode2"},
	}

	// Valida os resultados
	if len(episodes) != len(expectedEpisodes) {
		t.Fatalf("Expected %d episodes, got %d", len(expectedEpisodes), len(episodes))
	}

	for i, episode := range episodes {
		if episode.Number != expectedEpisodes[i].Number || episode.Num != expectedEpisodes[i].Num || episode.URL != expectedEpisodes[i].URL {
			t.Errorf("Expected episode %v, got %v", expectedEpisodes[i], episode)
		}
	}
}

// Helper functions for parseEpisodes test

func ParseEpisodes(doc *goquery.Document) []Episode {
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

func ParseEpisodeNumber(episodeNum string) (int, error) {
	numRe := regexp.MustCompile(`\d+`)
	numStr := numRe.FindString(episodeNum)
	if numStr == "" {
		return 0, errors.New("no episode number found")
	}
	return strconv.Atoi(numStr)
}
