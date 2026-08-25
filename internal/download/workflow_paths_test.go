package download

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// enrichRecorder records how HandleDownloadRequest drives the enrichment seam.
type enrichRecorder struct {
	called bool
	anime  *models.Anime
	err    error
}

// stubWorkflow points both workflow seams at offline fakes and turns on strict
// source resolution so providers.FetchEpisodes fails fast instead of guessing
// a live source. The anime URL is loopback, which SafeGet rejects before
// dialing (SSRF guard), so the legacy episode fetch also fails fast — every
// path is deterministic and touches no real network.
// These tests mutate package seams and process env, so none are parallel.
func stubWorkflow(t *testing.T, anime *models.Anime) *enrichRecorder {
	t.Helper()
	t.Setenv("GOANIME_STRICT_SOURCE", "1")

	rec := &enrichRecorder{}
	prevSearch, prevEnrich := workflowSearchFn, workflowEnrichFn
	prevMeta := player.GetMediaMeta()
	workflowSearchFn = func(_ string) (*models.Anime, error) { return anime, nil }
	workflowEnrichFn = func(_ context.Context, a *models.Anime) ([]metadata.SeasonMapping, error) {
		rec.called = true
		rec.anime = a
		return []metadata.SeasonMapping{}, rec.err
	}
	t.Cleanup(func() {
		workflowSearchFn = prevSearch
		workflowEnrichFn = prevEnrich
		player.SetMediaMeta(prevMeta)
	})
	return rec
}

// offlineAnime returns an anime whose URL is loopback (blocked by SafeGet) and
// whose source is unrecognizable (errors under GOANIME_STRICT_SOURCE).
func offlineAnime() *models.Anime {
	return &models.Anime{
		Name:      "Workflow Test Anime",
		URL:       "http://127.0.0.1:9/blocked",
		MediaType: models.MediaTypeAnime,
		AnilistID: 42,
		MalID:     7,
	}
}

func TestHandleDownloadRequest_SingleEpisode_LegacyFetchFails(t *testing.T) {
	anime := offlineAnime()
	rec := stubWorkflow(t, anime)

	req := &util.DownloadRequest{AnimeName: "test", EpisodeNum: 1}
	err := HandleDownloadRequest(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch episodes")
	assert.True(t, rec.called, "enrichment must run before the episode fetch")
	assert.Same(t, anime, rec.anime)
}

func TestHandleDownloadRequest_AllEpisodes_FallbackCascade(t *testing.T) {
	// Cascade: providers.FetchEpisodes fails (strict source) → batch download
	// is skipped → legacy fetch fails (loopback blocked) → error surfaces.
	stubWorkflow(t, offlineAnime())

	req := &util.DownloadRequest{AnimeName: "test", IsAll: true}
	err := HandleDownloadRequest(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch episodes")
}

func TestHandleDownloadRequest_Range_FallbackCascade(t *testing.T) {
	stubWorkflow(t, offlineAnime())

	req := &util.DownloadRequest{
		AnimeName:    "test",
		IsRange:      true,
		StartEpisode: 1,
		EndEpisode:   3,
	}
	err := HandleDownloadRequest(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch episodes")
}

func TestHandleDownloadRequest_SmartRange_RequiresAllAnimeSource(t *testing.T) {
	// AllAnimeSmart set but the anime is not from AllAnime: the smart branch
	// must be skipped and the normal range cascade taken instead.
	stubWorkflow(t, offlineAnime())

	req := &util.DownloadRequest{
		AnimeName:    "test",
		IsRange:      true,
		StartEpisode: 2,
		EndEpisode:   4,
		Source:       "animefire",
	}
	err := HandleDownloadRequest(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch episodes",
		"non-AllAnime smart request must fall through to the normal range path")
}

func TestHandleDownloadRequest_EnrichErrorIsNonFatal(t *testing.T) {
	anime := offlineAnime()
	rec := stubWorkflow(t, anime)
	rec.err = errors.New("anilist down")

	req := &util.DownloadRequest{AnimeName: "test", EpisodeNum: 1}
	err := HandleDownloadRequest(req)

	// The enrichment error is swallowed; the flow proceeds to the episode
	// fetch and fails there, not on enrichment.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch episodes")
	assert.True(t, rec.called)
}

func TestHandleDownloadRequest_MediaMetaPropagated(t *testing.T) {
	anime := offlineAnime()
	stubWorkflow(t, anime)

	req := &util.DownloadRequest{AnimeName: "test", EpisodeNum: 1, SeasonNum: 3, Quality: "720p"}
	_ = HandleDownloadRequest(req)

	meta := player.GetMediaMeta()
	require.NotNil(t, meta)
	assert.Equal(t, 42, meta.AnilistID)
	assert.Equal(t, 7, meta.MalID)
	assert.Equal(t, anime.OfficialTitle(), meta.OfficialTitle)
}
