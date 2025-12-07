// Example: Get episodes for a specific anime
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
	fmt.Println("Searching for 'Naruto'...")
	results, err := client.SearchAnime("Naruto", nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(results) == 0 {
		log.Fatal("No anime found")
	}

	// Select first result
	anime := results[0]
	fmt.Printf("\nSelected: %s [%s]\n\n", anime.Name, anime.Source)

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

	fmt.Printf("\nFound %d episodes:\n\n", len(episodes))

	// Display first 10 episodes
	displayCount := 10
	if len(episodes) < displayCount {
		displayCount = len(episodes)
	}

	for i := 0; i < displayCount; i++ {
		ep := episodes[i]
		title := "N/A"
		if ep.Title != nil {
			if ep.Title.English != "" {
				title = ep.Title.English
			} else if ep.Title.Romaji != "" {
				title = ep.Title.Romaji
			}
		}

		fmt.Printf("Episode %s: %s\n", ep.Number, title)

		if ep.IsFiller {
			fmt.Println("  [FILLER]")
		}
		if ep.IsRecap {
			fmt.Println("  [RECAP]")
		}
		if ep.Duration > 0 {
			fmt.Printf("  Duration: %d seconds\n", ep.Duration)
		}
	}

	if len(episodes) > displayCount {
		fmt.Printf("\n... and %d more episodes\n", len(episodes)-displayCount)
	}
}
