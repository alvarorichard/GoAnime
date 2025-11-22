// Example: Get stream URL for an episode
package main

import (
	"fmt"
	"log"

	"github.com/alvarorichard/Goanime/pkg/goanime"
	"github.com/alvarorichard/Goanime/pkg/goanime/types"
)

func main() {
	client := goanime.NewClient()

	// Search for anime
	animeName := "Demon Slayer"
	fmt.Printf("Searching for '%s'...\n", animeName)
	results, err := client.SearchAnime(animeName, nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(results) == 0 {
		log.Fatal("No anime found")
	}

	// Select first result
	anime := results[0]
	fmt.Printf("Selected: %s [%s]\n", anime.Name, anime.Source)

	// Parse source
	source, err := types.ParseSource(anime.Source)
	if err != nil {
		log.Fatal(err)
	}

	// Get episodes
	fmt.Println("Fetching episodes...")
	episodes, err := client.GetAnimeEpisodes(anime.URL, source)
	if err != nil {
		log.Fatal(err)
	}

	if len(episodes) == 0 {
		log.Fatal("No episodes found")
	}

	// Get stream URL for first episode
	episode := episodes[0]
	fmt.Printf("\nGetting stream URL for Episode %s...\n", episode.Number)

	streamURL, headers, err := client.GetStreamURL(episode.URL, source)
	if err != nil {
		log.Fatalf("Error getting stream URL: %v", err)
	}

	fmt.Println("\n=== Stream Information ===")
	fmt.Printf("Episode: %s\n", episode.Number)
	fmt.Printf("Stream URL: %s\n", streamURL)

	if len(headers) > 0 {
		fmt.Println("\nRequired Headers:")
		for key, value := range headers {
			fmt.Printf("  %s: %s\n", key, value)
		}
	} else {
		fmt.Println("\nNo special headers required")
	}

	fmt.Println("\nYou can use this URL with video players like mpv, vlc, or ffmpeg")
	fmt.Printf("Example: mpv \"%s\"\n", streamURL)
}
