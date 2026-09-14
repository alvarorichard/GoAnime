package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// aniListDisabledBody is AniList's own outage response, captured live on
// 2026-09-09 from graphql.anilist.co. Every request got it — the API
// User-Agent, no User-Agent, curl's default and a browser User-Agent alike —
// while anilist.co itself still served 200.
const aniListDisabledBody = `{"errors":[{"message":"The AniList API has been temporarily disabled due to severe stability issues.","status":403,"locations":[{"line":1,"column":1}]}],"data":null}`

func TestBodySaysAniListDisabled(t *testing.T) {
	t.Parallel()

	assert.True(t, bodySaysAniListDisabled([]byte(aniListDisabledBody)),
		"the real outage body must be recognised")

	// A reworded notice should still trip one of the two markers.
	assert.True(t, bodySaysAniListDisabled(
		[]byte(`{"errors":[{"message":"The AniList API has been temporarily disabled for maintenance."}]}`)))
	assert.True(t, bodySaysAniListDisabled(
		[]byte(`{"errors":[{"message":"Down: severe stability issues ongoing."}]}`)))

	// Ordinary rejections must NOT latch the breaker: a proxy block or a rate
	// limit is transient and retrying the next title is the right move.
	for _, body := range []string{
		"",
		`{"errors":[{"message":"Too Many Requests"}]}`,
		`{"errors":[{"message":"Not Found"}]}`,
		"<html><body>403 Forbidden</body></html>",
		`{"errors":[{"message":"validation error on field search"}]}`,
	} {
		assert.Falsef(t, bodySaysAniListDisabled([]byte(body)),
			"%q is an ordinary failure, not an announced outage", body)
	}
}

func TestAniListDisabledLatch(t *testing.T) {
	resetAniListDisabledForTest()
	t.Cleanup(resetAniListDisabledForTest)

	assert.False(t, aniListIsDisabled(), "starts closed")
	noteAniListDisabled()
	assert.True(t, aniListIsDisabled())
	noteAniListDisabled() // idempotent
	assert.True(t, aniListIsDisabled())
}

// TestFetchAniList_StopsOnAnnouncedOutage_2026_09_09 is the regression for the
// reported warning:
//
//	Metadata enrichment unavailable; continuing without it
//	error="AniList enrichment failed: AniList returned: 403 Forbidden"
//
// A bare "403 Forbidden" reads like a local misconfiguration. It was not: the
// API had been switched off upstream. The lookup must say so by name, and must
// not keep asking — neither the remaining search variations for this title nor
// AniList at all for the rest of the session.
func TestFetchAniList_StopsOnAnnouncedOutage_2026_09_09(t *testing.T) {
	resetAniListDisabledForTest()
	t.Cleanup(resetAniListDisabledForTest)

	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(aniListDisabledBody))
	}))
	defer srv.Close()

	prev := aniListEndpoint
	aniListEndpoint = srv.URL
	t.Cleanup(func() { aniListEndpoint = prev })

	_, err := FetchAnimeFromAniListWithURL("Naruto Shippuden", "https://animefire.plus/animes/naruto-shippuden")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAniListAPIDisabled,
		"the outage must be named, not reported as a bare 403")
	assert.Equal(t, 1, calls, "an API that says it is off must not be asked again for this title")

	// And the latch must spare the rest of the session.
	_, err = FetchAnimeFromAniListWithURL("One Piece", "https://animefire.plus/animes/one-piece")
	assert.ErrorIs(t, err, ErrAniListAPIDisabled)
	assert.Equal(t, 1, calls, "later titles must not re-ask a switched-off API")
}

// An ordinary 403 must keep the old behaviour: try the other search variations
// and leave the breaker closed, because it may be transient.
func TestFetchAniList_OrdinaryFailureDoesNotLatch(t *testing.T) {
	resetAniListDisabledForTest()
	t.Cleanup(resetAniListDisabledForTest)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"errors":[{"message":"Too Many Requests"}]}`))
	}))
	defer srv.Close()

	prev := aniListEndpoint
	aniListEndpoint = srv.URL
	t.Cleanup(func() { aniListEndpoint = prev })

	_, err := FetchAnimeFromAniListWithURL("Naruto Shippuden", "")
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrAniListAPIDisabled)
	assert.False(t, aniListIsDisabled(), "a transient rejection must not disable AniList for the session")
}

// TestFetchAnimeFromJikan_ParsesTheRealPayload runs the parser against a
// response captured live from api.jikan.moe, so the field mapping is pinned to
// what the service actually sends rather than to a hand-written guess.
func TestFetchAnimeFromJikan_ParsesTheRealPayload(t *testing.T) {
	payload, err := os.ReadFile("testdata/jikan_naruto_shippuden.json")
	require.NoError(t, err)

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	prev := jikanBaseURL
	jikanBaseURL = srv.URL
	t.Cleanup(func() { jikanBaseURL = prev })

	res, err := fetchAnimeFromJikan("[PT-BR] Naruto Shippuden (Dublado)")
	require.NoError(t, err)

	assert.Equal(t, "Naruto Shippuden", gotQuery,
		"the source's dub/language tags must be stripped before searching")

	m := res.Data.Media
	assert.Equal(t, 1735, m.IDMal)
	assert.Zero(t, m.ID, "Jikan indexes MyAnimeList and has no AniList id")
	assert.Equal(t, "Naruto: Shippuuden", m.Title.Romaji)
	assert.Equal(t, "Naruto Shippuden", m.Title.English)
	assert.NotEmpty(t, m.Title.Native)
	assert.Contains(t, m.Synonyms, "Naruto Hurricane Chronicles")
	assert.NotEmpty(t, m.CoverImage.Large, "the cover art is what the UI shows")
	assert.NotEmpty(t, m.Description)
	assert.NotEmpty(t, m.Genres)
	assert.Positive(t, m.AverageScore, "Jikan's 0-10 score must be rescaled to AniList's 0-100")
	assert.LessOrEqual(t, m.AverageScore, 100)
}

func TestFetchAnimeFromJikan_NoMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	prev := jikanBaseURL
	jikanBaseURL = srv.URL
	t.Cleanup(func() { jikanBaseURL = prev })

	_, err := fetchAnimeFromJikan("something that does not exist")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no matching anime")
}

// TestEnrichAnimeData_FallsBackToJikanDuringOutage is the end-to-end guard: with
// AniList switched off, enrichment must still fill the cover art, MAL id and
// titles rather than returning an error and leaving the anime bare.
func TestEnrichAnimeData_FallsBackToJikanDuringOutage(t *testing.T) {
	resetAniListDisabledForTest()
	t.Cleanup(resetAniListDisabledForTest)

	payload, err := os.ReadFile("testdata/jikan_naruto_shippuden.json")
	require.NoError(t, err)

	anilist := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(aniListDisabledBody))
	}))
	defer anilist.Close()
	jikan := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer jikan.Close()

	prevA, prevJ := aniListEndpoint, jikanBaseURL
	aniListEndpoint, jikanBaseURL = anilist.URL, jikan.URL
	t.Cleanup(func() { aniListEndpoint, jikanBaseURL = prevA, prevJ })

	anime := &models.Anime{Name: "[PT-BR] Naruto Shippuden (Dublado)", URL: "https://animefire.plus/animes/naruto-shippuden"}
	require.NoError(t, enrichAnimeData(anime), "an upstream AniList outage must not fail enrichment")

	assert.Equal(t, 1735, anime.MalID)
	assert.Zero(t, anime.AnilistID, "no AniList id is available, and both readers guard on > 0")
	assert.NotEmpty(t, anime.ImageURL, "the cover art must survive the outage")
	assert.Equal(t, "Naruto: Shippuuden", anime.Details.Title.Romaji)
}

// If MyAnimeList is down too — as it was while this was written — the error has
// to name both, so the log points at the outages rather than at GoAnime.
func TestEnrichAnimeData_ReportsBothOutages(t *testing.T) {
	resetAniListDisabledForTest()
	t.Cleanup(resetAniListDisabledForTest)

	anilist := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(aniListDisabledBody))
	}))
	defer anilist.Close()
	jikan := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
	}))
	defer jikan.Close()

	prevA, prevJ := aniListEndpoint, jikanBaseURL
	aniListEndpoint, jikanBaseURL = anilist.URL, jikan.URL
	t.Cleanup(func() { aniListEndpoint, jikanBaseURL = prevA, prevJ })

	err := enrichAnimeData(&models.Anime{Name: "Naruto Shippuden", URL: "https://animefire.plus/animes/x"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrAniListAPIDisabled), "the primary outage must stay identifiable")
	assert.Contains(t, err.Error(), "MyAnimeList fallback also failed")
	assert.Contains(t, err.Error(), "504")
}

// TestNoEarlierTestLeftAniListDisabled is a canary for the failure mode that
// turned this change into a broken suite.
//
// One test used to query the real graphql.anilist.co and shrug off the result.
// That was harmless only while AniList was up: once it started answering with
// its "API disabled" notice, that live call latched the process-wide breaker,
// and TestZenpenKouhenBugFix — which runs later and has nothing to do with
// outages — failed with an error it had never asked for.
//
// The latch is package state, so a test that trips it against the LIVE API
// poisons its neighbours. This asserts nobody has, at the point it runs; it
// only sees leaks from tests ordered before it, which is exactly where the
// offending one sat.
func TestNoEarlierTestLeftAniListDisabled(t *testing.T) {
	assert.False(t, aniListIsDisabled(),
		"a test latched the AniList outage breaker and did not reset it — "+
			"most likely by calling the live graphql.anilist.co, which unit tests must not do")
}

// TestEnrichAnimeData_LatchIsScopedToTheOutage pins that only an announced
// outage disables AniList for the session. Anything else — a timeout, a 500, a
// captive portal — must leave the next title free to try again.
func TestEnrichAnimeData_LatchIsScopedToTheOutage(t *testing.T) {
	resetAniListDisabledForTest()
	t.Cleanup(resetAniListDisabledForTest)

	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"message":"Internal Server Error"}]}`))
	}))
	defer srv.Close()

	prev := aniListEndpoint
	aniListEndpoint = srv.URL
	t.Cleanup(func() { aniListEndpoint = prev })

	for range 2 {
		_, err := FetchAnimeFromAniListWithURL("Some Show", "")
		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrAniListAPIDisabled)
	}
	assert.False(t, aniListIsDisabled(), "a 500 is transient; the session must keep trying")
	assert.GreaterOrEqual(t, hits, 2, "each title must get its own attempt")
}

// TestJikanSearch_RetriesTransientFailures covers the flakiness measured on
// 2026-09-09: a burst of 504s from api.jikan.moe followed, minutes later, by
// twelve consecutive successes with nothing changed on this side. A single
// attempt would have turned that blip into "no metadata".
func TestJikanSearch_RetriesTransientFailures(t *testing.T) {
	var attempts int
	payload, err := os.ReadFile("testdata/jikan_naruto_shippuden.json")
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < jikanAttempts {
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	prev := jikanBaseURL
	jikanBaseURL = srv.URL
	t.Cleanup(func() { jikanBaseURL = prev })

	res, err := fetchAnimeFromJikan("Naruto Shippuden")
	require.NoError(t, err, "a transient 504 must not cost the metadata")
	assert.Equal(t, 1735, res.Data.Media.IDMal)
	assert.Equal(t, jikanAttempts, attempts)
}

// A permanent rejection must fail on the first try: repeating it only delays a
// user who is waiting on playback.
func TestJikanSearch_DoesNotRetryPermanentFailures(t *testing.T) {
	var attempts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	prev := jikanBaseURL
	jikanBaseURL = srv.URL
	t.Cleanup(func() { jikanBaseURL = prev })

	_, err := fetchAnimeFromJikan("Naruto Shippuden")
	require.Error(t, err)
	assert.Equal(t, 1, attempts, "a 404 is an answer, not a blip")
}

// Retries are bounded: a service that is down must not stall enrichment.
func TestJikanSearch_GivesUp(t *testing.T) {
	var attempts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(http.StatusGatewayTimeout)
	}))
	defer srv.Close()

	prev := jikanBaseURL
	jikanBaseURL = srv.URL
	t.Cleanup(func() { jikanBaseURL = prev })

	_, err := fetchAnimeFromJikan("Naruto Shippuden")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "504")
	assert.Equal(t, jikanAttempts, attempts, "attempts must be capped")
}

func TestTransientJikanStatus(t *testing.T) {
	t.Parallel()
	for _, code := range []int{500, 502, 503, 504, 429} {
		assert.Truef(t, transientJikanStatus(code), "%d is worth another attempt", code)
	}
	for _, code := range []int{200, 400, 403, 404} {
		assert.Falsef(t, transientJikanStatus(code), "%d is a real answer", code)
	}
}
