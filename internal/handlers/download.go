// Package handlers provides HTTP-level request handlers for download and media operations.
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

	req := util.CurrentDownloadRequest()
	if req == nil {
		return fmt.Errorf("download request is nil")
	}

	if err := download.HandleDownloadRequest(req); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	return nil
}

// HandleMovieDownloadRequest processes movie/TV download requests from FlixHQ/SFlix
func HandleMovieDownloadRequest() error {
	// Initialize logger for download process
	util.InitLogger()

	req := util.CurrentDownloadRequest()
	if req == nil {
		return fmt.Errorf("movie download request is nil")
	}

	if err := download.HandleMovieDownloadRequest(req); err != nil {
		return fmt.Errorf("movie download failed: %w", err)
	}
	return nil
}
