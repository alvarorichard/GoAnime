package handlers

import (
	"fmt"

	"github.com/alvarorichard/Goanime/internal/download"
	"github.com/alvarorichard/Goanime/internal/util"
)

// HandleDownloadRequest processes download requests
func HandleDownloadRequest() error {
	// Initialize logger for download process
	util.InitLogger()

	if util.GlobalDownloadRequest == nil {
		return fmt.Errorf("download request is nil")
	}

	if err := download.HandleDownloadRequest(util.GlobalDownloadRequest); err != nil {
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

	if err := download.HandleMovieDownloadRequest(util.GlobalDownloadRequest); err != nil {
		return fmt.Errorf("movie download failed: %w", err)
	}
	return nil
}
