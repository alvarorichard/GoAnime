////package api
////
////import (
////	"fmt"
////	"github.com/hugolgst/rich-go/client"
////)
////
////func DiscordPresence(clientId string, anime Anime, isPaused bool) error {
////	err := client.Login(clientId)
////	if err != nil {
////		return err
////	}
////
////	// Ensure that there's at least one episode in the slice
////	var episodeNumber string
////	if len(anime.Episodes) > 0 {
////		episodeNumber = anime.Episodes[0].Number
////	} else {
////		episodeNumber = "N/A"
////	}
////
////	var state string
////	if isPaused {
////		state = fmt.Sprintf("Episode %s (Paused)", episodeNumber)
////	} else {
////		state = fmt.Sprintf("Watching Episode %s", episodeNumber)
////	}
////
////	err = client.SetActivity(client.Activity{
////		Details:    anime.Name,
////		State:      state,
////		LargeImage: "anime_placeholder", // Replace with your uploaded asset key
////		LargeText:  anime.Name,
////		// Buttons are omitted since IDs are not available
////	})
////	if err != nil {
////		return err
////	}
////	return nil
////}
//
//package api
//
//import (
//	"fmt"
//	"github.com/alvarorichard/Goanime/internal/util"
//	"github.com/hugolgst/rich-go/client"
//	"log"
//)
//
//// DiscordPresence updates the Discord Rich Presence with anime details
//func DiscordPresence(clientId string, anime Anime, isPaused bool) error {
//	// Login to Discord
//	err := client.Login(clientId)
//	if err != nil {
//		return fmt.Errorf("failed to login to Discord: %w", err)
//	}
//
//	// Define the state message based on whether the episode is paused
//	state := fmt.Sprintf("Assistindo %s", anime.Name)
//	if isPaused {
//		state += " (Pausado)"
//	}
//
//	// Set the activity for Discord Rich Presence
//	activity := client.Activity{
//		Details:    anime.Name,
//		State:      state + " | Ver capa no link abaixo",
//		LargeImage: "anime_placeholder",
//		LargeText:  anime.Name,
//		Buttons: []*client.Button{
//			{
//				Label: "Ver capa do anime",
//				Url:   anime.ImageURL,
//			},
//			{
//				Label: "Ver no site",
//				Url:   anime.URL,
//			},
//		},
//	}
//
//	err = client.SetActivity(activity)
//	if err != nil {
//		return fmt.Errorf("failed to set Discord activity: %w", err)
//	}
//
//	if util.IsDebug {
//		log.Println("Discord Rich Presence updated successfully with cover link.")
//		//return nil
//	}
//
//	return nil
//}
//

package api

import (
	"fmt"
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

	log.Println("Discord Rich Presence updated successfully with cover image link and other details.")
	return nil
}
