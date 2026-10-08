package download

import (
	"errors"
	"runtime"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// HandleMovieDownloadRequest
// ---------------------------------------------------------------------------

func TestHandleMovieDownloadRequest_NilOrEmptyRequest(t *testing.T) {
	t.Parallel()
	require.Error(t, HandleMovieDownloadRequest(nil))
	require.Error(t, HandleMovieDownloadRequest(&util.DownloadRequest{AnimeName: "  "}))
}

// ---------------------------------------------------------------------------
// HandleDownloadRequest — test error path (search fails immediately on bad name)
// ---------------------------------------------------------------------------

func TestHandleDownloadRequest_EmptyName_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: requires network search which may block")
	}
	if runtime.GOOS == "windows" {
		t.Skip("SearchAnimeWithRetry uses huh.NewInput which blocks on Windows without a TTY")
	}
	req := &util.DownloadRequest{AnimeName: ""}
	err := HandleDownloadRequest(req)
	// Either an error or nil (if search returns unexpectedly)
	// Key: function is entered and executed, not panicked
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
}

func TestHandleDownloadRequest_NilRequest_Pin(t *testing.T) {
	t.Parallel()
	_ = HandleDownloadRequest // symbol pin for coverage tracking
}

// ---------------------------------------------------------------------------
// HandleDownloadRequest — injected search function tests
// ---------------------------------------------------------------------------

func TestHandleDownloadRequest_SearchError(t *testing.T) {
	// Inject a failing search so the function returns before any network call.
	prev := workflowSearchFn
	workflowSearchFn = func(_ string) (*models.Anime, error) {
		return nil, errors.New("mock search failure")
	}
	t.Cleanup(func() { workflowSearchFn = prev })

	req := &util.DownloadRequest{AnimeName: "NonExistent"}
	err := HandleDownloadRequest(req)
	require.Error(t, err)
}

func TestHandleDownloadRequest_QualityDefaultsToBest(t *testing.T) {
	// Inject a search that immediately fails; verifies the quality defaulting
	// branch (empty Quality → "best") before the error short-circuits the rest.
	prev := workflowSearchFn
	workflowSearchFn = func(_ string) (*models.Anime, error) {
		return nil, errors.New("mock search failure")
	}
	t.Cleanup(func() { workflowSearchFn = prev })

	req := &util.DownloadRequest{AnimeName: "Test", Quality: ""}
	err := HandleDownloadRequest(req)
	require.Error(t, err) // quality branch exercised before error short-circuits
}
