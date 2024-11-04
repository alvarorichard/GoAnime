//package api
//
//import (
//	"fmt"
//	"github.com/hugolgst/rich-go/client"
//)
//
//func DiscordPresence(clientId string, anime Anime, isPaused bool) error {
//	err := client.Login(clientId)
//	if err != nil {
//		return err
//	}
//
//	// Ensure that there's at least one episode in the slice
//	var episodeNumber string
//	if len(anime.Episodes) > 0 {
//		episodeNumber = anime.Episodes[0].Number
//	} else {
//		episodeNumber = "N/A"
//	}
//
//	var state string
//	if isPaused {
//		state = fmt.Sprintf("Episode %s (Paused)", episodeNumber)
//	} else {
//		state = fmt.Sprintf("Watching Episode %s", episodeNumber)
//	}
//
//	err = client.SetActivity(client.Activity{
//		Details:    anime.Name,
//		State:      state,
//		LargeImage: "anime_placeholder", // Replace with your uploaded asset key
//		LargeText:  anime.Name,
//		// Buttons are omitted since IDs are not available
//	})
//	if err != nil {
//		return err
//	}
//	return nil
//}

package api

import (
	"fmt"
	"github.com/hugolgst/rich-go/client"
	"log"
)

// DiscordPresence updates the Discord Rich Presence with anime details
func DiscordPresence(clientId string, anime Anime, isPaused bool) error {
	// Login to Discord
	err := client.Login(clientId)
	if err != nil {
		return fmt.Errorf("failed to login to Discord: %w", err)
	}

	// Define the state message based on whether the episode is paused
	state := fmt.Sprintf("Assistindo %s", anime.Name)
	if isPaused {
		state += " (Pausado)"
	}

	// Set the activity for Discord Rich Presence
	activity := client.Activity{
		Details:    anime.Name,
		State:      state + " | Ver capa no link abaixo",
		LargeImage: "anime_placeholder",
		LargeText:  anime.Name,
		Buttons: []*client.Button{
			{
				Label: "Ver capa do anime",
				Url:   anime.ImageURL,
			},
			{
				Label: "Ver no site",
				Url:   anime.URL,
			},
		},
	}

	err = client.SetActivity(activity)
	if err != nil {
		return fmt.Errorf("failed to set Discord activity: %w", err)
	}

	log.Println("Discord Rich Presence updated successfully with cover link.")
	return nil
}

//func StartDiscordPresenceUpdater(clientId string, anime Anime, isPaused bool) {
//	ticker := time.NewTicker(5 * time.Second) // Update interval
//	defer ticker.Stop()
//
//	for {
//		select {
//		case <-ticker.C:
//			// Update Discord presence
//			err := DiscordPresence(clientId, anime, isPaused)
//			if err != nil {
//				log.Printf("Error updating Discord presence: %v", err)
//			}
//		}
//	}
//}
