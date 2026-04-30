package handlers

import (
	"fmt"

	"github.com/alvarorichard/Goanime/internal/download"
	"github.com/alvarorichard/Goanime/internal/util"
)

var (
	runAnimeDownload = download.HandleDownloadRequest
	runMovieDownload = download.HandleMovieDownloadRequest
)

// HandleDownloadRequest processes download requests
func HandleDownloadRequest() error {
	// Initialize logger for download process
	util.InitLogger()

	if util.GlobalDownloadRequest == nil {
		return fmt.Errorf("download request is nil")
	}

	if err := runAnimeDownload(util.GlobalDownloadRequest); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	return nil
}

// HandleMovieDownloadRequest processes movie/TV download requests from FlixHQ/SFlix
func HandleMovieDownloadRequest() error {
	// Initialize logger for download process
	util.InitLogger()

	if util.GlobalDownloadRequest == nil {
		return fmt.Errorf("movie download request is nil")
	}

	if err := runMovieDownload(util.GlobalDownloadRequest); err != nil {
		return fmt.Errorf("movie download failed: %w", err)
	}
	return nil
}
