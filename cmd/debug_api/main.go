package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type EpisodeResponse struct {
	Data struct {
		Episode struct {
			EpisodeString string `json:"episodeString"`
			SourceUrls    []struct {
				SourceUrl  string `json:"sourceUrl"`
				SourceName string `json:"sourceName"`
				Type       string `json:"type"`
			} `json:"sourceUrls"`
		} `json:"episode"`
	} `json:"data"`
}

func main() {
	// Test the exact API call that AllAnime makes
	animeID := "2oXgpDPd3xKWdgnoz"
	episodeNo := "1"
	mode := "sub"

	apiBase := "https://api.allanime.day"
	referer := "https://allanime.to"
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	episodeEmbedGQL := `query ($showId: String!, $translationType: VaildTranslationTypeEnumType!, $episodeString: String!) { episode( showId: $showId translationType: $translationType episodeString: $episodeString ) { episodeString sourceUrls }}`
	variables := fmt.Sprintf(`{"showId":"%s","translationType":"%s","episodeString":"%s"}`, animeID, mode, episodeNo)

	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", apiBase+"/api", nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return
	}

	q := req.URL.Query()
	q.Add("variables", variables)
	q.Add("query", episodeEmbedGQL)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Referer", referer)
	req.Header.Set("User-Agent", userAgent)

	fmt.Printf("Making request to: %s\n", req.URL.String())
	fmt.Printf("Variables: %s\n", variables)
	fmt.Printf("Query: %s\n", episodeEmbedGQL)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to make request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", string(body))

	// Try to parse the response
	var episodeResp EpisodeResponse
	if err := json.Unmarshal(body, &episodeResp); err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
	} else {
		fmt.Printf("Parsed successfully!\n")
		fmt.Printf("Episode String: %s\n", episodeResp.Data.Episode.EpisodeString)
		fmt.Printf("Number of source URLs: %d\n", len(episodeResp.Data.Episode.SourceUrls))

		for i, sourceUrl := range episodeResp.Data.Episode.SourceUrls {
			fmt.Printf("Source %d: %s (name: %s, type: %s)\n", i+1, sourceUrl.SourceUrl, sourceUrl.SourceName, sourceUrl.Type)
		}
	}
}
