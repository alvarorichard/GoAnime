package api

import (
	"fmt"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/hugolgst/rich-go/client"
	"log"
)

// DiscordPresence updates Discord Rich Presence with anime details and cover link
func DiscordPresence(clientId string, anime Anime, isPaused bool, timestamp int64) error {
	// Login to Discord
	err := client.Login(clientId)
	if err != nil {
		return fmt.Errorf("failed to login to Discord: %w", err)
	}

	// Define the state message based on whether the episode is paused
	var state string
	if isPaused {
		state = fmt.Sprintf("Watching Episode %d (Paused)", anime.Episodes[0].Num)
	} else {
		state = fmt.Sprintf("Watching Episode %d", anime.Episodes[0].Num)
	}

	// Set up the activity for Discord Rich Presence without a LargeImage key
	activity := client.Activity{
		Details:    anime.Name,
		LargeImage: anime.ImageURL,
		LargeText:  anime.Name,
		State:      state,
		Buttons: []*client.Button{
			{
				Label: "View on AniList",
				Url:   fmt.Sprintf("https://anilist.co/anime/%d", anime.AnilistID),
			},
			{
				Label: "View on MAL",                                                // Button label
				Url:   fmt.Sprintf("https://myanimelist.net/anime/%d", anime.MalID), // Button link
			},
		},
	}

	// Set the activity
	err = client.SetActivity(activity)
	if err != nil {
		return fmt.Errorf("failed to set Discord activity: %w", err)
	}

	if util.IsDebug {
		log.Println("Discord Rich Presence updated successfully with cover image link and other details.")

	}

	return nil
}
