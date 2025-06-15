// Package download provides high-level download workflow management
package download

import (
	"log"

	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/downloader"
	"github.com/alvarorichard/Goanime/internal/util"
)

// HandleDownloadRequest processes a download request from command line
func HandleDownloadRequest(request *util.DownloadRequest) error {
	util.Info("Starting download mode...")

	// Search for anime
	anime := appflow.SearchAnime(request.AnimeName)
	appflow.FetchAnimeDetails(anime)
	episodes := appflow.GetAnimeEpisodes(anime.URL)

	// Create downloader
	downloader := downloader.NewEpisodeDownloader(episodes, anime.URL)

	if request.IsRange {
		util.Infof("Downloading episodes %d-%d of %s",
			request.StartEpisode, request.EndEpisode, anime.Name)
		return downloader.DownloadEpisodeRange(request.StartEpisode, request.EndEpisode)
	} else {
		util.Infof("Downloading episode %d of %s",
			request.EpisodeNum, anime.Name)
		return downloader.DownloadSingleEpisode(request.EpisodeNum)
	}
}

// Example usage functions for documentation

// ExampleSingleDownload demonstrates single episode download
func ExampleSingleDownload() {
	// Command: goanime -d "My Hero Academia" 15
	// This would create a DownloadRequest like:
	request := &util.DownloadRequest{
		AnimeName:  "My Hero Academia",
		EpisodeNum: 15,
		IsRange:    false,
	}

	if err := HandleDownloadRequest(request); err != nil {
		log.Printf("Download failed: %v", err)
	}
}

// ExampleRangeDownload demonstrates episode range download
func ExampleRangeDownload() {
	// Command: goanime -d -r "Attack on Titan" 1-5
	// This would create a DownloadRequest like:
	request := &util.DownloadRequest{
		AnimeName:    "Attack on Titan",
		IsRange:      true,
		StartEpisode: 1,
		EndEpisode:   5,
	}

	if err := HandleDownloadRequest(request); err != nil {
		log.Printf("Range download failed: %v", err)
	}
}
