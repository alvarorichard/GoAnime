// Example: Get stream URL for an episode
package main

import (
	"fmt"
	"log"

	"github.com/alvarorichard/Goanime/v2/pkg/goanime"
	"github.com/alvarorichard/Goanime/v2/pkg/goanime/types"
)

func main() {
	client := goanime.NewClient()

	// Episode lists and stream URLs are only available for AnimeFire through
	// the library, so search that source rather than all of them.
	source := types.SourceAnimeFire
	animeName := "Demon Slayer"
	fmt.Printf("Searching for '%s'...\n", animeName)
	results, err := client.SearchAnime(animeName, &source)
	if err != nil {
		log.Fatal(err)
	}

	if len(results) == 0 {
		log.Fatal("No anime found")
	}

	// Select first result
	anime := results[0]
	fmt.Printf("Selected: %s [%s]\n", anime.Name, anime.Source)

	// Get episodes
	fmt.Println("Fetching episodes...")
	episodes, err := client.GetAnimeEpisodes(anime.URL, source)
	if err != nil {
		log.Fatal(err)
	}

	if len(episodes) == 0 {
		log.Fatal("No episodes found")
	}

	// Get the stream URL for the first episode
	episode := episodes[0]
	fmt.Printf("\nGetting stream URL for Episode %s...\n", episode.Number)

	streamURL, metadata, err := client.GetEpisodeStreamURL(anime, episode, nil)
	if err != nil {
		log.Fatalf("Error getting stream URL: %v", err)
	}

	fmt.Println("\n=== Stream Information ===")
	fmt.Printf("Episode: %s\n", episode.Number)
	fmt.Printf("Stream URL: %s\n", streamURL)

	if len(metadata) > 0 {
		fmt.Println("\nMetadata:")
		for key, value := range metadata {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	// AnimeFire's video host only needs a User-Agent, which mpv sends.
	fmt.Println("\nPlay it with mpv:")
	fmt.Printf("mpv \"%s\"\n", streamURL)
}
