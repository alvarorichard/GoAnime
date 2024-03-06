package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ktr0731/go-fuzzyfinder"

	"github.com/pkg/errors"
	"log"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
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

func SearchAnime(animeName string) (string, error) {
	currentPageURL := fmt.Sprintf("%s/pesquisar/%s", baseSiteURL, animeName)

	for {
		response, err := http.Get(currentPageURL)
		if err != nil {
			return "", fmt.Errorf("failed to perform search request: %v", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			if response.StatusCode == http.StatusForbidden {
				return "", fmt.Errorf("Connection refused: You need be in Brazil or use a VPN to access the server.")
			}
			return "", fmt.Errorf("Search failed, the server returned the error: %s", response.Status)
		}

		doc, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			return "", fmt.Errorf("failed to parse response: %v", err)
		}

		animes := make([]Anime, 0)
		doc.Find(".row.ml-1.mr-1 a").Each(func(i int, s *goquery.Selection) {
			anime := Anime{
				Name: strings.TrimSpace(s.Text()),
				URL:  s.AttrOr("href", ""),
			}
			animeName = strings.TrimSpace(s.Text())

			animes = append(animes, anime)
		})

		if len(animes) > 0 {
			selectedAnimeName, _ := selectAnimeWithGoFuzzyFinder(animes)
			for _, anime := range animes {
				if anime.Name == selectedAnimeName {
					return anime.URL, nil
				}
			}
		}

		nextPage, exists := doc.Find(".pagination .next a").Attr("href")
		if !exists || nextPage == "" {
			return "", fmt.Errorf("no anime found with the given name")
		}

		currentPageURL = baseSiteURL + nextPage
	}
}

func selectAnimeWithGoFuzzyFinder(animes []Anime) (string, error) {
	if len(animes) == 0 {
		return "", errors.New("no anime provided")
	}

	animeNames := make([]string, len(animes))
	for i, anime := range animes {
		animeNames[i] = anime.Name
	}

	idx, err := fuzzyfinder.Find(
		animeNames,
		func(i int) string {
			return animeNames[i]
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to select anime with go-fuzzyfinder: %v", err)
	}

	if idx < 0 || idx >= len(animes) {
		return "", errors.New("invalid index returned by fuzzyfinder")
	}

	return animes[idx].Name, nil
}

func IsDisallowedIP(hostIP string) bool {
	ip := net.ParseIP(hostIP)
	return ip.IsMulticast() || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate()
}

func SafeTransport(timeout time.Duration) *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(network, addr, timeout)
			if err != nil {
				return nil, err
			}
			ip, _, _ := net.SplitHostPort(c.RemoteAddr().String())
			if IsDisallowedIP(ip) {
				return nil, errors.New("ip address is not allowed")
			}
			return c, err
		},
		DialTLS: func(network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: timeout}
			c, err := tls.DialWithDialer(dialer, network, addr, &tls.Config{
				MinVersion: tls.VersionTLS12, // Set minimum TLS version to 1.3 or 1.2 in case break download
			})
			if err != nil {
				return nil, err
			}

			ip, _, _ := net.SplitHostPort(c.RemoteAddr().String())
			if IsDisallowedIP(ip) {
				return nil, errors.New("ip address is not allowed")
			}

			err = c.Handshake()
			if err != nil {
				return c, err
			}

			return c, c.Handshake()
		},
		TLSHandshakeTimeout: timeout,
	}
}

func GetAnimeEpisodes(animeURL string) ([]Episode, error) {
	resp, err := SafeGet(animeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime details: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse anime details: %v", err)
	}

	episodeContainer := doc.Find("a.lEp.epT.divNumEp.smallbox.px-2.mx-1.text-left.d-flex")

	var episodes []Episode
	episodeContainer.Each(func(i int, s *goquery.Selection) {
		episodeNum := s.Text()
		episodeURL, _ := s.Attr("href")

		// Parse episode number from episodeNum string
		numRe := regexp.MustCompile(`\d+`)
		numStr := numRe.FindString(episodeNum)
		if numStr == "" {
			numStr = "1"
		}
		num, err := strconv.Atoi(numStr)
		if err != nil {
			log.Printf("Error parsing episode number '%s': %v", episodeNum, err)
			return
		}

		episode := Episode{
			Number: episodeNum,
			Num:    num,
			URL:    episodeURL,
		}
		episodes = append(episodes, episode)
	})

	// Sort episodes by Num
	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].Num < episodes[j].Num
	})

	return episodes, nil
}

func SafeGet(url string) (*http.Response, error) {
	const clientConnectTimeout = time.Second * 10
	httpClient := &http.Client{
		Transport: SafeTransport(clientConnectTimeout),
	}
	return httpClient.Get(url)
}

func IsSeries(animeURL string) (bool, int, error) {
	episodes, err := GetAnimeEpisodes(animeURL)
	if err != nil {
		return false, 0, err
	}

	// Retorna true se o número de episódios for maior que 1, indicando uma série
	return len(episodes) > 1, len(episodes), nil
}
