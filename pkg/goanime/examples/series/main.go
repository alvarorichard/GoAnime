// Example: list the seasons of a series from a catalogue source (StartFlix),
// pick a season and an audio, and get the stream URL of its first episode.
package main

import (
	"fmt"
	"log"

	"github.com/alvarorichard/Goanime/v2/pkg/goanime"
	"github.com/alvarorichard/Goanime/v2/pkg/goanime/types"
)

func main() {
	client := goanime.NewClient()

	source := types.SourceStartFlix
	results, err := client.SearchAnime("breaking bad", &source)
	if err != nil {
		log.Fatal(err)
	}
	if len(results) == 0 {
		log.Fatal("no results")
	}
	title := results[0]
	fmt.Println(title.Name)

	seasons, err := client.GetSeasons(title)
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range seasons {
		fmt.Printf("  season %d: %v\n", s.Number, s.Audios)
	}
	if len(seasons) == 0 {
		log.Fatal("a movie has no seasons: use GetAnimeEpisodes")
	}

	// An empty audio takes the dub when the season has one.
	episodes, err := client.GetSeasonEpisodes(title, seasons[0].Number, "Legendado")
	if err != nil {
		log.Fatal(err) // ErrAudioUnavailable says which audios the season has
	}
	fmt.Printf("season %d: %d episodes\n", seasons[0].Number, len(episodes))

	streamURL, metadata, err := client.GetEpisodeStreamURL(title, episodes[0], &goanime.StreamOptions{Audio: "Legendado"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("stream:", streamURL)
	fmt.Println("audio:", metadata["audio"], "referer:", metadata["referer"])
}
