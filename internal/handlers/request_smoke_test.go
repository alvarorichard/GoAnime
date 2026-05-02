package handlers

import (
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func restoreHandlerSeams(t *testing.T) {
	t.Helper()

	origDownloadInit := handleDownloadInitLogger
	origDownload := handleDownloadRequestFn
	origMovieDownload := handleMovieDownloadRequestFn
	origUpdateInit := handleUpdateInitLogger
	origUpdateInfo := handleUpdateInfo
	origCheckUpdate := handleCheckAndPromptUpdateFn
	origRequest := util.GlobalDownloadRequest

	t.Cleanup(func() {
		handleDownloadInitLogger = origDownloadInit
		handleDownloadRequestFn = origDownload
		handleMovieDownloadRequestFn = origMovieDownload
		handleUpdateInitLogger = origUpdateInit
		handleUpdateInfo = origUpdateInfo
		handleCheckAndPromptUpdateFn = origCheckUpdate
		util.GlobalDownloadRequest = origRequest
	})

	handleDownloadInitLogger = func() {}
	handleUpdateInitLogger = func() {}
	handleUpdateInfo = func(any, ...any) {}
}

func TestHandleDownloadRequestSmoke(t *testing.T) {
	restoreHandlerSeams(t)

	util.GlobalDownloadRequest = &util.DownloadRequest{AnimeName: "Example", EpisodeNum: 1}
	called := false
	handleDownloadRequestFn = func(req *util.DownloadRequest) error {
		called = true
		require.Equal(t, "Example", req.AnimeName)
		require.Equal(t, 1, req.EpisodeNum)
		return nil
	}

	err := HandleDownloadRequest()
	require.NoError(t, err)
	assert.True(t, called)
}

func TestHandleDownloadRequestRequiresRequest(t *testing.T) {
	restoreHandlerSeams(t)

	util.GlobalDownloadRequest = nil
	err := HandleDownloadRequest()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "download request is nil")
}

func TestHandleMovieDownloadRequestWrapsErrors(t *testing.T) {
	restoreHandlerSeams(t)

	util.GlobalDownloadRequest = &util.DownloadRequest{AnimeName: "Movie", IsMovie: true}
	handleMovieDownloadRequestFn = func(req *util.DownloadRequest) error {
		return errors.New("backend failed")
	}

	err := HandleMovieDownloadRequest()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "movie download failed")
	assert.Contains(t, err.Error(), "backend failed")
}

func TestHandleUpdateRequestSmoke(t *testing.T) {
	restoreHandlerSeams(t)

	called := false
	handleCheckAndPromptUpdateFn = func() error {
		called = true
		return nil
	}

	err := HandleUpdateRequest()
	require.NoError(t, err)
	assert.True(t, called)
}
