package api

import (
	"log"
	"regexp"
	"sort"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

func GetAnimeEpisodes(animeURL string) ([]Episode, error) {
	resp, err := SafeGet(animeURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get anime details")
	}
	defer resp.Body.Close()

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
			log.Printf("Error parsing episode number '%s': %v", episodeNum, err)
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
		numStr = "1"
	}
	return strconv.Atoi(numStr)
}

func sortEpisodesByNum(episodes []Episode) {
	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].Num < episodes[j].Num
	})
}
